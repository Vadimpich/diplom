# Phase 3: Aggregated Baseline-Aware Profiles - Research

**Researched:** 2026-03-22
**Domain:** Core-owned examination aggregation with baseline-aware profiling and specialist trend history
**Confidence:** MEDIUM

## User Constraints

No `*-CONTEXT.md` was provided for this phase. Honor the direct scope from the request:

- Successful channel outputs must become an interpretable, versioned examination profile enriched with general and personal baseline deviation.
- Phase 2 facts must remain true:
  - `processing-status` currently reaches `terminal=true` when 3/3 channels succeed.
  - coarse examination status still stays `processing`.
  - PostgreSQL is authoritative for processing state.
  - channel results are unified and persisted by core backend.
  - workers are still stub services; real channel modeling is intentionally deferred.
  - `docs/00_project.md` recommends a separate Python baseline service and interpretable aggregation.
- Research must answer:
  - where aggregation logic should live now;
  - what normalized aggregated result schema should look like for later KESMI stability;
  - how baseline should be split between PostgreSQL persistence and a Python service contract;
  - how to move terminal channel success into `aggregating` without breaking current APIs/UI;
  - what minimum viable aggregation and baseline algorithms fit stub-channel constraints;
  - what tests should prove `AGGR-01..03`, `BASE-01..03`, `RSLT-03` without pretending ML models are final.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| AGGR-01 | Aggregator waits for all mandatory channels before forming final profile | Use a core-owned aggregation runner that selects only examinations whose mandatory channel runs are all `succeeded` and whose profile row does not yet exist |
| AGGR-02 | Aggregator normalizes channel outputs and stores a versioned aggregated profile | Introduce one canonical profile schema with `schema_version`, `aggregation_version`, normalized metrics, and stable JSON/row storage |
| AGGR-03 | System stores each channel contribution and human-readable explanation | Persist per-metric channel contribution rows plus explanation bullets generated from deterministic rules |
| BASE-01 | Baseline service computes deviation from general and personal norm | Keep baseline math in a separate Python service fed by core-owned aggregated metric vectors and historical profile snapshots |
| BASE-02 | Baseline profile stores refresh date, examination count, and algorithm version | Persist baseline profile metadata in PostgreSQL and snapshot returned baseline metadata onto each examination profile |
| BASE-03 | Baseline resists one-off anomalies and avoids uncontrolled updates | Use robust center/scale (`median`, `MAD`) plus outlier-gated baseline updates and bounded rolling history |
| RSLT-03 | Operator can inspect dynamics across examination history | Add specialist history/trend DTOs backed by aggregated profiles and baseline snapshots rather than raw channel payloads |
</phase_requirements>

## Summary

Phase 3 should not move aggregation into another ML worker. The current repository already stores every successful channel result in PostgreSQL, and Phase 2 deliberately keeps workflow ownership in the Go core backend. That same boundary should continue: core backend owns the examination state machine, decides when aggregation may start, builds the canonical aggregated profile, persists versioned result artifacts, and exposes stable HTTP DTOs for history and later KESMI delivery. This keeps PostgreSQL as the single source of truth and avoids turning aggregation into another transport-coupled queue workflow before the project has real ML payloads.

The separate service boundary belongs to baseline, not aggregation. `docs/00_project.md` explicitly calls out a standalone baseline module, and that split is useful now because baseline math is the part most likely to evolve from simple robust statistics into richer Python-centric logic. Core should send a compact, versioned baseline request containing the aggregated metric vector, general reference set version, and specialist history summary. The baseline service should never write PostgreSQL directly. It returns deviations, update eligibility, and the next candidate baseline profile; core persists all of that transactionally.

The main compatibility risk is status vocabulary. Phase 2 currently treats “all channels succeeded” as `terminal=true` in `/processing-status` while the coarse examination object still says `processing`. Phase 3 should fix that, but as an additive change: introduce `aggregating` and `aggregated` in the contract and UI in the same phase, keep `/processing-status` as the authoritative progress endpoint, and make history/status badges treat `ready_for_processing`, `processing`, and `aggregating` as “in progress” while `aggregated` becomes the pre-KESMI terminal success state.

**Primary recommendation:** Keep aggregation deterministic and core-owned; add a separate Python baseline service with a narrow HTTP contract; persist one canonical `aggregated examination profile` schema that later KESMI integration can reuse unchanged.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go core backend modules | repo-local | Aggregation eligibility, profile persistence, status transitions, API DTOs | Phase 2 already made core the owner of channel-result workflow and PostgreSQL truth |
| PostgreSQL JSONB + relational support tables | existing repo DB | Authoritative storage for one profile per examination, baseline snapshots, and trend projections | Stable persistence boundary for later Phase 4 integration and operator history |
| FastAPI | `0.135.1` (published 2026-03-01) | Separate Python baseline service HTTP boundary | Same Python service shell pattern already matches project architecture and existing worker stack |
| Pydantic | `2.12.5` (published 2025-11-26) | Versioned baseline request/response schemas | Strong contract validation for service boundary changes |
| NumPy | `2.4.3` (published 2026-03-09) | Vector math for baseline statistics | Standard numeric base for deterministic profile deviation math |
| SciPy | `1.17.1` (published 2026-02-23) | Robust statistics helpers such as MAD and winsorization | Avoids hand-rolling fragile baseline math |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/rabbitmq/amqp091-go` | `v1.10.0` (published 2024-05-08) | Keep Phase 2 queue infrastructure unchanged | Use only for existing channel result flow; do not add a second broker workflow for baseline in Phase 3 |
| `aio-pika` | `9.6.1` (published 2026-02-23) | Existing worker stack dependency | Keep for current channel workers; baseline service does not need it in MVP |
| TanStack Query | `5.94.5` (registry modified 2026-03-21) | Polling and history queries in frontend | Use for aggregate-status and trend-history views |
| Next.js | `16.2.1` (registry modified 2026-03-20) | Existing frontend platform | Keep existing operator history/result route structure and extend typed status unions |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Core-owned aggregation | Separate Python aggregator service | More architectural purity, but premature network boundary around deterministic orchestration and schema assembly |
| HTTP baseline service | RabbitMQ-based baseline worker | Better decoupling, but unnecessary queue complexity for small vector computations in this phase |
| Robust median/MAD baseline | Mean/stddev baseline | Mean/stddev is simpler but too sensitive to single anomalous stub examinations |
| Stable neutral metric names | Domain-heavy “stress/depression/anxiety” labels | Domain-heavy labels overclaim meaning while workers still emit proxies only |

**Installation:**
```bash
cd /home/katya/dimplom/ml-services/ml-baseline
pip install fastapi==0.135.1 pydantic==2.12.5 numpy==2.4.3 scipy==1.17.1 uvicorn==0.34.0
```

**Version verification:**
```bash
cd /home/katya/dimplom/core-backend && go list -m -json github.com/rabbitmq/amqp091-go@latest
cd /home/katya/dimplom/frontend && npm view @tanstack/react-query version time.modified && npm view next version time.modified
curl -s https://pypi.org/pypi/fastapi/json | jq -r '.info.version, .releases[.info.version][-1].upload_time_iso_8601'
curl -s https://pypi.org/pypi/pydantic/json | jq -r '.info.version, .releases[.info.version][-1].upload_time_iso_8601'
curl -s https://pypi.org/pypi/numpy/json | jq -r '.info.version, .releases[.info.version][-1].upload_time_iso_8601'
curl -s https://pypi.org/pypi/scipy/json | jq -r '.info.version, .releases[.info.version][-1].upload_time_iso_8601'
```

## Architecture Patterns

### Recommended Project Structure
```text
core-backend/
├── internal/aggregation/              # eligibility, normalization, explanation rules, profile persistence
├── internal/baselineclient/           # HTTP contract client for Python baseline service
├── internal/results/                  # aggregate + trend DTOs for operator-facing endpoints
└── migrations/                        # examination_profiles, metric rows, baseline tables

ml-services/ml-baseline/
├── app/main.py                        # FastAPI service
├── app/schemas.py                     # request/response envelopes
└── app/algorithms.py                  # robust center/scale and update gating

frontend/
├── app/(app)/operator/examinations/[id]/processing/
├── app/(app)/operator/examinations/[id]/results/
└── app/(app)/operator/specialists/[id]/   # history + trend view
```

### Pattern 1: Core-Owned Aggregation Runner
**What:** Core backend selects examinations whose mandatory channels all succeeded, marks them `aggregating`, computes the canonical profile, calls baseline service, persists outputs, and marks the examination `aggregated`.
**When to use:** Always. Aggregation is workflow orchestration plus stable contract assembly, not ML inference.
**Example:**
```go
// Source: project pattern derived from /home/katya/dimplom/core-backend/internal/channelresults/service.go
func (s *Service) OnChannelResultApplied(ctx context.Context, examinationID int64) error {
	ready, err := s.repo.IsAggregationReady(ctx, examinationID)
	if err != nil || !ready {
		return err
	}

	return s.repo.WithTx(ctx, func(tx Tx) error {
		if err := tx.MarkAggregating(ctx, examinationID); err != nil {
			return err
		}

		channelPayloads, err := tx.LoadSucceededChannelPayloads(ctx, examinationID)
		if err != nil {
			return err
		}

		profile := NormalizeProfile(channelPayloads)
		baseline := s.baselineClient.Calculate(ctx, BuildBaselineRequest(profile, tx.LoadHistory(ctx, examinationID)))

		if err := tx.SaveAggregatedProfile(ctx, profile, baseline); err != nil {
			return err
		}
		return tx.MarkAggregated(ctx, examinationID)
	})
}
```

### Pattern 2: Canonical Aggregated Profile Schema
**What:** Persist one stable profile shape per examination, independent from raw worker payload quirks.
**When to use:** For storage, operator history, and later KESMI delivery.
**Example:**
```json
{
  "schema_version": 1,
  "aggregation_version": "agg-v1",
  "baseline_version": "baseline-v1",
  "examination_id": 100,
  "specialist_id": 10,
  "generated_at": "2026-03-22T10:00:00Z",
  "status": "aggregated",
  "summary": {
    "overall_index": 0.58,
    "overall_band": "elevated",
    "explanation": "Повышение в основном связано с акустическими и паралингвистическими прокси."
  },
  "metrics": [
    {
      "key": "text_signal_score",
      "label": "Текстовый прокси-сигнал",
      "value": 0.44,
      "scale": "0..1",
      "interpretation": "higher_means_more_deviation"
    }
  ],
  "channel_contributions": [
    {
      "channel": "text",
      "weight": 0.33,
      "contribution": 0.15,
      "evidence": ["text_total_characters", "text_non_empty_answers"]
    }
  ],
  "baseline": {
    "general": { "delta": 0.21, "band": "mild", "reference_version": "general-v1" },
    "personal": { "delta": 0.37, "band": "moderate", "baseline_exam_count": 4 }
  }
}
```

### Pattern 3: Additive Status Expansion
**What:** Extend examination statuses to include `aggregating` and `aggregated`, but keep `/processing-status` as the backend-authoritative progress endpoint.
**When to use:** During Phase 3 migration from “channel success terminal” to “profile persisted terminal”.
**Example transition:**
```text
created
-> collecting_answers
-> ready_for_processing
-> processing
-> aggregating
-> aggregated
-> decision_pending   # Phase 4
-> completed          # Phase 4

any stage -> failed
```

### Pattern 4: Baseline Service is Pure Compute, Core Persists Everything
**What:** Core sends a metric vector plus bounded history summary; baseline service returns deviation outputs and an updated baseline candidate only.
**When to use:** Always. Baseline service must not own PostgreSQL state.
**Example request/response:**
```json
// Request
{
  "schema_version": 1,
  "algorithm_version": "baseline-v1",
  "specialist_id": 10,
  "examination_id": 100,
  "metrics": [
    { "key": "overall_index", "value": 0.58 },
    { "key": "text_signal_score", "value": 0.44 }
  ],
  "general_reference_version": "general-v1",
  "personal_history": [
    { "examination_id": 91, "captured_at": "2026-03-01T09:00:00Z", "metrics": { "overall_index": 0.31 } }
  ],
  "current_baseline": {
    "exam_count": 4,
    "centers": { "overall_index": 0.29 },
    "scales": { "overall_index": 0.07 }
  }
}

// Response
{
  "schema_version": 1,
  "algorithm_version": "baseline-v1",
  "general_deviation": {
    "overall_index": { "delta": 0.21, "robust_z": 1.8, "band": "mild" }
  },
  "personal_deviation": {
    "overall_index": { "delta": 0.29, "robust_z": 2.7, "band": "moderate" }
  },
  "baseline_update": {
    "eligible": false,
    "reason": "outlier_freeze",
    "exam_count_before": 4,
    "exam_count_after": 4
  },
  "next_baseline": {
    "last_refreshed_at": "2026-03-22T10:00:00Z",
    "exam_count": 4,
    "centers": { "overall_index": 0.29 },
    "scales": { "overall_index": 0.07 }
  }
}
```

### Anti-Patterns to Avoid
- **Separate aggregator worker queue in Phase 3:** adds a second orchestration transport without solving an actual current bottleneck.
- **Letting baseline service read/write PostgreSQL directly:** breaks the project’s service boundaries and makes idempotency harder.
- **Shipping raw worker payloads as the “final profile”:** locks KESMI to stub-specific fields like `audio_total_bytes`.
- **Using clinical-sounding metric names for stub-derived proxies:** overstates evidential quality and will force later contract churn.
- **Keeping `terminal=true` immediately after 3/3 channel success:** contradicts Phase 3 success criteria because the profile still does not exist.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Robust personal baseline | Custom ad hoc threshold spaghetti | Median + MAD + winsorized update gate | Standard robust statistics already solve one-off outlier sensitivity |
| Stable result contract | Channel-specific JSON blobs stitched in UI | Canonical aggregated profile schema | Prevents KESMI and frontend from binding to stub worker internals |
| Cross-service persistence | Python service writing core DB tables | Core-owned persistence after validated service response | Keeps PostgreSQL authoritative and preserves transaction boundaries |
| Trend projection | Frontend-side history math over raw exams | Backend DTO built from aggregated profile snapshots | One source of truth and repeatable operator history |

**Key insight:** Phase 3 is a schema-and-state-machine phase first, not a model-quality phase. Build the durable contract and robust update rules now so real ML models can replace stub scoring later without forcing workflow or API rewrites.

## Common Pitfalls

### Pitfall 1: Treating Stub Payloads as Domain Truth
**What goes wrong:** Fields such as `audio_total_bytes` or `speech_rate_proxy` leak directly into operator/KESMI contracts.
**Why it happens:** Those are the only values available today, so they are tempting to expose unchanged.
**How to avoid:** Convert all current worker outputs into neutral canonical metrics like `text_signal_score`, `acoustic_signal_score`, `paralinguistic_signal_score`, `overall_index`.
**Warning signs:** Final DTO contains worker field names verbatim or explanation text claims clinical meaning not justified by stubs.

### Pitfall 2: Breaking the UI with Premature Status Expansion
**What goes wrong:** History lists or badges fail once backend starts returning `aggregating` or `aggregated`.
**Why it happens:** Phase 2 frontend unions and routes are exhaustive over `created|collecting_answers|ready_for_processing|processing|failed`.
**How to avoid:** Update `docs/01_contract.md`, backend status checks, typed frontend unions, shared badges, and history navigation in the same phase slice.
**Warning signs:** TypeScript exhaustiveness errors or operator history linking `aggregating` examinations to the wrong screen.

### Pitfall 3: Uncontrolled Baseline Drift
**What goes wrong:** One anomalous examination drags the personal baseline and hides future deviations.
**Why it happens:** Naive mean/stddev updates and “always include current exam” logic.
**How to avoid:** Freeze updates when robust z-score exceeds threshold; use bounded recent history and robust center/scale instead of plain mean.
**Warning signs:** Baseline exam count increments on every exam, including clear outliers.

### Pitfall 4: Non-Idempotent Aggregation
**What goes wrong:** Replayed result messages or reruns create duplicate profiles or conflicting baseline updates.
**Why it happens:** Aggregation is triggered from every success result without a unique fence.
**How to avoid:** Enforce `UNIQUE (examination_id)` on the aggregated profile table and mark `aggregating` under row lock before work starts.
**Warning signs:** Multiple profiles for one examination or repeated baseline refreshes for the same exam.

## Code Examples

Verified patterns from project state and official docs:

### Robust Scale Calculation
```python
# Source: https://docs.scipy.org/doc/scipy/reference/generated/scipy.stats.median_abs_deviation.html
from scipy.stats import median_abs_deviation
import numpy as np

def robust_z_score(current: float, values: list[float]) -> float:
    center = float(np.median(values))
    scale = float(median_abs_deviation(values, scale="normal"))
    scale = max(scale, 1e-6)
    return (current - center) / scale
```

### Winsorized Update Gate
```python
# Source: https://docs.scipy.org/doc/scipy/reference/generated/scipy.stats.mstats.winsorize.html
from scipy.stats.mstats import winsorize
import numpy as np

def bounded_center(values: list[float]) -> float:
    clipped = winsorize(values, limits=(0.1, 0.1))
    return float(np.median(clipped))
```

### Stable Percentile Banding
```python
# Source: https://numpy.org/doc/stable/reference/generated/numpy.percentile.html
import numpy as np

def band_edges(values: list[float]) -> tuple[float, float]:
    return tuple(np.percentile(values, [25, 75]))
```

### Core-Side Explanation Rule
```go
// Source: repository pattern derived from current channel-results + processing services
func explain(profile AggregatedProfile) []string {
	lines := []string{}
	if profile.ChannelContributions["acoustic"] > 0.4 {
		lines = append(lines, "Основной вклад внёс акустический канал.")
	}
	if profile.Baseline.Personal.RobustZ >= 2.5 {
		lines = append(lines, "Профиль заметно отклоняется от личной нормы специалиста.")
	}
	return lines
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Channel success meant workflow success | Channel success only unlocks aggregation eligibility | Phase 3 | `terminal=true` must mean “profile persisted”, not “channels finished” |
| Raw channel payloads are the only persisted analytics | Canonical aggregated profile becomes the durable business result | Phase 3 | Later KESMI integration can bind to one stable schema |
| Personal baseline as undefined future idea | Separate Python service with robust statistics and persisted metadata | Phase 3 | Baseline becomes implementable now without locking in final ML semantics |

**Deprecated/outdated:**
- `processing-status terminal=true after 3/3 succeeded`: outdated for Phase 3 because it leaves no room for the required aggregation/baseline step.
- Using only coarse `processing` as the final success state: outdated once aggregated profile history becomes operator-visible.

## Open Questions

1. **Should `aggregated` or `profile_ready` be the post-Phase-3 success status?**
   - What we know: `docs/01_contract.md` already reserves future statuses like `aggregating`, `decision_pending`, `completed`.
   - What's unclear: the exact intermediate name desired before KESMI in Phase 4.
   - Recommendation: use `aggregated` now because it states exactly what Phase 3 delivers and maps cleanly to `decision_pending` in Phase 4.

2. **Which canonical metrics should be mandatory in v1?**
   - What we know: current workers expose only proxy/stub fields, not validated psychological constructs.
   - What's unclear: whether the diploma reviewers expect domain-rich labels or infrastructure-neutral proxies.
   - Recommendation: lock a small neutral set now: `text_signal_score`, `acoustic_signal_score`, `paralinguistic_signal_score`, `overall_index`.

3. **Should baseline service receive full history rows or only the current baseline snapshot plus recent values?**
   - What we know: direct DB access from Python service should be avoided.
   - What's unclear: whether later algorithms will need richer temporal context.
   - Recommendation: send both `current_baseline` and a bounded last-`N` history summary; it is still a narrow contract and avoids repainting the interface later.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + `net/http/httptest`; frontend `eslint` + `next build` + `tsc`; Python baseline service `pytest` |
| Config file | `none` for Go/frontend; `none — see Wave 0` for Python baseline |
| Quick run command | `cd /home/katya/dimplom/core-backend && go test ./internal/aggregation ./internal/results ./internal/http -run 'TestAggregationReadyOnlyAfterAllChannelsSucceeded|TestAggregationStoresVersionedProfile|TestProcessingStatusShowsAggregating|TestSpecialistTrendHistory' -count=1 && cd /home/katya/dimplom/frontend && npm run lint && npm run build && npx tsc --noEmit` |
| Full suite command | `cd /home/katya/dimplom/core-backend && go test ./... -count=1 && cd /home/katya/dimplom/frontend && npm run lint && npm run build && npx tsc --noEmit && cd /home/katya/dimplom/ml-services/ml-baseline && pytest -q` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| AGGR-01 | Aggregation does not start until all mandatory channels are `succeeded` | unit/service | `cd /home/katya/dimplom/core-backend && go test ./internal/aggregation -run TestAggregationReadyOnlyAfterAllChannelsSucceeded -count=1` | ❌ Wave 0 |
| AGGR-02 | One versioned canonical profile is persisted per examination | unit/repository | `cd /home/katya/dimplom/core-backend && go test ./internal/aggregation -run TestAggregationStoresVersionedProfile -count=1` | ❌ Wave 0 |
| AGGR-03 | Stored profile includes per-channel contributions and deterministic explanations | unit/service | `cd /home/katya/dimplom/core-backend && go test ./internal/aggregation -run TestAggregationIncludesContributionsAndExplanations -count=1` | ❌ Wave 0 |
| BASE-01 | Baseline response includes general and personal deviation for each canonical metric | unit/Python | `cd /home/katya/dimplom/ml-services/ml-baseline && pytest -q tests/test_service.py::test_returns_general_and_personal_deviation` | ❌ Wave 0 |
| BASE-02 | Baseline metadata persists refresh time, exam count, and algorithm version | integration | `cd /home/katya/dimplom/core-backend && go test ./internal/aggregation -run TestBaselineMetadataPersistedOnProfile -count=1` | ❌ Wave 0 |
| BASE-03 | Outlier exam does not automatically mutate the personal baseline | unit/Python | `cd /home/katya/dimplom/ml-services/ml-baseline && pytest -q tests/test_algorithms.py::test_outlier_freezes_baseline_update` | ❌ Wave 0 |
| RSLT-03 | Specialist history returns dynamics of key indicators across examinations | HTTP + frontend | `cd /home/katya/dimplom/core-backend && go test ./internal/http -run TestSpecialistTrendHistoryEndpoint -count=1 && cd /home/katya/dimplom/frontend && npm run lint && npm run build && npx tsc --noEmit` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `cd /home/katya/dimplom/core-backend && go test ./internal/aggregation ./internal/http -count=1`
- **Per wave merge:** `cd /home/katya/dimplom/core-backend && go test ./... -count=1 && cd /home/katya/dimplom/frontend && npm run lint && npm run build && npx tsc --noEmit`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `/home/katya/dimplom/core-backend/internal/aggregation/service_test.go` — covers `AGGR-01`, `AGGR-02`, `AGGR-03`
- [ ] `/home/katya/dimplom/core-backend/internal/http/results_handler_test.go` — covers `RSLT-03`
- [ ] `/home/katya/dimplom/ml-services/ml-baseline/tests/test_algorithms.py` — covers `BASE-01`, `BASE-03`
- [ ] `/home/katya/dimplom/ml-services/ml-baseline/tests/test_service.py` — covers baseline contract shape
- [ ] Framework install: `cd /home/katya/dimplom/ml-services/ml-baseline && pip install pytest`

## Sources

### Primary (HIGH confidence)
- `/home/katya/dimplom/docs/00_project.md` - target architecture, explicit Aggregator and Baseline Service split
- `/home/katya/dimplom/docs/01_contract.md` - current Phase 2 HTTP/AMQP contract, status vocabulary, PostgreSQL persistence model
- `/home/katya/dimplom/docs/02_implementation.md` - shipped Phase 2 implementation facts
- `/home/katya/dimplom/.planning/REQUIREMENTS.md` - exact requirement text for `AGGR-*`, `BASE-*`, `RSLT-03`
- `/home/katya/dimplom/.planning/ROADMAP.md` - Phase 3 goal and success criteria
- `/home/katya/dimplom/core-backend/internal/channelresults/service.go` - current result-ingestion ownership in core
- `/home/katya/dimplom/core-backend/internal/processing/repository.go` - current finish/status workflow and PostgreSQL truth pattern
- `https://docs.scipy.org/doc/scipy/reference/generated/scipy.stats.median_abs_deviation.html` - robust scale helper
- `https://docs.scipy.org/doc/scipy/reference/generated/scipy.stats.mstats.winsorize.html` - winsorization helper
- `https://numpy.org/doc/stable/reference/generated/numpy.percentile.html` - percentile banding helper

### Secondary (MEDIUM confidence)
- PyPI package metadata for FastAPI, Pydantic, NumPy, SciPy, aio-pika - current version verification as of 2026-03-22
- npm registry metadata for Next.js and TanStack Query - current frontend version verification as of 2026-03-22
- Go module metadata for `github.com/rabbitmq/amqp091-go` - current Go client version verification as of 2026-03-22

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - driven by current repo state plus verified package versions
- Architecture: HIGH - directly constrained by `docs/00_project.md`, `docs/01_contract.md`, and shipped Phase 2 code
- Pitfalls: MEDIUM - partly inferred from current code and stub-worker limitations, though strongly supported by architecture constraints

**Research date:** 2026-03-22
**Valid until:** 2026-04-21
