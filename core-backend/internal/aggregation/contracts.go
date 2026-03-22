package aggregation

import "time"

const (
	SchemaVersionV1          = 1
	AggregationVersionV1     = "agg-v1"
	BaselineAlgorithmVersion = "baseline-v1"

	MetricKeyOverallProxyIndex    = "overall_proxy_index"
	MetricKeyTextProxySignal      = "text_proxy_signal"
	MetricKeyAcousticProxySignal  = "acoustic_proxy_signal"
	MetricKeyParalinguisticSignal = "paralinguistic_proxy_signal"
	MetricKeySpeechStabilityProxy = "speech_stability_proxy"

	ExplanationKindSummary = "summary"
)

var StableMetricKeys = []string{
	MetricKeyOverallProxyIndex,
	MetricKeyTextProxySignal,
	MetricKeyAcousticProxySignal,
	MetricKeyParalinguisticSignal,
	MetricKeySpeechStabilityProxy,
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
	Delta                      float64 `json:"delta"`
	Band                       string  `json:"band"`
	ReferencePopulationVersion string  `json:"reference_population_version,omitempty"`
	BaselineExamCount          int     `json:"baseline_exam_count,omitempty"`
	UpdateEligible             bool    `json:"update_eligible,omitempty"`
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

type BaselineRequest struct {
	SchemaVersion                     int                   `json:"schema_version"`
	AlgorithmVersion                  string                `json:"algorithm_version"`
	SpecialistID                      int64                 `json:"specialist_id"`
	ExaminationID                     int64                 `json:"examination_id"`
	GeneratedAt                       time.Time             `json:"generated_at"`
	Metrics                           []BaselineMetricValue `json:"metrics"`
	History                           BaselineHistory       `json:"history"`
	GeneralReferencePopulationVersion string                `json:"general_reference_population_version"`
}

type BaselineScore struct {
	Score        float64                        `json:"score"`
	Band         string                         `json:"band"`
	MetricScores map[string]BaselineMetricScore `json:"metric_scores,omitempty"`
}

type BaselineUpdateEligibility struct {
	Eligible                     bool   `json:"eligible"`
	Reason                       string `json:"reason"`
	BaselineExamCountAfterUpdate int    `json:"baseline_exam_count_after_update"`
}

type BaselineMetricScore struct {
	Delta float64 `json:"delta"`
	Band  string  `json:"band"`
}

type NextBaseline struct {
	ExamCount int                `json:"exam_count"`
	Centers   map[string]float64 `json:"centers"`
	Scales    map[string]float64 `json:"scales"`
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
