package processing

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestOutboxRelayPublishesPendingMessages(t *testing.T) {
	command := ProcessingCommandEnvelope{
		MessageVersion: MessageVersionV1,
		MessageID:      "msg-text-1",
		CorrelationID:  "exam-100-text-v1",
		ExaminationID:  100,
		SpecialistID:   10,
		Channel:        ChannelText,
		Attempt:        1,
		MaxAttempts:    3,
		RequestedAt:    time.Unix(1_742_550_000, 0).UTC(),
		Answers: []CommandAnswerReference{{
			AnswerID:      1,
			QuestionID:    10,
			AudioS3Bucket: "diplom-audio",
			AudioS3Key:    "examinations/100/answers/1/audio.webm",
			AnswerText:    "Ответ",
		}},
	}
	payload, err := json.Marshal(command)
	if err != nil {
		t.Fatalf("marshal command: %v", err)
	}

	repo := &relayRepositoryStub{
		pending: []OutboxMessage{
			{
				ID:                1,
				ExaminationID:     100,
				ChannelRunID:      200,
				Channel:           ChannelText,
				ExchangeName:      commandsExchange,
				RoutingKey:        "processing.command.text",
				AttemptCount:      0,
				MaxAttempts:       3,
				MessageVersion:    MessageVersionV1,
				Payload:           payload,
				BrokerCorrelation: command.CorrelationID,
			},
		},
	}
	session := &sessionStub{confirm: true}
	relay := NewRelay(repo, RelayConfig{
		PollInterval: 100 * time.Millisecond,
		MaxAttempts:  3,
	})

	if err := relay.PublishPending(context.Background(), session); err != nil {
		t.Fatalf("publish pending: %v", err)
	}

	if len(repo.published) != 1 {
		t.Fatalf("expected 1 published row, got %d", len(repo.published))
	}
	if len(session.published) != 1 {
		t.Fatalf("expected broker publish, got %d calls", len(session.published))
	}
	if session.published[0].routingKey != "processing.command.text" {
		t.Fatalf("expected channel-specific routing key, got %s", session.published[0].routingKey)
	}
	queue := session.queues["qq.processing.text"]
	if queue.name == "" {
		t.Fatal("expected text queue to be declared")
	}
	if got := queue.args["x-queue-type"]; got != "quorum" {
		t.Fatalf("expected quorum queue, got %#v", got)
	}
	if got := queue.args["x-dead-letter-exchange"]; got != "processing.commands.dlx" {
		t.Fatalf("expected explicit DLX, got %#v", got)
	}
	if got := queue.args["x-delivery-limit"]; got != int32(3) {
		t.Fatalf("expected explicit delivery limit, got %#v", got)
	}
}

func TestRetryBudget(t *testing.T) {
	command := ProcessingCommandEnvelope{
		MessageVersion: MessageVersionV1,
		MessageID:      "msg-acoustic-1",
		CorrelationID:  "exam-100-acoustic-v1",
		ExaminationID:  100,
		SpecialistID:   10,
		Channel:        ChannelAcoustic,
		Attempt:        1,
		MaxAttempts:    3,
		RequestedAt:    time.Unix(1_742_550_000, 0).UTC(),
	}
	payload, err := json.Marshal(command)
	if err != nil {
		t.Fatalf("marshal command: %v", err)
	}

	tests := []struct {
		name           string
		attemptCount   int32
		expectedStatus string
	}{
		{name: "retryable", attemptCount: 0, expectedStatus: outboxStatusPending},
		{name: "exhausted", attemptCount: 2, expectedStatus: outboxStatusFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &relayRepositoryStub{
				pending: []OutboxMessage{{
					ID:             2,
					ExaminationID:  100,
					ChannelRunID:   201,
					Channel:        ChannelAcoustic,
					ExchangeName:   commandsExchange,
					RoutingKey:     "processing.command.acoustic",
					AttemptCount:   tt.attemptCount,
					MaxAttempts:    3,
					MessageVersion: MessageVersionV1,
					Payload:        payload,
				}},
			}
			relay := NewRelay(repo, RelayConfig{PollInterval: time.Second, MaxAttempts: 3})
			session := &sessionStub{publishErr: errors.New("broker unavailable")}

			if err := relay.PublishPending(context.Background(), session); err != nil {
				t.Fatalf("publish pending: %v", err)
			}
			if len(repo.failed) != 1 {
				t.Fatalf("expected one failure record, got %d", len(repo.failed))
			}
			if repo.failed[0].Status != tt.expectedStatus {
				t.Fatalf("expected status %s, got %s", tt.expectedStatus, repo.failed[0].Status)
			}
		})
	}
}

func TestFatalVsTemporaryError(t *testing.T) {
	command := ProcessingCommandEnvelope{
		MessageVersion: MessageVersionV1,
		MessageID:      "msg-para-1",
		CorrelationID:  "exam-100-paralinguistic-v1",
		ExaminationID:  100,
		SpecialistID:   10,
		Channel:        ChannelParalinguistic,
		Attempt:        1,
		MaxAttempts:    3,
		RequestedAt:    time.Unix(1_742_550_000, 0).UTC(),
	}
	payload, err := json.Marshal(command)
	if err != nil {
		t.Fatalf("marshal command: %v", err)
	}

	repo := &relayRepositoryStub{
		pending: []OutboxMessage{{
			ID:             3,
			ExaminationID:  100,
			ChannelRunID:   202,
			Channel:        ChannelParalinguistic,
			ExchangeName:   commandsExchange,
			RoutingKey:     "processing.command.paralinguistic",
			AttemptCount:   0,
			MaxAttempts:    3,
			MessageVersion: MessageVersionV1,
			Payload:        payload,
		}},
	}
	relay := NewRelay(repo, RelayConfig{PollInterval: time.Second, MaxAttempts: 3})

	if err := relay.PublishPending(context.Background(), &sessionStub{publishErr: errors.New("temporary")}); err != nil {
		t.Fatalf("temporary publish: %v", err)
	}
	if repo.failed[0].Status != outboxStatusPending {
		t.Fatalf("expected temporary error to remain pending, got %s", repo.failed[0].Status)
	}

	repo.failed = nil
	if err := relay.PublishPending(context.Background(), &sessionStub{publishErr: FatalPublishError{Err: errors.New("invalid payload")}}); err != nil {
		t.Fatalf("fatal publish: %v", err)
	}
	if repo.failed[0].Status != outboxStatusFailed {
		t.Fatalf("expected fatal error to fail row immediately, got %s", repo.failed[0].Status)
	}
}

func TestPublisherPreservesTraceContext(t *testing.T) {
	command := ProcessingCommandEnvelope{
		MessageVersion: MessageVersionV1,
		MessageID:      "msg-text-trace-1",
		CorrelationID:  "exam-100-text-v1",
		RequestID:      "req-100",
		TraceParent:    "00-8ec8c1b6409f4a6cb80cfcb4f74aa98c-5d7c1f97db7840b3-01",
		TraceState:     "tenant=diplom",
		ExaminationID:  100,
		SpecialistID:   10,
		Channel:        ChannelText,
		Attempt:        1,
		MaxAttempts:    3,
		RequestedAt:    time.Unix(1_742_550_000, 0).UTC(),
	}
	payload, err := json.Marshal(command)
	if err != nil {
		t.Fatalf("marshal command: %v", err)
	}

	repo := &relayRepositoryStub{
		pending: []OutboxMessage{{
			ID:                4,
			ExaminationID:     100,
			ChannelRunID:      203,
			Channel:           ChannelText,
			ExchangeName:      commandsExchange,
			RoutingKey:        "processing.command.text",
			AttemptCount:      0,
			MaxAttempts:       3,
			MessageVersion:    MessageVersionV1,
			Payload:           payload,
			BrokerCorrelation: command.CorrelationID,
		}},
	}
	session := &sessionStub{confirm: true}
	relay := NewRelay(repo, RelayConfig{PollInterval: time.Second, MaxAttempts: 3})

	if err := relay.PublishPending(context.Background(), session); err != nil {
		t.Fatalf("publish pending: %v", err)
	}
	if len(session.published) != 1 {
		t.Fatalf("expected one publish, got %d", len(session.published))
	}

	published := ProcessingCommandEnvelope{}
	if err := json.Unmarshal(session.published[0].body, &published); err != nil {
		t.Fatalf("unmarshal published payload: %v", err)
	}
	if published.RequestID != command.RequestID {
		t.Fatalf("expected request_id propagation, got %q", published.RequestID)
	}
	if published.TraceParent != command.TraceParent || published.TraceState != command.TraceState {
		t.Fatalf("expected trace context propagation, got traceparent=%q tracestate=%q", published.TraceParent, published.TraceState)
	}
}

type relayRepositoryStub struct {
	pending   []OutboxMessage
	published []PublishedOutboxUpdate
	failed    []FailedOutboxUpdate
}

func (s *relayRepositoryStub) ListPendingOutbox(context.Context, int32) ([]OutboxMessage, error) {
	return append([]OutboxMessage(nil), s.pending...), nil
}

func (s *relayRepositoryStub) MarkOutboxPublished(_ context.Context, update PublishedOutboxUpdate) error {
	s.published = append(s.published, update)
	return nil
}

func (s *relayRepositoryStub) MarkOutboxFailed(_ context.Context, update FailedOutboxUpdate) error {
	s.failed = append(s.failed, update)
	return nil
}

type sessionStub struct {
	confirm    bool
	publishErr error
	queues     map[string]declaredQueue
	published  []publishedMessage
}

type declaredQueue struct {
	name string
	args map[string]any
}

type publishedMessage struct {
	exchange   string
	routingKey string
	body       []byte
}

func (s *sessionStub) DeclareExchange(_ context.Context, name, kind string) error {
	return nil
}

func (s *sessionStub) DeclareQueue(_ context.Context, name string, args map[string]any) error {
	if s.queues == nil {
		s.queues = make(map[string]declaredQueue)
	}
	s.queues[name] = declaredQueue{name: name, args: args}
	return nil
}

func (s *sessionStub) BindQueue(context.Context, string, string, string) error {
	return nil
}

func (s *sessionStub) Publish(_ context.Context, exchange, routingKey string, body []byte) (bool, error) {
	if s.publishErr != nil {
		return false, s.publishErr
	}
	s.published = append(s.published, publishedMessage{
		exchange:   exchange,
		routingKey: routingKey,
		body:       body,
	})
	return s.confirm, nil
}
