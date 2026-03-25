# Summary 17-01: Rebuild admin users and questionnaires into compact table-first registries

## Done

- `/admin/users` reduced to a compact registry with inline filters and a dense clickable table.
- `/admin/questionnaires` reduced to a compact registry with inline filters and a dense clickable table.
- Explicit `Открыть` buttons, usage-focused metadata, summary counts, and descriptive registry copy were removed from both surfaces.

## Files

- `frontend/app/(app)/admin/users/page.tsx`
- `frontend/app/(app)/admin/questionnaires/page.tsx`
- `frontend/components/admin/user-list.tsx`
- `frontend/components/admin/questionnaire-list.tsx`

## Verification

- `cd frontend && npm run lint`
- `cd frontend && npm run build`
- `cd frontend && npx tsc --noEmit`

## Result

Phase 17 goal is met: users and questionnaires now behave like compact operational tables instead of card-led admin screens.
