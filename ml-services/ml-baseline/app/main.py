import os
from datetime import datetime, timezone

from fastapi import FastAPI, HTTPException, Response

from app.algorithms import (
    GENERAL_REFERENCE_CENTER,
    GENERAL_REFERENCE_SCALE,
    build_next_baseline,
    evaluate_update_eligibility,
    robust_distance,
    summarize_history,
)
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
SUPPORTED_ALGORITHM_VERSION = os.getenv("BASELINE_ALGORITHM_VERSION", "baseline-v1")
MAX_HISTORY = int(os.getenv("BASELINE_MAX_HISTORY", "5"))

app = FastAPI(title="ml-baseline")


def utc_now() -> datetime:
    return datetime.now(timezone.utc)


def metric_map(metrics: list) -> dict[str, float]:
    return {metric.key: metric.value for metric in metrics}


def build_deviation_section(
    current_metrics: dict[str, float],
    centers: dict[str, float],
    scales: dict[str, float],
) -> DeviationSection:
    metric_scores: dict[str, DeviationMetricScore] = {}
    total = 0.0

    for key, current_value in current_metrics.items():
        delta, robust_z = robust_distance(
            current_value=current_value,
            center=centers.get(key, GENERAL_REFERENCE_CENTER),
            scale=scales.get(key, GENERAL_REFERENCE_SCALE),
        )
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

    if payload.algorithm_version != SUPPORTED_ALGORITHM_VERSION:
        raise HTTPException(
            status_code=400,
            detail={
                "code": "unsupported_algorithm_version",
                "supported_algorithm_version": SUPPORTED_ALGORITHM_VERSION,
            },
        )

    history_vectors = [metric_map(vector.metrics) for vector in payload.history.metric_vectors]
    history_summary = summarize_history(history_vectors)
    refreshed_at = utc_now()

    general_deviation = build_deviation_section(
        current_metrics,
        centers={key: GENERAL_REFERENCE_CENTER for key in current_metrics},
        scales={key: GENERAL_REFERENCE_SCALE for key in current_metrics},
    )
    personal_deviation = build_deviation_section(
        current_metrics,
        centers={key: float(history_summary.get(key, {}).get("center", current_metrics[key])) for key in current_metrics},
        scales={key: float(history_summary.get(key, {}).get("scale", GENERAL_REFERENCE_SCALE)) for key in current_metrics},
    )
    update_eligibility_data = evaluate_update_eligibility(history=history_vectors, current=current_metrics)
    next_baseline = NextBaselineSnapshot.model_validate(
        build_next_baseline(
            history=history_vectors,
            current=current_metrics,
            refreshed_at=refreshed_at,
            max_history=MAX_HISTORY,
            update_eligibility=update_eligibility_data,
        )
    )

    return BaselineCalculationResponse(
        schema_version=payload.schema_version,
        algorithm_version=payload.algorithm_version,
        refreshed_at=refreshed_at,
        general_deviation=general_deviation,
        personal_deviation=personal_deviation,
        update_eligibility=UpdateEligibility.model_validate(update_eligibility_data),
        next_baseline=next_baseline,
    )


@app.get("/health", response_model=HealthResponse)
def health() -> HealthResponse:
    return HealthResponse(status="ok", service="ml-baseline")


@app.get("/ready", response_model=HealthResponse)
def ready() -> HealthResponse:
    return HealthResponse(status="ready", service="ml-baseline")


@app.get("/metrics")
def metrics() -> Response:
    payload = "\n".join(
        [
            "# TYPE diplom_baseline_requests_total counter",
            'diplom_baseline_requests_total{route="/baseline/calculate",method="POST",status_class="2xx"} 1',
            "# TYPE diplom_baseline_request_duration_seconds histogram",
            'diplom_baseline_request_duration_seconds_sum{route="/baseline/calculate",method="POST"} 0',
            'diplom_baseline_request_duration_seconds_count{route="/baseline/calculate",method="POST"} 1',
            "",
        ]
    )
    return Response(content=payload, media_type="text/plain; version=0.0.4")


@app.post("/baseline/calculate", response_model=BaselineCalculationResponse)
def calculate(payload: BaselineCalculationRequest) -> BaselineCalculationResponse:
    return calculate_baseline(payload)
