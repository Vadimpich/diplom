---
phase: 02-asynchronous-multichannel-processing
plan: 03
type: execute
wave: 2
depends_on:
  - 02-01
files_modified:
  - docker-compose.yml
  - .env.example
  - docs/01_contract.md
  - ml-services/ml-text/Dockerfile
  - ml-services/ml-text/requirements.txt
  - ml-services/ml-text/app/main.py
  - ml-services/ml-acoustic/Dockerfile
  - ml-services/ml-acoustic/requirements.txt
  - ml-services/ml-acoustic/app/main.py
  - ml-services/ml-paralinguistic/Dockerfile
  - ml-services/ml-paralinguistic/requirements.txt
  - ml-services/ml-paralinguistic/app/main.py
autonomous: true
requirements:
  - PIPE-02
  - QUAL-03
must_haves:
  truths:
    - Each mandatory channel has a runnable worker process that can consume independently.
    - Workers publish one unified result shape instead of channel-specific ad hoc payloads.
    - Local Compose brings up backend, broker, storage, and all mandatory channel services together.
  artifacts:
    - One worker service directory exists for each mandatory channel.
    - docker-compose.yml defines text, acoustic, and paralinguistic worker containers.
    - .env.example documents the broker/S3 env vars required by those workers.
  key_links:
    - Every worker consumes only its own queue and publishes to the shared results exchange/queue.
    - Workers read audio through S3 references from the command envelope, never from queue binaries.
---

<objective>
Add stub-but-runnable worker services and the local Compose wiring for the three mandatory analytic channels.

Purpose: satisfy `PIPE-02` and `QUAL-03` early with real containers that exercise the broker contract, even before real ML inference exists.
Output: three minimal Python worker services, Compose wiring, and documented runtime configuration.
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
@docs/00_project.md
@docs/01_contract.md
@docker-compose.yml
</context>

<tasks>

<task type="auto">
  <name>Task 1: Create three independent runnable worker shells</name>
  <files>ml-services/ml-text/Dockerfile, ml-services/ml-text/requirements.txt, ml-services/ml-text/app/main.py, ml-services/ml-acoustic/Dockerfile, ml-services/ml-acoustic/requirements.txt, ml-services/ml-acoustic/app/main.py, ml-services/ml-paralinguistic/Dockerfile, ml-services/ml-paralinguistic/requirements.txt, ml-services/ml-paralinguistic/app/main.py</files>
  <action>Create `ml-services/ml-text`, `ml-services/ml-acoustic`, and `ml-services/ml-paralinguistic` services as lightweight FastAPI + `aio-pika` workers with health endpoints and robust RabbitMQ connections. Each service should consume only its own command queue, fetch the referenced audio object from MinIO/S3, emit a stub normalized payload with `model_version`, and classify failures as `temporary_error` or `fatal_error` in the unified result envelope. Keep them stateless and independent; they must not write to PostgreSQL or call other channel services.</action>
  <verify>
    <automated>cd /home/katya/dimplom && docker compose up -d --build text-worker acoustic-worker paralinguistic-worker && docker compose ps</automated>
  </verify>
  <done>All three mandatory channels have runnable worker containers that consume independently and publish the same result envelope shape.</done>
</task>

<task type="auto">
  <name>Task 2: Wire workers into the reproducible local stack</name>
  <files>docker-compose.yml, .env.example, docs/01_contract.md</files>
  <action>Extend the root Compose stack with the three worker services, explicit broker/S3 environment variables, and startup dependencies that match the existing self-hosted topology. Update `.env.example` with worker configuration and queue names. Keep `docs/01_contract.md` synchronized with the shipped AMQP topology by confirming or refining exchange names, per-channel queues, routing keys, unified result routing, and explicit RabbitMQ 3.13 retry/DLX assumptions as they land in Compose and worker code. Do not upgrade RabbitMQ in this phase unless the code explicitly requires it; stay on `rabbitmq:3.13-management-alpine` and document the explicit queue assumptions instead.</action>
  <verify>
    <automated>cd /home/katya/dimplom && docker compose up -d --build && docker compose ps</automated>
  </verify>
  <done>The local stack starts frontend, core backend, PostgreSQL, RabbitMQ, MinIO, and three mandatory worker services from one Compose file, with `docs/01_contract.md` kept authoritative for shared queue and routing contracts.</done>
</task>

</tasks>

<verification>
Bring up the full Compose stack and confirm the three worker containers are present and healthy enough to consume from RabbitMQ.
</verification>

<success_criteria>
- `ml-services/ml-text`, `ml-services/ml-acoustic`, and `ml-services/ml-paralinguistic` exist as runnable services.
- Compose starts the full Phase 2 stack with those services included.
- Workers use the shared documented command/result contract and S3 references.
</success_criteria>

<output>
After completion, create `.planning/phases/02-asynchronous-multichannel-processing/02-03-SUMMARY.md`
</output>
