package specialists

import (
	"context"
	"errors"

	sqlcdb "diplom/db/sqlc"
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

func (r *SQLCRepository) Create(ctx context.Context, input CreateInput) (Specialist, error) {
	row, err := r.queries.CreateSpecialist(ctx, sqlcdb.CreateSpecialistParams{
		FullName:        input.FullName,
		PersonnelNumber: textValue(input.PersonnelNumber),
	})
	if err != nil {
		if isConflict(err) {
			return Specialist{}, repository.ErrConflict
		}
		return Specialist{}, err
	}

	return mapSpecialist(row), nil
}

func (r *SQLCRepository) List(ctx context.Context) ([]Specialist, error) {
	rows, err := r.queries.ListSpecialists(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Specialist, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapSpecialist(row))
	}

	return result, nil
}

func (r *SQLCRepository) GetByID(ctx context.Context, id int64) (Specialist, error) {
	row, err := r.queries.GetSpecialistByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Specialist{}, repository.ErrNotFound
		}
		return Specialist{}, err
	}

	return mapSpecialist(row), nil
}

func (r *SQLCRepository) Update(ctx context.Context, input UpdateInput) (Specialist, error) {
	row, err := r.queries.UpdateSpecialist(ctx, sqlcdb.UpdateSpecialistParams{
		ID:              input.ID,
		FullName:        input.FullName,
		PersonnelNumber: textValue(input.PersonnelNumber),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Specialist{}, repository.ErrNotFound
		}
		if isConflict(err) {
			return Specialist{}, repository.ErrConflict
		}
		return Specialist{}, err
	}

	return mapSpecialist(row), nil
}

func (r *SQLCRepository) Delete(ctx context.Context, id int64) error {
	deleted, err := r.queries.DeleteSpecialist(ctx, id)
	if err != nil {
		if isConflict(err) {
			return repository.ErrConflict
		}
		return err
	}
	if deleted == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func mapSpecialist(row sqlcdb.Specialist) Specialist {
	return Specialist{
		ID:              row.ID,
		FullName:        row.FullName,
		PersonnelNumber: nullableText(row.PersonnelNumber),
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}
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

func isConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && (pgErr.Code == "23505" || pgErr.Code == "23503")
}
