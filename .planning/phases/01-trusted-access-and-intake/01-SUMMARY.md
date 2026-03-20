---
phase: 01-trusted-access-and-intake
plan: 01
title: Auth Session Backend Foundation Summary
subsystem: auth
tags:
  - go
  - postgresql
  - sqlc
  - jwt
  - refresh-sessions
  - bff
requires: []
provides:
  - rotating refresh sessions backed by PostgreSQL
  - auth login refresh logout me endpoints
  - documented BFF-owned refresh token transport
affects:
  - core-backend/internal/auth
  - core-backend/internal/http
  - docs/01_contract.md
patterns:
  - short-lived access JWT plus opaque hashed refresh session
  - backend exposes token exchange while frontend BFF owns HttpOnly cookie transport
key_files:
  - /home/katya/dimplom/core-backend/internal/auth/service.go
  - /home/katya/dimplom/core-backend/internal/http/auth_handler.go
  - /home/katya/dimplom/core-backend/migrations/000003_refresh_sessions.up.sql
  - /home/katya/dimplom/docs/01_contract.md
decisions:
  - Keep browser cookie ownership out of core-backend; return refresh tokens to the trusted frontend BFF boundary instead.
  - Store only hashed opaque refresh tokens in PostgreSQL so logout and rotation remain server-authoritative.
metrics:
  started_at: 2026-03-20T10:11:24Z
  completed_at: 2026-03-20T10:19:20Z
  duration_seconds: 476
---

# Phase 01 Plan 01: Auth Session Backend Foundation Summary

**Short-lived access JWTs with PostgreSQL-backed opaque refresh rotation and BFF-owned cookie transport**

## Performance

- **Duration:** 8 min
- **Started:** 2026-03-20T10:11:24Z
- **Completed:** 2026-03-20T10:19:20Z
- **Tasks:** 2
- **Files modified:** 16

## Accomplishments

- Added `refresh_sessions` persistence with revocation, replacement linkage, IP/user-agent metadata, and SQLC query support.
- Extended core backend auth flow with `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`, and protected `GET /me` using short-lived access JWTs plus opaque refresh tokens.
- Added Wave 0 tests for login, refresh rotation, logout revocation, inactive-user rejection, and `/me`, then documented BFF-facing auth contracts in project docs.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add refresh-session schema and backend auth flows** - `2be10d9` (feat)
2. **Task 2: Add auth tests and update contracts** - `a40bc09` (test)

## Files Created/Modified

- `core-backend/migrations/000003_refresh_sessions.up.sql` - Creates server-managed refresh session storage and indexes.
- `core-backend/db/queries/auth.sql` - Adds refresh session SQL queries for create, lookup, rotate, touch, and revoke.
- `core-backend/internal/auth/service.go` - Implements login, refresh rotation, and logout revocation behavior.
- `core-backend/internal/http/auth_handler.go` - Exposes `/auth/login`, `/auth/refresh`, `/auth/logout`, and `/me`.
- `core-backend/internal/auth/service_test.go` - Covers refresh rotation, revoked-session rejection, and inactive-user login rejection.
- `core-backend/internal/http/auth_handler_test.go` - Covers login, logout revocation, and `/me`.
- `docs/01_contract.md` - Documents exact auth payloads and BFF-owned `HttpOnly` cookie boundary.
- `docs/02_implementation.md` - Records the new refresh-session model in the implementation log.

## Decisions Made

- Core backend returns `refresh_token` in JSON responses for a trusted frontend BFF instead of attempting to set browser cookies directly.
- Refresh token material is generated as opaque random bytes and persisted only as SHA-256 hashes in PostgreSQL.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Switched auth repository construction to use `pgxpool.Pool`**
- **Found during:** Task 1
- **Issue:** Refresh rotation needed a transaction-capable repository, but `auth.NewRepository` previously received only prebuilt SQLC queries.
- **Fix:** Updated `core-backend/internal/app/app.go` and auth repository construction to inject `db.Pool()` so rotation could lock, create, and revoke sessions atomically.
- **Files modified:** `core-backend/internal/app/app.go`, `core-backend/internal/auth/repository.go`
- **Verification:** `go test ./internal/http ./internal/auth -run 'TestLogin|TestRefresh|TestLogout|TestMe' -count=1`
- **Committed in:** `2be10d9`

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Necessary to keep refresh rotation atomic without changing the planned auth contract surface.

## Issues Encountered

- SQLC regeneration introduced new generated models and query types; after aligning repository imports and rotation parameters, the auth packages compiled and tests passed.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 1 now has the backend session foundation required for later RBAC and frontend BFF work.
- Remaining Phase 1 plans still need server-side role enforcement and examination intake hardening.
