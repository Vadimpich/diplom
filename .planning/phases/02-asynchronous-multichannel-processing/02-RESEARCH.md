# Phase 2: Asynchronous Multichannel Processing - Research

**Researched:** 2026-03-20
**Domain:** RabbitMQ-backed multichannel orchestration with Go core backend, Python ML workers, and frontend progress polling
**Confidence:** MEDIUM

## Summary

Phase 2 should extend the existing Phase 1 finish fence, not replace it. The current backend already guarantees that `POST /examinations/{id}/finish` transitions an examination to `ready_for_processing` only once and records `examination_processing_launches`. The missing piece is a durable handoff from PostgreSQL state into RabbitMQ plus persisted per-channel runtime state that the operator UI can poll. PostgreSQL must remain the source of truth for channel progress, retry counts, terminal failures, and final examination pipeline state. RabbitMQ should be treated as the transport for work distribution, not as the authoritative progress store.

The highest-risk design mistake is a DB+broker dual write in the finish request. If the examination state commits but queue publishing fails, the system will report `ready_for_processing` while no worker ever receives work. Use a transactional outbox in PostgreSQL, drained by a background publisher that uses RabbitMQ publisher confirms. Create one durable queue per mandatory channel (`text`, `acoustic`, `paralinguistic`), persist one channel-run row per examination/channel, and consume channel results back into core through a dedicated result exchange/queue. ML workers must stay stateless with respect to core data: fetch audio from S3, compute, publish a versioned result or classified error, and never update PostgreSQL directly.

There is one important version-specific constraint in the current repository: [docker-compose.yml](/home/vadim/diplom/docker-compose.yml) pins `rabbitmq:3.13-management-alpine`, while RabbitMQ 4.0+ changed quorum-queue defaults so `delivery-limit` now defaults to `20`. On RabbitMQ 3.13, quorum queues support `delivery-limit`, but there is no default. Phase 2 must therefore either upgrade Compose to RabbitMQ 4.x or explicitly declare/policy-set `delivery-limit` and DLX behavior in 3.13. Do not plan against 4.x defaults unless the stack is upgraded in this phase.

**Primary recommendation:** Implement Phase 2 as `finish fence -> PostgreSQL outbox + channel_runs -> RabbitMQ per-channel quorum queues -> Python workers -> result queue -> core result consumer -> frontend polling from backend progress DTO`.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| PIPE-01 | Core backend publishes analysis tasks for `text`, `acoustic`, `paralinguistic` with versioned payload and S3 audio reference | Use transactional outbox, publisher confirms, durable per-channel queues, and versioned message envelopes with `audio_s3_key` plus examination/answer IDs |
| PIPE-02 | Each ML service consumes independently and returns unified channel result | Use one queue per mandatory channel, one worker service per channel, unified result envelope on a result exchange/queue, no cross-channel runtime dependency |
| PIPE-03 | Retries are bounded and temporary vs fatal errors are persisted per channel | Persist retry counters and terminal status in PostgreSQL; use RabbitMQ quorum queue `delivery-limit`, DLX, and explicit error classification in worker results |
| PIPE-04 | If any mandatory channel exhausts retries, the examination becomes final error | Add examination pipeline state derived from persisted channel-run rows; one exhausted mandatory channel closes the examination into final pipeline error |
| RSLT-01 | Operator sees per-channel processing progress | Expose backend DTO with per-channel statuses and timestamps; frontend polls with TanStack Query and renders backend-authoritative channel badges |
| QUAL-03 | Local self-hosted stack is reproducible with frontend, core, PostgreSQL, RabbitMQ, S3, and analytic services | Extend root Compose with three minimal ML worker containers plus env/config for broker queues and S3 access; verify end-to-end startup locally |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| RabbitMQ | 3.13.x currently in repo; 4.x preferred if upgraded | Work distribution, redelivery limits, DLX, queue isolation per channel | Official broker already present in repo; quorum queues and delivery-limit cover poison-message and bounded-retry cases better than custom retry loops |
| `github.com/rabbitmq/amqp091-go` | `v1.10.0` (published 2024-05-08) | Go AMQP 0-9-1 client for core publisher and result consumer | Maintained by RabbitMQ team; natural choice for Go core backend |
| PostgreSQL outbox table + channel-run tables | project schema extension | Durable publication intent, retry ledger, channel progress, result persistence | Solves DB/broker dual write and keeps UI progress in one authoritative store |
| FastAPI | `0.135.1` (published 2026-03-01) | HTTP health/readiness and worker process shell for Python ML services | Already matches project architecture in `docs/00_project.md`; common for lightweight ML service shells |
| Pydantic | `2.12.5` (published 2025-11-26) | Versioned message/result schemas in Python workers | Strong schema validation and clear contract serialization |
| `aio-pika` | `9.6.1` (published 2026-02-23) | Async RabbitMQ client for ML workers | Fits FastAPI lifespan startup and robust async consumption model |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Uvicorn | `0.42.0` (published 2026-03-16) | Serve worker health endpoints and startup lifecycle | Use for each ML service container |
| TanStack Query | `5.91.3` (registry modified 2026-03-20) | Poll backend progress endpoints from the operator processing screen | Use for per-channel progress until terminal success/error |
| Next.js | `16.2.0` (registry modified 2026-03-19) | Existing frontend platform | Use existing app router shell; no need for websocket infrastructure in this phase |
| MinIO/S3 | repo-local existing stack | Audio object retrieval by workers | Continue passing only object references, never binaries through RabbitMQ |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `aio-pika` | `pika` | `pika` is workable, but `aio-pika` aligns better with async FastAPI lifespan and robust connection patterns |
| Result queue back into core | HTTP callback from workers into core | HTTP callback is simpler initially but weakens buffering and raises coupling during worker/core restarts |
| Quorum queues | Classic queues | Classic queues are lighter, but quorum queues give first-class poison-message handling and explicit `delivery-limit` support |
| Backend-authoritative polling | Broker-derived progress in UI | Reading broker state directly gives misleading progress and duplicates business state outside PostgreSQL |

**Installation:**
```bash
cd /home/vadim/diplom/core-backend
go get github.com/rabbitmq/amqp091-go@v1.10.0

# per Python worker service
pip install fastapi==0.135.1 pydantic==2.12.5 aio-pika==9.6.1 uvicorn==0.42.0
```

**Version verification:** Before planning tasks, re-check moving versions:
```bash
cd /home/vadim/diplom/core-backend && go list -m -json github.com/rabbitmq/amqp091-go@latest
cd /home/vadim/diplom/frontend && npm view @tanstack/react-query version && npm view next version
curl -s https://pypi.org/pypi/fastapi/json | jq -r '.info.version, .releases[.info.version][-1].upload_time_iso_8601'
curl -s https://pypi.org/pypi/pydantic/json | jq -r '.info.version, .releases[.info.version][-1].upload_time_iso_8601'
curl -s https://pypi.org/pypi/aio-pika/json | jq -r '.info.version, .releases[.info.version][-1].upload_time_iso_8601'
curl -s https://pypi.org/pypi/uvicorn/json | jq -r '.info.version, .releases[.info.version][-1].upload_time_iso_8601'
```

## Architecture Patterns

### Recommended Project Structure
```text
core-backend/
├── internal/processing/          # pipeline orchestration, channel-run state machine
├── internal/outbox/              # durable publish intents + relay
├── internal/rabbitmq/            # AMQP connection, topology, publisher, consumer
├── internal/channelresults/      # result ingestion and persistence
└── migrations/                   # channel-runs, outbox, result tables

ml-services/
├── shared/                       # common schemas, broker helpers, S3/audio helpers
├── text-service/
├── acoustic-service/
└── paralinguistic-service/

frontend/
└── app/(app)/operator/examinations/[id]/processing/   # backend-driven progress UI
```

### Pattern 1: Transactional Outbox After Finish Fence
**What:** Finish transaction creates channel-run rows and outbox messages in PostgreSQL; a background relay publishes them to RabbitMQ only after commit.
**When to use:** Every time examination pipeline state and queue publication must stay consistent.
**Example:**
```go
// Source: https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/transactional-outbox.html
// Source: https://www.rabbitmq.com/docs/confirms
func (s *Service) FinishAndEnqueue(ctx context.Context, examID int64) error {
	return s.repo.WithTx(ctx, func(tx Tx) error {
		exam, err := tx.FinishExam(ctx, examID)
		if err != nil {
			return err
		}

		for _, channel := range []string{"text", "acoustic", "paralinguistic"} {
			runID, err := tx.CreateChannelRun(ctx, exam.ID, channel)
			if err != nil {
				return err
			}
			if err := tx.InsertOutboxMessage(ctx, runID, channel); err != nil {
				return err
			}
		}
		return nil
	})
}
```

### Pattern 2: One Durable Queue Per Mandatory Channel
**What:** Declare separate durable queues and bindings for `text`, `acoustic`, and `paralinguistic`.
**When to use:** Always. The phase requirement explicitly needs independent channel runtime and scale.
**Example:**
```text
processing.commands (topic exchange)
├── processing.text        -> queue: qq.processing.text
├── processing.acoustic    -> queue: qq.processing.acoustic
└── processing.paralinguistic -> queue: qq.processing.paralinguistic

processing.results (topic exchange)
└── queue: qq.processing.results
```

### Pattern 3: Persisted Channel-Run State Machine
**What:** Model one row per examination/channel with explicit status and retry counters.
**When to use:** For backend progress DTOs, retry budgeting, and terminal failure propagation.
**Example:**
```text
channel_status:
- pending
- queued
- processing
- succeeded
- retry_scheduled
- failed_temporary
- failed_fatal
- exhausted
```

Recommended columns:
- `examination_id`
- `channel`
- `status`
- `attempt_count`
- `max_attempts`
- `message_version`
- `last_error_code`
- `last_error_message`
- `started_at`
- `finished_at`
- `last_heartbeat_at` if long-running workers need liveness later

### Pattern 4: Worker-Owned Classification, Core-Owned State Transition
**What:** Workers classify errors as temporary or fatal and publish a structured result; core alone updates examination/channel state in PostgreSQL.
**When to use:** For every result and failure path.
**Example:**
```json
{
  "message_version": 1,
  "examination_id": 42,
  "channel": "text",
  "attempt": 2,
  "status": "temporary_error",
  "error_code": "s3_timeout",
  "error_message": "object fetch timed out",
  "model_version": "stub-0.1.0",
  "trace_id": "..."
}
```

### Pattern 5: Polling UI From Backend DTO, Not From RabbitMQ
**What:** Add a backend endpoint that returns examination pipeline summary plus per-channel statuses and retry info; frontend polls it with TanStack Query.
**When to use:** Processing screen until every mandatory channel reaches success or a terminal error.
**Example:**
```ts
// Source: https://tanstack.com/query/v5/docs/framework/react/guides/query-retries
const query = useQuery({
  queryKey: ["exam-processing", examinationId],
  queryFn: () => apiClient.getExaminationProcessing(examinationId),
  refetchInterval: (q) => q.state.data?.is_terminal ? false : 3000,
  refetchIntervalInBackground: true,
  retry: false,
});
```

### Anti-Patterns to Avoid
- **Publishing directly inside `Finish` without outbox:** creates a dual-write inconsistency between PostgreSQL and RabbitMQ.
- **One shared work queue for all channels:** breaks independent scaling and complicates channel-specific retry/error accounting.
- **Letting ML workers write directly to core PostgreSQL:** violates project architecture and creates hidden coupling.
- **Using queue depth as operator progress:** queue depth says nothing about business completion or per-channel terminal status.
- **Infinite `nack(requeue=true)` loops:** produces poison-message churn and can exhaust broker disk/log resources.
- **Relying on RabbitMQ 4.x defaults while Compose stays on 3.13:** will silently undercut retry assumptions.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| DB + broker atomicity | Direct request-time publish with ad hoc retries | Transactional outbox | Dual writes fail in inconsistent ways; outbox is the standard mitigation |
| Worker transport client in Go | Custom AMQP protocol wrapper | `amqp091-go` | RabbitMQ-maintained Go client already covers AMQP 0-9-1 semantics |
| Worker reconnection in Python | Raw reconnect loop over sockets | `aio-pika` robust connection/channel | Existing library already models async reconnect and robust consumers |
| Retry/dead-letter mechanics | Custom sleep queues and homegrown retry counters only in memory | Quorum queue `delivery-limit` + DLX + persisted attempt ledger | Broker handles transport redelivery; PostgreSQL keeps business truth |
| UI polling scheduler | `setInterval` state machine per component | TanStack Query polling | Existing frontend stack already has battle-tested polling/backoff controls |

**Key insight:** Phase 2 is mostly an orchestration and state-modeling problem, not a model-inference problem. The safest plan is to reuse broker, schema, and polling primitives that already exist in the ecosystem and keep custom logic only in the domain state machine.

## Common Pitfalls

### Pitfall 1: Dual-Write Gap On Finish
**What goes wrong:** Examination status commits, but publish to RabbitMQ fails.
**Why it happens:** HTTP handler tries to write PostgreSQL and publish AMQP in one request path without a durable bridge.
**How to avoid:** Insert outbox rows in the same DB transaction as finish/channel-run creation; publish later with confirms.
**Warning signs:** Examinations stuck in `ready_for_processing` or `processing` with no broker message IDs and no channel rows.

### Pitfall 2: Retry Budget Exists Only In RabbitMQ
**What goes wrong:** Operator cannot see why a channel failed or how many attempts were consumed.
**Why it happens:** Delivery counts stay in headers/DLX only, never materialized in PostgreSQL.
**How to avoid:** Persist attempt count, last error class, and terminal reason on every channel run.
**Warning signs:** UI can only show generic “processing failed”.

### Pitfall 3: Requeue Loops With Prefetch Greater Than 1
**What goes wrong:** Multiple outstanding deliveries are repeatedly requeued and can be discarded together on quorum queues.
**Why it happens:** Worker uses broad concurrency with no per-consumer QoS discipline.
**How to avoid:** Use per-consumer QoS and start with `prefetch=1` for Phase 2 workers.
**Warning signs:** Sudden clustered failures, repeated redeliveries, and inconsistent attempt counts across messages.

### Pitfall 4: Progress Is Derived In Frontend Instead Of Backend
**What goes wrong:** UI invents statuses that diverge from actual pipeline state.
**Why it happens:** Current placeholder screen is tempted to infer progress from time elapsed or queue state.
**How to avoid:** Add an explicit processing-status endpoint and keep frontend as a renderer only.
**Warning signs:** Browser refresh changes visible status without any backend event.

### Pitfall 5: Contract Drift Between Go And Python
**What goes wrong:** Core publishes one message shape, workers expect another, or result envelopes drift per channel.
**Why it happens:** Contract changes are made in code without updating [docs/01_contract.md](/home/vadim/diplom/docs/01_contract.md).
**How to avoid:** Define one versioned command envelope and one versioned result envelope in docs before implementation.
**Warning signs:** Channel-specific parsers or undocumented optional fields start appearing.

### Pitfall 6: Compose Stack Passes Infra But Not Real Workers
**What goes wrong:** `docker compose up` works, but no analytic service consumes anything, so QUAL-03 is not actually satisfied.
**Why it happens:** Local stack includes broker/S3 only, not runnable worker containers.
**How to avoid:** Add three minimal worker services in Compose during this phase, even if they initially return stub features.
**Warning signs:** Processing screen never leaves queued state in local environments.

## Code Examples

Verified patterns from official sources:

### Go Publisher Relay With Confirms
```go
// Source: https://www.rabbitmq.com/docs/confirms
// Source: https://github.com/rabbitmq/amqp091-go
func publishOutbox(msg OutboxMessage, ch *amqp.Channel) error {
	if err := ch.Confirm(false); err != nil {
		return err
	}

	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	err := ch.PublishWithContext(
		context.Background(),
		"processing.commands",
		msg.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    msg.MessageID,
			Body:         msg.Body,
		},
	)
	if err != nil {
		return err
	}

	confirm := <-confirms
	if !confirm.Ack {
		return errors.New("publish not confirmed")
	}
	return nil
}
```

### Python Worker With Explicit Ack / Retry / Fatal Classification
```python
# Source: https://docs.aio-pika.com/apidoc.html
# Source: https://www.rabbitmq.com/docs/quorum-queues
async def handle_message(message: aio_pika.IncomingMessage) -> None:
    payload = json.loads(message.body)
    try:
        result = await run_channel(payload)
        await publish_result(status="succeeded", payload=result)
        await message.ack()
    except TemporaryWorkerError as exc:
        await publish_result(status="temporary_error", error_code=exc.code)
        await message.nack(requeue=True)
    except FatalWorkerError as exc:
        await publish_result(status="fatal_error", error_code=exc.code)
        await message.reject(requeue=False)
```

### FastAPI Lifespan For Worker Startup
```python
# Source: https://fastapi.tiangolo.com/advanced/events/
@asynccontextmanager
async def lifespan(app: FastAPI):
    app.state.connection = await aio_pika.connect_robust(settings.amqp_url)
    app.state.channel = await app.state.connection.channel()
    await app.state.channel.set_qos(prefetch_count=1)
    yield
    await app.state.connection.close()

app = FastAPI(lifespan=lifespan)
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Classic queue retry loops with weak poison-message handling | Quorum queues with `delivery-limit` and DLX | Available in 3.13, default limit added in 4.0 | Better bounded retry control and poison-message behavior |
| Placeholder frontend processing screen | Backend-authoritative per-channel progress DTO + polling | Needed now in Phase 2 | Operator sees real channel state instead of shell status |
| Direct publish after DB write | Transactional outbox + relay publisher confirms | Long-standing distributed-systems best practice | Removes finish-path inconsistency risk |

**Deprecated/outdated:**
- `tiangolo/uvicorn-gunicorn-fastapi` base image: deprecated by FastAPI docs; build worker images from the official Python image instead.
- Browser-derived examination progress: already inconsistent with the project’s Phase 1 decision to render backend workflow statuses as source of truth.

## Open Questions

1. **Should Phase 2 introduce a new examination-level status such as `processing` or keep `ready_for_processing` plus channel runs?**
   - What we know: current repo and contracts stop at `ready_for_processing`.
   - What's unclear: whether the user-facing workflow should gain a distinct in-flight examination status now or let per-channel statuses carry the state.
   - Recommendation: add a new examination pipeline status family now; otherwise `ready_for_processing` becomes semantically overloaded once work is actually running.

2. **Should the repository stay on RabbitMQ 3.13 for this phase or upgrade to 4.x?**
   - What we know: local Compose currently pins `rabbitmq:3.13-management-alpine`; 3.13 supports quorum `delivery-limit`, but RabbitMQ 4.0 adds the default limit of `20`.
   - What's unclear: whether upgrading the broker is acceptable scope for Phase 2.
   - Recommendation: if minimizing churn is more important, stay on 3.13 and configure `delivery-limit` explicitly by policy. If not, upgrade early so planning can rely on current RabbitMQ defaults.

3. **Do real analytic models belong in Phase 2, or are stub workers acceptable?**
   - What we know: phase requirements focus on async orchestration, retries, progress, and local runnable stack.
   - What's unclear: whether placeholder feature extraction is acceptable until Phase 3 aggregation.
   - Recommendation: implement minimal stub workers that consume, fetch S3 audio, emit deterministic normalized payloads, and exercise retry/error paths. Real model logic can evolve later without blocking pipeline architecture.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` + `httptest`; frontend static verification today, frontend behavioral test runner missing |
| Config file | none |
| Quick run command | `cd /home/vadim/diplom/core-backend && go test ./internal/examinations ./internal/http -count=1` |
| Full suite command | `cd /home/vadim/diplom/core-backend && go test ./... -count=1 && cd /home/vadim/diplom/frontend && npm run lint && npx tsc --noEmit && npm run build` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| PIPE-01 | Finish creates versioned outbox tasks with S3 references for all mandatory channels | integration | `cd /home/vadim/diplom/core-backend && go test ./internal/processing ./internal/http -run 'TestFinishCreatesOutboxForMandatoryChannels' -count=1` | ❌ Wave 0 |
| PIPE-02 | Each channel result is ingested independently and persisted without waiting for another channel runtime | integration | `cd /home/vadim/diplom/core-backend && go test ./internal/channelresults ./internal/processing -run 'TestIndependentChannelCompletion' -count=1` | ❌ Wave 0 |
| PIPE-03 | Temporary and fatal failures consume bounded retry budget and persist per-channel error state | unit/integration | `cd /home/vadim/diplom/core-backend && go test ./internal/processing -run 'TestRetryBudget|TestFatalVsTemporaryError' -count=1` | ❌ Wave 0 |
| PIPE-04 | Exhausted mandatory channel moves examination into final pipeline error state | integration | `cd /home/vadim/diplom/core-backend && go test ./internal/processing ./internal/examinations -run 'TestMandatoryChannelExhaustionFailsExamination' -count=1` | ❌ Wave 0 |
| RSLT-01 | Processing screen renders backend per-channel progress correctly | manual-only until frontend test runner exists | `cd /home/vadim/diplom/frontend && npm run lint && npx tsc --noEmit` | ❌ Wave 0 |
| QUAL-03 | Local container stack starts core, frontend, broker, storage, and worker services together | smoke | `cd /home/vadim/diplom && docker compose up -d --build && docker compose ps` | ⚠️ Partial: compose exists, worker services missing |

### Sampling Rate
- **Per task commit:** `cd /home/vadim/diplom/core-backend && go test ./internal/... -count=1`
- **Per wave merge:** `cd /home/vadim/diplom/core-backend && go test ./... -count=1 && cd /home/vadim/diplom/frontend && npm run lint && npx tsc --noEmit`
- **Phase gate:** Full suite green plus `docker compose up -d --build` with worker containers healthy before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `core-backend/internal/processing/*` — orchestration tests for outbox, retry ledger, and terminal failure propagation
- [ ] `core-backend/internal/channelresults/*` — result ingestion and idempotency tests
- [ ] `core-backend/internal/rabbitmq/*` — topology/publisher/consumer tests with stubs around confirm behavior
- [ ] `frontend` behavioral tests — add a lightweight runner for processing-screen status rendering; current repo has no first-party frontend tests
- [ ] Worker container definitions under Compose — required to satisfy QUAL-03 in practice, not just infra-only startup

## Sources

### Primary (HIGH confidence)
- Project docs: [docs/00_project.md](/home/vadim/diplom/docs/00_project.md), [docs/01_contract.md](/home/vadim/diplom/docs/01_contract.md), [docs/02_implementation.md](/home/vadim/diplom/docs/02_implementation.md) - target architecture, contract rules, and current baseline
- Project planning docs: [.planning/REQUIREMENTS.md](/home/vadim/diplom/.planning/REQUIREMENTS.md), [.planning/ROADMAP.md](/home/vadim/diplom/.planning/ROADMAP.md), [.planning/codebase/ARCHITECTURE.md](/home/vadim/diplom/.planning/codebase/ARCHITECTURE.md) - phase scope and current code structure
- RabbitMQ Quorum Queues: https://www.rabbitmq.com/docs/quorum-queues - poison-message handling, `delivery-limit`, prefetch caveats
- RabbitMQ Confirms: https://www.rabbitmq.com/docs/confirms - publisher confirms semantics and reliability expectations
- RabbitMQ Dead Letter Exchanges: https://www.rabbitmq.com/docs/dlx - DLX configuration, cycles, safety caveats
- RabbitMQ Consumers: https://www.rabbitmq.com/docs/next/consumers - consumer capacity and acknowledgement timeout behavior
- `amqp091-go` README: https://github.com/rabbitmq/amqp091-go and https://raw.githubusercontent.com/rabbitmq/amqp091-go/main/README.md - maintained Go client and reconnect non-goal
- FastAPI Lifespan docs: https://fastapi.tiangolo.com/advanced/events/ - recommended startup/shutdown pattern for worker services
- FastAPI Docker docs: https://fastapi.tiangolo.com/deployment/docker/ - current container guidance and deprecated base image clarification
- `aio-pika` API docs: https://docs.aio-pika.com/apidoc.html - robust connections, consumers, ack/reject API
- TanStack Query docs: https://tanstack.com/query/v5/docs/framework/react/guides/query-retries - polling and refetch guidance in current frontend stack

### Secondary (MEDIUM confidence)
- AWS Prescriptive Guidance - Transactional outbox pattern: https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/transactional-outbox.html - authoritative explanation of the dual-write problem and outbox mitigation
- PyPI package metadata for FastAPI, Pydantic, `aio-pika`, Uvicorn - current package versions and publish dates
- npm registry metadata for Next.js and TanStack Query - current frontend package versions

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - official docs and registry metadata confirm the broker/client/framework choices and current versions
- Architecture: MEDIUM - outbox, queue topology, and state-model recommendations are strongly supported but still partially inferred for this specific codebase
- Pitfalls: HIGH - supported by RabbitMQ official docs plus concrete constraints from the current repository

**Research date:** 2026-03-20
**Valid until:** 2026-04-19
