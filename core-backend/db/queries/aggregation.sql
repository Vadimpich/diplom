-- name: GetAggregationReadiness :one
SELECT
    examinations.id AS examination_id,
    examinations.specialist_id,
    examinations.status,
    COUNT(channel_runs.id) AS channels_total,
    COUNT(*) FILTER (WHERE channel_runs.status = 'succeeded') AS channels_succeeded,
    EXISTS (
        SELECT 1
        FROM aggregated_examination_profiles profiles
        WHERE profiles.examination_id = examinations.id
    ) AS already_aggregated
FROM examinations
LEFT JOIN examination_channel_runs channel_runs
    ON channel_runs.examination_id = examinations.id
WHERE examinations.id = $1
GROUP BY examinations.id;

-- name: MarkExaminationAggregating :one
UPDATE examinations
SET
    status = 'aggregating',
    updated_at = NOW()
WHERE id = $1
RETURNING id, specialist_id, created_by_user_id, questionnaire_id, status, created_at, started_at, finished_at, updated_at;

-- name: MarkExaminationAggregated :one
UPDATE examinations
SET
    status = 'aggregated',
    updated_at = NOW()
WHERE id = $1
RETURNING id, specialist_id, created_by_user_id, questionnaire_id, status, created_at, started_at, finished_at, updated_at;

-- name: CreateAggregatedProfile :one
INSERT INTO aggregated_examination_profiles (
    examination_id,
    specialist_id,
    schema_version,
    aggregation_version,
    baseline_algorithm_version,
    status,
    generated_at,
    baseline_refreshed_at,
    baseline_exam_count,
    overall_score,
    overall_band,
    primary_metric_key,
    neutral_recommendation_placeholder
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING
    id,
    examination_id,
    specialist_id,
    schema_version,
    aggregation_version,
    baseline_algorithm_version,
    status,
    generated_at,
    baseline_refreshed_at,
    baseline_exam_count,
    overall_score,
    overall_band,
    primary_metric_key,
    neutral_recommendation_placeholder,
    created_at,
    updated_at;

-- name: InsertAggregatedProfileMetric :one
INSERT INTO aggregated_profile_metrics (
    profile_id,
    metric_key,
    label,
    value,
    scale,
    direction
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, profile_id, metric_key, label, value, scale, direction, created_at;

-- name: InsertAggregatedProfileContribution :one
INSERT INTO aggregated_profile_channel_contributions (
    profile_id,
    channel,
    metric_key,
    weight,
    contribution,
    evidence_keys
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, profile_id, channel, metric_key, weight, contribution, evidence_keys, created_at;

-- name: InsertAggregatedProfileExplanation :one
INSERT INTO aggregated_profile_explanations (
    profile_id,
    position,
    kind,
    text
)
VALUES ($1, $2, $3, $4)
RETURNING id, profile_id, position, kind, text, created_at;

-- name: GetSpecialistBaselineState :one
SELECT
    id,
    specialist_id,
    algorithm_version,
    refreshed_at,
    baseline_exam_count,
    update_eligible,
    update_reason,
    metrics,
    created_at,
    updated_at
FROM specialist_baseline_states
WHERE specialist_id = $1;

-- name: UpsertSpecialistBaselineState :one
INSERT INTO specialist_baseline_states (
    specialist_id,
    algorithm_version,
    refreshed_at,
    baseline_exam_count,
    update_eligible,
    update_reason,
    metrics
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (specialist_id) DO UPDATE
SET
    algorithm_version = EXCLUDED.algorithm_version,
    refreshed_at = EXCLUDED.refreshed_at,
    baseline_exam_count = EXCLUDED.baseline_exam_count,
    update_eligible = EXCLUDED.update_eligible,
    update_reason = EXCLUDED.update_reason,
    metrics = EXCLUDED.metrics,
    updated_at = NOW()
RETURNING
    id,
    specialist_id,
    algorithm_version,
    refreshed_at,
    baseline_exam_count,
    update_eligible,
    update_reason,
    metrics,
    created_at,
    updated_at;

-- name: CreateExaminationBaselineSnapshot :one
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
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING
    id,
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
    update_reason,
    created_at;

-- name: ListSpecialistResultHistory :many
SELECT
    profiles.examination_id,
    profiles.specialist_id,
    profiles.generated_at,
    profiles.status,
    profiles.schema_version,
    profiles.aggregation_version,
    profiles.overall_score,
    profiles.overall_band,
    profiles.primary_metric_key,
    baseline.algorithm_version AS baseline_algorithm_version,
    baseline.refreshed_at AS baseline_refreshed_at,
    baseline.general_delta,
    baseline.personal_delta,
    baseline.baseline_exam_count
FROM aggregated_examination_profiles profiles
LEFT JOIN examination_baseline_snapshots baseline
    ON baseline.profile_id = profiles.id
WHERE profiles.specialist_id = $1
ORDER BY profiles.generated_at DESC, profiles.id DESC;
