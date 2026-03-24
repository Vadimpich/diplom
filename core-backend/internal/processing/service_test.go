package processing_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"diplom/internal/audit"
	"diplom/internal/examinations"
	"diplom/internal/processing"
)

func TestFinishCreatesOutboxForMandatoryChannels(t *testing.T) {
	repo := &finishRepoStub{
		finishResult: examinations.Examination{
			ID:           100,
			SpecialistID: 10,
			Status:       examinations.StatusReadyForProcessing,
		},
	}
	service := processing.NewService(repo)

	exam, err := service.Finish(context.Background(), 100)
	if err != nil {
		t.Fatalf("finish examination: %v", err)
	}
	if exam.Status != examinations.StatusReadyForProcessing {
		t.Fatalf("expected status %q, got %q", examinations.StatusReadyForProcessing, exam.Status)
	}

	if len(repo.outboxCommands) != len(processing.MandatoryChannels) {
		t.Fatalf("finish must create one pending command per mandatory channel: want %d, got %d", len(processing.MandatoryChannels), len(repo.outboxCommands))
	}

	for _, command := range repo.outboxCommands {
		if command.MessageVersion != processing.MessageVersionV1 {
			t.Fatalf("expected message_version=%d, got %d", processing.MessageVersionV1, command.MessageVersion)
		}
		if command.Channel == "" {
			t.Fatal("expected channel to be set on every pending command")
		}
		if len(command.Answers) == 0 {
			t.Fatal("expected pending command to carry S3 answer references")
		}
	}
}

func TestFinishIsIdempotentAndDoesNotDuplicatePendingWork(t *testing.T) {
	repo := &finishRepoStub{
		finishResult: examinations.Examination{
			ID:           100,
			SpecialistID: 10,
			Status:       examinations.StatusReadyForProcessing,
		},
	}
	service := processing.NewService(repo)

	first, err := service.Finish(context.Background(), 100)
	if err != nil {
		t.Fatalf("first finish: %v", err)
	}
	second, err := service.Finish(context.Background(), 100)
	if err != nil {
		t.Fatalf("second finish: %v", err)
	}

	if first.Status != examinations.StatusReadyForProcessing || second.Status != examinations.StatusReadyForProcessing {
		t.Fatal("expected repeated finish to remain idempotent")
	}
	if len(repo.outboxCommands) != len(processing.MandatoryChannels) {
		t.Fatalf("expected exactly %d persisted commands after repeated finish, got %d", len(processing.MandatoryChannels), len(repo.outboxCommands))
	}
	if repo.finishCalls != 2 {
		t.Fatalf("expected underlying finish flow to run on both attempts, got %d calls", repo.finishCalls)
	}
}

func TestPendingCommandsCarryIdentifiersAndS3References(t *testing.T) {
	repo := &finishRepoStub{
		finishResult: examinations.Examination{
			ID:           200,
			SpecialistID: 20,
			Status:       examinations.StatusReadyForProcessing,
		},
	}
	service := processing.NewService(repo)

	if _, err := service.Finish(context.Background(), 200); err != nil {
		t.Fatalf("finish examination: %v", err)
	}

	for _, command := range repo.outboxCommands {
		if command.ExaminationID != 200 {
			t.Fatalf("expected examination_id=200, got %d", command.ExaminationID)
		}
		if command.SpecialistID != 20 {
			t.Fatalf("expected specialist_id=20, got %d", command.SpecialistID)
		}
		if command.CorrelationID == "" {
			t.Fatal("expected correlation_id for retry-safe publication")
		}
		for _, answer := range command.Answers {
			if answer.AudioS3Bucket == "" || answer.AudioS3Key == "" {
				t.Fatal("expected answer references to include S3 bucket and key")
			}
			if answer.QuestionID == 0 || answer.AnswerID == 0 {
				t.Fatal("expected answer references to include identifiers")
			}
		}
	}
}

func TestChannelResultEnvelopeIsChannelNeutral(t *testing.T) {
	timestamp := time.Unix(1_742_550_000, 0).UTC()

	for _, channel := range processing.MandatoryChannels {
		channel := channel
		t.Run(channel, func(t *testing.T) {
			payload := processing.ChannelResultEnvelope{
				MessageVersion: processing.MessageVersionV1,
				MessageID:      "msg-1",
				CorrelationID:  "exam-100-" + channel + "-v1",
				ExaminationID:  100,
				Channel:        channel,
				Attempt:        1,
				Status:         processing.ResultStatusSucceeded,
				CompletedAt:    timestamp,
				ModelVersion:   channel + "-stub-0.1.0",
				Payload:        json.RawMessage(`{"summary":"ok"}`),
			}

			data, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("marshal result envelope: %v", err)
			}

			var decoded map[string]any
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal result envelope: %v", err)
			}

			if decoded["channel"] != channel {
				t.Fatalf("expected channel %q, got %#v", channel, decoded["channel"])
			}
			if _, ok := decoded["payload"]; !ok {
				t.Fatal("expected shared result envelope to always include payload field")
			}
		})
	}
}

type finishRepoStub struct {
	finishResult   examinations.Examination
	finishErr      error
	finishCalls    int
	outboxCommands []processing.ProcessingCommandEnvelope
}

func (s *finishRepoStub) FinishLaunch(context.Context, int64) (examinations.Examination, []processing.ProcessingCommandEnvelope, error) {
	s.finishCalls++
	if s.finishErr != nil {
		return examinations.Examination{}, nil, s.finishErr
	}
	if len(s.outboxCommands) == 0 {
		commands := make([]processing.ProcessingCommandEnvelope, 0, len(processing.MandatoryChannels))
		for idx, channel := range processing.MandatoryChannels {
			commands = append(commands, processing.ProcessingCommandEnvelope{
				MessageVersion: processing.MessageVersionV1,
				MessageID:      channel + "-msg",
				CorrelationID:  "exam-200-" + channel + "-v1",
				ExaminationID:  s.finishResult.ID,
				SpecialistID:   s.finishResult.SpecialistID,
				Channel:        channel,
				Attempt:        1,
				MaxAttempts:    3,
				RequestedAt:    time.Unix(1_742_550_000+int64(idx), 0).UTC(),
				Answers: []processing.CommandAnswerReference{
					{
						AnswerID:      int64(idx + 1),
						QuestionID:    int64(idx + 10),
						AudioS3Bucket: "diplom-audio",
						AudioS3Key:    "examinations/200/answers/1/audio.webm",
						AnswerText:    "Ответ обследуемого",
					},
				},
			})
		}
		s.outboxCommands = commands
	}
	return s.finishResult, s.outboxCommands, nil
}

func (s *finishRepoStub) GetProcessingStatus(context.Context, int64) (processing.ProcessingStatusResponse, error) {
	return processing.ProcessingStatusResponse{}, nil
}

func TestFinishPropagatesRepositoryError(t *testing.T) {
	expectedErr := errors.New("boom")
	service := processing.NewService(&finishRepoStub{finishErr: expectedErr})

	if _, err := service.Finish(context.Background(), 100); !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestProcessingLaunchWritesAuditEvent(t *testing.T) {
	repo := &finishRepoStub{
		finishResult: examinations.Examination{
			ID:           100,
			SpecialistID: 10,
			Status:       examinations.StatusReadyForProcessing,
		},
	}
	auditRepo := &processingAuditRepoStub{}
	service := processing.NewService(repo, audit.NewService(auditRepo))

	if _, err := service.Finish(context.Background(), 100); err != nil {
		t.Fatalf("finish examination: %v", err)
	}
	if len(auditRepo.events) != 1 {
		t.Fatalf("expected one audit event, got %d", len(auditRepo.events))
	}
	if auditRepo.events[0].Type != audit.EventTypeProcessingLaunch {
		t.Fatalf("expected processing.launch event, got %q", auditRepo.events[0].Type)
	}
}

type processingAuditRepoStub struct {
	events []audit.Event
}

func (s *processingAuditRepoStub) Append(_ context.Context, event audit.Event) error {
	s.events = append(s.events, event)
	return nil
}

func (s *processingAuditRepoStub) List(context.Context, audit.ListFilter) ([]audit.Event, error) {
	return append([]audit.Event(nil), s.events...), nil
}
