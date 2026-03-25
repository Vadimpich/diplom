# Phase 17: Admin Registries As Tables - Context

**Gathered:** 2026-03-25
**Status:** Ready for planning
**Mode:** Auto-generated (discuss skipped by direct user instruction)

<domain>
## Phase Boundary

Phase 17 covers only the admin registries:
- `/admin/users`
- `/admin/questionnaires`

Required direction:
- compact inline filters;
- dense tables instead of card-led surfaces;
- clickable rows instead of `Открыть` buttons;
- no descriptive copy, no usage summaries, no dashboard-style supporting blocks.

</domain>

<decisions>
## Implementation Decisions

### Keep existing data contracts if possible
Use current registry fields already present in frontend types and API responses. Do not add backend work unless a hard blocker appears.

### Compact means fewer columns
Users table should center on `login`, `role`, `status`, `last_login`.
Questionnaires table should center on `name`, `questions_count`, `status`, `updated_at`.

</decisions>

<code_context>
## Existing Code Insights

- `frontend/components/admin/user-list.tsx` already has filters and a table-like structure, but still wraps everything in card chrome, descriptive copy, counts, created/updated metadata and an explicit `Открыть` button.
- `frontend/components/admin/questionnaire-list.tsx` still exposes usage filters, usage counters, description snippets, last-used metadata and an explicit `Открыть` button.
- Page-level `PageHeader` descriptions on `/admin/users` and `/admin/questionnaires` are still verbose and should be shortened or removed.

</code_context>

<specifics>
## Specific Ideas

- keep top action buttons for create flows;
- make each row keyboard and pointer accessible through `router.push`;
- retain empty/error states, but shorten copy and remove ornamental framing where possible.

</specifics>

<deferred>
## Deferred Ideas

- audit and monitoring reduction in Phase 18;
- single-column edit surfaces in Phase 19.

</deferred>
