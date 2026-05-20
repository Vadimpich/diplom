from __future__ import annotations

import math
import os
from collections.abc import Iterable
from datetime import datetime

import numpy as np


MIN_PERSONAL_BASELINE_EXAMS = int(os.getenv("BASELINE_MIN_PERSONAL_EXAMS", "5"))
BASELINE_EWMA_ALPHA = float(os.getenv("BASELINE_EWMA_ALPHA", "0.2"))
DATA_RELIABILITY_THRESHOLD = float(os.getenv("BASELINE_DATA_RELIABILITY_THRESHOLD", "0.7"))
MIN_STD = float(os.getenv("BASELINE_MIN_STD", "0.05"))

GENERAL_REFERENCE = {
    "overall_deviation_index": {"mean": 0.26, "std": 0.08, "method": "general_reference"},
    "text_risk_signal": {"mean": 0.24, "std": 0.09, "method": "general_reference"},
    "acoustic_stress_signal": {"mean": 0.27, "std": 0.09, "method": "general_reference"},
    "paralinguistic_behavior_signal": {"mean": 0.23, "std": 0.08, "method": "general_reference"},
    "speech_stability_score": {"mean": 0.72, "std": 0.08, "method": "general_reference"},
}

CRITICAL_QUALITY_FLAGS = {
    "stt_failed",
    "empty_transcript",
    "feature_extraction_failed",
    "vad_failed",
    "no_speech_detected",
}


def general_metric_reference(key: str) -> dict[str, float | str]:
    return GENERAL_REFERENCE.get(key, {"mean": 0.5, "std": 0.1, "method": "general_reference"})


def robust_metric_summary(values: Iterable[float]) -> dict[str, float | list[float]]:
    sample = np.array(list(values), dtype=float)
    if sample.size == 0:
        return {"center": 0.0, "scale": MIN_STD, "robust_z_scores": []}

    center = float(np.median(sample))
    mad = float(np.median(np.abs(sample - center)))
    scale = max(mad, MIN_STD)
    robust_z_scores = (0.6745 * (sample - center)) / scale

    return {
        "center": round(center, 4),
        "scale": round(scale, 4),
        "robust_z_scores": [round(float(score), 4) for score in robust_z_scores],
    }


def z_distance(current_value: float, center: float, scale: float) -> tuple[float, float]:
    safe_scale = scale if scale > 0 else MIN_STD
    delta = round(current_value - center, 4)
    z_score = round(delta / safe_scale, 4)
    return delta, z_score


def summarize_history(history: list[dict[str, float]]) -> dict[str, dict[str, float | list[float]]]:
    metric_keys = sorted({key for vector in history for key in vector})
    return {
        key: robust_metric_summary(vector[key] for vector in history if key in vector)
        for key in metric_keys
    }


def has_critical_quality_flags(flags: list[str]) -> bool:
    return any(flag in CRITICAL_QUALITY_FLAGS for flag in flags)


def determine_deviation_level(z_score: float) -> str:
    absolute = abs(z_score)
    if absolute >= 2.5:
        return "high"
    if absolute >= 1.75:
        return "moderate"
    if absolute >= 1.0:
        return "mild"
    return "none"


def section_band(levels: Iterable[str]) -> str:
    ranked = {"none": 0, "mild": 1, "moderate": 2, "high": 3}
    best = max((ranked.get(level, 0) for level in levels), default=0)
    reverse = {value: key for key, value in ranked.items()}
    return reverse[best]


def baseline_state_from_history(
    history: list[dict[str, float]],
    refreshed_at: datetime,
    *,
    method: str,
) -> dict[str, dict[str, object]]:
    summary = summarize_history(history)
    result: dict[str, dict[str, object]] = {}
    for key, value in summary.items():
        result[key] = {
            "baseline_mean": float(value["center"]),
            "baseline_std": float(value["scale"]),
            "sample_count": len(history),
            "last_updated_at": refreshed_at,
            "method": method,
        }
    return result


def evaluate_update_eligibility(
    *,
    history_count: int,
    all_channels_done: bool,
    critical_quality_flags: list[str],
    data_reliability: float,
    overall_band: str,
) -> dict[str, int | bool | str | float]:
    if not all_channels_done:
        return {
            "eligible": False,
            "reason": "channels_incomplete",
            "baseline_exam_count_after_update": history_count,
            "data_reliability": round(data_reliability, 4),
        }
    if has_critical_quality_flags(critical_quality_flags):
        return {
            "eligible": False,
            "reason": "critical_quality_flags",
            "baseline_exam_count_after_update": history_count,
            "data_reliability": round(data_reliability, 4),
        }
    if data_reliability < DATA_RELIABILITY_THRESHOLD:
        return {
            "eligible": False,
            "reason": "low_data_reliability",
            "baseline_exam_count_after_update": history_count,
            "data_reliability": round(data_reliability, 4),
        }
    if overall_band == "high":
        return {
            "eligible": False,
            "reason": "high_deviation",
            "baseline_exam_count_after_update": history_count,
            "data_reliability": round(data_reliability, 4),
        }
    return {
        "eligible": True,
        "reason": "accepted",
        "baseline_exam_count_after_update": history_count + 1,
        "data_reliability": round(data_reliability, 4),
    }


def build_metric_scores(
    current_metrics: dict[str, float],
    baseline_metrics: dict[str, dict[str, object]],
    *,
    baseline_source: str,
    baseline_available: bool,
) -> tuple[dict[str, dict[str, object]], list[str]]:
    scores: dict[str, dict[str, object]] = {}
    significant: list[str] = []
    for key, current_value in current_metrics.items():
        if baseline_available and key in baseline_metrics:
            baseline = baseline_metrics[key]
            mean = float(baseline["baseline_mean"])
            std = max(float(baseline["baseline_std"]), MIN_STD)
            sample_count = int(baseline["sample_count"])
            method = str(baseline["method"])
        else:
            reference = general_metric_reference(key)
            mean = float(reference["mean"])
            std = max(float(reference["std"]), MIN_STD)
            sample_count = 0
            method = str(reference["method"])
        delta, z_score = z_distance(current_value, mean, std)
        deviation_level = determine_deviation_level(z_score)
        if deviation_level in {"moderate", "high"}:
            significant.append(key)
        scores[key] = {
            "baseline_available": baseline_available,
            "baseline_source": baseline_source,
            "baseline_mean": round(mean, 4),
            "baseline_std": round(std, 4),
            "sample_count": sample_count,
            "method": method,
            "delta": delta,
            "z_score": z_score,
            "deviation_level": deviation_level,
        }
    return scores, significant


def build_deviation_section(
    current_metrics: dict[str, float],
    baseline_metrics: dict[str, dict[str, object]],
    *,
    baseline_source: str,
    baseline_available: bool,
) -> dict[str, object]:
    metric_scores, significant = build_metric_scores(
        current_metrics,
        baseline_metrics,
        baseline_source=baseline_source,
        baseline_available=baseline_available,
    )
    mean_abs_z = round(
        sum(abs(float(item["z_score"])) for item in metric_scores.values()) / max(len(metric_scores), 1),
        4,
    )
    levels = [str(item["deviation_level"]) for item in metric_scores.values()]
    return {
        "score": mean_abs_z,
        "band": section_band(levels),
        "baseline_available": baseline_available,
        "baseline_source": baseline_source,
        "significant_deviations": significant,
        "metric_scores": metric_scores,
    }


def ewma_update_metric(
    previous_state: dict[str, object] | None,
    current_value: float,
    refreshed_at: datetime,
) -> dict[str, object]:
    if previous_state is None:
        return {
            "baseline_mean": round(current_value, 4),
            "baseline_std": MIN_STD,
            "sample_count": 1,
            "last_updated_at": refreshed_at,
            "method": "ewma_bootstrap",
        }

    previous_mean = float(previous_state["baseline_mean"])
    previous_std = max(float(previous_state["baseline_std"]), MIN_STD)
    sample_count = int(previous_state["sample_count"]) + 1
    new_mean = BASELINE_EWMA_ALPHA * current_value + (1 - BASELINE_EWMA_ALPHA) * previous_mean
    previous_variance = previous_std**2
    new_variance = (1 - BASELINE_EWMA_ALPHA) * (
        previous_variance + BASELINE_EWMA_ALPHA * (current_value - previous_mean) ** 2
    )
    return {
        "baseline_mean": round(new_mean, 4),
        "baseline_std": round(max(math.sqrt(max(new_variance, 0.0)), MIN_STD), 4),
        "sample_count": sample_count,
        "last_updated_at": refreshed_at,
        "method": "ewma",
    }


def build_next_baseline(
    *,
    current: dict[str, float],
    history: list[dict[str, float]],
    history_count: int,
    existing_baseline_metrics: dict[str, dict[str, object]],
    refreshed_at: datetime,
    update_eligibility: dict[str, int | bool | str | float],
) -> dict[str, object]:
    if not bool(update_eligibility["eligible"]):
        current_count = max(
            [int(item.get("sample_count", 0)) for item in existing_baseline_metrics.values()] or [history_count]
        )
        return {
            "exam_count": current_count,
            "baseline_available": current_count >= MIN_PERSONAL_BASELINE_EXAMS,
            "metrics": existing_baseline_metrics,
            "refreshed_at": refreshed_at,
        }

    metrics: dict[str, dict[str, object]] = {}
    if existing_baseline_metrics:
        for key, current_value in current.items():
            metrics[key] = ewma_update_metric(existing_baseline_metrics.get(key), current_value, refreshed_at)
    else:
        updated_history = [*history, current]
        metrics = baseline_state_from_history(updated_history, refreshed_at, method="ewma_bootstrap")
        effective_sample_count = history_count + 1
        for item in metrics.values():
            item["sample_count"] = max(int(item["sample_count"]), effective_sample_count)

    exam_count = max([int(item["sample_count"]) for item in metrics.values()] or [0])
    for key, current_value in current.items():
        if key not in metrics:
            metrics[key] = ewma_update_metric(None, current_value, refreshed_at)
    return {
        "exam_count": exam_count,
        "baseline_available": exam_count >= MIN_PERSONAL_BASELINE_EXAMS,
        "metrics": metrics,
        "refreshed_at": refreshed_at,
    }
