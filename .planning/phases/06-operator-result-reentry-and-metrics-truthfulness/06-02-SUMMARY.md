---
phase: 06-operator-result-reentry-and-metrics-truthfulness
plan: 02
subsystem: ui
tags: [frontend, observability, readiness, metrics, prometheus]
requires:
  - phase: 05-operational-trustworthiness
    provides: Frontend `/api/ready` and `/api/metrics` surfaces
provides:
  - Shared core readiness probe helper
  - Frontend readiness route backed by the shared probe
  - Truthful frontend dependency gauge in Prometheus output
affects: [observability, frontend-runtime, monitoring]
tech-stack:
  added: []
  patterns: [shared probe helper for multiple runtime routes, low-cardinality dependency metrics]
key-files:
  created:
    - frontend/lib/server/core-readiness.ts
  modified:
    - frontend/app/api/ready/route.ts
    - frontend/app/api/metrics/route.ts
    - frontend/lib/api/types.ts
key-decisions:
  - "Frontend `/api/ready` and `/api/metrics` now derive `core_backend` truth from the same live probe."
  - "The dependency gauge emits `0` or `1` only from the shared probe result and does not hardcode success."
patterns-established:
  - "Server-only probe logic lives under `frontend/lib/server/*` and is reused by multiple route handlers."
requirements-completed: [OBSV-02]
duration: 15min
completed: 2026-03-23
---

# 06-02 Summary

## Completed

- Added `frontend/lib/server/core-readiness.ts` as one shared probe for `core-backend` readiness.
- Refactored `frontend/app/api/ready/route.ts` to use the shared probe and return the existing contract through a typed response shape.
- Replaced the static dependency gauge in `frontend/app/api/metrics/route.ts` with a live `0/1` value derived from the shared readiness probe.

## Verification

- `cd /home/vadim/diplom/frontend && npm run lint`
- `cd /home/vadim/diplom/frontend && npm run build`
- `cd /home/vadim/diplom/frontend && npx tsc --noEmit`
- `cd /home/vadim/diplom/frontend && rg -n 'probeCoreBackendReadiness|core_backend|up|down' app/api/ready/route.ts lib/server/core-readiness.ts lib/api/types.ts`
- `cd /home/vadim/diplom/frontend && rg -n 'probeCoreBackendReadiness|diplom_frontend_dependency_up|text/plain; version=0.0.4' app/api/metrics/route.ts lib/server/core-readiness.ts`

## Result

Frontend observability no longer invents dependency health: readiness and metrics now expose the same backend-derived truth for `core_backend`.
