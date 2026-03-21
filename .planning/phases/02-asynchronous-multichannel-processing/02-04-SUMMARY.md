---
phase: 02-asynchronous-multichannel-processing
plan: 04
subsystem: api
tags: [go, postgres, rabbitmq, processing, http]
requires:
  - phase: 02-asynchronous-multichannel-processing
    provides: transactional outbox fan-out and runnable channel workers from plans 02-03
provides:
  - unified RabbitMQ result consumer wired into core backend
  - persisted channel-run retry and terminal failure state machine
  - backend-authoritative `/examinations/{id}/processing-status` projection from PostgreSQL
affects: [phase-02-plan-05, operator-processing-ui, processing-runtime]
tech-stack:
  added: []
  patterns: [core-owned channel result state machine, postgres-derived processing projection]
key-files:
  created:
    - core-backend/internal/channelresults/service.go
    - core-backend/internal/processing/results_consumer.go
    - .planning/phases/02-asynchronous-multichannel-processing/02-04-SUMMARY.md
  modified:
    - core-backend/internal/app/app.go
    - core-backend/internal/processing/repository.go
    - core-backend/internal/http/processing_status_test.go
    - core-backend/internal/http/helpers.go
    - core-backend/internal/examinations/service.go
    - docs/02_implementation.md
key-decisions:
  - "Unified result messages are consumed only by core backend; `channelresults.Service` owns persisted channel-run and examination failure transitions."
  - "Processing-status HTTP responses derive `started_at`, `failed_at`, terminal state, and conflict gating from PostgreSQL instead of broker state."
patterns-established:
  - "Result handling pattern: validate AMQP envelope in `processing`, commit domain state transitions in `channelresults`."
  - "HTTP runtime projection pattern: coarse examination status stays on `Examination`, rich per-channel detail lives only in `/processing-status`."
requirements-completed: [PIPE-03, PIPE-04, RSLT-01]
duration: 6 min
completed: 2026-03-21
---

# Phase 02 Plan 04: Result Ingestion Summary

**Unified result consumption with persisted retry ledger, mandatory-channel failure projection, and PostgreSQL-backed processing status DTO**

## Performance

- **Duration:** 6 min
- **Started:** 2026-03-21T07:18:48Z
- **Completed:** 2026-03-21T07:24:56Z
- **Tasks:** 2
- **Files modified:** 10

## Accomplishments
- Core backend now consumes `processing.results` and persists channel result outcomes through one authoritative state machine.
- Mandatory channel fatal errors or retry exhaustion now project the parent examination into final `failed` state.
- `GET /examinations/{id}/processing-status` now exposes per-channel runtime data and pipeline conflict handling from PostgreSQL only.

## Task Commits

Each task was committed atomically:

1. **Task 1: Persist unified channel results and bounded retry state** - `68e5e06` (test), `32772cf` (feat)
2. **Task 2: Expose backend-authoritative processing progress over HTTP** - `af16b9e` (test), `3fcb890` (feat)

## Files Created/Modified
- `core-backend/internal/channelresults/service.go` - Channel result application service plus transactional PostgreSQL repository adapter.
- `core-backend/internal/processing/results_consumer.go` - RabbitMQ consumer for unified channel result envelopes with envelope validation and ack/nack handling.
- `core-backend/internal/processing/repository.go` - PostgreSQL processing-status projection with conflict gating and derived timestamps.
- `core-backend/internal/app/app.go` - Runtime wiring for result consumer alongside the existing outbox relay.
- `core-backend/internal/http/processing_status_test.go` - Stronger endpoint tests for active progress, terminal failure payload, and unavailable pipeline conflict.
- `docs/02_implementation.md` - Append-only implementation log for async result ingestion and status projection.

## Decisions Made

- Core backend, not workers, is the sole owner of persisted channel result classification and examination failure projection, keeping RabbitMQ consumers thin and deterministic.
- `processing-status` returns `409 Conflict` before an examination enters the async pipeline, matching the documented contract instead of exposing pre-pipeline coarse statuses through the runtime endpoint.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Added explicit unavailable-pipeline conflict for processing status**
- **Found during:** Task 2 (Expose backend-authoritative processing progress over HTTP)
- **Issue:** `GetProcessingStatus` could return coarse examination rows even when the examination had not entered the processing pipeline, violating the documented `409 Conflict` contract.
- **Fix:** Added `processing.ErrProcessingStatusUnavailable`, mapped it to HTTP 409, and gated PostgreSQL projections to `ready_for_processing`, `processing`, and `failed`.
- **Files modified:** `core-backend/internal/processing/service.go`, `core-backend/internal/processing/repository.go`, `core-backend/internal/http/helpers.go`, `core-backend/internal/http/processing_status_test.go`
- **Verification:** `go test ./internal/http -run 'TestProcessingStatusEndpoint|TestProcessingStatusEndpointReturnsTerminalError|TestProcessingStatusEndpointRejectsUnavailablePipeline' -count=1`
- **Committed in:** `3fcb890`

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Fix kept implementation aligned with the existing contract and avoided leaking non-pipeline states through the operator progress surface.

## Issues Encountered

- `python` was unavailable in the shell while computing summary metrics; switched to `node` for timestamp math without affecting code changes.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 2 plan 05 can poll one stable backend endpoint for per-channel runtime and terminal failure details.
- Result ingestion, retry ledger, and examination failure projection are now persisted in core backend state; remaining Phase 2 work is frontend polling and full-stack verification.

## Self-Check: PASSED

---
*Phase: 02-asynchronous-multichannel-processing*
*Completed: 2026-03-21*
