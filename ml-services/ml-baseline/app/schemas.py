from datetime import datetime

from pydantic import BaseModel, ConfigDict, Field


class MetricValue(BaseModel):
    key: str
    value: float


class MetricVector(BaseModel):
    examination_id: int
    generated_at: datetime
    metrics: list[MetricValue] = Field(default_factory=list)


class HistoryPayload(BaseModel):
    baseline_exam_count: int = Field(ge=0)
    metric_vectors: list[MetricVector] = Field(default_factory=list)


class BaselineMetricState(BaseModel):
    baseline_mean: float
    baseline_std: float = Field(ge=0)
    sample_count: int = Field(ge=0)
    last_updated_at: datetime | None = None
    method: str


class ExistingBaselinePayload(BaseModel):
    baseline_available: bool = False
    metrics: dict[str, BaselineMetricState] = Field(default_factory=dict)


class BaselineContext(BaseModel):
    all_channels_done: bool = True
    critical_quality_flags: list[str] = Field(default_factory=list)
    data_reliability: float = Field(ge=0, le=1)
    overall_band: str = "stable"


class BaselineCalculationRequest(BaseModel):
    model_config = ConfigDict(extra="forbid")

    schema_version: int = Field(ge=1)
    algorithm_version: str
    specialist_id: int
    examination_id: int
    generated_at: datetime
    metrics: list[MetricValue] = Field(default_factory=list)
    history: HistoryPayload
    existing_baseline: ExistingBaselinePayload = Field(default_factory=ExistingBaselinePayload)
    context: BaselineContext
    general_reference_population_version: str


class DeviationMetricScore(BaseModel):
    baseline_available: bool
    baseline_source: str
    baseline_mean: float
    baseline_std: float
    sample_count: int = Field(ge=0)
    method: str
    delta: float
    z_score: float
    deviation_level: str


class DeviationSection(BaseModel):
    score: float
    band: str
    baseline_available: bool
    baseline_source: str
    significant_deviations: list[str] = Field(default_factory=list)
    metric_scores: dict[str, DeviationMetricScore] = Field(default_factory=dict)


class UpdateEligibility(BaseModel):
    eligible: bool
    reason: str
    baseline_exam_count_after_update: int = Field(ge=0)
    data_reliability: float = Field(ge=0, le=1)


class NextBaselineSnapshot(BaseModel):
    exam_count: int = Field(ge=0)
    baseline_available: bool = False
    metrics: dict[str, BaselineMetricState] = Field(default_factory=dict)
    refreshed_at: datetime


class BaselineCalculationResponse(BaseModel):
    schema_version: int
    algorithm_version: str
    refreshed_at: datetime
    general_deviation: DeviationSection
    personal_deviation: DeviationSection
    update_eligibility: UpdateEligibility
    next_baseline: NextBaselineSnapshot


class HealthResponse(BaseModel):
    status: str
    service: str
