---
phase: 03-aggregated-baseline-aware-profiles
plan: 03
type: execute
wave: 2
depends_on:
  - 03-01
files_modified:
  - ml-services/ml-baseline/app/main.py
  - ml-services/ml-baseline/app/schemas.py
  - ml-services/ml-baseline/app/algorithms.py
  - ml-services/ml-baseline/tests/test_service.py
  - ml-services/ml-baseline/tests/test_algorithms.py
  - ml-services/ml-baseline/requirements.txt
  - docker-compose.yml
  - .env.example
autonomous: true
requirements:
  - BASE-01
  - BASE-02
  - BASE-03
must_haves:
  truths:
    - Baseline deviation is computed by a separate Python service boundary rather than inside Go core.
    - The baseline response includes both general and personal deviation plus update metadata and algorithm versioning.
    - One anomalous examination can be rejected from baseline updates by explicit robust-statistics rules.
  artifacts:
    - ml-services/ml-baseline/app/schemas.py defines the versioned request and response models.
    - ml-services/ml-baseline/app/algorithms.py implements robust center/scale and outlier-gated update rules.
    - docker-compose.yml and .env.example can start the baseline service in the local stack.
  key_links:
    - The baseline service must never read or write PostgreSQL directly.
    - Core backend will depend on a narrow HTTP contract, so request and response schemas must be explicit and versioned.
    - Local compose wiring must expose the service only inside the internal stack, not as a new public surface.
---

<objective>
Create the standalone baseline Python service with a minimal robust-statistics algorithm and reproducible local runtime wiring.

Purpose: isolate the evolution-prone baseline math behind a narrow service boundary while keeping PostgreSQL ownership in core backend.
Output: runnable FastAPI baseline service, algorithm tests, and compose/env wiring for local execution.
</objective>

<execution_context>
@/home/katya/.codex/get-shit-done/workflows/execute-plan.md
@/home/katya/.codex/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/03-aggregated-baseline-aware-profiles/03-RESEARCH.md
@docs/00_project.md
@docs/01_contract.md
@docker-compose.yml
@.env.example

<interfaces>
From research baseline contract:
```json
{
  "schema_version": 1,
  "algorithm_version": "baseline-v1",
  "specialist_id": 10,
  "examination_id": 100,
  "metrics": [{ "key": "overall_index", "value": 0.58 }],
  "general_reference_version": "general-v1",
  "personal_history": [],
  "current_baseline": null
}
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Define the baseline service contract and FastAPI surface</name>
  <files>ml-services/ml-baseline/app/main.py, ml-services/ml-baseline/app/schemas.py, ml-services/ml-baseline/tests/test_service.py, ml-services/ml-baseline/requirements.txt</files>
  <behavior>
    - Test 1: POST baseline calculation returns both general and personal deviation sections for the requested canonical metrics.
    - Test 2: response includes `schema_version`, `algorithm_version`, refresh metadata, and update-eligibility fields.
    - Test 3: health endpoint reports service readiness without needing PostgreSQL access.
  </behavior>
  <action>Create the new `ml-services/ml-baseline` service with FastAPI and Pydantic models that mirror the contract already locked in `docs/01_contract.md`. Add one calculation endpoint and one health endpoint only. Package dependencies in `requirements.txt` with FastAPI, Pydantic, NumPy, SciPy, pytest, and uvicorn. Keep the service compute-only: no database client, no RabbitMQ consumer, and no attempts to own workflow state.</action>
  <verify>
    <automated>cd /home/katya/dimplom/ml-services/ml-baseline && pytest -q tests/test_service.py::test_returns_general_and_personal_deviation</automated>
  </verify>
  <done>The baseline service has a versioned HTTP surface, validated schemas, and passing contract tests for response shape and health behavior.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Implement robust update gating and local compose wiring</name>
  <files>ml-services/ml-baseline/app/algorithms.py, ml-services/ml-baseline/app/main.py, ml-services/ml-baseline/tests/test_algorithms.py, docker-compose.yml, .env.example</files>
  <behavior>
    - Test 1: robust z-score and baseline deltas are calculated from median and MAD, not mean and standard deviation.
    - Test 2: outlier examinations can freeze baseline updates while still returning deviations for the current exam.
    - Test 3: bounded history update keeps baseline exam count controlled and does not grow without rules.
  </behavior>
  <action>Implement the MVP algorithm from research: robust center/scale using median and MAD, bounded history updates, and an outlier-freeze rule that prevents uncontrolled personal-baseline drift. Wire the FastAPI app to call those helpers and return the `next_baseline` snapshot when eligible. Update `docker-compose.yml` and `.env.example` so `ml-services/ml-baseline` runs in the local stack with internal-only networking and explicit env for host, port, timeout, and algorithm version. Do not publish the service as a public browser-facing endpoint.</action>
  <verify>
    <automated>cd /home/katya/dimplom/ml-services/ml-baseline && pytest -q tests/test_algorithms.py::test_outlier_freezes_baseline_update tests/test_service.py::test_returns_general_and_personal_deviation && cd /home/katya/dimplom && docker compose config --services | rg '^ml-baseline$'</automated>
  </verify>
  <done>The baseline service computes robust deviations, guards baseline updates against outliers, and is reproducibly wired into the local compose stack.</done>
</task>

</tasks>

<verification>
Run the focused pytest suite and confirm `docker compose config` includes the new internal baseline service without breaking the existing stack definition.
</verification>

<success_criteria>
- `ml-services/ml-baseline` exists as a standalone FastAPI service with versioned schemas and robust-statistics logic.
- Baseline update eligibility is explicit and resistant to single-exam anomalies.
- The local stack can resolve the service through compose wiring and documented env values.
</success_criteria>

<output>
After completion, create `.planning/phases/03-aggregated-baseline-aware-profiles/03-03-SUMMARY.md`
</output>
