# Phase 4: Decision Delivery To Operator - Research

**Researched:** 2026-03-23
**Domain:** Core-owned decision delivery to external WiMi/KESMI with operator-facing normalized result and diagnostics
**Confidence:** MEDIUM

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Introduce two stable internal contracts: `decision_input` and `decision_result`.
- **D-02:** `decision_input` is built from the canonical aggregated profile, baseline data, and service metadata rather than from raw channel payloads.
- **D-03:** `decision_result` is the authoritative contract for PostgreSQL persistence and operator UI; raw WiMi responses are diagnostics only, not the primary app contract.
- **D-04:** After `aggregated`, the examination enters `decision_pending`; after a successful external decision response it moves to `completed`.
- **D-05:** Persist decision-delivery attempts separately from the examination aggregate, including correlation ID, payload version, raw external error fields, normalized recommendation, timestamps, and attempt number.
- **D-06:** The project does not introduce a separate operator-facing success contract directly from WiMi; normalized persistence and UI are owned by `core-backend`.
- **D-07:** Retry only temporary transport/availability failures: timeout, network failure, HTTP `5xx`, and WiMi pool exhaustion / busy pool.
- **D-08:** Do not retry business/contract failures such as bad parameters, type mismatch, missing model, or constraint violations.
- **D-09:** Retry budget stays intentionally small (`1-2` attempts maximum).
- **D-10:** If retries are exhausted, the system should not masquerade this as a successful decision; planner should preserve explicit integration exhaustion diagnostics in workflow/persistence.
- **D-11:** Until the real WiMi decision model exists, the operator UI should show one honest fixed outcome stating that the analysis is not implemented yet, rather than simulating or inventing a recommendation.
- **D-12:** Until the real model exists, the operator must not see synthetic `допуск / риск / недопуск`; the result page should instead render the fixed non-implemented message plus real delivery diagnostics when integration is attempted.
- **D-13:** Recommendation rendering and diagnostics must depend on the internal `decision_result` contract, not on raw WiMi response shapes.
- **D-14:** WiMi is a mandatory part of the project runtime, not an optional side dependency.
- **D-15:** WiMi should run together with the rest of the system in Compose at least for liveness/smoke verification (`starts and answers`), even before the final model is available.
- **D-16:** No separate custom wrapper service is required; integration stays inside a dedicated module in `core-backend`.

### the agent's Discretion
- Exact PostgreSQL table layout for decision attempts vs decision snapshots.
- Exact naming of the internal adapter module (`internal/kesmi` vs `internal/wimi`).
- Exact diagnostic payload fields stored verbatim vs normalized.
- Exact Compose wiring details for the WiMi container/process wrapper, provided WiMi starts with the rest of the stack and is reachable from `core-backend`.

### Deferred Ideas (OUT OF SCOPE)
- Final WiMi `modelID`, exact `inputParameters`, exact `outputParameters`, and final feature mapping are deferred until the external decision model is delivered.
- Final ML feature semantics and real productive channel models are deferred until the external model development stream finishes.
- Live contract verification against the real WiMi decision model is deferred; only the adapter seam, persistence, statuses, diagnostics, and Compose runtime requirement are locked now.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| KSMI-01 | Система передаёт итоговый нормализованный профиль обследования в КЭСМИ через изолированный integration layer | Core-owned `decision_input` DTO, dedicated adapter module in `core-backend`, outbox/attempt persistence, WiMi `POST /ModelCalc` boundary |
| KSMI-02 | Интеграция с КЭСМИ идемпотентно обрабатывает временные сетевые ошибки, различает transport и business errors и сохраняет correlation ID | Existing `processing` outbox/relay and `channelresults` error-taxonomy patterns, WiMi error classification docs, persisted attempt ledger |
| KSMI-03 | Оператор видит итоговую рекомендацию `допуск / риск / недопуск` вместе с причиной интеграционной ошибки, если внешний вызов не завершился успешно | Backend-authored `decision_result`, extended result/history DTOs, normalized diagnostics instead of raw WiMi payload |
| RSLT-02 | Оператор видит экран результата с итоговым решением, ключевыми показателями состояния, отклонением от baseline и вкладом каналов | Existing Phase 3 result page is reusable; Phase 4 must extend it with decision block and diagnostics without replacing aggregated/baseline content |
</phase_requirements>

## Summary

Phase 4 should be planned as a continuation of the existing `processing -> aggregation -> results` architecture, not as a fresh integration style. The current codebase already has the right building blocks: PostgreSQL as source of truth, an outbox/relay pattern for retryable delivery, a narrow compute/integration client boundary, and backend-authored result DTOs consumed directly by frontend pages. The safest plan is to mirror those patterns for WiMi decision delivery.

The highest-value planning choice is to keep decision delivery asynchronous and persisted. `aggregation.Service.TryAggregate` currently performs the Phase 3 success path inline; Phase 4 should stop treating `aggregated` as final success and instead persist a decision job/snapshot, move the examination to `decision_pending`, and let a dedicated core-owned relay/client perform the WiMi call with a small retry budget. On success, normalize the external response into `decision_result`, persist it, and move the examination to `completed`. On exhausted or business failures, keep explicit diagnostics in persistence and UI rather than pretending the examination completed successfully.

The real WiMi model is still missing, so the planner should optimize for stable seams: versioned DTOs, durable tables, transport/business error classification, correlation IDs, and honest operator messaging. Model-specific feature mapping is deferred and should not shape the internal contracts now.

**Primary recommendation:** Implement Phase 4 as a core-backend decision-delivery subsystem that reuses the existing outbox/retry/persistence patterns, stores normalized `decision_result` separately from aggregated profiles, and extends existing result/history DTOs with operator-safe decision diagnostics.

## Standard Stack

### Core
| Library / Component | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go runtime | repo targets `1.24`, local env has `1.26.0` | Core backend decision orchestration | Existing backend/runtime for all workflow logic |
| `github.com/go-chi/chi/v5` | repo-pinned `v5.2.3` | HTTP routing for result/status endpoints | Existing backend routing stack; no new server framework needed |
| `github.com/jackc/pgx/v5` | repo-pinned `v5.7.4` | PostgreSQL access and transactional persistence | Existing source-of-truth persistence path |
| `github.com/rabbitmq/amqp091-go` | repo-pinned `v1.10.0` | Existing outbox/relay messaging patterns | Already used for bounded retry delivery semantics |
| PostgreSQL | Compose `16-alpine` | Authoritative workflow, attempts, decision snapshot storage | Matches project rule: PostgreSQL is source of truth |
| WiMi server | local vendor package `0.1.7` | External decision engine | Mandatory runtime dependency for Phase 4 smoke/runtime |

### Supporting
| Library / Component | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Next.js | repo-pinned `15.5.14` | Operator result/history UI | Extend current operator pages; avoid frontend stack changes |
| `@tanstack/react-query` | repo-pinned `5.91.2` | Fetching result and history DTOs | Continue backend-authored DTO consumption |
| `zod` | repo-pinned `3.25.76` | Frontend contract validation/types | Reuse if frontend adds parsing for new DTO fields |
| `curl` | local `8.5.0` | Manual WiMi/core smoke probes | Useful before browser-level verification |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Core-owned adapter module + persisted relay | Direct sync HTTP call inside handler/service | Simpler initially, but loses idempotent retries, attempt history, and durable diagnostics |
| Separate decision snapshot + attempts tables | Mutating only aggregated profile row | Fewer tables, but mixes Phase 3 canonical profile with Phase 4 transport concerns |
| Backend-normalized DTO for UI | Raw WiMi JSON in frontend | Faster to prototype, but violates locked decisions and leaks external contract instability |

**Installation:**
```bash
# No new framework is required for planning.
# Phase 4 should extend the repo-pinned stack already present in the repository.
```

**Version verification:**
- `npm view next version` on 2026-03-23 returned `16.2.1` (modified `2026-03-22T23:32:01.413Z`), but the repo is pinned to `15.5.14`; Phase 4 should stay on the repo-pinned version.
- `npm view @tanstack/react-query version` on 2026-03-23 returned `5.95.0` (modified `2026-03-22T17:44:03.524Z`), but the repo is pinned to `5.91.2`.
- `npm view zod version` on 2026-03-23 returned `4.3.6` (modified `2026-01-25T21:51:57.252Z`), but the repo is pinned to `3.25.76`.
- Go module registry verification did not return usable output in this environment, so Go library confidence remains based on repo-pinned versions rather than latest-registry confirmation.

## Architecture Patterns

### Recommended Project Structure
```text
core-backend/internal/
├── decision/           # decision_input / decision_result contracts, service, repository
├── kesmi/              # WiMi/KESMI HTTP client + error classification
├── processing/         # existing relay/outbox pattern to mirror or extend
├── results/            # result/history DTO extension for operator UI
└── http/               # result endpoint and status/history handler updates
```

### Pattern 1: Persist Then Deliver
**What:** Move from `aggregated` to `decision_pending` only after the core backend has durably stored the canonical decision input and created a delivery record/job.
**When to use:** Every examination leaving Phase 3 success.
**Why:** Preserves PostgreSQL as source of truth and avoids losing a decision attempt when the process crashes between state transition and WiMi call.
**Example:**
```go
// Source: /home/vadim/diplom/core-backend/internal/processing/publisher.go
type OutboxMessage struct {
    ExaminationID     int64
    AttemptCount      int32
    MaxAttempts       int32
    Payload           []byte
    BrokerCorrelation string
}
```

### Pattern 2: Classify External Failures Before State Transition
**What:** Reuse the existing `temporary` vs `fatal` handling style from `channelresults`, but map it to WiMi transport/business semantics.
**When to use:** Every WiMi response, timeout, connection error, or malformed request.
**Why:** Phase 4 requirements explicitly require idempotent retry only for transport/availability failures.
**Example:**
```go
// Source: /home/vadim/diplom/core-backend/internal/channelresults/service.go
switch envelope.Status {
case processing.ResultStatusSucceeded:
    run.Status = RunStatusSucceeded
case processing.ResultStatusTemporaryError:
    if envelope.Attempt >= run.MaxAttempts {
        run.Status = RunStatusExhausted
    } else {
        run.Status = RunStatusRetryScheduled
    }
case processing.ResultStatusFatalError:
    run.Status = RunStatusFailedFatal
}
```

### Pattern 3: Backend-Authored Result Surfaces
**What:** Extend `GET /examinations/{id}/result` and likely `GET /specialists/{id}/result-history` from the backend, instead of having the frontend derive decision state from multiple endpoints or raw WiMi fields.
**When to use:** Operator-facing recommendation, diagnostics, and history rendering.
**Why:** Phase 3 already established canonical result DTOs; Phase 4 should add decision fields to those DTOs rather than bypass them.
**Example:**
```ts
// Source: /home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/results/page.tsx
const resultQuery = useQuery({
  queryKey: ["examination-result", examinationId],
  queryFn: () => apiClient.getExaminationResult(examinationId),
});
```

### Anti-Patterns to Avoid
- **Direct WiMi calls from frontend:** violates the isolation boundary and leaks credentials/transport details.
- **Reusing `failed` for decision delivery problems:** `failed` already means mandatory pipeline/post-processing failure; exhausted Phase 4 delivery should remain explicitly diagnosable, not merged into the old failure bucket.
- **Inline sync WiMi call during operator page load:** makes recommendation delivery non-idempotent and couples UX to network availability.
- **Mutating Phase 3 aggregated payload into raw WiMi shape:** breaks the locked `decision_input` seam and makes model-specific mapping infect the canonical profile.
- **Fake `допуск / риск / недопуск` labels before the real model exists:** directly contradicts D-11 and D-12.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Retry orchestration | Ad-hoc `for` loop around `http.Client.Do` inside service code | Persisted outbox/relay pattern modeled after `internal/processing` | Durable retries, crash safety, attempt history, bounded semantics |
| Error taxonomy | Free-text string matching in UI | Typed core-side transport/business classification | Keeps retry logic and UI diagnostics consistent |
| Decision history view | Client-side merge of aggregated result + ad-hoc diagnostics calls | Extend existing `results` repository/service DTOs | One canonical backend response, fewer race conditions |
| WiMi model introspection | Hardcoded assumptions about parameter IDs | Official WiMi `POST /ModelsParametersInfo` when the real model arrives | Prevents stale or invalid feature mapping |
| Success contract | Raw WiMi response as app DTO | Internal `decision_result` snapshot | Stable operator contract even when WiMi payload changes |

**Key insight:** Phase 4 looks like “just one HTTP integration”, but the dangerous complexity is not the HTTP call. It is idempotent delivery, retries, status transitions, correlation, persistence, and honest operator diagnostics. Those should be solved with the project’s existing orchestration patterns, not bespoke request code.

## Common Pitfalls

### Pitfall 1: Misclassifying WiMi business errors as retryable
**What goes wrong:** Missing model, bad parameter IDs, malformed payloads, or constraint violations keep retrying.
**Why it happens:** WiMi returns domain-specific JSON error bodies, not just plain HTTP transport failures.
**How to avoid:** Treat timeouts, connection failures, HTTP `5xx`, and pool exhaustion/busy signals as retryable; classify `5101`-`5105`, `5502`, and constraint-type model failures as non-retryable business errors.
**Warning signs:** Same correlation ID repeatedly fails with identical model/parameter error details.

### Pitfall 2: Losing the exact decision attempt that drove the UI
**What goes wrong:** The operator sees a diagnostic message, but PostgreSQL cannot show which attempt, payload version, or raw external error produced it.
**Why it happens:** Teams persist only the latest summary row.
**How to avoid:** Store a decision snapshot plus append-only attempt rows with attempt number, correlation ID, payload version, timestamps, raw error fields, and normalized outcome.
**Warning signs:** No durable attempt history or inability to explain a retry/exhaustion event after restart.

### Pitfall 3: Treating `aggregated` as terminal success after Phase 4 lands
**What goes wrong:** UI/history routes and workflow logic keep routing `aggregated` as final even though decision delivery still has to happen.
**Why it happens:** Phase 3 currently uses `aggregated` as success.
**How to avoid:** Expand contracts and routing to recognize `decision_pending` and `completed`, while keeping `aggregated` as the Phase 3 boundary only.
**Warning signs:** History page still links `aggregated` only, or processing/result logic stops before the decision block exists.

### Pitfall 4: Coupling internal DTO shape to the current stub WiMi absence
**What goes wrong:** Contracts are designed around “not implemented yet” placeholders instead of around the eventual normalized decision shape.
**Why it happens:** The real model is deferred, so temporary UI shortcuts feel safe.
**How to avoid:** Keep `decision_result` stable now, with a placeholder recommendation mode that is explicit and non-clinical.
**Warning signs:** Frontend hardcodes placeholder text without typed decision metadata from the backend.

## Code Examples

Verified project patterns from current sources:

### Persisted Relay Failure Classification
```go
// Source: /home/vadim/diplom/core-backend/internal/processing/publisher.go
func buildFailureUpdate(msg OutboxMessage, err error) FailedOutboxUpdate {
    status := outboxStatusPending
    errorCode := "temporary_publish_error"
    attemptCount := msg.AttemptCount + 1
    if attemptCount >= msg.MaxAttempts {
        status = outboxStatusFailed
    }
    var fatal FatalPublishError
    if errors.As(err, &fatal) {
        status = outboxStatusFailed
        errorCode = "fatal_publish_error"
        err = fatal.Err
    }
    return FailedOutboxUpdate{Status: status, ErrorCode: errorCode, ErrorMessage: err.Error()}
}
```

### Canonical Aggregated Result Boundary
```go
// Source: /home/vadim/diplom/core-backend/internal/results/service.go
type Repository interface {
    GetExaminationResult(context.Context, int64) (aggregation.AggregatedProfile, error)
    GetSpecialistHistory(context.Context, int64) (SpecialistHistoryResponse, error)
}
```

### Existing Operator Result Page Reuse Point
```tsx
// Source: /home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/results/page.tsx
<CardHeader>
  <CardTitle>Объяснение</CardTitle>
  <CardDescription>{result.summary.neutral_recommendation_placeholder}</CardDescription>
</CardHeader>
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `aggregated` treated as final success | `aggregated` becomes the handoff boundary into decision delivery | Phase 4 planning, 2026-03-23 | Workflow and UI must account for `decision_pending` and `completed` |
| Neutral placeholder in result summary | Backend-normalized `decision_result` with honest placeholder mode until real model exists | Phase 4 planning, 2026-03-23 | UI can stay truthful without coupling to raw WiMi payload |
| One profile row as the only final artifact | Separate aggregated profile, decision snapshot, and attempt ledger | Phase 4 planning, 2026-03-23 | Enables retries, diagnostics, auditability, and operator-safe rendering |

**Deprecated/outdated:**
- `neutral_recommendation_placeholder` as the only “decision” field: acceptable for Phase 3, but insufficient once Phase 4 starts persisting delivery attempts and diagnostics.
- Treating `GET /examinations/{id}/result` as Phase 3-only: it should remain the main result surface, but with Phase 4 extensions.

## Open Questions

1. **Exact persistence split between snapshot and attempts**
   - What we know: decisions must persist attempts separately from the examination aggregate.
   - What's unclear: whether the latest normalized `decision_result` lives on the examination row, a dedicated snapshot table, or the latest attempt row.
   - Recommendation: plan for two tables: one append-only attempts table plus one latest snapshot table keyed by `examination_id`.

2. **Exact normalized diagnostic envelope**
   - What we know: correlation ID, payload version, raw external error fields, normalized recommendation, timestamps, and attempt number are required.
   - What's unclear: how much of raw WiMi success payload should be stored verbatim.
   - Recommendation: persist raw response/error JSON in a JSONB column for diagnostics, but expose only normalized fields through API DTOs.

3. **How to run WiMi in Compose before a finalized model exists**
   - What we know: WiMi is mandatory runtime and should start with the local stack.
   - What's unclear: whether the repo will wrap the vendor `.deb` into a Docker image or mount a host-installed process.
   - Recommendation: planner should include an explicit Compose/devops task to choose one reproducible local runtime path and document the env/health contract.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | Core backend changes/tests | ✓ | `go1.26.0` | — |
| Node.js | Frontend changes/build | ✓ | `v22.22.1` | — |
| npm | Frontend dependency/build commands | ✓ | `11.12.0` | — |
| Python 3 | WiMi helper scripts / ML-side tooling / docs probes | ✓ | `3.12.3` | — |
| `curl` | Manual smoke checks against core/WiMi | ✓ | `8.5.0` | `wget` |
| `wget` | Health checks already used in Compose | ✓ | `1.21.4` | `curl` |
| Docker CLI / Compose v2 | Full local stack, WiMi compose smoke, end-to-end validation | ✗ | — | None in this environment |
| `pytest` (global) | Python test execution convenience | ✗ | — | Use per-service virtualenv as documented in `README.md` |
| `psql` | Direct DB inspection/manual migration validation | ✗ | — | Application-level tests or DB container exec once Docker exists |

**Missing dependencies with no fallback:**
- Docker / Docker Compose: blocks full stack verification, WiMi runtime smoke, and Compose-based phase validation in the current environment.

**Missing dependencies with fallback:**
- Global `pytest`: use the project’s service-specific virtualenv workflow from `README.md`.
- `psql`: query through app/repository tests or containerized postgres once Docker is available.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` + frontend `eslint/build/tsc` + Python `pytest` |
| Config file | `core-backend`: none; `frontend`: `package.json` scripts; `ml-services/ml-baseline`: none |
| Quick run command | `cd /home/vadim/diplom/core-backend && go test ./internal/channelresults ./internal/http ./internal/processing -count=1` |
| Full suite command | `cd /home/vadim/diplom/core-backend && go test ./... -count=1 && cd /home/vadim/diplom/frontend && npm run lint && npm run build && npx tsc --noEmit` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| KSMI-01 | Aggregated profile is transformed into stable `decision_input` and queued/persisted for delivery | unit/integration | `cd /home/vadim/diplom/core-backend && go test ./internal/decision ./internal/http -run 'TestCreateDecisionInput|TestDecisionPendingTransition' -count=1` | ❌ Wave 0 |
| KSMI-02 | WiMi delivery retries only transport failures, persists correlation ID and attempt history | unit/integration | `cd /home/vadim/diplom/core-backend && go test ./internal/decision ./internal/kesmi -run 'TestRetriesOnlyTransportFailures|TestPersistsCorrelationID' -count=1` | ❌ Wave 0 |
| KSMI-03 | Result endpoint exposes normalized recommendation/diagnostics without raw WiMi coupling | API/unit | `cd /home/vadim/diplom/core-backend && go test ./internal/http -run 'TestExaminationResultIncludesDecisionBlock|TestDecisionFailureDiagnostics' -count=1` | ❌ Wave 0 |
| RSLT-02 | Operator result screen shows decision, metrics, baseline, and channel contributions together | build/manual-smoke | `cd /home/vadim/diplom/frontend && npm run lint && npm run build && npx tsc --noEmit` | ⚠️ Existing build only |

### Sampling Rate
- **Per task commit:** `cd /home/vadim/diplom/core-backend && go test ./internal/decision ./internal/kesmi ./internal/http -count=1`
- **Per wave merge:** `cd /home/vadim/diplom/core-backend && go test ./... -count=1 && cd /home/vadim/diplom/frontend && npm run lint && npm run build && npx tsc --noEmit`
- **Phase gate:** Full suite green plus Docker-based WiMi/core smoke when Docker is available

### Wave 0 Gaps
- [ ] `core-backend/internal/decision/service_test.go` — pins `decision_input`, `decision_pending`, and `completed` behavior
- [ ] `core-backend/internal/kesmi/client_test.go` — pins WiMi transport/business error classification
- [ ] `core-backend/internal/http/results_handler_test.go` additions — verifies decision block in result DTO and honest placeholder mode
- [ ] Frontend result-page assertions or browser smoke notes for decision diagnostics — current frontend has build coverage only
- [ ] Docker-based WiMi smoke harness — required because Phase 4 runtime contract includes WiMi in Compose

## Sources

### Primary (HIGH confidence)
- `/home/vadim/diplom/docs/00_project.md` - Phase 4 workflow, statuses `decision_pending` / `completed`, KESMI integration constraints, config requirements
- `/home/vadim/diplom/docs/01_contract.md` - Current Phase 3 result/history contracts and explicit Phase 4 gap
- `/home/vadim/diplom/docs/04_wimi_guide.md` - Project-specific WiMi deployment and integration guidance
- `/home/vadim/diplom/wimi-server/doc/REST API/Методы/POST.md` - Official local WiMi POST methods (`/ModelCalc`, `/ModelsParametersInfo`)
- `/home/vadim/diplom/wimi-server/doc/REST API/Методы/GET.md` - Official local WiMi `GET /Models` behavior
- `/home/vadim/diplom/wimi-server/doc/REST API/Классификация-состояний.md` - Official local WiMi error families and error IDs
- `/home/vadim/diplom/core-backend/internal/processing/publisher.go` - Existing persisted relay/retry pattern
- `/home/vadim/diplom/core-backend/internal/channelresults/service.go` - Existing temporary vs fatal failure pattern
- `/home/vadim/diplom/core-backend/internal/results/repository.go` - Existing canonical result/history projection surface
- `/home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/results/page.tsx` - Existing operator result UI reuse point

### Secondary (MEDIUM confidence)
- `npm view next version time.modified` - latest registry version observed on 2026-03-23
- `npm view @tanstack/react-query version time.modified` - latest registry version observed on 2026-03-23
- `npm view zod version time.modified` - latest registry version observed on 2026-03-23
- `/home/vadim/diplom/README.md` - current local validation workflow and environment expectations

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: MEDIUM - core stack is clear, but Go registry latest versions could not be verified in this environment and WiMi runtime packaging is not finalized
- Architecture: HIGH - existing codebase patterns and project docs strongly constrain the correct Phase 4 shape
- Pitfalls: HIGH - supported by current WiMi docs, Phase 3 contracts, and existing retry/state patterns

**Research date:** 2026-03-23
**Valid until:** 2026-04-22
