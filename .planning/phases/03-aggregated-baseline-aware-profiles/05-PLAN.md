---
phase: 03-aggregated-baseline-aware-profiles
plan: 05
type: execute
wave: 4
depends_on:
  - 03-04
files_modified:
  - core-backend/internal/results/service.go
  - core-backend/internal/results/repository.go
  - core-backend/internal/http/results_handler.go
  - core-backend/internal/http/results_handler_test.go
  - core-backend/internal/http/router.go
autonomous: true
requirements:
  - AGGR-03
  - RSLT-03
must_haves:
  truths:
    - Operator can fetch one examination’s aggregated profile, explanations, and baseline deviation through a stable HTTP DTO.
    - Operator can fetch specialist result history with trend/dynamics data built from persisted aggregated profiles, not raw channel payloads.
    - Backend result surfaces remain Phase 3-scoped and do not claim KESMI recommendation delivery yet.
  artifacts:
    - core-backend/internal/results/service.go assembles result and history DTOs from canonical persisted data.
    - core-backend/internal/http/results_handler.go exposes the new result endpoints.
    - core-backend/internal/http/results_handler_test.go proves the DTOs and route behavior.
  key_links:
    - `/examinations/{id}/result` must read the canonical profile plus baseline snapshot persisted in Phase 3, not recompute on request.
    - `/specialists/{id}/result-history` must derive dynamics from stored aggregated snapshots and timestamps.
    - Result DTOs must stay recommendation-free until Phase 4 KESMI delivery is implemented.
---

<objective>
Expose operator-facing result and specialist-history DTOs from the core backend.

Purpose: satisfy the backend half of `RSLT-03` and make the frontend consume one stable Phase 3 result surface.
Output: result/history services, HTTP handlers, and route tests for aggregated profile retrieval.
</objective>

<execution_context>
@/home/vadim/.codex/get-shit-done/workflows/execute-plan.md
@/home/vadim/.codex/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@docs/01_contract.md
@core-backend/internal/http/router.go
@core-backend/internal/http/examinations_handler.go
@core-backend/internal/aggregation/repository.go
@frontend/app/(app)/operator/examinations/[id]/results/page.tsx
@frontend/app/(app)/operator/specialists/[id]/page.tsx

<interfaces>
Existing operator routes in core-backend/internal/http/router.go:
```go
operatorOrAdmin.Get("/specialists/{id}/examinations", examinationsHandler.ListBySpecialist)
operatorOrAdmin.Get("/examinations/{id}", examinationsHandler.GetByID)
operatorOrAdmin.Get("/examinations/{id}/processing-status", examinationsHandler.ProcessingStatus)
```

Phase 3 adds:
```text
GET /examinations/{id}/result
GET /specialists/{id}/result-history
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Serve the canonical examination result DTO</name>
  <files>core-backend/internal/results/service.go, core-backend/internal/results/repository.go, core-backend/internal/http/results_handler.go, core-backend/internal/http/results_handler_test.go, core-backend/internal/http/router.go</files>
  <behavior>
    - Test 1: `GET /examinations/{id}/result` returns the persisted canonical profile, baseline snapshot, contributions, and explanation bullets.
    - Test 2: non-aggregated examinations are rejected with a contract-appropriate error instead of placeholder data.
    - Test 3: no KESMI recommendation fields are required yet in the response.
  </behavior>
  <action>Create a dedicated `internal/results` package that reads the canonical profile and baseline snapshot persisted in earlier plans and maps them into the DTO locked in `docs/01_contract.md`. Add a dedicated HTTP handler and route for `GET /examinations/{id}/result`. Return only Phase 3 data: normalized metrics, channel contributions, explanation bullets, baseline deviations, algorithm metadata, and timestamps. Do not pull in Phase 4 recommendation or KESMI delivery state.</action>
  <verify>
    <automated>cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/results -run 'TestGetExaminationResultEndpoint' -count=1</automated>
  </verify>
  <done>The backend exposes one stable examination result DTO backed by persisted aggregated data and rejects unavailable results cleanly.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Serve specialist result-history dynamics for operator trends</name>
  <files>core-backend/internal/results/service.go, core-backend/internal/results/repository.go, core-backend/internal/http/results_handler.go, core-backend/internal/http/results_handler_test.go, core-backend/internal/http/router.go</files>
  <behavior>
    - Test 1: `GET /specialists/{id}/result-history` returns ordered aggregated examinations with key indicator dynamics and baseline deltas.
    - Test 2: history excludes examinations that never reached `aggregated`.
    - Test 3: trend data is derived server-side from persisted aggregated snapshots rather than browser math over raw channel data.
  </behavior>
  <action>Extend `internal/results` with the specialist-history projection described in `docs/01_contract.md`. Query only aggregated examination profiles for the specialist, order them by examination completion time, and return trend items that include key indicator values, general/personal deviations, explanation summary, and profile metadata. Wire the new route in `router.go` for operator/admin access. Keep the DTO intentionally result-facing and interpretable; do not expose raw channel payload blobs or invent client-side trend formulas.</action>
  <verify>
    <automated>cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/results -run 'TestSpecialistResultHistoryEndpoint' -count=1</automated>
  </verify>
  <done>The backend now provides specialist result-history dynamics that the frontend can render directly for `RSLT-03`.</done>
</task>

</tasks>

<verification>
Run the targeted result/history endpoint tests and confirm they read persisted Phase 3 artifacts only and stay free of Phase 4 decision-delivery fields.
</verification>

<success_criteria>
- `/examinations/{id}/result` returns the canonical aggregated profile with baseline snapshot and explanations.
- `/specialists/{id}/result-history` returns ordered trend data for aggregated examinations only.
- Backend result endpoints are ready for frontend consumption without contract drift.
</success_criteria>

<output>
After completion, create `.planning/phases/03-aggregated-baseline-aware-profiles/03-05-SUMMARY.md`
</output>
