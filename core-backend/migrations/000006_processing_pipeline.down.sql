DROP TABLE IF EXISTS channel_results;
DROP TABLE IF EXISTS processing_outbox;
DROP TABLE IF EXISTS examination_channel_runs;

ALTER TABLE examinations
    DROP CONSTRAINT IF EXISTS examinations_status_check;

ALTER TABLE examinations
    ADD CONSTRAINT examinations_status_check
        CHECK (status IN ('created', 'collecting_answers', 'ready_for_processing'));
