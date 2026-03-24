package questionnaires

import (
	"context"
	"errors"

	sqlcdb "diplom/db/sqlc"
	"diplom/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLCRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *SQLCRepository {
	return &SQLCRepository{pool: pool}
}

func (r *SQLCRepository) List(ctx context.Context) ([]Questionnaire, error) {
	rows, err := sqlcdb.New(r.pool).ListQuestionnaireDetails(ctx)
	if err != nil {
		return nil, err
	}
	return mapQuestionnaireRows(rows), nil
}

func (r *SQLCRepository) GetByID(ctx context.Context, id int64) (Questionnaire, error) {
	rows, err := sqlcdb.New(r.pool).GetQuestionnaireDetailsByID(ctx, id)
	if err != nil {
		return Questionnaire{}, err
	}
	if len(rows) == 0 {
		return Questionnaire{}, repository.ErrNotFound
	}
	return mapQuestionnaireDetailsRows(rows)[0], nil
}

func (r *SQLCRepository) Create(ctx context.Context, input CreateInput) (Questionnaire, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Questionnaire{}, err
	}
	defer tx.Rollback(ctx)

	queries := sqlcdb.New(tx)
	item, err := queries.CreateQuestionnaire(ctx, sqlcdb.CreateQuestionnaireParams{
		Title:       input.Title,
		Description: textValue(input.Description),
		IsActive:    input.IsActive,
	})
	if err != nil {
		return Questionnaire{}, err
	}

	if err := upsertQuestions(ctx, queries, item.ID, input.Questions); err != nil {
		return Questionnaire{}, err
	}

	rows, err := queries.GetQuestionnaireDetailsByID(ctx, item.ID)
	if err != nil {
		return Questionnaire{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Questionnaire{}, err
	}

	return mapQuestionnaireDetailsRows(rows)[0], nil
}

func (r *SQLCRepository) Update(ctx context.Context, input UpdateInput) (Questionnaire, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Questionnaire{}, err
	}
	defer tx.Rollback(ctx)

	queries := sqlcdb.New(tx)
	if _, err := queries.GetQuestionnaireByID(ctx, input.ID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Questionnaire{}, repository.ErrNotFound
		}
		return Questionnaire{}, err
	}

	if _, err := queries.UpdateQuestionnaire(ctx, sqlcdb.UpdateQuestionnaireParams{
		ID:          input.ID,
		Title:       input.Title,
		Description: textValue(input.Description),
		IsActive:    input.IsActive,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Questionnaire{}, repository.ErrNotFound
		}
		return Questionnaire{}, err
	}

	if err := queries.DeleteQuestionnaireQuestions(ctx, input.ID); err != nil {
		return Questionnaire{}, err
	}
	if err := queries.DeleteOrphanQuestions(ctx); err != nil {
		return Questionnaire{}, err
	}
	if err := upsertQuestions(ctx, queries, input.ID, input.Questions); err != nil {
		return Questionnaire{}, err
	}

	rows, err := queries.GetQuestionnaireDetailsByID(ctx, input.ID)
	if err != nil {
		return Questionnaire{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Questionnaire{}, err
	}

	return mapQuestionnaireDetailsRows(rows)[0], nil
}

func upsertQuestions(ctx context.Context, queries *sqlcdb.Queries, questionnaireID int64, questions []QuestionInput) error {
	for idx, question := range questions {
		row, err := queries.CreateQuestion(ctx, question.Text)
		if err != nil {
			return err
		}
		if err := queries.AddQuestionToQuestionnaire(ctx, sqlcdb.AddQuestionToQuestionnaireParams{
			QuestionnaireID: questionnaireID,
			QuestionID:      row.ID,
			Position:        int32(idx + 1),
		}); err != nil {
			return err
		}
	}
	return nil
}

func mapQuestionnaireRows(rows []sqlcdb.ListQuestionnaireDetailsRow) []Questionnaire {
	result := make([]Questionnaire, 0)
	indexByID := make(map[int64]int)

	for _, row := range rows {
		idx, exists := indexByID[row.ID]
		if !exists {
			result = append(result, Questionnaire{
				ID:          row.ID,
				Title:       row.Title,
				Description: nullableText(row.Description),
				IsActive:    row.IsActive,
				Questions:   make([]Question, 0),
				CreatedAt:   row.CreatedAt.Time,
				UpdatedAt:   row.UpdatedAt.Time,
			})
			idx = len(result) - 1
			indexByID[row.ID] = idx
		}

		if row.QuestionID.Valid {
			result[idx].Questions = append(result[idx].Questions, Question{
				ID:       row.QuestionID.Int64,
				Text:     row.QuestionText.String,
				Position: row.Position.Int32,
			})
		}
	}

	return result
}

func mapQuestionnaireDetailsRows(rows []sqlcdb.GetQuestionnaireDetailsByIDRow) []Questionnaire {
	result := make([]Questionnaire, 0)
	indexByID := make(map[int64]int)

	for _, row := range rows {
		idx, exists := indexByID[row.ID]
		if !exists {
			result = append(result, Questionnaire{
				ID:          row.ID,
				Title:       row.Title,
				Description: nullableText(row.Description),
				IsActive:    row.IsActive,
				Questions:   make([]Question, 0),
				CreatedAt:   row.CreatedAt.Time,
				UpdatedAt:   row.UpdatedAt.Time,
			})
			idx = len(result) - 1
			indexByID[row.ID] = idx
		}

		if row.QuestionID.Valid {
			result[idx].Questions = append(result[idx].Questions, Question{
				ID:       row.QuestionID.Int64,
				Text:     row.QuestionText.String,
				Position: row.Position.Int32,
			})
		}
	}

	return result
}

func textValue(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func nullableText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}
