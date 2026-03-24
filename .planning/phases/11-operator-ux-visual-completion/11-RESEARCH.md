# Phase 11: Operator UX & Visual Completion - Research

**Researched:** 2026-03-24
**Domain:** Next.js operator/admin UX completion on existing brownfield frontend
**Confidence:** HIGH

## User Constraints

- Milestone remains frontend-first.
- Only note backend changes if strictly necessary.
- Keep recommendations grounded in the current repo, not a speculative redesign.
- Propose a clean plan decomposition for Phase 11 with likely 3-5 executable plans.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| OPRX-01 | Оператор видит results screen с понятной интерпретацией результата, baseline-отклонениями, вкладом каналов и отдельно вынесенными техническими деталями | Repo audit identifies result-page information-architecture debt, raw contract wording, and a frontend-only restructuring path using existing DTOs and card primitives |
| OPRX-02 | Сценарий обследования даёт оператору явный feedback для записи, загрузки, сохранения, завершения и recoverable ошибок | Repo audit identifies missing success/confirmation feedback in examination start/upload/finish and specialist edit/delete flows; existing `sonner`, `Alert`, `ConfirmDialog`, and button pending states are sufficient |
| OPRX-03 | Ключевые operator и admin страницы корректно обрабатывают `loading`, `empty` и `error` состояния без blank или placeholder UI | Repo audit shows remaining gaps on operator history, specialists, examination create, results, and some admin/read surfaces despite Phase 9-10 improvements |
| OPRX-04 | Ключевые edit и destructive действия в operator/admin интерфейсах требуют подтверждения там, где это необходимо, и показывают явный success/error feedback после выполнения | Existing shared confirm/toast primitives are available; gaps remain mainly in operator specialist edit/delete and examination completion flow |
| DSGN-02 | В shipped frontend не остаются временные заглушки, черновые элементы и визуально незавершённые экраны в основном пользовательском потоке | Repo audit shows lingering contract/debug copy, “shell/stub” phrasing, and unfinished explanatory blocks on operator/admin primary pages |
</phase_requirements>

## Summary

Phase 11 should be executed as a brownfield polish phase on top of the shipped Next.js 15 frontend, not as a redesign. The repo already has the right primitives: contour shells, shared `PageHeader`, `Alert`, `EmptyState`, `Skeleton`, app-level `sonner` toaster, and a reusable `ConfirmDialog`. The remaining work is mostly composition, copy cleanup, and state coverage on operator-first screens.

The largest debt is on operator surfaces that still speak in contract/debug language instead of operator language. The worst offender is the examination results screen, but the same problem appears on processing, history, specialist detail/history, operator dashboard, and several admin descriptions that still explain milestone internals instead of the user task. This is a frontend information-architecture issue, not a backend contract issue.

The cleanest Phase 11 decomposition is four plans: results information architecture, examination-flow feedback and confirmations, remaining loading/empty/error coverage, and final unfinished-surface/copy cleanup with verification. No backend change is strictly required for the current goal.

**Primary recommendation:** Keep the current stack and contracts; use Phase 11 to normalize copy, state handling, and feedback across shipped operator/admin flows with four sequential frontend plans.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Next.js App Router | 15.5.14 in repo; latest registry 16.2.1 verified 2026-03-20 | Route layouts, `loading.tsx`, client/server page structure | Already shipped in repo; Phase 11 is polish work, not framework migration |
| React | 19.2.4 | UI runtime | Current repo baseline; no need to change for UX completion |
| Tailwind CSS | 3.4.19 in repo; latest registry 4.2.2 verified 2026-03-18 | Tokens and utility styling | Existing design-token layer already built in Phase 8 |
| TanStack Query | 5.91.2 in repo; latest registry 5.95.2 verified 2026-03-23 | Query state for loading/error/refetch | Already powers all list/detail pages and should remain the state backbone |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| React Hook Form | 7.71.2 in repo; latest registry 7.72.0 verified 2026-03-22 | Form state | Keep for create/edit pages and settings |
| Zod | 3.25.76 in repo; latest registry 4.3.6 verified 2026-01-22 | Validation | Keep existing schemas; do not upgrade in this phase |
| Radix Alert Dialog | 1.1.15 | Confirmation modal primitive | Reuse for destructive or irreversible actions |
| Sonner | 2.0.7 | Toast/success feedback | Reuse for mutation success and recoverable action feedback |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Existing Next 15/Tailwind 3 stack | Upgrade to Next 16/Tailwind 4 during Phase 11 | Wrong phase boundary; introduces migration risk unrelated to UX debt |
| Existing shared primitives | Page-local custom modals/toasts/states | Reintroduces UI drift that Phase 8 explicitly removed |
| Frontend-only copy/state cleanup | New backend DTOs for operator wording | Unnecessary for current gaps; existing DTOs already contain enough data |

**Installation:**
```bash
npm install
```

**Version verification:** Verified on 2026-03-24 with:
```bash
npm view next version
npm view react version
npm view @tanstack/react-query version
npm view react-hook-form version
npm view zod version
npm view @radix-ui/react-alert-dialog version
npm view sonner version
npm view tailwindcss version
```

## Project Constraints

- `docs/00_project.md` remains the product and architecture source of truth.
- If any HTTP/API contract changes are introduced, `docs/01_contract.md` must be updated in the same phase.
- After execution, `docs/02_implementation.md` must receive an append-only implementation note.
- Keep changes minimal; do not refactor unrelated areas.
- Respect operator/admin contour separation.
- Do not add backend work unless the UI goal cannot be met with existing contracts.
- Do not surface secrets, raw stack traces, or sensitive data.

## Architecture Patterns

### Recommended Project Structure
```text
frontend/
├── app/(app)/operator/          # Operator routes and route-level loading/error UI
├── app/(app)/admin/             # Admin routes and route-level loading/error UI
├── components/ui/               # Shared primitives: alert, empty, confirm, skeleton, toaster
├── components/operator/         # Operator-specific reusable workflow blocks
└── components/admin/            # Admin-specific reusable list/form blocks
```

### Pattern 1: Results Page Uses Two Layers Of Detail
**What:** Put operator interpretation first; move correlation/debug/runtime fields into a secondary “technical details” block.
**When to use:** Examination result and processing screens.
**Example:**
```tsx
<Card>
  <CardHeader>
    <CardTitle>Итог обследования</CardTitle>
    <CardDescription>{operatorSummary}</CardDescription>
  </CardHeader>
  <CardContent>{/* score, baseline, explanations, contributions */}</CardContent>
</Card>

<Card>
  <CardHeader>
    <CardTitle>Технические детали</CardTitle>
    <CardDescription>Для диагностики и повторной проверки результата.</CardDescription>
  </CardHeader>
  <CardContent>{/* correlation_id, attempts, diagnostics */}</CardContent>
</Card>
```

### Pattern 2: Query-Driven State Branching At The List Boundary
**What:** Branch `loading`, `error`, `empty`, `filtered-empty`, and `populated` before rendering list rows.
**When to use:** Operator history, specialists, examination creation pickers, monitoring read surfaces.
**Example:**
```tsx
if (query.isLoading) return <SkeletonList />;
if (query.isError) return <Alert variant="danger">Не удалось загрузить данные.</Alert>;
if (!items.length && filterApplied) return <EmptyState title="Ничего не найдено" description="Измените фильтр." />;
if (!items.length) return <EmptyState title="Список пуст" description="Добавьте первую запись." />;
return <ActualList items={items} />;
```

### Pattern 3: Mutation Feedback Uses Pending + Success/Error + Confirm
**What:** Pending button state for all mutations, success toast or inline success alert after completion, confirm dialog for destructive or irreversible actions.
**When to use:** Specialist delete, specialist edit, examination finish, answer save, questionnaire question removal.
**Example:**
```tsx
<ConfirmDialog
  title="Завершить сбор ответов?"
  description="После подтверждения обследование перейдёт в обработку."
  confirmLabel="Завершить"
  confirmVariant="danger"
  onConfirm={() => finishMutation.mutate() }
  trigger={<Button>Завершить сбор ответов</Button>}
/>
```

### Anti-Patterns to Avoid
- **Contract-first copy in operator UI:** Do not expose `DTO`, `backend-authoritative`, `decision layer`, raw endpoint names, or milestone references in user descriptions.
- **One-card mixed abstraction:** Do not keep operator summary and diagnostics in one equally weighted block on results/process screens.
- **Blank-card loading:** Do not render an empty container while queries are unresolved.
- **Silent success:** Do not redirect after create/update/delete without explicit user feedback unless the destination page itself clearly acknowledges success.
- **New local feedback widgets:** Do not invent page-specific confirmation/toast components; reuse Phase 8 primitives.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Confirm modals | New page-local modal implementation | `components/ui/confirm-dialog.tsx` | Already integrated with Radix semantics and design tokens |
| Toasts | Ad hoc inline banners everywhere | App-level Sonner toaster | Keeps mutation feedback consistent across contours |
| Loading placeholders | Per-page custom gray boxes | Shared `Skeleton` plus route/page-level branching | Faster and visually consistent |
| Empty-state cards | Repeated dashed cards with local markup | `EmptyState` extended only if needed | Keeps state language and layout consistent |
| Result remapping | Backend DTO changes for wording only | Frontend presenter/copy layer on existing DTO | Cheaper, lower risk, no contract churn |

**Key insight:** Phase 11 should mostly compose the primitives that Phase 8-10 already shipped; custom one-off UX solutions would create new debt while trying to remove old debt.

## Common Pitfalls

### Pitfall 1: “Fixing” UX With Backend Expansion
**What goes wrong:** Planner introduces new endpoints or DTO fields for issues that are purely copy/layout/state problems.
**Why it happens:** Raw contract wording on results feels like a contract problem, but the repo already has enough data to present it better.
**How to avoid:** Treat result-processing-specialist wording as presenter-layer work first. Only escalate to backend if a user-facing label truly cannot be derived from existing DTOs.
**Warning signs:** Proposed tasks mention new result fields, translator endpoints, or DTO redesign before any UI restructuring.

### Pitfall 2: Inconsistent State Taxonomy
**What goes wrong:** Some pages get `loading` + `error`, others only `empty`, and filtered-empty states collapse into generic warnings.
**Why it happens:** Lists are still implemented page by page.
**How to avoid:** Define one explicit state matrix per page: initial loading, fetch error, empty dataset, filtered empty, populated.
**Warning signs:** Components still map `(query.data?.items ?? [])` directly or show only one warning alert for all zero-item cases.

### Pitfall 3: Silent Operator Mutations
**What goes wrong:** Start/save/finish/delete actions happen, but the operator gets no success acknowledgment or no confirmation before irreversible actions.
**Why it happens:** Redirects and query invalidation are used as implicit feedback.
**How to avoid:** Add toast/inline success states and confirm destructive transitions.
**Warning signs:** `onSuccess` only invalidates query or redirects, with no feedback side effect.

### Pitfall 4: Debug Copy Survives Final Polish
**What goes wrong:** Pages still mention phases, shells, DTOs, raw statuses, or backend ownership after “visual completion”.
**Why it happens:** Technical descriptions were useful during earlier milestone delivery and were never replaced.
**How to avoid:** Rewrite all operator/admin descriptions around user intent; keep technical language only in dedicated technical sections.
**Warning signs:** Page headers contain endpoint names, phase numbers, or terms like `backend-authoritative`, `shell`, `placeholder`, `DTO`.

## Code Examples

Verified patterns from official sources and current repo:

### Route-Level Loading For App Router
```tsx
// app/(app)/operator/loading.tsx
export default function Loading() {
  return <OperatorSkeleton />;
}
```
Source: https://nextjs.org/docs/app/building-your-application/routing/loading-ui-and-streaming

### Query State Branching
```tsx
const query = useQuery({ queryKey: ["items"], queryFn: fetchItems });

if (query.isLoading) return <SkeletonList />;
if (query.isError) return <Alert variant="danger">Ошибка загрузки.</Alert>;
if (!query.data?.items.length) return <EmptyState title="Пока пусто" description="Добавьте первую запись." />;
return <List items={query.data.items} />;
```
Source: https://tanstack.com/query/latest/docs/framework/react/reference/useQuery

### Accessible Confirmation Dialog
```tsx
<ConfirmDialog
  title="Удалить карточку?"
  description="Действие нельзя отменить."
  confirmLabel="Удалить"
  confirmVariant="danger"
  onConfirm={handleDelete}
  trigger={<Button variant="danger">Удалить</Button>}
/>
```
Source: https://www.radix-ui.com/primitives/docs/components/alert-dialog

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Technical result/debug copy shown inline to operator | Operator-first summary with separate technical details block | Needed now in Phase 11 | Increases trust and readability without contract changes |
| Page-level direct array mapping | Explicit state boundaries around query lifecycle | Already used on admin users/questionnaires in Phase 9 | Phase 11 should extend this to remaining operator/admin pages |
| Implicit mutation success via redirect | Explicit success feedback via toast/alert plus redirect | Already used on admin users/questionnaires/settings | Phase 11 should bring operator flows to the same standard |

**Deprecated/outdated:**
- Phase and endpoint names in `PageHeader` descriptions: outdated for shipped UI; keep them in code comments or docs, not operator/admin-visible copy.
- “Shell”, “placeholder”, “backend-authoritative”, “DTO”, “decision layer” in user-visible descriptions: outdated for a completion phase.

## Recommended Plan Breakdown

### Plan 11-01: Operator Results Information Architecture
- Rewrite `/operator/examinations/[id]/results` around operator language.
- Split interpretation, baseline deviations, explanations, channel contributions, and technical diagnostics into clear sections.
- Normalize result-history copy on `/operator/specialists/[id]` and remove raw English labels.
- Preserve existing backend DTOs; no contract change unless a true blocker is discovered.

### Plan 11-02: Examination Flow Feedback And Confirmations
- Add explicit success feedback for start, answer upload/save, specialist create/update, and examination completion.
- Add confirmation for destructive or irreversible actions: specialist delete and examination finish at minimum.
- Tighten recoverable-error messaging around recording/microphone/upload failures.
- Reuse `ConfirmDialog`, `Alert`, pending buttons, and `sonner`; do not introduce new primitives.

### Plan 11-03: Remaining Loading, Empty, Error Coverage
- Close gaps on operator history, specialists, examination creation pickers, result loading fallback, and remaining admin read surfaces.
- Distinguish empty dataset from filtered-empty on searchable pages.
- Replace blank or weak fallback cards with intentional shared state components.

### Plan 11-04: Unfinished Surface And Copy Cleanup + Phase Verification
- Remove milestone/debug wording from operator dashboard, processing page, specialist detail, admin overview, monitoring/settings explanatory text where it leaks internals.
- Eliminate visible “stub/shell/placeholder” language from main flows.
- Run focused frontend verification: lint, tests, build, and manual pass over operator/admin critical paths.

## Open Questions

1. **Should technical diagnostics on results be always visible or collapsed by default?**
   - What we know: Technical data is required for support/debugging, but current inline presentation hurts operator readability.
   - What's unclear: Whether operators use those fields routinely or only in exceptional cases.
   - Recommendation: Plan for a secondary “technical details” section with lower visual priority; default-collapsed is acceptable if it remains easily discoverable.

2. **How much admin copy should stay technical on monitoring/settings pages?**
   - What we know: Admin pages can be more technical than operator pages, but several current descriptions still talk about milestone internals rather than user tasks.
   - What's unclear: Whether the user wants pure admin task language or a light technical explanation.
   - Recommendation: Keep technical nouns where they help operation, but remove phase/history language and endpoint-oriented prose.

## Sources

### Primary (HIGH confidence)
- Local repo audit: `/home/vadim/diplom/.planning/PROJECT.md`, `/home/vadim/diplom/.planning/ROADMAP.md`, `/home/vadim/diplom/.planning/REQUIREMENTS.md`, `/home/vadim/diplom/.planning/phases/07-verification-evidence-and-requirement-revalidation/07-UI-REVIEW.md`
- Local product/docs audit: `/home/vadim/diplom/docs/00_project.md`, `/home/vadim/diplom/docs/01_contract.md`, `/home/vadim/diplom/docs/02_implementation.md`
- Local frontend code audit: operator/admin routes and shared components under `/home/vadim/diplom/frontend/app` and `/home/vadim/diplom/frontend/components`
- Next.js official docs: https://nextjs.org/docs/app/building-your-application/routing/loading-ui-and-streaming
- TanStack Query official docs: https://tanstack.com/query/latest/docs/framework/react/reference/useQuery
- Radix Alert Dialog official docs: https://www.radix-ui.com/primitives/docs/components/alert-dialog
- Package versions verified on 2026-03-24 with local `npm view` calls against npm registry

### Secondary (MEDIUM confidence)
- None needed; repo and official docs were sufficient for this phase

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - directly verified from `frontend/package.json` and npm registry
- Architecture: HIGH - phase scope is grounded in current repo structure and milestone docs
- Pitfalls: HIGH - derived from direct code audit plus the shipped Phase 07 UI review

**Research date:** 2026-03-24
**Valid until:** 2026-04-23
