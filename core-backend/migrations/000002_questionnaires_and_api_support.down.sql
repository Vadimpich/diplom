DROP INDEX IF EXISTS idx_examinations_questionnaire_id;

ALTER TABLE examinations
    DROP COLUMN IF EXISTS questionnaire_id;

DROP INDEX IF EXISTS idx_questionnaire_questions_questionnaire_id;
DROP TABLE IF EXISTS questionnaire_questions;
DROP TABLE IF EXISTS questions;
DROP TABLE IF EXISTS questionnaires;
