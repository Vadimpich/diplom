-- name: CreateSpecialist :one
INSERT INTO specialists (
    full_name,
    personnel_number
) VALUES (
    $1, $2
)
RETURNING id, full_name, personnel_number, created_at, updated_at;

-- name: ListSpecialists :many
SELECT
    s.id,
    s.full_name,
    s.personnel_number,
    COALESCE(exam_stats.examinations_count, 0)::BIGINT AS examinations_count,
    COALESCE(last_exam.id, 0)::BIGINT AS last_examination_id,
    COALESCE(last_exam.activity_at, 'epoch'::timestamptz) AS last_examination_at,
    COALESCE(last_exam.status, '') AS last_examination_status,
    profile.overall_score AS last_overall_score,
    profile.overall_band AS last_overall_band,
    COALESCE(baseline.baseline_exam_count, 0)::INTEGER AS baseline_exam_count,
    baseline.refreshed_at AS baseline_refreshed_at,
    s.created_at,
    s.updated_at
FROM specialists s
LEFT JOIN (
    SELECT
        e.specialist_id,
        COUNT(*)::BIGINT AS examinations_count
    FROM examinations e
    GROUP BY e.specialist_id
) exam_stats ON exam_stats.specialist_id = s.id
LEFT JOIN LATERAL (
    SELECT
        e.id,
        COALESCE(e.finished_at, e.started_at, e.created_at) AS activity_at,
        e.status
    FROM examinations e
    WHERE e.specialist_id = s.id
    ORDER BY COALESCE(e.finished_at, e.started_at, e.created_at) DESC, e.id DESC
    LIMIT 1
) last_exam ON TRUE
LEFT JOIN aggregated_examination_profiles profile ON profile.examination_id = last_exam.id
LEFT JOIN specialist_baseline_states baseline ON baseline.specialist_id = s.id
ORDER BY s.id DESC;

-- name: GetSpecialistByID :one
SELECT
    s.id,
    s.full_name,
    s.personnel_number,
    COALESCE(exam_stats.examinations_count, 0)::BIGINT AS examinations_count,
    COALESCE(last_exam.id, 0)::BIGINT AS last_examination_id,
    COALESCE(last_exam.activity_at, 'epoch'::timestamptz) AS last_examination_at,
    COALESCE(last_exam.status, '') AS last_examination_status,
    profile.overall_score AS last_overall_score,
    profile.overall_band AS last_overall_band,
    COALESCE(baseline.baseline_exam_count, 0)::INTEGER AS baseline_exam_count,
    baseline.refreshed_at AS baseline_refreshed_at,
    s.created_at,
    s.updated_at
FROM specialists s
LEFT JOIN (
    SELECT
        e.specialist_id,
        COUNT(*)::BIGINT AS examinations_count
    FROM examinations e
    GROUP BY e.specialist_id
) exam_stats ON exam_stats.specialist_id = s.id
LEFT JOIN LATERAL (
    SELECT
        e.id,
        COALESCE(e.finished_at, e.started_at, e.created_at) AS activity_at,
        e.status
    FROM examinations e
    WHERE e.specialist_id = s.id
    ORDER BY COALESCE(e.finished_at, e.started_at, e.created_at) DESC, e.id DESC
    LIMIT 1
) last_exam ON TRUE
LEFT JOIN aggregated_examination_profiles profile ON profile.examination_id = last_exam.id
LEFT JOIN specialist_baseline_states baseline ON baseline.specialist_id = s.id
WHERE s.id = $1;

-- name: UpdateSpecialist :one
UPDATE specialists
SET
    full_name = $2,
    personnel_number = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING id, full_name, personnel_number, created_at, updated_at;

-- name: DeleteSpecialist :execrows
DELETE FROM specialists
WHERE id = $1;

