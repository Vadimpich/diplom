DROP TABLE IF EXISTS examination_baseline_snapshots;
DROP TABLE IF EXISTS specialist_baseline_states;
DROP TABLE IF EXISTS aggregated_profile_explanations;
DROP TABLE IF EXISTS aggregated_profile_channel_contributions;
DROP TABLE IF EXISTS aggregated_profile_metrics;
DROP TABLE IF EXISTS aggregated_examination_profiles;

ALTER TABLE examinations
    DROP CONSTRAINT IF EXISTS examinations_status_check;

ALTER TABLE examinations
    ADD CONSTRAINT examinations_status_check
        CHECK (status IN (
            'created',
            'collecting_answers',
            'ready_for_processing',
            'processing',
            'failed'
        ));
