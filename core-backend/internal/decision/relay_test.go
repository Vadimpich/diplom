package decision

import (
	"context"
	"testing"
)

func TestDecisionRelayResumesPendingSnapshots(t *testing.T) {
	repo := newDecisionRepoStub()
	repo.pending = []Snapshot{{
		ID:                 1,
		ExaminationID:      101,
		SpecialistID:       55,
		State:              DecisionStatePending,
		PayloadVersion:     PayloadVersionV1,
		AggregationVersion: "agg-v1",
		MaxAttempts:        2,
		CorrelationID:      "exam-101-kesmi-1",
		RequestID:          "req-101",
		TraceParent:        "00-8ec8c1b6409f4a6cb80cfcb4f74aa98c-5d7c1f97db7840b3-01",
	}}

	service := NewService(repo, decisionExecutorStub{
		response: ExecutionResponse{HTTPStatus: 200, RawResponse: []byte(`{"requiredExploredParameters":[{"id":"p32","value":"allow"},{"id":"p33","value":"risk=low; decision=allow; patterns=none"}]}`)},
	}, Config{MaxAttempts: 2})
	relay := NewRelay(repo, service, 0)

	if err := relay.runOnce(context.Background()); err != nil {
		t.Fatalf("run relay once: %v", err)
	}
	if repo.markedSuccess != 1 {
		t.Fatalf("expected one pending snapshot to be completed, got %d", repo.markedSuccess)
	}
	if repo.appendedAttempts != 1 {
		t.Fatalf("expected one append-only attempt to be recorded, got %d", repo.appendedAttempts)
	}
}

func TestDecisionRelayPreservesTraceContext(t *testing.T) {
	repo := newDecisionRepoStub()
	repo.pending = []Snapshot{{
		ID:                 1,
		ExaminationID:      101,
		SpecialistID:       55,
		State:              DecisionStatePending,
		PayloadVersion:     PayloadVersionV1,
		AggregationVersion: "agg-v1",
		MaxAttempts:        2,
		CorrelationID:      "exam-101-kesmi-1",
		RequestID:          "req-101",
		TraceParent:        "00-8ec8c1b6409f4a6cb80cfcb4f74aa98c-5d7c1f97db7840b3-01",
	}}

	var gotRequestID string
	var gotTraceParent string
	service := NewService(repo, decisionExecutorStub{
		execute: func(_ context.Context, _ DecisionInput, _ string) (ExecutionResponse, error) {
			gotRequestID = repo.pending[0].RequestID
			gotTraceParent = repo.pending[0].TraceParent
			return ExecutionResponse{HTTPStatus: 200, RawResponse: []byte(`{"requiredExploredParameters":[{"id":"p32","value":"allow"},{"id":"p33","value":"risk=low; decision=allow; patterns=none"}]}`)}, nil
		},
	}, Config{MaxAttempts: 2})
	relay := NewRelay(repo, service, 0)

	if err := relay.runOnce(context.Background()); err != nil {
		t.Fatalf("run relay once: %v", err)
	}
	if gotRequestID != "req-101" || gotTraceParent == "" {
		t.Fatalf("expected request/trace continuity into decision execution, got request_id=%q traceparent=%q", gotRequestID, gotTraceParent)
	}
}
