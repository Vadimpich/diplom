ALTER TABLE users
    ADD COLUMN last_login_at TIMESTAMPTZ;

ALTER TABLE questionnaires
    ADD COLUMN last_edited_by_user_id BIGINT REFERENCES users (id) ON DELETE SET NULL,
    ADD COLUMN last_edited_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX idx_users_last_login_at ON users (last_login_at DESC);
CREATE INDEX idx_questionnaires_last_edited_by_user_id ON questionnaires (last_edited_by_user_id);
