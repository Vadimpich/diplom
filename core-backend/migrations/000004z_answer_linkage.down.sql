DROP INDEX IF EXISTS idx_answers_specialist_id;
DROP INDEX IF EXISTS idx_answers_examination_question_unique;

ALTER TABLE answers
    DROP COLUMN IF EXISTS specialist_id,
    DROP COLUMN IF EXISTS examination_question_id;
