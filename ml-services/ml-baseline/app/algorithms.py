from collections.abc import Iterable
from datetime import datetime

import numpy as np


GENERAL_REFERENCE_CENTER = 0.5
GENERAL_REFERENCE_SCALE = 0.1
OUTLIER_THRESHOLD = 3.5


def robust_metric_summary(values: Iterable[float]) -> dict[str, float | list[float]]:
    sample = np.array(list(values), dtype=float)
    if sample.size == 0:
        return {"center": 0.0, "scale": 0.0, "robust_z_scores": []}

    center = float(np.median(sample))
    mad = float(np.median(np.abs(sample - center)))
    scale = mad if mad > 0 else 1e-6
    robust_z_scores = (0.6745 * (sample - center)) / scale

    return {
        "center": round(center, 4),
        "scale": round(scale, 4),
        "robust_z_scores": [round(float(score), 4) for score in robust_z_scores],
    }


def robust_distance(current_value: float, center: float, scale: float) -> tuple[float, float]:
    safe_scale = scale if scale > 0 else 1e-6
    delta = round(current_value - center, 4)
    robust_z = round((0.6745 * delta) / safe_scale, 4)
    return delta, robust_z


def summarize_history(history: list[dict[str, float]]) -> dict[str, dict[str, float | list[float]]]:
    metric_keys = sorted({key for vector in history for key in vector})
    return {
        key: robust_metric_summary(vector[key] for vector in history if key in vector)
        for key in metric_keys
    }


def evaluate_update_eligibility(history: list[dict[str, float]], current: dict[str, float]) -> dict[str, int | bool | str]:
    if not history:
        return {
            "eligible": True,
            "reason": "accepted",
            "baseline_exam_count_after_update": 1,
        }

    history_summary = summarize_history(history)
    for key, value in current.items():
        metric_summary = history_summary.get(key)
        if metric_summary is None:
            continue
        _, robust_z = robust_distance(
            current_value=value,
            center=float(metric_summary["center"]),
            scale=float(metric_summary["scale"]),
        )
        if abs(robust_z) > OUTLIER_THRESHOLD:
            return {
                "eligible": False,
                "reason": "outlier_detected",
                "baseline_exam_count_after_update": len(history),
            }

    return {
        "eligible": True,
        "reason": "accepted",
        "baseline_exam_count_after_update": len(history) + 1,
    }


def apply_baseline_update(
    history: list[dict[str, float]],
    current: dict[str, float],
    max_history: int,
) -> dict[str, object]:
    updated_history = [*history, current]
    if len(updated_history) > max_history:
        updated_history = updated_history[-max_history:]

    return {
        "history": updated_history,
        "baseline_exam_count_after_update": len(updated_history),
        "summary": summarize_history(updated_history),
    }


def build_next_baseline(
    history: list[dict[str, float]],
    current: dict[str, float],
    refreshed_at: datetime,
    max_history: int,
    update_eligibility: dict[str, int | bool | str],
) -> dict[str, object]:
    if bool(update_eligibility["eligible"]):
        updated = apply_baseline_update(history=history, current=current, max_history=max_history)
        summary = updated["summary"]
        exam_count = int(updated["baseline_exam_count_after_update"])
    else:
        summary = summarize_history(history)
        exam_count = len(history)

    return {
        "exam_count": exam_count,
        "centers": {key: float(value["center"]) for key, value in summary.items()},
        "scales": {key: float(value["scale"]) for key, value in summary.items()},
        "refreshed_at": refreshed_at,
    }
