# Phase 08: ui-contours-design-foundation - Research

**Researched:** 2026-03-24
**Domain:** Next.js App Router frontend contour separation and shared design-system foundation
**Confidence:** MEDIUM

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
### UI Contour Identity
- **D-01:** `operator` and `admin` must be two clearly different working contours on top of one shared frontend codebase, not one identical shell with different menu items.
- **D-02:** The `operator` contour should feel task-driven, compact, and focused on fast examination work with minimal distraction.
- **D-03:** The `admin` contour should feel calmer, more informational, and oriented around management, overview, and system control.

### Design-System Foundation
- **D-04:** Phase 8 must lock not only design tokens, but also shared primitives used by both contours.
- **D-05:** The design foundation must cover color, typography, spacing, radius, surfaces, page headers, cards, forms, alerts, badges, and base loading/empty/error states.
- **D-06:** Visual direction is professional, strict, and corporate: modern minimalism, high readability, clear hierarchy, low visual fatigue, and no decorative excess or visual noise.
- **D-07:** Status colors are semantically meaningful and should consistently signal state through green / yellow / red conventions.
- **D-08:** Interactions should feel responsive and polished through restrained transitions, hover states, and motion without adding gratuitous animation.
- **D-09:** Loading skeletons are preferred over generic spinners for primary page and content loading states.
- **D-10:** The system must present explicit loading, empty, and error states, plus alerts, toasts, confirmations, and visible feedback for user actions as shared UX primitives.

### Navigation Model
- **D-11:** Both contours keep a left-sidebar navigation as the shared structural pattern for the application shell.
- **D-12:** The operator sidebar should stay short and action-oriented, emphasizing fast access to the primary workstation flows.
- **D-13:** The admin sidebar may be denser and more system-oriented, including a stronger sense of sections and secondary operational information.

### Role Entry and Landing Behavior
- **D-14:** Role-based redirects remain explicit: operators enter the operator workstation, admins enter the admin contour.
- **D-15:** `operator` should continue to land on an operator dashboard/workstation entry page.
- **D-16:** `admin` should no longer treat `/admin/users` as the de facto home of the whole contour; Phase 8 should introduce an explicit admin home/overview entry point, even if its initial content is minimal.

### Claude's Discretion
- Exact component names, file/module boundaries, and how the design tokens are organized across CSS variables and reusable components.
- Exact visual density, typography values, and spacing scale, provided they remain consistent with the strict corporate design direction captured above.
- Exact implementation details of sidebar sectioning, hover motion, skeleton patterns, and toast/confirmation primitives.

### Deferred Ideas (OUT OF SCOPE)
- Detailed operator results readability fixes from the Phase 07 UI review are executed in Phase 11, not in Phase 8.
- Admin CRUD completion, system settings, monitoring, and audit log delivery belong to Phases 9 and 10 after the contour foundation is in place.
- Advanced charts and richer result visualizations remain future work under `RICH-01`.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CNTR-01 | Пользователь с ролью `operator` работает в отдельном operator-контуре с собственной навигацией, layout и без admin-элементов управления | Recommends contour-specific shell wrappers, operator-only nav model, task-first landing hierarchy, and redirect preservation to `/operator` |
| CNTR-02 | Пользователь с ролью `admin` работает в отдельном admin-контуре с собственной навигацией, layout и без operator-сценариев проведения обследования | Recommends explicit `/admin` overview page, denser admin sidebar, separate admin shell treatment, and redirect updates away from `/admin/users` |
| DSGN-01 | Operator и admin интерфейсы используют единый набор design tokens и переиспользуемых компонентов для типографики, цветов, отступов, форм и состояний | Recommends token-first normalization in `globals.css`/Tailwind theme plus shared primitives for shell, page header, card, empty/error/loading, toast, and confirmation dialog |
</phase_requirements>

## Summary

Phase 8 should be planned as a frontend composition phase, not a broad redesign or dependency migration. The shipped app already has the right macro-structure for this work: Next.js App Router route groups, separate `operator` and `admin` layouts, shared `components/ui/*` primitives, and an `AppShell` that centralizes navigation and session UI. The gap is that contour identity is currently too thin. The two layouts mainly differ by menu arrays and label strings, while root redirects and guard logic still treat `/admin/users` as the admin home.

The correct planning focus is to split the current shared shell into a stable foundation plus contour-specific wrappers, then normalize the base primitives that every later admin/operator page will reuse. That includes tokens in `frontend/app/globals.css`, Tailwind theme mappings, page header rules, cards, badges, forms, empty/error/loading states, and explicit feedback primitives. The phase should also introduce the minimal admin overview landing page now, because both `frontend/app/page.tsx` and `frontend/app/(app)/admin/page.tsx` currently route admins to `/admin/users`, which directly conflicts with Decision `D-16`.

Do not turn this phase into a Tailwind v4 migration, Next 16 upgrade, or page-by-page visual rewrite. The current repo is on Next `15.5.14`, Tailwind `3.4.19`, React Query `5.91.2`, React Hook Form `7.71.2`, and no `components.json` exists for shadcn initialization. The safest plan is to keep the existing shipped stack, align it to the approved UI spec, and add only the missing shared primitives required for later phases.

**Primary recommendation:** Build Phase 8 around `shared foundation + OperatorShell + AdminShell + admin overview redirect fix`, while keeping the existing Next 15/Tailwind 3 stack unchanged.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Next.js | repo pinned `15.5.14` (latest upstream `16.2.1`, modified 2026-03-23) | App Router layouts, route groups, redirects, loading/error boundaries | Existing app already uses App Router correctly; route groups and nested layouts are the native way to separate contours |
| React | repo pinned `19.2.4` | Component model for shared primitives and contour wrappers | Already shipped in repo; no phase need justifies runtime migration |
| Tailwind CSS | repo pinned `3.4.19` (latest upstream `4.2.2`, modified 2026-03-23) | Token-to-utility styling and shell/page composition | Existing theme already maps CSS variables into utilities; foundation work can land without upgrading to v4 |
| shadcn/ui conventions | repo-local components, `components.json` absent | Source-owned primitive patterns for button/card/input/alert/skeleton/dialog/toast/sidebar composition | Matches approved UI-SPEC, keeps primitives in repo control, avoids introducing a heavy runtime design system |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `@tanstack/react-query` | repo pinned `5.91.2` (latest upstream `5.95.2`, modified 2026-03-23) | Data fetching and async UI states | Keep for page-level data states; pair with skeleton/empty/error primitives instead of ad hoc placeholders |
| `react-hook-form` | repo pinned `7.71.2` (latest upstream `7.72.0`, modified 2026-03-22) | Form state | Keep for shared form primitives and admin forms in later phases |
| `zod` | repo pinned `3.25.76` | Schema validation | Already used on login; continue for form-level validation messaging |
| `lucide-react` | repo pinned `0.511.0` (registry reported newer `1.0.1`, modified 2026-03-24) | Shared iconography | Keep current repo version in this phase; no icon-library upgrade is needed to satisfy contour goals |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Keep Next 15 + Tailwind 3 for Phase 8 | Upgrade to Next 16 + Tailwind 4 now | Higher churn, broader QA surface, no direct requirement value for CNTR-01/CNTR-02/DSGN-01 |
| Shared primitives with contour wrappers | Fully separate operator/admin component trees | More duplication and drift risk; later phases would fork every primitive unnecessarily |
| shadcn-style source-owned primitives | Third-party runtime component library | Faster initial scaffolding, but conflicts with existing repo structure and increases visual override work |

**Installation:**
```bash
# Prefer no framework upgrade in Phase 8.
# Add only missing primitive dependencies if implementation chooses official shadcn patterns.
npm install @radix-ui/react-alert-dialog sonner
```

**Version verification:** Verified via `npm view` on 2026-03-24 for `next`, `tailwindcss`, `@tanstack/react-query`, `react-hook-form`, and `lucide-react`. `zod` remained repo-pinned because registry metadata queries were unreliable during research; confidence there is MEDIUM.

## Architecture Patterns

### Recommended Project Structure
```text
frontend/
├── app/
│   ├── (app)/
│   │   ├── admin/
│   │   │   ├── layout.tsx      # AdminShell wrapper + role guard
│   │   │   └── page.tsx        # Admin overview landing
│   │   └── operator/
│   │       ├── layout.tsx      # OperatorShell wrapper + role guard
│   │       └── page.tsx        # Operator workstation landing
│   └── globals.css             # canonical tokens and motion/state variables
├── components/
│   ├── layout/
│   │   ├── shell-frame.tsx     # shared shell primitive
│   │   ├── operator-shell.tsx  # contour wrapper
│   │   ├── admin-shell.tsx     # contour wrapper
│   │   └── route-guard.tsx     # role gate, no contour visuals
│   └── ui/
│       ├── page-header.tsx
│       ├── empty-state.tsx
│       ├── skeleton.tsx
│       ├── confirm-dialog.tsx
│       ├── toaster.tsx
│       └── badge/alert/card/form primitives
└── tailwind.config.ts          # token wiring for utilities
```

### Pattern 1: Shared Foundation, Contour-Specific Shells
**What:** Keep one structural shell primitive and wrap it in `OperatorShell` and `AdminShell` components that supply contour-specific nav grouping, density, color balance, header treatment, and landing emphasis.
**When to use:** Always for `/operator/*` and `/admin/*`; never pass raw menu arrays directly from layouts into one generic shell and call the job done.
**Example:**
```tsx
// Source inspiration: https://nextjs.org/docs/app/building-your-application/routing/route-groups
// Source inspiration: https://ui.shadcn.com/docs/components/sidebar
export function OperatorLayout({ children }: { children: React.ReactNode }) {
  return (
    <RouteGuard requiredRole="operator">
      <OperatorShell>{children}</OperatorShell>
    </RouteGuard>
  );
}
```

### Pattern 2: Token-First Foundation
**What:** Define approved colors, spacing, radius, and motion in CSS variables first; then consume them via Tailwind theme aliases and shared primitives.
**When to use:** Before touching page-level styling. Phase 8 should normalize `globals.css` and `tailwind.config.ts` before restyling operator/admin screens.
**Example:**
```css
/* Source inspiration: local frontend/app/globals.css + approved 08-UI-SPEC.md */
:root {
  --background: 214 33% 97%;
  --primary: 215 32% 18%;
  --accent: 201 94% 92%;
  --success: 145 63% 42%;
  --warning: 38 92% 50%;
  --danger: 0 72% 52%;
  --radius-card: 1.5rem;
  --motion-fast: 150ms;
}
```

### Pattern 3: Segment-Level Loading/Error States
**What:** Use App Router loading/error boundaries and skeleton blocks for primary content states.
**When to use:** For contour landing pages, data lists, and route-guard/session transitions. Keep spinners only for inline blocking actions.
**Example:**
```tsx
// Source inspiration: https://nextjs.org/docs/app/api-reference/file-conventions/loading
// Source inspiration: https://ui.shadcn.com/docs/components/skeleton
export default function Loading() {
  return (
    <div className="space-y-4">
      <Skeleton className="h-8 w-64" />
      <Skeleton className="h-32 w-full" />
      <Skeleton className="h-32 w-full" />
    </div>
  );
}
```

### Pattern 4: Explicit Redirect Ownership
**What:** Centralize role-to-home mapping and make `/admin` the admin home instead of `/admin/users`.
**When to use:** In `app/page.tsx`, login success redirects, and `RouteGuard` fallback redirects.
**Example:**
```tsx
// Source inspiration: https://nextjs.org/docs/app/api-reference/functions/redirect
if (role === "admin") {
  redirect("/admin");
}
if (role === "operator") {
  redirect("/operator");
}
```

### Anti-Patterns to Avoid
- **Contour-by-menu-only:** Current `AppShell` setup makes contours look almost identical; this fails the phase intent even if route trees stay separate.
- **Page-first restyling:** Restyling admin/users, operator/dashboard, and login independently before locking primitives will create drift that later phases must unwind.
- **Spinner-first loading:** `RouteGuard` currently centers a spinner for access checks; full-page shell/content loading should move toward skeleton-compatible states.
- **Stack migration inside UI foundation:** Tailwind v4 or Next 16 migration expands scope without helping the required contour split.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Confirm destructive actions | Custom boolean modal state and ad hoc overlay markup everywhere | `AlertDialog`-style shared confirmation primitive | Accessibility, focus trapping, and consistent destructive copy are easy to get wrong |
| Toast notifications | Bespoke event bus and floating DOM container | `sonner` with one shared app toaster | Standard placement, stacking, dismissal, and reduced custom state |
| Loading placeholders | One-off gray rectangles on each page | Shared `Skeleton` primitive + route `loading.tsx` | Keeps state UX consistent and satisfies D-09 cleanly |
| Sidebar/nav behavior | Many page-owned nav blocks | Shared shell/sidebar primitive with contour wrapper inputs | Prevents contour drift and repeated active-state logic |
| Form field states | Manual label/error/helper composition on every screen | Shared field primitives around existing `Input`, `Textarea`, `Label` | Later admin CRUD phases depend on consistent field/error rendering |

**Key insight:** Phase 8 should standardize primitives once and reuse them. Hand-rolled per-page UI will make Phases 9-11 slower and more brittle.

## Common Pitfalls

### Pitfall 1: Admin Home Never Actually Changes
**What goes wrong:** Teams add a new overview page but leave redirects pointing to `/admin/users`.
**Why it happens:** The current code has redirect ownership in three places: root page, login success, and `RouteGuard`.
**How to avoid:** Plan a single task that updates all role-home redirects together and verifies `admin` lands on `/admin`.
**Warning signs:** An admin can manually visit `/admin`, but fresh login and session mismatch still bounce to `/admin/users`.

### Pitfall 2: Design Tokens Exist, But Primitives Still Drift
**What goes wrong:** `globals.css` changes land, but button/card/header/empty-state visuals keep inconsistent padding, radii, and typography.
**Why it happens:** Token work is done without a pass over the actual shared primitives in `components/ui/*`.
**How to avoid:** Make primitive normalization an explicit plan item before page refinements.
**Warning signs:** `PageHeader`, `Card`, `Button`, and empty/error states still use mismatched sizes after token updates.

### Pitfall 3: Route Guard Feels Like a Different App
**What goes wrong:** Contour pages look polished, but auth/session/loading states still use old spinner-only UI and break the new visual language.
**Why it happens:** Guard/session surfaces are often treated as infrastructure instead of user-facing UI.
**How to avoid:** Include `RouteGuard`, login redirect flow, and auth loading/error states in the same foundation pass.
**Warning signs:** Transition into the app shows centered spinner boxes that do not match contour shells.

### Pitfall 4: Admin and Operator Diverge by Copy, Not by Hierarchy
**What goes wrong:** Colors change slightly, but both shells still share the same sidebar density, same hero weight, and same reading order.
**Why it happens:** Teams optimize for reuse too aggressively and under-encode contour identity.
**How to avoid:** Plan separate wrappers with different sidebar sectioning, title/subtitle treatment, and first-read surface priority.
**Warning signs:** Screenshots of both contours are hard to distinguish after menu text is blurred out.

### Pitfall 5: Phase 8 Quietly Expands into CRUD Completion
**What goes wrong:** The team starts finishing admin features instead of defining the reusable foundation.
**Why it happens:** Existing admin pages are easy targets for polish work.
**How to avoid:** Keep Phase 8 scoped to shell, redirects, overview page, tokens, and primitives only.
**Warning signs:** New domain forms or monitoring widgets appear before toast/dialog/skeleton/page-header work is complete.

## Code Examples

Verified patterns from official sources:

### Contour Redirect Mapping
```tsx
// Source: https://nextjs.org/docs/app/api-reference/functions/redirect
import { redirect } from "next/navigation";

export function redirectToRoleHome(role: "operator" | "admin") {
  redirect(role === "admin" ? "/admin" : "/operator");
}
```

### Shared Loading Skeleton
```tsx
// Source: https://nextjs.org/docs/app/api-reference/file-conventions/loading
// Source: https://ui.shadcn.com/docs/components/skeleton
export default function AdminLoading() {
  return (
    <section className="space-y-4">
      <Skeleton className="h-8 w-56" />
      <div className="grid gap-4 md:grid-cols-2">
        <Skeleton className="h-40 w-full" />
        <Skeleton className="h-40 w-full" />
      </div>
    </section>
  );
}
```

### Shared Confirmation and Toast Entry
```tsx
// Source: https://ui.shadcn.com/docs/components/alert-dialog
// Source: https://ui.shadcn.com/docs/components/radix/sonner
<AlertDialog>
  <AlertDialogTrigger asChild>
    <Button variant="ghost">Выйти</Button>
  </AlertDialogTrigger>
  <AlertDialogContent>{/* canonical destructive copy */}</AlertDialogContent>
</AlertDialog>
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| One generic shell with different nav arrays | Shared foundation plus contour-specific shell wrappers | Current best fit for this repo and UI-SPEC | Meets CNTR-01/CNTR-02 without duplicating all primitives |
| Admin default home is `/admin/users` | Admin default home should be `/admin` overview | Locked in 08-CONTEXT D-16 on 2026-03-24 | Redirect logic and admin index page must change together |
| Centered generic spinner for major page loading | Skeleton-first loading for primary page/content states | Locked in 08-CONTEXT D-09 and 08-UI-SPEC | Requires shared skeleton primitive and selective `loading.tsx` adoption |
| Page-specific visual fixes | Token-first and primitive-first normalization | Best practice for foundation phases | Prevents later admin/operator screens from diverging |

**Deprecated/outdated:**
- Treating `/admin/users` as contour home: outdated for this milestone; replace with explicit admin overview page.
- Defining contour separation only via copy/menu arrays: outdated for this phase because contour identity is now a requirement, not a nice-to-have.

## Open Questions

1. **Should Phase 8 formally initialize shadcn in the repo or keep manual source-owned primitives?**
   - What we know: `08-UI-SPEC.md` locks shadcn as the design-system basis, but `frontend/components.json` is absent and the repo already contains compatible source-owned primitives.
   - What's unclear: Whether planner should include a one-time shadcn init step or just normalize existing components to shadcn conventions.
   - Recommendation: Prefer normalizing existing source-owned primitives first; only initialize shadcn if the executor needs generator support for `sidebar`, `alert-dialog`, or `sonner`.

2. **How far should loading/error normalization go in Phase 8?**
   - What we know: Shared loading/empty/error primitives are in scope, but Phase 11 owns broad list-state remediation.
   - What's unclear: Whether to retrofit every existing route now or only shell-critical pages and shared primitives.
   - Recommendation: Scope Phase 8 to shell-critical routes, landing pages, route guard, and reusable primitives; let Phase 11 finish broad per-page rollout.

## Sources

### Primary (HIGH confidence)
- Local source: `/home/vadim/diplom/.planning/phases/08-ui-contours-design-foundation/08-CONTEXT.md` - locked decisions and phase boundary
- Local source: `/home/vadim/diplom/.planning/phases/08-ui-contours-design-foundation/08-UI-SPEC.md` - approved visual contract and component inventory
- Local source: `/home/vadim/diplom/docs/00_project.md` - operator/admin contour architecture and frontend stack
- Local source: `/home/vadim/diplom/docs/01_contract.md` - status vocabulary and runtime/admin data constraints
- Local source: `/home/vadim/diplom/frontend/components/layout/app-shell.tsx` - current shell implementation to evolve
- https://nextjs.org/docs/app/building-your-application/routing/route-groups - App Router route-group and layout pattern
- https://nextjs.org/docs/app/api-reference/file-conventions/loading - segment loading boundaries
- https://nextjs.org/docs/app/api-reference/functions/redirect - redirect behavior for role-home mapping
- https://ui.shadcn.com/docs/components/sidebar - sidebar composition pattern
- https://ui.shadcn.com/docs/components/skeleton - skeleton primitive pattern
- https://ui.shadcn.com/docs/components/alert-dialog - confirmation dialog pattern
- https://ui.shadcn.com/docs/components/radix/sonner - toast primitive pattern

### Secondary (MEDIUM confidence)
- npm registry `npm view` on 2026-03-24 for `next`, `tailwindcss`, `@tanstack/react-query`, `react-hook-form`, `lucide-react` - version currency and modification dates

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: MEDIUM - local stack is clear, but latest upstream currency was only partially registry-verified and no stack upgrade is recommended
- Architecture: HIGH - strongly supported by local codebase, phase context, UI-SPEC, and official Next.js layout/redirect patterns
- Pitfalls: HIGH - directly derived from current repo behavior and locked phase decisions

**Research date:** 2026-03-24
**Valid until:** 2026-04-23
