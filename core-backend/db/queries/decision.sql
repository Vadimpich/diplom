-- name: CreateDecisionSnapshot :one
INSERT INTO decision_snapshots (
    examination_id,
    specialist_id,
    status,
    payload_version,
    aggregation_version,
    recommendation,
    message,
    correlation_id,
    max_attempts,
    attempt_count,
    last_attempt_at,
    completed_at,
    diagnostics_json,
    raw_response_json
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
RETURNING
    id,
    examination_id,
    specialist_id,
    status,
    payload_version,
    aggregation_version,
    recommendation,
    message,
    correlation_id,
    max_attempts,
    attempt_count,
    last_attempt_at,
    completed_at,
    diagnostics_json,
    raw_response_json,
    created_at,
    updated_at;

-- name: GetDecisionSnapshotByExamination :one
SELECT
    id,
    examination_id,
    specialist_id,
    status,
    payload_version,
    aggregation_version,
    recommendation,
    message,
    correlation_id,
    max_attempts,
    attempt_count,
    last_attempt_at,
    completed_at,
    diagnostics_json,
    raw_response_json,
    created_at,
    updated_at
FROM decision_snapshots
WHERE examination_id = $1;

-- name: InsertDecisionAttempt :one
INSERT INTO decision_attempts (
    decision_snapshot_id,
    attempt_number,
    correlation_id,
    request_payload_json,
    response_payload_json,
    error_code,
    error_message,
    error_class,
    retryable,
    http_status,
    started_at,
    finished_at
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING
    id,
    decision_snapshot_id,
    attempt_number,
    correlation_id,
    request_payload_json,
    response_payload_json,
    error_code,
    error_message,
    error_class,
    retryable,
    http_status,
    started_at,
    finished_at,
    created_at;

-- name: ListPendingDecisionSnapshots :many
SELECT
    id,
    examination_id,
    specialist_id,
    status,
    payload_version,
    aggregation_version,
    recommendation,
    message,
    correlation_id,
    max_attempts,
    attempt_count,
    last_attempt_at,
    completed_at,
    diagnostics_json,
    raw_response_json,
    created_at,
    updated_at
FROM decision_snapshots
WHERE status = 'pending'
ORDER BY updated_at ASC, id ASC;

-- name: MarkDecisionAttemptSucceeded :one
UPDATE decision_snapshots
SET
    status = 'succeeded',
    recommendation = $2,
    message = $3,
    correlation_id = $4,
    attempt_count = $5,
    last_attempt_at = $6,
    completed_at = $7,
    diagnostics_json = $8,
    raw_response_json = $9,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    examination_id,
    specialist_id,
    status,
    payload_version,
    aggregation_version,
    recommendation,
    message,
    correlation_id,
    max_attempts,
    attempt_count,
    last_attempt_at,
    completed_at,
    diagnostics_json,
    raw_response_json,
    created_at,
    updated_at;

-- name: MarkDecisionAttemptFailed :one
UPDATE decision_snapshots
SET
    status = $2,
    recommendation = $3,
    message = $4,
    correlation_id = $5,
    attempt_count = $6,
    last_attempt_at = $7,
    completed_at = $8,
    diagnostics_json = $9,
    raw_response_json = $10,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    examination_id,
    specialist_id,
    status,
    payload_version,
    aggregation_version,
    recommendation,
    message,
    correlation_id,
    max_attempts,
    attempt_count,
    last_attempt_at,
    completed_at,
    diagnostics_json,
    raw_response_json,
    created_at,
    updated_at;

-- name: IncrementDecisionAttemptCount :one
UPDATE decision_snapshots
SET
    attempt_count = attempt_count + 1,
    correlation_id = $2,
    last_attempt_at = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    examination_id,
    specialist_id,
    status,
    payload_version,
    aggregation_version,
    recommendation,
    message,
    correlation_id,
    max_attempts,
    attempt_count,
    last_attempt_at,
    completed_at,
    diagnostics_json,
    raw_response_json,
    created_at,
    updated_at;

-- name: GetExaminationDecisionProjection :one
SELECT
    examinations.id AS examination_id,
    examinations.status AS examination_status,
    snapshots.id AS decision_snapshot_id,
    snapshots.status AS decision_status,
    snapshots.recommendation,
    snapshots.message,
    snapshots.correlation_id,
    snapshots.attempt_count,
    snapshots.max_attempts,
    snapshots.last_attempt_at,
    snapshots.completed_at,
    snapshots.diagnostics_json,
    (snapshots.raw_response_json IS NOT NULL) AS raw_response_available
FROM examinations
LEFT JOIN decision_snapshots snapshots
    ON snapshots.examination_id = examinations.id
WHERE examinations.id = $1;
