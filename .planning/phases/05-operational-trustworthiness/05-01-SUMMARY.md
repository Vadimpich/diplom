# 05-01 Summary

## Completed

- Published the Phase 5 operational contract in `docs/01_contract.md`:
  - append-only `audit_event` schema
  - required audit event types
  - `/health`, `/ready`, `/metrics` semantics for `core-backend`, `frontend`, workers, and `ml-baseline`
  - low-cardinality metric rules
  - `request_id` + `traceparent` propagation rules across HTTP, RabbitMQ, baseline, and WiMi
- Added `.planning/phases/05-operational-trustworthiness/05-VALIDATION.md` as the canonical Phase 5 validation checklist.
- Added RED scaffolds:
  - `core-backend/internal/audit/audit_test.go`
  - `core-backend/internal/http/observability_test.go`
  - trace-propagation expectations in `core-backend/internal/processing/publisher_test.go`
  - trace-propagation expectations in `core-backend/internal/channelresults/service_test.go`
  - trace-propagation expectations in `core-backend/internal/decision/relay_test.go`
  - Python observability tests for `ml-text`, `ml-acoustic`, `ml-paralinguistic`, `ml-baseline`

## Verification

- `rg -n 'audit_event|/ready|/metrics|traceparent|request_id|processing.launch|decision.failed' docs/01_contract.md .planning/phases/05-operational-trustworthiness/05-VALIDATION.md`
- `go test ./internal/audit ./internal/http ./internal/processing ./internal/channelresults ./internal/decision -run 'TestAudit|TestReadiness|TestMetrics|TestTrace|TestCorrelation' -count=1`
- `pytest -q tests/test_observability.py` in Python service directories

## Result

RED state confirmed:
- `core-backend` has no `audit` package yet.
- processing and decision DTOs do not yet carry `request_id` / `traceparent`.
- `/ready` and `/metrics` do not exist yet.
- Python services do not expose `/ready` / `/metrics` yet.
- shell-level `pytest` is not on PATH in this WSL session, so Python verification must later run via per-service virtualenv or container path.
