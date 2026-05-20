package settings

import (
	"context"
	"errors"
	"time"

	"diplom/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{pool: pool}
}

func (r *SQLRepository) Get(ctx context.Context) (RuntimeSettings, error) {
	const query = `
SELECT
	audio_retention_ttl_days,
	processing_max_attempts,
	kesmi_max_retries,
	created_at,
	updated_at
FROM system_settings
WHERE singleton = TRUE`

	var item RuntimeSettings
	if err := r.pool.QueryRow(ctx, query).Scan(
		&item.AudioRetentionTTLDays,
		&item.ProcessingMaxAttempts,
		&item.KESMIMaxRetries,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RuntimeSettings{}, repository.ErrNotFound
		}
		return RuntimeSettings{}, err
	}
	return item, nil
}

func (r *SQLRepository) Ensure(ctx context.Context, defaults Defaults) (RuntimeSettings, error) {
	const query = `
INSERT INTO system_settings (
	singleton,
	audio_retention_ttl_days,
	processing_max_attempts,
	kesmi_max_retries
)
VALUES (TRUE, $1, $2, $3)
ON CONFLICT (singleton) DO NOTHING`

	if _, err := r.pool.Exec(ctx, query, defaults.AudioRetentionTTLDays, defaults.ProcessingMaxAttempts, defaults.KESMIMaxRetries); err != nil {
		return RuntimeSettings{}, err
	}
	return r.Get(ctx)
}

func (r *SQLRepository) Update(ctx context.Context, input UpdateInput) (RuntimeSettings, error) {
	const query = `
UPDATE system_settings
SET
	audio_retention_ttl_days = $1,
	processing_max_attempts = $2,
	kesmi_max_retries = $3,
	updated_at = $4
WHERE singleton = TRUE
RETURNING audio_retention_ttl_days, processing_max_attempts, kesmi_max_retries, created_at, updated_at`

	var item RuntimeSettings
	now := time.Now().UTC()
	if err := r.pool.QueryRow(ctx, query, input.AudioRetentionTTLDays, input.ProcessingMaxAttempts, input.KESMIMaxRetries, now).Scan(
		&item.AudioRetentionTTLDays,
		&item.ProcessingMaxAttempts,
		&item.KESMIMaxRetries,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RuntimeSettings{}, repository.ErrNotFound
		}
		return RuntimeSettings{}, err
	}
	return item, nil
}
