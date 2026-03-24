package baselineclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"diplom/internal/aggregation"
	"diplom/internal/observability"
)

func TestBaselineClientPropagatesTraceContext(t *testing.T) {
	var gotRequestID string
	var gotTraceParent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequestID = r.Header.Get("X-Request-Id")
		gotTraceParent = r.Header.Get("traceparent")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"schema_version":1,"algorithm_version":"baseline-v1","refreshed_at":"2026-03-23T18:10:12Z","general_deviation":{"score":0,"band":"low","metric_scores":{}},"personal_deviation":{"score":0,"band":"low","metric_scores":{}},"update_eligibility":{"eligible":false,"reason":"none","baseline_exam_count_after_update":0},"next_baseline":{"exam_count":0,"metric_keys":[],"centers":{},"scales":{},"metric_vectors":[],"refreshed_at":"2026-03-23T18:10:12Z"}}`))
	}))
	defer server.Close()

	client := New(server.URL, time.Second)
	ctx := observability.WithTraceContext(context.Background(), observability.BuildTraceContext("req-100", "00-8ec8c1b6409f4a6cb80cfcb4f74aa98c-5d7c1f97db7840b3-01", ""))
	_, err := client.Calculate(ctx, aggregation.BaselineRequest{SchemaVersion: 1, AlgorithmVersion: "baseline-v1"})
	if err != nil {
		t.Fatalf("calculate baseline: %v", err)
	}
	if gotRequestID != "req-100" || gotTraceParent == "" {
		t.Fatalf("expected propagated headers, got request_id=%q traceparent=%q", gotRequestID, gotTraceParent)
	}
}
