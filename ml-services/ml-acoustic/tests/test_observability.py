from fastapi.testclient import TestClient

from app.main import app


def test_health_ready_and_metrics_contract() -> None:
    client = TestClient(app)

    health = client.get("/health")
    ready = client.get("/ready")
    metrics = client.get("/metrics")

    assert health.status_code == 200
    assert ready.status_code == 200
    assert metrics.status_code == 200
    assert "diplom_worker_messages_total" in metrics.text


def test_result_envelope_keeps_trace_context() -> None:
    from app.main import ProcessingCommandEnvelope, WorkerState, load_config

    state = WorkerState(load_config())
    command = ProcessingCommandEnvelope.model_validate(
        {
            "message_version": 1,
            "message_id": "msg-1",
            "correlation_id": "exam-100-acoustic-v1",
            "request_id": "req-100",
            "traceparent": "00-8ec8c1b6409f4a6cb80cfcb4f74aa98c-5d7c1f97db7840b3-01",
            "tracestate": "tenant=diplom",
            "examination_id": 100,
            "specialist_id": 55,
            "channel": "acoustic",
            "attempt": 1,
            "max_attempts": 3,
            "requested_at": "2026-03-23T18:10:12Z",
            "answers": [
                {
                    "answer_id": 1,
                    "question_id": 2,
                    "audio_s3_bucket": "bucket",
                    "audio_s3_key": "key",
                    "answer_text": "text",
                }
            ],
        }
    )

    result = state.build_result(command, "succeeded", payload={"summary": "ok"})

    assert result["request_id"] == "req-100"
    assert result["traceparent"]
    assert result["tracestate"] == "tenant=diplom"
