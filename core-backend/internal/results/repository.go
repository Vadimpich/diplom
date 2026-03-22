package results

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"dimplom/internal/aggregation"
	"dimplom/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{pool: pool}
}

func (r *SQLRepository) GetExaminationResult(ctx context.Context, examinationID int64) (aggregation.AggregatedProfile, error) {
	const profileQuery = `
SELECT
	p.examination_id,
	p.specialist_id,
	p.schema_version,
	p.aggregation_version,
	p.status,
	p.generated_at,
	p.overall_score,
	p.overall_band,
	p.primary_metric_key,
	p.neutral_recommendation_placeholder,
	b.algorithm_version,
	b.refreshed_at,
	b.general_delta,
	b.general_band,
	b.general_reference_population_version,
	b.personal_delta,
	b.personal_band,
	b.baseline_exam_count,
	b.update_eligible
FROM aggregated_examination_profiles p
LEFT JOIN examination_baseline_snapshots b
	ON b.profile_id = p.id
WHERE p.examination_id = $1
  AND p.status = 'aggregated'`

	var profile aggregation.AggregatedProfile
	var generalRef *string
	var refreshedAt *time.Time
	var generalDelta, personalDelta *float64
	var generalBand, personalBand *string
	var baselineExamCount *int
	var updateEligible *bool
	err := r.pool.QueryRow(ctx, profileQuery, examinationID).Scan(
		&profile.ExaminationID,
		&profile.SpecialistID,
		&profile.SchemaVersion,
		&profile.AggregationVersion,
		&profile.Status,
		&profile.GeneratedAt,
		&profile.Summary.OverallScore,
		&profile.Summary.OverallBand,
		&profile.Summary.PrimaryMetricKey,
		&profile.Summary.NeutralRecommendationPlaceholder,
		&profile.BaselineSnapshot.AlgorithmVersion,
		&refreshedAt,
		&generalDelta,
		&generalBand,
		&generalRef,
		&personalDelta,
		&personalBand,
		&baselineExamCount,
		&updateEligible,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return aggregation.AggregatedProfile{}, repository.ErrNotFound
		}
		return aggregation.AggregatedProfile{}, err
	}
	if refreshedAt != nil {
		profile.BaselineSnapshot.RefreshedAt = refreshedAt.UTC()
	}
	if generalDelta != nil {
		profile.BaselineSnapshot.General.Delta = *generalDelta
	}
	if generalBand != nil {
		profile.BaselineSnapshot.General.Band = *generalBand
	}
	if generalRef != nil {
		profile.BaselineSnapshot.General.ReferencePopulationVersion = *generalRef
	}
	if personalDelta != nil {
		profile.BaselineSnapshot.Personal.Delta = *personalDelta
	}
	if personalBand != nil {
		profile.BaselineSnapshot.Personal.Band = *personalBand
	}
	if baselineExamCount != nil {
		profile.BaselineSnapshot.Personal.BaselineExamCount = *baselineExamCount
	}
	if updateEligible != nil {
		profile.BaselineSnapshot.Personal.UpdateEligible = *updateEligible
	}

	metrics, err := r.loadMetrics(ctx, examinationID)
	if err != nil {
		return aggregation.AggregatedProfile{}, err
	}
	profile.Metrics = metrics
	contributions, err := r.loadContributions(ctx, examinationID)
	if err != nil {
		return aggregation.AggregatedProfile{}, err
	}
	profile.ChannelContributions = contributions
	explanations, err := r.loadExplanations(ctx, examinationID)
	if err != nil {
		return aggregation.AggregatedProfile{}, err
	}
	profile.Explanations = explanations
	return profile, nil
}

func (r *SQLRepository) GetSpecialistHistory(ctx context.Context, specialistID int64) (SpecialistHistoryResponse, error) {
	const specialistQuery = `SELECT 1 FROM specialists WHERE id = $1`
	if err := r.pool.QueryRow(ctx, specialistQuery, specialistID).Scan(new(int)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SpecialistHistoryResponse{}, repository.ErrNotFound
		}
		return SpecialistHistoryResponse{}, err
	}

	const query = `
SELECT
	p.examination_id,
	p.generated_at,
	p.status,
	p.overall_score,
	p.overall_band,
	COALESCE(b.algorithm_version, ''),
	b.refreshed_at,
	COALESCE(b.general_delta, 0),
	COALESCE(b.personal_delta, 0),
	COALESCE(b.baseline_exam_count, 0)
FROM aggregated_examination_profiles p
LEFT JOIN examination_baseline_snapshots b
	ON b.profile_id = p.id
WHERE p.specialist_id = $1
  AND p.status = 'aggregated'
ORDER BY p.generated_at DESC, p.id DESC`
	rows, err := r.pool.Query(ctx, query, specialistID)
	if err != nil {
		return SpecialistHistoryResponse{}, err
	}
	defer rows.Close()

	items := make([]SpecialistHistoryItem, 0)
	previousMetrics := map[string]float64{}
	for rows.Next() {
		var item SpecialistHistoryItem
		var generatedAt time.Time
		var refreshedAt *time.Time
		if err := rows.Scan(
			&item.ExaminationID,
			&generatedAt,
			&item.Status,
			&item.Summary.OverallScore,
			&item.Summary.OverallBand,
			&item.BaselineSnapshot.AlgorithmVersion,
			&refreshedAt,
			&item.BaselineSnapshot.GeneralDelta,
			&item.BaselineSnapshot.PersonalDelta,
			&item.BaselineSnapshot.BaselineExamCount,
		); err != nil {
			return SpecialistHistoryResponse{}, err
		}
		item.GeneratedAt = generatedAt.UTC().Format(time.RFC3339)
		if refreshedAt != nil {
			item.BaselineSnapshot.RefreshedAt = refreshedAt.UTC().Format(time.RFC3339)
		}

		metrics, err := r.loadMetrics(ctx, item.ExaminationID)
		if err != nil {
			return SpecialistHistoryResponse{}, err
		}
		item.KeyMetrics = make([]HistoryMetric, 0, len(metrics))
		for _, metric := range metrics {
			record := HistoryMetric{
				Key:   metric.Key,
				Label: metric.Label,
				Value: metric.Value,
			}
			if previous, ok := previousMetrics[metric.Key]; ok {
				prev := previous
				delta := metric.Value - previous
				record.PreviousValue = &prev
				record.DeltaFromPrevious = &delta
			}
			previousMetrics[metric.Key] = metric.Value
			item.KeyMetrics = append(item.KeyMetrics, record)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return SpecialistHistoryResponse{}, err
	}

	return SpecialistHistoryResponse{
		SpecialistID: specialistID,
		Items:        items,
	}, nil
}

func (r *SQLRepository) loadMetrics(ctx context.Context, examinationID int64) ([]aggregation.Metric, error) {
	const query = `
SELECT m.metric_key, m.label, m.value, m.scale, m.direction
FROM aggregated_profile_metrics m
JOIN aggregated_examination_profiles p
	ON p.id = m.profile_id
WHERE p.examination_id = $1
ORDER BY m.metric_key ASC`
	rows, err := r.pool.Query(ctx, query, examinationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]aggregation.Metric, 0)
	for rows.Next() {
		var item aggregation.Metric
		if err := rows.Scan(&item.Key, &item.Label, &item.Value, &item.Scale, &item.Direction); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) loadContributions(ctx context.Context, examinationID int64) ([]aggregation.ChannelContribution, error) {
	const query = `
SELECT c.channel, c.metric_key, c.weight, c.contribution, c.evidence_keys
FROM aggregated_profile_channel_contributions c
JOIN aggregated_examination_profiles p
	ON p.id = c.profile_id
WHERE p.examination_id = $1
ORDER BY c.contribution DESC, c.channel ASC`
	rows, err := r.pool.Query(ctx, query, examinationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]aggregation.ChannelContribution, 0)
	for rows.Next() {
		var item aggregation.ChannelContribution
		var evidence []byte
		if err := rows.Scan(&item.Channel, &item.MetricKey, &item.Weight, &item.Contribution, &evidence); err != nil {
			return nil, err
		}
		if len(evidence) > 0 {
			if err := json.Unmarshal(evidence, &item.EvidenceKeys); err != nil {
				return nil, err
			}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) loadExplanations(ctx context.Context, examinationID int64) ([]aggregation.Explanation, error) {
	const query = `
SELECT e.position, e.kind, e.text
FROM aggregated_profile_explanations e
JOIN aggregated_examination_profiles p
	ON p.id = e.profile_id
WHERE p.examination_id = $1
ORDER BY e.position ASC`
	rows, err := r.pool.Query(ctx, query, examinationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]aggregation.Explanation, 0)
	for rows.Next() {
		var item aggregation.Explanation
		if err := rows.Scan(&item.Position, &item.Kind, &item.Text); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
