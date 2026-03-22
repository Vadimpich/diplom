from datetime import datetime, timezone

from fastapi import FastAPI, HTTPException

from app.schemas import (
    BaselineCalculationRequest,
    BaselineCalculationResponse,
    DeviationMetricScore,
    DeviationSection,
    HealthResponse,
    NextBaselineSnapshot,
    UpdateEligibility,
)


REQUIRED_METRIC_KEYS = {"overall_proxy_index", "speech_stability_proxy"}

app = FastAPI(title="ml-baseline")


def utc_now() -> datetime:
    return datetime.now(timezone.utc)


def metric_map(metrics: list) -> dict[str, float]:
    return {metric.key: metric.value for metric in metrics}


def build_deviation_section(
    current_metrics: dict[str, float],
    reference_metrics: dict[str, float],
) -> DeviationSection:
    metric_scores: dict[str, DeviationMetricScore] = {}
    total = 0.0

    for key, current_value in current_metrics.items():
        reference_value = reference_metrics.get(key, 0.0)
        delta = round(current_value - reference_value, 4)
        robust_z = round(delta / 0.1, 4)
        band = "low"
        if abs(robust_z) >= 3:
            band = "high"
        elif abs(robust_z) >= 2:
            band = "moderate"
        elif abs(robust_z) >= 1:
            band = "mild"

        metric_scores[key] = DeviationMetricScore(delta=delta, robust_z=robust_z, band=band)
        total += abs(robust_z)

    score = round(total / max(len(metric_scores), 1), 4)
    section_band = "low"
    if score >= 3:
        section_band = "high"
    elif score >= 2:
        section_band = "moderate"
    elif score >= 1:
        section_band = "mild"

    return DeviationSection(score=score, band=section_band, metric_scores=metric_scores)


def calculate_baseline(payload: BaselineCalculationRequest) -> BaselineCalculationResponse:
    current_metrics = metric_map(payload.metrics)
    if missing_keys := sorted(REQUIRED_METRIC_KEYS - current_metrics.keys()):
        raise HTTPException(
            status_code=422,
            detail={
                "code": "missing_required_metrics",
                "missing_metric_keys": missing_keys,
            },
        )

    history_vectors = payload.history.metric_vectors
    if history_vectors:
        last_vector = metric_map(history_vectors[-1].metrics)
        mean_reference = {
            key: round(
                sum(metric_map(vector.metrics).get(key, 0.0) for vector in history_vectors) / len(history_vectors),
                4,
            )
            for key in current_metrics
        }
    else:
        last_vector = {key: 0.0 for key in current_metrics}
        mean_reference = {key: 0.0 for key in current_metrics}

    general_deviation = build_deviation_section(current_metrics, last_vector)
    personal_deviation = build_deviation_section(current_metrics, mean_reference)

    next_exam_count = len(history_vectors) + 1
    next_baseline = NextBaselineSnapshot(
        exam_count=next_exam_count,
        centers={key: round((mean_reference.get(key, 0.0) + value) / 2, 4) for key, value in current_metrics.items()},
        scales={key: round(abs(value - mean_reference.get(key, 0.0)), 4) for key, value in current_metrics.items()},
        refreshed_at=utc_now(),
    )

    return BaselineCalculationResponse(
        schema_version=payload.schema_version,
        algorithm_version=payload.algorithm_version,
        refreshed_at=utc_now(),
        general_deviation=general_deviation,
        personal_deviation=personal_deviation,
        update_eligibility=UpdateEligibility(
            eligible=True,
            reason="accepted",
            baseline_exam_count_after_update=next_exam_count,
        ),
        next_baseline=next_baseline,
    )


@app.get("/health", response_model=HealthResponse)
def health() -> HealthResponse:
    return HealthResponse(status="ok", service="ml-baseline")


@app.post("/baseline/calculate", response_model=BaselineCalculationResponse)
def calculate(payload: BaselineCalculationRequest) -> BaselineCalculationResponse:
    return calculate_baseline(payload)
