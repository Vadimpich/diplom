---
phase: 02-asynchronous-multichannel-processing
plan: 04
type: execute
wave: 3
depends_on:
  - 02-02
  - 02-03
files_modified:
  - core-backend/internal/app/app.go
  - core-backend/internal/http/router.go
  - core-backend/internal/http/examinations_handler.go
  - core-backend/internal/http/processing_status_test.go
  - core-backend/internal/processing/results_consumer.go
  - core-backend/internal/processing/service.go
  - core-backend/internal/processing/repository.go
  - core-backend/internal/channelresults/service.go
  - core-backend/internal/channelresults/service_test.go
autonomous: true
requirements:
  - PIPE-03
  - PIPE-04
  - RSLT-01
must_haves:
  truths:
    - Channel completion and failure are persisted independently of one another.
    - Retry exhaustion on any mandatory channel drives the examination into a final error state.
    - The operator-facing progress endpoint reports per-channel runtime directly from backend state.
  artifacts:
    - core-backend/internal/processing/results_consumer.go ingests unified result messages from RabbitMQ.
    - core-backend/internal/channelresults/service.go owns result classification and examination state transitions.
    - core-backend/internal/http/examinations_handler.go exposes a processing-status endpoint.
  key_links:
    - Result messages update PostgreSQL channel-run rows before any UI status is exposed.
    - An exhausted mandatory channel transitions the examination into a terminal pipeline error visible through HTTP.
---

<objective>
Consume channel results back into core, persist retry/error state, and expose backend-authoritative progress.

Purpose: close the async loop so independent workers update one authoritative runtime ledger that the UI can poll.
Output: result consumer, channel-run state machine, examination terminal-error projection, and progress endpoint.
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
@core-backend/internal/http/router.go
@core-backend/internal/http/examinations_handler.go
@core-backend/internal/app/app.go

<interfaces>
From frontend/lib/api/types.ts:
```ts
export interface Examination {
  id: number;
  specialist_id: number;
  created_by_user_id: number;
  questionnaire_id: number | null;
  status: "created" | "collecting_answers" | "ready_for_processing";
}
```

Required extension in this plan:
```text
Persist and document examination-level `processing` / `failed` statuses, but keep rich per-channel runtime detail in a dedicated processing-status DTO instead of overloading the existing examination list item.
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Persist unified channel results and bounded retry state</name>
  <files>core-backend/internal/processing/results_consumer.go, core-backend/internal/processing/service.go, core-backend/internal/processing/repository.go, core-backend/internal/channelresults/service.go, core-backend/internal/channelresults/service_test.go, core-backend/internal/app/app.go</files>
  <behavior>
    - Test 1: one successful channel result marks only that channel as succeeded and leaves other channels untouched.
    - Test 2: temporary failures increment attempt_count and preserve retryability until the configured max is reached.
    - Test 3: fatal failures or exhausted retries produce terminal channel state with persisted error code/message.
    - Test 4: if any mandatory channel becomes exhausted, the parent examination enters a final error pipeline state.
  </behavior>
  <action>Implement a shared result consumer in core backend that subscribes to the unified results queue, validates the documented envelope, and hands it to a `channelresults` service. That service must own PostgreSQL updates for channel-run status transitions, normalized result persistence, retry counters, and terminal examination failure projection. Keep workers dumb: they classify and publish, but core decides persisted state transitions.</action>
  <verify>
    <automated>cd /home/vadim/diplom/core-backend && go test ./internal/channelresults ./internal/processing -run 'TestIndependentChannelCompletion|TestRetryBudget|TestFatalVsTemporaryError|TestMandatoryChannelExhaustionFailsExamination' -count=1</automated>
  </verify>
  <done>Core backend persists channel success/failure independently and derives final examination error state from mandatory-channel exhaustion.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Expose backend-authoritative processing progress over HTTP</name>
  <files>core-backend/internal/http/router.go, core-backend/internal/http/examinations_handler.go, core-backend/internal/http/processing_status_test.go</files>
  <behavior>
    - Test 1: the progress endpoint returns pipeline status and three mandatory channel entries for an active examination.
    - Test 2: terminal failure is returned with the failing channel and persisted error details.
    - Test 3: the endpoint works without reading broker management state.
  </behavior>
  <action>Add `GET /examinations/{id}/processing-status` to the operator/admin backend surface. The handler should return a DTO built entirely from PostgreSQL projection: overall pipeline status, `is_terminal`, per-channel statuses, attempt metadata, timestamps, and last error details. Align runtime status handling with the Phase 2 contract decision: examination state may progress through `processing` and `failed`, but rich per-channel detail is exposed only through this dedicated endpoint rather than overloaded into generic list/history payloads.</action>
  <verify>
    <automated>cd /home/vadim/diplom/core-backend && go test ./internal/http -run 'TestProcessingStatusEndpoint|TestProcessingStatusEndpointReturnsTerminalError' -count=1</automated>
  </verify>
  <done>Frontend has one stable HTTP endpoint that exposes authoritative per-channel progress and final failure state.</done>
</task>

</tasks>

<verification>
Run both result-state and HTTP projection tests; confirm terminal pipeline error is observable without inspecting RabbitMQ.
</verification>

<success_criteria>
- Unified result messages drive persisted channel-run transitions in PostgreSQL.
- Retry exhaustion or fatal failure on a mandatory channel produces a final examination `failed` state.
- `GET /examinations/{id}/processing-status` exposes per-channel progress for UI polling while remaining the rich runtime surface.
</success_criteria>

<output>
After completion, create `.planning/phases/02-asynchronous-multichannel-processing/02-04-SUMMARY.md`
</output>
