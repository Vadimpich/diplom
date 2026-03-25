# Summary 18-01: Reduce admin audit and monitoring to plain operational status surfaces

## Done

- Audit page was reduced to a plain filter row and a compact expandable log table.
- Monitoring page was reduced to one honest status list without summary cards, action cards, or explanatory panels.
- Dashboard framing and developer-facing prose were removed from both read-only admin surfaces.

## Files

- `frontend/app/(app)/admin/audit/page.tsx`
- `frontend/components/admin/audit-filter-form.tsx`
- `frontend/components/admin/audit-event-list.tsx`
- `frontend/app/(app)/admin/monitoring/page.tsx`

## Verification

- `cd frontend && npm run lint`
- `cd frontend && npm run build`
- `cd frontend && npx tsc --noEmit`

## Result

Phase 18 goal is met: audit and monitoring now behave like quiet operational read tools instead of admin dashboards.
