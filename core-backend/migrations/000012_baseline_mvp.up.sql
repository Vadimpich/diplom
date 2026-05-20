ALTER TABLE examination_baseline_snapshots
    ADD COLUMN general_baseline_available BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN general_baseline_source TEXT NOT NULL DEFAULT 'general',
    ADD COLUMN personal_baseline_available BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN personal_baseline_source TEXT NOT NULL DEFAULT 'general',
    ADD COLUMN data_reliability DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN significant_deviations JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN candidate_metrics JSONB NOT NULL DEFAULT '{}'::jsonb;
