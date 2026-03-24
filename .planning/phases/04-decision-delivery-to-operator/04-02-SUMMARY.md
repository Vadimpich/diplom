---
phase: 04-decision-delivery-to-operator
plan: 02
subsystem: api
tags: [kesmi, relay, retries, postgres, go]
requires:
  - phase: 04-decision-delivery-to-operator
    provides: Phase 4 contracts and decision schema skeleton
provides:
  - Persisted decision snapshot creation after aggregation
  - Restart-safe decision relay inside core-backend
  - WiMi classifier and client boundary
  - Honest pre-model completion semantics with diagnostics
affects: [04-03, 04-04, results, operator-ui]
tech-stack:
  added: []
  patterns: [core-owned decision relay, normalized external failure classification]
key-files:
  created:
    - core-backend/internal/decision/service.go
    - core-backend/internal/decision/repository.go
    - core-backend/internal/decision/relay.go
    - core-backend/internal/decision/relay_test.go
    - core-backend/internal/kesmi/client.go
    - core-backend/internal/kesmi/classifier.go
  modified:
    - core-backend/internal/decision/service_test.go
    - core-backend/internal/aggregation/service.go
    - core-backend/internal/aggregation/service_test.go
    - core-backend/internal/config/config.go
    - core-backend/internal/app/app.go
    - docs/02_implementation.md
key-decisions:
  - "Decision delivery remains inside core-backend as a module plus background relay, not a standalone wrapper service."
  - "Before the real WiMi model exists, terminal success still uses recommendation=unavailable and message=analysis_not_implemented_yet."
patterns-established:
  - "Aggregation finalization immediately creates a pending decision snapshot and flips the examination into decision_pending."
  - "Retryable transport failures remain pending until the retry budget is exhausted; terminal business and exhausted transport failures complete the examination with diagnostics."
requirements-completed: [KSMI-01, KSMI-02]
duration: 35min
completed: 2026-03-23
---

# Phase 4 Plan 02 Summary

**Core-owned decision relay with persisted pending snapshots, WiMi error classification, and honest completed projection before a real model exists**

## Accomplishments

- Added a persisted decision delivery subsystem in `core-backend` via `service.go`, `repository.go`, and `relay.go`.
- Wired aggregation finalization to start decision delivery immediately after Phase 3 completion by creating a `decision_pending` snapshot.
- Added `internal/kesmi` client/classifier boundary so retryable transport failures and terminal business failures are normalized before they touch operator-facing contracts.
- Extended runtime config and app wiring so the decision relay starts with the backend process and uses env-driven `KESMI_*` settings.

## Verification

- `cd /home/vadim/diplom/core-backend && go test ./internal/decision ./internal/kesmi ./internal/aggregation -run 'TestDecisionPendingTransition|TestRetriesOnlyTransportFailures|TestDecisionSuccessMarksCompleted|TestDecisionBusinessErrorMarksCompleted|TestDecisionTransportExhaustedMarksCompleted|TestDecisionRelayResumesPendingSnapshots|TestAggregationStartsDecisionDeliveryAfterFinalize' -count=1`
- `cd /home/vadim/diplom/core-backend && go test ./internal/app -count=1`

## Issues Encountered

- The worktree remains dirty, so this plan was documented without an atomic commit to avoid mixing unrelated user changes.

## Next Phase Readiness

- `04-03` can now project real `decision_result` data into `/examinations/{id}/result` and the operator UI.
- `04-04` can wire WiMi into Compose knowing the backend already consumes `KESMI_*` runtime config and has a dedicated client boundary.
