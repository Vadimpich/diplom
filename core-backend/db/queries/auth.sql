-- name: GetUserByLogin :one
SELECT
    u.id,
    u.login,
    u.password_hash,
    u.is_active,
    u.created_at,
    u.updated_at,
    r.id AS role_id,
    r.slug AS role_slug,
    r.name AS role_name,
    r.created_at AS role_created_at
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.login = $1;

-- name: GetUserByID :one
SELECT
    u.id,
    u.login,
    u.password_hash,
    u.is_active,
    u.created_at,
    u.updated_at,
    r.id AS role_id,
    r.slug AS role_slug,
    r.name AS role_name,
    r.created_at AS role_created_at
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.id = $1;

-- name: ListUsers :many
SELECT
    u.id,
    u.login,
    u.password_hash,
    u.is_active,
    u.created_at,
    u.updated_at,
    r.id AS role_id,
    r.slug AS role_slug,
    r.name AS role_name,
    r.created_at AS role_created_at
FROM users u
JOIN roles r ON r.id = u.role_id
ORDER BY u.id;

-- name: GetRoleBySlug :one
SELECT id, slug, name, created_at
FROM roles
WHERE slug = $1;

-- name: CreateUser :one
INSERT INTO users (
    login,
    password_hash,
    role_id,
    is_active
) VALUES (
    $1, $2, $3, TRUE
)
RETURNING id, login, password_hash, role_id, is_active, created_at, updated_at;

-- name: UpdateUser :one
UPDATE users
SET
    login = $2,
    role_id = $3,
    is_active = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING id, login, password_hash, role_id, is_active, created_at, updated_at;

-- name: UpsertUser :one
INSERT INTO users (
    login,
    password_hash,
    role_id,
    is_active
) VALUES (
    $1, $2, $3, TRUE
)
ON CONFLICT (login) DO UPDATE
SET
    password_hash = EXCLUDED.password_hash,
    role_id = EXCLUDED.role_id,
    is_active = TRUE,
    updated_at = NOW()
RETURNING id, login, password_hash, role_id, is_active, created_at, updated_at;

-- name: CreateRefreshSession :one
INSERT INTO refresh_sessions (
    user_id,
    token_hash,
    expires_at,
    created_by_ip,
    user_agent,
    last_used_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING id, user_id, token_hash, expires_at, revoked_at, replaced_by_session_id, created_by_ip, user_agent, last_used_at, created_at;

-- name: GetRefreshSessionByHash :one
SELECT
    id,
    user_id,
    token_hash,
    expires_at,
    revoked_at,
    replaced_by_session_id,
    created_by_ip,
    user_agent,
    last_used_at,
    created_at
FROM refresh_sessions
WHERE token_hash = $1
FOR UPDATE;

-- name: RotateRefreshSession :exec
UPDATE refresh_sessions
SET
    revoked_at = $2,
    replaced_by_session_id = $3,
    last_used_at = $4
WHERE id = $1;

-- name: TouchRefreshSession :exec
UPDATE refresh_sessions
SET
    last_used_at = $2
WHERE id = $1;

-- name: RevokeRefreshSession :exec
UPDATE refresh_sessions
SET
    revoked_at = $2
WHERE id = $1;
