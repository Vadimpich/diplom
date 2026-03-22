---
phase: 03-aggregated-baseline-aware-profiles
plan: 01
subsystem: api
tags: [postgres, go, pytest, contracts, aggregation, baseline]
requires:
  - phase: 02-asynchronous-multichannel-processing
    provides: versioned channel result envelopes, processing-status projection, persisted channel results
provides:
  - Phase 3 HTTP contracts for aggregated result, result history, and baseline service
  - PostgreSQL schema skeleton for aggregated profiles and baseline snapshots
  - RED tests that pin aggregating/aggregated semantics before implementation
affects: [frontend, core-backend, ml-baseline, result-history, baseline]
tech-stack:
  added: []
  patterns: [core-owned aggregation persistence, compute-only baseline boundary, canonical aggregated DTO]
key-files:
  created: [core-backend/internal/aggregation/service_test.go, core-backend/internal/http/results_handler_test.go, ml-baseline/tests/test_service.py, ml-baseline/tests/test_algorithms.py, core-backend/migrations/000007_aggregated_profiles.up.sql, core-backend/migrations/000007_aggregated_profiles.down.sql, core-backend/db/queries/aggregation.sql, core-backend/internal/aggregation/contracts.go, .planning/phases/03-aggregated-baseline-aware-profiles/03-01-SUMMARY.md]
  modified: [docs/01_contract.md, docs/02_implementation.md]
key-decisions:
  - "Aggregation remains core-owned in Go; only baseline crosses into a Python compute boundary."
  - "Phase 3 terminal success is `aggregated`, not merely all channels succeeded."
  - "Metric names stay neutral and proxy-oriented until real ML semantics exist."
patterns-established:
  - "Canonical aggregated profile contracts are published in docs before backend/frontend implementation."
  - "Result and history surfaces must consume persisted aggregated snapshots instead of raw worker payloads."
requirements-completed: [AGGR-02, AGGR-03, BASE-01, BASE-02, RSLT-03]
duration: 10 min
completed: 2026-03-22
---

# Phase 3 Plan 01: Aggregated Profile Contracts, RED Tests, and Persistence Skeleton Summary

**Canonical aggregated result contracts, RED verification scaffolds, and PostgreSQL schema skeletons for baseline-aware profiles**

## Performance

- **Duration:** 10 min
- **Started:** 2026-03-22T10:58:00Z
- **Completed:** 2026-03-22T11:08:28Z
- **Tasks:** 2
- **Files modified:** 11

## Accomplishments

- Published Phase 3 contracts for `aggregating` / `aggregated`, `GET /examinations/{id}/result`, `GET /specialists/{id}/result-history`, and `POST /baseline/calculate`.
- Added RED tests in Go and Python that now fail on missing aggregation, missing result-history endpoint, and missing baseline implementation instead of missing contract artifacts.
- Added the authoritative Phase 3 database skeleton for aggregated profiles, metric snapshots, channel contributions, explanation rows, specialist baseline state, and examination baseline snapshots.

## Task Commits

Each task was committed atomically:

1. **Task 1: Lock the Phase 3 contracts and status vocabulary** - `f4bda7e` (test)
2. **Task 2: Add the authoritative PostgreSQL schema and query skeletons** - `2030e7a` (feat)

## Files Created/Modified

- `docs/01_contract.md` - Phase 3 status, result, history, and baseline-service contracts.
- `core-backend/internal/aggregation/service_test.go` - RED scaffold for canonical profile persistence.
- `core-backend/internal/http/results_handler_test.go` - RED HTTP coverage for `aggregating` semantics and specialist result history.
- `ml-baseline/tests/test_service.py` - RED baseline service contract test.
- `ml-baseline/tests/test_algorithms.py` - RED baseline outlier-gating test.
- `core-backend/migrations/000007_aggregated_profiles.up.sql` - Phase 3 schema for aggregated profile and baseline tables.
- `core-backend/migrations/000007_aggregated_profiles.down.sql` - rollback for Phase 3 schema and status vocabulary.
- `core-backend/db/queries/aggregation.sql` - query skeletons for readiness, persistence, baseline state, and history projection.
- `core-backend/internal/aggregation/contracts.go` - stable Go DTOs and metric keys for aggregation and baseline boundaries.
- `docs/02_implementation.md` - append-only implementation log for this plan.

## Decisions Made

- Aggregation stays inside core backend because Phase 2 already made PostgreSQL and Go orchestration authoritative for channel-result workflow.
- Baseline remains a narrow Python service boundary and must not read or write core PostgreSQL directly.
- Public result/history DTOs intentionally hide raw stub-worker payloads and expose only canonical proxy-oriented fields.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- Python verification could not run because the environment does not have `pytest` installed (`/usr/bin/python3: No module named pytest`).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Ready for Plan 02 to implement aggregation readiness, `aggregating` transitions, and canonical profile persistence against the published contracts and schema.
- Remaining expected RED failures are intentional: aggregation logic, result-history routing, and baseline service implementation are still absent.

## Self-Check

PASSED

---
*Phase: 03-aggregated-baseline-aware-profiles*
*Completed: 2026-03-22*
