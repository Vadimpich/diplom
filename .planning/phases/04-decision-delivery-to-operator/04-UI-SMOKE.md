# Phase 4 UI Smoke

## Manual Probe

1. Start the application stack needed for the operator result page.
2. Open `/operator/examinations/{id}/results`.
3. Verify the decision card is rendered above the existing metrics section.
4. Verify the page shows `Не реализовано` and does not show synthetic `Допуск`, `Риск`, or `Недопуск`.
5. Verify the page still shows metrics, baseline, explanations, and channel contributions below the decision card.

## Expected Result

- The decision card remains honest about the missing model output.
- Existing metrics and channel contributions remain visible together with the Phase 4 diagnostics area.

