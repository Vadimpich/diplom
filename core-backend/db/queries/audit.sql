-- name: CreateAuditLog :one
INSERT INTO audit_logs (
    event_type,
    event_key,
    outcome,
    happened_at,
    request_id,
    trace_id,
    traceparent,
    tracestate,
    correlation_id,
    actor_user_id,
    actor_login,
    actor_role_slug,
    actor_ip,
    actor_user_agent,
    resource_kind,
    resource_id,
    examination_id,
    specialist_id,
    questionnaire_id,
    channel,
    decision_snapshot_id,
    payload
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9,
    $10, $11, $12, $13, $14, $15, $16, $17, $18,
    $19, $20, $21, $22
)
ON CONFLICT (event_key) WHERE event_key IS NOT NULL DO NOTHING
RETURNING *;

-- name: ListAuditLogs :many
SELECT *
FROM audit_logs
WHERE
    (sqlc.narg(event_type)::text IS NULL OR event_type = sqlc.narg(event_type)::text)
    AND (sqlc.narg(resource_kind)::text IS NULL OR resource_kind = sqlc.narg(resource_kind)::text)
    AND (sqlc.narg(actor_user_id)::bigint IS NULL OR actor_user_id = sqlc.narg(actor_user_id)::bigint)
    AND (sqlc.narg(examination_id)::bigint IS NULL OR examination_id = sqlc.narg(examination_id)::bigint)
    AND happened_at >= sqlc.arg(from_at)
    AND happened_at <= sqlc.arg(to_at)
ORDER BY happened_at DESC
LIMIT sqlc.arg(limit_count);
