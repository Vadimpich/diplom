# Phase 09: Admin Users & Questionnaires - Research

**Researched:** 2026-03-24
**Domain:** Frontend-first completion of admin user and questionnaire CRUD on the existing Next.js + Go stack
**Confidence:** HIGH

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| ADMN-01 | Администратор может просматривать список пользователей с их ролями и статусом доступа | Existing `GET /users` + admin-only RBAC already satisfy the data contract; work is primarily list-state completion and clearer status presentation. |
| ADMN-02 | Администратор может создавать, редактировать и изменять роль пользователя через UI с серверной валидацией прав | Existing `POST /users`, `GET /users/{id}`, `PUT /users/{id}` and admin-only router wiring support the workflow; remaining work is UX hardening plus one minimal backend fix around audit behavior. |
| QSTR-01 | Администратор может создавать опросник с названием, описанием и списком вопросов | Existing questionnaire contract and create screen already support this; remaining work is state coverage, feedback consistency, and builder ergonomics. |
| QSTR-02 | Администратор может редактировать состав, порядок и состояние публикации опросника | `PUT /questionnaires/{id}` already replaces the full ordered array and persists `is_active`; remaining work is explicit reorder UX and guardrails around destructive/question ordering actions. |
</phase_requirements>

## Summary

Phase 9 is not a greenfield admin build. The core backend already exposes admin-only user and questionnaire CRUD routes, and the frontend already has list/create/edit pages for both domains. The main gap is that these surfaces are still stage-3 level shells: they lack complete loading/empty/error handling, they expose contract-oriented copy, questionnaire ordering is only implicit through array order, and feedback patterns are inconsistent with the Phase 8 design-system foundation.

The backend should remain mostly unchanged. RBAC is already enforced in the router, user create/update validation exists in `auth.Service`, and questionnaire create/update semantics are already contract-defined. The only backend issue worth bundling into Phase 9 is a correctness bug in `GetUserByID`: it currently emits an `admin.user_updated` audit event on a read path, which will pollute audit history once the admin UI begins opening user detail pages routinely.

The clean plan boundary is three waves: first normalize users list/detail/create UX on top of existing APIs; second finish the questionnaire builder with explicit ordered-question editing and publication-state UX; third run phase-level polish and regression verification focused on list states, mutation feedback, and contract fidelity. This keeps the milestone frontend-first and avoids backend refactors or scope expansion into settings, monitoring, ML, or decision surfaces.

**Primary recommendation:** Treat Phase 9 as frontend completion over shipped admin contracts, with one targeted backend bugfix for erroneous audit emission on `GET /users/{id}`.

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Next.js | `15.5.14` | App Router admin pages and route-level loading boundaries | Already shipped in the repo and aligned with Phase 8 contour structure |
| React | `19.2.4` | Client UI for admin CRUD flows | Current app baseline; no migration belongs in this phase |
| TypeScript | `5.9.3` | Contract-safe admin forms and API types | Existing repo-wide typed contract discipline |
| `@tanstack/react-query` | `5.91.2` | Query/mutation orchestration and cache invalidation | Already used across admin pages and fits list/detail mutation refresh |
| `react-hook-form` | `7.71.2` | Local form state for user/questionnaire editors | Already used on all current admin forms |
| `zod` + `@hookform/resolvers` | `3.25.76` + `3.10.0` | Client-side validation aligned with backend payload shape | Existing repo pattern for forms |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| shadcn/ui primitives | repo-local | Cards, alerts, inputs, buttons, page headers | Default UI layer for all admin surfaces in this milestone |
| `sonner` | `2.0.7` | Global toast feedback | Use for mutation success/error acknowledgement beyond inline alerts |
| `@radix-ui/react-alert-dialog` | `1.1.15` | Confirmation dialogs | Use for destructive/question-removal flows where confirmation is warranted |
| `chi` | `5.2.3` | Backend admin route grouping and RBAC middleware | Keep existing route structure; no handler framework changes |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Existing RHF field-array builder | Separate drag-and-drop library | Not needed for Phase 9; explicit move up/down controls satisfy ordered-question requirements with less surface area |
| Existing query invalidation pattern | Global form state/store | Unnecessary for this phase; current pages are local CRUD views, not cross-app collaborative state |

**Installation:** None. Use the currently shipped workspace dependencies.

## Architecture Patterns

### Recommended Project Structure

```text
frontend/app/(app)/admin/
├── users/
│   ├── page.tsx          # list with loading/empty/error coverage
│   ├── new/page.tsx      # create form
│   └── [id]/page.tsx     # edit form
├── questionnaires/
│   ├── page.tsx          # list with loading/empty/error coverage
│   ├── new/page.tsx      # create builder
│   └── [id]/page.tsx     # edit builder
frontend/components/admin/
├── user-list.tsx
├── user-form.tsx
├── questionnaire-list.tsx
└── questionnaire-builder.tsx
```

Use extracted admin components only when they remove duplication between `new` and `[id]` routes. Do not introduce a generic entity framework.

### Pattern 1: Keep Admin CRUD Contract-First

**What:** Frontend forms should map directly to the current backend payloads instead of inventing view-specific data models.
**When to use:** All Phase 9 user and questionnaire mutations.
**Example:**

```ts
const updateMutation = useMutation({
  mutationFn: (values: UpdateUserValues) => apiClient.updateUser(userId, values),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ["users"] });
    await queryClient.invalidateQueries({ queryKey: ["user", userId] });
  },
});
```

Source: existing repo pattern in `frontend/app/(app)/admin/users/[id]/page.tsx` plus TanStack Query invalidation guidance at `https://tanstack.com/query/latest/docs/framework/react/guides/invalidations-from-mutations`

### Pattern 2: Ordered Questions Stay as a Single Full-Replacement Array

**What:** The frontend should treat questionnaire editing as editing one ordered array of question texts, because `PUT /questionnaires/{id}` replaces the entire `questions` composition and recalculates `position` from array order.
**When to use:** Questionnaire create/edit and reorder interactions.
**Example:**

```ts
const fieldArray = useFieldArray({
  control: form.control,
  name: "questions",
});

fieldArray.move(index, index - 1);
fieldArray.move(index, index + 1);
```

Source: current full-replacement contract in `docs/01_contract.md` and React Hook Form `useFieldArray` docs at `https://react-hook-form.com/docs/usefieldarray`

### Pattern 3: Route-Level Skeleton, Page-Level Empty/Error

**What:** Keep contour-level `loading.tsx` for navigation transitions, but add explicit in-page loading, empty, and error states for data-backed list/detail surfaces.
**When to use:** `/admin/users`, `/admin/questionnaires`, and detail pages fetching by `id`.
**Example:**

```tsx
if (usersQuery.isLoading) {
  return <AdminUsersListSkeleton />;
}

if (usersQuery.isError) {
  return <Alert variant="danger">Не удалось загрузить пользователей.</Alert>;
}

if ((usersQuery.data?.items.length ?? 0) === 0) {
  return <EmptyState title="Пользователей пока нет" />;
}
```

Source: Phase 7 UI review requirements for admin lists and existing `frontend/app/(app)/admin/loading.tsx`

### Pattern 4: Shared Mutation Feedback Primitives

**What:** Use one feedback contract across admin forms: disabled submit while pending, inline field validation, toast or success alert after save, and destructive confirmations where content can be removed.
**When to use:** User create/edit, questionnaire create/edit, question removal, leaving dirty forms if that is introduced.

### Anti-Patterns to Avoid

- **Do not add a new backend questionnaire reorder endpoint.** The current contract already encodes ordering via array position.
- **Do not turn Phase 9 into settings/monitoring/audit UI work.** Those belong to Phase 10.
- **Do not redesign admin routing.** Phase 8 already established the contour shell and `/admin` overview entry.
- **Do not hand-roll client-side derived role permissions.** Backend RBAC is already authoritative.
- **Do not infer “published” from a new field.** The current publication-state contract is `is_active`; preserve that unless a contract change is explicitly approved.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Async cache refresh after save | Custom local mirror state across list/detail pages | TanStack Query invalidation | Existing stack already supports reliable refresh after mutation |
| Ordered questionnaire editing | Manual DOM reordering state outside form model | `react-hook-form` `useFieldArray` | Keeps ordering and validation inside one form source of truth |
| Confirmation UX | Page-local modal logic per route | Shared `ConfirmDialog` primitive from Phase 8 | Prevents feedback drift across admin pages |
| Success/error notifications | Ad hoc inline copy per page only | Shared alert/toast pattern | Phase 8 already introduced common feedback primitives |

**Key insight:** Phase 9 is mostly about finishing existing patterns consistently, not inventing new infrastructure.

## Current State Audit

### What Already Exists

| Area | Existing Capability | Evidence | Confidence |
|------|---------------------|----------|------------|
| Backend RBAC | `/users` and `/questionnaires` are admin-only | `core-backend/internal/http/router.go:89-98` | HIGH |
| User list/create/edit APIs | `GET /users`, `GET /users/{id}`, `POST /users`, `PUT /users/{id}` are wired and consumed by frontend | `frontend/lib/api/client.ts`, `core-backend/internal/http/auth_handler.go` | HIGH |
| User status data | User DTO already exposes `role` and `is_active` | `frontend/lib/api/types.ts`, current list page | HIGH |
| Questionnaire list/create/edit APIs | `GET/POST/GET by id/PUT /questionnaires` exist and are already used | `docs/01_contract.md`, `frontend/lib/api/client.ts` | HIGH |
| Questionnaire publication state | Contract and forms already use `is_active` | `docs/01_contract.md`, questionnaire pages | HIGH |
| Ordered-question persistence | Backend stores position by incoming array order | `docs/01_contract.md:1635-1640`, `core-backend/internal/questionnaires/repository.go:123-137` | HIGH |

### Gaps Remaining by Requirement

| Requirement | Remaining Gap | Notes |
|------------|---------------|-------|
| ADMN-01 | User list lacks explicit loading, empty, and error states; access status is shown, but presentation is minimal and tied to raw fetch success | Current page maps directly over `usersQuery.data?.items ?? []` |
| ADMN-02 | Forms exist, but success/error UX is uneven, pending states are thin, and backend read path incorrectly emits update audit events | Server validation is already present; UX completion is the main work |
| QSTR-01 | Create form exists, but builder ergonomics are basic and list/detail states are incomplete | No major backend gap |
| QSTR-02 | Edit form supports add/remove and `is_active`, but reorder is only implicit; there are no explicit move controls or confirmations for destructive edits | Full-replacement semantics are already contract-safe |

## Common Pitfalls

### Pitfall 1: Blank admin lists during loading or empty fetches

**What goes wrong:** A list card renders nothing while the query is loading or when there are zero items.
**Why it happens:** Current list pages map directly over optional arrays without distinct state branches.
**How to avoid:** Add explicit `isLoading`, `isError`, and zero-length branches before rendering rows.
**Warning signs:** Empty white card on first load; no copy explaining empty system state.

### Pitfall 2: Reorder UI that does not match backend semantics

**What goes wrong:** The UI pretends to support drag/reorder, but the saved payload does not reflect the visible order.
**Why it happens:** `PUT /questionnaires/{id}` is full replacement by array position, so any builder state outside the field-array ordering becomes a drift source.
**How to avoid:** Keep order inside `useFieldArray` and serialize the field order directly into `questions`.
**Warning signs:** Reorder controls mutate local display only, or `position` is computed separately from submission order.

### Pitfall 3: Polluting audit history from read-only admin navigation

**What goes wrong:** Opening a user detail page records `admin.user_updated` even when no change happened.
**Why it happens:** `auth.Service.GetUserByID` currently appends an update audit event on the read path.
**How to avoid:** Remove audit emission from `GetUserByID`; only `CreateUser` and `UpdateUser` should emit mutation events.
**Warning signs:** Audit log grows after page navigation without form submission.

### Pitfall 4: Scope creep into Phase 10

**What goes wrong:** Admin CRUD work expands into monitoring/settings/audit list features.
**Why it happens:** Admin shell now has overview entry points, so adjacent surfaces are visually nearby.
**How to avoid:** Restrict Phase 9 to users/questionnaires only; keep monitoring/settings/audit visibility untouched.
**Warning signs:** New API requests to `/ready`, `/metrics`, or audit endpoints appear in Phase 9 plans.

## Code Examples

Verified patterns to reuse:

### User Mutation with Cache Refresh

```tsx
const updateMutation = useMutation({
  mutationFn: (values: UpdateUserValues) => apiClient.updateUser(userId, values),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ["users"] });
    await queryClient.invalidateQueries({ queryKey: ["user", userId] });
  },
});
```

Source: existing repo pattern in `frontend/app/(app)/admin/users/[id]/page.tsx` and TanStack Query React docs

### Ordered Questionnaire Builder

```tsx
const fieldArray = useFieldArray({
  control: form.control,
  name: "questions",
});

<Button type="button" onClick={() => fieldArray.move(index, index - 1)}>
  Выше
</Button>
<Button type="button" onClick={() => fieldArray.move(index, index + 1)}>
  Ниже
</Button>
```

Source: existing questionnaire form structure in `frontend/app/(app)/admin/questionnaires/[id]/page.tsx` and React Hook Form `useFieldArray` docs

## Recommended Plan Decomposition

### Wave 1: Users CRUD Completion

**Scope**
- Finish `/admin/users` list-state UX.
- Normalize `/admin/users/new` and `/admin/users/[id]` feedback, pending states, and validation messaging.
- Apply the minimal backend fix removing erroneous mutation audit emission from `GetUserByID`.

**Why first**
- `ADMN-01` and `ADMN-02` are the most contract-stable part of the phase.
- The backend fix is isolated and low-risk.
- This wave creates the feedback/state pattern that questionnaires should reuse.

### Wave 2: Questionnaire Builder Completion

**Scope**
- Finish `/admin/questionnaires` list-state UX.
- Refactor questionnaire create/edit into a shared builder shape if it removes duplication.
- Add explicit move up/down ordering controls.
- Add confirmation around destructive question removal when it would discard existing text.
- Make publication state UX explicit around `is_active`.

**Why second**
- Question ordering and destructive editing need deliberate UX choices, but no backend redesign.
- This wave closes `QSTR-01` and `QSTR-02` cleanly on the current contract.

### Wave 3: Phase Polish and Verification

**Scope**
- End-to-end admin smoke across users and questionnaires.
- Cross-page consistency pass for empty/error/loading/mutation feedback.
- Contract/doc touch-ups only if implementation changes payloads or semantics.

**Why third**
- Prevents each CRUD page from inventing its own feedback/state behavior.
- Keeps verification focused on the exact Phase 9 success criteria.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Stage-3 placeholder admin CRUD pages | Phase 8 foundation + contour-aware admin shell with shared feedback primitives | 2026-03-24 | Phase 9 should complete CRUD UX on top of the new shared shell instead of reworking layout foundations |
| Client-only admin gating risk | Backend-enforced admin route groups | Shipped before v1.1 planning | Server already validates permissions; UI should trust API `403` boundaries |

**Deprecated/outdated:**
- Blank list cards as implicit loading/empty state: explicitly called out as a shipped UX issue in the Phase 7 review.
- Implicit questionnaire ordering by form position alone: contract still uses array order, but UX now needs explicit reorder controls to satisfy `QSTR-02`.

## Open Questions

1. **Should Phase 9 include delete for users or questionnaires?**
   - What we know: Neither the roadmap nor the current contracts require delete for this phase.
   - What's unclear: The admin UX may feel incomplete without destructive actions.
   - Recommendation: Do not add delete in Phase 9 unless the user explicitly expands scope; current requirements are satisfied by list/create/edit/role/publication flows.

2. **Should questionnaire “published” copy remain mapped to `is_active`?**
   - What we know: The current backend contract only exposes `is_active`.
   - What's unclear: Product copy may prefer “published/draft”.
   - Recommendation: UI copy can label `is_active` as publication state, but do not rename the contract field in this phase.

## Environment Availability

Step 2.6: SKIPPED (no external dependencies identified). Phase 9 is code/UI work on the existing project stack.

## Sources

### Primary (HIGH confidence)
- Repo sources:
  - `/home/vadim/diplom/.planning/REQUIREMENTS.md`
  - `/home/vadim/diplom/.planning/ROADMAP.md`
  - `/home/vadim/diplom/.planning/phases/07-verification-evidence-and-requirement-revalidation/07-UI-REVIEW.md`
  - `/home/vadim/diplom/frontend/app/(app)/admin/users/page.tsx`
  - `/home/vadim/diplom/frontend/app/(app)/admin/users/[id]/page.tsx`
  - `/home/vadim/diplom/frontend/app/(app)/admin/users/new/page.tsx`
  - `/home/vadim/diplom/frontend/app/(app)/admin/questionnaires/page.tsx`
  - `/home/vadim/diplom/frontend/app/(app)/admin/questionnaires/[id]/page.tsx`
  - `/home/vadim/diplom/frontend/app/(app)/admin/questionnaires/new/page.tsx`
  - `/home/vadim/diplom/frontend/lib/api/client.ts`
  - `/home/vadim/diplom/frontend/lib/api/types.ts`
  - `/home/vadim/diplom/core-backend/internal/http/router.go`
  - `/home/vadim/diplom/core-backend/internal/http/auth_handler.go`
  - `/home/vadim/diplom/core-backend/internal/http/questionnaires_handler.go`
  - `/home/vadim/diplom/core-backend/internal/auth/service.go`
  - `/home/vadim/diplom/core-backend/internal/questionnaires/service.go`
  - `/home/vadim/diplom/core-backend/internal/questionnaires/repository.go`
  - `/home/vadim/diplom/docs/01_contract.md`

### Secondary (MEDIUM confidence)
- TanStack Query React docs: `https://tanstack.com/query/latest/docs/framework/react/guides/invalidations-from-mutations`
- React Hook Form docs: `https://react-hook-form.com/docs/usefieldarray`

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - versions and usage are directly visible in the workspace.
- Architecture: HIGH - route structure, form patterns, and backend contracts are already shipped.
- Pitfalls: HIGH - most are directly evidenced by current code or Phase 7 review findings.

**Research date:** 2026-03-24
**Valid until:** 2026-04-23
