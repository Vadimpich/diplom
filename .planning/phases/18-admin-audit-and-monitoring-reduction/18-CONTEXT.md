# Phase 18: Admin Audit And Monitoring Reduction - Context

**Gathered:** 2026-03-25
**Status:** Ready for planning
**Mode:** Auto-generated (discuss skipped by direct user instruction)

<domain>
## Phase Boundary

Phase 18 covers:
- `/admin/audit`
- `/admin/monitoring`

Required direction:
- audit becomes a plain log table with filters and expandable rows;
- monitoring becomes a simple status list;
- remove summary cards, action cards, explanatory text blocks, and observability/dashboard framing.

</domain>

<decisions>
## Implementation Decisions

### Read-only surfaces stay read-only
No new admin actions should be introduced in this phase.

### Use current data only
Monitoring must stay honest and use only already available health/readiness/metrics data.

</decisions>

<code_context>
## Existing Code Insights

- `frontend/app/(app)/admin/audit/page.tsx` still includes explanatory header copy and an extra alert block.
- `frontend/components/admin/audit-event-list.tsx` is card-based and not table-based.
- `frontend/components/admin/audit-filter-form.tsx` is wrapped in card chrome and explanatory text.
- `frontend/app/(app)/admin/monitoring/page.tsx` is still a multi-card dashboard with summaries, actions, and explanatory sections.

</code_context>

<specifics>
## Specific Ideas

- keep one compact filter row above audit table;
- make audit details collapsible per row using native `details`/`summary` or a secondary row;
- monitoring should be one compact table plus optional last-updated stamp.

</specifics>

<deferred>
## Deferred Ideas

- user/questionnaire/settings edit-surface simplification in Phase 19.

</deferred>
