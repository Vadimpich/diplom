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
from app.analysis import MODEL_VERSION, analyze_audio_bytes
from app.telemetry import metrics_payload, ready_payload, test_mode


CHANNEL = "acoustic"
COMMAND_QUEUE = os.getenv("WORKER_QUEUE_NAME", "qq.processing.acoustic")
COMMAND_ROUTING_KEY = os.getenv("WORKER_COMMAND_ROUTING_KEY", "processing.command.acoustic")


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
        app_name=os.getenv("WORKER_APP_NAME", "acoustic-worker"),
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
        command: ProcessingCommandEnvelope | None = None
        started_at = time.perf_counter()
        try:
            command = ProcessingCommandEnvelope.model_validate_json(body)
            if command.channel != CHANNEL:
                raise FatalProcessingError("channel_mismatch", f"expected {CHANNEL}, got {command.channel}")
            if not command.answers:
                raise FatalProcessingError("empty_answers", "command does not contain answer references")

            self.logger.info(
                "processing_started examination_id=%s channel=%s attempt=%s answers=%s correlation_id=%s message_id=%s",
                command.examination_id,
                command.channel,
                command.attempt,
                len(command.answers),
                command.correlation_id,
                command.message_id,
            )
            fetched_objects = await asyncio.to_thread(self.fetch_answer_objects, command.answers)
            self.logger.info(
                "audio_objects_loaded examination_id=%s channel=%s attempt=%s answers=%s bytes=%s correlation_id=%s",
                command.examination_id,
                command.channel,
                command.attempt,
                len(fetched_objects),
                sum(item["bytes"] for item in fetched_objects),
                command.correlation_id,
            )
            payload = build_payload(command, fetched_objects)
            result = self.build_result(command, "succeeded", payload=payload)
            self.logger.info(
                "processing_succeeded examination_id=%s channel=%s attempt=%s processing_time_ms=%s correlation_id=%s",
                command.examination_id,
                command.channel,
                command.attempt,
                payload.get("processing_time_ms"),
                command.correlation_id,
            )
            return result
        except ValidationError as exc:
            self.logger.warning("processing_invalid_command channel=%s error=%s", CHANNEL, exc)
            return self.build_validation_error(exc)
        except FatalProcessingError as exc:
            self.log_failure(command, "fatal_error", exc.code, exc.message, started_at)
            return self.build_result_from_error(body, "fatal_error", exc.code, exc.message)
        except TemporaryProcessingError as exc:
            self.log_failure(command, "temporary_error", exc.code, exc.message, started_at)
            return self.build_result_from_error(body, "temporary_error", exc.code, exc.message)
        except Exception as exc:
            self.log_failure(command, "temporary_error", "unexpected_error", str(exc), started_at, with_trace=True)
            return self.build_result_from_error(body, "temporary_error", "unexpected_error", str(exc))

    def log_failure(
        self,
        command: ProcessingCommandEnvelope | None,
        status: str,
        code: str,
        message: str,
        started_at: float,
        *,
        with_trace: bool = False,
    ) -> None:
        duration_ms = int(round((time.perf_counter() - started_at) * 1000))
        if command is None:
            self.logger.warning(
                "processing_failed channel=%s status=%s error_code=%s duration_ms=%s error=%s",
                CHANNEL,
                status,
                code,
                duration_ms,
                message,
                exc_info=with_trace,
            )
            return
        self.logger.warning(
            "processing_failed examination_id=%s channel=%s attempt=%s status=%s error_code=%s duration_ms=%s correlation_id=%s error=%s",
            command.examination_id,
            command.channel,
            command.attempt,
            status,
            code,
            duration_ms,
            command.correlation_id,
            message,
            exc_info=with_trace,
        )

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
        self.logger.info(
            "result_published examination_id=%s channel=%s attempt=%s status=%s correlation_id=%s message_id=%s",
            result.get("examination_id"),
            result.get("channel"),
            result.get("attempt"),
            result.get("status"),
            result.get("correlation_id"),
            result.get("message_id"),
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
        result = analyze_audio_bytes(item["content"])
        answer_results.append(
            {
                "answer_id": item["answer_id"],
                "question_id": item["question_id"],
                "audio_s3_key": item["audio_s3_key"],
                "bytes": item["bytes"],
                "result": result,
            }
        )

    payload = aggregate_answer_results(answer_results)
    payload["examination_id"] = command.examination_id
    payload["answer_id"] = answer_results[0]["answer_id"] if len(answer_results) == 1 else None
    payload["processing_time_ms"] = int(round((time.perf_counter() - started_at) * 1000))
    payload["error"] = None
    return payload


def aggregate_answer_results(answer_results: list[dict[str, Any]]) -> dict[str, Any]:
    if not answer_results:
        return analyze_audio_bytes(b"")

    results = [item["result"] for item in answer_results]
    features_list = [item["features"] for item in results]
    scores_list = [item["scores"] for item in results]
    emotion_list = [item.get("emotion_probs", {}) for item in results]
    total_duration = sum(item["duration_ms"] for item in features_list)
    quality_flags = sorted({flag for result in results for flag in result["quality_flags"]})

    features = {
        "duration_ms": total_duration,
        "sample_rate": max((int(item.get("sample_rate") or 0) for item in features_list), default=0),
        "rms_energy_mean": weighted_mean(features_list, "rms_energy_mean"),
    }
    scores = {
        "acoustic_stress_score": round(sum(item["acoustic_stress_score"] for item in scores_list) / len(scores_list), 4),
        "voice_stability_score": round(sum(item["voice_stability_score"] for item in scores_list) / len(scores_list), 4),
        "intensity_variability_score": round(
            sum(item["intensity_variability_score"] for item in scores_list) / len(scores_list),
            4,
        ),
    }
    emotion_probs = weighted_emotion_probs(features_list, emotion_list)
    dominant_emotion = max(emotion_probs.items(), key=lambda item: item[1])[0] if emotion_probs else "neutral"
    evidence = [
        f"Доминирующая эмоция: {dominant_emotion}.",
        f"Средняя энергия RMS: {features['rms_energy_mean']}.",
        f"Обработано аудиоответов: {len(answer_results)}.",
        f"Индекс акустического напряжения: {scores['acoustic_stress_score']}.",
    ]
    if quality_flags:
        evidence.append(f"Флаги качества: {', '.join(quality_flags)}.")

    return {
        "channel": CHANNEL,
        "status": "done",
        "examination_id": None,
        "answer_id": None,
        "features": features,
        "scores": scores,
        "emotion_probs": emotion_probs,
        "dominant_emotion": dominant_emotion,
        "quality_flags": quality_flags,
        "evidence": evidence,
        "model_version": MODEL_VERSION,
        "processing_time_ms": 0,
        "error": None,
        "answers": answer_results,
    }


def weighted_mean(items: list[dict[str, Any]], key: str) -> float:
    weighted_sum = 0.0
    total_weight = 0
    for item in items:
        value = item.get(key)
        if value is None:
            continue
        weight = max(int(item.get("duration_ms") or 0), 1)
        weighted_sum += float(value) * weight
        total_weight += weight
    return round(weighted_sum / total_weight, 6) if total_weight > 0 else 0.0


def weighted_emotion_probs(features_list: list[dict[str, Any]], emotion_list: list[dict[str, Any]]) -> dict[str, float]:
    if not emotion_list:
        return {}

    all_keys = sorted({key for emotion in emotion_list for key in emotion.keys()})
    result: dict[str, float] = {}
    for key in all_keys:
        weighted_sum = 0.0
        total_weight = 0
        for features, emotion in zip(features_list, emotion_list, strict=False):
            value = emotion.get(key)
            if value is None:
                continue
            weight = max(int(features.get("duration_ms") or 0), 1)
            weighted_sum += float(value) * weight
            total_weight += weight
        result[key] = round(weighted_sum / total_weight, 4) if total_weight > 0 else 0.0
    return result


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
