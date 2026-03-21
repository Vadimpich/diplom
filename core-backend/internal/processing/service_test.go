package processing_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"dimplom/internal/examinations"
	"dimplom/internal/processing"
)

func TestFinishCreatesOutboxForMandatoryChannels(t *testing.T) {
	repo := &finishRepoStub{
		finishResult: examinations.Examination{
			ID:           100,
			SpecialistID: 10,
			Status:       examinations.StatusReadyForProcessing,
		},
	}
	service := examinations.NewService(repo)

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
	outboxCommands []processing.ProcessingCommandEnvelope
}

func (s *finishRepoStub) Create(context.Context, examinations.CreateInput) (examinations.Examination, error) {
	return examinations.Examination{}, nil
}

func (s *finishRepoStub) List(context.Context) ([]examinations.Examination, error) {
	return nil, nil
}

func (s *finishRepoStub) GetByID(context.Context, int64) (examinations.Examination, error) {
	return examinations.Examination{}, nil
}

func (s *finishRepoStub) ListBySpecialistID(context.Context, int64) ([]examinations.Examination, error) {
	return nil, nil
}

func (s *finishRepoStub) UpdateStatus(context.Context, int64, string) (examinations.Examination, error) {
	return examinations.Examination{}, nil
}

func (s *finishRepoStub) Finish(context.Context, int64) (examinations.Examination, error) {
	return s.finishResult, s.finishErr
}
