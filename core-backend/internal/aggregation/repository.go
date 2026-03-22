package aggregation

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"dimplom/internal/processing"
	"dimplom/internal/repository"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{pool: pool}
}

type dbtx interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func (r *SQLRepository) GetReadiness(ctx context.Context, examinationID int64) (Readiness, error) {
	const query = `
SELECT
	e.id,
	e.specialist_id,
	e.status,
	COUNT(*) FILTER (WHERE cr.status = 'succeeded') AS channels_succeeded,
	EXISTS (
		SELECT 1
		FROM aggregated_examination_profiles profiles
		WHERE profiles.examination_id = e.id
		  AND profiles.status = 'aggregated'
	) AS already_aggregated
FROM examinations e
LEFT JOIN examination_channel_runs cr
	ON cr.examination_id = e.id
WHERE e.id = $1
GROUP BY e.id, e.specialist_id, e.status`

	var readiness Readiness
	err := r.queryRow(ctx, query, examinationID).Scan(
		&readiness.ExaminationID,
		&readiness.SpecialistID,
		&readiness.Status,
		&readiness.ChannelsSucceeded,
		&readiness.AlreadyAggregated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Readiness{}, repository.ErrNotFound
		}
		return Readiness{}, err
	}
	return readiness, nil
}

func (r *SQLRepository) LoadSucceededResults(ctx context.Context, examinationID int64) ([]PersistedChannelResult, error) {
	const query = `
SELECT
	e.specialist_id,
	cr.examination_id,
	cr.channel,
	cr.model_version,
	cr.payload,
	cr.completed_at
FROM channel_results cr
JOIN examinations e
	ON e.id = cr.examination_id
JOIN (
	SELECT channel, MAX(completed_at) AS completed_at
	FROM channel_results
	WHERE examination_id = $1
	  AND status = 'succeeded'
	GROUP BY channel
) latest
	ON latest.channel = cr.channel
	AND latest.completed_at = cr.completed_at
WHERE cr.examination_id = $1
  AND cr.status = 'succeeded'
ORDER BY cr.channel ASC`

	rows, err := r.query(ctx, query, examinationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]PersistedChannelResult, 0, len(processing.MandatoryChannels))
	for rows.Next() {
		var item PersistedChannelResult
		if err := rows.Scan(
			&item.SpecialistID,
			&item.ExaminationID,
			&item.Channel,
			&item.ModelVersion,
			&item.Payload,
			&item.CompletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}

func (r *SQLRepository) SaveAggregatingProfile(ctx context.Context, input PersistInput) (bool, error) {
	tx, err := r.begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	profile := input.Profile
	const statusQuery = `
UPDATE examinations
SET status = 'aggregating', updated_at = NOW()
WHERE id = $1
  AND status IN ('processing', 'ready_for_processing', 'aggregating')`
	if _, err := tx.Exec(ctx, statusQuery, profile.ExaminationID); err != nil {
		return false, err
	}

	const insertProfile = `
INSERT INTO aggregated_examination_profiles (
	examination_id,
	specialist_id,
	schema_version,
	aggregation_version,
	baseline_algorithm_version,
	status,
	generated_at,
	baseline_exam_count,
	overall_score,
	overall_band,
	primary_metric_key,
	neutral_recommendation_placeholder
)
VALUES ($1,$2,$3,$4,$5,'aggregating',$6,0,$7,$8,$9,$10)
ON CONFLICT (examination_id) DO NOTHING
RETURNING id`

	var profileID int64
	err = tx.QueryRow(
		ctx,
		insertProfile,
		profile.ExaminationID,
		profile.SpecialistID,
		profile.SchemaVersion,
		profile.AggregationVersion,
		profile.BaselineSnapshot.AlgorithmVersion,
		profile.GeneratedAt,
		profile.Summary.OverallScore,
		profile.Summary.OverallBand,
		profile.Summary.PrimaryMetricKey,
		profile.Summary.NeutralRecommendationPlaceholder,
	).Scan(&profileID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return false, tx.Commit(ctx)
	}

	const metricQuery = `
INSERT INTO aggregated_profile_metrics (profile_id, metric_key, label, value, scale, direction)
VALUES ($1,$2,$3,$4,$5,$6)`
	for _, metric := range profile.Metrics {
		if _, err := tx.Exec(ctx, metricQuery, profileID, metric.Key, metric.Label, metric.Value, metric.Scale, metric.Direction); err != nil {
			return false, err
		}
	}

	const contributionQuery = `
INSERT INTO aggregated_profile_channel_contributions (profile_id, channel, metric_key, weight, contribution, evidence_keys)
VALUES ($1,$2,$3,$4,$5,$6::jsonb)`
	for _, item := range profile.ChannelContributions {
		evidence, _ := json.Marshal(item.EvidenceKeys)
		if _, err := tx.Exec(ctx, contributionQuery, profileID, item.Channel, item.MetricKey, item.Weight, item.Contribution, evidence); err != nil {
			return false, err
		}
	}

	const explanationQuery = `
INSERT INTO aggregated_profile_explanations (profile_id, position, kind, text)
VALUES ($1,$2,$3,$4)`
	for _, item := range profile.Explanations {
		if _, err := tx.Exec(ctx, explanationQuery, profileID, item.Position, item.Kind, item.Text); err != nil {
			return false, err
		}
	}

	return true, tx.Commit(ctx)
}

func (r *SQLRepository) LoadBaselineHistory(ctx context.Context, specialistID int64, currentExaminationID int64) (BaselineHistory, error) {
	const query = `
SELECT
	p.examination_id,
	p.generated_at,
	m.metric_key,
	m.value
FROM aggregated_examination_profiles p
JOIN aggregated_profile_metrics m
	ON m.profile_id = p.id
WHERE p.specialist_id = $1
  AND p.status = 'aggregated'
  AND p.examination_id <> $2
ORDER BY p.generated_at ASC, p.id ASC, m.metric_key ASC`

	rows, err := r.query(ctx, query, specialistID, currentExaminationID)
	if err != nil {
		return BaselineHistory{}, err
	}
	defer rows.Close()

	vectors := make([]BaselineHistoryVector, 0)
	indexByExam := map[int64]int{}
	for rows.Next() {
		var examinationID int64
		var generatedAt time.Time
		var key string
		var value float64
		if err := rows.Scan(&examinationID, &generatedAt, &key, &value); err != nil {
			return BaselineHistory{}, err
		}
		idx, ok := indexByExam[examinationID]
		if !ok {
			idx = len(vectors)
			indexByExam[examinationID] = idx
			vectors = append(vectors, BaselineHistoryVector{
				ExaminationID: examinationID,
				GeneratedAt:   generatedAt.UTC(),
				Metrics:       []BaselineMetricValue{},
			})
		}
		vectors[idx].Metrics = append(vectors[idx].Metrics, BaselineMetricValue{Key: key, Value: value})
	}
	if err := rows.Err(); err != nil {
		return BaselineHistory{}, err
	}

	return BaselineHistory{
		BaselineExamCount: len(vectors),
		MetricVectors:     vectors,
	}, nil
}

func (r *SQLRepository) FinalizeAggregatedProfile(ctx context.Context, input FinalizeInput) error {
	tx, err := r.begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const selectProfile = `
SELECT id
FROM aggregated_examination_profiles
WHERE examination_id = $1
FOR UPDATE`
	var profileID int64
	if err := tx.QueryRow(ctx, selectProfile, input.Profile.ExaminationID).Scan(&profileID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrNotFound
		}
		return err
	}

	nextBaseline, err := json.Marshal(input.Baseline.NextBaseline)
	if err != nil {
		return err
	}

	const updateProfile = `
UPDATE aggregated_examination_profiles
SET
	status = 'aggregated',
	baseline_algorithm_version = $2,
	baseline_refreshed_at = $3,
	baseline_exam_count = $4,
	updated_at = NOW()
WHERE id = $1`
	if _, err := tx.Exec(
		ctx,
		updateProfile,
		profileID,
		input.Baseline.AlgorithmVersion,
		input.Baseline.RefreshedAt.UTC(),
		input.Baseline.UpdateEligibility.BaselineExamCountAfterUpdate,
	); err != nil {
		return err
	}

	const snapshotQuery = `
INSERT INTO examination_baseline_snapshots (
	examination_id,
	profile_id,
	algorithm_version,
	refreshed_at,
	general_delta,
	general_band,
	general_reference_population_version,
	personal_delta,
	personal_band,
	baseline_exam_count,
	update_eligible,
	update_reason
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
ON CONFLICT (examination_id) DO NOTHING`
	if _, err := tx.Exec(
		ctx,
		snapshotQuery,
		input.Profile.ExaminationID,
		profileID,
		input.Baseline.AlgorithmVersion,
		input.Baseline.RefreshedAt.UTC(),
		input.Baseline.GeneralDeviation.Score,
		input.Baseline.GeneralDeviation.Band,
		"general-v1",
		input.Baseline.PersonalDeviation.Score,
		input.Baseline.PersonalDeviation.Band,
		input.Baseline.UpdateEligibility.BaselineExamCountAfterUpdate,
		input.Baseline.UpdateEligibility.Eligible,
		input.Baseline.UpdateEligibility.Reason,
	); err != nil {
		return err
	}

	const baselineStateQuery = `
INSERT INTO specialist_baseline_states (
	specialist_id,
	algorithm_version,
	refreshed_at,
	baseline_exam_count,
	update_eligible,
	update_reason,
	metrics
)
VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb)
ON CONFLICT (specialist_id) DO UPDATE
SET
	algorithm_version = EXCLUDED.algorithm_version,
	refreshed_at = EXCLUDED.refreshed_at,
	baseline_exam_count = EXCLUDED.baseline_exam_count,
	update_eligible = EXCLUDED.update_eligible,
	update_reason = EXCLUDED.update_reason,
	metrics = EXCLUDED.metrics,
	updated_at = NOW()`
	if _, err := tx.Exec(
		ctx,
		baselineStateQuery,
		input.Profile.SpecialistID,
		input.Baseline.AlgorithmVersion,
		input.Baseline.RefreshedAt.UTC(),
		input.Baseline.UpdateEligibility.BaselineExamCountAfterUpdate,
		input.Baseline.UpdateEligibility.Eligible,
		input.Baseline.UpdateEligibility.Reason,
		nextBaseline,
	); err != nil {
		return err
	}

	const markAggregated = `
UPDATE examinations
SET
	status = 'aggregated',
	finished_at = $2,
	updated_at = NOW()
WHERE id = $1`
	if _, err := tx.Exec(ctx, markAggregated, input.Profile.ExaminationID, input.Baseline.RefreshedAt.UTC()); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *SQLRepository) begin(ctx context.Context) (pgx.Tx, error) {
	if tx, ok := txFromContext(ctx); ok {
		return tx, nil
	}
	return r.pool.BeginTx(ctx, pgx.TxOptions{})
}

func (r *SQLRepository) query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	if tx, ok := txFromContext(ctx); ok {
		return tx.Query(ctx, query, args...)
	}
	return r.pool.Query(ctx, query, args...)
}

func (r *SQLRepository) queryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if tx, ok := txFromContext(ctx); ok {
		return tx.QueryRow(ctx, query, args...)
	}
	return r.pool.QueryRow(ctx, query, args...)
}

type txContextKey struct{}

func txFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txContextKey{}).(pgx.Tx)
	return tx, ok
}
