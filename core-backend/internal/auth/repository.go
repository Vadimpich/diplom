package auth

import (
	"context"
	"errors"
	"time"

	sqlcdb "diplom/db/sqlc"
	"diplom/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func (r *SQLCRepository) GetUserByLogin(ctx context.Context, login string) (StoredUser, error) {
	row, err := r.queries.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StoredUser{}, repository.ErrNotFound
		}
		return StoredUser{}, err
	}

	return mapStoredUserByLogin(row), nil
}

func (r *SQLCRepository) GetUserByID(ctx context.Context, id int64) (StoredUser, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StoredUser{}, repository.ErrNotFound
		}
		return StoredUser{}, err
	}

	return mapStoredUserByID(row), nil
}

func (r *SQLCRepository) ListUsers(ctx context.Context) ([]StoredUser, error) {
	rows, err := r.queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]StoredUser, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapStoredUserByList(row))
	}

	return result, nil
}

func (r *SQLCRepository) GetRoleBySlug(ctx context.Context, slug string) (Role, error) {
	row, err := r.queries.GetRoleBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Role{}, repository.ErrNotFound
		}
		return Role{}, err
	}

	return Role{
		ID:   row.ID,
		Slug: row.Slug,
		Name: row.Name,
	}, nil
}

func (r *SQLCRepository) CreateUser(ctx context.Context, params CreateUserParams) (StoredUser, error) {
	row, err := r.queries.CreateUser(ctx, sqlcdb.CreateUserParams{
		Login:        params.Login,
		PasswordHash: params.PasswordHash,
		RoleID:       params.RoleID,
	})
	if err != nil {
		if isConflict(err) {
			return StoredUser{}, repository.ErrConflict
		}
		return StoredUser{}, err
	}
	return r.GetUserByID(ctx, row.ID)
}

func (r *SQLCRepository) UpdateUser(ctx context.Context, params UpdateUserParams) (StoredUser, error) {
	passwordHash := pgtype.Text{}
	if params.PasswordHash != nil {
		passwordHash = textValue(*params.PasswordHash)
	}

	row, err := r.queries.UpdateUser(ctx, sqlcdb.UpdateUserParams{
		ID:           params.ID,
		Login:        params.Login,
		RoleID:       params.RoleID,
		IsActive:     params.IsActive,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StoredUser{}, repository.ErrNotFound
		}
		if isConflict(err) {
			return StoredUser{}, repository.ErrConflict
		}
		return StoredUser{}, err
	}

	return r.GetUserByID(ctx, row.ID)
}

func (r *SQLCRepository) UpsertUser(ctx context.Context, params UpsertUserParams) (StoredUser, error) {
	row, err := r.queries.UpsertUser(ctx, sqlcdb.UpsertUserParams{
		Login:        params.Login,
		PasswordHash: params.PasswordHash,
		RoleID:       params.RoleID,
	})
	if err != nil {
		return StoredUser{}, err
	}
	return r.GetUserByID(ctx, row.ID)
}

func (r *SQLCRepository) CreateRefreshSession(ctx context.Context, params CreateRefreshSessionParams) (RefreshSession, error) {
	row, err := r.queries.CreateRefreshSession(ctx, sqlcdb.CreateRefreshSessionParams{
		UserID:      params.UserID,
		TokenHash:   params.TokenHash,
		ExpiresAt:   pgtype.Timestamptz{Time: params.ExpiresAt, Valid: true},
		CreatedByIp: textValue(params.CreatedByIP),
		UserAgent:   textValue(params.UserAgent),
		LastUsedAt:  pgtype.Timestamptz{Time: params.LastUsedAt, Valid: true},
	})
	if err != nil {
		if isConflict(err) {
			return RefreshSession{}, repository.ErrConflict
		}
		return RefreshSession{}, err
	}
	return mapRefreshSession(row), nil
}

func (r *SQLCRepository) GetRefreshSessionByHash(ctx context.Context, tokenHash []byte) (RefreshSession, error) {
	row, err := r.queries.GetRefreshSessionByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RefreshSession{}, repository.ErrNotFound
		}
		return RefreshSession{}, err
	}
	return mapRefreshSession(row), nil
}

func (r *SQLCRepository) RotateRefreshSession(ctx context.Context, params RotateRefreshSessionParams) (RefreshSession, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return RefreshSession{}, err
	}
	defer tx.Rollback(ctx)

	queries := sqlcdb.New(tx)
	current, err := queries.GetRefreshSessionByHash(ctx, params.CurrentTokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RefreshSession{}, repository.ErrNotFound
		}
		return RefreshSession{}, err
	}
	if current.ID != params.SessionID || current.RevokedAt.Valid || params.Now.After(current.ExpiresAt.Time) {
		return RefreshSession{}, repository.ErrNotFound
	}

	created, err := queries.CreateRefreshSession(ctx, sqlcdb.CreateRefreshSessionParams{
		UserID:      current.UserID,
		TokenHash:   params.NewTokenHash,
		ExpiresAt:   pgtype.Timestamptz{Time: params.ExpiresAt, Valid: true},
		CreatedByIp: textValue(params.CreatedByIP),
		UserAgent:   textValue(params.UserAgent),
		LastUsedAt:  pgtype.Timestamptz{Time: params.Now, Valid: true},
	})
	if err != nil {
		if isConflict(err) {
			return RefreshSession{}, repository.ErrConflict
		}
		return RefreshSession{}, err
	}

	if err := queries.RotateRefreshSession(ctx, sqlcdb.RotateRefreshSessionParams{
		ID:                  current.ID,
		RevokedAt:           pgtype.Timestamptz{Time: params.Now, Valid: true},
		ReplacedBySessionID: pgtype.Int8{Int64: created.ID, Valid: true},
		LastUsedAt:          pgtype.Timestamptz{Time: params.Now, Valid: true},
	}); err != nil {
		return RefreshSession{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return RefreshSession{}, err
	}

	return mapRefreshSession(created), nil
}

func (r *SQLCRepository) TouchRefreshSession(ctx context.Context, sessionID int64, touchedAt time.Time) error {
	return r.queries.TouchRefreshSession(ctx, sqlcdb.TouchRefreshSessionParams{
		ID:         sessionID,
		LastUsedAt: pgtype.Timestamptz{Time: touchedAt, Valid: true},
	})
}

func (r *SQLCRepository) RevokeRefreshSession(ctx context.Context, params RevokeRefreshSessionParams) error {
	return r.queries.RevokeRefreshSession(ctx, sqlcdb.RevokeRefreshSessionParams{
		ID:        params.SessionID,
		RevokedAt: pgtype.Timestamptz{Time: params.RevokedAt, Valid: true},
	})
}

func (r *SQLCRepository) UpdateUserLastLogin(ctx context.Context, userID int64, loggedAt time.Time) error {
	return r.queries.UpdateUserLastLogin(ctx, sqlcdb.UpdateUserLastLoginParams{
		ID:          userID,
		LastLoginAt: pgtype.Timestamptz{Time: loggedAt, Valid: true},
	})
}

func isConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func mapStoredUserByLogin(row sqlcdb.GetUserByLoginRow) StoredUser {
	return StoredUser{
		ID:           row.ID,
		Login:        row.Login,
		PasswordHash: row.PasswordHash,
		IsActive:     row.IsActive,
		LastLoginAt:  nullableTime(row.LastLoginAt),
		Created:      row.CreatedAt.Time,
		Updated:      row.UpdatedAt.Time,
		Role: Role{
			ID:   row.RoleID,
			Slug: row.RoleSlug,
			Name: row.RoleName,
		},
	}
}

func mapStoredUserByID(row sqlcdb.GetUserByIDRow) StoredUser {
	return StoredUser{
		ID:           row.ID,
		Login:        row.Login,
		PasswordHash: row.PasswordHash,
		IsActive:     row.IsActive,
		LastLoginAt:  nullableTime(row.LastLoginAt),
		Created:      row.CreatedAt.Time,
		Updated:      row.UpdatedAt.Time,
		Role: Role{
			ID:   row.RoleID,
			Slug: row.RoleSlug,
			Name: row.RoleName,
		},
	}
}

func mapStoredUserByList(row sqlcdb.ListUsersRow) StoredUser {
	return StoredUser{
		ID:           row.ID,
		Login:        row.Login,
		PasswordHash: row.PasswordHash,
		IsActive:     row.IsActive,
		LastLoginAt:  nullableTime(row.LastLoginAt),
		Created:      row.CreatedAt.Time,
		Updated:      row.UpdatedAt.Time,
		Role: Role{
			ID:   row.RoleID,
			Slug: row.RoleSlug,
			Name: row.RoleName,
		},
	}
}

func mapRefreshSession(row sqlcdb.RefreshSession) RefreshSession {
	return RefreshSession{
		ID:                  row.ID,
		UserID:              row.UserID,
		TokenHash:           row.TokenHash,
		ExpiresAt:           row.ExpiresAt.Time,
		RevokedAt:           nullableTime(row.RevokedAt),
		ReplacedBySessionID: nullableInt64(row.ReplacedBySessionID),
		CreatedByIP:         nullableText(row.CreatedByIp),
		UserAgent:           nullableText(row.UserAgent),
		LastUsedAt:          nullableTime(row.LastUsedAt),
		CreatedAt:           row.CreatedAt.Time,
	}
}

func textValue(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func nullableText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func nullableTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func nullableInt64(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Int64
	return &result
}
