---
phase: 03-aggregated-baseline-aware-profiles
plan: 04
type: execute
wave: 3
depends_on:
  - 03-02
  - 03-03
files_modified:
  - core-backend/internal/baselineclient/client.go
  - core-backend/internal/baselineclient/client_test.go
  - core-backend/internal/aggregation/service.go
  - core-backend/internal/aggregation/repository.go
  - core-backend/internal/config/config.go
  - core-backend/internal/app/app.go
  - core-backend/internal/processing/repository.go
  - core-backend/internal/processing/service.go
autonomous: true
requirements:
  - AGGR-01
  - BASE-01
  - BASE-02
  - BASE-03
must_haves:
  truths:
    - Core backend calls the baseline service after canonical profile assembly and persists the returned deviations and baseline metadata in PostgreSQL.
    - Examination reaches terminal success only after baseline-enriched persistence completes and the status becomes `aggregated`.
    - Personal baseline updates remain idempotent and controlled even when result ingestion is replayed.
  artifacts:
    - core-backend/internal/baselineclient/client.go performs the narrow HTTP call with timeout and transport-safe error handling.
    - core-backend/internal/aggregation/repository.go stores examination-level baseline snapshots and specialist baseline state.
    - core-backend/internal/app/app.go and config wire the baseline client into the runtime.
  key_links:
    - Go core must remain the only writer of baseline state in PostgreSQL.
    - `processing-status` terminal semantics must change only after baseline persistence succeeds.
    - Baseline transport errors must not surface raw Python traces or partially written profile state.
---

<objective>
Integrate the Go aggregation flow with the Python baseline service and finalize `aggregated` as the Phase 3 terminal success state.

Purpose: close the loop between canonical aggregation and persisted baseline-aware results while preserving PostgreSQL as the source of truth.
Output: baseline HTTP client, persisted deviations and metadata, and final aggregated status semantics.
</objective>

<execution_context>
@/home/vadim/.codex/get-shit-done/workflows/execute-plan.md
@/home/vadim/.codex/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/03-aggregated-baseline-aware-profiles/03-RESEARCH.md
@docs/01_contract.md
@core-backend/internal/app/app.go
@core-backend/internal/config/config.go
@core-backend/internal/aggregation/service.go
@core-backend/internal/processing/repository.go
@docker-compose.yml

<interfaces>
From core-backend/internal/app/app.go:
```go
processingRepository := processing.NewRepository(db.Pool(), cfg.S3Bucket)
processingService := processing.NewService(processingRepository)
```

Phase 3 adds:
```text
baseline client config
aggregation service that can call the baseline client
aggregated terminal success after baseline snapshot persistence
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Add the baseline HTTP client and runtime configuration</name>
  <files>core-backend/internal/baselineclient/client.go, core-backend/internal/baselineclient/client_test.go, core-backend/internal/config/config.go, core-backend/internal/app/app.go</files>
  <behavior>
    - Test 1: client sends the versioned request envelope expected by `ml-services/ml-baseline`.
    - Test 2: client maps transport and malformed-response errors to safe backend errors without leaking stack traces.
    - Test 3: runtime wiring can inject baseline URL, timeout, and algorithm version through env/config.
  </behavior>
  <action>Create `internal/baselineclient` as a narrow HTTP adapter with explicit timeout handling and JSON schema validation against the contract from plan 01. Add config fields and env support for baseline base URL, timeout, general reference version, and algorithm version defaults. Wire the client into `internal/app/app.go` so later aggregation logic can call it without reaching into compose or env directly. Keep this client transport-only; baseline state must still be persisted by the aggregation repository, not by the client.</action>
  <verify>
    <automated>cd /home/vadim/diplom/core-backend && go test ./internal/baselineclient -run 'TestClientSendsBaselineRequest|TestClientHandlesTransportFailure' -count=1</automated>
  </verify>
  <done>The core backend has a tested baseline HTTP client and runtime config needed to talk to `ml-services/ml-baseline` safely.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Persist baseline snapshots and finalize `aggregated` status semantics</name>
  <files>core-backend/internal/aggregation/service.go, core-backend/internal/aggregation/repository.go, core-backend/internal/processing/repository.go, core-backend/internal/processing/service.go, core-backend/internal/aggregation/service_test.go</files>
  <behavior>
    - Test 1: successful aggregation calls baseline, stores general and personal deviations, and marks the examination `aggregated`.
    - Test 2: `processing-status` becomes terminal only after the baseline-enriched profile is persisted.
    - Test 3: repeated triggers for the same examination do not double-apply baseline updates or duplicate snapshots.
  </behavior>
  <action>Extend the aggregation flow from plan 02 so it loads the bounded specialist history and current baseline state, builds the baseline request, calls the Python service, and persists both the examination-level baseline snapshot and the updated specialist baseline state inside PostgreSQL. After that transaction completes, transition the examination to `aggregated` and make `/processing-status` treat `aggregated` as terminal success. Preserve idempotency with the existing unique profile fence and baseline-state versioning, and surface transport or validation failures as controlled technical errors instead of partial success.</action>
  <verify>
    <automated>cd /home/vadim/diplom/core-backend && go test ./internal/aggregation ./internal/processing -run 'TestAggregationPersistsBaselineSnapshot|TestBaselineMetadataPersistedOnProfile|TestProcessingStatusShowsAggregatedAfterBaseline' -count=1</automated>
  </verify>
  <done>Core backend now stores baseline-aware aggregated results, updates specialist baseline state safely, and reports `aggregated` as the terminal success status.</done>
</task>

</tasks>

<verification>
Run the targeted aggregation, baseline-client, and processing tests and confirm terminal success is now bound to persisted baseline-enriched profiles, not just channel completion.
</verification>

<success_criteria>
- Core backend integrates with `ml-services/ml-baseline` through a narrow tested client.
- Baseline metadata and deviations are persisted in PostgreSQL under core ownership.
- `aggregated` is the Phase 3 terminal success state for processing-status and examination workflow.
</success_criteria>

<output>
After completion, create `.planning/phases/03-aggregated-baseline-aware-profiles/03-04-SUMMARY.md`
</output>
