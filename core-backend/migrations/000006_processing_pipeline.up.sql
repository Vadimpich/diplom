ALTER TABLE examinations
    DROP CONSTRAINT IF EXISTS examinations_status_check;

ALTER TABLE examinations
    ADD CONSTRAINT examinations_status_check
        CHECK (status IN ('created', 'collecting_answers', 'ready_for_processing', 'processing', 'failed'));

CREATE TABLE examination_channel_runs (
    id BIGSERIAL PRIMARY KEY,
    examination_id BIGINT NOT NULL REFERENCES examinations (id) ON DELETE CASCADE,
    channel TEXT NOT NULL CHECK (channel IN ('text', 'acoustic', 'paralinguistic')),
    status TEXT NOT NULL CHECK (status IN (
        'pending',
        'queued',
        'processing',
        'succeeded',
        'retry_scheduled',
        'failed_temporary',
        'failed_fatal',
        'exhausted'
    )),
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    max_attempts INTEGER NOT NULL DEFAULT 3 CHECK (max_attempts > 0),
    message_version INTEGER NOT NULL CHECK (message_version > 0),
    last_error_code TEXT,
    last_error_message TEXT,
    broker_message_id TEXT,
    broker_correlation_id TEXT,
    queued_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (examination_id, channel)
);

CREATE INDEX idx_examination_channel_runs_examination_id
    ON examination_channel_runs (examination_id);

CREATE INDEX idx_examination_channel_runs_status
    ON examination_channel_runs (status);

CREATE TABLE processing_outbox (
    id BIGSERIAL PRIMARY KEY,
    examination_id BIGINT NOT NULL REFERENCES examinations (id) ON DELETE CASCADE,
    channel_run_id BIGINT NOT NULL REFERENCES examination_channel_runs (id) ON DELETE CASCADE,
    channel TEXT NOT NULL CHECK (channel IN ('text', 'acoustic', 'paralinguistic')),
    exchange_name TEXT NOT NULL,
    routing_key TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'published', 'failed')),
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    max_attempts INTEGER NOT NULL DEFAULT 3 CHECK (max_attempts > 0),
    message_version INTEGER NOT NULL CHECK (message_version > 0),
    payload JSONB NOT NULL,
    broker_message_id TEXT,
    broker_correlation_id TEXT,
    last_error_code TEXT,
    last_error_message TEXT,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (channel_run_id)
);

CREATE INDEX idx_processing_outbox_status_created_at
    ON processing_outbox (status, created_at);

CREATE INDEX idx_processing_outbox_examination_id
    ON processing_outbox (examination_id);

CREATE TABLE channel_results (
    id BIGSERIAL PRIMARY KEY,
    examination_id BIGINT NOT NULL REFERENCES examinations (id) ON DELETE CASCADE,
    channel_run_id BIGINT NOT NULL REFERENCES examination_channel_runs (id) ON DELETE CASCADE,
    channel TEXT NOT NULL CHECK (channel IN ('text', 'acoustic', 'paralinguistic')),
    message_version INTEGER NOT NULL CHECK (message_version > 0),
    attempt INTEGER NOT NULL CHECK (attempt > 0),
    status TEXT NOT NULL CHECK (status IN ('succeeded', 'temporary_error', 'fatal_error')),
    model_version TEXT NOT NULL,
    payload JSONB NOT NULL,
    error_code TEXT,
    error_message TEXT,
    broker_message_id TEXT,
    broker_correlation_id TEXT,
    completed_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (channel_run_id, attempt)
);

CREATE INDEX idx_channel_results_examination_id
    ON channel_results (examination_id);
