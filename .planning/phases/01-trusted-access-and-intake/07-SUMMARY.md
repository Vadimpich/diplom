---
phase: 01-trusted-access-and-intake
plan: 07
subsystem: ui
tags: [nextjs, operator, statuses]
requirements-completed: [EXAM-04]
completed: 2026-03-20
---

# Phase 1 Plan 07 Summary

Specialist history UI now renders authoritative backend workflow statuses.

## Task Commits

1. `90a793a` `feat(01-07): show specialist history from backend statuses`

## Outcome

- Specialist detail history uses `GET /specialists/{id}/examinations`.
- Status badges render `created`, `collecting_answers`, `ready_for_processing`.
- Backend HTTP tests verify status history stays authoritative.
