# Phase 16: Admin Shell And Home Simplification - Context

**Gathered:** 2026-03-25
**Status:** Ready for planning
**Mode:** Auto-generated (discuss skipped by direct user instruction)

<domain>
## Phase Boundary

Phase 16 covers only the admin shell chrome and the `/admin` entry surface.

Required direction:
- no dashboard on `/admin`;
- no explanatory copy in the admin shell;
- no descriptive sublabels in navigation;
- fixed-height sidebar with logout pinned to the bottom;
- only navigation and actions that help the administrator reach working screens.

Out of phase:
- users/questionnaires registry restructuring;
- audit and monitoring reduction;
- user/questionnaire/settings edit-surface simplification.

</domain>

<decisions>
## Implementation Decisions

### Separate milestone, strict scope
This work belongs to `v1.3` and must not re-open operator surfaces from `v1.2`.

### Minimal entry model
`/admin` may become a quiet redirect or a minimal navigation surface, but not a summary dashboard.

### No contract work unless necessary
Phase 16 should stay frontend-only unless a blocker is found.

</decisions>

<code_context>
## Existing Code Insights

- `frontend/components/layout/app-shell.tsx` still renders admin subtitle text, section titles, bordered nav blocks, and a session card that add non-essential noise.
- `frontend/components/layout/admin-shell.tsx` still models the admin contour as an overview-led shell with grouped sections.
- `frontend/app/(app)/admin/page.tsx` is still a dashboard entry that mounts `AdminSummaryStrip`.
- `frontend/components/admin/admin-summary-strip.tsx` contains the current dashboard layer and is a likely removal target for this phase.

</code_context>

<specifics>
## Specific Ideas

- compress the admin shell to label-only nav;
- remove nav section titles if they do not improve wayfinding;
- keep one header row in content and avoid any card grid on `/admin`;
- prefer a direct redirect from `/admin` to the most actionable registry if that yields the quietest UX.

</specifics>

<deferred>
## Deferred Ideas

- table-first user and questionnaire registries in Phase 17;
- audit and monitoring structural reduction in Phase 18;
- single-column edit surfaces in Phase 19.

</deferred>
