# Phase 8: UI Contours & Design Foundation - Context

**Gathered:** 2026-03-24
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 8 separates the shipped frontend into two clearly distinct operator and admin UI contours while establishing one shared design-system foundation. The phase defines the shell structure, navigation model, landing behavior, and baseline visual rules that later phases must reuse; it does not expand ML, decision, or infrastructure scope.

</domain>

<decisions>
## Implementation Decisions

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

### the agent's Discretion
- Exact component names, file/module boundaries, and how the design tokens are organized across CSS variables and reusable components.
- Exact visual density, typography values, and spacing scale, provided they remain consistent with the strict corporate design direction captured above.
- Exact implementation details of sidebar sectioning, hover motion, skeleton patterns, and toast/confirmation primitives.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Product and architecture
- `docs/00_project.md` §4.1 — operator responsibilities and expectations for the operator workstation.
- `docs/00_project.md` §4.2 — administrator responsibilities and expected administrative capabilities.
- `docs/00_project.md` §7.1 — frontend stack, operator/admin contour split, and UI architecture constraints.

### Milestone and phase scope
- `.planning/PROJECT.md` — milestone `v1.1 UI & Admin Completion` goals and frontend-first constraints.
- `.planning/REQUIREMENTS.md` — `CNTR-01`, `CNTR-02`, `DSGN-01` and milestone-wide UX/design expectations.
- `.planning/ROADMAP.md` — Phase 8 goal, dependency chain, and success criteria.
- `.planning/STATE.md` — active milestone position and current planning focus.

### UX quality baseline
- `.planning/phases/07-verification-evidence-and-requirement-revalidation/07-UI-REVIEW.md` — baseline audit findings that motivate the stricter design foundation and contour separation.

### Contracts that affect contour behavior
- `docs/01_contract.md` — authoritative status vocabulary, runtime health/readiness/metrics surfaces, and role-sensitive operator/admin data presentation constraints.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `frontend/components/layout/app-shell.tsx`: current shared shell component that already centralizes sidebar, user session panel, and logout action; likely base for contour-specific shells or a shared shell primitive.
- `frontend/components/layout/route-guard.tsx`: existing role gate and session-state handling that Phase 8 must preserve while improving contour separation.
- `frontend/components/ui/page-header.tsx`: current shared page heading primitive that can become part of the canonical design foundation.
- `frontend/components/ui/*`: existing button/input/card/alert/badge/spinner primitives that can be normalized instead of replaced wholesale.
- `frontend/app/globals.css`: current token source and shell background patterns; natural place to tighten color, spacing, radius, and typography rules.

### Established Patterns
- The app already uses Next.js App Router route groups with separate `frontend/app/(app)/operator/*` and `frontend/app/(app)/admin/*` layouts.
- Both current contours use the same `AppShell` with different titles, subtitles, and nav items, so Phase 8 is an evolution of an existing pattern rather than a greenfield redesign.
- Frontend data/state work is already contract-first and centered on shared client utilities, so contour work should stay presentation- and composition-focused.

### Integration Points
- `frontend/app/(app)/operator/layout.tsx`: current operator shell entry point and role-gated layout.
- `frontend/app/(app)/admin/layout.tsx`: current admin shell entry point and role-gated layout.
- `frontend/app/page.tsx`: current role-based landing redirect that must be updated for the admin home decision.
- `frontend/app/(auth)/login/page.tsx`: sets the visual tone of entry into the application and should stay aligned with the new contour foundation.

</code_context>

<specifics>
## Specific Ideas

- The target look is strict, professional, and corporate rather than decorative or playful.
- Modern minimalism is acceptable only when readability and data interpretation improve.
- Shared UX primitives must include skeletons, alerts, toasts, confirmations, and explicit action feedback.
- Interface consistency across operator and admin matters more than adding more unique one-off components.
- The admin contour should have its own true landing page rather than inheriting `/admin/users` as a default just because it exists today.

</specifics>

<deferred>
## Deferred Ideas

- Detailed operator results readability fixes from the Phase 07 UI review are executed in Phase 11, not in Phase 8.
- Admin CRUD completion, system settings, monitoring, and audit log delivery belong to Phases 9 and 10 after the contour foundation is in place.
- Advanced charts and richer result visualizations remain future work under `RICH-01`.

</deferred>

---

*Phase: 08-ui-contours-design-foundation*
*Context gathered: 2026-03-24*
