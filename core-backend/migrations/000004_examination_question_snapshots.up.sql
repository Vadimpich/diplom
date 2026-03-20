CREATE TABLE examination_questions (
    id BIGSERIAL PRIMARY KEY,
    examination_id BIGINT NOT NULL REFERENCES examinations (id) ON DELETE CASCADE,
    specialist_id BIGINT NOT NULL REFERENCES specialists (id) ON DELETE RESTRICT,
    questionnaire_id BIGINT NOT NULL REFERENCES questionnaires (id) ON DELETE RESTRICT,
    source_question_id BIGINT REFERENCES questions (id) ON DELETE SET NULL,
    position INTEGER NOT NULL CHECK (position > 0),
    question_text TEXT NOT NULL,
    UNIQUE (examination_id, position)
);

CREATE INDEX idx_examination_questions_examination_id
    ON examination_questions (examination_id);
