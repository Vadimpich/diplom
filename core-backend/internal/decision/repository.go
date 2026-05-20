package decision

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sqlcdb "diplom/db/sqlc"
	"diplom/internal/aggregation"
	"diplom/internal/processing"
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
	DecisionCode   string
	RiskClass      string
	Patterns       []string
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
	channelPayloads, err := r.loadChannelPayloads(ctx, examinationID)
	if err != nil {
		return DecisionInput{}, err
	}
	baselineMetricScores, err := r.loadBaselineMetricScores(ctx, examinationID)
	if err != nil {
		return DecisionInput{}, err
	}
	payload, err := aggregation.BuildDecisionPayload(profile, channelPayloads, baselineMetricScores)
	if err != nil {
		return DecisionInput{}, err
	}
	return NewDecisionInput(profile, payload, DecisionServiceMetadata{
		TargetSystem: "kesmi",
		DeliveryMode: "canonical_kesmi_payload",
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
	if _, err := r.pool.Exec(ctx, `
UPDATE examinations
SET status = 'completed', finished_at = $2, updated_at = NOW()
WHERE id = $1`, input.ExaminationID, input.CompletedAt.UTC()); err != nil {
		return err
	}
	return r.applyBaselineUpdate(ctx, input.ExaminationID, input.Recommendation, input.CompletedAt.UTC(), true)
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
	if _, err := r.pool.Exec(ctx, `
UPDATE examinations
SET status = 'completed', finished_at = $2, updated_at = NOW()
WHERE id = $1`, input.ExaminationID, input.CompletedAt.UTC()); err != nil {
		return err
	}
	return r.applyBaselineUpdate(ctx, input.ExaminationID, input.Recommendation, input.CompletedAt.UTC(), false)
}

type baselineCandidateRow struct {
	SpecialistID        int64
	AlgorithmVersion    string
	BaselineExamCount   int
	UpdateEligible      bool
	UpdateReason        *string
	OverallBand         string
	CandidateMetricsRaw []byte
}

func (r *SQLRepository) applyBaselineUpdate(
	ctx context.Context,
	examinationID int64,
	recommendation string,
	completedAt time.Time,
	decisionSucceeded bool,
) error {
	const query = `
SELECT
    profiles.specialist_id,
    baseline.algorithm_version,
    baseline.baseline_exam_count,
    baseline.update_eligible,
    baseline.update_reason,
    profiles.overall_band,
    baseline.candidate_metrics
FROM examination_baseline_snapshots baseline
JOIN aggregated_examination_profiles profiles
    ON profiles.id = baseline.profile_id
WHERE baseline.examination_id = $1`

	var row baselineCandidateRow
	err := r.pool.QueryRow(ctx, query, examinationID).Scan(
		&row.SpecialistID,
		&row.AlgorithmVersion,
		&row.BaselineExamCount,
		&row.UpdateEligible,
		&row.UpdateReason,
		&row.OverallBand,
		&row.CandidateMetricsRaw,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}

	finalEligible := row.UpdateEligible && decisionSucceeded && !blocksBaselineRecommendation(recommendation)
	finalReason := "accepted"
	if !row.UpdateEligible {
		finalReason = stringOrDefault(row.UpdateReason, "baseline_ineligible")
	} else if !decisionSucceeded {
		finalReason = "decision_not_succeeded"
	} else if blocksBaselineRecommendation(recommendation) {
		finalReason = "decision_blocked"
	}

	finalExamCount := row.BaselineExamCount
	if finalEligible {
		metrics, err := extractBaselineMetrics(row.CandidateMetricsRaw)
		if err != nil {
			return err
		}
		metricsJSON, err := json.Marshal(metrics)
		if err != nil {
			return err
		}
		finalExamCount = baselineExamCount(metrics)
		const upsertState = `
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
		if _, err := r.pool.Exec(
			ctx,
			upsertState,
			row.SpecialistID,
			row.AlgorithmVersion,
			completedAt,
			finalExamCount,
			true,
			finalReason,
			metricsJSON,
		); err != nil {
			return err
		}
	}

	const updateSnapshot = `
UPDATE examination_baseline_snapshots
SET
    baseline_exam_count = $2,
    update_eligible = $3,
    update_reason = $4
WHERE examination_id = $1`
	_, err = r.pool.Exec(ctx, updateSnapshot, examinationID, finalExamCount, finalEligible, finalReason)
	return err
}

func blocksBaselineRecommendation(recommendation string) bool {
	return recommendation == DecisionRecommendationDenied
}

func extractBaselineMetrics(raw []byte) (map[string]aggregation.ExistingBaselineMetric, error) {
	var payload aggregation.NextBaseline
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("decode next baseline candidate: %w", err)
	}
	return payload.Metrics, nil
}

func baselineExamCount(metrics map[string]aggregation.ExistingBaselineMetric) int {
	count := 0
	for _, item := range metrics {
		if item.SampleCount > count {
			count = item.SampleCount
		}
	}
	return count
}

func stringOrDefault(value *string, fallback string) string {
	if value == nil || *value == "" {
		return fallback
	}
	return *value
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
	b.general_baseline_available,
	b.general_baseline_source,
	b.general_reference_population_version,
	b.personal_delta,
	b.personal_band,
	b.personal_baseline_available,
	b.personal_baseline_source,
	b.baseline_exam_count,
	b.update_eligible,
	b.data_reliability,
	b.significant_deviations
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
	var generalAvailable, personalAvailable *bool
	var generalSource, personalSource *string
	var baselineExamCount *int
	var updateEligible *bool
	var dataReliability *float64
	var significantDeviationsRaw []byte
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
		&generalAvailable,
		&generalSource,
		&generalRef,
		&personalDelta,
		&personalBand,
		&personalAvailable,
		&personalSource,
		&baselineExamCount,
		&updateEligible,
		&dataReliability,
		&significantDeviationsRaw,
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
	if generalAvailable != nil {
		profile.BaselineSnapshot.General.BaselineAvailable = *generalAvailable
	}
	if generalSource != nil {
		profile.BaselineSnapshot.General.BaselineSource = *generalSource
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
	if personalAvailable != nil {
		profile.BaselineSnapshot.Personal.BaselineAvailable = *personalAvailable
	}
	if personalSource != nil {
		profile.BaselineSnapshot.Personal.BaselineSource = *personalSource
	}
	if baselineExamCount != nil {
		profile.BaselineSnapshot.Personal.BaselineExamCount = *baselineExamCount
	}
	if updateEligible != nil {
		profile.BaselineSnapshot.Personal.UpdateEligible = *updateEligible
	}
	if dataReliability != nil {
		profile.BaselineSnapshot.Personal.DataReliability = *dataReliability
	}
	if len(significantDeviationsRaw) > 0 {
		if err := json.Unmarshal(significantDeviationsRaw, &profile.BaselineSnapshot.Personal.SignificantDeviations); err != nil {
			return aggregation.AggregatedProfile{}, err
		}
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

func (r *SQLRepository) loadChannelPayloads(ctx context.Context, examinationID int64) (aggregation.ChannelPayloadMap, error) {
	const query = `
SELECT channel, payload
FROM channel_results
WHERE examination_id = $1
  AND status = 'succeeded'
ORDER BY completed_at DESC`

	rows, err := r.pool.Query(ctx, query, examinationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make(aggregation.ChannelPayloadMap, len(processing.MandatoryChannels))
	for rows.Next() {
		var channel string
		var raw []byte
		if err := rows.Scan(&channel, &raw); err != nil {
			return nil, err
		}
		if _, exists := items[channel]; exists {
			continue
		}
		var payload processing.CanonicalChannelPayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, fmt.Errorf("decode channel payload %s: %w", channel, err)
		}
		items[channel] = payload
	}
	return items, rows.Err()
}

func (r *SQLRepository) loadBaselineMetricScores(ctx context.Context, examinationID int64) (map[string]aggregation.BaselineMetricScore, error) {
	const query = `
SELECT baseline_metric_scores
FROM examination_baseline_snapshots
WHERE examination_id = $1`

	var raw []byte
	err := r.pool.QueryRow(ctx, query, examinationID).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return map[string]aggregation.BaselineMetricScore{}, nil
		}
		return nil, err
	}
	if len(raw) == 0 {
		return map[string]aggregation.BaselineMetricScore{}, nil
	}
	var scores map[string]aggregation.BaselineMetricScore
	if err := json.Unmarshal(raw, &scores); err != nil {
		return nil, fmt.Errorf("decode baseline metric scores: %w", err)
	}
	return scores, nil
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
