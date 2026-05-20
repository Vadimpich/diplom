from datetime import datetime, timezone

from app.algorithms import build_deviation_section, build_next_baseline, evaluate_update_eligibility, robust_metric_summary


def test_robust_metric_summary_uses_median_and_mad() -> None:
    summary = robust_metric_summary([0.2, 0.21, 0.22, 0.9])

    assert summary["center"] == 0.215
    assert summary["scale"] > 0
    assert round(summary["robust_z_scores"][-1], 2) > 9


def test_personal_baseline_becomes_available_after_min_exam_count() -> None:
    refreshed_at = datetime(2026, 5, 18, tzinfo=timezone.utc)
    current = {"overall_deviation_index": 0.38, "speech_stability_score": 0.66}
    history = [
        {"overall_deviation_index": 0.30, "speech_stability_score": 0.71},
        {"overall_deviation_index": 0.31, "speech_stability_score": 0.70},
        {"overall_deviation_index": 0.29, "speech_stability_score": 0.73},
        {"overall_deviation_index": 0.32, "speech_stability_score": 0.69},
    ]
    eligibility = {
        "eligible": True,
        "reason": "accepted",
        "baseline_exam_count_after_update": 5,
        "data_reliability": 0.91,
    }

    next_baseline = build_next_baseline(
        current=current,
        history=history,
        history_count=4,
        existing_baseline_metrics={},
        refreshed_at=refreshed_at,
        update_eligibility=eligibility,
    )

    assert next_baseline["exam_count"] == 5
    assert next_baseline["baseline_available"] is True
    assert next_baseline["metrics"]["overall_deviation_index"]["sample_count"] == 5


def test_high_deviation_blocks_baseline_update() -> None:
    result = evaluate_update_eligibility(
        history_count=4,
        all_channels_done=True,
        critical_quality_flags=[],
        data_reliability=0.92,
        overall_band="high",
    )

    assert result["eligible"] is False
    assert result["reason"] == "high_deviation"
    assert result["baseline_exam_count_after_update"] == 4


def test_z_score_and_deviation_level_are_calculated() -> None:
    current = {"overall_deviation_index": 0.49, "speech_stability_score": 0.52}
    baseline = {
        "overall_deviation_index": {
            "baseline_mean": 0.29,
            "baseline_std": 0.08,
            "sample_count": 6,
            "last_updated_at": None,
            "method": "ewma",
        },
        "speech_stability_score": {
            "baseline_mean": 0.70,
            "baseline_std": 0.08,
            "sample_count": 6,
            "last_updated_at": None,
            "method": "ewma",
        },
    }

    section = build_deviation_section(
        current,
        baseline,
        baseline_source="personal",
        baseline_available=True,
    )

    assert section["metric_scores"]["overall_deviation_index"]["z_score"] == 2.5
    assert section["metric_scores"]["overall_deviation_index"]["deviation_level"] == "high"
    assert section["metric_scores"]["speech_stability_score"]["deviation_level"] == "moderate"
    assert "overall_deviation_index" in section["significant_deviations"]


def test_fallback_to_general_baseline_when_personal_missing() -> None:
    current = {"overall_deviation_index": 0.32, "speech_stability_score": 0.68}

    section = build_deviation_section(
        current,
        {},
        baseline_source="general",
        baseline_available=False,
    )

    assert section["baseline_available"] is False
    assert section["baseline_source"] == "general"
    assert section["metric_scores"]["overall_deviation_index"]["sample_count"] == 0
    assert section["metric_scores"]["overall_deviation_index"]["method"] == "general_reference"
