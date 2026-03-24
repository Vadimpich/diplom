CREATE TABLE roles (
    id BIGSERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO roles (slug, name)
VALUES
    ('admin', 'Administrator'),
    ('operator', 'Operator');

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role_id BIGINT NOT NULL REFERENCES roles (id),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_role_id ON users (role_id);

CREATE TABLE specialists (
    id BIGSERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    personnel_number TEXT UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE examinations (
    id BIGSERIAL PRIMARY KEY,
    specialist_id BIGINT NOT NULL REFERENCES specialists (id) ON DELETE RESTRICT,
    created_by_user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    status TEXT NOT NULL CHECK (status IN ('created', 'collecting_answers', 'ready_for_processing')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_examinations_specialist_id ON examinations (specialist_id);
CREATE INDEX idx_examinations_created_by_user_id ON examinations (created_by_user_id);
CREATE INDEX idx_examinations_status ON examinations (status);

CREATE TABLE answers (
    id BIGSERIAL PRIMARY KEY,
    examination_id BIGINT NOT NULL REFERENCES examinations (id) ON DELETE CASCADE,
    created_by_user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    answer_text TEXT NOT NULL,
    audio_s3_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_answers_examination_id ON answers (examination_id);
CREATE INDEX idx_answers_created_by_user_id ON answers (created_by_user_id);

