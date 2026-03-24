package answers

import (
	"context"
	"errors"
	"time"

	sqlcdb "diplom/db/sqlc"
	"diplom/internal/examinations"
	"diplom/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type SQLCRepository struct {
	queries *sqlcdb.Queries
}

func NewRepository(queries *sqlcdb.Queries) *SQLCRepository {
	return &SQLCRepository{queries: queries}
}

func (r *SQLCRepository) ReserveID(ctx context.Context) (int64, error) {
	return r.queries.NextAnswerID(ctx)
}

func (r *SQLCRepository) Create(ctx context.Context, params CreateParams) (Answer, error) {
	row, err := r.queries.CreateAnswer(ctx, sqlcdb.CreateAnswerParams{
		ID:                    params.ID,
		ExaminationID:         params.ExaminationID,
		ExaminationQuestionID: pgtype.Int8{Int64: params.ExaminationQuestionID, Valid: true},
		SpecialistID:          pgtype.Int8{Int64: params.SpecialistID, Valid: true},
		CreatedByUserID:       params.CreatedByUserID,
		AnswerText:            params.Text,
		AudioS3Key:            params.AudioS3Key,
	})
	if err != nil {
		if isConflict(err) {
			return Answer{}, repository.ErrConflict
		}
		return Answer{}, err
	}
	return Answer{
		ID:                    row.ID,
		ExaminationID:         row.ExaminationID,
		ExaminationQuestionID: int64Value(row.ExaminationQuestionID),
		SpecialistID:          int64Value(row.SpecialistID),
		CreatedByUserID:       row.CreatedByUserID,
		Text:                  row.AnswerText,
		AudioS3Key:            row.AudioS3Key,
		CreatedAt:             row.CreatedAt.Time,
	}, nil
}

func (r *SQLCRepository) GetExaminationByID(ctx context.Context, id int64) (examinations.Examination, error) {
	row, err := r.queries.GetExaminationByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return examinations.Examination{}, repository.ErrNotFound
		}
		return examinations.Examination{}, err
	}
	return examinations.Examination{
		ID:              row.ID,
		SpecialistID:    row.SpecialistID,
		CreatedByUserID: row.CreatedByUserID,
		Status:          row.Status,
		CreatedAt:       row.CreatedAt.Time,
		StartedAt:       timestampPtr(row.StartedAt),
		FinishedAt:      timestampPtr(row.FinishedAt),
		UpdatedAt:       row.UpdatedAt.Time,
	}, nil
}

func (r *SQLCRepository) GetExaminationQuestionByID(ctx context.Context, id int64) (examinations.ExaminationQuestion, error) {
	row, err := r.queries.GetExaminationQuestionByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return examinations.ExaminationQuestion{}, repository.ErrNotFound
		}
		return examinations.ExaminationQuestion{}, err
	}
	return examinations.ExaminationQuestion{
		ID:               row.ID,
		ExaminationID:    row.ExaminationID,
		SpecialistID:     row.SpecialistID,
		QuestionnaireID:  row.QuestionnaireID,
		SourceQuestionID: int64Ptr(row.SourceQuestionID),
		Position:         row.Position,
		QuestionText:     row.QuestionText,
	}, nil
}

func timestampPtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func int64Ptr(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Int64
	return &result
}

func int64Value(value pgtype.Int8) int64 {
	if !value.Valid {
		return 0
	}
	return value.Int64
}

func isConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
