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
    assert "diplom_baseline_requests_total" in metrics.text
