# Summary 16-01: Simplify admin shell chrome and remove the admin dashboard layer

## Done

- Admin shell navigation was reduced to label-only links without overview grouping or descriptive subtitles.
- `/admin` was turned into a redirect to `/admin/users`, removing the dashboard entry surface entirely.
- Admin loading state was rebuilt as a quiet registry-like skeleton instead of dashboard cards.

## Files

- `frontend/components/layout/app-shell.tsx`
- `frontend/components/layout/admin-shell.tsx`
- `frontend/app/(app)/admin/page.tsx`
- `frontend/app/(app)/admin/loading.tsx`

## Verification

- `cd frontend && npm run lint`
- `cd frontend && npm run build`
- `cd frontend && npx tsc --noEmit`

## Result

Phase 16 goal is met: the admin contour now enters through a quiet working shell and no longer exposes a dashboard-style `/admin` home.
