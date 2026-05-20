package results

import "context"

type Repository interface {
	GetExaminationResult(context.Context, int64) (ExaminationResultResponse, error)
	GetSpecialistHistory(context.Context, int64) (SpecialistHistoryResponse, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetExaminationResult(ctx context.Context, examinationID int64) (ExaminationResultResponse, error) {
	return s.repo.GetExaminationResult(ctx, examinationID)
}

func (s *Service) GetSpecialistHistory(ctx context.Context, specialistID int64) (SpecialistHistoryResponse, error) {
	return s.repo.GetSpecialistHistory(ctx, specialistID)
}

type ExaminationResultResponse struct {
	SchemaVersion        int                         `json:"schema_version"`
	AggregationVersion   string                      `json:"aggregation_version"`
	ExaminationID        int64                       `json:"examination_id"`
	SpecialistID         int64                       `json:"specialist_id"`
	Status               string                      `json:"status"`
	GeneratedAt          string                      `json:"generated_at"`
	Summary              ExaminationSummary          `json:"summary"`
	Metrics              []ExaminationMetric         `json:"metrics"`
	ChannelContributions []ChannelContribution       `json:"channel_contributions"`
	ChannelReports       []ChannelReport             `json:"channel_reports"`
	Explanations         []Explanation               `json:"explanations"`
	BaselineSnapshot     ExaminationBaselineSnapshot `json:"baseline_snapshot"`
	Decision             DecisionResultView          `json:"decision"`
}

type ExaminationSummary struct {
	OverallScore                     float64 `json:"overall_score"`
	OverallBand                      string  `json:"overall_band"`
	PrimaryMetricKey                 string  `json:"primary_metric_key"`
	NeutralRecommendationPlaceholder string  `json:"neutral_recommendation_placeholder"`
}

type ExaminationMetric struct {
	Key       string  `json:"key"`
	Label     string  `json:"label"`
	Value     float64 `json:"value"`
	Scale     string  `json:"scale"`
	Direction string  `json:"direction"`
}

type ChannelContribution struct {
	Channel      string   `json:"channel"`
	MetricKey    string   `json:"metric_key"`
	Weight       float64  `json:"weight"`
	Contribution float64  `json:"contribution"`
	EvidenceKeys []string `json:"evidence_keys"`
}

type ChannelReport struct {
	Channel      string               `json:"channel"`
	ModelVersion string               `json:"model_version"`
	QualityFlags []string             `json:"quality_flags"`
	Evidence     []string             `json:"evidence"`
	Scores       []ChannelReportScore `json:"scores"`
}

type ChannelReportScore struct {
	Key   string  `json:"key"`
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

type Explanation struct {
	Position int    `json:"position"`
	Kind     string `json:"kind"`
	Text     string `json:"text"`
}

type BaselineDeviation struct {
	Delta                      float64 `json:"delta"`
	Band                       string  `json:"band"`
	BaselineAvailable          bool    `json:"baseline_available,omitempty"`
	BaselineSource             string  `json:"baseline_source,omitempty"`
	ReferencePopulationVersion string  `json:"reference_population_version,omitempty"`
	BaselineExamCount          int     `json:"baseline_exam_count,omitempty"`
	UpdateEligible             bool    `json:"update_eligible,omitempty"`
	DataReliability            float64 `json:"data_reliability,omitempty"`
}

type ExaminationBaselineSnapshot struct {
	AlgorithmVersion string            `json:"algorithm_version"`
	RefreshedAt      string            `json:"refreshed_at"`
	General          BaselineDeviation `json:"general"`
	Personal         BaselineDeviation `json:"personal"`
}

type DecisionDiagnosticsView struct {
	ErrorClass   *string `json:"error_class"`
	ErrorCode    *string `json:"error_code"`
	ErrorMessage *string `json:"error_message"`
	HTTPStatus   *int    `json:"http_status"`
	Retryable    bool    `json:"retryable"`
}

type DecisionResultView struct {
	State                string                  `json:"state"`
	Recommendation       string                  `json:"recommendation"`
	Message              string                  `json:"message"`
	DecisionCode         string                  `json:"decision_code,omitempty"`
	RiskClass            string                  `json:"risk_class,omitempty"`
	Patterns             []string                `json:"patterns,omitempty"`
	CorrelationID        string                  `json:"correlation_id"`
	AttemptCount         int32                   `json:"attempt_count"`
	MaxAttempts          int32                   `json:"max_attempts"`
	LastAttemptAt        *string                 `json:"last_attempt_at"`
	Diagnostics          DecisionDiagnosticsView `json:"diagnostics"`
	RawResponseAvailable bool                    `json:"raw_response_available"`
}

type SpecialistHistoryResponse struct {
	SpecialistID int64                   `json:"specialist_id"`
	Items        []SpecialistHistoryItem `json:"items"`
}

type SpecialistHistoryItem struct {
	ExaminationID    int64            `json:"examination_id"`
	GeneratedAt      string           `json:"generated_at"`
	Status           string           `json:"status"`
	Summary          Summary          `json:"summary"`
	BaselineSnapshot BaselineSnapshot `json:"baseline_snapshot"`
	KeyMetrics       []HistoryMetric  `json:"key_metrics"`
}

type Summary struct {
	OverallScore float64 `json:"overall_score"`
	OverallBand  string  `json:"overall_band"`
}

type BaselineSnapshot struct {
	AlgorithmVersion  string  `json:"algorithm_version"`
	RefreshedAt       string  `json:"refreshed_at"`
	GeneralDelta      float64 `json:"general_delta"`
	PersonalDelta     float64 `json:"personal_delta"`
	BaselineExamCount int     `json:"baseline_exam_count"`
}

type HistoryMetric struct {
	Key               string   `json:"key"`
	Label             string   `json:"label"`
	Value             float64  `json:"value"`
	PreviousValue     *float64 `json:"previous_value"`
	DeltaFromPrevious *float64 `json:"delta_from_previous"`
}
