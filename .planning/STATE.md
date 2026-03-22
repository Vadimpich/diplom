---
gsd_state_version: 1.0
milestone: v1.0
milestone_name: milestone
status: unknown
stopped_at: Completed 03-aggregated-baseline-aware-profiles-03-PLAN.md
last_updated: "2026-03-22T11:20:27.125Z"
progress:
  total_phases: 5
  completed_phases: 2
  total_plans: 20
  completed_plans: 16
---

# Project State

## Project Reference

See: `.planning/PROJECT.md` (updated 2026-03-20)

**Core value:** Система должна давать оператору надёжный, интерпретируемый и воспроизводимый результат обследования специалиста, основанный на полном мультимодальном анализе речевых ответов, а не на ручной субъективной оценке.
**Current focus:** Phase 03 — aggregated-baseline-aware-profiles

## Current Position

Phase: 03 (aggregated-baseline-aware-profiles) — EXECUTING
Plan: 2 of 6

## Performance Metrics

**Velocity:**

- Total plans completed: 16
- Average duration: 9 min
- Total execution time: 2.4 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01-trusted-access-and-intake | 8 | 3500s | 437s |
| 02-asynchronous-multichannel-processing | 6 | 4343s | 724s |
| 03-aggregated-baseline-aware-profiles | 2 | 1152s | 576s |

**Recent Trend:**

- Last 5 plans: 02-02, 02-04, 02-05, 02-06, 03-01
- Trend: Improving

| Phase 02-asynchronous-multichannel-processing P01 | 600 | 2 tasks | 8 files |
| Phase 02-asynchronous-multichannel-processing P03 | 420 | 2 tasks | 13 files |
| Phase 02-asynchronous-multichannel-processing P02 | 2040 | 2 tasks | 13 files |
| Phase 02-asynchronous-multichannel-processing P04 | 6 min | 2 tasks | 10 files |
| Phase 02-asynchronous-multichannel-processing P05 | 338 | 2 tasks | 5 files |
| Phase 02-asynchronous-multichannel-processing P06 | 585 | 2 tasks | 8 files |
| Phase 03-aggregated-baseline-aware-profiles P01 | 600 | 2 tasks | 11 files |
| Phase 03-aggregated-baseline-aware-profiles P03 | 552 | 2 tasks | 12 files |

## Accumulated Context

### Decisions

Decisions are logged in `.planning/PROJECT.md`.
Recent decisions affecting current work:

- Roadmap compressed to 5 phases because `config.json` sets coarse granularity.
- Existing operator/admin CRUD and intake flows are treated as brownfield baseline; roadmap covers only target-architecture gaps.
- [Phase 01-trusted-access-and-intake]: Core backend returns refresh tokens to a trusted frontend BFF, which owns HttpOnly browser cookies.
- [Phase 01-trusted-access-and-intake]: Refresh tokens are opaque random values stored only as SHA-256 hashes in PostgreSQL refresh_sessions.
- [Phase 01-trusted-access-and-intake]: Examination creation snapshots questionnaire questions into immutable examination-scoped rows.
- [Phase 01-trusted-access-and-intake]: Answer uploads are bound to examination, snapshot question, and specialist; duplicate answers per question are rejected.
- [Phase 01-trusted-access-and-intake]: Finish is idempotent and fenced at the DB layer before processing starts.
- [Phase 01-trusted-access-and-intake]: Specialist history UI renders backend workflow statuses as the source of truth.
- [Phase 02-asynchronous-multichannel-processing]: Phase 2 uses one versioned command envelope and one versioned result envelope for all mandatory channels.
- [Phase 02-asynchronous-multichannel-processing]: PostgreSQL, not RabbitMQ, remains the source of truth for processing progress and terminal failures.
- [Phase 02-asynchronous-multichannel-processing]: The existing examination_processing_launches fence stays in place and is extended by channel runs plus outbox rows.
- [Phase 02-asynchronous-multichannel-processing]: Each mandatory channel now runs as an independent FastAPI plus aio-pika worker that consumes only its own queue and publishes the same result envelope shape.
- [Phase 02-asynchronous-multichannel-processing]: Workers bind their own queue and routing key from env so the local Compose stack remains reproducible without hidden broker bootstrap steps.
- [Phase 02-asynchronous-multichannel-processing]: Local stack stays on rabbitmq:3.13-management-alpine; retry and DLX behavior remain explicit contract assumptions instead of relying on RabbitMQ 4 defaults.
- [Phase 02-asynchronous-multichannel-processing]: Finish persists channel runs and outbox rows in the same PostgreSQL transaction as the processing launch fence.
- [Phase 02-asynchronous-multichannel-processing]: RabbitMQ command topology is declared explicitly for 3.13 with quorum queues, DLX, delivery-limit, and publisher confirms.
- [Phase 02-asynchronous-multichannel-processing]: Unified result messages are consumed only by core backend; channelresults owns persisted channel-run and examination failure transitions. — This keeps RabbitMQ workers dumb and preserves PostgreSQL as the single source of truth for retry/error state.
- [Phase 02-asynchronous-multichannel-processing]: Processing-status HTTP responses derive started_at, failed_at, terminal state, and conflict gating from PostgreSQL instead of broker state. — The operator progress endpoint must remain backend-authoritative and usable without RabbitMQ management or inferred frontend workflow.
- [Phase 02-asynchronous-multichannel-processing]: Frontend mirrors backend snake_case processing progress DTOs and examination statuses directly in typed operator UI.
- [Phase 02-asynchronous-multichannel-processing]: TanStack Query polls GET /examinations/{id}/processing-status every 3 seconds and stops only when terminal=true.
- [Phase 02-asynchronous-multichannel-processing]: Core backend compose wiring now sets explicit RABBITMQ_URL because the Go runtime ignores host/port fragments without a full broker URL.
- [Phase 02-asynchronous-multichannel-processing]: Workers declare quorum/DLX queue arguments identical to the relay so RabbitMQ topology remains reproducible across repeated local startups.
- [Phase 02-asynchronous-multichannel-processing]: Frontend Phase 2 validation is documented as lint -> build -> tsc because App Router type artifacts are generated by build.
- [Phase 03-aggregated-baseline-aware-profiles]: Aggregation remains core-owned in Go; baseline stays a separate Python compute boundary.
- [Phase 03-aggregated-baseline-aware-profiles]: Phase 3 terminal success is `aggregated`, not mere channel completion.
- [Phase 03-aggregated-baseline-aware-profiles]: Result and history DTOs use canonical proxy-oriented snapshots instead of raw worker payloads.
- [Phase 03-aggregated-baseline-aware-profiles]: Baseline service remains compute-only and never accesses PostgreSQL directly.
- [Phase 03-aggregated-baseline-aware-profiles]: Baseline refresh eligibility uses median plus MAD with outlier freeze instead of mean/stddev.
- [Phase 03-aggregated-baseline-aware-profiles]: Local compose wiring keeps ml-baseline internal-only without published host ports.

### Pending Todos

None yet.

### Blockers/Concerns

- Aggregator, baseline, KЭСМИ integration, and final operator result views are still not implemented; Phase 3 starts from a now-reproducible async pipeline baseline.
- RED coverage now includes aggregation contract scaffolds, result-history endpoint expectations, and baseline update gating, but the corresponding implementations are still absent.

## Session Continuity

Last session: 2026-03-22T11:20:27.122Z
Stopped at: Completed 03-aggregated-baseline-aware-profiles-03-PLAN.md
Resume file: None
