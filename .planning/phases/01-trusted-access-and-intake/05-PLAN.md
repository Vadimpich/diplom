---
phase: 01-trusted-access-and-intake
plan: 05
title: Answer Linkage And Intake Contracts
type: execute
wave: 4
depends_on:
  - 04-PLAN.md
files_modified:
  - /home/katya/dimplom/core-backend/db/queries/answers.sql
  - /home/katya/dimplom/core-backend/db/sqlc/answers.sql.go
  - /home/katya/dimplom/core-backend/db/sqlc/models.go
  - /home/katya/dimplom/core-backend/internal/answers/service.go
  - /home/katya/dimplom/core-backend/internal/answers/repository.go
  - /home/katya/dimplom/core-backend/internal/http/answers_handler.go
  - /home/katya/dimplom/core-backend/internal/answers/service_test.go
  - /home/katya/dimplom/docs/01_contract.md
  - /home/katya/dimplom/docs/02_implementation.md
autonomous: true
requirements_addressed:
  - EXAM-02
must_haves:
  truths:
    - "Each uploaded answer is linked to examination, examination question, and specialist."
    - "The backend rejects answers that reference a question outside the examination snapshot."
  artifacts:
    - path: /home/katya/dimplom/core-backend/internal/answers/service.go
      provides: answer linkage validation
    - path: /home/katya/dimplom/docs/01_contract.md
      provides: intake payload contracts
  key_links:
    - from: /home/katya/dimplom/core-backend/internal/http/answers_handler.go
      to: /home/katya/dimplom/core-backend/internal/answers/service.go
      via: POST /answers
---

# Objective

Persist answer uploads with stable question and specialist linkage based on the examination snapshot.

<tasks>

<task id="1-05-01" type="auto">
  <name>Task 1: Persist answer linkage and validate snapshot ownership</name>
  <files>
    /home/katya/dimplom/core-backend/db/queries/answers.sql
    /home/katya/dimplom/core-backend/db/sqlc/answers.sql.go
    /home/katya/dimplom/core-backend/db/sqlc/models.go
    /home/katya/dimplom/core-backend/internal/answers/service.go
    /home/katya/dimplom/core-backend/internal/answers/repository.go
    /home/katya/dimplom/core-backend/internal/http/answers_handler.go
    /home/katya/dimplom/core-backend/internal/answers/service_test.go
  </files>
  <read_first>
    /home/katya/dimplom/AGENTS.md
    /home/katya/dimplom/core-backend/internal/answers/service.go
    /home/katya/dimplom/core-backend/internal/http/answers_handler.go
    /home/katya/dimplom/core-backend/db/queries/answers.sql
  </read_first>
  <action>
    Extend the answer write path so `POST /answers` requires `examination_id`, `examination_question_id`, `specialist_id`, `text`, and audio upload data, and persists `answers.examination_question_id` plus `answers.specialist_id` with validation that the referenced `examination_question_id` belongs to the same examination and specialist. Add a uniqueness fence such as `UNIQUE (examination_id, examination_question_id)` if one answer per assigned question is required. Create or extend `internal/answers/service_test.go` for success, mismatched question rejection, and duplicate answer rejection.
  </action>
  <verify>
    <automated>cd /home/katya/dimplom/core-backend && go test ./internal/answers -run 'TestCreateAnswerPersistsQuestionLink|TestCreateAnswerRejectsMismatchedQuestion|TestCreateAnswerRejectsDuplicateQuestionAnswer' -count=1</automated>
  </verify>
  <acceptance_criteria>
    `rg -n "examination_question_id|specialist_id" /home/katya/dimplom/core-backend/db/queries/answers.sql /home/katya/dimplom/core-backend/internal/http/answers_handler.go`
  </acceptance_criteria>
  <done>
    Answer persistence includes examination/question/specialist linkage and validation tests pass.
  </done>
</task>

<task id="1-05-02" type="auto">
  <name>Task 2: Document answer-linkage contracts and implementation state</name>
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
    Update `docs/01_contract.md` to reflect the exact `POST /answers` payload and linkage semantics, then append the implementation note to `docs/02_implementation.md`. Treat the documentation duties required by `AGENTS.md` as part of completion.
  </action>
  <verify>
    <automated>cd /home/katya/dimplom/core-backend && go test ./internal/answers -run 'TestCreateAnswerPersistsQuestionLink|TestCreateAnswerRejectsMismatchedQuestion|TestCreateAnswerRejectsDuplicateQuestionAnswer' -count=1</automated>
  </verify>
  <acceptance_criteria>
    `rg -n "POST /answers|examination_question_id|specialist_id" /home/katya/dimplom/docs/01_contract.md`
  </acceptance_criteria>
  <done>
    Answer-linkage contract and implementation updates are recorded in `docs/01_contract.md` and `docs/02_implementation.md` per `AGENTS.md`.
  </done>
</task>

</tasks>
