# Phase 5: Operational Trustworthiness - Research

**Researched:** 2026-03-23
**Domain:** audit trail, observability, trace correlation, and workflow trustworthiness for a brownfield Go + Next.js + FastAPI system
**Confidence:** HIGH

## Summary

The repo already has the hard part of the domain model: PostgreSQL is the source of truth for workflow state, async processing already carries `correlation_id`, decision delivery persists attempt history, and the codebase has a useful pattern of contract-first RED tests before each implementation slice. Phase 5 therefore should not redesign the workflow. It should add the operational surfaces that make the existing workflow explainable and safe to run: durable audit events, structured and correlated logs, split health/readiness/metrics endpoints, and targeted regression coverage for the workflow transitions that matter in production.

The current gap is operational, not architectural. `core-backend` has only a coarse `GET /health`, `chi` `RequestID`, and plain `log.Printf`; the Python workers expose `/health` with an internal `consumer_ready` flag but no `/ready`, no `/metrics`, and no trace/log propagation. `docs/01_contract.md` does not yet define observability contracts or an audit event schema. Test coverage is strong for individual modules, but there is no phase-wide contract that pins audit side effects, readiness degradation rules, or cross-service correlation.

**Primary recommendation:** plan Phase 5 in this order: `contracts + RED scaffolds -> core audit persistence -> shared correlation/logging plumbing -> readiness/metrics endpoints + minimal compose observability -> workflow regression hardening + docs`.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| OBSV-01 | System keeps an audit trail for login, critical admin actions, examination creation/finish, processing launch, and result receipt | Use a dedicated PostgreSQL-backed audit layer in `core-backend`, append audit rows in the same transaction as the source-of-truth state change where possible, and define a versioned internal audit event schema in `docs/01_contract.md` |
| OBSV-02 | All services publish health/readiness and basic metrics for HTTP, DB, queues, S3, and errors | Add `/health`, `/ready`, and `/metrics` contracts per service; use low-cardinality Prometheus metrics and dependency-aware readiness checks; verify with compose smoke |
| OBSV-03 | Logs and traces correlate by request ID or trace ID across the flow | Standardize on trace context plus request ID, propagate through HTTP, RabbitMQ envelopes, and KESMI calls, and emit structured JSON logs with trace/request/examination fields |
| QUAL-01 | All HTTP and inter-service contracts are fixed and versioned in `docs/01_contract.md` | Publish observability endpoint payloads, audit event schema, metric naming conventions, and correlation rules before implementation |
| QUAL-02 | Critical workflow parts are covered by unit/integration/API tests including transitions, idempotency, and failures | Add RED scaffolds for audit emission, readiness degradation, result-receipt traces, and status/idempotency regressions around finish, retries, aggregation, and decision delivery |
</phase_requirements>

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `go.opentelemetry.io/otel` | `v1.42.0` | Trace context propagation, tracer setup, resource attributes in Go | Official OpenTelemetry Go SDK; matches the project architecture in `docs/00_project.md` and supports OTLP exporter env config |
| `opentelemetry-sdk` | `1.40.0` | Python tracing SDK for FastAPI workers | Official Python SDK for manual/code-based instrumentation on Python 3.12 services |
| `opentelemetry-exporter-otlp` | `1.40.0` | OTLP export from Python services to Collector | Official OTLP path; keeps transport uniform across Go and Python |
| `opentelemetry-instrumentation-fastapi` | `0.61b0` | FastAPI HTTP span instrumentation | Avoids hand-rolled request span middleware in the ML services |
| `prometheus-client` | `0.24.1` | `/metrics` exposition for Python services | Official Prometheus Python client with standard Counter/Gauge/Histogram types |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| OpenTelemetry Collector Contrib | Verify exact tag before implementation | Central OTLP ingest and export fan-out | Use when Phase 5 needs one shared local telemetry sink instead of direct exporters from each service |
| Prometheus | Existing stack addition; verify exact image tag before implementation | Scrape `/metrics` endpoints | Use for local smoke and service-level operational verification |
| `zap` or `zerolog` | Pick one in Wave 1 and keep it repo-wide | Structured JSON application logs in Go | Replace `log.Printf`; do not mix logging styles |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Code-based OpenTelemetry in app code | eBPF/zero-code instrumentation | Official OBI docs explicitly note manual instrumentation is still recommended for custom attributes and events; Phase 5 needs custom `examination_id`, actor, and audit semantics |
| Minimal collector + Prometheus | Full Prometheus + Grafana + Loki + Tempo stack | Full stack is closer to target architecture, but it is larger scope than the phase requirements strictly require |
| Dedicated admin audit UI now | Backend-only audit persistence plus optional read API | UI is useful later, but Phase 5 can satisfy trustworthiness with durable storage, docs, and tests first |

**Installation:**

```bash
cd /home/vadim/diplom/core-backend
go get go.opentelemetry.io/otel@v1.42.0

cd /home/vadim/diplom/ml-services/ml-text
./.venv/bin/pip install opentelemetry-sdk==1.40.0 opentelemetry-exporter-otlp==1.40.0 opentelemetry-instrumentation-fastapi==0.61b0 prometheus-client==0.24.1
```

**Version verification:** verified on 2026-03-23 from the module proxy / registries.

- `go.opentelemetry.io/otel` -> `v1.42.0` published `2026-03-06T19:13:23Z`
- `opentelemetry-sdk` -> `1.40.0` uploaded `2026-03-04T14:17:17Z`
- `opentelemetry-exporter-otlp` -> `1.40.0` uploaded `2026-03-04T14:17:03Z`
- `opentelemetry-instrumentation-fastapi` -> `0.61b0` uploaded `2026-03-04T14:19:30Z`
- `prometheus-client` -> `0.24.1` uploaded `2026-01-14T15:26:24Z`

## Architecture Patterns

### Recommended Project Structure

```text
core-backend/
├── internal/
│   ├── audit/            # event schema, repository, append helpers
│   ├── observability/    # tracer setup, logger setup, metrics registry, propagation
│   ├── http/             # request/trace middleware, health/readiness/metrics handlers
│   └── ...
ml-services/
├── ml-text/app/
│   ├── telemetry.py      # FastAPI instrumentation, metrics, shared log fields
│   └── main.py
├── ml-acoustic/app/
├── ml-paralinguistic/app/
└── ml-baseline/app/
```

### Pattern 1: Transactional Audit Append At The Source-Of-Truth Boundary

**What:** write audit rows from `core-backend` when the authoritative state change is persisted, not from derived logs or downstream consumers.

**When to use:** auth events, user/admin mutations, examination creation/start/finish, processing launch fence, channel result receipt, aggregation completion, and decision result persistence.

**Plan-shaping rule:** for events backed by a PostgreSQL transaction, the audit append must happen in the same transaction as the domain write. For events that represent failed auth or transport-level failures without a domain transaction, append immediately with explicit `outcome=failed`.

**Example:**

```go
// Source: project pattern derived from PostgreSQL-as-source-of-truth + OTel trace context docs
type AuditEvent struct {
    EventType      string
    ActorUserID    *int64
    ExaminationID  *int64
    SpecialistID   *int64
    RequestID      string
    TraceID        string
    Outcome        string
    Payload        []byte
    HappenedAt     time.Time
}

func (s *Service) FinishExamination(ctx context.Context, examID int64) error {
    return s.repo.WithTx(ctx, func(tx Tx) error {
        exam, err := tx.FinishLaunch(ctx, examID)
        if err != nil {
            return err
        }
        return tx.AppendAudit(ctx, AuditEvent{
            EventType:     "examination.finish_requested",
            ExaminationID: &exam.ID,
            SpecialistID:  &exam.SpecialistID,
            RequestID:     requestid.FromContext(ctx),
            TraceID:       trace.SpanFromContext(ctx).SpanContext().TraceID().String(),
            Outcome:       "succeeded",
            HappenedAt:    time.Now().UTC(),
        })
    })
}
```

### Pattern 2: One Correlation Model Across HTTP, RabbitMQ, And External Calls

**What:** every request gets both a request ID and a trace/span context; async envelopes must carry the correlation information required to continue the trace or at minimum link log records unambiguously.

**When to use:** all core HTTP handlers, outbox publish path, RabbitMQ result consumer, worker processing path, baseline HTTP call, and KESMI client.

**Plan-shaping rule:** keep `correlation_id` for business/debug continuity because it already exists in contracts, but make `traceparent` the canonical distributed trace transport. Do not replace existing `correlation_id`; augment the envelope.

### Pattern 3: Split Liveness/Health/Readiness By Dependency Role

**What:** health answers “is the process alive enough to answer”; readiness answers “can this service safely serve traffic/work now”.

**When to use:** every service in the compose stack.

**Recommended contract:**

- `/health`: process alive, no deep dependency fan-out, cheap enough for container healthcheck
- `/ready`: dependency-aware check
- `/metrics`: Prometheus exposition only, no JSON

**Service-specific readiness gates:**

- `core-backend`: PostgreSQL required; RabbitMQ, MinIO, baseline service, and WiMi should degrade readiness only for features that depend on them, but for this phase it is acceptable to mark global readiness failed if a required Phase 1-4 dependency is unavailable
- `text/acoustic/paralinguistic-worker`: RabbitMQ channel/consumer ready and MinIO client reachable
- `ml-baseline`: HTTP process up; no DB/RabbitMQ dependency
- `frontend`: app up; no need for deep backend dependency readiness in this phase

### Pattern 4: Observability Package, Not Per-Handler Ad Hoc Instrumentation

**What:** centralize logger/tracer/meter initialization and middleware so each module adds business attributes, not transport boilerplate.

**When to use:** immediately after contracts/Wave 0. This keeps later plans small and avoids touching every handler twice.

### Anti-Patterns to Avoid

- **Using application logs as audit trail:** logs are mutable, noisy, and retention differs; audit must be a dedicated persisted layer.
- **Recording metrics with `examination_id`, `specialist_id`, or `user_id` labels:** Prometheus labels are easily abused; keep metrics low-cardinality.
- **Treating `RequestID` as distributed tracing:** current `chi` request ID stops at the process boundary; Phase 5 needs propagation into RabbitMQ and outbound HTTP.
- **Emitting audit from workers for core-owned state changes:** this creates duplicates and drifts from PostgreSQL source of truth.
- **Making `/health` expensive:** container healthchecks must stay cheap; put dependency fan-in into `/ready`.

## Recommended Sequencing

1. **Wave 0: contracts and RED scaffolds**
   Publish audit schema, observability endpoint payloads, metric naming rules, and correlation rules in `docs/01_contract.md`. Add failing tests for audit append, readiness degradation, metrics endpoint availability, and correlation propagation.
2. **Wave 1: core backend audit + structured logging base**
   Add `internal/audit` and `internal/observability`, replace `log.Printf` request logging, and append durable audit rows for auth, admin/user, examination, processing launch, result receipt, and decision completion/failure.
3. **Wave 2: correlation propagation across async edges**
   Carry trace context and request metadata through outbox publish, result consume, baseline/KESMI clients, and worker processing; enrich JSON logs with `trace_id`, `request_id`, `examination_id`, `specialist_id`, and `event_type`.
4. **Wave 3: readiness + metrics + minimal compose observability**
   Add `/ready` and `/metrics` to all services, instrument HTTP durations/error counts and dependency health metrics, then add minimal Collector/Prometheus compose wiring and smoke docs.
5. **Wave 4: workflow regression hardening**
   Expand Go and Python tests for idempotency, retries, failure classification, audit side effects, and endpoint contracts; update `README.md` and `docs/02_implementation.md`.

## Acceptance Boundaries

### Must Be In Scope

- Durable audit persistence for the critical actions listed in `OBSV-01`
- Versioned observability contracts in `docs/01_contract.md`
- Structured, correlated logs in `core-backend` and Python services
- `/health`, `/ready`, and `/metrics` on all runtime services
- Automated tests for status transitions, idempotency, failure handling, and new observability contracts
- A compose-verifiable local path for metrics/traces collection, if Docker is available

### Should Stay Out Of Scope Unless A Requirement Forces It

- Full admin audit UI
- Production-grade dashboards and alerts for every service
- Long-term log retention policy implementation in Loki
- Reworking Phase 1-4 business logic that already passes
- Replacing existing `correlation_id` contracts with a new identifier scheme

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Distributed trace propagation | Custom headers or manual string concatenation across services | OpenTelemetry context propagation and OTLP exporters | Official trace context is standardized and interoperable |
| Metrics exposition | Homegrown JSON counters endpoint | Prometheus `/metrics` via official clients/exporters | Prometheus expects standard exposition semantics and low-cardinality usage |
| Audit storage | Free-form log scraping | Dedicated `audit_logs` table with typed columns + JSON payload | Audit needs queryability, retention separation, and stable semantics |
| Structured logging | `fmt`/`log.Printf` parsing | One JSON logger (`zap` or `zerolog`) with shared fields | Parsing plain strings later is brittle and breaks correlation |
| Readiness checks | Copy-paste handler logic in each service | Small shared health/readiness helpers per runtime | Keeps dependency semantics consistent and testable |

**Key insight:** the dangerous custom code in this phase is not the business logic; it is the instrumentation plumbing. Use official telemetry primitives and keep custom code focused on domain fields, audit semantics, and dependency gating.

## Common Pitfalls

### Pitfall 1: Duplicate Audit Events On Retry

**What goes wrong:** a retried HTTP request or consumer delivery writes the same semantic audit event multiple times.
**Why it happens:** audit append is tied to request arrival rather than the idempotent state transition fence.
**How to avoid:** append audit after the fenced domain write succeeds, and include a stable event key or reference to the underlying state-change row when needed.
**Warning signs:** multiple `processing.launch` or `examination.finish` audit rows for one examination without a matching state delta.

### Pitfall 2: Metrics Cardinality Explosion

**What goes wrong:** Prometheus becomes noisy or unusable because labels include identifiers such as examination, specialist, file path, or raw error text.
**Why it happens:** metrics are misused as logs.
**How to avoid:** keep identifiers in logs/spans/audit payloads; metrics should use bounded labels like `service`, `route`, `channel`, `status_class`, `dependency`.
**Warning signs:** a metric family where series count grows linearly with traffic or data volume.

### Pitfall 3: Broken Async Trace Continuity

**What goes wrong:** HTTP request spans exist, worker spans exist, but the trace graph breaks at RabbitMQ or outbound HTTP.
**Why it happens:** only `correlation_id` is propagated, not standardized trace context.
**How to avoid:** inject/extract W3C trace context through message headers or a documented equivalent field while preserving existing business correlation IDs.
**Warning signs:** logs show the same `correlation_id` but tracing backend shows unrelated traces.

### Pitfall 4: Overloading `/health`

**What goes wrong:** container healthchecks flap because `/health` performs deep checks on DB, broker, S3, and external services.
**Why it happens:** liveness and readiness semantics are collapsed.
**How to avoid:** keep `/health` cheap and move dependency fan-in to `/ready`.
**Warning signs:** container restarts during transient dependency blips even though the process itself is fine.

### Pitfall 5: Tests Pin Output Text Instead Of Contract Semantics

**What goes wrong:** tests become brittle and block harmless logging/message wording changes.
**Why it happens:** assertions target whole log lines instead of structured fields and behavior.
**How to avoid:** assert event types, IDs, states, counters, and payload fields.
**Warning signs:** observability tests fail after logger formatting changes with no functional regression.

## Code Examples

Verified patterns from official or project-authoritative sources:

### FastAPI HTTP Instrumentation

```python
# Source: https://opentelemetry-python-contrib.readthedocs.io/en/latest/instrumentation/fastapi/fastapi.html
from fastapi import FastAPI
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor

app = FastAPI()
FastAPIInstrumentor.instrument_app(app)
```

### Trace ID Extraction For Log Enrichment

```go
// Source: https://opentelemetry.io/docs/concepts/signals/traces/
spanCtx := trace.SpanFromContext(ctx).SpanContext()
if spanCtx.IsValid() {
    logger.Info("processing result received",
        zap.String("trace_id", spanCtx.TraceID().String()),
        zap.String("span_id", spanCtx.SpanID().String()),
    )
}
```

### Prometheus-Friendly Request Duration Metric

```go
// Source: Prometheus client-library guidance + project low-cardinality requirements
var requestDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "diplom_http_request_duration_seconds",
        Help: "HTTP request latency by route and status class.",
    },
    []string{"route", "method", "status_class"},
)
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Single coarse `GET /health` | Split `/health`, `/ready`, `/metrics` with dependency-aware readiness | Widely standardized by current platform/runtime practice | Prevents restart loops and makes ops checks meaningful |
| Plain text request logs + local request ID | Structured JSON logs enriched with trace/request/domain fields | Current OTel + JSON log practice | Makes cross-service debugging realistic |
| Per-module behavior tests only | Contract-first regression suites that pin workflow, audit, and observability semantics | Already established in Phases 2-4 and should continue here | Lets observability changes stay safe and reviewable |
| Zero-code-only observability hopes | Manual/code-based instrumentation for domain-specific attributes and events | Current OTel guidance still favors manual instrumentation for custom business attributes | Required because this project needs `examination_id`, `specialist_id`, actor, and audit semantics |

**Deprecated/outdated:**

- Parsing `log.Printf` strings as an operations interface: inadequate for auditability and trace correlation.
- Using `RequestID` alone as distributed tracing: useful locally, insufficient across RabbitMQ and outbound HTTP.

## Likely Work Breakdown

| Slice | Main Deliverable | Notes |
|------|------------------|-------|
| 00 | Phase 5 contracts + RED scaffold tests | Mirrors Phase 4 Wave 0 pattern; easiest way to keep scope honest |
| 01 | `core-backend` audit schema, repo, service hooks, and structured logger bootstrap | Highest leverage because core owns source-of-truth state |
| 02 | Trace/request propagation through RabbitMQ, baseline, KESMI, and Python workers | Finish correlation before adding dashboards so data is meaningful |
| 03 | `/ready` + `/metrics` across all services and minimal compose telemetry stack | Focus on operability, not dashboard polish |
| 04 | Regression coverage + README/docs implementation log refresh | Closes `QUAL-02` and keeps future phases safe |

## Risks

- Docker is unavailable in the current WSL environment, so compose-based observability smoke cannot be executed here without restoring Docker integration.
- Audit scope can sprawl into UI/reporting; keep Phase 5 centered on persistence, contracts, and queryability first.
- Introducing telemetry in every service can create noisy diffs; a shared observability package in Go and one repeated pattern in Python is essential.
- If trace propagation is added after metrics/logging, planners may end up re-touching the same files twice. Sequence correlation before endpoint polish.

## Open Questions

1. **Should Phase 5 include an admin read API for audit logs?**
   - What we know: `docs/00_project.md` says audit access should be admin-restricted, but Phase 5 requirements only demand the trail itself.
   - What's unclear: whether “auditable enough to operate” requires a backend read surface now or only durable storage.
   - Recommendation: keep write-path mandatory; make a minimal paginated admin read endpoint optional if planner needs a concrete operability surface.

2. **How much observability infrastructure should land in this phase?**
   - What we know: target architecture recommends Prometheus/Grafana/Loki/Tempo or Jaeger.
   - What's unclear: whether full dashboards/log retention are required for milestone acceptance.
   - Recommendation: minimum accepted scope is Collector + Prometheus or equivalent scrape/export path; treat Grafana/Loki/Tempo dashboards as stretch unless explicitly required by the planner.

3. **Should frontend admin monitoring be expanded in Phase 5?**
   - What we know: current admin page only reads `GET /health`.
   - What's unclear: whether user-facing operational screens are needed before the milestone closes.
   - Recommendation: do not make frontend expansion a blocker; backend observability contracts and tests have higher value for this phase.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | core-backend tests and new telemetry deps | ✓ | `go1.26.0` | — |
| Node.js | frontend verification and docs tooling | ✓ | `v22.22.1` | — |
| npm | frontend verification | ✓ | `11.12.0` | — |
| Python 3 | ML service instrumentation/tests | ✓ | `3.12.3` | — |
| `pytest` | ML baseline regression and possible worker tests | ✓ | `9.0.2` via `ml-baseline/.venv` | use project `.venv`, not global install |
| Docker / Docker Compose | compose smoke for observability stack | ✗ | — | none for real runtime smoke |
| `curl` | endpoint smoke checks | ✓ | `8.5.0` | `wget` where already present in containers |

**Missing dependencies with no fallback:**

- Docker Desktop / Docker CLI integration in this WSL distro. This blocks real compose verification for Prometheus/Collector and any end-to-end runtime smoke.

**Missing dependencies with fallback:**

- Global `pytest` is absent, but the tracked project virtual environments under `ml-services/*/.venv` are available.

## Sources

### Primary (HIGH confidence)

- Local project sources:
  - `docs/00_project.md` - target architecture, logging/audit requirements, recommended observability stack
  - `docs/01_contract.md` - current HTTP/AMQP contracts and present observability gaps
  - `docs/02_implementation.md` - implemented phases and validation history
  - `core-backend/internal/http/router.go`, `core-backend/internal/http/health.go` - current core runtime surface
  - `ml-services/ml-text/app/main.py` - current worker health/logging pattern representative of text/acoustic/paralinguistic services
- OpenTelemetry Go exporters docs: https://opentelemetry.io/docs/languages/go/exporters/
- OpenTelemetry traces concepts: https://opentelemetry.io/docs/concepts/signals/traces/
- Prometheus client-library guidance: https://prometheus.io/docs/instrumenting/writing_clientlibs/
- FastAPI instrumentation docs: https://opentelemetry-python-contrib.readthedocs.io/en/latest/instrumentation/fastapi/fastapi.html
- PyPI package pages / registry metadata:
  - https://pypi.org/project/opentelemetry-sdk/
  - https://pypi.org/project/opentelemetry-exporter-otlp/
  - https://pypi.org/project/opentelemetry-instrumentation-fastapi/
  - https://pypi.org/project/prometheus-client/
  - https://proxy.golang.org/go.opentelemetry.io/otel/@latest

### Secondary (MEDIUM confidence)

- OpenTelemetry Collector deployment patterns: https://opentelemetry.io/docs/collector/deploy/
- OpenTelemetry zero-code / OBI docs: https://opentelemetry.io/docs/zero-code/obi/

### Tertiary (LOW confidence)

- None.

## Metadata

**Confidence breakdown:**

- Standard stack: HIGH - official docs and registry metadata were available for the recommended telemetry packages; collector image version remains a verify-at-implementation item
- Architecture: HIGH - recommendations are constrained by the repo’s existing PostgreSQL source-of-truth workflow and by `docs/00_project.md`
- Pitfalls: HIGH - directly derived from current brownfield gaps plus official Prometheus/OpenTelemetry guidance

**Research date:** 2026-03-23
**Valid until:** 2026-04-22
