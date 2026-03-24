CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    event_type TEXT NOT NULL,
    event_key TEXT,
    outcome TEXT NOT NULL,
    happened_at TIMESTAMPTZ NOT NULL,
    request_id TEXT,
    trace_id TEXT,
    traceparent TEXT,
    tracestate TEXT,
    correlation_id TEXT,
    actor_user_id BIGINT,
    actor_login TEXT,
    actor_role_slug TEXT,
    actor_ip TEXT,
    actor_user_agent TEXT,
    resource_kind TEXT NOT NULL,
    resource_id BIGINT,
    examination_id BIGINT,
    specialist_id BIGINT,
    questionnaire_id BIGINT,
    channel TEXT,
    decision_snapshot_id BIGINT,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX ux_audit_logs_event_key
    ON audit_logs (event_key)
    WHERE event_key IS NOT NULL;

CREATE INDEX idx_audit_logs_event_type_happened_at
    ON audit_logs (event_type, happened_at DESC);

CREATE INDEX idx_audit_logs_actor_user_id_happened_at
    ON audit_logs (actor_user_id, happened_at DESC);

CREATE INDEX idx_audit_logs_examination_id_happened_at
    ON audit_logs (examination_id, happened_at DESC);

CREATE INDEX idx_audit_logs_resource_happened_at
    ON audit_logs (resource_kind, resource_id, happened_at DESC);
