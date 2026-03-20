ALTER TABLE answers
    ADD COLUMN examination_question_id BIGINT REFERENCES examination_questions (id) ON DELETE RESTRICT,
    ADD COLUMN specialist_id BIGINT REFERENCES specialists (id) ON DELETE RESTRICT;

CREATE UNIQUE INDEX idx_answers_examination_question_unique
    ON answers (examination_id, examination_question_id);

CREATE INDEX idx_answers_specialist_id ON answers (specialist_id);
