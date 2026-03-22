from app.algorithms import (
    apply_baseline_update,
    evaluate_update_eligibility,
    robust_metric_summary,
)


def test_robust_metric_summary_uses_median_and_mad() -> None:
    summary = robust_metric_summary([0.2, 0.21, 0.22, 0.9])

    assert summary["center"] == 0.215
    assert summary["scale"] > 0
    assert round(summary["robust_z_scores"][-1], 2) > 45


def test_outlier_freezes_baseline_update() -> None:
    history = [
        {"overall_proxy_index": 0.40, "speech_stability_proxy": 0.38},
        {"overall_proxy_index": 0.43, "speech_stability_proxy": 0.39},
        {"overall_proxy_index": 0.41, "speech_stability_proxy": 0.37},
        {"overall_proxy_index": 0.42, "speech_stability_proxy": 0.40},
    ]
    current = {"overall_proxy_index": 0.95, "speech_stability_proxy": 0.15}

    result = evaluate_update_eligibility(history=history, current=current)

    assert result["eligible"] is False
    assert result["reason"] == "outlier_detected"
    assert result["baseline_exam_count_after_update"] == 4


def test_bounded_history_update_keeps_recent_window() -> None:
    history = [
        {"overall_proxy_index": 0.33, "speech_stability_proxy": 0.42},
        {"overall_proxy_index": 0.34, "speech_stability_proxy": 0.44},
        {"overall_proxy_index": 0.35, "speech_stability_proxy": 0.43},
        {"overall_proxy_index": 0.36, "speech_stability_proxy": 0.45},
        {"overall_proxy_index": 0.37, "speech_stability_proxy": 0.46},
    ]
    current = {"overall_proxy_index": 0.38, "speech_stability_proxy": 0.47}

    updated = apply_baseline_update(history=history, current=current, max_history=5)

    assert len(updated["history"]) == 5
    assert updated["history"][0]["overall_proxy_index"] == 0.34
    assert updated["history"][-1]["speech_stability_proxy"] == 0.47
    assert updated["baseline_exam_count_after_update"] == 5
