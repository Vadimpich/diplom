---
phase: 04
slug: decision-delivery-to-operator
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-23
---

# Phase 04 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` + frontend `eslint/build/tsc` + Python `pytest` |
| **Config file** | `core-backend`: none; `frontend/package.json` scripts; `ml-services/ml-baseline`: none |
| **Quick run command** | `cd /home/vadim/diplom/core-backend && go test ./internal/decision ./internal/kesmi ./internal/results ./internal/http -run 'TestCreateDecisionInput|TestDecisionPendingTransition|TestRetriesOnlyTransportFailures|TestExaminationResultIncludesDecisionBlock|TestDecisionFailureDiagnostics|TestCompletedResultIncludesPlaceholderDecision|TestDecisionSuccessMarksCompleted|TestDecisionRelayResumesPendingSnapshots' -count=1` |
| **Full suite command** | `cd /home/vadim/diplom/core-backend && go test ./... -count=1 && cd /home/vadim/diplom/frontend && npm run lint && npm run build && npx tsc --noEmit` |
| **Estimated runtime** | ~90 seconds |

---

## Sampling Rate

- **After every task commit:** Run the plan-specific narrow command from the verification map below
- **After every plan wave:** Run `cd /home/vadim/diplom/core-backend && go test ./... -count=1 && cd /home/vadim/diplom/frontend && npm run lint && npm run build && npx tsc --noEmit`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 90 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 04-00-01 | 00 | 0 | KSMI-01, KSMI-02, KSMI-03, RSLT-02 | red-scaffold | `cd /home/vadim/diplom && rg -n 'TestCreateDecisionInput|TestDecisionPendingTransition|TestDecisionSuccessMarksCompleted' core-backend/internal/decision/service_test.go && rg -n 'TestRetriesOnlyTransportFailures' core-backend/internal/kesmi/client_test.go && rg -n 'TestExaminationResultIncludesDecisionBlock|TestDecisionFailureDiagnostics|TestCompletedResultIncludesPlaceholderDecision' core-backend/internal/http/results_handler_test.go && rg -n 'docker compose up -d --build wimi core-backend|docker compose exec core-backend|curl -fsS http://localhost:8080/health' wimi-server/scripts/smoke.sh && rg -n '/operator/examinations/\\{id\\}/results|Не реализовано|channel contributions' .planning/phases/04-decision-delivery-to-operator/04-UI-SMOKE.md` | ❌ W0 | ⬜ pending |
| 04-01-01 | 01 | 1 | KSMI-01 | unit/integration | `cd /home/vadim/diplom/core-backend && go test ./internal/decision ./internal/http -run 'TestCreateDecisionInput|TestDecisionPendingTransition' -count=1` | ❌ W0 | ⬜ pending |
| 04-01-02 | 01 | 1 | KSMI-02 | unit/integration | `cd /home/vadim/diplom/core-backend && go test ./internal/decision ./internal/kesmi -run 'TestRetriesOnlyTransportFailures|TestExaminationResultIncludesDecisionBlock' -count=1` | ❌ W0 | ⬜ pending |
| 04-02-01 | 02 | 2 | KSMI-01, KSMI-02 | unit/integration | `cd /home/vadim/diplom/core-backend && go test ./internal/decision ./internal/kesmi ./internal/aggregation -run 'TestDecisionSuccessMarksCompleted|TestDecisionRelayResumesPendingSnapshots' -count=1` | ❌ W0 | ⬜ pending |
| 04-03-01 | 03 | 4 | KSMI-03, RSLT-02 | API/unit | `cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/results -run 'TestExaminationResultIncludesDecisionBlock|TestDecisionFailureDiagnostics|TestCompletedResultIncludesPlaceholderDecision' -count=1` | ❌ W0 | ⬜ pending |
| 04-03-02 | 03 | 4 | RSLT-02 | build/assertion | `cd /home/vadim/diplom/frontend && npm run lint && npm run build && npx tsc --noEmit && rg -n 'Не реализовано|raw_response_available|correlation_id|attempt_count' app/'(app)'/operator/examinations/'[id]'/results/page.tsx lib/api/types.ts` | ✅ build | ⬜ pending |
| 04-04-01 | 04 | 3 | KSMI-01, KSMI-02 | smoke/integration | `docker compose up -d --build wimi core-backend && docker compose exec core-backend sh -lc 'wget -qO- http://wimi:8081/Models >/dev/null' && curl -fsS http://localhost:8080/health` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `core-backend/internal/decision/service_test.go` — stubs for `KSMI-01`
- [ ] `core-backend/internal/kesmi/client_test.go` — stubs for `KSMI-02`
- [ ] `core-backend/internal/http/results_handler_test.go` additions — response coverage for `KSMI-03`
- [ ] `.planning/phases/04-decision-delivery-to-operator/04-UI-SMOKE.md` — documented manual probe for `RSLT-02` via `04-00-PLAN.md`
- [ ] `wimi-server/scripts/smoke.sh` — startup and response verification for mandatory Compose runtime via `04-00-PLAN.md`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Operator sees fixed non-implemented message instead of fake recommendation before the real model exists | KSMI-03, RSLT-02 | Current frontend stack has build validation and grep assertions, but no browser automation in phase baseline | 1. Start stack with `docker compose up -d --build wimi core-backend frontend`. 2. Open `/operator/examinations/{id}/results`. 3. Confirm the decision card shows `Не реализовано`, displays `correlation_id`, `attempt_count`, and `raw_response_available`. 4. Confirm metrics, baseline, explanations, and channel contributions stay visible below the decision card. |
| WiMi mandatory runtime starts and answers inside Compose | KSMI-01, KSMI-02 | Requires Docker runtime and vendor binary startup not available in this environment | 1. Start `wimi-server` with project compose. 2. Verify container/service is healthy. 3. Call `GET /Models` on WiMi and confirm an HTTP response is returned. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 180s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending

Wave 0 artifact references:
- `wimi-server/scripts/smoke.sh`
- `.planning/phases/04-decision-delivery-to-operator/04-UI-SMOKE.md`
