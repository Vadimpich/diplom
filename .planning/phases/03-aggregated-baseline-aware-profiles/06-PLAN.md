---
phase: 03-aggregated-baseline-aware-profiles
plan: 06
type: execute
wave: 5
depends_on:
  - 03-05
files_modified:
  - frontend/lib/api/types.ts
  - frontend/lib/api/client.ts
  - frontend/components/operator/status-badge.tsx
  - frontend/app/(app)/operator/specialists/[id]/page.tsx
  - frontend/app/(app)/operator/examinations/[id]/results/page.tsx
  - frontend/app/(app)/operator/history/page.tsx
  - README.md
  - .planning/phases/03-aggregated-baseline-aware-profiles/03-VALIDATION.md
  - docs/02_implementation.md
autonomous: true
requirements:
  - AGGR-03
  - RSLT-03
must_haves:
  truths:
    - Operator result screen shows the real Phase 3 aggregated profile, baseline deviation, and explanation instead of the placeholder shell.
    - Specialist detail and history surfaces let the operator inspect dynamics across aggregated examinations against baseline.
    - Phase 3 verification and runbook docs reflect the shipped result/baseline flow, and implementation history is appended once execution is complete.
  artifacts:
    - frontend/lib/api/types.ts and frontend/lib/api/client.ts define the result/history DTOs and `aggregating`/`aggregated` statuses.
    - frontend/app/(app)/operator/examinations/[id]/results/page.tsx renders the real baseline-aware result view.
    - frontend/app/(app)/operator/specialists/[id]/page.tsx and frontend/app/(app)/operator/history/page.tsx expose trend navigation and status handling.
    - docs/02_implementation.md gains one append-only Phase 3 entry after verification passes.
  key_links:
    - Frontend must render backend snake_case DTOs directly to avoid contract drift.
    - `aggregating` still routes to processing progress while `aggregated` routes to the result screen.
    - Validation and README commands must include the new baseline service and result endpoints.
---

<objective>
Connect the operator UI to the new Phase 3 result surfaces and close out validation/runbook documentation.

Purpose: deliver the actual operator-facing baseline-aware result experience and leave the phase in a verifiable, documented state.
Output: typed frontend integration, result/history UI, refreshed validation docs, and append-only implementation log entry.
</objective>

<execution_context>
@/home/katya/.codex/get-shit-done/workflows/execute-plan.md
@/home/katya/.codex/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/03-aggregated-baseline-aware-profiles/03-VALIDATION.md
@docs/01_contract.md
@docs/02_implementation.md
@frontend/lib/api/types.ts
@frontend/lib/api/client.ts
@frontend/app/(app)/operator/examinations/[id]/results/page.tsx
@frontend/app/(app)/operator/specialists/[id]/page.tsx
@frontend/components/operator/status-badge.tsx
@README.md

<interfaces>
Current frontend status union in frontend/lib/api/types.ts:
```ts
export type ExaminationStatus =
  | "created"
  | "collecting_answers"
  | "ready_for_processing"
  | "processing"
  | "failed";
```

Phase 3 expands this with:
```text
aggregating
aggregated
result and specialist-history DTOs from the backend
```
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Add typed Phase 3 result contracts to the frontend boundary</name>
  <files>frontend/lib/api/types.ts, frontend/lib/api/client.ts, frontend/components/operator/status-badge.tsx</files>
  <action>Extend the shared frontend API types with the Phase 3 result DTOs, specialist result-history DTOs, and the additive `aggregating` / `aggregated` status values from `docs/01_contract.md`. Add typed client methods for `GET /examinations/{id}/result` and `GET /specialists/{id}/result-history`. Update the shared status badge so all contract-visible examination statuses render exhaustively and route logic can distinguish `aggregating` from terminal `aggregated`. Keep backend snake_case field names intact and do not add frontend-only enum aliases.</action>
  <verify>
    <automated>cd /home/katya/dimplom/frontend && npm run lint && npm run build && npx tsc --noEmit</automated>
  </verify>
  <done>The frontend boundary knows about the Phase 3 result/history DTOs and all current backend workflow statuses.</done>
</task>

<task type="auto">
  <name>Task 2: Replace placeholders with real result and trend views</name>
  <files>frontend/app/(app)/operator/examinations/[id]/results/page.tsx, frontend/app/(app)/operator/specialists/[id]/page.tsx, frontend/app/(app)/operator/history/page.tsx</files>
  <action>Replace the placeholder results page with a real TanStack Query view that renders the aggregated profile summary, baseline deviations, channel contributions, and explanation bullets from `GET /examinations/{id}/result`. Extend the specialist detail page and operator history surface to fetch `GET /specialists/{id}/result-history` and show dynamics of key indicators across aggregated examinations. Keep navigation contract-driven: `aggregating` examinations still lead to the processing screen, while `aggregated` examinations lead to the results screen. Do not compute baseline or trend math in the browser beyond simple presentation formatting.</action>
  <verify>
    <automated>cd /home/katya/dimplom/frontend && npm run lint && npm run build && npx tsc --noEmit</automated>
  </verify>
  <done>The operator UI now shows the real Phase 3 result and history dynamics rather than shells, and navigation respects backend statuses.</done>
</task>

<task type="auto">
  <name>Task 3: Close out Phase 3 validation, runbook, and implementation history</name>
  <files>README.md, .planning/phases/03-aggregated-baseline-aware-profiles/03-VALIDATION.md, docs/02_implementation.md</files>
  <action>After the Phase 3 automated checks pass, update `README.md` with the local run steps and verification commands that include the baseline service and result endpoints. Refresh `03-VALIDATION.md` so the per-task and phase-level verification commands match the shipped files and route names. Append one concise Phase 3 implementation entry to `docs/02_implementation.md` describing the delivered aggregation, baseline, result, and trend capabilities. Keep `docs/02_implementation.md` append-only and do not rewrite prior history.</action>
  <verify>
    <automated>cd /home/katya/dimplom/frontend && npm run lint && npm run build && npx tsc --noEmit && cd /home/katya/dimplom/core-backend && go test ./... -count=1 && cd /home/katya/dimplom/ml-baseline && pytest -q && cd /home/katya/dimplom && rg -n 'aggregated|baseline|result-history|/examinations/\\{id\\}/result' README.md .planning/phases/03-aggregated-baseline-aware-profiles/03-VALIDATION.md docs/02_implementation.md</automated>
  </verify>
  <done>Phase 3 has current runbook and validation docs, and the implementation log includes one new append-only entry for the shipped result flow.</done>
</task>

</tasks>

<verification>
Run the frontend checks plus the full backend and baseline test suites, then confirm the docs mention the actual Phase 3 routes, statuses, and local runtime.
</verification>

<success_criteria>
- Frontend renders the real aggregated result and specialist trend history.
- Shared status handling supports `aggregating` and `aggregated` without contract drift.
- README, `03-VALIDATION.md`, and `docs/02_implementation.md` reflect the shipped Phase 3 flow.
</success_criteria>

<output>
After completion, create `.planning/phases/03-aggregated-baseline-aware-profiles/03-06-SUMMARY.md`
</output>
