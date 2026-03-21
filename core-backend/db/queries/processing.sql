-- name: CreateProcessingChannelRuns :many
WITH mandatory_channels AS (
    SELECT
        unnest($2::text[]) AS channel
)
INSERT INTO examination_channel_runs (
    examination_id,
    channel,
    status,
    attempt_count,
    max_attempts,
    message_version
)
SELECT
    $1,
    mandatory_channels.channel,
    'pending',
    0,
    $3,
    $4
FROM mandatory_channels
JOIN examination_processing_launches launches
    ON launches.examination_id = $1
ON CONFLICT (examination_id, channel) DO UPDATE
SET
    max_attempts = EXCLUDED.max_attempts,
    message_version = EXCLUDED.message_version,
    updated_at = NOW()
RETURNING
    id,
    examination_id,
    channel,
    status,
    attempt_count,
    max_attempts,
    message_version,
    last_error_code,
    last_error_message,
    broker_message_id,
    broker_correlation_id,
    queued_at,
    started_at,
    finished_at,
    created_at,
    updated_at;

-- name: CreateProcessingOutboxRows :many
INSERT INTO processing_outbox (
    examination_id,
    channel_run_id,
    channel,
    exchange_name,
    routing_key,
    status,
    attempt_count,
    max_attempts,
    message_version,
    payload,
    broker_correlation_id
)
SELECT
    channel_runs.examination_id,
    channel_runs.id,
    channel_runs.channel,
    'processing.commands',
    CASE channel_runs.channel
        WHEN 'text' THEN 'processing.command.text'
        WHEN 'acoustic' THEN 'processing.command.acoustic'
        ELSE 'processing.command.paralinguistic'
    END,
    'pending',
    channel_runs.attempt_count,
    channel_runs.max_attempts,
    channel_runs.message_version,
    $2::jsonb,
    CASE channel_runs.channel
        WHEN 'text' THEN format('exam-%s-text-v%s', channel_runs.examination_id, channel_runs.message_version)
        WHEN 'acoustic' THEN format('exam-%s-acoustic-v%s', channel_runs.examination_id, channel_runs.message_version)
        ELSE format('exam-%s-paralinguistic-v%s', channel_runs.examination_id, channel_runs.message_version)
    END
FROM examination_channel_runs channel_runs
WHERE channel_runs.examination_id = $1
ON CONFLICT (channel_run_id) DO NOTHING
RETURNING
    id,
    examination_id,
    channel_run_id,
    channel,
    exchange_name,
    routing_key,
    status,
    attempt_count,
    max_attempts,
    message_version,
    payload,
    broker_message_id,
    broker_correlation_id,
    last_error_code,
    last_error_message,
    published_at,
    created_at,
    updated_at;

-- name: ListPendingProcessingOutboxRows :many
SELECT
    id,
    examination_id,
    channel_run_id,
    channel,
    exchange_name,
    routing_key,
    status,
    attempt_count,
    max_attempts,
    message_version,
    payload,
    broker_message_id,
    broker_correlation_id,
    last_error_code,
    last_error_message,
    published_at,
    created_at,
    updated_at
FROM processing_outbox
WHERE status = 'pending'
ORDER BY created_at ASC, id ASC;

-- name: GetExaminationProcessingProjection :many
SELECT
    examinations.id AS examination_id,
    examinations.status AS examination_status,
    examinations.finished_at AS examination_finished_at,
    examinations.updated_at AS examination_updated_at,
    channel_runs.id AS channel_run_id,
    channel_runs.channel,
    channel_runs.status AS channel_status,
    channel_runs.attempt_count,
    channel_runs.max_attempts,
    channel_runs.message_version,
    channel_runs.last_error_code,
    channel_runs.last_error_message,
    channel_runs.broker_message_id,
    channel_runs.broker_correlation_id,
    channel_runs.queued_at,
    channel_runs.started_at,
    channel_runs.finished_at,
    processing_outbox.status AS outbox_status,
    processing_outbox.published_at,
    processing_outbox.updated_at AS outbox_updated_at
FROM examinations
LEFT JOIN examination_channel_runs channel_runs
    ON channel_runs.examination_id = examinations.id
LEFT JOIN processing_outbox
    ON processing_outbox.channel_run_id = channel_runs.id
WHERE examinations.id = $1
ORDER BY channel_runs.channel ASC;
