ALTER TABLE examination_baseline_snapshots
    DROP COLUMN IF EXISTS candidate_metrics,
    DROP COLUMN IF EXISTS significant_deviations,
    DROP COLUMN IF EXISTS data_reliability,
    DROP COLUMN IF EXISTS personal_baseline_source,
    DROP COLUMN IF EXISTS personal_baseline_available,
    DROP COLUMN IF EXISTS general_baseline_source,
    DROP COLUMN IF EXISTS general_baseline_available;
