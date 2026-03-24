package kesmi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequestID = r.Header.Get("X-Request-Id")
		gotTraceParent = r.Header.Get("traceparent")
		gotCorrelationID = r.Header.Get("X-Correlation-Id")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "model-1", time.Second)
	ctx := observability.WithTraceContext(context.Background(), observability.BuildTraceContext("req-100", "00-8ec8c1b6409f4a6cb80cfcb4f74aa98c-5d7c1f97db7840b3-01", ""))
	_, err := client.Execute(ctx, decision.DecisionInput{PayloadVersion: decision.PayloadVersionV1}, "exam-101-kesmi-1")
	if err != nil {
		t.Fatalf("kesmi execute: %v", err)
	}
	if gotRequestID != "req-100" || gotTraceParent == "" || gotCorrelationID != "exam-101-kesmi-1" {
		t.Fatalf("expected propagated headers, got request_id=%q traceparent=%q correlation_id=%q", gotRequestID, gotTraceParent, gotCorrelationID)
	}
}
