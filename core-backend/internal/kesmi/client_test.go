package kesmi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"diplom/internal/aggregation"
	"diplom/internal/decision"
	"diplom/internal/observability"
)

func TestRetriesOnlyTransportFailures(t *testing.T) {
	cases := []struct {
		name          string
		err           error
		httpStatus    int
		errorCode     string
		wantRetryable bool
		wantState     string
	}{
		{name: "network timeout", err: errors.New("i/o timeout"), wantRetryable: true, wantState: decision.DecisionStateTransportExhausted},
		{name: "http 503", httpStatus: 503, wantRetryable: true, wantState: decision.DecisionStateTransportExhausted},
		{name: "pool_busy", errorCode: "pool_busy", wantRetryable: true, wantState: decision.DecisionStateTransportExhausted},
		{name: "pool_exhausted", errorCode: "pool_exhausted", wantRetryable: true, wantState: decision.DecisionStateTransportExhausted},
		{name: "unknown model", httpStatus: 404, errorCode: "unknown_model", wantRetryable: false, wantState: decision.DecisionStateBusinessError},
		{name: "parameter mismatch", httpStatus: 400, errorCode: "parameter_type_mismatch", wantRetryable: false, wantState: decision.DecisionStateBusinessError},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := ClassifyFailure(tc.err, tc.httpStatus, tc.errorCode)
			if result.Retryable != tc.wantRetryable {
				t.Fatalf("expected retryable=%t, got %t", tc.wantRetryable, result.Retryable)
			}
			if result.State != tc.wantState {
				t.Fatalf("expected state=%q, got %q", tc.wantState, result.State)
			}
		})
	}
}

func TestKESMIClientPropagatesTraceContext(t *testing.T) {
	var gotRequestID string
	var gotTraceParent string
	var gotCorrelationID string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequestID = r.Header.Get("X-Request-Id")
		gotTraceParent = r.Header.Get("traceparent")
		gotCorrelationID = r.Header.Get("X-Correlation-Id")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "model-1", time.Second)
	ctx := observability.WithTraceContext(context.Background(), observability.BuildTraceContext("req-100", "00-8ec8c1b6409f4a6cb80cfcb4f74aa98c-5d7c1f97db7840b3-01", ""))
	_, err := client.Execute(ctx, sampleDecisionInput(), "exam-101-kesmi-1")
	if err != nil {
		t.Fatalf("kesmi execute: %v", err)
	}
	if gotRequestID != "req-100" || gotTraceParent == "" || gotCorrelationID != "exam-101-kesmi-1" {
		t.Fatalf("expected propagated headers, got request_id=%q traceparent=%q correlation_id=%q", gotRequestID, gotTraceParent, gotCorrelationID)
	}
	if gotBody["modelID"] != "model-1" {
		t.Fatalf("expected modelID=model-1, got %#v", gotBody["modelID"])
	}
	outputParameters, ok := gotBody["outputParameters"].([]any)
	if !ok || len(outputParameters) != 2 || outputParameters[0] != outputFinalDecisionID || outputParameters[1] != outputFinalSummaryID {
		t.Fatalf("unexpected outputParameters: %#v", gotBody["outputParameters"])
	}
	service, ok := gotBody["service"].(map[string]any)
	if !ok {
		t.Fatalf("expected service object, got %#v", gotBody["service"])
	}
	outputFields, ok := service["outputFields"].([]any)
	if !ok || len(outputFields) != 1 || outputFields[0] != "requiredExploredParameters" {
		t.Fatalf("unexpected outputFields: %#v", service["outputFields"])
	}
	incoming, ok := gotBody["incommingParameters"].([]any)
	if !ok || len(incoming) != len(parameterIDs) {
		t.Fatalf("unexpected incommingParameters: %#v", gotBody["incommingParameters"])
	}
	first, ok := incoming[0].(map[string]any)
	if !ok || first["id"] != "i1" || first["value"] != 0.28 {
		t.Fatalf("unexpected first input parameter: %#v", first)
	}
}

func sampleDecisionInput() decision.DecisionInput {
	return decision.DecisionInput{
		PayloadVersion:  decision.PayloadVersionV1,
		DataReliability: 0.91,
		Channels: map[string]aggregation.DecisionChannel{
			"text": {
				Scores: map[string]float64{
					"text_negativity_score": 0.28,
					"text_anxiety_score":    0.34,
					"text_confidence_score": 0.72,
					"text_coherence_score":  0.82,
					"text_evasion_score":    0.18,
				},
			},
			"acoustic": {
				Scores: map[string]float64{
					"acoustic_stress_score":       0.46,
					"voice_stability_score":       0.62,
					"intensity_variability_score": 0.31,
				},
			},
			"paralinguistic": {
				Scores: map[string]float64{
					"hesitation_score":             0.33,
					"speech_disorganization_score": 0.22,
				},
			},
		},
		Baseline: aggregation.DecisionBaseline{
			Available:              false,
			BaselineDeviationIndex: 0.37,
			ZScores: map[string]float64{
				"overall_deviation_index":        1.5,
				"text_risk_signal":               2.1,
				"acoustic_stress_signal":         1.7,
				"paralinguistic_behavior_signal": 1.2,
				"speech_stability_score":         -0.4,
			},
		},
		DerivedIndicators: aggregation.DecisionDerivedIndicators{
			SemanticStressIndex:        0.32,
			AcousticActivationIndex:    0.41,
			SpeechDisorganizationIndex: 0.28,
			BaselineShiftIndex:         0.39,
		},
	}
}
