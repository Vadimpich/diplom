package channelresults_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"dimplom/internal/channelresults"
	"dimplom/internal/processing"
)

func TestIndependentChannelCompletion(t *testing.T) {
	repo := newRepositoryStub()
	service := channelresults.NewService(repo, nil)

	err := service.ApplyResult(context.Background(), processing.ChannelResultEnvelope{
		MessageVersion: processing.MessageVersionV1,
		MessageID:      "msg-text-1",
		CorrelationID:  "exam-100-text-v1",
		ExaminationID:  100,
		Channel:        processing.ChannelText,
		Attempt:        1,
		Status:         processing.ResultStatusSucceeded,
		CompletedAt:    time.Unix(1_742_550_000, 0).UTC(),
		ModelVersion:   "text-stub-0.1.0",
		Payload:        json.RawMessage(`{"summary":"ok"}`),
	})
	if err != nil {
		t.Fatalf("apply successful result: %v", err)
	}

	textRun := repo.runs[processing.ChannelText]
	if textRun.Status != channelresults.RunStatusSucceeded {
		t.Fatalf("expected text channel to succeed, got %q", textRun.Status)
	}
	if textRun.AttemptCount != 1 {
		t.Fatalf("expected text attempt_count=1, got %d", textRun.AttemptCount)
	}

	acousticRun := repo.runs[processing.ChannelAcoustic]
	if acousticRun.Status != channelresults.RunStatusQueued {
		t.Fatalf("expected unrelated channel to remain queued, got %q", acousticRun.Status)
	}
	if acousticRun.AttemptCount != 0 {
		t.Fatalf("expected unrelated attempt_count=0, got %d", acousticRun.AttemptCount)
	}
}

func TestRetryBudget(t *testing.T) {
	repo := newRepositoryStub()
	service := channelresults.NewService(repo, nil)

	err := service.ApplyResult(context.Background(), processing.ChannelResultEnvelope{
		MessageVersion: processing.MessageVersionV1,
		MessageID:      "msg-text-1",
		CorrelationID:  "exam-100-text-v1",
		ExaminationID:  100,
		Channel:        processing.ChannelText,
		Attempt:        1,
		Status:         processing.ResultStatusTemporaryError,
		CompletedAt:    time.Unix(1_742_550_000, 0).UTC(),
		ModelVersion:   "text-stub-0.1.0",
		ErrorCode:      stringPtr("s3_unavailable"),
		ErrorMessage:   stringPtr("temporary outage"),
		Payload:        json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatalf("apply temporary error result: %v", err)
	}

	run := repo.runs[processing.ChannelText]
	if run.Status != channelresults.RunStatusRetryScheduled {
		t.Fatalf("expected retry_scheduled, got %q", run.Status)
	}
	if run.AttemptCount != 1 {
		t.Fatalf("expected attempt_count=1, got %d", run.AttemptCount)
	}
	if got := deref(run.LastErrorCode); got != "s3_unavailable" {
		t.Fatalf("expected error code to persist, got %q", got)
	}
	if repo.examinationStatus != "processing" {
		t.Fatalf("expected examination to stay processing, got %q", repo.examinationStatus)
	}
}

func TestFatalVsTemporaryError(t *testing.T) {
	repo := newRepositoryStub()
	service := channelresults.NewService(repo, nil)

	err := service.ApplyResult(context.Background(), processing.ChannelResultEnvelope{
		MessageVersion: processing.MessageVersionV1,
		MessageID:      "msg-text-2",
		CorrelationID:  "exam-100-text-v1",
		ExaminationID:  100,
		Channel:        processing.ChannelText,
		Attempt:        2,
		Status:         processing.ResultStatusFatalError,
		CompletedAt:    time.Unix(1_742_550_020, 0).UTC(),
		ModelVersion:   "text-stub-0.1.0",
		ErrorCode:      stringPtr("invalid_payload"),
		ErrorMessage:   stringPtr("unsupported document"),
		Payload:        json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatalf("apply fatal error result: %v", err)
	}

	run := repo.runs[processing.ChannelText]
	if run.Status != channelresults.RunStatusFailedFatal {
		t.Fatalf("expected failed_fatal, got %q", run.Status)
	}
	if got := deref(run.LastErrorMessage); got != "unsupported document" {
		t.Fatalf("expected fatal error message to persist, got %q", got)
	}
	if len(repo.savedResults) != 1 {
		t.Fatalf("expected normalized result row to be persisted, got %d rows", len(repo.savedResults))
	}
}

func TestMandatoryChannelExhaustionFailsExamination(t *testing.T) {
	repo := newRepositoryStub()
	repo.runs[processing.ChannelText] = channelresults.ChannelRun{
		ExaminationID: 100,
		Channel:       processing.ChannelText,
		Status:        channelresults.RunStatusRetryScheduled,
		AttemptCount:  2,
		MaxAttempts:   3,
	}
	service := channelresults.NewService(repo, nil)

	err := service.ApplyResult(context.Background(), processing.ChannelResultEnvelope{
		MessageVersion: processing.MessageVersionV1,
		MessageID:      "msg-text-3",
		CorrelationID:  "exam-100-text-v1",
		ExaminationID:  100,
		Channel:        processing.ChannelText,
		Attempt:        3,
		Status:         processing.ResultStatusTemporaryError,
		CompletedAt:    time.Unix(1_742_550_040, 0).UTC(),
		ModelVersion:   "text-stub-0.1.0",
		ErrorCode:      stringPtr("timeout"),
		ErrorMessage:   stringPtr("retry budget exhausted"),
		Payload:        json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatalf("apply exhausted retry result: %v", err)
	}

	run := repo.runs[processing.ChannelText]
	if run.Status != channelresults.RunStatusExhausted {
		t.Fatalf("expected exhausted, got %q", run.Status)
	}
	if repo.examinationStatus != channelresults.ExaminationStatusFailed {
		t.Fatalf("expected examination to fail, got %q", repo.examinationStatus)
	}
	if repo.failedAt == nil {
		t.Fatal("expected failed_at to be recorded")
	}
}

type repositoryStub struct {
	runs              map[string]channelresults.ChannelRun
	examinationStatus string
	failedAt          *time.Time
	savedResults      []channelresults.StoredResult
}

func newRepositoryStub() *repositoryStub {
	return &repositoryStub{
		runs: map[string]channelresults.ChannelRun{
			processing.ChannelText: {
				ExaminationID: 100,
				Channel:       processing.ChannelText,
				Status:        channelresults.RunStatusQueued,
				AttemptCount:  0,
				MaxAttempts:   3,
			},
			processing.ChannelAcoustic: {
				ExaminationID: 100,
				Channel:       processing.ChannelAcoustic,
				Status:        channelresults.RunStatusQueued,
				AttemptCount:  0,
				MaxAttempts:   3,
			},
			processing.ChannelParalinguistic: {
				ExaminationID: 100,
				Channel:       processing.ChannelParalinguistic,
				Status:        channelresults.RunStatusQueued,
				AttemptCount:  0,
				MaxAttempts:   3,
			},
		},
		examinationStatus: "processing",
		savedResults:      make([]channelresults.StoredResult, 0, 1),
	}
}

func (r *repositoryStub) GetChannelRun(_ context.Context, examinationID int64, channel string) (channelresults.ChannelRun, error) {
	run := r.runs[channel]
	run.ExaminationID = examinationID
	return run, nil
}

func (r *repositoryStub) SaveResult(_ context.Context, result channelresults.StoredResult) error {
	r.savedResults = append(r.savedResults, result)
	return nil
}

func (r *repositoryStub) UpdateChannelRun(_ context.Context, run channelresults.ChannelRun) error {
	r.runs[run.Channel] = run
	return nil
}

func (r *repositoryStub) MarkExaminationFailed(_ context.Context, examinationID int64, failedAt time.Time) error {
	r.examinationStatus = channelresults.ExaminationStatusFailed
	r.failedAt = &failedAt
	return nil
}

func stringPtr(value string) *string {
	return &value
}

func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
