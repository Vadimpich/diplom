status: passed

# Phase 17 Verification

## Checks

- `cd frontend && npm run lint` — passed
- `cd frontend && npm run build` — passed
- `cd frontend && npx tsc --noEmit` — passed

## Outcome

The admin registries now satisfy the phase rules:
- compact inline filters;
- quiet, dense tables;
- clickable rows instead of `Открыть` actions;
- no summary or usage-oriented dashboard chrome.
