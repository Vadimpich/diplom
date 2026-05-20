package aggregation

import (
	"time"

	"diplom/internal/processing"
)

const (
	SchemaVersionV1          = 1
	AggregationVersionV1     = "agg-v1"
	BaselineAlgorithmVersion = "baseline-v1"

	MetricKeyOverallDeviationIndex  = "overall_deviation_index"
	MetricKeyTextRiskSignal         = "text_risk_signal"
	MetricKeyAcousticStressSignal   = "acoustic_stress_signal"
	MetricKeyParalinguisticBehavior = "paralinguistic_behavior_signal"
	MetricKeySpeechStabilityScore   = "speech_stability_score"

	ExplanationKindSummary = "summary"
)

var StableMetricKeys = []string{
	MetricKeyOverallDeviationIndex,
	MetricKeyTextRiskSignal,
	MetricKeyAcousticStressSignal,
	MetricKeyParalinguisticBehavior,
	MetricKeySpeechStabilityScore,
}

type ProfileSummary struct {
	OverallScore                     float64 `json:"overall_score"`
	OverallBand                      string  `json:"overall_band"`
	PrimaryMetricKey                 string  `json:"primary_metric_key"`
	NeutralRecommendationPlaceholder string  `json:"neutral_recommendation_placeholder"`
}

type Metric struct {
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

type Explanation struct {
	Position int    `json:"position"`
	Kind     string `json:"kind"`
	Text     string `json:"text"`
}

type BaselineDeviation struct {
	Delta                      float64  `json:"delta"`
	Band                       string   `json:"band"`
	BaselineAvailable          bool     `json:"baseline_available,omitempty"`
	BaselineSource             string   `json:"baseline_source,omitempty"`
	ReferencePopulationVersion string   `json:"reference_population_version,omitempty"`
	BaselineExamCount          int      `json:"baseline_exam_count,omitempty"`
	UpdateEligible             bool     `json:"update_eligible,omitempty"`
	DataReliability            float64  `json:"data_reliability,omitempty"`
	SignificantDeviations      []string `json:"significant_deviations,omitempty"`
}

type BaselineSnapshot struct {
	AlgorithmVersion string            `json:"algorithm_version"`
	RefreshedAt      time.Time         `json:"refreshed_at"`
	General          BaselineDeviation `json:"general"`
	Personal         BaselineDeviation `json:"personal"`
}

type AggregatedProfile struct {
	SchemaVersion        int                   `json:"schema_version"`
	AggregationVersion   string                `json:"aggregation_version"`
	ExaminationID        int64                 `json:"examination_id"`
	SpecialistID         int64                 `json:"specialist_id"`
	Status               string                `json:"status"`
	GeneratedAt          time.Time             `json:"generated_at"`
	Summary              ProfileSummary        `json:"summary"`
	Metrics              []Metric              `json:"metrics"`
	ChannelContributions []ChannelContribution `json:"channel_contributions"`
	Explanations         []Explanation         `json:"explanations"`
	BaselineSnapshot     BaselineSnapshot      `json:"baseline_snapshot"`
}

type DecisionChannel struct {
	Scores       map[string]float64 `json:"scores"`
	QualityFlags []string           `json:"quality_flags"`
}

type DecisionBaseline struct {
	Source                 string             `json:"source"`
	Available              bool               `json:"available"`
	BaselineDeviationIndex float64            `json:"baseline_deviation_index"`
	SignificantDeviations  []string           `json:"significant_deviations"`
	ZScores                map[string]float64 `json:"z_scores"`
}

type DecisionDerivedIndicators struct {
	SemanticStressIndex        float64 `json:"semantic_stress_index"`
	AcousticActivationIndex    float64 `json:"acoustic_activation_index"`
	SpeechDisorganizationIndex float64 `json:"speech_disorganization_index"`
	BaselineShiftIndex         float64 `json:"baseline_shift_index"`
}

type DecisionPayload struct {
	ExaminationID     int64                      `json:"examination_id"`
	SpecialistID      int64                      `json:"specialist_id"`
	DataReliability   float64                    `json:"data_reliability"`
	Channels          map[string]DecisionChannel `json:"channels"`
	Baseline          DecisionBaseline           `json:"baseline"`
	DerivedIndicators DecisionDerivedIndicators  `json:"derived_indicators"`
	Evidence          []string                   `json:"evidence"`
}

type BaselineMetricValue struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

type BaselineHistoryVector struct {
	ExaminationID int64                 `json:"examination_id"`
	GeneratedAt   time.Time             `json:"generated_at"`
	Metrics       []BaselineMetricValue `json:"metrics"`
}

type BaselineHistory struct {
	BaselineExamCount int                     `json:"baseline_exam_count"`
	MetricVectors     []BaselineHistoryVector `json:"metric_vectors"`
}

type ExistingBaselineMetric struct {
	BaselineMean  float64    `json:"baseline_mean"`
	BaselineStd   float64    `json:"baseline_std"`
	SampleCount   int        `json:"sample_count"`
	LastUpdatedAt *time.Time `json:"last_updated_at,omitempty"`
	Method        string     `json:"method"`
}

type ExistingBaselinePayload struct {
	BaselineAvailable bool                              `json:"baseline_available"`
	Metrics           map[string]ExistingBaselineMetric `json:"metrics"`
}

type BaselineContext struct {
	AllChannelsDone      bool     `json:"all_channels_done"`
	CriticalQualityFlags []string `json:"critical_quality_flags"`
	DataReliability      float64  `json:"data_reliability"`
	OverallBand          string   `json:"overall_band"`
}

type BaselineRequest struct {
	SchemaVersion                     int                     `json:"schema_version"`
	AlgorithmVersion                  string                  `json:"algorithm_version"`
	SpecialistID                      int64                   `json:"specialist_id"`
	ExaminationID                     int64                   `json:"examination_id"`
	GeneratedAt                       time.Time               `json:"generated_at"`
	Metrics                           []BaselineMetricValue   `json:"metrics"`
	History                           BaselineHistory         `json:"history"`
	ExistingBaseline                  ExistingBaselinePayload `json:"existing_baseline"`
	Context                           BaselineContext         `json:"context"`
	GeneralReferencePopulationVersion string                  `json:"general_reference_population_version"`
}

type BaselineScore struct {
	Score                 float64                        `json:"score"`
	Band                  string                         `json:"band"`
	BaselineAvailable     bool                           `json:"baseline_available"`
	BaselineSource        string                         `json:"baseline_source"`
	SignificantDeviations []string                       `json:"significant_deviations,omitempty"`
	MetricScores          map[string]BaselineMetricScore `json:"metric_scores,omitempty"`
}

type BaselineUpdateEligibility struct {
	Eligible                     bool    `json:"eligible"`
	Reason                       string  `json:"reason"`
	BaselineExamCountAfterUpdate int     `json:"baseline_exam_count_after_update"`
	DataReliability              float64 `json:"data_reliability"`
}

type BaselineMetricScore struct {
	BaselineAvailable bool    `json:"baseline_available"`
	BaselineSource    string  `json:"baseline_source"`
	BaselineMean      float64 `json:"baseline_mean"`
	BaselineStd       float64 `json:"baseline_std"`
	SampleCount       int     `json:"sample_count"`
	Method            string  `json:"method"`
	Delta             float64 `json:"delta"`
	ZScore            float64 `json:"z_score"`
	DeviationLevel    string  `json:"deviation_level"`
}

type NextBaseline struct {
	ExamCount         int                               `json:"exam_count"`
	BaselineAvailable bool                              `json:"baseline_available"`
	Metrics           map[string]ExistingBaselineMetric `json:"metrics"`
	RefreshedAt       time.Time                         `json:"refreshed_at"`
}

type BaselineResponse struct {
	SchemaVersion     int                       `json:"schema_version"`
	AlgorithmVersion  string                    `json:"algorithm_version"`
	RefreshedAt       time.Time                 `json:"refreshed_at"`
	GeneralDeviation  BaselineScore             `json:"general_deviation"`
	PersonalDeviation BaselineScore             `json:"personal_deviation"`
	UpdateEligibility BaselineUpdateEligibility `json:"update_eligibility"`
	NextBaseline      NextBaseline              `json:"next_baseline"`
}

type ChannelPayloadMap map[string]processing.CanonicalChannelPayload
