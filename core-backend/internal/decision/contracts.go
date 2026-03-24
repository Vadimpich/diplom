package decision

import (
	"diplom/internal/aggregation"
)

const SchemaVersionV1 = 1
const PayloadVersionV1 = "decision-input-v1"
const DecisionStatePending = "pending"
const DecisionStateSucceeded = "succeeded"
const DecisionStateTransportExhausted = "transport_exhausted"
const DecisionStateBusinessError = "business_error"
const DecisionRecommendationUnavailable = "unavailable"
const DecisionRecommendationAllowed = "allowed"
const DecisionRecommendationRisk = "risk"
const DecisionRecommendationDenied = "denied"

type DecisionSummary struct {
	OverallScore     float64 `json:"overall_score"`
	OverallBand      string  `json:"overall_band"`
	PrimaryMetricKey string  `json:"primary_metric_key"`
}

type DecisionMetric struct {
	Key       string  `json:"key"`
	Label     string  `json:"label"`
	Value     float64 `json:"value"`
	Scale     string  `json:"scale"`
	Direction string  `json:"direction"`
}

type DecisionContribution struct {
	Channel      string   `json:"channel"`
	MetricKey    string   `json:"metric_key"`
	Weight       float64  `json:"weight"`
	Contribution float64  `json:"contribution"`
	EvidenceKeys []string `json:"evidence_keys"`
}

type DecisionBaselineDeviation struct {
	Delta                      float64 `json:"delta"`
	Band                       string  `json:"band"`
	ReferencePopulationVersion string  `json:"reference_population_version,omitempty"`
	BaselineExamCount          int     `json:"baseline_exam_count,omitempty"`
	UpdateEligible             bool    `json:"update_eligible,omitempty"`
}

type DecisionBaselineSnapshot struct {
	AlgorithmVersion string                    `json:"algorithm_version"`
	General          DecisionBaselineDeviation `json:"general"`
	Personal         DecisionBaselineDeviation `json:"personal"`
}

type DecisionServiceMetadata struct {
	TargetSystem string `json:"target_system"`
	DeliveryMode string `json:"delivery_mode"`
	Message      string `json:"message"`
}

type DecisionInput struct {
	SchemaVersion        int                      `json:"schema_version"`
	PayloadVersion       string                   `json:"payload_version"`
	AggregationVersion   string                   `json:"aggregation_version"`
	ExaminationID        int64                    `json:"examination_id"`
	SpecialistID         int64                    `json:"specialist_id"`
	GeneratedAt          string                   `json:"generated_at"`
	Summary              DecisionSummary          `json:"summary"`
	Metrics              []DecisionMetric         `json:"metrics"`
	ChannelContributions []DecisionContribution   `json:"channel_contributions"`
	BaselineSnapshot     DecisionBaselineSnapshot `json:"baseline_snapshot"`
	ServiceMetadata      DecisionServiceMetadata  `json:"service_metadata"`
}

type DecisionDiagnostics struct {
	ErrorClass   *string `json:"error_class"`
	ErrorCode    *string `json:"error_code"`
	ErrorMessage *string `json:"error_message"`
	HTTPStatus   *int    `json:"http_status"`
	Retryable    bool    `json:"retryable"`
}

type DecisionResult struct {
	State                string              `json:"state"`
	Recommendation       string              `json:"recommendation"`
	Message              string              `json:"message"`
	CorrelationID        string              `json:"correlation_id"`
	AttemptCount         int32               `json:"attempt_count"`
	MaxAttempts          int32               `json:"max_attempts"`
	LastAttemptAt        *string             `json:"last_attempt_at"`
	Diagnostics          DecisionDiagnostics `json:"diagnostics"`
	RawResponseAvailable bool                `json:"raw_response_available"`
}

func NewDecisionInput(profile aggregation.AggregatedProfile, metadata DecisionServiceMetadata) DecisionInput {
	input := DecisionInput{
		SchemaVersion:      SchemaVersionV1,
		PayloadVersion:     PayloadVersionV1,
		AggregationVersion: profile.AggregationVersion,
		ExaminationID:      profile.ExaminationID,
		SpecialistID:       profile.SpecialistID,
		GeneratedAt:        profile.GeneratedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		Summary: DecisionSummary{
			OverallScore:     profile.Summary.OverallScore,
			OverallBand:      profile.Summary.OverallBand,
			PrimaryMetricKey: profile.Summary.PrimaryMetricKey,
		},
		BaselineSnapshot: DecisionBaselineSnapshot{
			AlgorithmVersion: profile.BaselineSnapshot.AlgorithmVersion,
			General: DecisionBaselineDeviation{
				Delta:                      profile.BaselineSnapshot.General.Delta,
				Band:                       profile.BaselineSnapshot.General.Band,
				ReferencePopulationVersion: profile.BaselineSnapshot.General.ReferencePopulationVersion,
			},
			Personal: DecisionBaselineDeviation{
				Delta:             profile.BaselineSnapshot.Personal.Delta,
				Band:              profile.BaselineSnapshot.Personal.Band,
				BaselineExamCount: profile.BaselineSnapshot.Personal.BaselineExamCount,
				UpdateEligible:    profile.BaselineSnapshot.Personal.UpdateEligible,
			},
		},
		ServiceMetadata: metadata,
	}

	input.Metrics = make([]DecisionMetric, 0, len(profile.Metrics))
	for _, metric := range profile.Metrics {
		input.Metrics = append(input.Metrics, DecisionMetric{
			Key:       metric.Key,
			Label:     metric.Label,
			Value:     metric.Value,
			Scale:     metric.Scale,
			Direction: metric.Direction,
		})
	}

	input.ChannelContributions = make([]DecisionContribution, 0, len(profile.ChannelContributions))
	for _, contribution := range profile.ChannelContributions {
		input.ChannelContributions = append(input.ChannelContributions, DecisionContribution{
			Channel:      contribution.Channel,
			MetricKey:    contribution.MetricKey,
			Weight:       contribution.Weight,
			Contribution: contribution.Contribution,
			EvidenceKeys: append([]string(nil), contribution.EvidenceKeys...),
		})
	}

	return input
}
