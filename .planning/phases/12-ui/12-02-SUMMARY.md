---
phase: 12-ui
plan: 02
subsystem: api
tags: [postgres, sqlc, go, contracts, nextjs]
requires:
  - phase: 11-operator-ux-visual-completion
    provides: operator/admin UI surfaces that needed denser truthful registry data
provides:
  - enriched user DTOs with persisted last login timestamps
  - questionnaire registry metadata with usage and last editor fields
  - specialist registry summaries derived from examinations, profiles, and baseline state
affects: [12-03, 12-04, operator-ui, admin-ui]
tech-stack:
  added: []
  patterns: [persist only UI-critical metadata, derive registry summaries in SQL read models]
key-files:
  created:
    - core-backend/migrations/000011_ui_registry_support.up.sql
    - core-backend/migrations/000011_ui_registry_support.down.sql
  modified:
    - docs/01_contract.md
    - core-backend/db/queries/auth.sql
    - core-backend/db/queries/questionnaires.sql
    - core-backend/db/queries/specialists.sql
    - frontend/lib/api/types.ts
key-decisions:
  - "User last login is persisted on users.last_login_at and updated only after successful POST /auth/login."
  - "Questionnaire last editor metadata is persisted on questionnaires, while usage_count and specialist summaries stay derived from existing examination/profile/baseline tables."
patterns-established:
  - "Registry-only metadata is added to existing list/detail DTOs instead of creating new summary endpoints."
  - "Nullable SQL read-model facts can use internal sentinel values and repository mapping when sqlc nullability inference is too strict, while the external JSON contract stays nullable."
requirements-completed: [ADMN-01, QSTR-02, MONR-01, RICH-01, AMON-01]
duration: 18 min
completed: 2026-03-25
---

# Phase 12 Plan 02 Summary

**Registry-facing backend contracts now expose real login, questionnaire, and specialist summary metadata for dense admin/operator UI surfaces**

## Performance

- **Duration:** 18 min
- **Started:** 2026-03-25T08:15:00Z
- **Completed:** 2026-03-25T08:33:24Z
- **Tasks:** 3
- **Files modified:** 23

## Accomplishments
- Added persisted support for `last_login_at` and questionnaire editor metadata via migration `000011_ui_registry_support`.
- Enriched backend list/detail DTOs so `/users`, `/questionnaires`, and `/specialists` return truthful registry metadata from PostgreSQL-backed reads.
- Extended frontend API types to expose the new snake_case contract fields without introducing UI-side derivations.

## Task Commits

1. **Task 1: Publish the enriched UI-support contracts per Phase 12 scope** - `f0f573a` (feat)
2. **Task 2: Implement backend DTO enrichments for users, questionnaires, and specialists** - `5ba7133` (feat)
3. **Task 2 TDD follow-up: cover last login auth metadata** - `6231255` (test)
4. **Task 3: Wire the enriched contracts into the frontend API boundary** - `d4ab811` (feat)

## Files Created/Modified
- `core-backend/migrations/000011_ui_registry_support.up.sql` - adds persisted columns for user last login and questionnaire editor metadata
- `core-backend/migrations/000011_ui_registry_support.down.sql` - rolls back the UI registry support schema
- `docs/01_contract.md` - documents enriched DTO fields for users, questionnaires, and specialists
- `core-backend/db/queries/auth.sql` - reads and updates `last_login_at`
- `core-backend/db/queries/questionnaires.sql` - exposes usage and editor metadata on questionnaire reads
- `core-backend/db/queries/specialists.sql` - derives registry summary fields from examinations, aggregated profiles, and baseline state
- `core-backend/internal/auth/service.go` - updates last login timestamp after successful login
- `core-backend/internal/questionnaires/repository.go` - maps usage/editor metadata into API DTOs
- `core-backend/internal/specialists/repository.go` - maps specialist summary fields into API DTOs
- `frontend/lib/api/types.ts` - adds typed contract fields for enriched registries

## Decisions Made
- Persisted only the metadata that cannot be reconstructed honestly at read time: `users.last_login_at` and questionnaire editor tracking.
- Kept specialist summary data read-only and SQL-derived to avoid a new summary service or handler-side N+1 logic.
- Reused existing list/detail endpoints so Phase 12 UI work can densify existing screens without new route contracts.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Worked around sqlc nullability inference for registry read models**
- **Found during:** Task 2
- **Issue:** `sqlc` inferred some nullable aggregate/join fields as non-nullable or `interface{}`, which would either fail scans or leak awkward types into repositories.
- **Fix:** Used internal SQL sentinel values for absent timestamps/ids and normalized them back to nullable Go pointers in repository mapping.
- **Files modified:** `core-backend/db/queries/questionnaires.sql`, `core-backend/db/queries/specialists.sql`, `core-backend/internal/questionnaires/repository.go`, `core-backend/internal/specialists/repository.go`
- **Verification:** `go test ./internal/http ./internal/auth ./internal/questionnaires ./internal/specialists`
- **Committed in:** `5ba7133`

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Kept scope unchanged and preserved the nullable external contract without adding new architectural layers.

## Issues Encountered

- The repository started from a dirty worktree, including unrelated frontend/admin artifacts in several files. Task commits were limited to `12-02` work only.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 12 UI plans can now render dense registries with real metadata for user activity, questionnaire usage, and specialist recency/baseline context.
- No backend blocker remains for registry density inside the existing `/users`, `/questionnaires`, and `/specialists` routes.

## Self-Check: PASSED
