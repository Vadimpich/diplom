---
phase: 01-trusted-access-and-intake
plan: 01
title: Auth Session Backend Foundation
type: execute
wave: 1
depends_on: []
files_modified:
  - /home/vadim/diplom/core-backend/migrations/000003_refresh_sessions.up.sql
  - /home/vadim/diplom/core-backend/migrations/000003_refresh_sessions.down.sql
  - /home/vadim/diplom/core-backend/db/queries/auth.sql
  - /home/vadim/diplom/core-backend/db/sqlc/auth.sql.go
  - /home/vadim/diplom/core-backend/db/sqlc/models.go
  - /home/vadim/diplom/core-backend/internal/auth/service.go
  - /home/vadim/diplom/core-backend/internal/auth/repository.go
  - /home/vadim/diplom/core-backend/internal/auth/jwt.go
  - /home/vadim/diplom/core-backend/internal/http/auth_handler.go
  - /home/vadim/diplom/core-backend/internal/http/router.go
  - /home/vadim/diplom/core-backend/internal/http/auth_handler_test.go
  - /home/vadim/diplom/core-backend/internal/auth/service_test.go
  - /home/vadim/diplom/docs/01_contract.md
  - /home/vadim/diplom/docs/02_implementation.md
autonomous: true
requirements_addressed:
  - ACCS-01
  - ACCS-02
  - ACCS-03
must_haves:
  truths:
    - "User can log in with login/password and receives a short-lived access token plus a server-managed refresh session token intended for the frontend BFF boundary."
    - "User can refresh access without re-entering credentials while the refresh session is active."
    - "User can log out and the revoked refresh session cannot be reused."
  artifacts:
    - path: /home/vadim/diplom/core-backend/migrations/000003_refresh_sessions.up.sql
      provides: refresh session persistence
    - path: /home/vadim/diplom/core-backend/internal/http/auth_handler.go
      provides: /auth/login, /auth/refresh, /auth/logout handlers
    - path: /home/vadim/diplom/docs/01_contract.md
      provides: authoritative auth contracts and BFF-facing refresh semantics
  key_links:
    - from: /home/vadim/diplom/core-backend/internal/http/auth_handler.go
      to: /home/vadim/diplom/core-backend/internal/auth/service.go
      via: login/refresh/logout handler calls
---

# Objective

Create the backend auth/session foundation for secure login, refresh rotation, and logout revocation with PostgreSQL as the source of truth while leaving browser cookie ownership to the frontend BFF.

<tasks>

<task id="1-01-01" type="auto">
  <name>Task 1: Add refresh-session schema and backend auth flows</name>
  <files>
    /home/vadim/diplom/core-backend/migrations/000003_refresh_sessions.up.sql
    /home/vadim/diplom/core-backend/migrations/000003_refresh_sessions.down.sql
    /home/vadim/diplom/core-backend/db/queries/auth.sql
    /home/vadim/diplom/core-backend/db/sqlc/auth.sql.go
    /home/vadim/diplom/core-backend/db/sqlc/models.go
    /home/vadim/diplom/core-backend/internal/auth/service.go
    /home/vadim/diplom/core-backend/internal/auth/repository.go
    /home/vadim/diplom/core-backend/internal/auth/jwt.go
    /home/vadim/diplom/core-backend/internal/http/auth_handler.go
    /home/vadim/diplom/core-backend/internal/http/router.go
  </files>
  <read_first>
    /home/vadim/diplom/AGENTS.md
    /home/vadim/diplom/docs/00_project.md
    /home/vadim/diplom/docs/01_contract.md
    /home/vadim/diplom/.planning/phases/01-trusted-access-and-intake/01-RESEARCH.md
  </read_first>
  <action>
    Add migration `000003_refresh_sessions` creating table `refresh_sessions` with exact columns `id BIGSERIAL PRIMARY KEY`, `user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE`, `token_hash BYTEA NOT NULL UNIQUE`, `expires_at TIMESTAMPTZ NOT NULL`, `revoked_at TIMESTAMPTZ`, `replaced_by_session_id BIGINT REFERENCES refresh_sessions(id) ON DELETE SET NULL`, `created_by_ip TEXT`, `user_agent TEXT`, `last_used_at TIMESTAMPTZ`, and `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`. Extend `core-backend/db/queries/auth.sql` and regenerated SQLC code with `CreateRefreshSession`, `GetRefreshSessionByHash`, `RotateRefreshSession`, `TouchRefreshSession`, and `RevokeRefreshSession`. Update `internal/auth/service.go`, `internal/auth/repository.go`, `internal/auth/jwt.go`, `internal/http/auth_handler.go`, and `internal/http/router.go` so the backend exposes `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`, and `GET /me`; `login` and `refresh` issue a short-lived access JWT plus a refresh token value for the trusted frontend BFF to store in an `HttpOnly` browser cookie, while `logout` revokes the refresh session without making core-backend the browser cookie authority.
  </action>
  <verify>
    <automated>cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/auth -run 'TestLogin|TestRefresh|TestLogout|TestMe' -count=1</automated>
  </verify>
  <acceptance_criteria>
    `rg -n "CREATE TABLE refresh_sessions|token_hash BYTEA|replaced_by_session_id" /home/vadim/diplom/core-backend/migrations/000003_refresh_sessions.up.sql`
  </acceptance_criteria>
  <done>
    `refresh_sessions` schema, SQL queries, handlers, and service methods exist; `/auth/login`, `/auth/refresh`, `/auth/logout`, and `/me` are wired in router; backend refresh semantics are documented for BFF consumption rather than direct browser cookie writes; automated verify command passes.
  </done>
</task>

<task id="1-01-02" type="auto">
  <name>Task 2: Add auth tests and update contracts</name>
  <files>
    /home/vadim/diplom/core-backend/internal/http/auth_handler_test.go
    /home/vadim/diplom/core-backend/internal/auth/service_test.go
    /home/vadim/diplom/docs/01_contract.md
    /home/vadim/diplom/docs/02_implementation.md
  </files>
  <read_first>
    /home/vadim/diplom/AGENTS.md
    /home/vadim/diplom/.planning/phases/01-trusted-access-and-intake/01-VALIDATION.md
    /home/vadim/diplom/docs/01_contract.md
    /home/vadim/diplom/docs/02_implementation.md
  </read_first>
  <action>
    Create Wave 0 tests in `core-backend/internal/http/auth_handler_test.go` and `core-backend/internal/auth/service_test.go` covering successful login, refresh rotation, revoked refresh rejection, inactive-user rejection, logout revocation, and `/me` behavior. Update `docs/01_contract.md` with exact payloads for `/auth/login`, `/auth/refresh`, `/auth/logout`, and `/me`, explicitly stating that the frontend BFF owns browser cookie transport while core-backend exposes refresh-token exchange to that trusted boundary. Append a concise entry to `docs/02_implementation.md` describing the refresh-session model. Treat the documentation update required by `AGENTS.md` as part of completion.
  </action>
  <verify>
    <automated>cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/auth -run 'TestLogin|TestRefreshRotation|TestLogoutRevokesSession|TestMe' -count=1</automated>
  </verify>
  <acceptance_criteria>
    `rg -n "/auth/login|/auth/refresh|/auth/logout|HttpOnly|refresh session" /home/vadim/diplom/docs/01_contract.md`
  </acceptance_criteria>
  <done>
    Auth tests referenced in `01-VALIDATION.md` exist and pass; `docs/01_contract.md` documents the changed contracts; `docs/02_implementation.md` records the implementation status in compliance with `AGENTS.md`.
  </done>
</task>

</tasks>
