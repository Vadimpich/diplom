from importlib import import_module


def test_returns_general_and_personal_deviation():
    service = import_module("app.service")

    payload = {
        "schema_version": 1,
        "algorithm_version": "baseline-v1",
        "specialist_id": 55,
        "examination_id": 101,
        "generated_at": "2026-03-22T10:02:08Z",
        "metrics": [
            {"key": "overall_proxy_index", "value": 0.58},
            {"key": "speech_stability_proxy", "value": 0.41},
        ],
        "history": {
            "baseline_exam_count": 4,
            "metric_vectors": [
                {
                    "examination_id": 91,
                    "generated_at": "2026-03-20T10:02:08Z",
                    "metrics": [{"key": "overall_proxy_index", "value": 0.46}],
                }
            ],
        },
        "general_reference_population_version": "general-v1",
    }

    response = service.calculate_baseline(payload)

    assert response["general_deviation"]["score"] >= 0
    assert response["personal_deviation"]["score"] >= 0
    assert response["algorithm_version"] == "baseline-v1"
    assert response["refreshed_at"]
    assert "update_eligibility" in response
