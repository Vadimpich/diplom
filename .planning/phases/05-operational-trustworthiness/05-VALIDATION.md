# Phase 05 Validation

## Contract and RED scaffolds

```bash
cd /home/vadim/diplom
rg -n 'audit_event|/ready|/metrics|traceparent|request_id|processing.launch|decision.failed' docs/01_contract.md .planning/phases/05-operational-trustworthiness/05-VALIDATION.md
cd /home/vadim/diplom/core-backend
go test ./internal/audit ./internal/http ./internal/processing ./internal/channelresults ./internal/decision -run 'TestAudit|TestReadiness|TestMetrics|TestTrace|TestCorrelation' -count=1
cd /home/vadim/diplom/ml-services/ml-baseline
./.venv/bin/pytest -q tests/test_observability.py
cd /home/vadim/diplom/ml-services/ml-text
./.venv/bin/pytest -q tests/test_observability.py
cd /home/vadim/diplom/ml-services/ml-acoustic
./.venv/bin/pytest -q tests/test_observability.py
cd /home/vadim/diplom/ml-services/ml-paralinguistic
./.venv/bin/pytest -q tests/test_observability.py
```

## Audit implementation

```bash
cd /home/vadim/diplom/core-backend
sqlc generate
go test ./internal/audit -run 'TestAppendAuditEvent|TestAppendAuditEventUsesStableEventKey|TestListAuditEvents' -count=1
go test ./internal/auth ./internal/questionnaires ./internal/examinations ./internal/processing ./internal/channelresults ./internal/decision ./internal/audit -run 'TestLoginWritesAuditEvent|TestFailedLoginWritesAuditEvent|TestUserMutationWritesAuditEvent|TestQuestionnaireMutationWritesAuditEvent|TestExaminationFinishWritesSingleAuditEvent|TestProcessingLaunchWritesAuditEvent|TestResultReceiptWritesAuditEvent|TestDecisionTerminalStateWritesAuditEvent' -count=1
```

## Correlation and logging

```bash
cd /home/vadim/diplom/core-backend
go test ./internal/http ./internal/observability -run 'TestRequestLoggingEmitsStructuredFields|TestRequestContextIncludesTraceAndRequestIDs|TestTraceContextMiddleware' -count=1
go test ./internal/http ./internal/observability ./internal/processing ./internal/channelresults ./internal/decision ./internal/baselineclient ./internal/kesmi -run 'TestRequestLoggingEmitsStructuredFields|TestRequestContextIncludesTraceAndRequestIDs|TestTraceContextMiddleware|TestPublisherPreservesTraceContext|TestResultsConsumerContinuesTrace|TestBaselineClientPropagatesTraceContext|TestKESMIClientPropagatesTraceContext' -count=1
```

## Runtime endpoints and metrics

```bash
cd /home/vadim/diplom/core-backend
go test ./internal/http ./internal/observability -run 'TestHealthEndpointIsCheap|TestReadinessDegradesOnDependencyFailure|TestMetricsEndpointExposesLowCardinalityFamilies' -count=1
cd /home/vadim/diplom/ml-services/ml-baseline
./.venv/bin/pytest -q tests/test_observability.py
cd /home/vadim/diplom/ml-services/ml-text
./.venv/bin/pytest -q tests/test_observability.py
cd /home/vadim/diplom/ml-services/ml-acoustic
./.venv/bin/pytest -q tests/test_observability.py
cd /home/vadim/diplom/ml-services/ml-paralinguistic
./.venv/bin/pytest -q tests/test_observability.py
cd /home/vadim/diplom/frontend
npm run lint
npx tsc --noEmit
npm run build
```

## Compose smoke

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

## Final regression

```bash
cd /home/vadim/diplom/core-backend
go test ./internal/auth ./internal/examinations ./internal/processing ./internal/channelresults ./internal/aggregation ./internal/decision ./internal/http -count=1
```
