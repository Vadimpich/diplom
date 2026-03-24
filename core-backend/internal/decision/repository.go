package decision

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	sqlcdb "diplom/db/sqlc"
	"diplom/internal/aggregation"
	"diplom/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLRepository struct {
	pool    *pgxpool.Pool
	queries *sqlcdb.Queries
}

type FinalizeInput struct {
	SnapshotID     int64
	ExaminationID  int64
	State          string
	Recommendation string
	Message        string
	CorrelationID  string
	AttemptCount   int32
	LastAttemptAt  time.Time
	CompletedAt    time.Time
	Diagnostics    DecisionDiagnostics
	RawResponse    []byte
}

func NewRepository(pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{pool: pool, queries: sqlcdb.New(pool)}
}

func (r *SQLRepository) CreatePendingSnapshot(ctx context.Context, profile aggregation.AggregatedProfile, _ DecisionInput, maxAttempts int32) (Snapshot, error) {
	existing, err := r.queries.GetDecisionSnapshotByExamination(ctx, profile.ExaminationID)
	if err == nil {
		return snapshotFromDB(existing), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Snapshot{}, err
	}

	diagnostics, _ := json.Marshal(DecisionDiagnostics{})
	row, err := r.queries.CreateDecisionSnapshot(ctx, sqlcdb.CreateDecisionSnapshotParams{
		ExaminationID:      profile.ExaminationID,
		SpecialistID:       profile.SpecialistID,
		Status:             DecisionStatePending,
		PayloadVersion:     PayloadVersionV1,
		AggregationVersion: profile.AggregationVersion,
		Recommendation:     DecisionRecommendationUnavailable,
		Message:            PlaceholderMessage,
		MaxAttempts:        maxAttempts,
		AttemptCount:       0,
		DiagnosticsJson:    diagnostics,
	})
	if err != nil {
		return Snapshot{}, err
	}
	if _, err := r.pool.Exec(ctx, `
UPDATE examinations
SET status = 'decision_pending', updated_at = NOW()
WHERE id = $1 AND status IN ('aggregated', 'decision_pending')`, profile.ExaminationID); err != nil {
		return Snapshot{}, err
	}
	return snapshotFromDB(row), nil
}

func (r *SQLRepository) ListPendingSnapshots(ctx context.Context) ([]Snapshot, error) {
	rows, err := r.queries.ListPendingDecisionSnapshots(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]Snapshot, 0, len(rows))
	for _, row := range rows {
		items = append(items, snapshotFromDB(row))
	}
	return items, nil
}

func (r *SQLRepository) LoadDecisionInput(ctx context.Context, examinationID int64) (DecisionInput, error) {
	profile, err := r.loadAggregatedProfile(ctx, examinationID)
	if err != nil {
		return DecisionInput{}, err
	}
	return NewDecisionInput(profile, DecisionServiceMetadata{
		TargetSystem: "kesmi",
		DeliveryMode: "placeholder",
		Message:      PlaceholderMessage,
	}), nil
}

func (r *SQLRepository) AppendAttempt(ctx context.Context, input AttemptRecord) error {
	params := sqlcdb.InsertDecisionAttemptParams{
		DecisionSnapshotID: input.SnapshotID,
		AttemptNumber:      input.AttemptNumber,
		RequestPayloadJson: input.RequestPayload,
		Retryable:          input.Retryable,
		StartedAt:          pgtype.Timestamptz{Time: input.StartedAt.UTC(), Valid: true},
		FinishedAt:         pgtype.Timestamptz{Time: input.FinishedAt.UTC(), Valid: true},
	}
	if input.CorrelationID != "" {
		params.CorrelationID = pgtype.Text{String: input.CorrelationID, Valid: true}
	}
	if len(input.ResponsePayload) > 0 {
		params.ResponsePayloadJson = input.ResponsePayload
	}
	if input.ErrorCode != "" {
		params.ErrorCode = pgtype.Text{String: input.ErrorCode, Valid: true}
	}
	if input.ErrorMessage != "" {
		params.ErrorMessage = pgtype.Text{String: input.ErrorMessage, Valid: true}
	}
	if input.ErrorClass != "" {
		params.ErrorClass = pgtype.Text{String: input.ErrorClass, Valid: true}
	}
	if input.HTTPStatus > 0 {
		params.HttpStatus = pgtype.Int4{Int32: int32(input.HTTPStatus), Valid: true}
	}
	_, err := r.queries.InsertDecisionAttempt(ctx, params)
	return err
}

func (r *SQLRepository) MarkRetryPending(ctx context.Context, input RetryPendingInput) error {
	diagnostics, err := json.Marshal(input.Diagnostics)
	if err != nil {
		return err
	}
	if _, err := r.queries.IncrementDecisionAttemptCount(ctx, sqlcdb.IncrementDecisionAttemptCountParams{
		ID:            input.SnapshotID,
		CorrelationID: pgtype.Text{String: input.CorrelationID, Valid: input.CorrelationID != ""},
		LastAttemptAt: pgtype.Timestamptz{Time: input.LastAttemptAt.UTC(), Valid: true},
	}); err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
UPDATE decision_snapshots
SET diagnostics_json = $2, updated_at = NOW()
WHERE id = $1`, input.SnapshotID, diagnostics)
	return err
}

func (r *SQLRepository) MarkSucceeded(ctx context.Context, input FinalizeInput) error {
	diagnostics, err := json.Marshal(input.Diagnostics)
	if err != nil {
		return err
	}
	if _, err := r.queries.MarkDecisionAttemptSucceeded(ctx, sqlcdb.MarkDecisionAttemptSucceededParams{
		ID:              input.SnapshotID,
		Recommendation:  input.Recommendation,
		Message:         input.Message,
		CorrelationID:   pgtype.Text{String: input.CorrelationID, Valid: input.CorrelationID != ""},
		AttemptCount:    input.AttemptCount,
		LastAttemptAt:   pgtype.Timestamptz{Time: input.LastAttemptAt.UTC(), Valid: true},
		CompletedAt:     pgtype.Timestamptz{Time: input.CompletedAt.UTC(), Valid: true},
		DiagnosticsJson: diagnostics,
		RawResponseJson: input.RawResponse,
	}); err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
UPDATE examinations
SET status = 'completed', finished_at = $2, updated_at = NOW()
WHERE id = $1`, input.ExaminationID, input.CompletedAt.UTC())
	return err
}

func (r *SQLRepository) MarkFailed(ctx context.Context, input FinalizeInput) error {
	diagnostics, err := json.Marshal(input.Diagnostics)
	if err != nil {
		return err
	}
	if _, err := r.queries.MarkDecisionAttemptFailed(ctx, sqlcdb.MarkDecisionAttemptFailedParams{
		ID:              input.SnapshotID,
		Status:          input.State,
		Recommendation:  input.Recommendation,
		Message:         input.Message,
		CorrelationID:   pgtype.Text{String: input.CorrelationID, Valid: input.CorrelationID != ""},
		AttemptCount:    input.AttemptCount,
		LastAttemptAt:   pgtype.Timestamptz{Time: input.LastAttemptAt.UTC(), Valid: true},
		CompletedAt:     pgtype.Timestamptz{Time: input.CompletedAt.UTC(), Valid: true},
		DiagnosticsJson: diagnostics,
		RawResponseJson: input.RawResponse,
	}); err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
UPDATE examinations
SET status = 'completed', finished_at = $2, updated_at = NOW()
WHERE id = $1`, input.ExaminationID, input.CompletedAt.UTC())
	return err
}

func snapshotFromDB(row sqlcdb.DecisionSnapshot) Snapshot {
	item := Snapshot{
		ID:                 row.ID,
		ExaminationID:      row.ExaminationID,
		SpecialistID:       row.SpecialistID,
		State:              row.Status,
		PayloadVersion:     row.PayloadVersion,
		AggregationVersion: row.AggregationVersion,
		Recommendation:     row.Recommendation,
		Message:            row.Message,
		AttemptCount:       row.AttemptCount,
		MaxAttempts:        row.MaxAttempts,
	}
	if row.CorrelationID.Valid {
		item.CorrelationID = row.CorrelationID.String
	}
	if row.LastAttemptAt.Valid {
		value := row.LastAttemptAt.Time.UTC()
		item.LastAttemptAt = &value
	}
	return item
}

func (r *SQLRepository) loadAggregatedProfile(ctx context.Context, examinationID int64) (aggregation.AggregatedProfile, error) {
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

	profile.Metrics, err = r.loadMetrics(ctx, examinationID)
	if err != nil {
		return aggregation.AggregatedProfile{}, err
	}
	profile.ChannelContributions, err = r.loadContributions(ctx, examinationID)
	if err != nil {
		return aggregation.AggregatedProfile{}, err
	}
	profile.Explanations, err = r.loadExplanations(ctx, examinationID)
	if err != nil {
		return aggregation.AggregatedProfile{}, err
	}
	return profile, nil
}

func (r *SQLRepository) loadMetrics(ctx context.Context, examinationID int64) ([]aggregation.Metric, error) {
	rows, err := r.pool.Query(ctx, `
SELECT m.metric_key, m.label, m.value, m.scale, m.direction
FROM aggregated_profile_metrics m
JOIN aggregated_examination_profiles p ON p.id = m.profile_id
WHERE p.examination_id = $1
ORDER BY m.metric_key ASC`, examinationID)
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
	rows, err := r.pool.Query(ctx, `
SELECT c.channel, c.metric_key, c.weight, c.contribution, c.evidence_keys
FROM aggregated_profile_channel_contributions c
JOIN aggregated_examination_profiles p ON p.id = c.profile_id
WHERE p.examination_id = $1
ORDER BY c.contribution DESC, c.channel ASC`, examinationID)
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
	rows, err := r.pool.Query(ctx, `
SELECT e.position, e.kind, e.text
FROM aggregated_profile_explanations e
JOIN aggregated_examination_profiles p ON p.id = e.profile_id
WHERE p.examination_id = $1
ORDER BY e.position ASC`, examinationID)
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
