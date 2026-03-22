from importlib import import_module


def test_outlier_freezes_baseline_update():
    algorithms = import_module("app.algorithms")

    history = [
        {"overall_proxy_index": 0.40},
        {"overall_proxy_index": 0.43},
        {"overall_proxy_index": 0.41},
        {"overall_proxy_index": 0.42},
    ]
    current = {"overall_proxy_index": 0.95}

    result = algorithms.evaluate_update_eligibility(history=history, current=current)

    assert result["eligible"] is False
    assert result["reason"] == "outlier_detected"
    assert result["baseline_exam_count_after_update"] == 4
