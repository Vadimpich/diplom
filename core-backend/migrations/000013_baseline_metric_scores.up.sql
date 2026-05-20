ALTER TABLE examination_baseline_snapshots
    ADD COLUMN baseline_metric_scores JSONB NOT NULL DEFAULT '{}'::jsonb;
