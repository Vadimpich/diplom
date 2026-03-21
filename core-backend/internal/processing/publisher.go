package processing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	outboxStatusPending   = "pending"
	outboxStatusPublished = "published"
	outboxStatusFailed    = "failed"
	commandExchangeName   = "processing.commands"
	commandDLXName        = "processing.commands.dlx"
)

var queueNamesByChannel = map[string]string{
	ChannelText:           "qq.processing.text",
	ChannelAcoustic:       "qq.processing.acoustic",
	ChannelParalinguistic: "qq.processing.paralinguistic",
}

type RelayConfig struct {
	BrokerURL    string
	PollInterval time.Duration
	MaxAttempts  int32
}

type OutboxMessage struct {
	ID                int64
	ExaminationID     int64
	ChannelRunID      int64
	Channel           string
	ExchangeName      string
	RoutingKey        string
	AttemptCount      int32
	MaxAttempts       int32
	MessageVersion    int
	Payload           []byte
	BrokerCorrelation string
}

type PublishedOutboxUpdate struct {
	OutboxID          int64
	ExaminationID     int64
	ChannelRunID      int64
	BrokerMessageID   string
	BrokerCorrelation string
	PublishedAt       time.Time
	AttemptCount      int32
}

type FailedOutboxUpdate struct {
	OutboxID     int64
	AttemptCount int32
	Status       string
	ErrorCode    string
	ErrorMessage string
}

type RelayRepository interface {
	ListPendingOutbox(context.Context, int32) ([]OutboxMessage, error)
	MarkOutboxPublished(context.Context, PublishedOutboxUpdate) error
	MarkOutboxFailed(context.Context, FailedOutboxUpdate) error
}

type Session interface {
	DeclareExchange(context.Context, string, string) error
	DeclareQueue(context.Context, string, map[string]any) error
	BindQueue(context.Context, string, string, string) error
	Publish(context.Context, string, string, []byte) (bool, error)
}

type Relay struct {
	repo      RelayRepository
	config    RelayConfig
	logger    *log.Logger
	sessionFn func(context.Context) (Session, func() error, error)
}

func NewRelay(repo RelayRepository, cfg RelayConfig) *Relay {
	return &Relay{
		repo:   repo,
		config: cfg,
		logger: log.Default(),
	}
}

func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.config.PollInterval)
	defer ticker.Stop()

	for {
		if err := r.runOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
			r.logger.Printf("processing relay error=%v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r *Relay) PublishPending(ctx context.Context, session Session) error {
	if err := declareCommandTopology(ctx, session, r.config.MaxAttempts); err != nil {
		return err
	}

	messages, err := r.repo.ListPendingOutbox(ctx, r.config.MaxAttempts)
	if err != nil {
		return err
	}

	for _, msg := range messages {
		if err := r.publishOne(ctx, session, msg); err != nil {
			return err
		}
	}
	return nil
}

func (r *Relay) runOnce(ctx context.Context) error {
	if r.sessionFn == nil {
		r.sessionFn = func(ctx context.Context) (Session, func() error, error) {
			return openAMQPSession(ctx, r.config.BrokerURL)
		}
	}

	session, closeFn, err := r.sessionFn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if closeFn != nil {
			_ = closeFn()
		}
	}()

	return r.PublishPending(ctx, session)
}

func (r *Relay) publishOne(ctx context.Context, session Session, msg OutboxMessage) error {
	command := ProcessingCommandEnvelope{}
	if err := json.Unmarshal(msg.Payload, &command); err != nil {
		return r.repo.MarkOutboxFailed(ctx, FailedOutboxUpdate{
			OutboxID:     msg.ID,
			AttemptCount: msg.AttemptCount + 1,
			Status:       outboxStatusFailed,
			ErrorCode:    "invalid_payload",
			ErrorMessage: err.Error(),
		})
	}

	confirmed, err := session.Publish(ctx, msg.ExchangeName, msg.RoutingKey, msg.Payload)
	if err != nil {
		return r.repo.MarkOutboxFailed(ctx, buildFailureUpdate(msg, err))
	}
	if !confirmed {
		return r.repo.MarkOutboxFailed(ctx, buildFailureUpdate(msg, errors.New("publisher confirm not acknowledged")))
	}

	return r.repo.MarkOutboxPublished(ctx, PublishedOutboxUpdate{
		OutboxID:          msg.ID,
		ExaminationID:     msg.ExaminationID,
		ChannelRunID:      msg.ChannelRunID,
		BrokerMessageID:   command.MessageID,
		BrokerCorrelation: command.CorrelationID,
		PublishedAt:       time.Now().UTC(),
		AttemptCount:      msg.AttemptCount + 1,
	})
}

func buildFailureUpdate(msg OutboxMessage, err error) FailedOutboxUpdate {
	status := outboxStatusPending
	errorCode := "temporary_publish_error"
	attemptCount := msg.AttemptCount + 1
	if attemptCount >= msg.MaxAttempts {
		status = outboxStatusFailed
	}

	var fatal FatalPublishError
	if errors.As(err, &fatal) {
		status = outboxStatusFailed
		errorCode = "fatal_publish_error"
		err = fatal.Err
	}

	return FailedOutboxUpdate{
		OutboxID:     msg.ID,
		AttemptCount: attemptCount,
		Status:       status,
		ErrorCode:    errorCode,
		ErrorMessage: err.Error(),
	}
}

type FatalPublishError struct {
	Err error
}

func (e FatalPublishError) Error() string {
	return e.Err.Error()
}

func declareCommandTopology(ctx context.Context, session Session, maxAttempts int32) error {
	if err := session.DeclareExchange(ctx, commandExchangeName, "topic"); err != nil {
		return err
	}
	if err := session.DeclareExchange(ctx, commandDLXName, "topic"); err != nil {
		return err
	}

	for channel, queue := range queueNamesByChannel {
		args := map[string]any{
			"x-queue-type":           "quorum",
			"x-dead-letter-exchange": commandDLXName,
			"x-delivery-limit":       maxAttempts,
		}
		if err := session.DeclareQueue(ctx, queue, args); err != nil {
			return err
		}
		if err := session.BindQueue(ctx, queue, routingKeyForChannel(channel), commandExchangeName); err != nil {
			return err
		}
	}
	return nil
}

type amqpSession struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	confirms   <-chan amqp.Confirmation
}

func openAMQPSession(ctx context.Context, brokerURL string) (Session, func() error, error) {
	if brokerURL == "" {
		return nil, nil, fmt.Errorf("RABBITMQ_URL is required")
	}
	conn, err := amqp.Dial(brokerURL)
	if err != nil {
		return nil, nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}
	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, err
	}
	session := &amqpSession{
		connection: conn,
		channel:    ch,
		confirms:   ch.NotifyPublish(make(chan amqp.Confirmation, 1)),
	}
	return session, func() error {
		if err := ch.Close(); err != nil {
			_ = conn.Close()
			return err
		}
		return conn.Close()
	}, nil
}

func (s *amqpSession) DeclareExchange(_ context.Context, name, kind string) error {
	return s.channel.ExchangeDeclare(name, kind, true, false, false, false, nil)
}

func (s *amqpSession) DeclareQueue(_ context.Context, name string, args map[string]any) error {
	_, err := s.channel.QueueDeclare(name, true, false, false, false, amqp.Table(args))
	return err
}

func (s *amqpSession) BindQueue(_ context.Context, name, routingKey, exchange string) error {
	return s.channel.QueueBind(name, routingKey, exchange, false, nil)
}

func (s *amqpSession) Publish(ctx context.Context, exchange, routingKey string, body []byte) (bool, error) {
	if err := s.channel.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	}); err != nil {
		return false, err
	}

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case confirm := <-s.confirms:
		return confirm.Ack, nil
	}
}
