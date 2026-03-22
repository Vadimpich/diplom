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
            'failed'
        ));

CREATE TABLE aggregated_examination_profiles (
    id BIGSERIAL PRIMARY KEY,
    examination_id BIGINT NOT NULL REFERENCES examinations (id) ON DELETE CASCADE,
    specialist_id BIGINT NOT NULL REFERENCES specialists (id) ON DELETE CASCADE,
    schema_version INTEGER NOT NULL CHECK (schema_version > 0),
    aggregation_version TEXT NOT NULL,
    baseline_algorithm_version TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('aggregating', 'aggregated')),
    generated_at TIMESTAMPTZ NOT NULL,
    baseline_refreshed_at TIMESTAMPTZ,
    baseline_exam_count INTEGER NOT NULL DEFAULT 0 CHECK (baseline_exam_count >= 0),
    overall_score DOUBLE PRECISION,
    overall_band TEXT,
    primary_metric_key TEXT,
    neutral_recommendation_placeholder TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (examination_id)
);

CREATE INDEX idx_aggregated_profiles_specialist_generated_at
    ON aggregated_examination_profiles (specialist_id, generated_at DESC);

CREATE TABLE aggregated_profile_metrics (
    id BIGSERIAL PRIMARY KEY,
    profile_id BIGINT NOT NULL REFERENCES aggregated_examination_profiles (id) ON DELETE CASCADE,
    metric_key TEXT NOT NULL,
    label TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    scale TEXT NOT NULL,
    direction TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (profile_id, metric_key)
);

CREATE INDEX idx_aggregated_profile_metrics_profile_id
    ON aggregated_profile_metrics (profile_id);

CREATE TABLE aggregated_profile_channel_contributions (
    id BIGSERIAL PRIMARY KEY,
    profile_id BIGINT NOT NULL REFERENCES aggregated_examination_profiles (id) ON DELETE CASCADE,
    channel TEXT NOT NULL CHECK (channel IN ('text', 'acoustic', 'paralinguistic')),
    metric_key TEXT NOT NULL,
    weight DOUBLE PRECISION NOT NULL,
    contribution DOUBLE PRECISION NOT NULL,
    evidence_keys JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (profile_id, channel, metric_key)
);

CREATE INDEX idx_aggregated_profile_contributions_profile_id
    ON aggregated_profile_channel_contributions (profile_id);

CREATE TABLE aggregated_profile_explanations (
    id BIGSERIAL PRIMARY KEY,
    profile_id BIGINT NOT NULL REFERENCES aggregated_examination_profiles (id) ON DELETE CASCADE,
    position INTEGER NOT NULL CHECK (position > 0),
    kind TEXT NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (profile_id, position)
);

CREATE INDEX idx_aggregated_profile_explanations_profile_id
    ON aggregated_profile_explanations (profile_id);

CREATE TABLE specialist_baseline_states (
    id BIGSERIAL PRIMARY KEY,
    specialist_id BIGINT NOT NULL REFERENCES specialists (id) ON DELETE CASCADE,
    algorithm_version TEXT NOT NULL,
    refreshed_at TIMESTAMPTZ NOT NULL,
    baseline_exam_count INTEGER NOT NULL DEFAULT 0 CHECK (baseline_exam_count >= 0),
    update_eligible BOOLEAN NOT NULL DEFAULT FALSE,
    update_reason TEXT,
    metrics JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (specialist_id)
);

CREATE TABLE examination_baseline_snapshots (
    id BIGSERIAL PRIMARY KEY,
    examination_id BIGINT NOT NULL REFERENCES examinations (id) ON DELETE CASCADE,
    profile_id BIGINT NOT NULL REFERENCES aggregated_examination_profiles (id) ON DELETE CASCADE,
    algorithm_version TEXT NOT NULL,
    refreshed_at TIMESTAMPTZ NOT NULL,
    general_delta DOUBLE PRECISION NOT NULL,
    general_band TEXT NOT NULL,
    general_reference_population_version TEXT,
    personal_delta DOUBLE PRECISION NOT NULL,
    personal_band TEXT NOT NULL,
    baseline_exam_count INTEGER NOT NULL DEFAULT 0 CHECK (baseline_exam_count >= 0),
    update_eligible BOOLEAN NOT NULL DEFAULT FALSE,
    update_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (examination_id),
    UNIQUE (profile_id)
);

CREATE INDEX idx_examination_baseline_snapshots_profile_id
    ON examination_baseline_snapshots (profile_id);
