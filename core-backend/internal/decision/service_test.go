package decision

import (
	"context"
	"testing"
	"time"

	"diplom/internal/aggregation"
	"diplom/internal/audit"
)

func TestCreateDecisionInput(t *testing.T) {
	profile := sampleAggregatedProfile()

	input := NewDecisionInput(profile, DecisionServiceMetadata{
		TargetSystem: "kesmi",
		DeliveryMode: "placeholder",
		Message:      PlaceholderMessage,
	})

	if input.SchemaVersion != SchemaVersionV1 {
		t.Fatalf("expected schema_version=%d, got %d", SchemaVersionV1, input.SchemaVersion)
	}
	if input.PayloadVersion != PayloadVersionV1 {
		t.Fatalf("expected payload_version=%q, got %q", PayloadVersionV1, input.PayloadVersion)
	}
	if input.AggregationVersion != profile.AggregationVersion {
		t.Fatalf("expected aggregation_version=%q, got %q", profile.AggregationVersion, input.AggregationVersion)
	}
	if input.ExaminationID != profile.ExaminationID || input.SpecialistID != profile.SpecialistID {
		t.Fatalf("expected identifiers from aggregated profile, got examination=%d specialist=%d", input.ExaminationID, input.SpecialistID)
	}
	if input.Summary.PrimaryMetricKey != aggregation.MetricKeyOverallProxyIndex {
		t.Fatalf("expected primary metric key to propagate, got %q", input.Summary.PrimaryMetricKey)
	}
	if len(input.Metrics) != 1 || len(input.ChannelContributions) != 1 {
		t.Fatalf("expected metrics and contributions to propagate, got metrics=%d contributions=%d", len(input.Metrics), len(input.ChannelContributions))
	}
	if input.BaselineSnapshot.Personal.BaselineExamCount != 4 {
		t.Fatalf("expected personal baseline exam count=4, got %d", input.BaselineSnapshot.Personal.BaselineExamCount)
	}
	if input.ServiceMetadata.TargetSystem != "kesmi" {
		t.Fatalf("expected service metadata target system to propagate, got %q", input.ServiceMetadata.TargetSystem)
	}
}

func TestDecisionPendingTransition(t *testing.T) {
	repo := newDecisionRepoStub()
	service := NewService(repo, decisionExecutorStub{}, Config{MaxAttempts: 2})

	if err := service.CreatePendingDecision(context.Background(), sampleAggregatedProfile()); err != nil {
		t.Fatalf("create pending decision: %v", err)
	}
	if repo.createdSnapshot == nil {
		t.Fatal("expected decision snapshot to be created")
	}
	if repo.createdSnapshot.State != DecisionStatePending {
		t.Fatalf("expected decision state=%q, got %q", DecisionStatePending, repo.createdSnapshot.State)
	}
	if repo.examinationStatus != "decision_pending" {
		t.Fatalf("expected examination status to move to decision_pending, got %q", repo.examinationStatus)
	}
	if repo.createdSnapshot.Recommendation != DecisionRecommendationUnavailable {
		t.Fatalf("expected placeholder recommendation=%q, got %q", DecisionRecommendationUnavailable, repo.createdSnapshot.Recommendation)
	}
}

func TestDecisionSuccessMarksCompleted(t *testing.T) {
	repo := newDecisionRepoStub()
	repo.pending = []Snapshot{{
		ID:                 1,
		ExaminationID:      101,
		SpecialistID:       55,
		State:              DecisionStatePending,
		PayloadVersion:     PayloadVersionV1,
		AggregationVersion: "agg-v1",
		MaxAttempts:        2,
	}}
	service := NewService(repo, decisionExecutorStub{
		response: ExecutionResponse{HTTPStatus: 200, RawResponse: []byte(`{"ok":true}`)},
	}, Config{MaxAttempts: 2})

	if err := service.DeliverPending(context.Background(), repo.pending[0]); err != nil {
		t.Fatalf("deliver pending decision: %v", err)
	}
	if repo.markedSuccess != 1 {
		t.Fatalf("expected success projection to be recorded once, got %d", repo.markedSuccess)
	}
	if repo.examinationStatus != "completed" {
		t.Fatalf("expected examination status to become completed, got %q", repo.examinationStatus)
	}
	if repo.lastFinalize.State != DecisionStateSucceeded {
		t.Fatalf("expected decision state=%q, got %q", DecisionStateSucceeded, repo.lastFinalize.State)
	}
	if repo.lastFinalize.Recommendation != DecisionRecommendationUnavailable {
		t.Fatalf("expected placeholder recommendation=%q, got %q", DecisionRecommendationUnavailable, repo.lastFinalize.Recommendation)
	}
	if repo.lastFinalize.Message != PlaceholderMessage {
		t.Fatalf("expected message=%q, got %q", PlaceholderMessage, repo.lastFinalize.Message)
	}
}

func TestDecisionBusinessErrorMarksCompleted(t *testing.T) {
	repo := newDecisionRepoStub()
	snapshot := Snapshot{ID: 1, ExaminationID: 101, SpecialistID: 55, State: DecisionStatePending, PayloadVersion: PayloadVersionV1, AggregationVersion: "agg-v1", MaxAttempts: 2}
	service := NewService(repo, decisionExecutorStub{
		response: ExecutionResponse{
			State:        DecisionStateBusinessError,
			ErrorClass:   "business",
			ErrorCode:    "unknown_model",
			ErrorMessage: "unknown model",
			HTTPStatus:   404,
		},
	}, Config{MaxAttempts: 2})

	if err := service.DeliverPending(context.Background(), snapshot); err != nil {
		t.Fatalf("deliver pending decision: %v", err)
	}
	if repo.markedFailed != 1 {
		t.Fatalf("expected business failure to be recorded once, got %d", repo.markedFailed)
	}
	if repo.examinationStatus != "completed" {
		t.Fatalf("expected examination status to become completed, got %q", repo.examinationStatus)
	}
	if repo.lastFinalize.State != DecisionStateBusinessError {
		t.Fatalf("expected state=%q, got %q", DecisionStateBusinessError, repo.lastFinalize.State)
	}
}

func TestDecisionTransportExhaustedMarksCompleted(t *testing.T) {
	repo := newDecisionRepoStub()
	snapshot := Snapshot{ID: 1, ExaminationID: 101, SpecialistID: 55, State: DecisionStatePending, PayloadVersion: PayloadVersionV1, AggregationVersion: "agg-v1", AttemptCount: 1, MaxAttempts: 2}
	service := NewService(repo, decisionExecutorStub{
		response: ExecutionResponse{
			State:        DecisionStateTransportExhausted,
			ErrorClass:   "transport",
			ErrorCode:    "pool_busy",
			ErrorMessage: "busy",
			HTTPStatus:   503,
			Retryable:    true,
		},
	}, Config{MaxAttempts: 2})

	if err := service.DeliverPending(context.Background(), snapshot); err != nil {
		t.Fatalf("deliver pending decision: %v", err)
	}
	if repo.markedFailed != 1 {
		t.Fatalf("expected exhausted transport failure to be recorded once, got %d", repo.markedFailed)
	}
	if repo.lastFinalize.State != DecisionStateTransportExhausted {
		t.Fatalf("expected state=%q, got %q", DecisionStateTransportExhausted, repo.lastFinalize.State)
	}
	if repo.examinationStatus != "completed" {
		t.Fatalf("expected examination status to become completed, got %q", repo.examinationStatus)
	}
}

func TestDecisionTerminalStateWritesAuditEvent(t *testing.T) {
	repo := newDecisionRepoStub()
	auditRepo := &decisionAuditRepoStub{}
	service := NewService(repo, decisionExecutorStub{
		response: ExecutionResponse{HTTPStatus: 200, RawResponse: []byte(`{"ok":true}`)},
	}, Config{MaxAttempts: 2}, audit.NewService(auditRepo))

	snapshot := Snapshot{ID: 1, ExaminationID: 101, SpecialistID: 55, State: DecisionStatePending, PayloadVersion: PayloadVersionV1, AggregationVersion: "agg-v1", MaxAttempts: 2}
	if err := service.DeliverPending(context.Background(), snapshot); err != nil {
		t.Fatalf("deliver pending decision: %v", err)
	}
	if len(auditRepo.events) != 1 {
		t.Fatalf("expected one audit event, got %d", len(auditRepo.events))
	}
	if auditRepo.events[0].Type != audit.EventTypeDecisionCompleted {
		t.Fatalf("expected decision.completed event, got %q", auditRepo.events[0].Type)
	}
}

type decisionExecutorStub struct {
	response ExecutionResponse
	err      error
	execute  func(context.Context, DecisionInput, string) (ExecutionResponse, error)
}

func (s decisionExecutorStub) Execute(ctx context.Context, input DecisionInput, correlationID string) (ExecutionResponse, error) {
	if s.execute != nil {
		return s.execute(ctx, input, correlationID)
	}
	return s.response, s.err
}

type decisionRepoStub struct {
	input             DecisionInput
	pending           []Snapshot
	createdSnapshot   *Snapshot
	examinationStatus string
	appendedAttempts  int
	markedSuccess     int
	markedFailed      int
	lastFinalize      FinalizeInput
}

func newDecisionRepoStub() *decisionRepoStub {
	return &decisionRepoStub{
		input: NewDecisionInput(sampleAggregatedProfile(), DecisionServiceMetadata{
			TargetSystem: "kesmi",
			DeliveryMode: "placeholder",
			Message:      PlaceholderMessage,
		}),
	}
}

func (r *decisionRepoStub) CreatePendingSnapshot(_ context.Context, profile aggregation.AggregatedProfile, _ DecisionInput, maxAttempts int32) (Snapshot, error) {
	snapshot := Snapshot{
		ID:                 1,
		ExaminationID:      profile.ExaminationID,
		SpecialistID:       profile.SpecialistID,
		State:              DecisionStatePending,
		PayloadVersion:     PayloadVersionV1,
		AggregationVersion: profile.AggregationVersion,
		Recommendation:     DecisionRecommendationUnavailable,
		Message:            PlaceholderMessage,
		MaxAttempts:        maxAttempts,
	}
	r.createdSnapshot = &snapshot
	r.examinationStatus = "decision_pending"
	return snapshot, nil
}

func (r *decisionRepoStub) ListPendingSnapshots(context.Context) ([]Snapshot, error) {
	return r.pending, nil
}

func (r *decisionRepoStub) LoadDecisionInput(context.Context, int64) (DecisionInput, error) {
	return r.input, nil
}

func (r *decisionRepoStub) AppendAttempt(context.Context, AttemptRecord) error {
	r.appendedAttempts++
	return nil
}

func (r *decisionRepoStub) MarkRetryPending(_ context.Context, input RetryPendingInput) error {
	r.examinationStatus = "decision_pending"
	if len(r.pending) > 0 {
		r.pending[0].AttemptCount = input.AttemptCount
	}
	return nil
}

func (r *decisionRepoStub) MarkSucceeded(_ context.Context, input FinalizeInput) error {
	r.markedSuccess++
	r.lastFinalize = input
	r.examinationStatus = "completed"
	return nil
}

func (r *decisionRepoStub) MarkFailed(_ context.Context, input FinalizeInput) error {
	r.markedFailed++
	r.lastFinalize = input
	r.examinationStatus = "completed"
	return nil
}

func sampleAggregatedProfile() aggregation.AggregatedProfile {
	return aggregation.AggregatedProfile{
		AggregationVersion: "agg-v1",
		ExaminationID:      101,
		SpecialistID:       55,
		GeneratedAt:        time.Date(2026, 3, 23, 10, 0, 0, 0, time.UTC),
		Summary: aggregation.ProfileSummary{
			OverallScore:     0.58,
			OverallBand:      "elevated",
			PrimaryMetricKey: aggregation.MetricKeyOverallProxyIndex,
		},
		Metrics: []aggregation.Metric{{
			Key:       aggregation.MetricKeyOverallProxyIndex,
			Label:     "Сводный прокси-индекс",
			Value:     0.58,
			Scale:     "0..1",
			Direction: "higher_means_more_deviation",
		}},
		ChannelContributions: []aggregation.ChannelContribution{{
			Channel:      "text",
			MetricKey:    aggregation.MetricKeyOverallProxyIndex,
			Weight:       0.33,
			Contribution: 0.17,
			EvidenceKeys: []string{"text_proxy_signal"},
		}},
		BaselineSnapshot: aggregation.BaselineSnapshot{
			AlgorithmVersion: aggregation.BaselineAlgorithmVersion,
			General: aggregation.BaselineDeviation{
				Delta: 0.21,
				Band:  "mild",
			},
			Personal: aggregation.BaselineDeviation{
				Delta:             0.37,
				Band:              "moderate",
				BaselineExamCount: 4,
				UpdateEligible:    false,
			},
		},
	}
}

type decisionAuditRepoStub struct {
	events []audit.Event
}

func (s *decisionAuditRepoStub) Append(_ context.Context, event audit.Event) error {
	s.events = append(s.events, event)
	return nil
}

func (s *decisionAuditRepoStub) List(context.Context, audit.ListFilter) ([]audit.Event, error) {
	return append([]audit.Event(nil), s.events...), nil
}
