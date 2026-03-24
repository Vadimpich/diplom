package audit

import (
	"context"
	"testing"
	"time"
)

func TestAppendAuditEvent(t *testing.T) {
	repo := &repositoryStub{}
	service := NewService(repo)

	event := Event{
		Type:          EventTypeProcessingLaunch,
		Key:           "processing-launch:examination:101",
		Outcome:       OutcomeSucceeded,
		HappenedAt:    time.Date(2026, 3, 23, 18, 10, 12, 0, time.UTC),
		RequestID:     "req-123",
		TraceID:       "8ec8c1b6409f4a6cb80cfcb4f74aa98c",
		TraceParent:   "00-8ec8c1b6409f4a6cb80cfcb4f74aa98c-5d7c1f97db7840b3-01",
		CorrelationID: "exam-101-text-v1",
		Resource: ResourceRef{
			Kind: "examination",
			ID:   101,
		},
		DomainRefs: DomainRefs{
			ExaminationID: int64Ptr(101),
			SpecialistID:  int64Ptr(55),
			Channel:       stringPtr("text"),
		},
	}

	if err := service.Append(context.Background(), event); err != nil {
		t.Fatalf("append audit event: %v", err)
	}
	if len(repo.appended) != 1 {
		t.Fatalf("expected one appended audit event, got %d", len(repo.appended))
	}
	if repo.appended[0].Type != EventTypeProcessingLaunch {
		t.Fatalf("expected processing.launch event type, got %q", repo.appended[0].Type)
	}
	if repo.appended[0].TraceParent == "" || repo.appended[0].CorrelationID == "" {
		t.Fatal("expected trace and correlation metadata to be persisted")
	}
}

func TestAppendAuditEventUsesStableEventKey(t *testing.T) {
	repo := &repositoryStub{}
	service := NewService(repo)

	event := Event{
		Type:       EventTypeExaminationFinished,
		Key:        "examination-finished:101",
		Outcome:    OutcomeSucceeded,
		HappenedAt: time.Date(2026, 3, 23, 18, 10, 12, 0, time.UTC),
		Resource: ResourceRef{
			Kind: "examination",
			ID:   101,
		},
	}

	if err := service.Append(context.Background(), event); err != nil {
		t.Fatalf("first append: %v", err)
	}
	if err := service.Append(context.Background(), event); err != nil {
		t.Fatalf("second append with same key should be idempotent: %v", err)
	}
	if len(repo.appended) != 1 {
		t.Fatalf("expected event key to fence duplicates, got %d appends", len(repo.appended))
	}
}

func TestListAuditEvents(t *testing.T) {
	now := time.Date(2026, 3, 23, 18, 10, 12, 0, time.UTC)
	repo := &repositoryStub{
		listResult: []Event{
			{
				Type:       EventTypeDecisionCompleted,
				Outcome:    OutcomeSucceeded,
				HappenedAt: now,
				Resource: ResourceRef{
					Kind: "decision_snapshot",
					ID:   44,
				},
			},
		},
	}
	service := NewService(repo)

	events, err := service.List(context.Background(), ListFilter{
		ResourceKind: stringPtr("decision_snapshot"),
		From:         now.Add(-time.Minute),
		To:           now.Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("list audit events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one event, got %d", len(events))
	}
	if events[0].Type != EventTypeDecisionCompleted {
		t.Fatalf("expected decision.completed, got %q", events[0].Type)
	}
}

type repositoryStub struct {
	appended   []Event
	listResult []Event
	seenKeys   map[string]struct{}
}

func (r *repositoryStub) Append(_ context.Context, event Event) error {
	if r.seenKeys == nil {
		r.seenKeys = map[string]struct{}{}
	}
	if event.Key != "" {
		if _, exists := r.seenKeys[event.Key]; exists {
			return nil
		}
		r.seenKeys[event.Key] = struct{}{}
	}
	r.appended = append(r.appended, event)
	return nil
}

func (r *repositoryStub) List(context.Context, ListFilter) ([]Event, error) {
	return append([]Event(nil), r.listResult...), nil
}

func int64Ptr(value int64) *int64 {
	return &value
}

func stringPtr(value string) *string {
	return &value
}
