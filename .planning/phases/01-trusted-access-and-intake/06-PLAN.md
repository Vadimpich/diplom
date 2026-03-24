---
phase: 01-trusted-access-and-intake
plan: 06
title: Backend Finish Idempotency Fence
type: execute
wave: 5
depends_on:
  - 05-PLAN.md
files_modified:
  - /home/vadim/diplom/core-backend/migrations/000005_examination_finish_fence.up.sql
  - /home/vadim/diplom/core-backend/migrations/000005_examination_finish_fence.down.sql
  - /home/vadim/diplom/core-backend/db/queries/examinations.sql
  - /home/vadim/diplom/core-backend/db/sqlc/examinations.sql.go
  - /home/vadim/diplom/core-backend/internal/examinations/service.go
  - /home/vadim/diplom/core-backend/internal/examinations/repository.go
  - /home/vadim/diplom/core-backend/internal/http/examinations_handler.go
  - /home/vadim/diplom/core-backend/internal/examinations/service_test.go
  - /home/vadim/diplom/core-backend/internal/http/examinations_handler_test.go
  - /home/vadim/diplom/docs/01_contract.md
  - /home/vadim/diplom/docs/02_implementation.md
autonomous: true
requirements_addressed:
  - EXAM-03
must_haves:
  truths:
    - "Finishing an examination transitions it to ready_for_processing only once even if finish is retried."
  artifacts:
    - path: /home/vadim/diplom/core-backend/migrations/000005_examination_finish_fence.up.sql
      provides: processing launch dedupe fence
    - path: /home/vadim/diplom/core-backend/internal/examinations/service.go
      provides: transactional finish logic
  key_links:
    - from: /home/vadim/diplom/core-backend/internal/http/examinations_handler.go
      to: /home/vadim/diplom/core-backend/internal/examinations/service.go
      via: POST /examinations/{id}/finish
---

# Objective

Guarantee that finishing an examination is transactional and idempotent at the database layer.

<tasks>

<task id="1-06-01" type="auto">
  <name>Task 1: Add transactional finish fence and idempotency checks</name>
  <files>
    /home/vadim/diplom/core-backend/migrations/000005_examination_finish_fence.up.sql
    /home/vadim/diplom/core-backend/migrations/000005_examination_finish_fence.down.sql
    /home/vadim/diplom/core-backend/db/queries/examinations.sql
    /home/vadim/diplom/core-backend/db/sqlc/examinations.sql.go
    /home/vadim/diplom/core-backend/internal/examinations/service.go
    /home/vadim/diplom/core-backend/internal/examinations/repository.go
    /home/vadim/diplom/core-backend/internal/http/examinations_handler.go
    /home/vadim/diplom/core-backend/internal/examinations/service_test.go
    /home/vadim/diplom/core-backend/internal/http/examinations_handler_test.go
  </files>
  <read_first>
    /home/vadim/diplom/AGENTS.md
    /home/vadim/diplom/docs/00_project.md
    /home/vadim/diplom/docs/01_contract.md
    /home/vadim/diplom/.planning/phases/01-trusted-access-and-intake/01-RESEARCH.md
  </read_first>
  <action>
    Add migration `000005_examination_finish_fence` creating `examination_processing_launches (examination_id BIGINT PRIMARY KEY REFERENCES examinations(id) ON DELETE CASCADE, launched_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`. Extend examination SQL and repository code with `GetExaminationForUpdate`, `CountAnswersForExamination`, `CountExaminationQuestions`, `CreateExaminationProcessingLaunch`, and a finish update that uses `COALESCE(finished_at, NOW())`. Update `internal/examinations/service.go` and repository transaction logic so `POST /examinations/{id}/finish` locks the examination row, verifies `collecting_answers`, verifies answer completeness against `examination_questions`, inserts the launch fence with `ON CONFLICT DO NOTHING`, and updates status to `ready_for_processing` only inside the same transaction. Add or extend Wave 0 tests for missing answers and repeated finish.
  </action>
  <verify>
    <automated>cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/examinations -run 'TestFinishIsIdempotent|TestFinishRequiresAllAnswers' -count=1</automated>
  </verify>
  <acceptance_criteria>
    `rg -n "CREATE TABLE examination_processing_launches|ON CONFLICT DO NOTHING|FOR UPDATE" /home/vadim/diplom/core-backend/migrations/000005_examination_finish_fence.up.sql /home/vadim/diplom/core-backend/db/queries/examinations.sql`
  </acceptance_criteria>
  <done>
    Transactional finish fence exists, idempotency and completeness tests pass, and backend finish logic prevents duplicate processing launches.
  </done>
</task>

<task id="1-06-02" type="auto">
  <name>Task 2: Document finish-idempotency contracts and implementation state</name>
  <files>
    /home/vadim/diplom/docs/01_contract.md
    /home/vadim/diplom/docs/02_implementation.md
  </files>
  <read_first>
    /home/vadim/diplom/AGENTS.md
    /home/vadim/diplom/docs/01_contract.md
    /home/vadim/diplom/docs/02_implementation.md
  </read_first>
  <action>
    Update `docs/01_contract.md` with exact finish idempotency semantics for `POST /examinations/{id}/finish`, then append the implementation note to `docs/02_implementation.md`. Treat the documentation duties required by `AGENTS.md` as part of completion.
  </action>
  <verify>
    <automated>cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/examinations -run 'TestFinishIsIdempotent|TestFinishRequiresAllAnswers' -count=1</automated>
  </verify>
  <acceptance_criteria>
    `rg -n "idempotent|ready_for_processing|POST /examinations/\\{id\\}/finish" /home/vadim/diplom/docs/01_contract.md`
  </acceptance_criteria>
  <done>
    Finish-idempotency contract and implementation updates are recorded in `docs/01_contract.md` and `docs/02_implementation.md` per `AGENTS.md`.
  </done>
</task>

</tasks>
