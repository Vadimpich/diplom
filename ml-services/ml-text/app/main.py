import asyncio
import json
import logging
import os
import signal
import time
from contextlib import asynccontextmanager
from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Any
from uuid import uuid4

import aio_pika
from aio_pika import ExchangeType, IncomingMessage, Message, RobustChannel, RobustConnection
from fastapi import FastAPI, Response
from fastapi.responses import JSONResponse
from minio import Minio
from minio.error import S3Error
from pydantic import BaseModel, ConfigDict, Field, ValidationError
from app.stt import FasterWhisperSTT, STTError, failed_stt_result, load_stt_config
from app.text_analysis import TEXT_MODEL_VERSION, TextAnalyzer, load_text_analysis_config
from app.telemetry import metrics_payload, ready_payload, test_mode


CHANNEL = "text"
MODEL_VERSION = TEXT_MODEL_VERSION
COMMAND_QUEUE = os.getenv("WORKER_QUEUE_NAME", "qq.processing.text")
COMMAND_ROUTING_KEY = os.getenv("WORKER_COMMAND_ROUTING_KEY", "processing.command.text")



def utc_now() -> datetime:
    return datetime.now(timezone.utc)


def env_bool(name: str, default: bool) -> bool:
    raw = os.getenv(name)
    if raw is None:
        return default
    return raw.strip().lower() in {"1", "true", "yes", "on"}


class AnswerReference(BaseModel):
    answer_id: int
    question_id: int
    audio_s3_bucket: str
    audio_s3_key: str
    answer_text: str = ""


class ProcessingCommandEnvelope(BaseModel):
    model_config = ConfigDict(populate_by_name=True)

    message_version: int
    message_id: str
    correlation_id: str
    request_id: str = ""
    traceparent: str = ""
    tracestate: str = ""
    examination_id: int
    specialist_id: int
    channel: str
    attempt: int
    max_attempts: int
    requested_at: datetime
    answers: list[AnswerReference] = Field(default_factory=list)


@dataclass
class WorkerConfig:
    app_name: str
    listen_port: int
    prefetch_count: int
    max_delivery_attempts: int
    rabbitmq_url: str
    command_exchange: str
    result_exchange: str
    result_routing_key: str
    command_queue: str
    s3_endpoint: str
    s3_access_key: str
    s3_secret_key: str
    s3_use_ssl: bool


class TemporaryProcessingError(Exception):
    def __init__(self, code: str, message: str) -> None:
        super().__init__(message)
        self.code = code
        self.message = message


class FatalProcessingError(Exception):
    def __init__(self, code: str, message: str) -> None:
        super().__init__(message)
        self.code = code
        self.message = message


def build_rabbitmq_url() -> str:
    raw = os.getenv("RABBITMQ_URL")
    if raw:
        return raw

    host = os.getenv("RABBITMQ_HOST", "rabbitmq")
    port = os.getenv("RABBITMQ_PORT", "5672")
    user = os.getenv("RABBITMQ_USER", "diplom")
    password = os.getenv("RABBITMQ_PASSWORD", "diplom_dev_password")
    vhost = os.getenv("RABBITMQ_VHOST", "/").lstrip("/")
    return f"amqp://{user}:{password}@{host}:{port}/{vhost}"


def load_config() -> WorkerConfig:
    return WorkerConfig(
        app_name=os.getenv("WORKER_APP_NAME", "text-worker"),
        listen_port=int(os.getenv("WORKER_PORT", "8080")),
        prefetch_count=int(os.getenv("WORKER_PREFETCH_COUNT", "1")),
        max_delivery_attempts=int(os.getenv("PROCESSING_OUTBOX_MAX_ATTEMPTS", "3")),
        rabbitmq_url=build_rabbitmq_url(),
        command_exchange=os.getenv("WORKER_COMMAND_EXCHANGE", "processing.commands"),
        result_exchange=os.getenv("WORKER_RESULT_EXCHANGE", "processing.results"),
        result_routing_key=os.getenv("WORKER_RESULT_ROUTING_KEY", "processing.result"),
        command_queue=COMMAND_QUEUE,
        s3_endpoint=os.getenv("MINIO_ENDPOINT", "minio:9000").replace("http://", "").replace("https://", ""),
        s3_access_key=os.getenv("MINIO_ACCESS_KEY_ID", os.getenv("MINIO_ROOT_USER", "")),
        s3_secret_key=os.getenv("MINIO_SECRET_ACCESS_KEY", os.getenv("MINIO_ROOT_PASSWORD", "")),
        s3_use_ssl=env_bool("MINIO_USE_SSL", False),
    )


class WorkerState:
    def __init__(self, config: WorkerConfig) -> None:
        self.config = config
        self.logger = logging.getLogger(config.app_name)
        self.connection: RobustConnection | None = None
        self.channel: RobustChannel | None = None
        self.consumer_tag: str | None = None
        self.command_exchange: aio_pika.abc.AbstractRobustExchange | None = None
        self.result_exchange: aio_pika.abc.AbstractRobustExchange | None = None
        self.minio_client = Minio(
            config.s3_endpoint,
            access_key=config.s3_access_key,
            secret_key=config.s3_secret_key,
            secure=config.s3_use_ssl,
        )
        self.stop_event = asyncio.Event()
        self.consumer_ready = False
        self.last_error: str | None = None
        self.stt = FasterWhisperSTT(load_stt_config())
        self.text_analyzer = TextAnalyzer(load_text_analysis_config())

    async def start(self) -> None:
        self.connection = await aio_pika.connect_robust(self.config.rabbitmq_url)
        self.channel = await self.connection.channel()
        await self.channel.set_qos(prefetch_count=self.config.prefetch_count)
        self.command_exchange = await self.channel.declare_exchange(
            self.config.command_exchange,
            ExchangeType.TOPIC,
            durable=True,
        )
        self.result_exchange = await self.channel.declare_exchange(
            self.config.result_exchange,
            ExchangeType.TOPIC,
            durable=True,
        )
        queue = await self.channel.declare_queue(
            self.config.command_queue,
            durable=True,
            arguments={
                "x-queue-type": "quorum",
                "x-dead-letter-exchange": f"{self.config.command_exchange}.dlx",
                "x-delivery-limit": self.config.max_delivery_attempts,
            },
        )
        await queue.bind(self.command_exchange, routing_key=COMMAND_ROUTING_KEY)
        self.consumer_tag = await queue.consume(self.on_message)
        self.consumer_ready = True
        self.last_error = None
        self.logger.info("worker started queue=%s channel=%s", self.config.command_queue, CHANNEL)

    async def stop(self) -> None:
        self.stop_event.set()
        self.consumer_ready = False
        if self.channel and self.consumer_tag:
            await self.channel.cancel(self.consumer_tag)
            self.consumer_tag = None
        if self.channel and not self.channel.is_closed:
            await self.channel.close()
        if self.connection and not self.connection.is_closed:
            await self.connection.close()

    async def on_message(self, message: IncomingMessage) -> None:
        async with message.process(ignore_processed=True):
            result = await self.process_delivery(message.body)
            await self.publish_result(result)

    async def process_delivery(self, body: bytes) -> dict[str, Any]:
        try:
            command = ProcessingCommandEnvelope.model_validate_json(body)
            if command.channel != CHANNEL:
                raise FatalProcessingError("channel_mismatch", f"expected {CHANNEL}, got {command.channel}")
            if not command.answers:
                raise FatalProcessingError("empty_answers", "command does not contain answer references")

            fetched_objects = await asyncio.to_thread(self.fetch_answer_objects, command.answers)
            payload = build_payload(command, fetched_objects)
            return self.build_result(command, "succeeded", payload=payload)
        except ValidationError as exc:
            return self.build_validation_error(exc)
        except FatalProcessingError as exc:
            return self.build_result_from_error(body, "fatal_error", exc.code, exc.message)
        except TemporaryProcessingError as exc:
            return self.build_result_from_error(body, "temporary_error", exc.code, exc.message)
        except Exception as exc:
            self.logger.exception("unexpected worker failure")
            return self.build_result_from_error(body, "temporary_error", "unexpected_error", str(exc))

    def fetch_answer_objects(self, answers: list[AnswerReference]) -> list[dict[str, Any]]:
        objects: list[dict[str, Any]] = []
        for answer in answers:
            response = None
            try:
                response = self.minio_client.get_object(answer.audio_s3_bucket, answer.audio_s3_key)
                content = response.read()
                objects.append(
                    {
                        "answer_id": answer.answer_id,
                        "question_id": answer.question_id,
                        "audio_s3_bucket": answer.audio_s3_bucket,
                        "audio_s3_key": answer.audio_s3_key,
                        "answer_text": answer.answer_text,
                        "bytes": len(content),
                        "content": content,
                    }
                )
            except S3Error as exc:
                if exc.code in {"NoSuchBucket", "NoSuchKey", "AccessDenied"}:
                    raise FatalProcessingError("s3_object_missing", f"{exc.code}: {answer.audio_s3_key}") from exc
                raise TemporaryProcessingError("s3_read_failed", f"{exc.code}: {exc.message}") from exc
            except Exception as exc:
                raise TemporaryProcessingError("s3_read_failed", str(exc)) from exc
            finally:
                try:
                    if response is not None:
                        response.close()
                        response.release_conn()
                except Exception:
                    pass
        return objects

    def build_validation_error(self, exc: ValidationError) -> dict[str, Any]:
        error = exc.errors()[0]
        return {
            "message_version": 1,
            "message_id": str(uuid4()),
            "correlation_id": "",
            "request_id": "",
            "traceparent": "",
            "tracestate": "",
            "examination_id": 0,
            "channel": CHANNEL,
            "attempt": 0,
            "status": "fatal_error",
            "completed_at": utc_now().isoformat(),
            "model_version": MODEL_VERSION,
            "error_code": "invalid_command",
            "error_message": f"{error['loc']}: {error['msg']}",
            "payload": {},
        }

    def build_result(
        self,
        command: ProcessingCommandEnvelope,
        status: str,
        payload: dict[str, Any] | None = None,
        error_code: str | None = None,
        error_message: str | None = None,
    ) -> dict[str, Any]:
        return {
            "message_version": command.message_version,
            "message_id": str(uuid4()),
            "correlation_id": command.correlation_id,
            "request_id": command.request_id,
            "traceparent": command.traceparent,
            "tracestate": command.tracestate,
            "examination_id": command.examination_id,
            "channel": command.channel,
            "attempt": command.attempt,
            "status": status,
            "completed_at": utc_now().isoformat(),
            "model_version": MODEL_VERSION,
            "error_code": error_code,
            "error_message": error_message,
            "payload": payload or {},
        }

    def build_result_from_error(self, body: bytes, status: str, code: str, message: str) -> dict[str, Any]:
        try:
            command = ProcessingCommandEnvelope.model_validate_json(body)
            return self.build_result(command, status, error_code=code, error_message=message)
        except ValidationError:
            return {
                "message_version": 1,
                "message_id": str(uuid4()),
                "correlation_id": "",
                "request_id": "",
                "traceparent": "",
                "tracestate": "",
                "examination_id": 0,
                "channel": CHANNEL,
                "attempt": 0,
                "status": status,
                "completed_at": utc_now().isoformat(),
                "model_version": MODEL_VERSION,
                "error_code": code,
                "error_message": message,
                "payload": {},
            }

    async def publish_result(self, result: dict[str, Any]) -> None:
        if self.result_exchange is None:
            raise RuntimeError("result exchange is not ready")
        await self.result_exchange.publish(
            Message(
                body=json.dumps(result).encode("utf-8"),
                content_type="application/json",
                delivery_mode=aio_pika.DeliveryMode.PERSISTENT,
                message_id=result["message_id"],
                correlation_id=result["correlation_id"] or None,
                timestamp=utc_now(),
            ),
            routing_key=self.config.result_routing_key,
        )

    def health(self) -> dict[str, Any]:
        return {
            "status": "ok" if self.consumer_ready else "starting",
            "worker": self.config.app_name,
            "channel": CHANNEL,
            "queue": self.config.command_queue,
            "consumer_ready": self.consumer_ready,
            "last_error": self.last_error,
            "model_version": MODEL_VERSION,
        }


def build_payload(command: ProcessingCommandEnvelope, fetched_objects: list[dict[str, Any]]) -> dict[str, Any]:
    started_at = time.perf_counter()
    answer_results = []
    for item in fetched_objects:
        result = transcribe_item(item, examination_id=command.examination_id)
        answer_results.append(
            {
                "answer_id": item["answer_id"],
                "question_id": item["question_id"],
                "audio_s3_key": item["audio_s3_key"],
                "bytes": item["bytes"],
                "stt": result["stt"],
                "quality_flags": result["quality_flags"],
                "evidence": result["evidence"],
            }
        )

    payload = aggregate_answer_results(answer_results, examination_id=command.examination_id)
    payload["examination_id"] = command.examination_id
    payload["answer_id"] = answer_results[0]["answer_id"] if len(answer_results) == 1 else None
    payload["processing_time_ms"] = int(round((time.perf_counter() - started_at) * 1000))
    payload["error"] = None
    return payload


def transcribe_item(item: dict[str, Any], *, examination_id: int) -> dict[str, Any]:
    suffix = suffix_from_s3_key(item["audio_s3_key"])
    try:
        stt = state.stt.transcribe_bytes(item["content"], suffix=suffix)
        quality_flags: list[str] = []
        evidence = [
            f"STT transcript length: {len(stt['transcript'])} characters.",
            f"STT word count: {stt['word_count']}.",
            f"STT segments: {len(stt['segments'])}.",
        ]
        if not stt["transcript"]:
            quality_flags.append("empty_transcript")
            evidence.append("STT returned empty transcript.")
        maybe_log_stt_transcript(
            examination_id=examination_id,
            answer_id=item["answer_id"],
            question_id=item["question_id"],
            transcript=stt["transcript"],
        )
        return {"stt": stt, "quality_flags": quality_flags, "evidence": evidence}
    except STTError as exc:
        stt = failed_stt_result(str(exc), model_version=state.stt.model_version)
        return {
            "stt": stt,
            "quality_flags": ["stt_failed"],
            "evidence": [f"STT failed: {exc}."],
        }


def aggregate_answer_results(answer_results: list[dict[str, Any]], *, examination_id: int) -> dict[str, Any]:
    transcripts = [item["stt"]["transcript"].strip() for item in answer_results if item["stt"]["transcript"].strip()]
    transcript = "\n".join(transcripts).strip()
    segments = []
    duration_ms = 0
    for item in answer_results:
        for segment in item["stt"]["segments"]:
            enriched = dict(segment)
            enriched["answer_id"] = item["answer_id"]
            enriched["question_id"] = item["question_id"]
            segments.append(enriched)
        if item["stt"]["duration_ms"] is not None:
            duration_ms += int(item["stt"]["duration_ms"])

    quality_flags = sorted({flag for item in answer_results for flag in item["quality_flags"]})
    languages = [item["stt"]["language"] for item in answer_results if item["stt"]["language"]]
    model_versions = sorted({item["stt"]["model_version"] for item in answer_results if item["stt"]["model_version"]})
    word_count = sum(int(item["stt"]["word_count"]) for item in answer_results)
    stt = {
        "transcript": transcript,
        "language": most_common(languages),
        "segments": segments,
        "duration_ms": duration_ms if duration_ms > 0 else None,
        "word_count": word_count,
        "model_version": model_versions[0] if model_versions else state.stt.model_version,
    }
    maybe_log_stt_transcript(
        examination_id=examination_id,
        answer_id=None,
        question_id=None,
        transcript=transcript,
        aggregate=True,
    )
    evidence = [
        f"STT total transcript length: {len(transcript)} characters.",
        f"STT total word count: {word_count}.",
        f"STT total segments: {len(segments)}.",
    ]
    if quality_flags:
        evidence.append(f"Quality flags: {', '.join(quality_flags)}.")
    text_analysis = state.text_analyzer.analyze(transcript, stt_word_count=word_count)
    quality_flags = sorted(set(quality_flags + text_analysis["quality_flags"]))
    evidence.extend(text_analysis["evidence"])
    return {
        "channel": CHANNEL,
        "status": "done",
        "examination_id": None,
        "answer_id": None,
        "stt": stt,
        "features": text_analysis["features"],
        "scores": text_analysis["scores"],
        "emotion_probs": text_analysis["emotion_probs"],
        "quality_flags": quality_flags,
        "evidence": evidence,
        "model_version": MODEL_VERSION,
        "emotion_model_version": text_analysis["emotion_model_version"],
        "processing_time_ms": 0,
        "error": None,
        "answers": answer_results,
    }


def suffix_from_s3_key(key: str) -> str:
    _, ext = os.path.splitext(key)
    return ext if ext else ".webm"


def most_common(values: list[str]) -> str | None:
    if not values:
        return None
    return max(set(values), key=values.count)


def maybe_log_stt_transcript(
    *,
    examination_id: int,
    answer_id: int | None,
    question_id: int | None,
    transcript: str,
    aggregate: bool = False,
) -> None:
    if not env_bool("TEXT_DEBUG_LOG_STT_TRANSCRIPTS", False):
        return

    max_chars_raw = os.getenv("TEXT_DEBUG_LOG_STT_MAX_CHARS", "1200").strip()
    try:
        max_chars = max(0, int(max_chars_raw))
    except ValueError:
        max_chars = 1200

    normalized = " ".join(transcript.split())
    if max_chars and len(normalized) > max_chars:
        normalized = normalized[:max_chars] + "..."

    if aggregate:
        state.logger.info(
            "stt_debug aggregate examination_id=%s transcript=%r",
            examination_id,
            normalized,
        )
        return

    state.logger.info(
        "stt_debug answer examination_id=%s answer_id=%s question_id=%s transcript=%r",
        examination_id,
        answer_id,
        question_id,
        normalized,
    )


logging.basicConfig(
    level=os.getenv("LOG_LEVEL", "INFO").upper(),
    format="%(asctime)s %(levelname)s %(name)s %(message)s",
)
config = load_config()
state = WorkerState(config)


async def run_worker() -> None:
    while not state.stop_event.is_set():
        try:
            await state.start()
            await state.stop_event.wait()
        except asyncio.CancelledError:
            raise
        except Exception as exc:
            state.last_error = str(exc)
            state.consumer_ready = False
            state.logger.exception("worker loop failed, retrying")
            await asyncio.sleep(5)


@asynccontextmanager
async def lifespan(_: FastAPI):
    if test_mode():
        yield
        return
    worker_task = asyncio.create_task(run_worker())
    try:
        yield
    finally:
        await state.stop()
        worker_task.cancel()
        try:
            await worker_task
        except asyncio.CancelledError:
            pass


app = FastAPI(title=config.app_name, lifespan=lifespan)


@app.get("/health")
async def health() -> dict[str, Any]:
    return state.health()


@app.get("/ready")
async def ready() -> JSONResponse:
    payload, status_code = ready_payload(state, CHANNEL)
    return JSONResponse(content=payload, status_code=status_code)


@app.get("/metrics")
async def metrics() -> Response:
    return Response(metrics_payload(CHANNEL), media_type="text/plain; version=0.0.4")


def _handle_signal(_: int, __: Any) -> None:
    state.stop_event.set()


signal.signal(signal.SIGTERM, _handle_signal)
signal.signal(signal.SIGINT, _handle_signal)
