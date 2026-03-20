package examinations

import (
	"context"
	"errors"

	sqlcdb "dimplom/db/sqlc"
	"dimplom/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLCRepository struct {
	pool    *pgxpool.Pool
	queries *sqlcdb.Queries
}

func NewRepository(pool *pgxpool.Pool) *SQLCRepository {
	return &SQLCRepository{
		pool:    pool,
		queries: sqlcdb.New(pool),
	}
}

func (r *SQLCRepository) Create(ctx context.Context, input CreateInput) (Examination, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Examination{}, err
	}
	defer tx.Rollback(ctx)

	queries := sqlcdb.New(tx)
	if _, err := queries.GetSpecialistByID(ctx, input.SpecialistID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Examination{}, repository.ErrNotFound
		}
		return Examination{}, err
	}
	if _, err := queries.GetQuestionnaireByID(ctx, *input.QuestionnaireID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Examination{}, repository.ErrNotFound
		}
		return Examination{}, err
	}

	row, err := queries.CreateExamination(ctx, sqlcdb.CreateExaminationParams{
		SpecialistID:    input.SpecialistID,
		CreatedByUserID: input.CreatedByUserID,
		QuestionnaireID: int8Value(input.QuestionnaireID),
	})
	if err != nil {
		return Examination{}, err
	}

	snapshots, err := queries.CreateExaminationQuestionSnapshots(ctx, sqlcdb.CreateExaminationQuestionSnapshotsParams{
		ExaminationID: row.ID,
		SpecialistID:  input.SpecialistID,
		QuestionnaireID: *input.QuestionnaireID,
	})
	if err != nil {
		return Examination{}, err
	}
	if len(snapshots) == 0 {
		return Examination{}, ErrInvalidInput
	}
	if err := tx.Commit(ctx); err != nil {
		return Examination{}, err
	}
	return mapCreateExamination(row), nil
}

func (r *SQLCRepository) List(ctx context.Context) ([]Examination, error) {
	rows, err := r.queries.ListExaminations(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Examination, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapListExamination(row))
	}

	return result, nil
}

func (r *SQLCRepository) GetByID(ctx context.Context, id int64) (Examination, error) {
	row, err := r.queries.GetExaminationByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Examination{}, repository.ErrNotFound
		}
		return Examination{}, err
	}
	return mapGetExamination(row), nil
}

func (r *SQLCRepository) ListBySpecialistID(ctx context.Context, specialistID int64) ([]Examination, error) {
	if _, err := r.queries.GetSpecialistByID(ctx, specialistID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	rows, err := r.queries.ListExaminationsBySpecialistID(ctx, specialistID)
	if err != nil {
		return nil, err
	}

	result := make([]Examination, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapListExaminationBySpecialist(row))
	}

	return result, nil
}

func (r *SQLCRepository) UpdateStatus(ctx context.Context, id int64, status string) (Examination, error) {
	row, err := r.queries.UpdateExaminationStatus(ctx, sqlcdb.UpdateExaminationStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Examination{}, repository.ErrNotFound
		}
		return Examination{}, err
	}
	return mapUpdateExamination(row), nil
}

func mapCreateExamination(row sqlcdb.CreateExaminationRow) Examination {
	return buildExamination(
		row.ID,
		row.SpecialistID,
		row.CreatedByUserID,
		row.QuestionnaireID,
		row.Status,
		row.CreatedAt,
		row.StartedAt,
		row.FinishedAt,
		row.UpdatedAt,
	)
}

func mapGetExamination(row sqlcdb.GetExaminationByIDRow) Examination {
	return buildExamination(
		row.ID,
		row.SpecialistID,
		row.CreatedByUserID,
		row.QuestionnaireID,
		row.Status,
		row.CreatedAt,
		row.StartedAt,
		row.FinishedAt,
		row.UpdatedAt,
	)
}

func mapListExamination(row sqlcdb.ListExaminationsRow) Examination {
	return buildExamination(
		row.ID,
		row.SpecialistID,
		row.CreatedByUserID,
		row.QuestionnaireID,
		row.Status,
		row.CreatedAt,
		row.StartedAt,
		row.FinishedAt,
		row.UpdatedAt,
	)
}

func mapListExaminationBySpecialist(row sqlcdb.ListExaminationsBySpecialistIDRow) Examination {
	return buildExamination(
		row.ID,
		row.SpecialistID,
		row.CreatedByUserID,
		row.QuestionnaireID,
		row.Status,
		row.CreatedAt,
		row.StartedAt,
		row.FinishedAt,
		row.UpdatedAt,
	)
}

func mapUpdateExamination(row sqlcdb.UpdateExaminationStatusRow) Examination {
	return buildExamination(
		row.ID,
		row.SpecialistID,
		row.CreatedByUserID,
		row.QuestionnaireID,
		row.Status,
		row.CreatedAt,
		row.StartedAt,
		row.FinishedAt,
		row.UpdatedAt,
	)
}

func buildExamination(
	id int64,
	specialistID int64,
	createdByUserID int64,
	questionnaireID pgtype.Int8,
	status string,
	createdAt pgtype.Timestamptz,
	startedAt pgtype.Timestamptz,
	finishedAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
) Examination {
	exam := Examination{
		ID:              id,
		SpecialistID:    specialistID,
		CreatedByUserID: createdByUserID,
		Status:          status,
		CreatedAt:       createdAt.Time,
		UpdatedAt:       updatedAt.Time,
	}
	if startedAt.Valid {
		exam.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		exam.FinishedAt = &finishedAt.Time
	}
	if questionnaireID.Valid {
		exam.QuestionnaireID = &questionnaireID.Int64
	}
	return exam
}

func int8Value(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *value, Valid: true}
}
