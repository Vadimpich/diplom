# Summary 19-01: Simplify admin edit surfaces into single-column forms and ordered lists

## Done

- User create/edit screens were reduced to one-column forms with compact inline metadata instead of side panels.
- Questionnaire create/edit screens were reduced to one-column forms, and question ordering now supports native drag-and-drop.
- Settings screen was reduced to one vertical form without a secondary side panel or explanatory prose.

## Files

- `frontend/app/(app)/admin/users/new/page.tsx`
- `frontend/app/(app)/admin/users/[id]/page.tsx`
- `frontend/components/admin/user-form.tsx`
- `frontend/app/(app)/admin/questionnaires/new/page.tsx`
- `frontend/app/(app)/admin/questionnaires/[id]/page.tsx`
- `frontend/components/admin/questionnaire-builder.tsx`
- `frontend/app/(app)/admin/settings/page.tsx`

## Verification

- `cd frontend && npm run lint`
- `cd frontend && npm run build`
- `cd frontend && npx tsc --noEmit`

## Result

Phase 19 goal is met: the admin edit surfaces now behave as single-column action-first forms instead of summary-led admin screens.
