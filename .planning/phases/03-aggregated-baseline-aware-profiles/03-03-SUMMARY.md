---
phase: 03-aggregated-baseline-aware-profiles
plan: 03
subsystem: api
tags: [fastapi, pydantic, numpy, docker-compose, baseline]
requires:
  - phase: 03-aggregated-baseline-aware-profiles
    provides: published Phase 3 contracts and RED scaffolds for baseline integration
provides:
  - standalone ml-baseline FastAPI service
  - median and MAD based baseline deviation and update gating
  - internal-only compose wiring for baseline runtime
affects: [core-backend, aggregation, operator-results, compose]
tech-stack:
  added: [fastapi, pydantic, numpy, scipy, pytest, uvicorn]
  patterns: [compute-only service boundary, robust statistics gating, internal compose service wiring]
key-files:
  created: [ml-baseline/app/main.py, ml-baseline/app/schemas.py, ml-baseline/app/algorithms.py, ml-baseline/Dockerfile]
  modified: [ml-baseline/tests/test_service.py, ml-baseline/tests/test_algorithms.py, docker-compose.yml, .env.example, docs/01_contract.md, docs/02_implementation.md]
key-decisions:
  - "Baseline service remains compute-only and does not access PostgreSQL or RabbitMQ directly."
  - "Baseline refresh eligibility uses median plus MAD with an outlier freeze threshold instead of mean and standard deviation."
  - "Local compose wiring keeps ml-baseline internal-only by omitting host port publication."
patterns-established:
  - "Versioned Python service contracts mirror docs/01_contract.md and return persistence-ready metadata."
  - "Baseline snapshots are bounded by a recent-history window before computing centers and scales."
requirements-completed: [BASE-01, BASE-02, BASE-03]
duration: 9 min
completed: 2026-03-22
---

# Phase 3 Plan 3: Baseline Service Summary

**Standalone FastAPI baseline service with versioned schemas, median/MAD deviation scoring, and internal-only compose runtime**

## Performance

- **Duration:** 9 min
- **Started:** 2026-03-22T11:09:50Z
- **Completed:** 2026-03-22T11:19:02Z
- **Tasks:** 2
- **Files modified:** 12

## Accomplishments
- Added `ml-baseline` as a standalone FastAPI service with `POST /baseline/calculate` and `GET /health`.
- Implemented robust baseline math using median, MAD, bounded history, and outlier-gated updates.
- Wired the service into local compose with explicit `BASELINE_*` env vars and kept it internal-only.
- Synced `docs/01_contract.md` and appended the implementation log in `docs/02_implementation.md`.

## Task Commits

1. **Task 1: Define the baseline service contract and FastAPI surface** - `e1fee01` (test), `28bb06b` (feat)
2. **Task 2: Implement robust update gating and local compose wiring** - `510d041` (test), `791aa4c` (feat)

## Files Created/Modified
- `ml-baseline/app/main.py` - FastAPI app and baseline calculation endpoint.
- `ml-baseline/app/schemas.py` - Versioned request and response models.
- `ml-baseline/app/algorithms.py` - Robust statistics helpers, outlier freeze, and bounded update logic.
- `ml-baseline/Dockerfile` - Container runtime for the baseline service.
- `docker-compose.yml` - Internal-only `ml-baseline` service wiring.
- `.env.example` - Baseline runtime environment variables.
- `docs/01_contract.md` - Updated baseline response contract with metric scores and `next_baseline`.
- `docs/02_implementation.md` - Append-only implementation log entry for plan 03.

## Decisions Made
- Kept the baseline boundary HTTP-only and compute-only so PostgreSQL ownership stays in core backend.
- Returned `metric_scores` and `next_baseline` in the response so core can persist richer baseline snapshots without re-deriving math.
- Used a bounded `BASELINE_MAX_HISTORY` window to prevent uncontrolled baseline growth.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added a dedicated Dockerfile for `ml-baseline`**
- **Found during:** Task 2 (Implement robust update gating and local compose wiring)
- **Issue:** Compose wiring for the new baseline service could not be runnable without a container build definition.
- **Fix:** Added `ml-baseline/Dockerfile` following the existing Python service pattern and parameterized host/port startup via `BASELINE_*`.
- **Files modified:** `ml-baseline/Dockerfile`
- **Verification:** Focused pytest suite passed and compose YAML parse confirmed `ml-baseline` exists as a service.
- **Committed in:** `791aa4c` (part of task commit)

**2. [Rule 3 - Blocking] Used an isolated temporary Python venv for verification**
- **Found during:** Task 1 and Task 2 verification
- **Issue:** System Python in the local environment lacked `pytest`, which blocked required test execution.
- **Fix:** Created `/tmp/dimplom-ml-baseline-venv` outside the repository and installed only the packages needed to run the focused `ml-baseline` tests.
- **Files modified:** None in repository
- **Verification:** `pytest` commands for service and algorithm tests passed in the isolated environment.
- **Committed in:** No repository files changed

---

**Total deviations:** 2 auto-fixed (2 blocking)
**Impact on plan:** Both fixes were necessary to verify the new service and keep the runtime wiring reproducible. No scope creep beyond the plan boundary.

## Issues Encountered
- `docker compose config` could not run because the `docker` CLI is unavailable in this WSL distro. As a fallback, the compose file was parsed via Python to confirm that `ml-baseline` is present and has no published `ports`.
- Installing `scipy` into the temporary verification venv failed with a wheel filesystem error, but the repository dependency was still pinned in `ml-baseline/requirements.txt` and the focused tests did not require SciPy at runtime.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Core aggregation can now call a narrow baseline HTTP boundary and persist `update_eligibility`, `metric_scores`, and `next_baseline`.
- Compose and env wiring already reserve the internal service name and baseline runtime knobs for the upcoming core integration work.

## Self-Check
PASSED
