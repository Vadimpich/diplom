package processing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"diplom/internal/examinations"
	"diplom/internal/observability"
	"diplom/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxAttempts = int32(3)
	commandsExchange   = "processing.commands"
)

type SQLRepository struct {
	pool                *pgxpool.Pool
	bucket              string
	nowFunc             func() time.Time
	maxAttemptsProvider MaxAttemptsProvider
}

type MaxAttemptsProvider interface {
	ProcessingMaxAttempts(context.Context) (int32, error)
}

func NewRepository(pool *pgxpool.Pool, bucket string, providers ...MaxAttemptsProvider) *SQLRepository {
	var provider MaxAttemptsProvider
	if len(providers) > 0 {
		provider = providers[0]
	}
	return &SQLRepository{
		pool:                pool,
		bucket:              bucket,
		nowFunc:             func() time.Time { return time.Now().UTC() },
		maxAttemptsProvider: provider,
	}
}

func (r *SQLRepository) FinishLaunch(ctx context.Context, examinationID int64) (examinations.Examination, []ProcessingCommandEnvelope, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return examinations.Examination{}, nil, err
	}
	defer tx.Rollback(ctx)

	exam, err := lockExamination(ctx, tx, examinationID)
	if err != nil {
		return examinations.Examination{}, nil, err
	}

	switch exam.Status {
	case examinations.StatusCollectingAnswers:
		if err := ensureAnswersComplete(ctx, tx, examinationID); err != nil {
			return examinations.Examination{}, nil, err
		}
		finishedAt, updatedAt, err := finishExamination(ctx, tx, examinationID)
		if err != nil {
			return examinations.Examination{}, nil, err
		}
		exam.Status = examinations.StatusReadyForProcessing
		exam.FinishedAt = &finishedAt
		exam.UpdatedAt = updatedAt
	case examinations.StatusReadyForProcessing:
	default:
		return examinations.Examination{}, nil, examinations.ErrInvalidTransition
	}

	if err := createLaunchFence(ctx, tx, examinationID); err != nil {
		return examinations.Examination{}, nil, err
	}

	answerRefs, err := loadAnswerReferences(ctx, tx, examinationID)
	if err != nil {
		return examinations.Examination{}, nil, err
	}
	runs, err := r.ensureChannelRuns(ctx, tx, exam)
	if err != nil {
		return examinations.Examination{}, nil, err
	}
	commands, err := ensureOutboxRows(ctx, tx, r.bucket, r.nowFunc(), exam, runs, answerRefs)
	if err != nil {
		return examinations.Examination{}, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return examinations.Examination{}, nil, err
	}
	return exam, commands, nil
}

type channelRunRecord struct {
	ID             int64
	Channel        string
	AttemptCount   int32
	MaxAttempts    int32
	MessageVersion int32
}

type answerReferenceRecord struct {
	AnswerID   int64
	QuestionID int64
	AudioS3Key string
	AnswerText string
}

func lockExamination(ctx context.Context, tx pgx.Tx, examinationID int64) (examinations.Examination, error) {
	const query = `
SELECT id, specialist_id, created_by_user_id, questionnaire_id, status, created_at, started_at, finished_at, updated_at
FROM examinations
WHERE id = $1
FOR UPDATE`

	var (
		exam            examinations.Examination
		questionnaireID pgtype.Int8
		startedAt       pgtype.Timestamptz
		finishedAt      pgtype.Timestamptz
		updatedAt       pgtype.Timestamptz
	)
	err := tx.QueryRow(ctx, query, examinationID).Scan(
		&exam.ID,
		&exam.SpecialistID,
		&exam.CreatedByUserID,
		&questionnaireID,
		&exam.Status,
		&exam.CreatedAt,
		&startedAt,
		&finishedAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return examinations.Examination{}, repository.ErrNotFound
		}
		return examinations.Examination{}, err
	}
	if questionnaireID.Valid {
		exam.QuestionnaireID = &questionnaireID.Int64
	}
	if startedAt.Valid {
		exam.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		exam.FinishedAt = &finishedAt.Time
	}
	if updatedAt.Valid {
		exam.UpdatedAt = updatedAt.Time
	}
	return exam, nil
}

func ensureAnswersComplete(ctx context.Context, tx pgx.Tx, examinationID int64) error {
	const query = `
SELECT
	(SELECT COUNT(*) FROM answers WHERE examination_id = $1),
	(SELECT COUNT(*) FROM examination_questions WHERE examination_id = $1)`
	var answerCount, questionCount int64
	if err := tx.QueryRow(ctx, query, examinationID).Scan(&answerCount, &questionCount); err != nil {
		return err
	}
	if answerCount < questionCount {
		return examinations.ErrAnswersIncomplete
	}
	return nil
}

func finishExamination(ctx context.Context, tx pgx.Tx, examinationID int64) (time.Time, time.Time, error) {
	const query = `
UPDATE examinations
SET
	status = 'ready_for_processing',
	finished_at = COALESCE(finished_at, NOW()),
	updated_at = NOW()
WHERE id = $1
RETURNING finished_at, updated_at`
	var finishedAt, updatedAt time.Time
	if err := tx.QueryRow(ctx, query, examinationID).Scan(&finishedAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, time.Time{}, repository.ErrNotFound
		}
		return time.Time{}, time.Time{}, err
	}
	return finishedAt.UTC(), updatedAt.UTC(), nil
}

func createLaunchFence(ctx context.Context, tx pgx.Tx, examinationID int64) error {
	const query = `
INSERT INTO examination_processing_launches (examination_id)
VALUES ($1)
ON CONFLICT DO NOTHING`
	_, err := tx.Exec(ctx, query, examinationID)
	return err
}

func loadAnswerReferences(ctx context.Context, tx pgx.Tx, examinationID int64) ([]answerReferenceRecord, error) {
	const query = `
SELECT
	a.id,
	a.examination_question_id,
	a.audio_s3_key,
	a.answer_text
FROM answers a
JOIN examination_questions eq
	ON eq.id = a.examination_question_id
WHERE a.examination_id = $1
ORDER BY eq.position ASC, a.id ASC`

	rows, err := tx.Query(ctx, query, examinationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	refs := make([]answerReferenceRecord, 0)
	for rows.Next() {
		var ref answerReferenceRecord
		if err := rows.Scan(&ref.AnswerID, &ref.QuestionID, &ref.AudioS3Key, &ref.AnswerText); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return refs, nil
}

func (r *SQLRepository) ensureChannelRuns(ctx context.Context, tx pgx.Tx, exam examinations.Examination) ([]channelRunRecord, error) {
	const query = `
INSERT INTO examination_channel_runs (
	examination_id,
	channel,
	status,
	attempt_count,
	max_attempts,
	message_version
)
VALUES ($1, $2, 'pending', 0, $3, $4)
ON CONFLICT (examination_id, channel) DO UPDATE
SET
	max_attempts = EXCLUDED.max_attempts,
	message_version = EXCLUDED.message_version,
	updated_at = NOW()
RETURNING id, channel, attempt_count, max_attempts, message_version`

	runs := make([]channelRunRecord, 0, len(MandatoryChannels))
	maxAttempts := defaultMaxAttempts
	if r.maxAttemptsProvider != nil {
		value, err := r.maxAttemptsProvider.ProcessingMaxAttempts(ctx)
		if err != nil {
			return nil, err
		}
		if value > 0 {
			maxAttempts = value
		}
	}
	for _, channel := range MandatoryChannels {
		var run channelRunRecord
		if err := tx.QueryRow(ctx, query, exam.ID, channel, maxAttempts, MessageVersionV1).Scan(
			&run.ID,
			&run.Channel,
			&run.AttemptCount,
			&run.MaxAttempts,
			&run.MessageVersion,
		); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, nil
}

func ensureOutboxRows(
	ctx context.Context,
	tx pgx.Tx,
	bucket string,
	requestedAt time.Time,
	exam examinations.Examination,
	runs []channelRunRecord,
	answerRefs []answerReferenceRecord,
) ([]ProcessingCommandEnvelope, error) {
	const insertQuery = `
INSERT INTO processing_outbox (
	examination_id,
	channel_run_id,
	channel,
	exchange_name,
	routing_key,
	status,
	attempt_count,
	max_attempts,
	message_version,
	payload,
	broker_correlation_id
)
VALUES ($1, $2, $3, $4, $5, 'pending', $6, $7, $8, $9, $10)
ON CONFLICT (channel_run_id) DO NOTHING`

	commands := make([]ProcessingCommandEnvelope, 0, len(runs))
	trace := observability.TraceFromContext(ctx)
	for _, run := range runs {
		command := buildCommandEnvelope(bucket, requestedAt, exam, run, answerRefs, trace)
		payload, err := json.Marshal(command)
		if err != nil {
			return nil, fmt.Errorf("marshal command payload: %w", err)
		}
		if _, err := tx.Exec(
			ctx,
			insertQuery,
			exam.ID,
			run.ID,
			run.Channel,
			commandsExchange,
			routingKeyForChannel(run.Channel),
			run.AttemptCount,
			run.MaxAttempts,
			run.MessageVersion,
			payload,
			command.CorrelationID,
		); err != nil {
			return nil, err
		}
		commands = append(commands, command)
	}
	return commands, nil
}

func buildCommandEnvelope(
	bucket string,
	requestedAt time.Time,
	exam examinations.Examination,
	run channelRunRecord,
	answerRefs []answerReferenceRecord,
	trace observability.TraceContext,
) ProcessingCommandEnvelope {
	answers := make([]CommandAnswerReference, 0, len(answerRefs))
	for _, ref := range answerRefs {
		answers = append(answers, CommandAnswerReference{
			AnswerID:      ref.AnswerID,
			QuestionID:    ref.QuestionID,
			AudioS3Bucket: bucket,
			AudioS3Key:    ref.AudioS3Key,
			AnswerText:    ref.AnswerText,
		})
	}

	return ProcessingCommandEnvelope{
		MessageVersion: int(run.MessageVersion),
		MessageID:      uuid.NewString(),
		CorrelationID:  fmt.Sprintf("exam-%d-%s-v%d", exam.ID, run.Channel, run.MessageVersion),
		RequestID:      trace.RequestID,
		TraceParent:    trace.TraceParent,
		TraceState:     trace.TraceState,
		ExaminationID:  exam.ID,
		SpecialistID:   exam.SpecialistID,
		Channel:        run.Channel,
		Attempt:        run.AttemptCount + 1,
		MaxAttempts:    run.MaxAttempts,
		RequestedAt:    requestedAt,
		Answers:        answers,
	}
}

func routingKeyForChannel(channel string) string {
	switch channel {
	case ChannelText:
		return "processing.command.text"
	case ChannelAcoustic:
		return "processing.command.acoustic"
	default:
		return "processing.command.paralinguistic"
	}
}

func (r *SQLRepository) GetProcessingStatus(ctx context.Context, examinationID int64) (ProcessingStatusResponse, error) {
	const query = `
SELECT
	e.id,
	e.status,
	e.finished_at,
	e.updated_at,
	cr.channel,
	cr.status,
	cr.attempt_count,
	cr.max_attempts,
	cr.message_version,
	cr.last_error_code,
	cr.last_error_message,
	cr.broker_message_id,
	cr.broker_correlation_id,
	cr.queued_at,
	cr.started_at,
	cr.finished_at,
	po.status,
	po.published_at
FROM examinations e
LEFT JOIN examination_channel_runs cr
	ON cr.examination_id = e.id
LEFT JOIN processing_outbox po
	ON po.channel_run_id = cr.id
WHERE e.id = $1
ORDER BY cr.channel ASC`

	rows, err := r.pool.Query(ctx, query, examinationID)
	if err != nil {
		return ProcessingStatusResponse{}, err
	}
	defer rows.Close()

	response := ProcessingStatusResponse{
		ExaminationID:  examinationID,
		MessageVersion: MessageVersionV1,
		ChannelsTotal:  len(MandatoryChannels),
		Channels:       make([]ChannelStatusDTO, 0, len(MandatoryChannels)),
	}

	var found bool
	var firstActivityAt *time.Time
	var failedAt *time.Time
	for rows.Next() {
		found = true
		var (
			examID          int64
			examStatus      string
			examFinishedAt  pgtype.Timestamptz
			examUpdatedAt   pgtype.Timestamptz
			channel         pgtype.Text
			channelStatus   pgtype.Text
			attemptCount    pgtype.Int4
			maxAttempts     pgtype.Int4
			messageVersion  pgtype.Int4
			lastErrorCode   pgtype.Text
			lastErrorMsg    pgtype.Text
			brokerMessageID pgtype.Text
			brokerCorrID    pgtype.Text
			queuedAt        pgtype.Timestamptz
			startedAt       pgtype.Timestamptz
			finishedAt      pgtype.Timestamptz
			outboxStatus    pgtype.Text
			publishedAt     pgtype.Timestamptz
		)
		if err := rows.Scan(
			&examID,
			&examStatus,
			&examFinishedAt,
			&examUpdatedAt,
			&channel,
			&channelStatus,
			&attemptCount,
			&maxAttempts,
			&messageVersion,
			&lastErrorCode,
			&lastErrorMsg,
			&brokerMessageID,
			&brokerCorrID,
			&queuedAt,
			&startedAt,
			&finishedAt,
			&outboxStatus,
			&publishedAt,
		); err != nil {
			return ProcessingStatusResponse{}, err
		}

		response.ExaminationID = examID
		response.Status = examStatus
		if examUpdatedAt.Valid {
			response.UpdatedAt = examUpdatedAt.Time
		}
		if examFinishedAt.Valid && (examStatus == examinations.StatusAggregated || examStatus == "decision_pending" || examStatus == "completed") {
			response.FinishedAt = &examFinishedAt.Time
		}

		if channel.Valid {
			runtimeStatus, runtimeStartedAt := resolveChannelRuntimeStatus(
				channelStatus.String,
				timePtrFromPG(queuedAt),
				timePtrFromPG(startedAt),
				timePtrFromPG(finishedAt),
				textFromPG(outboxStatus),
				timePtrFromPG(publishedAt),
			)
			dto := ChannelStatusDTO{
				Channel:        channel.String,
				Status:         runtimeStatus,
				AttemptCount:   attemptCount.Int32,
				MaxAttempts:    maxAttempts.Int32,
				MessageVersion: int(messageVersion.Int32),
			}
			if queuedAt.Valid {
				dto.QueuedAt = &queuedAt.Time
			}
			if runtimeStartedAt != nil {
				dto.StartedAt = runtimeStartedAt
			}
			if finishedAt.Valid {
				dto.FinishedAt = &finishedAt.Time
			}
			if lastErrorCode.Valid {
				dto.LastErrorCode = &lastErrorCode.String
			}
			if lastErrorMsg.Valid {
				dto.LastErrorMessage = &lastErrorMsg.String
			}
			if brokerMessageID.Valid {
				dto.BrokerMessageID = &brokerMessageID.String
			}
			if brokerCorrID.Valid {
				dto.BrokerCorrelationID = &brokerCorrID.String
			}
			response.Channels = append(response.Channels, dto)
			if queuedAt.Valid {
				firstActivityAt = earlierTime(firstActivityAt, queuedAt.Time)
			}
			if runtimeStartedAt != nil {
				firstActivityAt = earlierTime(firstActivityAt, *runtimeStartedAt)
			}
			if dto.Status == "succeeded" {
				response.ChannelsComplete++
			}
			if response.Status == examinations.StatusFailed && finishedAt.Valid && isFailureChannelStatus(dto.Status) {
				failedAt = laterTime(failedAt, finishedAt.Time)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return ProcessingStatusResponse{}, err
	}
	if !found {
		return ProcessingStatusResponse{}, repository.ErrNotFound
	}
	switch response.Status {
	case examinations.StatusReadyForProcessing, examinations.StatusProcessing, examinations.StatusAggregating, examinations.StatusAggregated, "decision_pending", "completed", examinations.StatusFailed:
	default:
		return ProcessingStatusResponse{}, ErrProcessingStatusUnavailable
	}
	response.StartedAt = firstActivityAt
	response.FailedAt = failedAt
	response.Terminal = response.Status == examinations.StatusFailed || response.Status == "completed"
	return response, nil
}

func resolveChannelRuntimeStatus(
	channelStatus string,
	queuedAt *time.Time,
	startedAt *time.Time,
	finishedAt *time.Time,
	outboxStatus string,
	publishedAt *time.Time,
) (string, *time.Time) {
	if startedAt != nil {
		value := startedAt.UTC()
		return channelStatus, &value
	}
	if finishedAt != nil {
		return channelStatus, nil
	}
	if channelStatus == "queued" && outboxStatus == "published" && publishedAt != nil {
		value := publishedAt.UTC()
		return "processing", &value
	}
	if channelStatus == "queued" && queuedAt != nil {
		return channelStatus, nil
	}
	return channelStatus, nil
}

func timePtrFromPG(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	item := value.Time.UTC()
	return &item
}

func textFromPG(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func isTerminalChannelStatus(status string) bool {
	switch status {
	case "succeeded", "failed_fatal", "exhausted":
		return true
	default:
		return false
	}
}

func isFailureChannelStatus(status string) bool {
	switch status {
	case "failed_fatal", "exhausted":
		return true
	default:
		return false
	}
}

func earlierTime(current *time.Time, candidate time.Time) *time.Time {
	value := candidate.UTC()
	if current == nil || value.Before(current.UTC()) {
		return &value
	}
	return current
}

func laterTime(current *time.Time, candidate time.Time) *time.Time {
	value := candidate.UTC()
	if current == nil || value.After(current.UTC()) {
		return &value
	}
	return current
}

func (r *SQLRepository) ListPendingOutbox(ctx context.Context, limit int32) ([]OutboxMessage, error) {
	const query = `
SELECT
	id,
	examination_id,
	channel_run_id,
	channel,
	exchange_name,
	routing_key,
	attempt_count,
	max_attempts,
	message_version,
	payload,
	COALESCE(broker_correlation_id, '')
FROM processing_outbox
WHERE status = 'pending'
	AND attempt_count < max_attempts
ORDER BY created_at ASC, id ASC
LIMIT $1`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]OutboxMessage, 0)
	for rows.Next() {
		var item OutboxMessage
		var payload []byte
		if err := rows.Scan(
			&item.ID,
			&item.ExaminationID,
			&item.ChannelRunID,
			&item.Channel,
			&item.ExchangeName,
			&item.RoutingKey,
			&item.AttemptCount,
			&item.MaxAttempts,
			&item.MessageVersion,
			&payload,
			&item.BrokerCorrelation,
		); err != nil {
			return nil, err
		}
		item.Payload = payload
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *SQLRepository) MarkOutboxPublished(ctx context.Context, update PublishedOutboxUpdate) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const outboxQuery = `
UPDATE processing_outbox
SET
	status = 'published',
	attempt_count = $2,
	broker_message_id = $3,
	broker_correlation_id = $4,
	last_error_code = NULL,
	last_error_message = NULL,
	published_at = $5,
	updated_at = NOW()
WHERE id = $1`
	if _, err := tx.Exec(
		ctx,
		outboxQuery,
		update.OutboxID,
		update.AttemptCount,
		update.BrokerMessageID,
		update.BrokerCorrelation,
		update.PublishedAt,
	); err != nil {
		return err
	}

	const runQuery = `
UPDATE examination_channel_runs
SET
	status = 'queued',
	attempt_count = $2,
	broker_message_id = $3,
	broker_correlation_id = $4,
	last_error_code = NULL,
	last_error_message = NULL,
	queued_at = $5,
	updated_at = NOW()
WHERE id = $1`
	if _, err := tx.Exec(
		ctx,
		runQuery,
		update.ChannelRunID,
		update.AttemptCount,
		update.BrokerMessageID,
		update.BrokerCorrelation,
		update.PublishedAt,
	); err != nil {
		return err
	}

	const examQuery = `
UPDATE examinations
SET
	status = 'processing',
	updated_at = NOW()
WHERE id = $1
	AND status = 'ready_for_processing'`
	if _, err := tx.Exec(ctx, examQuery, update.ExaminationID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *SQLRepository) MarkOutboxFailed(ctx context.Context, update FailedOutboxUpdate) error {
	status := update.Status
	if status == "" {
		status = outboxStatusPending
	}

	const query = `
UPDATE processing_outbox
SET
	status = $2,
	attempt_count = $3,
	last_error_code = $4,
	last_error_message = $5,
	updated_at = NOW()
WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, update.OutboxID, status, update.AttemptCount, update.ErrorCode, update.ErrorMessage)
	return err
}
