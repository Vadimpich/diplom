import logging
import os
from datetime import datetime, timezone

from fastapi import FastAPI, HTTPException, Response

from app.algorithms import (
    build_next_baseline,
    evaluate_update_eligibility,
    build_deviation_section,
    general_metric_reference,
)
from app.schemas import (
    BaselineCalculationRequest,
    BaselineCalculationResponse,
    DeviationSection,
    HealthResponse,
    NextBaselineSnapshot,
    UpdateEligibility,
)


REQUIRED_METRIC_KEYS = {"overall_deviation_index", "speech_stability_score"}
SUPPORTED_ALGORITHM_VERSION = os.getenv("BASELINE_ALGORITHM_VERSION", "baseline-v1")

app = FastAPI(title="ml-baseline")
logger = logging.getLogger("ml-baseline")


def utc_now() -> datetime:
    return datetime.now(timezone.utc)


def metric_map(metrics: list) -> dict[str, float]:
    return {metric.key: metric.value for metric in metrics}


def calculate_baseline(payload: BaselineCalculationRequest) -> BaselineCalculationResponse:
    current_metrics = metric_map(payload.metrics)
    logger.info(
        "baseline_calculation_started examination_id=%s specialist_id=%s algorithm_version=%s metrics=%s history_count=%s personal_baseline_available=%s data_reliability=%.4f overall_band=%s",
        payload.examination_id,
        payload.specialist_id,
        payload.algorithm_version,
        len(current_metrics),
        payload.history.baseline_exam_count,
        payload.existing_baseline.baseline_available,
        payload.context.data_reliability,
        payload.context.overall_band,
    )
    if missing_keys := sorted(REQUIRED_METRIC_KEYS - current_metrics.keys()):
        logger.warning(
            "baseline_calculation_rejected examination_id=%s specialist_id=%s reason=missing_required_metrics missing_metric_keys=%s",
            payload.examination_id,
            payload.specialist_id,
            ",".join(missing_keys),
        )
        raise HTTPException(
            status_code=422,
            detail={
                "code": "missing_required_metrics",
                "missing_metric_keys": missing_keys,
            },
        )

    if payload.algorithm_version != SUPPORTED_ALGORITHM_VERSION:
        logger.warning(
            "baseline_calculation_rejected examination_id=%s specialist_id=%s reason=unsupported_algorithm_version requested=%s supported=%s",
            payload.examination_id,
            payload.specialist_id,
            payload.algorithm_version,
            SUPPORTED_ALGORITHM_VERSION,
        )
        raise HTTPException(
            status_code=400,
            detail={
                "code": "unsupported_algorithm_version",
                "supported_algorithm_version": SUPPORTED_ALGORITHM_VERSION,
            },
        )

    history_vectors = [metric_map(vector.metrics) for vector in payload.history.metric_vectors]
    refreshed_at = utc_now()
    existing_baseline_metrics = {
        key: value.model_dump()
        for key, value in payload.existing_baseline.metrics.items()
    }

    general_reference_metrics = {
        key: {
            "baseline_mean": float(general_metric_reference(key)["mean"]),
            "baseline_std": float(general_metric_reference(key)["std"]),
            "sample_count": 0,
            "last_updated_at": None,
            "method": str(general_metric_reference(key)["method"]),
        }
        for key in current_metrics
    }
    general_deviation = DeviationSection.model_validate(
        build_deviation_section(
            current_metrics,
            general_reference_metrics,
            baseline_source="general",
            baseline_available=True,
        )
    )
    personal_available = payload.existing_baseline.baseline_available and bool(existing_baseline_metrics)
    personal_deviation = DeviationSection.model_validate(
        build_deviation_section(
            current_metrics,
            existing_baseline_metrics if personal_available else general_reference_metrics,
            baseline_source="personal" if personal_available else "general",
            baseline_available=personal_available,
        )
    )
    history_count = max(payload.history.baseline_exam_count, max((item.get("sample_count", 0) for item in existing_baseline_metrics.values()), default=0))
    update_eligibility_data = evaluate_update_eligibility(
        history_count=history_count,
        all_channels_done=payload.context.all_channels_done,
        critical_quality_flags=payload.context.critical_quality_flags,
        data_reliability=payload.context.data_reliability,
        overall_band=payload.context.overall_band,
    )
    next_baseline = NextBaselineSnapshot.model_validate(
        build_next_baseline(
            current=current_metrics,
            history=history_vectors,
            history_count=history_count,
            existing_baseline_metrics=existing_baseline_metrics,
            refreshed_at=refreshed_at,
            update_eligibility=update_eligibility_data,
        )
    )

    response = BaselineCalculationResponse(
        schema_version=payload.schema_version,
        algorithm_version=payload.algorithm_version,
        refreshed_at=refreshed_at,
        general_deviation=general_deviation,
        personal_deviation=personal_deviation,
        update_eligibility=UpdateEligibility.model_validate(update_eligibility_data),
        next_baseline=next_baseline,
    )
    logger.info(
        "baseline_calculation_succeeded examination_id=%s specialist_id=%s general_band=%s personal_band=%s personal_source=%s update_eligible=%s update_reason=%s next_exam_count=%s next_baseline_available=%s significant_general=%s significant_personal=%s",
        payload.examination_id,
        payload.specialist_id,
        response.general_deviation.band,
        response.personal_deviation.band,
        response.personal_deviation.baseline_source,
        response.update_eligibility.eligible,
        response.update_eligibility.reason,
        response.next_baseline.exam_count,
        response.next_baseline.baseline_available,
        ",".join(response.general_deviation.significant_deviations) or "none",
        ",".join(response.personal_deviation.significant_deviations) or "none",
    )
    return response


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
    try:
        return calculate_baseline(payload)
    except HTTPException:
        raise
    except Exception:
        logger.exception(
            "baseline_calculation_failed examination_id=%s specialist_id=%s",
            payload.examination_id,
            payload.specialist_id,
        )
        raise
