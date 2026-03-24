package decision

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"diplom/internal/aggregation"
	"diplom/internal/audit"
)

const PlaceholderMessage = "analysis_not_implemented_yet"

type Executor interface {
	Execute(context.Context, DecisionInput, string) (ExecutionResponse, error)
}

type Repository interface {
	CreatePendingSnapshot(context.Context, aggregation.AggregatedProfile, DecisionInput, int32) (Snapshot, error)
	ListPendingSnapshots(context.Context) ([]Snapshot, error)
	LoadDecisionInput(context.Context, int64) (DecisionInput, error)
	AppendAttempt(context.Context, AttemptRecord) error
	MarkRetryPending(context.Context, RetryPendingInput) error
	MarkSucceeded(context.Context, FinalizeInput) error
	MarkFailed(context.Context, FinalizeInput) error
}

type Config struct {
	MaxAttempts int32
}

type Service struct {
	repo    Repository
	exec    Executor
	config  Config
	auditor *audit.Service
	nowFunc func() time.Time
}

type Snapshot struct {
	ID                 int64
	ExaminationID      int64
	SpecialistID       int64
	State              string
	PayloadVersion     string
	AggregationVersion string
	Recommendation     string
	Message            string
	CorrelationID      string
	RequestID          string
	TraceParent        string
	TraceState         string
	AttemptCount       int32
	MaxAttempts        int32
	LastAttemptAt      *time.Time
}

type ExecutionResponse struct {
	State        string
	ErrorClass   string
	ErrorCode    string
	ErrorMessage string
	HTTPStatus   int
	Retryable    bool
	RawResponse  []byte
}

type AttemptRecord struct {
	SnapshotID      int64
	AttemptNumber   int32
	CorrelationID   string
	RequestPayload  []byte
	ResponsePayload []byte
	ErrorClass      string
	ErrorCode       string
	ErrorMessage    string
	Retryable       bool
	HTTPStatus      int
	StartedAt       time.Time
	FinishedAt      time.Time
}

type RetryPendingInput struct {
	SnapshotID    int64
	CorrelationID string
	AttemptCount  int32
	LastAttemptAt time.Time
	Diagnostics   DecisionDiagnostics
}

func NewService(repo Repository, exec Executor, cfg Config, auditors ...*audit.Service) *Service {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 2
	}
	var auditor *audit.Service
	if len(auditors) > 0 {
		auditor = auditors[0]
	}
	return &Service{
		repo:    repo,
		exec:    exec,
		config:  cfg,
		auditor: auditor,
		nowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *Service) CreatePendingDecision(ctx context.Context, profile aggregation.AggregatedProfile) error {
	if s.repo == nil {
		return nil
	}
	input := NewDecisionInput(profile, DecisionServiceMetadata{
		TargetSystem: "kesmi",
		DeliveryMode: "placeholder",
		Message:      PlaceholderMessage,
	})
	_, err := s.repo.CreatePendingSnapshot(ctx, profile, input, s.config.MaxAttempts)
	return err
}

func (s *Service) DeliverPending(ctx context.Context, snapshot Snapshot) error {
	if s.repo == nil || s.exec == nil {
		return nil
	}

	input, err := s.repo.LoadDecisionInput(ctx, snapshot.ExaminationID)
	if err != nil {
		return err
	}
	requestPayload, err := json.Marshal(input)
	if err != nil {
		return err
	}

	attemptNumber := snapshot.AttemptCount + 1
	startedAt := s.nowFunc()
	correlationID := buildCorrelationID(snapshot.ExaminationID, attemptNumber)

	response, execErr := s.exec.Execute(ctx, input, correlationID)
	finishedAt := s.nowFunc()
	if execErr != nil {
		response = ExecutionResponse{
			State:        DecisionStateTransportExhausted,
			ErrorClass:   "transport",
			ErrorCode:    "execution_error",
			ErrorMessage: execErr.Error(),
			Retryable:    true,
		}
	}

	if err := s.repo.AppendAttempt(ctx, AttemptRecord{
		SnapshotID:      snapshot.ID,
		AttemptNumber:   attemptNumber,
		CorrelationID:   correlationID,
		RequestPayload:  requestPayload,
		ResponsePayload: response.RawResponse,
		ErrorClass:      response.ErrorClass,
		ErrorCode:       response.ErrorCode,
		ErrorMessage:    response.ErrorMessage,
		Retryable:       response.Retryable,
		HTTPStatus:      response.HTTPStatus,
		StartedAt:       startedAt,
		FinishedAt:      finishedAt,
	}); err != nil {
		return err
	}

	diagnostics := diagnosticsFromExecution(response)

	if response.State == "" {
		finalize := FinalizeInput{
			SnapshotID:     snapshot.ID,
			ExaminationID:  snapshot.ExaminationID,
			State:          DecisionStateSucceeded,
			Recommendation: DecisionRecommendationUnavailable,
			Message:        PlaceholderMessage,
			CorrelationID:  correlationID,
			AttemptCount:   attemptNumber,
			LastAttemptAt:  finishedAt,
			CompletedAt:    finishedAt,
			Diagnostics:    diagnostics,
			RawResponse:    response.RawResponse,
		}
		if err := s.repo.MarkSucceeded(ctx, finalize); err != nil {
			return err
		}
		s.appendAudit(ctx, snapshot, finalize)
		return nil
	}

	if response.Retryable && attemptNumber < snapshot.MaxAttempts {
		return s.repo.MarkRetryPending(ctx, RetryPendingInput{
			SnapshotID:    snapshot.ID,
			CorrelationID: correlationID,
			AttemptCount:  attemptNumber,
			LastAttemptAt: finishedAt,
			Diagnostics:   diagnostics,
		})
	}

	state := DecisionStateBusinessError
	if response.Retryable {
		state = DecisionStateTransportExhausted
	}

	finalize := FinalizeInput{
		SnapshotID:     snapshot.ID,
		ExaminationID:  snapshot.ExaminationID,
		State:          state,
		Recommendation: DecisionRecommendationUnavailable,
		Message:        PlaceholderMessage,
		CorrelationID:  correlationID,
		AttemptCount:   attemptNumber,
		LastAttemptAt:  finishedAt,
		CompletedAt:    finishedAt,
		Diagnostics:    diagnostics,
		RawResponse:    response.RawResponse,
	}
	if err := s.repo.MarkFailed(ctx, finalize); err != nil {
		return err
	}
	s.appendAudit(ctx, snapshot, finalize)
	return nil
}

func buildCorrelationID(examinationID int64, attempt int32) string {
	return fmt.Sprintf("exam-%d-kesmi-%d", examinationID, attempt)
}

func diagnosticsFromExecution(response ExecutionResponse) DecisionDiagnostics {
	var errorClass *string
	if response.ErrorClass != "" {
		value := response.ErrorClass
		errorClass = &value
	}
	var errorCode *string
	if response.ErrorCode != "" {
		value := response.ErrorCode
		errorCode = &value
	}
	var errorMessage *string
	if response.ErrorMessage != "" {
		value := response.ErrorMessage
		errorMessage = &value
	}
	var httpStatus *int
	if response.HTTPStatus > 0 {
		value := response.HTTPStatus
		httpStatus = &value
	}
	return DecisionDiagnostics{
		ErrorClass:   errorClass,
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		HTTPStatus:   httpStatus,
		Retryable:    response.Retryable,
	}
}

func (s *Service) appendAudit(ctx context.Context, snapshot Snapshot, finalize FinalizeInput) {
	if s.auditor == nil {
		return
	}
	eventType := audit.EventTypeDecisionCompleted
	eventOutcome := audit.OutcomeSucceeded
	if finalize.State != DecisionStateSucceeded {
		eventType = audit.EventTypeDecisionFailed
		eventOutcome = audit.OutcomeFailed
	}
	channel := "kesmi"
	_ = s.auditor.AppendFromContext(ctx, audit.Event{
		Type:          eventType,
		Key:           fmt.Sprintf("decision-terminal:%d:%s:%d", snapshot.ExaminationID, finalize.State, finalize.AttemptCount),
		Outcome:       eventOutcome,
		CorrelationID: finalize.CorrelationID,
		Resource: audit.ResourceRef{
			Kind: "decision_snapshot",
			ID:   snapshot.ID,
		},
		DomainRefs: audit.DomainRefs{
			ExaminationID:      &snapshot.ExaminationID,
			SpecialistID:       &snapshot.SpecialistID,
			Channel:            &channel,
			DecisionSnapshotID: &snapshot.ID,
		},
		Payload: mustDecisionJSON(map[string]any{
			"state":          finalize.State,
			"attempt_count":  finalize.AttemptCount,
			"recommendation": finalize.Recommendation,
		}),
	})
}

func mustDecisionJSON(payload map[string]any) json.RawMessage {
	data, _ := json.Marshal(payload)
	return data
}
