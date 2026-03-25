# Phase 12: ui - Research

**Researched:** 2026-03-25
**Domain:** Frontend UI refinement and audit-closure on the existing Next.js operator/admin application
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

No explicit `## Decisions` section exists in `/home/vadim/diplom/.planning/phases/12-ui/12-CONTEXT.md`; the phase context itself is the locked scope source.

Verbatim locked scope excerpts:

> Довести frontend до более строгого, плотного, корпоративного и прикладного уровня:
> - убрать presentation/demo-style и developer-facing copy;
> - повысить плотность полезных данных;
> - сделать operator/admin экраны рабочими инструментами, а не описанием системы;
> - закрыть остаточные визуальные, информационные и UX-недостатки после Phase 11.

> Если какие-то пункты уже выполнены, например:
> - skeletons;
> - подтверждения действий;
> - success feedback;
> - empty states;
>
> их нужно просто проверить и сохранить, а не переделывать заново без причины.

> Phase 12 должна использовать этот документ как подробный scope source.
>
> При последующем `plan-phase 12` сокращать этот контекст до общих формулировок нельзя: в планировании нужно сохранить поэкранную детализацию и использовать её как явный checklist coverage.

The full per-screen checklist remains locked in:
- `/home/vadim/diplom/.planning/phases/12-ui/12-CONTEXT.md`

### Claude's Discretion

No explicit `## Claude's Discretion` section exists in `/home/vadim/diplom/.planning/phases/12-ui/12-CONTEXT.md`.

Interpretation for planning:
- frontend implementation details may vary;
- stack migration is not allowed;
- scope coverage must still map back to the locked per-screen checklist.

### Deferred Ideas (OUT OF SCOPE)

No explicit `## Deferred Ideas` section exists in `/home/vadim/diplom/.planning/phases/12-ui/12-CONTEXT.md`.
</user_constraints>

## Project Constraints (from AGENTS.md)

- Frontend work must use only contracts from `/home/vadim/diplom/docs/01_contract.md`; do not invent new API fields in UI code.
- Do not change backend or contracts as incidental frontend work.
- If a contract change becomes necessary, `/home/vadim/diplom/docs/01_contract.md` must be updated in the same phase.
- After execution, append a concise entry to `/home/vadim/diplom/docs/02_implementation.md`.
- Keep changes minimal; do not refactor unrelated areas.
- Critical UX flows must preserve observability, error handling, and explicit failure behavior.
- Do not expose secrets, internal-only details, or raw stack traces in UI.

## Summary

Phase 12 is a refinement phase, not a platform phase. The repo already has the required frontend foundation: separate operator/admin contours, shared shell/layout primitives, shared `Skeleton`/`EmptyState`/`Alert`/`ConfirmDialog`/toast flows, TanStack Query for async state, and tokenized Tailwind styling. The correct planning stance is to preserve that foundation and replace presentation-heavy, developer-facing, low-density screens with denser operator/admin work surfaces.

The biggest planning risk is assuming every audit comment is frontend-only. Current contracts support some density upgrades immediately, especially on operator history, specialist detail, examination status/progress, and admin registries. But several requested fields do not exist in the current DTOs: user last login, questionnaire usage/editor metadata, dashboard-level active user/service/error aggregates, richer monitoring metrics, and specialist-level summary fields beyond what can be derived from current result-history and examination-history endpoints. The plan should therefore separate:
- frontend-only fixes using existing DTOs;
- optional backend-support tasks only where the locked UI scope cannot be honestly satisfied otherwise.

Another important constraint is that Phase 11 already shipped honest state handling for many screens. Phase 12 should not rebuild those patterns. It should keep shared feedback/state primitives and tighten copy, layout density, hierarchy, and operational usefulness screen by screen.

**Primary recommendation:** Plan Phase 12 as a screen-by-screen densification pass on the existing Next 15 + Tailwind 3 stack, with explicit contract-gap callouts instead of hidden UI assumptions.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Next.js App Router | `15.5.14` in repo; latest registry `16.2.1` on 2026-03-24 | routing, layouts, `loading.tsx`, route-level error handling | Locked project decision keeps current Next 15 stack; App Router conventions already underpin contour/layout/loading behavior |
| React | `19.2.4` in repo and latest registry `19.2.4` on 2026-03-24 | component runtime | Existing app already uses client/server boundary patterns compatible with current UI work |
| Tailwind CSS | `3.4.19` in repo; latest registry `4.2.2` on 2026-03-24 | tokens, density, responsive layout | Locked project decision keeps Tailwind 3; design tokens are already encoded in `globals.css` and `tailwind.config.ts` |
| TanStack Query | `5.91.2` in repo; latest registry `5.95.2` on 2026-03-23 | async data loading, refetching, honest query state handling | Existing app already standardizes fetch state via Query; Phase 12 should deepen that usage instead of custom state machines |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| React Hook Form | `7.71.2` in repo; latest registry `7.72.0` on 2026-03-22 | admin/operator forms | Keep for create/edit pages; do not replace form state tooling during a UI-only phase |
| Zod | `3.25.76` in repo; latest registry `4.3.6` on 2026-01-25 | schema validation | Keep existing form validation path; no schema migration in this phase |
| Sonner | `2.0.7` in repo and latest registry `2.0.7` on 2025-08-02 | success/error toasts | Already mounted app-wide; use for post-action feedback |
| Radix Alert Dialog | `1.1.15` in repo | confirmation modal primitive | Reuse via existing `ConfirmDialog`; do not build page-local confirms |
| Lucide React | `0.511.0` in repo | restrained iconography | Use sparingly; audit scope explicitly calls for less decorative noise |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Existing card/grid/list primitives | TanStack Table or AG Grid | Overkill for this phase; adds migration cost and new abstraction surface when current lists are still small and contract-limited |
| Existing shared feedback primitives | Page-local modals/toasts/spinners | Causes UX drift and duplicates solved Phase 8/11 work |
| Existing Next 15 + Tailwind 3 stack | Next 16/Tailwind 4 upgrade | Explicitly out of scope; stack migration would dominate a polish phase |

**Installation:**
```bash
# No new packages are recommended for Phase 12.
# Stay on the current frontend dependency set.
```

**Version verification:** Upstream package versions were verified on 2026-03-25 with `npm view`. Latest registry versions are newer than the repo for `next`, `@tanstack/react-query`, `tailwindcss`, `react-hook-form`, and `zod`, but the locked project decision is to keep the current Next 15 + Tailwind 3 stack for this milestone.

## Architecture Patterns

### Recommended Project Structure
```text
frontend/
├── app/
│   ├── (auth)/login/              # compact auth screen
│   └── (app)/
│       ├── operator/              # operator pages only
│       └── admin/                 # admin pages only
├── components/
│   ├── layout/                    # AppShell, contour shells, guards
│   ├── ui/                        # shared primitives and state surfaces
│   ├── operator/                  # operator-specific summary/status widgets
│   └── admin/                     # admin registries, filters, builders
└── lib/
    ├── api/                       # DTOs and client calls
    ├── navigation/                # role-home and route helpers
    └── operator/                  # examination flow helpers
```

### Pattern 1: Preserve Contour Shells, Simplify Their Content
**What:** Keep `AppShell`, `OperatorShell`, and `AdminShell`, but remove verbose nav descriptions and decorative summary blocks called out by Phase 12 scope.
**When to use:** Any navigation/sidebar change in this phase.
**Example:**
```tsx
// Source: local project pattern + locked Phase 12 scope
<AppShell
  contour="operator"
  navSections={[
    { items: [
      { href: "/operator", label: "Рабочее место" },
      { href: "/operator/specialists", label: "Специалисты" },
      { href: "/operator/history", label: "История" },
    ]},
  ]}
>
  {children}
</AppShell>
```

### Pattern 2: Registry-First List Screens
**What:** Prefer dense registries with sortable/filterable headers, narrow status cells, and row-level navigation over large marketing-style cards.
**When to use:** `specialists`, `users`, `questionnaires`, `history`, and likely portions of dashboard screens.
**Example:**
```tsx
// Source: https://tailwindcss.com/docs/table-layout
<div className="overflow-x-auto rounded-2xl border border-border/70">
  <table className="min-w-full table-fixed text-sm">
    <thead className="bg-secondary/50 text-xs uppercase tracking-[0.16em] text-muted-foreground">
      <tr>
        <th className="w-[32%] px-4 py-3 text-left">Специалист</th>
        <th className="w-[18%] px-4 py-3 text-left">Статус</th>
        <th className="w-[18%] px-4 py-3 text-left">Последнее обследование</th>
        <th className="w-[16%] px-4 py-3 text-left">Отклонение</th>
        <th className="w-[16%] px-4 py-3 text-right">Действие</th>
      </tr>
    </thead>
  </table>
</div>
```

### Pattern 3: Query-State Trident
**What:** Handle async surfaces in the order `pending -> error -> success`, and only render empty states inside the success branch.
**When to use:** Every page or shared list driven by TanStack Query.
**Example:**
```tsx
// Source: https://tanstack.com/query/latest/docs/framework/react/guides/queries
const { isPending, isError, data, error } = useQuery({
  queryKey: ["specialists"],
  queryFn: apiClient.getSpecialists,
});

if (isPending) return <SpecialistsSkeleton />;
if (isError) return <Alert variant="danger">{error.message}</Alert>;
if (!data.items.length) return <EmptyState title="Список пуст" description="..." />;

return <SpecialistsRegistry items={data.items} />;
```

### Pattern 4: Route-Level Loading and Error Surfaces
**What:** Use `loading.tsx` for instant route feedback and `error.tsx` for segment-level uncaught failures; reserve page-local alerts for handled query/mutation errors.
**When to use:** Initial route transitions and high-level page failure boundaries.
**Example:**
```tsx
// Source: https://nextjs.org/docs/app/api-reference/file-conventions/loading
export default function Loading() {
  return <DashboardSkeleton />;
}
```

### Anti-Patterns to Avoid
- **Developer-facing copy in user UI:** `backend`, `runtime`, `source of truth`, `contract`, `placeholder`, `mutation event` must not appear on shipped screens.
- **Decorative side panels on work pages:** operator/admin screens should privilege current actions and current status, not explanatory prose.
- **Error-as-empty:** do not turn failed queries into `0` counters, empty registries, or fake "everything is fine" cards.
- **Global primitive churn for page-local polish:** avoid broad token or shared-component rewrites unless multiple screens need the exact same correction.
- **Invented fields:** do not show “last login”, “questionnaire in use”, “queue backlog”, or “active users” unless a real DTO supports them.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Confirmation for destructive actions | page-local modal implementations | existing `ConfirmDialog` | Already integrated and consistent with Phase 8/11 flows |
| Success feedback | ad hoc banners per page | shared Sonner toaster | Centralized, already mounted at app level |
| Loading placeholders | page-specific spinner-only blocks | `loading.tsx` + shared `Skeleton` layouts | Next App Router supports instant route loading; skeletons preserve final layout rhythm |
| Async fetch state machines | manual `useState` fetch flags everywhere | TanStack Query | Existing app is already query-driven; custom state invites drift and hidden error states |
| Dense registry framework | new table/grid dependency | Tailwind table/grid + existing badges/buttons/cards | Current dataset size and phase scope do not justify a table-library introduction |
| Role-aware redirects | page-local hardcoded redirects | existing role-home helper and route guard | Already canonical in the repo; changing this would be unrelated scope |

**Key insight:** Phase 12 should spend effort on information hierarchy and truthful operator/admin language, not on replacing solved infrastructure primitives.

## Common Pitfalls

### Pitfall 1: Trying to Satisfy Density Requirements with More Text
**What goes wrong:** screens become even busier, but still not more useful.
**Why it happens:** explanatory cards are easier to write than data-dense registries.
**How to avoid:** every major block should answer who/what/when/status/next action.
**Warning signs:** long `CardDescription` blocks, hero panels, or sidebars taking more space than actionable data.

### Pitfall 2: Planning UI Fields That the Current Contract Cannot Supply
**What goes wrong:** plan includes impossible columns like user last login or questionnaire usage.
**Why it happens:** the audit scope asks for richer data than current DTOs expose.
**How to avoid:** check `/home/vadim/diplom/docs/01_contract.md` and `frontend/lib/api/types.ts` before committing any screen-level task.
**Warning signs:** UI copy references fields absent from `User`, `Questionnaire`, `Specialist`, `Examination`, `SystemSettings`, or monitoring DTOs.

### Pitfall 3: Regressing Already-Fixed Honest States
**What goes wrong:** skeletons, confirms, and empty states get replaced by simpler but weaker polish.
**Why it happens:** teams treat a polish phase as a rewrite instead of a tighten-and-preserve pass.
**How to avoid:** keep existing feedback primitives and check the Phase 12 “already done, preserve it” constraint.
**Warning signs:** removing `ConfirmDialog`, replacing skeletons with blank white cards, or dropping toasts after mutations.

### Pitfall 4: Using Global Shared Primitives as a Dumping Ground for Page-Specific Fixes
**What goes wrong:** small fixes on one screen destabilize typography, spacing, and controls everywhere.
**Why it happens:** page-level density issues are pushed into `Card`, `PageHeader`, or `Button` too early.
**How to avoid:** change shared primitives only when the same defect clearly appears across multiple audited screens.
**Warning signs:** touching `globals.css`, `card.tsx`, `button.tsx`, or `page-header.tsx` to solve a single-screen complaint.

### Pitfall 5: Keeping Dashboard Screens as Feature Descriptions
**What goes wrong:** dashboards still read like onboarding or architecture overviews.
**Why it happens:** placeholder overview cards survive longer than they should.
**How to avoid:** dashboards must become “what needs attention now” surfaces, not “what the product can do” surfaces.
**Warning signs:** boxes titled like “Поток обследования”, “Фокус текущего этапа”, or “Что остаётся вне панели”.

## Code Examples

Verified patterns from official sources:

### Instant Route Loading
```tsx
// Source: https://nextjs.org/docs/app/api-reference/file-conventions/loading
export default function Loading() {
  return <PageSkeleton />;
}
```

### Segment Error Boundary
```tsx
// Source: https://nextjs.org/docs/app/getting-started/error-handling
"use client";

export default function Error({
  error,
  reset,
}: {
  error: Error;
  reset: () => void;
}) {
  return (
    <div className="space-y-4">
      <p className="text-sm text-danger">{error.message}</p>
      <button type="button" onClick={reset}>Повторить</button>
    </div>
  );
}
```

### Honest Query State Ordering
```tsx
// Source: https://tanstack.com/query/latest/docs/framework/react/guides/queries
const { isPending, isError, data, error } = useQuery({
  queryKey: ["users"],
  queryFn: apiClient.getUsers,
});

if (isPending) return <UsersSkeleton />;
if (isError) return <Alert variant="danger">{error.message}</Alert>;
if (!data.items.length) return <EmptyState title="Пользователей нет" description="..." />;

return <UsersRegistry items={data.items} />;
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Presentation-heavy contour shell with descriptive nav copy | Shared contour shells with role-specific wrappers and reusable state primitives | Phase 8, 2026-03-24 | Phase 12 should simplify shell content, not replace contour architecture |
| Generic empty/error handling and weak transition states | Skeleton-first, explicit `loading` / `empty` / `error` patterns on key flows | Phase 11, 2026-03-25 | Preserve these mechanics while making the screens denser |
| DTO/debug wording in operator/admin surfaces | Operator-facing/admin-facing copy that hides backend internals | Phase 11, 2026-03-25, but incomplete | Phase 12 must finish the copy cleanup across remaining screens |

**Deprecated/outdated:**
- Dashboard explainer cards that describe product capabilities instead of current work.
- Login promo language mentioning backend contracts, JWT, runtime semantics, or placeholder honesty.
- Admin copy using `operational status`, `runtime`, `source of truth`, and similar internal phrasing.

## Open Questions

1. **Should Phase 12 include contract support for missing registry/dashboard fields?**
   - What we know: current contracts do not provide user last login, questionnaire usage/editor metadata, active user counts, queue depth, or failure counters.
   - What's unclear: whether the milestone expects those items to be approximated from existing data or explicitly added to backend contracts.
   - Recommendation: plan frontend-only coverage first, then add a clearly scoped backend-support subplan only for blockers that prevent honest completion of the locked checklist.

2. **How far should operator/admin dashboards move toward derived aggregates using existing endpoints?**
   - What we know: `GET /examinations`, `GET /specialists`, and result-history endpoints let the frontend derive some counts and recent items.
   - What's unclear: whether deriving these client-side is acceptable for Phase 12 or whether dedicated summary endpoints are preferred.
   - Recommendation: derive lightweight counts from existing endpoints for now; only request backend summary endpoints if the derived view becomes inaccurate or too expensive.

3. **Should dense registries remain card-based or become true tables?**
   - What we know: the audit scope asks for corporate, dense, register-like screens, and Tailwind officially supports fixed-layout tables.
   - What's unclear: whether all target screens should become tables, or whether hybrid “table-like cards” are enough.
   - Recommendation: use real tables where columns matter (`specialists`, `users`, `questionnaires`, `history`); keep cards for narrative detail pages and summaries.

## Sources

### Primary (HIGH confidence)
- `/home/vadim/diplom/.planning/phases/12-ui/12-CONTEXT.md` - locked phase scope and checklist
- `/home/vadim/diplom/.planning/phases/11-operator-ux-visual-completion/11-UI-REVIEW.md` - prior audit findings and unresolved UI risks
- `/home/vadim/diplom/.planning/REQUIREMENTS.md` - shipped v1 scope and out-of-scope boundaries
- `/home/vadim/diplom/docs/00_project.md` - target operator/admin UX model and locked frontend stack
- `/home/vadim/diplom/docs/01_contract.md` - current DTO and endpoint limits for Phase 12
- `/home/vadim/diplom/frontend/package.json` - actual repo stack versions
- `/home/vadim/diplom/frontend/app/globals.css` and `/home/vadim/diplom/frontend/tailwind.config.ts` - current token and theme foundation
- `/home/vadim/diplom/frontend/components/layout/app-shell.tsx` - current contour shell architecture
- `/home/vadim/diplom/frontend/components/ui/*.tsx` - shared feedback and layout primitives
- https://nextjs.org/docs/app/api-reference/file-conventions/loading - `loading.js` behavior and instant loading states
- https://nextjs.org/docs/app/getting-started/error-handling - route error boundaries and manual event-handler error treatment
- https://tanstack.com/query/latest/docs/framework/react/guides/queries - recommended query-state handling order
- https://tailwindcss.com/docs/table-layout - dense fixed-layout table support
- npm registry via `npm view` on 2026-03-25 for `next`, `react`, `@tanstack/react-query`, `tailwindcss`, `react-hook-form`, `zod`, `sonner`

### Secondary (MEDIUM confidence)
- None.

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - locked by project docs, Phase 8 decisions, repo package.json, and verified registry versions
- Architecture: HIGH - grounded in existing repo patterns plus official Next/TanStack/Tailwind guidance
- Pitfalls: HIGH - directly supported by Phase 12 context, Phase 11 UI review, and current screen code

**Research date:** 2026-03-25
**Valid until:** 2026-04-24
