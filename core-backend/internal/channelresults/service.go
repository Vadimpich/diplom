package channelresults

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"dimplom/internal/processing"
	"dimplom/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	RunStatusQueued         = "queued"
	RunStatusRetryScheduled = "retry_scheduled"
	RunStatusSucceeded      = "succeeded"
	RunStatusFailedFatal    = "failed_fatal"
	RunStatusExhausted      = "exhausted"
	ExaminationStatusFailed = "failed"
)

type ChannelRun struct {
	ExaminationID    int64
	ChannelRunID     int64
	Channel          string
	Status           string
	AttemptCount     int32
	MaxAttempts      int32
	MessageVersion   int32
	LastErrorCode    *string
	LastErrorMessage *string
	FinishedAt       *time.Time
}

type StoredResult struct {
	ExaminationID       int64
	ChannelRunID        int64
	Channel             string
	MessageVersion      int
	Attempt             int32
	Status              string
	ModelVersion        string
	Payload             json.RawMessage
	ErrorCode           *string
	ErrorMessage        *string
	BrokerMessageID     string
	BrokerCorrelationID string
	CompletedAt         time.Time
}

type Repository interface {
	GetChannelRun(context.Context, int64, string) (ChannelRun, error)
	SaveResult(context.Context, StoredResult) error
	UpdateChannelRun(context.Context, ChannelRun) error
	MarkExaminationFailed(context.Context, int64, time.Time) error
}

type AggregationTrigger interface {
	TryAggregate(context.Context, int64) error
}

type Service struct {
	repo    Repository
	trigger AggregationTrigger
}

func NewService(repo Repository, trigger AggregationTrigger) *Service {
	return &Service{repo: repo, trigger: trigger}
}

type Handler struct {
	repo    *SQLRepository
	service *Service
}

func NewHandler(repo *SQLRepository, trigger AggregationTrigger) *Handler {
	service := NewService(repo, trigger)
	return &Handler{
		repo:    repo,
		service: service,
	}
}

func (s *Service) ApplyResult(ctx context.Context, envelope processing.ChannelResultEnvelope) error {
	run, err := s.repo.GetChannelRun(ctx, envelope.ExaminationID, envelope.Channel)
	if err != nil {
		return err
	}

	stored := StoredResult{
		ExaminationID:       envelope.ExaminationID,
		ChannelRunID:        run.ChannelRunID,
		Channel:             envelope.Channel,
		MessageVersion:      envelope.MessageVersion,
		Attempt:             envelope.Attempt,
		Status:              envelope.Status,
		ModelVersion:        envelope.ModelVersion,
		Payload:             envelope.Payload,
		ErrorCode:           envelope.ErrorCode,
		ErrorMessage:        envelope.ErrorMessage,
		BrokerMessageID:     envelope.MessageID,
		BrokerCorrelationID: envelope.CorrelationID,
		CompletedAt:         envelope.CompletedAt.UTC(),
	}
	if err := s.repo.SaveResult(ctx, stored); err != nil {
		return err
	}

	run.AttemptCount = envelope.Attempt
	run.MessageVersion = int32(envelope.MessageVersion)
	run.LastErrorCode = envelope.ErrorCode
	run.LastErrorMessage = envelope.ErrorMessage
	finishedAt := envelope.CompletedAt.UTC()
	run.FinishedAt = &finishedAt

	terminalFailure := false
	switch envelope.Status {
	case processing.ResultStatusSucceeded:
		run.Status = RunStatusSucceeded
		run.LastErrorCode = nil
		run.LastErrorMessage = nil
	case processing.ResultStatusTemporaryError:
		if envelope.Attempt >= run.MaxAttempts {
			run.Status = RunStatusExhausted
			terminalFailure = true
		} else {
			run.Status = RunStatusRetryScheduled
		}
	case processing.ResultStatusFatalError:
		run.Status = RunStatusFailedFatal
		terminalFailure = true
	default:
		return fmt.Errorf("unsupported result status %q", envelope.Status)
	}

	if err := s.repo.UpdateChannelRun(ctx, run); err != nil {
		return err
	}
	if terminalFailure {
		if err := s.repo.MarkExaminationFailed(ctx, envelope.ExaminationID, finishedAt); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) HandleResult(ctx context.Context, envelope processing.ChannelResultEnvelope) error {
	return h.repo.HandleResult(ctx, h.service, envelope)
}

type SQLRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{pool: pool}
}

func (r *SQLRepository) GetChannelRun(ctx context.Context, examinationID int64, channel string) (ChannelRun, error) {
	const query = `
SELECT id, examination_id, channel, status, attempt_count, max_attempts, message_version, last_error_code, last_error_message, finished_at
FROM examination_channel_runs
WHERE examination_id = $1 AND channel = $2
FOR UPDATE`

	tx, ok := txFromContext(ctx)
	if !ok {
		return ChannelRun{}, errors.New("channelresults: transactional context required")
	}

	var run ChannelRun
	var finishedAt *time.Time
	if err := tx.QueryRow(ctx, query, examinationID, channel).Scan(
		&run.ChannelRunID,
		&run.ExaminationID,
		&run.Channel,
		&run.Status,
		&run.AttemptCount,
		&run.MaxAttempts,
		&run.MessageVersion,
		&run.LastErrorCode,
		&run.LastErrorMessage,
		&finishedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ChannelRun{}, repository.ErrNotFound
		}
		return ChannelRun{}, err
	}
	run.FinishedAt = finishedAt
	return run, nil
}

func (r *SQLRepository) SaveResult(ctx context.Context, result StoredResult) error {
	const query = `
INSERT INTO channel_results (
	examination_id,
	channel_run_id,
	channel,
	message_version,
	attempt,
	status,
	model_version,
	payload,
	error_code,
	error_message,
	broker_message_id,
	broker_correlation_id,
	completed_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT (channel_run_id, attempt) DO NOTHING`

	tx, ok := txFromContext(ctx)
	if !ok {
		return errors.New("channelresults: transactional context required")
	}
	_, err := tx.Exec(
		ctx,
		query,
		result.ExaminationID,
		result.ChannelRunID,
		result.Channel,
		result.MessageVersion,
		result.Attempt,
		result.Status,
		result.ModelVersion,
		result.Payload,
		result.ErrorCode,
		result.ErrorMessage,
		result.BrokerMessageID,
		result.BrokerCorrelationID,
		result.CompletedAt,
	)
	return err
}

func (r *SQLRepository) UpdateChannelRun(ctx context.Context, run ChannelRun) error {
	const query = `
UPDATE examination_channel_runs
SET
	status = $2,
	attempt_count = $3,
	message_version = $4,
	last_error_code = $5,
	last_error_message = $6,
	finished_at = $7,
	updated_at = NOW()
WHERE id = $1`

	tx, ok := txFromContext(ctx)
	if !ok {
		return errors.New("channelresults: transactional context required")
	}
	_, err := tx.Exec(
		ctx,
		query,
		run.ChannelRunID,
		run.Status,
		run.AttemptCount,
		run.MessageVersion,
		run.LastErrorCode,
		run.LastErrorMessage,
		run.FinishedAt,
	)
	return err
}

func (r *SQLRepository) MarkExaminationFailed(ctx context.Context, examinationID int64, failedAt time.Time) error {
	const query = `
UPDATE examinations
SET
	status = 'failed',
	updated_at = $2
WHERE id = $1`

	tx, ok := txFromContext(ctx)
	if !ok {
		return errors.New("channelresults: transactional context required")
	}
	_, err := tx.Exec(ctx, query, examinationID, failedAt.UTC())
	return err
}

type txContextKey struct{}

func txFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txContextKey{}).(pgx.Tx)
	return tx, ok
}

func withTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

func (r *SQLRepository) HandleResult(ctx context.Context, service *Service, envelope processing.ChannelResultEnvelope) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := service.ApplyResult(withTx(ctx, tx), envelope); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	if envelope.Status == processing.ResultStatusSucceeded && service.trigger != nil {
		return service.trigger.TryAggregate(ctx, envelope.ExaminationID)
	}
	return nil
}
