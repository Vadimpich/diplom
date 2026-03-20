---
phase: 01-trusted-access-and-intake
plan: 04
title: Examination Snapshot Persistence
type: execute
wave: 3
depends_on:
  - 02-PLAN.md
wave_reason: operator intake routes should already be role-scoped before expanding examination creation semantics
files_modified:
  - /home/katya/dimplom/core-backend/migrations/000004_examination_question_snapshots.up.sql
  - /home/katya/dimplom/core-backend/migrations/000004_examination_question_snapshots.down.sql
  - /home/katya/dimplom/core-backend/db/queries/examinations.sql
  - /home/katya/dimplom/core-backend/db/sqlc/examinations.sql.go
  - /home/katya/dimplom/core-backend/db/sqlc/models.go
  - /home/katya/dimplom/core-backend/internal/examinations/service.go
  - /home/katya/dimplom/core-backend/internal/examinations/repository.go
  - /home/katya/dimplom/core-backend/internal/http/examinations_handler.go
  - /home/katya/dimplom/core-backend/internal/examinations/service_test.go
  - /home/katya/dimplom/docs/01_contract.md
  - /home/katya/dimplom/docs/02_implementation.md
autonomous: true
requirements_addressed:
  - EXAM-01
must_haves:
  truths:
    - "Creating an examination in Phase 1 requires `questionnaire_id` and snapshots the ordered question set for that examination."
  artifacts:
    - path: /home/katya/dimplom/core-backend/migrations/000004_examination_question_snapshots.up.sql
      provides: examination_questions schema
    - path: /home/katya/dimplom/core-backend/internal/examinations/repository.go
      provides: transactional exam creation with question snapshots
  key_links:
    - from: /home/katya/dimplom/core-backend/internal/http/examinations_handler.go
      to: /home/katya/dimplom/core-backend/internal/examinations/service.go
      via: POST /examinations
---

# Objective

Snapshot questionnaire questions at examination creation so intake is anchored to immutable examination-scoped questions.

<tasks>

<task id="1-04-01" type="auto">
  <name>Task 1: Add examination question snapshot schema and creation flow</name>
  <files>
    /home/katya/dimplom/core-backend/migrations/000004_examination_question_snapshots.up.sql
    /home/katya/dimplom/core-backend/migrations/000004_examination_question_snapshots.down.sql
    /home/katya/dimplom/core-backend/db/queries/examinations.sql
    /home/katya/dimplom/core-backend/db/sqlc/examinations.sql.go
    /home/katya/dimplom/core-backend/db/sqlc/models.go
    /home/katya/dimplom/core-backend/internal/examinations/service.go
    /home/katya/dimplom/core-backend/internal/examinations/repository.go
    /home/katya/dimplom/core-backend/internal/http/examinations_handler.go
    /home/katya/dimplom/core-backend/internal/examinations/service_test.go
  </files>
  <read_first>
    /home/katya/dimplom/AGENTS.md
    /home/katya/dimplom/docs/00_project.md
    /home/katya/dimplom/docs/01_contract.md
    /home/katya/dimplom/.planning/phases/01-trusted-access-and-intake/01-RESEARCH.md
  </read_first>
  <action>
    Add migration `000004_examination_question_snapshots` creating `examination_questions` with exact columns `id BIGSERIAL PRIMARY KEY`, `examination_id BIGINT NOT NULL REFERENCES examinations(id) ON DELETE CASCADE`, `specialist_id BIGINT NOT NULL REFERENCES specialists(id) ON DELETE RESTRICT`, `questionnaire_id BIGINT NOT NULL REFERENCES questionnaires(id) ON DELETE RESTRICT`, `source_question_id BIGINT REFERENCES questions(id) ON DELETE SET NULL`, `position INTEGER NOT NULL CHECK (position > 0)`, and `question_text TEXT NOT NULL`, plus `UNIQUE (examination_id, position)`. Extend examination SQL and repository code so `POST /examinations` requires non-null `questionnaire_id` for Phase 1, rejects missing questionnaire assignments with a deterministic validation error, creates the examination, and inserts snapshot rows from `questionnaire_questions` in one transaction. Add Wave 0 test coverage in `internal/examinations/service_test.go` for snapshot creation, missing-questionnaire rejection, and invalid questionnaire input.
  </action>
  <verify>
    <automated>cd /home/katya/dimplom/core-backend && go test ./internal/examinations -run 'TestCreateExaminationSnapshotsQuestions|TestCreateExaminationRejectsInvalidQuestionnaire' -count=1</automated>
  </verify>
  <acceptance_criteria>
    `rg -n "CREATE TABLE examination_questions|UNIQUE \\(examination_id, position\\)" /home/katya/dimplom/core-backend/migrations/000004_examination_question_snapshots.up.sql`
  </acceptance_criteria>
  <done>
    Examination creation requires `questionnaire_id`, snapshots questions transactionally, related tests pass, and the implementation no longer relies on mutable questionnaire rows to represent historical questions.
  </done>
</task>

<task id="1-04-02" type="auto">
  <name>Task 2: Document examination snapshot contract changes</name>
  <files>
    /home/katya/dimplom/docs/01_contract.md
    /home/katya/dimplom/docs/02_implementation.md
  </files>
  <read_first>
    /home/katya/dimplom/AGENTS.md
    /home/katya/dimplom/docs/01_contract.md
    /home/katya/dimplom/docs/02_implementation.md
  </read_first>
  <action>
    Update `docs/01_contract.md` with the examination creation contract changes caused by question snapshotting, explicitly marking `questionnaire_id` as required for Phase 1 creation flow, and append the implementation note to `docs/02_implementation.md`. Treat the documentation duties required by `AGENTS.md` as part of completion.
  </action>
  <verify>
    <automated>cd /home/katya/dimplom/core-backend && go test ./internal/examinations -run 'TestCreateExaminationSnapshotsQuestions|TestCreateExaminationRejectsInvalidQuestionnaire' -count=1</automated>
  </verify>
  <acceptance_criteria>
    `rg -n "examination_questions|POST /examinations" /home/katya/dimplom/docs/01_contract.md`
  </acceptance_criteria>
  <done>
    Snapshot-related contract and implementation updates are recorded in `docs/01_contract.md` and `docs/02_implementation.md` per `AGENTS.md`.
  </done>
</task>

</tasks>
