# Phase 06 Validation

## Automated

```bash
cd /home/vadim/diplom/frontend
npm run lint
npm run build
npx tsc --noEmit
npm run test -- --run lib/operator/examination-navigation.test.ts lib/server/core-readiness.test.ts
```

## Manual

1. Open `/operator/history`.
2. Select an examination in `aggregated` or `completed`.
3. Confirm the link lands on `/operator/examinations/{id}/results`.
4. Confirm the result screen shows the same decision, baseline, metrics, and channel-contribution surface as the direct result flow.
5. Force the shared frontend probe down by pointing `NEXT_PUBLIC_API_URL` at an unreachable core-backend or by stopping `core-backend`.
6. Request `/api/metrics` and confirm it contains `diplom_frontend_dependency_up{dependency="core_backend"} 0`.
