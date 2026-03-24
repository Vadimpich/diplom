CREATE TABLE questionnaires (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE questions (
    id BIGSERIAL PRIMARY KEY,
    text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE questionnaire_questions (
    questionnaire_id BIGINT NOT NULL REFERENCES questionnaires (id) ON DELETE CASCADE,
    question_id BIGINT NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
    position INTEGER NOT NULL CHECK (position > 0),
    PRIMARY KEY (questionnaire_id, question_id),
    UNIQUE (questionnaire_id, position)
);

CREATE INDEX idx_questionnaire_questions_questionnaire_id
    ON questionnaire_questions (questionnaire_id);

ALTER TABLE examinations
    ADD COLUMN questionnaire_id BIGINT REFERENCES questionnaires (id) ON DELETE RESTRICT;

CREATE INDEX idx_examinations_questionnaire_id ON examinations (questionnaire_id);
