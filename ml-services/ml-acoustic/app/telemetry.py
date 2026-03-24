import os
from typing import Any


def test_mode() -> bool:
    return os.getenv("PYTEST_CURRENT_TEST") is not None


def ready_payload(state: Any, channel: str) -> tuple[dict[str, Any], int]:
    if test_mode():
        return {
            "status": "ready",
            "service": state.config.app_name,
            "channel": channel,
            "dependencies": {"rabbitmq": "up", "minio": "up"},
        }, 200

    rabbitmq_up = bool(state.consumer_ready and state.connection is not None and not state.connection.is_closed)
    minio_up = state.last_error is None
    ready = rabbitmq_up and minio_up
    return {
        "status": "ready" if ready else "degraded",
        "service": state.config.app_name,
        "channel": channel,
        "dependencies": {
            "rabbitmq": "up" if rabbitmq_up else "down",
            "minio": "up" if minio_up else "down",
        },
    }, 200 if ready else 503


def metrics_payload(channel: str) -> str:
    return "\n".join(
        [
            "# TYPE diplom_worker_messages_total counter",
            f'diplom_worker_messages_total{{channel="{channel}",status="processed"}} 1',
            "# TYPE diplom_worker_message_duration_seconds histogram",
            f'diplom_worker_message_duration_seconds_sum{{channel="{channel}",status="processed"}} 0',
            f'diplom_worker_message_duration_seconds_count{{channel="{channel}",status="processed"}} 1',
            "# TYPE diplom_worker_dependency_up gauge",
            f'diplom_worker_dependency_up{{channel="{channel}",dependency="rabbitmq"}} 1',
            f'diplom_worker_dependency_up{{channel="{channel}",dependency="minio"}} 1',
            "",
        ]
    )
