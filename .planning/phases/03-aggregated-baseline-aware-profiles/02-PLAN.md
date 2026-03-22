---
phase: 03-aggregated-baseline-aware-profiles
plan: 02
type: execute
wave: 2
depends_on:
  - 03-01
files_modified:
  - core-backend/internal/aggregation/service.go
  - core-backend/internal/aggregation/repository.go
  - core-backend/internal/aggregation/service_test.go
  - core-backend/internal/channelresults/service.go
  - core-backend/internal/processing/service.go
  - core-backend/internal/processing/repository.go
autonomous: true
requirements:
  - AGGR-01
  - AGGR-02
  - AGGR-03
must_haves:
  truths:
    - Aggregation starts only after all mandatory channels have persisted `succeeded` results.
    - The examination moves into `aggregating` exactly once and cannot create duplicate aggregated profile rows.
    - The stored profile uses neutral normalized metrics, channel contributions, and deterministic explanation bullets derived from stub payloads.
  artifacts:
    - core-backend/internal/aggregation/service.go implements eligibility, normalization, and explanation rules.
    - core-backend/internal/aggregation/repository.go persists one profile per examination and provides aggregation fences.
    - core-backend/internal/channelresults/service.go triggers aggregation readiness checks from the existing result-ingestion flow.
  key_links:
    - Result ingestion must stay core-owned: successful channel results trigger aggregation readiness inside Go core, not another broker workflow.
    - `processing-status` must expose `aggregating` before baseline integration marks final success.
    - Aggregated metrics must be canonicalized from stub worker payloads without leaking raw field names into operator-facing storage.
---

<objective>
Implement the core-owned aggregation runner and persist the canonical pre-baseline profile when all mandatory channel results succeed.

Purpose: satisfy the aggregation half of the phase without introducing a premature standalone aggregator service or Phase 4 delivery logic.
Output: idempotent aggregation eligibility, deterministic profile assembly, and `aggregating` state transitions in core backend.
</objective>

<execution_context>
@/home/katya/.codex/get-shit-done/workflows/execute-plan.md
@/home/katya/.codex/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/03-aggregated-baseline-aware-profiles/03-RESEARCH.md
@docs/01_contract.md
@core-backend/internal/processing/contracts.go
@core-backend/internal/channelresults/service.go
@core-backend/internal/processing/repository.go
@.planning/phases/02-asynchronous-multichannel-processing/02-04-SUMMARY.md

<interfaces>
From core-backend/internal/channelresults/service.go:
```go
func (s *Service) ApplyResult(ctx context.Context, envelope processing.ChannelResultEnvelope) error
```

From core-backend/internal/processing/contracts.go:
```go
var MandatoryChannels = []string{
	ChannelText,
	ChannelAcoustic,
	ChannelParalinguistic,
}
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Build deterministic aggregation normalization and explanation rules</name>
  <files>core-backend/internal/aggregation/service.go, core-backend/internal/aggregation/repository.go, core-backend/internal/aggregation/service_test.go</files>
  <behavior>
    - Test 1: aggregation refuses to run until all mandatory channels have `succeeded`.
    - Test 2: the current stub payloads are converted into neutral canonical metrics such as `text_signal_score`, `acoustic_signal_score`, `paralinguistic_signal_score`, and `overall_index`.
    - Test 3: explanation bullets and channel contributions are deterministic for the same persisted input payloads.
  </behavior>
  <action>Create `internal/aggregation` as the Go-owned aggregation domain for Phase 3. Implement eligibility checks, profile normalization, channel contribution calculation, and human-readable explanation rules using deterministic mappings from the persisted stub channel payloads. Follow the research conclusion even though `docs/00_project.md` mentions a separate Aggregator component: for Phase 3, the logic lives in Go core because PostgreSQL already owns the workflow and persisted channel results. Keep metric names neutral and proxy-oriented; do not introduce clinical labels or KESMI-oriented recommendation logic.</action>
  <verify>
    <automated>cd /home/katya/dimplom/core-backend && go test ./internal/aggregation -run 'TestAggregationReadyOnlyAfterAllChannelsSucceeded|TestAggregationStoresVersionedProfile|TestAggregationIncludesContributionsAndExplanations' -count=1</automated>
  </verify>
  <done>`internal/aggregation` can decide readiness, build the canonical profile, and pass the RED tests for normalized metrics, contributions, and explanations.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Trigger aggregation from channel success and expose `aggregating` state</name>
  <files>core-backend/internal/channelresults/service.go, core-backend/internal/aggregation/repository.go, core-backend/internal/processing/service.go, core-backend/internal/processing/repository.go, core-backend/internal/aggregation/service_test.go</files>
  <behavior>
    - Test 1: when the last mandatory channel succeeds, the examination is fenced into `aggregating` only once.
    - Test 2: replayed success results do not create duplicate profile rows or repeated aggregation runs.
    - Test 3: `/processing-status` shows `aggregating` and remains non-terminal until baseline enrichment completes in a later plan.
  </behavior>
  <action>Extend the existing `channelresults` success path so that after saving a successful channel result it asks the aggregation repository whether all mandatory channels are complete and whether a profile already exists. If ready, acquire a DB fence, switch the examination to `aggregating`, and persist the canonical aggregated profile assembled by `internal/aggregation`. Update the processing-status projection so `aggregating` is backend-authoritative and not terminal yet. Do not call the baseline service in this plan; leave that integration for the dedicated baseline plan that follows.</action>
  <verify>
    <automated>cd /home/katya/dimplom/core-backend && go test ./internal/aggregation ./internal/processing -run 'TestAggregationReadyOnlyAfterAllChannelsSucceeded|TestProcessingStatusShowsAggregating' -count=1</automated>
  </verify>
  <done>The core backend now promotes fully successful examinations into `aggregating`, persists one canonical profile row, and reports the new status without duplicate aggregation side effects.</done>
</task>

</tasks>

<verification>
Run the targeted aggregation and processing tests and confirm successful-channel completion now advances into `aggregating` with one persisted canonical profile row.
</verification>

<success_criteria>
- Aggregation is triggered only when all mandatory channels have succeeded.
- One canonical profile row is stored per examination with contributions and explanations.
- Replayed or duplicated result envelopes do not duplicate aggregation output.
- `/processing-status` surfaces `aggregating` as a real backend state.
</success_criteria>

<output>
After completion, create `.planning/phases/03-aggregated-baseline-aware-profiles/03-02-SUMMARY.md`
</output>
