---
phase: 02-asynchronous-multichannel-processing
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - docs/01_contract.md
  - core-backend/migrations/000006_processing_pipeline.up.sql
  - core-backend/migrations/000006_processing_pipeline.down.sql
  - core-backend/db/queries/processing.sql
  - core-backend/internal/processing/contracts.go
  - core-backend/internal/processing/service_test.go
  - core-backend/internal/http/processing_status_test.go
autonomous: true
requirements:
  - PIPE-01
must_haves:
  truths:
    - Finish-path persistence is defined before any broker code ships.
    - Command, result, and progress DTO contracts are versioned in one documented place.
    - Backend tests describe the mandatory channel fan-out and progress surface before implementation.
  artifacts:
    - docs/01_contract.md documents versioned envelopes and the processing-status HTTP DTO.
    - core-backend/migrations/000006_processing_pipeline.up.sql creates outbox and per-channel runtime tables.
    - core-backend/internal/processing/contracts.go exports the mandatory channel names and envelope types.
  key_links:
    - POST /examinations/{id}/finish must create outbox rows and channel-run rows in the same transaction.
    - The progress endpoint must project PostgreSQL state, not RabbitMQ state.
---

<objective>
Publish the Phase 2 contracts and persistence skeleton that every later plan will implement against.

Purpose: remove contract drift and dual-write ambiguity before broker or worker code lands.
Output: documented message/API contracts, DB schema for outbox and channel runs, and failing tests that pin the behavior.
</objective>

<execution_context>
@/home/vadim/.codex/get-shit-done/workflows/execute-plan.md
@/home/vadim/.codex/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/02-asynchronous-multichannel-processing/02-RESEARCH.md
@.planning/phases/02-asynchronous-multichannel-processing/02-VALIDATION.md
@docs/00_project.md
@docs/01_contract.md
@docs/02_implementation.md
@core-backend/internal/examinations/service.go
@core-backend/internal/http/examinations_handler.go
@core-backend/internal/http/router.go

<interfaces>
From core-backend/internal/examinations/service.go:
```go
const (
	StatusCreated            = "created"
	StatusCollectingAnswers  = "collecting_answers"
	StatusReadyForProcessing = "ready_for_processing"
)

func (s *Service) Finish(ctx context.Context, id int64) (Examination, error)
```

From docs/00_project.md and research:
```text
mandatory channels = text, acoustic, paralinguistic
transport payloads must reference S3 objects, never inline audio bytes
PostgreSQL remains the source of truth for progress and failures
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Lock Phase 2 contracts and test expectations</name>
  <files>docs/01_contract.md, core-backend/internal/processing/contracts.go, core-backend/internal/processing/service_test.go, core-backend/internal/http/processing_status_test.go</files>
  <behavior>
    - Test 1: finishing an examination creates exactly one pending command per mandatory channel with message_version and audio S3 references.
    - Test 2: processing status DTO returns pipeline state plus per-channel status, attempt_count, max_attempts, timestamps, and last error fields.
    - Test 3: result ingestion is channel-neutral: the same envelope shape is accepted for text, acoustic, and paralinguistic.
  </behavior>
  <action>Update `docs/01_contract.md` first with one versioned processing command envelope, one versioned channel result envelope, explicit shared AMQP topology (`processing.commands`, `processing.results`, per-channel routing keys and queue names, unified result routing, and RabbitMQ 3.13 retry/DLX assumptions), and one backend-authoritative `GET /examinations/{id}/processing-status` response. In the same contract update, explicitly broaden examination runtime status vocabulary to include `processing` and `failed` while reserving rich per-channel detail for the dedicated processing-status DTO. Create `core-backend/internal/processing/contracts.go` with exported channel constants and envelope structs mirroring the docs exactly. Add failing tests in `core-backend/internal/processing/service_test.go` and `core-backend/internal/http/processing_status_test.go` that pin the documented shapes, status vocabulary, and mandatory-channel fan-out. Do not introduce websocket contracts or queue-derived UI state.</action>
  <verify>
    <automated>cd /home/vadim/diplom/core-backend && go test ./internal/processing ./internal/http -run 'TestFinishCreatesOutboxForMandatoryChannels|TestProcessingStatusEndpoint' -count=1</automated>
  </verify>
  <done>`docs/01_contract.md` is the single source of truth for Phase 2 envelopes, topology, runtime statuses, and DTOs, and the new tests fail until implementation is added in later plans.</done>
</task>

<task type="auto">
  <name>Task 2: Add the PostgreSQL schema for outbox and per-channel runtime state</name>
  <files>core-backend/migrations/000006_processing_pipeline.up.sql, core-backend/migrations/000006_processing_pipeline.down.sql, core-backend/db/queries/processing.sql</files>
  <action>Create a migration that adds durable processing tables needed by the research-backed design: processing outbox rows, per-examination per-channel run rows, and normalized channel result rows. Include explicit fields for `channel`, `status`, `attempt_count`, `max_attempts`, `message_version`, `last_error_code`, `last_error_message`, broker identifiers, and timestamps required by the progress DTO. Add SQL queries for inserting channel runs/outbox rows atomically and reading progress projections. Preserve Phase 1 finish idempotency by extending the existing launch fence instead of replacing it.</action>
  <verify>
    <automated>cd /home/vadim/diplom/core-backend && go test ./internal/processing ./internal/http -run 'TestFinishCreatesOutboxForMandatoryChannels|TestProcessingStatusEndpoint' -count=1</automated>
  </verify>
  <done>The DB now has a durable schema for command publication intent and per-channel progress that later plans can implement without revisiting the contract.</done>
</task>

</tasks>

<verification>
Run the new targeted tests and confirm they fail only because publisher/result-consumer implementation is not present yet, not because contracts or SQL are inconsistent.
</verification>

<success_criteria>
- `docs/01_contract.md` describes the Phase 2 HTTP and RabbitMQ contracts in versioned form.
- The shared contract also fixes the examination runtime status decision for Phase 2: `processing` and `failed` are valid examination states, while per-channel detail lives under `/processing-status`.
- The migration creates all durable tables required for outbox publishing and channel progress.
- New tests define the finish fan-out and progress DTO behavior before implementation starts.
</success_criteria>

<output>
After completion, create `.planning/phases/02-asynchronous-multichannel-processing/02-01-SUMMARY.md`
</output>
