---
phase: 04-decision-delivery-to-operator
plan: 01
subsystem: api
tags: [kesmi, contracts, postgres, sqlc, decision]
requires:
  - phase: 04-decision-delivery-to-operator
    provides: Wave 0 RED scaffolds and smoke artifacts
provides:
  - Phase 4 decision status vocabulary and public contracts
  - Stable Go DTOs for decision_input and decision_result
  - PostgreSQL decision snapshot plus attempt ledger schema
  - Generated sqlc query layer for future decision repository
affects: [04-02, 04-03, decision, kesmi, results]
tech-stack:
  added: [sqlc]
  patterns: [contract-first decision DTOs, append-only attempt ledger]
key-files:
  created:
    - core-backend/internal/decision/contracts.go
    - core-backend/migrations/000008_decision_delivery.up.sql
    - core-backend/migrations/000008_decision_delivery.down.sql
    - core-backend/db/queries/decision.sql
    - core-backend/db/sqlc/decision.sql.go
  modified:
    - docs/01_contract.md
    - core-backend/internal/decision/service_test.go
    - core-backend/internal/kesmi/client_test.go
    - core-backend/internal/http/results_handler_test.go
    - docs/02_implementation.md
key-decisions:
  - "Phase 4 public API stays backend-owned and exposes normalized decision_result instead of raw WiMi payloads."
  - "Decision delivery persistence uses one latest decision snapshot plus append-only decision_attempts ledger."
patterns-established:
  - "DecisionInput is built from aggregated profile plus baseline snapshot and versioned independently from WiMi model parameters."
  - "Decision states use pending/succeeded/transport_exhausted/business_error, while coarse examination status advances to decision_pending/completed."
requirements-completed: [KSMI-01, KSMI-02]
duration: 25min
completed: 2026-03-23
---

# Phase 4 Plan 01 Summary

**Phase 4 contract-first decision delivery foundation with normalized DTOs, durable snapshot-plus-attempt schema, and generated sqlc accessors**

## Performance

- **Duration:** 25 min
- **Tasks:** 2
- **Files modified:** 10+

## Accomplishments

- Phase 4 status vocabulary in [docs/01_contract.md](/home/vadim/diplom/docs/01_contract.md) now includes `decision_pending` and `completed`, plus explicit `decision_input` and `decision_result`.
- Added [contracts.go](/home/vadim/diplom/core-backend/internal/decision/contracts.go) as the stable Go contract for decision DTOs and constants.
- Added migration/query skeletons for `decision_snapshots` and append-only `decision_attempts`, then generated [decision.sql.go](/home/vadim/diplom/core-backend/db/sqlc/decision.sql.go).
- Replaced placeholder Phase 4 test names with contract-aware RED scaffolds in decision, kesmi, and HTTP result handler tests.

## Verification

- `sqlc generate`
- `rg -n 'decision_input|decision_result|decision_pending|completed|correlation_id|transport_exhausted' docs/01_contract.md`
- `rg -n 'type DecisionInput struct|type DecisionResult struct|const DecisionStatePending = "pending"|const DecisionRecommendationUnavailable = "unavailable"' core-backend/internal/decision/contracts.go`
- `rg -n 'CreateDecisionSnapshot|InsertDecisionAttempt|ListPendingDecisionSnapshots|GetExaminationDecisionProjection' core-backend/db/queries/decision.sql core-backend/db/sqlc/decision.sql.go`
- `go test ./internal/decision ./internal/kesmi ./internal/http -run 'TestCreateDecisionInput|TestDecisionPendingTransition|TestRetriesOnlyTransportFailures|TestExaminationResultIncludesDecisionBlock' -count=1` currently fails by design because runtime delivery and result projection are not implemented yet.

## Issues Encountered

- `sqlc` was missing in the environment at first; generation succeeded after the tool became available locally.
- The worktree is still heavily dirty, so this plan was documented without an atomic git commit to avoid mixing unrelated changes.

## Next Phase Readiness

- `04-02` can now implement the real decision service, classifier, relay, and DB repository without inventing new statuses, DTO fields, or table names.
- `04-03` can project the already documented `decision` block into `/examinations/{id}/result` and operator UI.
