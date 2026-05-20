package decision

import "diplom/internal/aggregation"

const SchemaVersionV1 = 1
const PayloadVersionV1 = "decision-input-v2"
const DecisionStatePending = "pending"
const DecisionStateSucceeded = "succeeded"
const DecisionStateTransportExhausted = "transport_exhausted"
const DecisionStateBusinessError = "business_error"
const DecisionRecommendationUnavailable = "unavailable"
const DecisionRecommendationAllowed = "allowed"
const DecisionRecommendationRisk = "risk"
const DecisionRecommendationDenied = "denied"
const DecisionCodeAllow = "allow"
const DecisionCodeMonitoring = "monitoring"
const DecisionCodeExtendedCheck = "extended_check"
const DecisionCodeNoAccess = "no_access"

type DecisionServiceMetadata struct {
	TargetSystem string `json:"target_system"`
	DeliveryMode string `json:"delivery_mode"`
	Message      string `json:"message"`
}

type DecisionInput struct {
	SchemaVersion      int                                    `json:"schema_version"`
	PayloadVersion     string                                 `json:"payload_version"`
	AggregationVersion string                                 `json:"aggregation_version"`
	ExaminationID      int64                                  `json:"examination_id"`
	SpecialistID       int64                                  `json:"specialist_id"`
	GeneratedAt        string                                 `json:"generated_at"`
	DataReliability    float64                                `json:"data_reliability"`
	Channels           map[string]aggregation.DecisionChannel `json:"channels"`
	Baseline           aggregation.DecisionBaseline           `json:"baseline"`
	DerivedIndicators  aggregation.DecisionDerivedIndicators  `json:"derived_indicators"`
	Evidence           []string                               `json:"evidence"`
	ServiceMetadata    DecisionServiceMetadata                `json:"service_metadata"`
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
	DecisionCode         string              `json:"decision_code,omitempty"`
	RiskClass            string              `json:"risk_class,omitempty"`
	Patterns             []string            `json:"patterns,omitempty"`
	CorrelationID        string              `json:"correlation_id"`
	AttemptCount         int32               `json:"attempt_count"`
	MaxAttempts          int32               `json:"max_attempts"`
	LastAttemptAt        *string             `json:"last_attempt_at"`
	Diagnostics          DecisionDiagnostics `json:"diagnostics"`
	RawResponseAvailable bool                `json:"raw_response_available"`
}

func NewDecisionInput(profile aggregation.AggregatedProfile, payload aggregation.DecisionPayload, metadata DecisionServiceMetadata) DecisionInput {
	return DecisionInput{
		SchemaVersion:      SchemaVersionV1,
		PayloadVersion:     PayloadVersionV1,
		AggregationVersion: profile.AggregationVersion,
		ExaminationID:      profile.ExaminationID,
		SpecialistID:       profile.SpecialistID,
		GeneratedAt:        profile.GeneratedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		DataReliability:    payload.DataReliability,
		Channels:           payload.Channels,
		Baseline:           payload.Baseline,
		DerivedIndicators:  payload.DerivedIndicators,
		Evidence:           append([]string(nil), payload.Evidence...),
		ServiceMetadata:    metadata,
	}
}
