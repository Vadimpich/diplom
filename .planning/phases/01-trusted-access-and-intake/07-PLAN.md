---
phase: 01-trusted-access-and-intake
plan: 07
title: Specialist History Status UI
type: execute
wave: 6
depends_on:
  - 08-PLAN.md
  - 06-PLAN.md
files_modified:
  - /home/vadim/diplom/core-backend/internal/http/examinations_handler_test.go
  - /home/vadim/diplom/frontend/app/(app)/operator/specialists/[id]/page.tsx
  - /home/vadim/diplom/frontend/components/operator/examination-summary.tsx
  - /home/vadim/diplom/frontend/components/operator/status-badge.tsx
  - /home/vadim/diplom/frontend/lib/api/client.ts
  - /home/vadim/diplom/frontend/lib/api/types.ts
  - /home/vadim/diplom/docs/01_contract.md
  - /home/vadim/diplom/docs/02_implementation.md
autonomous: true
requirements_addressed:
  - EXAM-04
must_haves:
  truths:
    - "Specialist history reads current workflow statuses from GET /specialists/{id}/examinations."
    - "Operator UI renders authoritative backend statuses rather than deriving them from local draft state."
  artifacts:
    - path: /home/vadim/diplom/core-backend/internal/http/examinations_handler_test.go
      provides: backend proof that specialist history stays authoritative
    - path: /home/vadim/diplom/frontend/app/(app)/operator/specialists/[id]/page.tsx
      provides: specialist history page wired to backend statuses
    - path: /home/vadim/diplom/frontend/components/operator/status-badge.tsx
      provides: authoritative workflow status rendering
  key_links:
    - from: /home/vadim/diplom/frontend/app/(app)/operator/specialists/[id]/page.tsx
      to: /home/vadim/diplom/frontend/lib/api/client.ts
      via: GET /specialists/{id}/examinations fetch
    - from: /home/vadim/diplom/core-backend/internal/http/examinations_handler_test.go
      to: /home/vadim/diplom/core-backend/internal/http/examinations_handler.go
      via: authoritative history endpoint verification
---

# Objective

Render specialist examination history from the authoritative backend status API.

<tasks>

<task id="1-07-01" type="auto">
  <name>Task 1: Wire specialist history page to authoritative examination statuses</name>
  <files>
    /home/vadim/diplom/core-backend/internal/http/examinations_handler_test.go
    /home/vadim/diplom/frontend/app/(app)/operator/specialists/[id]/page.tsx
    /home/vadim/diplom/frontend/components/operator/examination-summary.tsx
    /home/vadim/diplom/frontend/components/operator/status-badge.tsx
    /home/vadim/diplom/frontend/lib/api/client.ts
    /home/vadim/diplom/frontend/lib/api/types.ts
  </files>
  <read_first>
    /home/vadim/diplom/AGENTS.md
    /home/vadim/diplom/core-backend/internal/http/examinations_handler_test.go
    /home/vadim/diplom/frontend/components/operator/examination-summary.tsx
    /home/vadim/diplom/frontend/components/operator/status-badge.tsx
    /home/vadim/diplom/frontend/lib/api/client.ts
  </read_first>
  <action>
    Extend `core-backend/internal/http/examinations_handler_test.go` with an explicit backend verification that `GET /specialists/{id}/examinations` returns current PostgreSQL-backed workflow statuses after the snapshot, answer-linkage, and finish-fence changes. Implement or update `frontend/app/(app)/operator/specialists/[id]/page.tsx` so the specialist card/history view fetches `GET /specialists/{id}/examinations` through `frontend/lib/api/client.ts` and renders backend `status`, `started_at`, and `finished_at` values. Update `frontend/components/operator/examination-summary.tsx` and `frontend/components/operator/status-badge.tsx` so they display `created`, `collecting_answers`, and `ready_for_processing` from the API without deriving workflow state from local drafts, cookies, or optimistic client flags. Extend `frontend/lib/api/types.ts` to keep the response contract aligned with the backend.
  </action>
  <verify>
    <automated>cd /home/vadim/diplom/core-backend && go test ./internal/http -run TestListBySpecialistReturnsCurrentStatuses -count=1 && cd /home/vadim/diplom/frontend && npm run lint && npx tsc --noEmit</automated>
  </verify>
  <acceptance_criteria>
    `rg -n "specialists/\\[id\\]/page.tsx|ready_for_processing|collecting_answers|created" /home/vadim/diplom/frontend`
  </acceptance_criteria>
  <done>
    Backend test proves `GET /specialists/{id}/examinations` remains authoritative after Phase 1 changes, specialist history UI is wired to that endpoint, and status components render authoritative workflow states.
  </done>
</task>

<task id="1-07-02" type="auto">
  <name>Task 2: Document specialist-history status semantics</name>
  <files>
    /home/vadim/diplom/docs/01_contract.md
    /home/vadim/diplom/docs/02_implementation.md
  </files>
  <read_first>
    /home/vadim/diplom/AGENTS.md
    /home/vadim/diplom/docs/01_contract.md
    /home/vadim/diplom/docs/02_implementation.md
  </read_first>
  <action>
    Update `docs/01_contract.md` with status-history expectations for `GET /specialists/{id}/examinations` and append the implementation note to `docs/02_implementation.md`. Treat the documentation duties required by `AGENTS.md` as part of completion.
  </action>
  <verify>
    <automated>cd /home/vadim/diplom/frontend && npm run lint && npx tsc --noEmit</automated>
  </verify>
  <acceptance_criteria>
    `rg -n "GET /specialists/\\{id\\}/examinations|ready_for_processing|collecting_answers|created" /home/vadim/diplom/docs/01_contract.md`
  </acceptance_criteria>
  <done>
    Specialist-history status semantics are recorded in `docs/01_contract.md` and `docs/02_implementation.md` per `AGENTS.md`.
  </done>
</task>

</tasks>
