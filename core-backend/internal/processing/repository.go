package processing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"dimplom/internal/examinations"
	"dimplom/internal/repository"
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
	pool    *pgxpool.Pool
	bucket  string
	nowFunc func() time.Time
}

func NewRepository(pool *pgxpool.Pool, bucket string) *SQLRepository {
	return &SQLRepository{
		pool:    pool,
		bucket:  bucket,
		nowFunc: func() time.Time { return time.Now().UTC() },
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
	runs, err := ensureChannelRuns(ctx, tx, exam)
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

func ensureChannelRuns(ctx context.Context, tx pgx.Tx, exam examinations.Examination) ([]channelRunRecord, error) {
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
	for _, channel := range MandatoryChannels {
		var run channelRunRecord
		if err := tx.QueryRow(ctx, query, exam.ID, channel, defaultMaxAttempts, MessageVersionV1).Scan(
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
	for _, run := range runs {
		command := buildCommandEnvelope(bucket, requestedAt, exam, run, answerRefs)
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
