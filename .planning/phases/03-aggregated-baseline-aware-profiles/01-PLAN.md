---
phase: 03-aggregated-baseline-aware-profiles
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - docs/01_contract.md
  - core-backend/migrations/000007_aggregated_profiles.up.sql
  - core-backend/migrations/000007_aggregated_profiles.down.sql
  - core-backend/db/queries/aggregation.sql
  - core-backend/internal/aggregation/contracts.go
  - core-backend/internal/aggregation/service_test.go
  - core-backend/internal/http/results_handler_test.go
  - ml-services/ml-baseline/tests/test_service.py
  - ml-services/ml-baseline/tests/test_algorithms.py
autonomous: true
requirements:
  - AGGR-02
  - AGGR-03
  - BASE-01
  - BASE-02
  - RSLT-03
must_haves:
  truths:
    - Aggregated examination results have one canonical schema that does not expose raw stub-worker payload fields as the business contract.
    - Baseline computation uses a documented Python service boundary instead of direct PostgreSQL access from Python.
    - Result and history HTTP DTOs are specified before frontend or backend implementations begin.
  artifacts:
    - docs/01_contract.md documents Phase 3 HTTP, DB, and baseline-service contracts.
    - core-backend/migrations/000007_aggregated_profiles.up.sql creates authoritative tables for aggregated profiles, metric snapshots, and baseline state.
    - core-backend/internal/aggregation/service_test.go and ml-services/ml-baseline/tests/*.py pin the expected behaviors before implementation.
  key_links:
    - Channel success must unlock aggregation, but terminal success cannot be reported until baseline-enriched profile persistence completes.
    - Baseline request and response envelopes must remain versioned and narrow because the Python service is compute-only.
    - Operator result and history screens must consume canonical aggregated DTOs, not raw channel payloads.
---

<objective>
Publish the Phase 3 contracts, persistence skeleton, and RED verification scaffolds before any aggregation or baseline code lands.

Purpose: prevent contract drift across Go core, Python baseline, and frontend result surfaces.
Output: documented Phase 3 contracts, migration/query skeletons, and failing tests for aggregation, baseline, and result-history behavior.
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
@.planning/phases/03-aggregated-baseline-aware-profiles/03-VALIDATION.md
@docs/00_project.md
@docs/01_contract.md
@docs/02_implementation.md
@core-backend/internal/processing/contracts.go
@core-backend/internal/http/examinations_handler.go
@frontend/lib/api/types.ts

<interfaces>
From core-backend/internal/processing/contracts.go:
```go
type ProcessingStatusResponse struct {
	ExaminationID    int64
	Status           string
	MessageVersion   int
	ChannelsTotal    int
	ChannelsComplete int
	Terminal         bool
	StartedAt        *time.Time
	UpdatedAt        time.Time
	FinishedAt       *time.Time
	FailedAt         *time.Time
	Channels         []ChannelStatusDTO
}
```

From research constraints:
```text
aggregation stays in Go core backend for Phase 3
baseline stays a separate Python service boundary
real channel semantics remain deferred; use neutral proxy metric names
Phase 4 KESMI delivery must not be pulled into this phase
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Lock the Phase 3 contracts and status vocabulary</name>
  <files>docs/01_contract.md, core-backend/internal/aggregation/service_test.go, core-backend/internal/http/results_handler_test.go, ml-services/ml-baseline/tests/test_service.py, ml-services/ml-baseline/tests/test_algorithms.py</files>
  <behavior>
    - Test 1: aggregation result DTO returns one versioned canonical profile with normalized metrics, per-channel contributions, explanation bullets, and baseline snapshot metadata.
    - Test 2: `/processing-status` treats `aggregating` as in-progress and `aggregated` as terminal success, while `failed` remains terminal failure.
    - Test 3: baseline service response includes general deviation, personal deviation, algorithm version, refresh timestamp, and update-eligibility data.
    - Test 4: specialist result-history DTO exposes dynamics of key indicators across prior aggregated examinations.
  </behavior>
  <action>Update `docs/01_contract.md` first. Add the canonical aggregated profile schema, baseline request/response envelopes, `GET /examinations/{id}/result`, and `GET /specialists/{id}/result-history`. Expand the Phase 3 examination status vocabulary to `created`, `collecting_answers`, `ready_for_processing`, `processing`, `aggregating`, `aggregated`, and `failed`, and update `GET /examinations/{id}/processing-status` so `terminal=true` means `aggregated` or `failed`, not merely “all channels succeeded.” Keep Phase 4 concerns out: do not add KESMI request/response contracts or final recommendation fields beyond neutral placeholders reserved for later phases. In the same task, add RED tests in Go and Python that pin these contracts without pretending the current stub payloads have final clinical semantics.</action>
  <verify>
    <automated>cd /home/katya/dimplom/core-backend && go test ./internal/aggregation ./internal/http -run 'TestAggregationStoresVersionedProfile|TestProcessingStatusShowsAggregating|TestSpecialistResultHistoryEndpoint' -count=1 && cd /home/katya/dimplom/ml-services/ml-baseline && pytest -q tests/test_service.py::test_returns_general_and_personal_deviation tests/test_algorithms.py::test_outlier_freezes_baseline_update</automated>
  </verify>
  <done>`docs/01_contract.md` becomes the single source of truth for Phase 3 result, history, status, and baseline-service contracts, and the new tests fail only because implementation is not present yet.</done>
</task>

<task type="auto">
  <name>Task 2: Add the authoritative PostgreSQL schema and query skeletons</name>
  <files>core-backend/migrations/000007_aggregated_profiles.up.sql, core-backend/migrations/000007_aggregated_profiles.down.sql, core-backend/db/queries/aggregation.sql, core-backend/internal/aggregation/contracts.go</files>
  <action>Create the Phase 3 migration for one authoritative aggregated profile per examination plus the supporting data needed for baseline and history. Include tables for the canonical profile row, per-metric snapshots, per-channel contributions, explanation lines, specialist baseline state, and examination-level baseline snapshots. Enforce `UNIQUE (examination_id)` on the aggregated profile table and store `schema_version`, `aggregation_version`, `baseline_algorithm_version`, refresh timestamps, and counted examinations as first-class fields. Add query skeletons for eligibility checks, profile persistence, baseline-state reads/writes, and result-history projections. Create `core-backend/internal/aggregation/contracts.go` to export the stable metric keys, profile structs, and baseline payload structs that later plans implement against.</action>
  <verify>
    <automated>cd /home/katya/dimplom/core-backend && go test ./internal/aggregation ./internal/http -run 'TestAggregationStoresVersionedProfile|TestSpecialistResultHistoryEndpoint' -count=1</automated>
  </verify>
  <done>The repo has a durable schema and Go contract types for aggregated profiles, baseline snapshots, and result-history projections without revisiting the contract later.</done>
</task>

</tasks>

<verification>
Run the targeted Go and Python tests and confirm they now fail for missing implementation rather than missing contracts, table definitions, or route expectations.
</verification>

<success_criteria>
- `docs/01_contract.md` documents the canonical profile, baseline service, result/history endpoints, and additive status expansion for Phase 3.
- PostgreSQL schema skeleton exists for aggregated profiles, contributions, baseline state, and history snapshots.
- RED tests exist for aggregation, baseline update gating, result endpoint shape, and specialist trend history.
</success_criteria>

<output>
After completion, create `.planning/phases/03-aggregated-baseline-aware-profiles/03-01-SUMMARY.md`
</output>
