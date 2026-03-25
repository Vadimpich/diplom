DROP INDEX IF EXISTS idx_questionnaires_last_edited_by_user_id;
DROP INDEX IF EXISTS idx_users_last_login_at;

ALTER TABLE questionnaires
    DROP COLUMN IF EXISTS last_edited_at,
    DROP COLUMN IF EXISTS last_edited_by_user_id;

ALTER TABLE users
    DROP COLUMN IF EXISTS last_login_at;
