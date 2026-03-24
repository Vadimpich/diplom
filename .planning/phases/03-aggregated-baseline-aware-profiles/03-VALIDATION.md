---
phase: 3
slug: aggregated-baseline-aware-profiles
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-22
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` + `net/http/httptest`, frontend lint/build/typecheck, Python `pytest` for baseline service |
| **Config file** | none |
| **Quick run command** | `cd /home/vadim/diplom/core-backend && go test ./internal/aggregation ./internal/http -run 'TestAggregationReadyOnlyAfterAllChannelsSucceeded|TestAggregationStoresVersionedProfile|TestAggregationIncludesContributionsAndExplanations|TestProcessingStatusShowsAggregating|TestSpecialistResultHistoryEndpoint' -count=1 && cd /home/vadim/diplom/frontend && npm run lint && npm run build && npx tsc --noEmit` |
| **Full suite command** | `cd /home/vadim/diplom/core-backend && go test ./... -count=1 && cd /home/vadim/diplom/frontend && npm run lint && npm run build && npx tsc --noEmit && source /tmp/diplom-ml-baseline-venv/bin/activate && cd /home/vadim/diplom/ml-services/ml-baseline && pytest -q` |
| **Estimated runtime** | ~210 seconds |

---

## Sampling Rate

- **After every task commit:** Run `cd /home/vadim/diplom/core-backend && go test ./internal/aggregation ./internal/http -count=1`
- **After every plan wave:** Run `cd /home/vadim/diplom/core-backend && go test ./... -count=1 && cd /home/vadim/diplom/frontend && npm run lint && npm run build && npx tsc --noEmit && source /tmp/diplom-ml-baseline-venv/bin/activate && cd /home/vadim/diplom/ml-services/ml-baseline && pytest -q`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 210 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 3-01-01 | 01 | 1 | AGGR-02, AGGR-03, BASE-01 | contracts + RED tests | `cd /home/vadim/diplom/core-backend && go test ./internal/aggregation ./internal/http -run 'TestAggregationStoresVersionedProfile|TestProcessingStatusShowsAggregating|TestSpecialistResultHistoryEndpoint' -count=1 && source /tmp/diplom-ml-baseline-venv/bin/activate && cd /home/vadim/diplom/ml-services/ml-baseline && pytest -q tests/test_service.py::test_returns_general_and_personal_deviation tests/test_algorithms.py::test_outlier_freezes_baseline_update` | ✅ | ✅ green |
| 3-01-02 | 01 | 1 | AGGR-02, BASE-02, RSLT-03 | schema + contracts | `cd /home/vadim/diplom/core-backend && go test ./internal/aggregation ./internal/http -run 'TestAggregationStoresVersionedProfile|TestSpecialistResultHistoryEndpoint' -count=1` | ✅ | ✅ green |
| 3-02-01 | 02 | 2 | AGGR-01, AGGR-02, AGGR-03 | normalization + explanations | `cd /home/vadim/diplom/core-backend && go test ./internal/aggregation -run 'TestAggregationReadyOnlyAfterAllChannelsSucceeded|TestAggregationStoresVersionedProfile|TestAggregationIncludesContributionsAndExplanations' -count=1` | ✅ | ✅ green |
| 3-02-02 | 02 | 2 | AGGR-01 | aggregating state transition | `cd /home/vadim/diplom/core-backend && go test ./internal/aggregation ./internal/processing -run 'TestAggregationReadyOnlyAfterAllChannelsSucceeded|TestProcessingStatusShowsAggregating' -count=1` | ✅ | ✅ green |
| 3-03-01 | 03 | 2 | BASE-01, BASE-02 | baseline HTTP contract | `source /tmp/diplom-ml-baseline-venv/bin/activate && cd /home/vadim/diplom/ml-services/ml-baseline && pytest -q tests/test_service.py::test_returns_general_and_personal_deviation` | ✅ | ✅ green |
| 3-03-02 | 03 | 2 | BASE-03 | robust update gating + compose wiring | `source /tmp/diplom-ml-baseline-venv/bin/activate && cd /home/vadim/diplom/ml-services/ml-baseline && pytest -q tests/test_algorithms.py::test_outlier_freezes_baseline_update tests/test_service.py::test_returns_general_and_personal_deviation` | ✅ | ✅ green |
| 3-04-01 | 04 | 3 | BASE-01 | baseline HTTP client | `cd /home/vadim/diplom/core-backend && go test ./internal/baselineclient -count=1` | ✅ | ✅ green |
| 3-04-02 | 04 | 3 | BASE-01, BASE-02, BASE-03, AGGR-01 | baseline persistence + aggregated terminal status | `cd /home/vadim/diplom/core-backend && go test ./internal/aggregation ./internal/processing -count=1` | ✅ | ✅ green |
| 3-05-01 | 05 | 4 | AGGR-03 | result endpoint | `cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/results -run 'TestSpecialistResultHistoryEndpoint' -count=1` | ✅ | ✅ green |
| 3-05-02 | 05 | 4 | RSLT-03 | specialist result-history endpoint | `cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/results -run 'TestSpecialistResultHistoryEndpoint' -count=1` | ✅ | ✅ green |
| 3-06-01 | 06 | 5 | AGGR-03, RSLT-03 | frontend typed contracts | `cd /home/vadim/diplom/frontend && npm run lint && npm run build && npx tsc --noEmit` | ✅ | ✅ green |
| 3-06-02 | 06 | 5 | RSLT-03 | frontend result/history/processing navigation | `cd /home/vadim/diplom/frontend && npm run lint && npm run build && npx tsc --noEmit` | ✅ | ✅ green |
| 3-06-03 | 06 | 5 | AGGR-03, RSLT-03 | docs + full regression | `cd /home/vadim/diplom/core-backend && go test ./... -count=1 && cd /home/vadim/diplom/frontend && npm run lint && npm run build && npx tsc --noEmit && source /tmp/diplom-ml-baseline-venv/bin/activate && cd /home/vadim/diplom/ml-services/ml-baseline && pytest -q && cd /home/vadim/diplom && rg -n 'aggregated|baseline|result-history|/examinations/\\{id\\}/result' README.md .planning/phases/03-aggregated-baseline-aware-profiles/03-VALIDATION.md docs/02_implementation.md` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `/home/vadim/diplom/core-backend/internal/aggregation/service_test.go` — covers `AGGR-01`, `AGGR-02`, `AGGR-03`
- [x] `/home/vadim/diplom/core-backend/internal/http/results_handler_test.go` or equivalent trend/history HTTP tests — covers `RSLT-03`
- [x] `/home/vadim/diplom/ml-services/ml-baseline/tests/test_service.py` — covers `BASE-01`
- [x] `/home/vadim/diplom/ml-services/ml-baseline/tests/test_algorithms.py` — covers `BASE-03`
- [x] `/home/vadim/diplom/core-backend/internal/aggregation/` integration seam for baseline client responses — required so baseline persistence is not validated only by mocks
- [x] `/home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/processing/page.tsx` compatibility coverage — required so `aggregating` continues to route/render correctly during the Phase 3 status expansion

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Operator can inspect specialist dynamics against baseline across multiple examinations | RSLT-03 | Requires realistic browser navigation through history/result surfaces and multi-examination seeded data | Create at least two successful examinations for one specialist, open the specialist history/results UI, and confirm the displayed trend values and baseline deltas match backend DTOs rather than client-side guesses |
| Aggregated explanations remain interpretable despite stub worker payloads | AGGR-03 | Human interpretability cannot be fully asserted by automated tests alone | Inspect one aggregated result payload and operator result screen, verify explanation bullets name neutral proxy metrics and channel contributions without overclaiming clinical meaning |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 210s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** passed
