---
phase: 02-asynchronous-multichannel-processing
plan: 02
type: execute
wave: 2
depends_on:
  - 02-01
files_modified:
  - core-backend/go.mod
  - core-backend/go.sum
  - core-backend/internal/config/config.go
  - core-backend/internal/app/app.go
  - core-backend/internal/processing/service.go
  - core-backend/internal/processing/repository.go
  - core-backend/internal/processing/publisher.go
  - core-backend/internal/processing/service_test.go
autonomous: true
requirements:
  - PIPE-01
  - PIPE-03
must_haves:
  truths:
    - A finished examination durably fans out three mandatory commands without publishing inside the request transaction.
    - Broker publishing uses explicit RabbitMQ 3.13-compatible topology and bounded retry semantics.
    - Publisher-side failures are visible in PostgreSQL and can be retried safely.
  artifacts:
    - core-backend/internal/processing/service.go extends finish orchestration with outbox creation.
    - core-backend/internal/processing/publisher.go drains pending outbox rows with publisher confirms.
    - core-backend/internal/config/config.go exposes broker and retry settings needed by the publisher.
  key_links:
    - Finish handler delegates to processing service, which writes channel runs and outbox rows atomically.
    - The outbox relay publishes to per-channel queues declared explicitly for RabbitMQ 3.13, not 4.x defaults.
---

<objective>
Implement durable command publication from the core backend using a transactional outbox and explicit broker topology.

Purpose: satisfy `PIPE-01` without a DB+broker dual write and establish the retry budget primitives for later failure handling.
Output: finish-path outbox orchestration, background publisher, and RabbitMQ configuration in the core backend.
</objective>

<execution_context>
@/home/katya/.codex/get-shit-done/workflows/execute-plan.md
@/home/katya/.codex/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/02-asynchronous-multichannel-processing/01-PLAN.md
@.planning/phases/02-asynchronous-multichannel-processing/02-RESEARCH.md
@.planning/phases/02-asynchronous-multichannel-processing/02-VALIDATION.md
@core-backend/internal/app/app.go
@core-backend/internal/config/config.go
@core-backend/internal/examinations/service.go
@core-backend/internal/http/examinations_handler.go
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Extend the finish fence into outbox-backed processing launch</name>
  <files>core-backend/internal/processing/service.go, core-backend/internal/processing/repository.go, core-backend/internal/app/app.go, core-backend/internal/processing/service_test.go</files>
  <behavior>
    - Test 1: first finish call creates exactly three channel runs and three outbox rows.
    - Test 2: repeated finish remains idempotent and does not duplicate channel runs or outbox rows.
    - Test 3: queued rows capture message_version, examination identifiers, and S3 references needed by workers.
  </behavior>
  <action>Introduce a dedicated processing service/repository that wraps the Phase 1 finish fence and, in the same transaction, creates one channel-run row plus one outbox row for each mandatory channel. Wire this service into `app.go` and the examinations handler path so the public finish endpoint still returns promptly after persistence. Keep PostgreSQL as the source of truth; do not publish AMQP messages directly from the HTTP request path.</action>
  <verify>
    <automated>cd /home/katya/dimplom/core-backend && go test ./internal/processing ./internal/http -run 'TestFinishCreatesOutboxForMandatoryChannels' -count=1</automated>
  </verify>
  <done>Finishing an examination durably records the mandatory processing work in PostgreSQL exactly once.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Build the RabbitMQ outbox relay with explicit 3.13 queue semantics</name>
  <files>core-backend/go.mod, core-backend/go.sum, core-backend/internal/config/config.go, core-backend/internal/app/app.go, core-backend/internal/processing/publisher.go, core-backend/internal/processing/service_test.go</files>
  <behavior>
    - Test 1: pending outbox rows are published to channel-specific routing keys and marked as published only after confirm.
    - Test 2: publish failure leaves the row retryable instead of silently dropping it.
    - Test 3: queue declaration uses explicit quorum/DLX/delivery-limit settings compatible with RabbitMQ 3.13.
  </behavior>
  <action>Add `github.com/rabbitmq/amqp091-go`, then implement a background outbox relay that opens AMQP channels on startup, declares one durable queue per mandatory channel plus the command exchange, and publishes with publisher confirms. Expose config knobs for broker DSN, outbox polling interval, and max attempts. Because Compose pins `rabbitmq:3.13-management-alpine`, explicitly set quorum-queue arguments and `delivery-limit`/DLX behavior instead of assuming RabbitMQ 4.x defaults.</action>
  <verify>
    <automated>cd /home/katya/dimplom/core-backend && go test ./internal/processing -run 'TestOutboxRelayPublishesPendingMessages|TestRetryBudget|TestFatalVsTemporaryError' -count=1</automated>
  </verify>
  <done>Core backend can publish versioned commands to per-channel RabbitMQ queues from the outbox relay with explicit, bounded broker semantics.</done>
</task>

</tasks>

<verification>
Run the targeted processing tests and inspect that idempotent finish plus confirmed outbox publish paths are both covered.
</verification>

<success_criteria>
- `POST /examinations/{id}/finish` persists channel runs and outbox rows transactionally.
- A background relay publishes those rows into explicit channel queues with publisher confirms.
- Broker retry semantics are declared explicitly for RabbitMQ 3.13.
</success_criteria>

<output>
After completion, create `.planning/phases/02-asynchronous-multichannel-processing/02-02-SUMMARY.md`
</output>
