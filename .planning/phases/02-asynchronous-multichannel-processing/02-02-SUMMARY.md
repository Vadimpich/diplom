---
phase: 02-asynchronous-multichannel-processing
plan: 02
subsystem: api
tags: [postgres, rabbitmq, outbox, amqp091-go, retries]
requires:
  - phase: 02-asynchronous-multichannel-processing
    provides: versioned processing contracts, pipeline tables, and finish-fence scaffolding from plan 01
provides:
  - transactional finish fan-out into examination channel runs and processing outbox rows
  - background RabbitMQ relay with publisher confirms and explicit 3.13 queue arguments
  - backend processing-status route wired to PostgreSQL channel state
affects: [02-04, 02-05, processing, operator-ui]
tech-stack:
  added: [github.com/rabbitmq/amqp091-go]
  patterns: [transactional outbox, publisher confirms, PostgreSQL-backed retry ledger]
key-files:
  created: [core-backend/internal/processing/service.go, core-backend/internal/processing/publisher.go, core-backend/internal/processing/publisher_test.go]
  modified: [core-backend/internal/processing/repository.go, core-backend/internal/config/config.go, core-backend/internal/http/examinations_handler.go, core-backend/internal/http/router.go, docs/01_contract.md, docs/02_implementation.md]
key-decisions:
  - "Finish continues to use PostgreSQL as the single source of truth by creating channel runs and outbox rows in the same transaction as the launch fence."
  - "RabbitMQ topology is declared explicitly for 3.13 compatibility with quorum queues, DLX, and delivery-limit instead of relying on broker defaults."
patterns-established:
  - "Processing launch pattern: HTTP finish persists intent only; background relay performs broker publication asynchronously."
  - "Retry visibility pattern: publish confirms and failures update PostgreSQL outbox/channel-run state before any UI reads progress."
requirements-completed: [PIPE-01, PIPE-03]
duration: 34 min
completed: 2026-03-21
---

# Phase 02 Plan 02: Transactional Outbox Relay Summary

**Transactional finish fan-out with PostgreSQL outbox persistence and a confirmed RabbitMQ relay for mandatory processing channels**

## Performance

- **Duration:** 34 min
- **Started:** 2026-03-21T06:39:00Z
- **Completed:** 2026-03-21T07:13:34Z
- **Tasks:** 2
- **Files modified:** 13

## Accomplishments
- `POST /examinations/{id}/finish` now persists mandatory channel runs and versioned outbox commands atomically, without publishing from the request path.
- Core backend now runs a background outbox relay that declares RabbitMQ 3.13-compatible command topology and marks rows published only after publisher confirms.
- Processing progress plumbing is wired through `GET /examinations/{id}/processing-status`, backed by PostgreSQL channel state.

## Task Commits

Each task was committed atomically:

1. **Task 1: Extend the finish fence into outbox-backed processing launch** - `bd37b43` (feat)
2. **Task 2: Build the RabbitMQ outbox relay with explicit 3.13 queue semantics** - `a3f78be` (feat)

## Files Created/Modified
- `core-backend/internal/processing/service.go` - Processing service for finish orchestration and status reads.
- `core-backend/internal/processing/repository.go` - Transactional finish/outbox persistence plus relay state updates and processing-status projection.
- `core-backend/internal/processing/publisher.go` - Background RabbitMQ relay with topology declaration, confirms, and retry classification.
- `core-backend/internal/processing/publisher_test.go` - Relay tests for confirms, retry budget, and fatal vs temporary failures.
- `core-backend/internal/config/config.go` - Broker URL, outbox polling interval, and retry budget env config.
- `core-backend/internal/http/examinations_handler.go` - Finish delegation to processing service and processing-status handler.
- `core-backend/internal/http/router.go` - Route wiring for finish relay and processing-status endpoint.
- `docs/01_contract.md` - Updated broker topology and env contract.
- `docs/02_implementation.md` - Append-only implementation log entries for outbox relay work.

## Decisions Made
- Finish remains fast and idempotent by storing broker intent in PostgreSQL, then letting a background relay publish later.
- Relay success updates both `processing_outbox` and `examination_channel_runs`, while temporary and fatal publish errors stay visible in PostgreSQL for safe retries.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Wired the missing `processing-status` route and handler**
- **Found during:** Task 2 verification
- **Issue:** `go test ./...` failed because the phase already had a contract/test for `GET /examinations/{id}/processing-status`, but router wiring was absent.
- **Fix:** Added handler wiring plus PostgreSQL-backed status read path, with safe router fallback for tests that construct `Dependencies` without a processing service.
- **Files modified:** `core-backend/internal/processing/service.go`, `core-backend/internal/processing/repository.go`, `core-backend/internal/http/examinations_handler.go`, `core-backend/internal/http/router.go`, `docs/02_implementation.md`
- **Verification:** `cd /home/vadim/diplom/core-backend && go test ./...`
- **Committed in:** `a3f78be`

---

**Total deviations:** 1 auto-fixed (1 missing critical)
**Impact on plan:** Auto-fix stayed inside the same processing slice and removed a correctness gap already covered by repository tests.

## Issues Encountered
- `go test ./...` exposed a nil-interface trap in router wiring for optional processing dependencies; constructor logic now sets interfaces only when a concrete processing service is present.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Core backend now produces and relays durable channel commands with bounded publish retries, so result-consumer and progress phases can build on persisted broker IDs and channel-run state.
- Examination failure promotion and full channel-result ingestion remain for later plans.

## Self-Check: PASSED
