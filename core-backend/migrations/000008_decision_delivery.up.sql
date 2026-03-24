ALTER TABLE examinations
    DROP CONSTRAINT IF EXISTS examinations_status_check;

ALTER TABLE examinations
    ADD CONSTRAINT examinations_status_check
        CHECK (status IN (
            'created',
            'collecting_answers',
            'ready_for_processing',
            'processing',
            'aggregating',
            'aggregated',
            'decision_pending',
            'completed',
            'failed'
        ));

CREATE TABLE decision_snapshots (
    id BIGSERIAL PRIMARY KEY,
    examination_id BIGINT NOT NULL REFERENCES examinations (id) ON DELETE CASCADE,
    specialist_id BIGINT NOT NULL REFERENCES specialists (id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('pending', 'succeeded', 'transport_exhausted', 'business_error')),
    payload_version TEXT NOT NULL,
    aggregation_version TEXT NOT NULL,
    recommendation TEXT NOT NULL CHECK (recommendation IN ('unavailable', 'allowed', 'risk', 'denied')),
    message TEXT NOT NULL,
    correlation_id TEXT,
    max_attempts INTEGER NOT NULL CHECK (max_attempts > 0),
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    last_attempt_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    diagnostics_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    raw_response_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (examination_id)
);

CREATE INDEX idx_decision_snapshots_status_updated_at
    ON decision_snapshots (status, updated_at ASC);

CREATE TABLE decision_attempts (
    id BIGSERIAL PRIMARY KEY,
    decision_snapshot_id BIGINT NOT NULL REFERENCES decision_snapshots (id) ON DELETE CASCADE,
    attempt_number INTEGER NOT NULL CHECK (attempt_number > 0),
    correlation_id TEXT,
    request_payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    response_payload_json JSONB,
    error_code TEXT,
    error_message TEXT,
    error_class TEXT,
    retryable BOOLEAN NOT NULL,
    http_status INTEGER,
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (decision_snapshot_id, attempt_number)
);

CREATE INDEX idx_decision_attempts_snapshot_attempt
    ON decision_attempts (decision_snapshot_id, attempt_number DESC);
