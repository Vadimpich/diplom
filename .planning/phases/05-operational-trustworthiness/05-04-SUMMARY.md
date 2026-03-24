# 05-04 Summary

## Completed

- `core-backend` runtime endpoints:
  - cheap `GET /health`
  - dependency-aware `GET /ready`
  - Prometheus `GET /metrics`
- Added metrics registry in `core-backend/internal/observability/metrics.go`.
- Added frontend runtime endpoints:
  - `frontend/app/api/health/route.ts`
  - `frontend/app/api/ready/route.ts`
  - `frontend/app/api/metrics/route.ts`
- Updated admin monitoring UI to the new health contract.
- Added Python observability surfaces:
  - worker `telemetry.py` helpers for `ml-text`, `ml-acoustic`, `ml-paralinguistic`
  - `/ready` and `/metrics` in all workers
  - `/ready` and `/metrics` in `ml-baseline`
- Added Prometheus local scrape path:
  - `ops/prometheus/prometheus.yml`
  - `prometheus` service in `docker-compose.yml`
- Updated local runtime env/docs:
  - `.env.example`
  - `README.md`

## Verification

- `cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/observability -run 'TestHealthEndpointIsCheap|TestReadinessDegradesOnDependencyFailure|TestMetricsEndpointExposesLowCardinalityFamilies' -count=1`
- `cd /home/vadim/diplom/ml-services/ml-baseline && ./.venv/bin/pytest -q tests/test_observability.py`
- `cd /home/vadim/diplom/ml-services/ml-text && ./.venv/bin/pytest -q tests/test_observability.py`
- `cd /home/vadim/diplom/ml-services/ml-acoustic && ./.venv/bin/pytest -q tests/test_observability.py`
- `cd /home/vadim/diplom/ml-services/ml-paralinguistic && ./.venv/bin/pytest -q tests/test_observability.py`
- `cd /home/vadim/diplom/frontend && npm run lint && npx tsc --noEmit && npm run build`
- `cd /home/vadim/diplom && docker compose config`
- `cd /home/vadim/diplom && docker compose up -d --build frontend core-backend text-worker acoustic-worker paralinguistic-worker ml-baseline wimi prometheus`
- `curl -fsS http://localhost:3000/api/health`
- `curl -fsS http://localhost:3000/api/ready`
- `curl -fsS http://localhost:3000/api/metrics`
- `curl -fsS http://localhost:8080/health`
- `curl -fsS http://localhost:8080/ready`
- `curl -fsS http://localhost:8080/metrics`
- `curl -fsS http://localhost:9090/-/healthy`
- `curl -fsS http://localhost:9090/api/v1/targets`

## Result

The shipped runtime now exposes explicit health/readiness/metrics surfaces across frontend, core backend, and Python services, and the real compose smoke confirms healthy frontend/core readiness plus active Prometheus scraping across core/frontend/workers/baseline. During smoke, a false-negative readiness issue was found and fixed by aligning `BASELINE_SERVICE_URL` with the actual `ml-baseline:8080` runtime endpoint.
