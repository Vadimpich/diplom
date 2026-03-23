# Phase 05 Verification

## Scope

This artifact is the canonical verification record for Phase 5 operational trustworthiness. It lifts the exact commands from `.planning/phases/05-operational-trustworthiness/05-VALIDATION.md` and anchors them to the requirements and contracts that Phase 5 was intended to satisfy.

## Evidence Sources

- `.planning/phases/05-operational-trustworthiness/05-VALIDATION.md`
- `.planning/phases/05-operational-trustworthiness/05-01-SUMMARY.md`
- `.planning/phases/05-operational-trustworthiness/05-02-SUMMARY.md`
- `.planning/phases/05-operational-trustworthiness/05-03-SUMMARY.md`
- `.planning/phases/05-operational-trustworthiness/05-04-SUMMARY.md`
- `.planning/phases/05-operational-trustworthiness/05-05-SUMMARY.md`
- `.planning/phases/06-operator-result-reentry-and-metrics-truthfulness/06-02-SUMMARY.md`
- `.planning/phases/06-operator-result-reentry-and-metrics-truthfulness/06-03-SUMMARY.md`
- `.planning/REQUIREMENTS.md`
- `.planning/v1.0-v1.0-MILESTONE-AUDIT.md`
- `docs/00_project.md`
- `docs/01_contract.md`

## Requirement Matrix

| Requirement | Status | Evidence owner | Command-backed evidence | Notes |
|-------------|--------|----------------|-------------------------|-------|
| `OBSV-01` | Verified | Phase 5 | `cd /home/vadim/diplom/core-backend && go test ./internal/audit -run 'TestAppendAuditEvent|TestAppendAuditEventUsesStableEventKey|TestListAuditEvents' -count=1` | Confirms append-only audit storage and stable `audit_event` behavior for core-owned source-of-truth transitions. |
| `OBSV-01` critical transition coverage | Verified | Phase 5 | `cd /home/vadim/diplom/core-backend && go test ./internal/auth ./internal/questionnaires ./internal/examinations ./internal/processing ./internal/channelresults ./internal/decision ./internal/audit -run 'TestLoginWritesAuditEvent|TestFailedLoginWritesAuditEvent|TestUserMutationWritesAuditEvent|TestQuestionnaireMutationWritesAuditEvent|TestExaminationFinishWritesSingleAuditEvent|TestProcessingLaunchWritesAuditEvent|TestResultReceiptWritesAuditEvent|TestDecisionTerminalStateWritesAuditEvent' -count=1` | Proves that audit events are emitted from the backend boundaries listed in `docs/01_contract.md`, including `processing.launch`, `processing.result_received`, `decision.completed`, and `decision.failed`. |
| `OBSV-03` | Verified | Phase 5 | `cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/observability ./internal/processing ./internal/channelresults ./internal/decision ./internal/baselineclient ./internal/kesmi -run 'TestRequestLoggingEmitsStructuredFields|TestRequestContextIncludesTraceAndRequestIDs|TestTraceContextMiddleware|TestPublisherPreservesTraceContext|TestResultsConsumerContinuesTrace|TestBaselineClientPropagatesTraceContext|TestKESMIClientPropagatesTraceContext' -count=1` | Confirms `request_id`, `traceparent`, and correlated trace continuity across HTTP, RabbitMQ, baseline, and WiMi/KESMI. |
| `QUAL-01` | Verified | Phase 5 | `cd /home/vadim/diplom && rg -n 'audit_event|traceparent|request_id|/ready|/metrics|processing.launch|decision.failed' docs/01_contract.md .planning/phases/05-operational-trustworthiness/05-VALIDATION.md` | Proves `docs/01_contract.md` is the versioned source of truth for audit, readiness, metrics, and propagation contracts used by the code and tests. |
| `QUAL-02` | Verified | Phase 5 | `cd /home/vadim/diplom/core-backend && go test ./internal/auth ./internal/examinations ./internal/processing ./internal/channelresults ./internal/aggregation ./internal/decision ./internal/http -count=1` | Final regression command from `05-VALIDATION.md` confirms critical workflow coverage around status transitions, idempotency fences, and error handling. |
| `OBSV-02` | Cross-phase note | Phase 5 + Phase 6 | `cd /home/vadim/diplom/frontend && npm run test -- --run lib/operator/examination-navigation.test.ts lib/server/core-readiness.test.ts` | Phase 5 shipped the frontend `/api/ready` and `/api/metrics` surfaces; Phase 6 closed the truthful `core_backend` dependency gauge gap recorded in the milestone audit. |

## Exact Evidence Commands

### Contracts and requirement anchors

```bash
cd /home/vadim/diplom
rg -n '05-VALIDATION.md|audit_event|traceparent|request_id|docs/01_contract.md|/ready|/metrics|processing.launch|decision.failed' .planning/phases/05-operational-trustworthiness/05-VERIFICATION.md docs/01_contract.md .planning/phases/05-operational-trustworthiness/05-VALIDATION.md
```

This is the index command for the shipped operational contract surface:

- append-only `audit_event`;
- propagated `traceparent` and `request_id`;
- runtime `/ready` and `/metrics` contracts;
- the expectation that `docs/01_contract.md` remains authoritative.

### Audit trail evidence

```bash
cd /home/vadim/diplom/core-backend
sqlc generate
go test ./internal/audit -run 'TestAppendAuditEvent|TestAppendAuditEventUsesStableEventKey|TestListAuditEvents' -count=1
go test ./internal/auth ./internal/questionnaires ./internal/examinations ./internal/processing ./internal/channelresults ./internal/decision ./internal/audit -run 'TestLoginWritesAuditEvent|TestFailedLoginWritesAuditEvent|TestUserMutationWritesAuditEvent|TestQuestionnaireMutationWritesAuditEvent|TestExaminationFinishWritesSingleAuditEvent|TestProcessingLaunchWritesAuditEvent|TestResultReceiptWritesAuditEvent|TestDecisionTerminalStateWritesAuditEvent' -count=1
```

These commands are the direct proof for `OBSV-01`. They match the required `audit_event` contract in `docs/01_contract.md` and the Phase 5 summaries that describe source-of-truth audit emission.

### Trace propagation and structured logging evidence

```bash
cd /home/vadim/diplom/core-backend
go test ./internal/http ./internal/observability -run 'TestRequestLoggingEmitsStructuredFields|TestRequestContextIncludesTraceAndRequestIDs|TestTraceContextMiddleware' -count=1
go test ./internal/http ./internal/observability ./internal/processing ./internal/channelresults ./internal/decision ./internal/baselineclient ./internal/kesmi -run 'TestRequestLoggingEmitsStructuredFields|TestRequestContextIncludesTraceAndRequestIDs|TestTraceContextMiddleware|TestPublisherPreservesTraceContext|TestResultsConsumerContinuesTrace|TestBaselineClientPropagatesTraceContext|TestKESMIClientPropagatesTraceContext' -count=1
```

These commands are the direct proof for `OBSV-03`. They verify that one examination flow can be reconstructed through `request_id`, `traceparent`, and low-sensitivity correlation data.

### Runtime endpoints, metrics, and service evidence

```bash
cd /home/vadim/diplom/core-backend
go test ./internal/http ./internal/observability -run 'TestHealthEndpointIsCheap|TestReadinessDegradesOnDependencyFailure|TestMetricsEndpointExposesLowCardinalityFamilies' -count=1
```

```bash
cd /home/vadim/diplom/ml-services/ml-baseline
./.venv/bin/pytest -q tests/test_observability.py
cd /home/vadim/diplom/ml-services/ml-text
./.venv/bin/pytest -q tests/test_observability.py
cd /home/vadim/diplom/ml-services/ml-acoustic
./.venv/bin/pytest -q tests/test_observability.py
cd /home/vadim/diplom/ml-services/ml-paralinguistic
./.venv/bin/pytest -q tests/test_observability.py
```

```bash
cd /home/vadim/diplom/frontend
npm run lint
npx tsc --noEmit
npm run build
```

```bash
cd /home/vadim/diplom
docker compose up -d --build frontend core-backend text-worker acoustic-worker paralinguistic-worker ml-baseline wimi prometheus
curl -fsS http://localhost:3000/api/health >/dev/null
curl -fsS http://localhost:3000/api/ready >/dev/null
curl -fsS http://localhost:3000/api/metrics >/dev/null
curl -fsS http://localhost:8080/health >/dev/null
curl -fsS http://localhost:8080/ready >/dev/null
curl -fsS http://localhost:8080/metrics >/dev/null
docker compose exec text-worker python -c "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8080/ready', timeout=3).read(); urllib.request.urlopen('http://127.0.0.1:8080/metrics', timeout=3).read()"
docker compose exec acoustic-worker python -c "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8080/ready', timeout=3).read(); urllib.request.urlopen('http://127.0.0.1:8080/metrics', timeout=3).read()"
docker compose exec paralinguistic-worker python -c "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8080/ready', timeout=3).read(); urllib.request.urlopen('http://127.0.0.1:8080/metrics', timeout=3).read()"
docker compose exec ml-baseline python -c "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8080/ready', timeout=3).read(); urllib.request.urlopen('http://127.0.0.1:8080/metrics', timeout=3).read()"
curl -fsS http://localhost:9090/-/healthy >/dev/null
curl -fsS 'http://localhost:9090/api/v1/targets' | rg 'core-backend|frontend|text-worker|acoustic-worker|paralinguistic-worker|ml-baseline'
```

These commands are the operational proof that the Phase 5 runtime surfaces exist, follow the contract, and are runnable from the repository exactly as documented.

### Final regression evidence

```bash
cd /home/vadim/diplom/core-backend
go test ./internal/auth ./internal/examinations ./internal/processing ./internal/channelresults ./internal/aggregation ./internal/decision ./internal/http -count=1
```

This is the canonical `QUAL-02` regression command from `05-VALIDATION.md`.

## Cross-Phase Note For OBSV-02

Phase 5 shipped the frontend readiness and metrics surfaces and documented them in `docs/01_contract.md`, `05-VALIDATION.md`, and `05-04-SUMMARY.md`. The milestone audit later found a remaining truthfulness gap: frontend `/api/metrics` still hardcoded the `core_backend` dependency gauge.

That specific gap is closed in Phase 6 and must be read alongside this Phase 5 artifact:

- `.planning/phases/06-operator-result-reentry-and-metrics-truthfulness/06-02-SUMMARY.md` records the shared readiness probe and live `diplom_frontend_dependency_up{dependency="core_backend"}` behavior.
- `.planning/phases/06-operator-result-reentry-and-metrics-truthfulness/06-03-SUMMARY.md` records the focused regression test coverage and validation artifact.

Supporting command:

```bash
cd /home/vadim/diplom/frontend
npm run test -- --run lib/operator/examination-navigation.test.ts lib/server/core-readiness.test.ts
```

This does not weaken Phase 5 ownership. It clarifies the audit boundary:

- Phase 5 owns the contract and the shipped runtime surfaces.
- Phase 6 closes the later-discovered frontend dependency-truthfulness gap for `OBSV-02`.

## Verdict

Phase 5 now has durable, command-backed evidence for the requirements the milestone audit could not previously satisfy:

- `OBSV-01` is verified from append-only `audit_event` storage plus transition-level audit tests.
- `OBSV-03` is verified from structured logging and distributed trace propagation tests spanning HTTP, RabbitMQ, baseline, and WiMi/KESMI.
- `QUAL-01` is verified because `docs/01_contract.md` is explicitly referenced and grep-verifiable as the source of truth for these contracts.
- `QUAL-02` is verified from the documented repository regression commands across backend, Python services, frontend build checks, and compose smoke.

The missing-file audit blocker for Phase 5 is removed because operational trustworthiness now has a canonical `VERIFICATION.md` that points to exact commands instead of relying on summaries alone.
