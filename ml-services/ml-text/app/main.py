import asyncio
import hashlib
import json
import logging
import os
import signal
from contextlib import asynccontextmanager
from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Any
from uuid import uuid4

import aio_pika
from aio_pika import ExchangeType, IncomingMessage, Message, RobustChannel, RobustConnection
from fastapi import FastAPI
from minio import Minio
from minio.error import S3Error
from pydantic import BaseModel, ConfigDict, Field, ValidationError


CHANNEL = "text"
MODEL_VERSION = "text-stub-0.1.0"
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
    user = os.getenv("RABBITMQ_USER", "dimplom")
    password = os.getenv("RABBITMQ_PASSWORD", "dimplom_dev_password")
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
                        "sha256": hashlib.sha256(content).hexdigest(),
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
    total_bytes = sum(item["bytes"] for item in fetched_objects)
    total_chars = sum(len(item["answer_text"].strip()) for item in fetched_objects)
    combined_digest = hashlib.sha256("".join(item["sha256"] for item in fetched_objects).encode("utf-8")).hexdigest()
    return {
        "summary": "text worker stub result",
        "answers_processed": len(fetched_objects),
        "audio_total_bytes": total_bytes,
        "text_total_characters": total_chars,
        "text_non_empty_answers": sum(1 for item in fetched_objects if item["answer_text"].strip()),
        "artifact_digest": combined_digest[:16],
        "requested_at": command.requested_at.isoformat(),
    }


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


def _handle_signal(_: int, __: Any) -> None:
    state.stop_event.set()


signal.signal(signal.SIGTERM, _handle_signal)
signal.signal(signal.SIGINT, _handle_signal)
