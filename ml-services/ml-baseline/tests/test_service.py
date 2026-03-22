from fastapi.testclient import TestClient

from app.main import app


def build_payload() -> dict:
    return {
        "schema_version": 1,
        "algorithm_version": "baseline-v1",
        "specialist_id": 55,
        "examination_id": 101,
        "generated_at": "2026-03-22T10:02:08Z",
        "metrics": [
            {"key": "overall_proxy_index", "value": 0.49},
            {"key": "speech_stability_proxy", "value": 0.405},
        ],
        "history": {
            "baseline_exam_count": 4,
            "metric_vectors": [
                {
                    "examination_id": 91,
                    "generated_at": "2026-03-20T10:02:08Z",
                    "metrics": [
                        {"key": "overall_proxy_index", "value": 0.46},
                        {"key": "speech_stability_proxy", "value": 0.39},
                    ],
                },
                {
                    "examination_id": 92,
                    "generated_at": "2026-03-21T10:02:08Z",
                    "metrics": [
                        {"key": "overall_proxy_index", "value": 0.48},
                        {"key": "speech_stability_proxy", "value": 0.40},
                    ],
                },
            ],
        },
        "general_reference_population_version": "general-v1",
    }


def test_returns_general_and_personal_deviation() -> None:
    client = TestClient(app)

    response = client.post("/baseline/calculate", json=build_payload())

    assert response.status_code == 200
    body = response.json()
    assert body["schema_version"] == 1
    assert body["algorithm_version"] == "baseline-v1"
    assert body["general_deviation"]["score"] >= 0
    assert body["general_deviation"]["metric_scores"]["overall_proxy_index"]["band"]
    assert body["personal_deviation"]["score"] >= 0
    assert body["personal_deviation"]["metric_scores"]["speech_stability_proxy"]["delta"] != 0


def test_response_includes_metadata_and_update_eligibility() -> None:
    client = TestClient(app)

    response = client.post("/baseline/calculate", json=build_payload())

    assert response.status_code == 200
    body = response.json()
    assert body["refreshed_at"]
    assert body["update_eligibility"] == {
        "eligible": True,
        "reason": "accepted",
        "baseline_exam_count_after_update": 3,
    }
    assert body["next_baseline"]["exam_count"] == 3
    assert body["next_baseline"]["centers"]["overall_proxy_index"] > 0
    assert body["next_baseline"]["scales"]["speech_stability_proxy"] >= 0


def test_health_endpoint_reports_ready() -> None:
    client = TestClient(app)

    response = client.get("/health")

    assert response.status_code == 200
    assert response.json() == {"status": "ok", "service": "ml-baseline"}
