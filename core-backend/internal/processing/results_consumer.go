package processing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"diplom/internal/audit"
	"diplom/internal/observability"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	resultExchangeName = "processing.results"
	resultQueueName    = "qq.processing.results"
	resultRoutingKey   = "processing.result"
)

type ResultHandler interface {
	HandleResult(context.Context, ChannelResultEnvelope) error
}

type ResultsConsumer struct {
	brokerURL string
	handler   ResultHandler
	logger    *log.Logger
}

func NewResultsConsumer(brokerURL string, handler ResultHandler) *ResultsConsumer {
	return &ResultsConsumer{
		brokerURL: brokerURL,
		handler:   handler,
		logger:    log.Default(),
	}
}

func (c *ResultsConsumer) Run(ctx context.Context) {
	if c == nil || c.handler == nil {
		return
	}

	for {
		if err := c.consumeOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
			c.logger.Printf("processing results consumer error=%v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func (c *ResultsConsumer) consumeOnce(ctx context.Context) error {
	if c.brokerURL == "" {
		return fmt.Errorf("RABBITMQ_URL is required")
	}

	conn, err := amqp.Dial(c.brokerURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(resultExchangeName, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(resultQueueName, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(resultQueueName, resultRoutingKey, resultExchangeName, false, nil); err != nil {
		return err
	}
	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}

	deliveries, err := ch.Consume(resultQueueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return nil
			}
			if err := c.handleDelivery(ctx, delivery); err != nil {
				c.logger.Printf("processing results delivery failed message_id=%s err=%v", delivery.MessageId, err)
			}
		}
	}
}

func (c *ResultsConsumer) handleDelivery(ctx context.Context, delivery amqp.Delivery) error {
	var envelope ChannelResultEnvelope
	if err := json.Unmarshal(delivery.Body, &envelope); err != nil {
		_ = delivery.Ack(false)
		return fmt.Errorf("decode envelope: %w", err)
	}
	if err := validateResultEnvelope(envelope); err != nil {
		_ = delivery.Ack(false)
		return err
	}
	trace := observability.BuildTraceContext(envelope.RequestID, envelope.TraceParent, envelope.TraceState)
	ctx = observability.WithTraceContext(ctx, trace)
	meta := audit.MetadataFromContext(ctx)
	meta.RequestID = trace.RequestID
	meta.TraceID = trace.TraceID
	meta.TraceParent = trace.TraceParent
	meta.TraceState = trace.TraceState
	meta.CorrelationID = envelope.CorrelationID
	ctx = audit.WithMetadata(ctx, meta)
	if err := c.handler.HandleResult(ctx, envelope); err != nil {
		_ = delivery.Nack(false, true)
		return err
	}
	return delivery.Ack(false)
}

func validateResultEnvelope(envelope ChannelResultEnvelope) error {
	if envelope.MessageVersion != MessageVersionV1 {
		return fmt.Errorf("unsupported message_version %d", envelope.MessageVersion)
	}
	if envelope.ExaminationID <= 0 {
		return fmt.Errorf("examination_id is required")
	}
	switch envelope.Channel {
	case ChannelText, ChannelAcoustic, ChannelParalinguistic:
	default:
		return fmt.Errorf("unsupported channel %q", envelope.Channel)
	}
	switch envelope.Status {
	case ResultStatusSucceeded, ResultStatusTemporaryError, ResultStatusFatalError:
	default:
		return fmt.Errorf("unsupported result status %q", envelope.Status)
	}
	if envelope.Attempt <= 0 {
		return fmt.Errorf("attempt must be positive")
	}
	if envelope.ModelVersion == "" {
		return fmt.Errorf("model_version is required")
	}
	if envelope.CompletedAt.IsZero() {
		return fmt.Errorf("completed_at is required")
	}
	if envelope.Status != ResultStatusSucceeded && (envelope.ErrorCode == nil || envelope.ErrorMessage == nil) {
		return fmt.Errorf("error_code and error_message are required for failed results")
	}
	return nil
}
