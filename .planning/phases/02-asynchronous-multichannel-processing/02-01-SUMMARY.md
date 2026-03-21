---
phase: 02-asynchronous-multichannel-processing
plan: 01
subsystem: api
tags: [rabbitmq, postgres, outbox, processing, testing]
requires:
  - phase: 01-trusted-access-and-intake
    provides: idempotent finish fence and processing-ready examination workflow
provides:
  - versioned Phase 2 HTTP and AMQP processing contracts
  - PostgreSQL schema skeleton for outbox, channel runs, and channel results
  - RED tests that pin mandatory fan-out and processing-status behavior
affects: [phase-02-plan-02, phase-02-plan-04, frontend-processing-ui]
tech-stack:
  added: [PostgreSQL migration 000006, sqlc query skeletons]
  patterns: [transactional outbox, backend-authoritative processing progress, versioned processing envelopes]
key-files:
  created:
    - core-backend/internal/processing/contracts.go
    - core-backend/internal/processing/service_test.go
    - core-backend/internal/http/processing_status_test.go
    - core-backend/migrations/000006_processing_pipeline.up.sql
    - core-backend/migrations/000006_processing_pipeline.down.sql
    - core-backend/db/queries/processing.sql
  modified:
    - docs/01_contract.md
    - docs/02_implementation.md
key-decisions:
  - "Phase 2 uses one versioned command envelope and one versioned result envelope for all mandatory channels."
  - "PostgreSQL, not RabbitMQ, remains the source of truth for processing progress and terminal failures."
  - "The existing examination_processing_launches fence stays in place and is extended by channel runs plus outbox rows."
patterns-established:
  - "Transactional outbox rows are created in the same finish transaction as channel-run state."
  - "Processing detail is exposed through GET /examinations/{id}/processing-status instead of overloading Examination."
requirements-completed: [PIPE-01]
duration: 10min
completed: 2026-03-21
---

# Phase 2 Plan 1: Asynchronous Multichannel Processing Summary

**Versioned processing contracts, RED fan-out tests, and PostgreSQL outbox/channel-run schema for the mandatory text, acoustic, and paralinguistic pipeline**

## Performance

- **Duration:** 10 min
- **Started:** 2026-03-21T06:46:00Z
- **Completed:** 2026-03-21T06:56:09Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments
- Documented Phase 2 HTTP and RabbitMQ contracts in one place, including versioned envelopes, topology, status vocabulary, and `/processing-status`.
- Added shared Go contract types plus RED tests that pin missing finish fan-out and missing processing-status endpoint behavior.
- Added durable PostgreSQL schema and sqlc query skeletons for outbox publishing intent, per-channel runtime state, and normalized channel results.

## Task Commits

Each task was committed atomically:

1. **Task 1: Lock Phase 2 contracts and test expectations** - `26bd802` (feat)
2. **Task 2: Add the PostgreSQL schema for outbox and per-channel runtime state** - `1b299b6` (feat)

## Files Created/Modified
- `docs/01_contract.md` - Phase 2 processing API, AMQP envelopes, topology, status vocabulary, and persistence contract.
- `docs/02_implementation.md` - Append-only implementation log entry for Phase 2 plan 01.
- `core-backend/internal/processing/contracts.go` - Shared channel constants and DTO/envelope types.
- `core-backend/internal/processing/service_test.go` - RED contract test for mandatory finish fan-out and channel-neutral result envelope.
- `core-backend/internal/http/processing_status_test.go` - RED endpoint test for backend-authoritative processing progress.
- `core-backend/migrations/000006_processing_pipeline.up.sql` - Outbox, channel-run, and channel-result schema plus expanded examination statuses.
- `core-backend/migrations/000006_processing_pipeline.down.sql` - Rollback for Phase 2 processing pipeline schema.
- `core-backend/db/queries/processing.sql` - SQL skeleton for atomic channel-run/outbox creation and processing status projection.

## Decisions Made
- Used `processing` and `failed` as coarse-grained examination statuses while keeping per-channel detail exclusively in `/processing-status`.
- Standardized on one unified result routing key and queue for all mandatory channels so result ingestion stays channel-neutral.
- Kept the existing `examination_processing_launches` fence as the idempotency guard instead of replacing it with a new mechanism.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- Existing migration filenames in the repository use both sequential and suffixed numbering (`000004z_*`), so the new migration was added as `000006_processing_pipeline` without renaming prior files.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 2 plan 02 can now extend `POST /examinations/{id}/finish` into real transactional outbox fan-out against the published contracts and schema.
- The targeted tests are intentionally RED and currently fail only because the actual publisher orchestration and `/processing-status` implementation are still absent.

## Self-Check: PASSED

- Found summary file: `.planning/phases/02-asynchronous-multichannel-processing/02-01-SUMMARY.md`
- Verified task commit `26bd802`
- Verified task commit `1b299b6`
