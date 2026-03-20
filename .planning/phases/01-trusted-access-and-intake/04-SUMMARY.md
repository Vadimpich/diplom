---
phase: 01-trusted-access-and-intake
plan: 04
subsystem: examinations
tags: [go, postgres, snapshots]
requirements-completed: [EXAM-01]
completed: 2026-03-20
---

# Phase 1 Plan 04 Summary

Examination creation now snapshots questionnaire questions into immutable examination-scoped rows.

## Task Commits

1. `fc8bd78` `feat(01-04): snapshot questionnaire questions per examination`

## Outcome

- Added `examination_questions`.
- `POST /examinations` now requires `questionnaire_id`.
- Snapshot rows are created transactionally on examination creation.
