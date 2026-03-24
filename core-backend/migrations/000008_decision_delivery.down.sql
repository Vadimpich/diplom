DROP TABLE IF EXISTS decision_attempts;
DROP TABLE IF EXISTS decision_snapshots;

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
