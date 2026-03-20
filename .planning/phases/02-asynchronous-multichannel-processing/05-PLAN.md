---
phase: 02-asynchronous-multichannel-processing
plan: 05
type: execute
wave: 4
depends_on:
  - 02-04
files_modified:
  - docs/01_contract.md
  - frontend/lib/api/types.ts
  - frontend/lib/api/client.ts
  - frontend/app/(app)/operator/examinations/[id]/processing/page.tsx
  - frontend/app/(app)/operator/history/page.tsx
  - frontend/components/operator/status-badge.tsx
autonomous: true
requirements:
  - RSLT-01
must_haves:
  truths:
    - Operator can see per-channel progress for text, acoustic, and paralinguistic channels.
    - UI polling stops automatically when the backend reports a terminal state.
    - Frontend renders backend statuses directly instead of inventing local workflow.
  artifacts:
    - frontend/lib/api/types.ts defines the processing-status DTO and expands generic examination statuses for `processing` / `failed`.
    - frontend/lib/api/client.ts adds one typed API method for the progress endpoint.
    - frontend/app/(app)/operator/examinations/[id]/processing/page.tsx renders the real channel progress.
    - Existing status consumers keep working when generic examination payloads include `processing` / `failed`.
  key_links:
    - TanStack Query polls GET /examinations/{id}/processing-status until is_terminal is true.
    - Status rendering maps backend status values without client-side derivation.
---

<objective>
Replace the placeholder processing screen with a minimal contract-driven progress view.

Purpose: satisfy `RSLT-01` without introducing speculative UI patterns or backend logic into the frontend.
Output: typed API client support and a polling processing page that renders backend-authoritative channel status.
</objective>

<execution_context>
@/home/katya/.codex/get-shit-done/workflows/execute-plan.md
@/home/katya/.codex/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/02-asynchronous-multichannel-processing/02-RESEARCH.md
@docs/01_contract.md
@frontend/lib/api/types.ts
@frontend/lib/api/client.ts
@frontend/app/(app)/operator/examinations/[id]/processing/page.tsx
@frontend/components/operator/status-badge.tsx

<interfaces>
From frontend/lib/api/client.ts:
```ts
getExamination(id: number) {
  return request<Examination>(`/examinations/${id}`);
}
```

Implement the same pattern for:
```ts
getExaminationProcessingStatus(id: number)
```
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Add typed frontend contract support for processing progress and expanded examination statuses</name>
  <files>docs/01_contract.md, frontend/lib/api/types.ts, frontend/lib/api/client.ts</files>
  <action>Extend `frontend/lib/api/types.ts` with the dedicated processing-status DTO from `docs/01_contract.md`, including overall pipeline fields and the per-channel entries. In the same task, align the generic `Examination.status` union with the Phase 2 contract decision so `processing` and `failed` are valid typed values for existing list/detail consumers. Add one typed client method in `frontend/lib/api/client.ts` for `GET /examinations/{id}/processing-status`. Keep snake_case fields aligned with backend JSON and avoid introducing frontend-only enum aliases that could drift from the contract.</action>
  <verify>
    <automated>cd /home/katya/dimplom/frontend && npm run lint && npx tsc --noEmit</automated>
  </verify>
  <done>Frontend can request the processing-status endpoint through the shared typed API boundary, and existing typed examination consumers accept `processing` / `failed` without contract drift.</done>
</task>

<task type="auto">
  <name>Task 2: Render live channel progress and keep shared status UI compatible</name>
  <files>frontend/app/(app)/operator/examinations/[id]/processing/page.tsx, frontend/app/(app)/operator/history/page.tsx, frontend/components/operator/status-badge.tsx</files>
  <action>Replace the placeholder processing page with TanStack Query polling against the new progress endpoint. Render overall pipeline status plus one card/row each for `text`, `acoustic`, and `paralinguistic`, including attempt counters and last error text when present. Poll every few seconds until `is_terminal` is true, then stop. Reuse and extend the existing badge/history rendering so any generic examination payloads carrying `processing` or `failed` still render correctly in operator history and related shared status UI. Do not derive optimistic progress in the browser.</action>
  <verify>
    <automated>cd /home/katya/dimplom/frontend && npm run lint && npx tsc --noEmit && npm run build</automated>
  </verify>
  <done>The operator processing page shows real per-channel progress from backend polling and stops polling on terminal completion or error, while shared status UI remains compatible with Phase 2 examination statuses.</done>
</task>

</tasks>

<verification>
Run the frontend static checks and confirm the page builds against the backend contract without placeholder-only text.
</verification>

<success_criteria>
- Frontend has one typed method for the processing-status endpoint and accepts `processing` / `failed` in generic examination DTOs.
- The processing page renders all mandatory channels and their backend statuses.
- Polling is contract-driven and halts on terminal state, and existing history/status surfaces remain compatible.
</success_criteria>

<output>
After completion, create `.planning/phases/02-asynchronous-multichannel-processing/02-05-SUMMARY.md`
</output>
