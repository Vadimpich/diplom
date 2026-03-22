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


class BaselineCalculationRequest(BaseModel):
    model_config = ConfigDict(extra="forbid")

    schema_version: int = Field(ge=1)
    algorithm_version: str
    specialist_id: int
    examination_id: int
    generated_at: datetime
    metrics: list[MetricValue] = Field(default_factory=list)
    history: HistoryPayload
    general_reference_population_version: str


class DeviationMetricScore(BaseModel):
    delta: float
    robust_z: float
    band: str


class DeviationSection(BaseModel):
    score: float
    band: str
    metric_scores: dict[str, DeviationMetricScore] = Field(default_factory=dict)


class UpdateEligibility(BaseModel):
    eligible: bool
    reason: str
    baseline_exam_count_after_update: int = Field(ge=0)


class NextBaselineSnapshot(BaseModel):
    exam_count: int = Field(ge=0)
    centers: dict[str, float] = Field(default_factory=dict)
    scales: dict[str, float] = Field(default_factory=dict)
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
