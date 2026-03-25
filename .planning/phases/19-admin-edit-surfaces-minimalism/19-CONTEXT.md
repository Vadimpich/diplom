# Phase 19: Admin Edit Surfaces Minimalism - Context

**Gathered:** 2026-03-25
**Status:** Ready for planning
**Mode:** Auto-generated (discuss skipped by direct user instruction)

<domain>
## Phase Boundary

Phase 19 covers:
- user create/edit
- questionnaire create/edit
- settings

Required direction:
- single-column forms;
- no right-side summary panels;
- no technical helper prose in the UI;
- questionnaire ordering through drag-and-drop;
- short, operational copy only.

</domain>

<decisions>
## Implementation Decisions

### Bounded to frontend
No contract changes are expected; use current admin APIs and fields.

### Drag-and-drop can be native
Avoid introducing a new dependency if native browser drag-and-drop is enough for ordered questions.

</decisions>

<code_context>
## Existing Code Insights

- User edit and questionnaire edit pages still render two-column layouts with secondary info cards.
- `QuestionnaireBuilder` still relies on noisy `Вверх`/`Вниз` controls and explanatory text.
- Settings page still uses a two-column layout plus section descriptions and a side panel.

</code_context>

<specifics>
## Specific Ideas

- keep timestamps as a compact inline metadata row instead of a side card;
- use one visible save action at the end of each form;
- retain validation, alerts, confirm-dialogs, and unsaved-changes warning if they already help action flow.

</specifics>

<deferred>
## Deferred Ideas

- milestone-level verification and archive workflow after Phase 19.

</deferred>
