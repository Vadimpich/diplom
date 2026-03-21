---
phase: 2
slug: asynchronous-multichannel-processing
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-03-20
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` + `net/http/httptest`, Next.js lint/build/typecheck, Docker Compose smoke |
| **Config file** | none |
| **Quick run command** | `cd /home/katya/dimplom/core-backend && go test ./internal/processing ./internal/http -run 'TestFinishCreatesOutboxForMandatoryChannels|TestProcessingStatusEndpoint' -count=1 && go test ./internal/channelresults ./internal/processing -run 'TestIndependentChannelCompletion|TestRetryBudget|TestFatalVsTemporaryError|TestMandatoryChannelExhaustionFailsExamination' -count=1` |
| **Full suite command** | `cd /home/katya/dimplom && docker compose up -d --build && docker compose ps && cd /home/katya/dimplom/core-backend && go test ./... -count=1 && cd /home/katya/dimplom/frontend && npm run lint && npm run build && npx tsc --noEmit` |
| **Estimated runtime** | ~180 seconds |

---

## Sampling Rate

- **After every task commit:** Run `cd /home/katya/dimplom/core-backend && go test ./internal/processing ./internal/http -run 'TestFinishCreatesOutboxForMandatoryChannels|TestProcessingStatusEndpoint' -count=1 && go test ./internal/channelresults ./internal/processing -run 'TestIndependentChannelCompletion|TestRetryBudget|TestFatalVsTemporaryError|TestMandatoryChannelExhaustionFailsExamination' -count=1`
- **After every plan wave:** Run `cd /home/katya/dimplom && docker compose up -d --build && docker compose ps && cd /home/katya/dimplom/core-backend && go test ./... -count=1 && cd /home/katya/dimplom/frontend && npm run lint && npm run build && npx tsc --noEmit`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 180 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 2-01-01 | 01 | 1 | PIPE-01 | contract + behavior pinning | `cd /home/katya/dimplom/core-backend && go test ./internal/processing ./internal/http -run 'TestFinishCreatesOutboxForMandatoryChannels|TestProcessingStatusEndpoint' -count=1` | ❌ W0 | ⬜ pending |
| 2-01-02 | 01 | 1 | PIPE-01 | schema + projection | `cd /home/katya/dimplom/core-backend && go test ./internal/processing ./internal/http -run 'TestFinishCreatesOutboxForMandatoryChannels|TestProcessingStatusEndpoint' -count=1` | ❌ W0 | ⬜ pending |
| 2-02-01 | 02 | 2 | PIPE-01 | finish orchestration | `cd /home/katya/dimplom/core-backend && go test ./internal/processing ./internal/http -run 'TestFinishCreatesOutboxForMandatoryChannels' -count=1` | ❌ W0 | ⬜ pending |
| 2-02-02 | 02 | 2 | PIPE-01, PIPE-03 | AMQP relay + broker semantics | `cd /home/katya/dimplom/core-backend && go test ./internal/processing -run 'TestOutboxRelayPublishesPendingMessages|TestRetryBudget|TestFatalVsTemporaryError' -count=1` | ❌ W0 | ⬜ pending |
| 2-03-01 | 03 | 2 | PIPE-02 | worker shells | `cd /home/katya/dimplom && docker compose up -d --build text-worker acoustic-worker paralinguistic-worker && docker compose ps` | ❌ W0 | ⬜ pending |
| 2-03-02 | 03 | 2 | PIPE-02, QUAL-03 | compose wiring + shared contracts | `cd /home/katya/dimplom && docker compose up -d --build && docker compose ps` | ❌ W0 | ⬜ pending |
| 2-04-01 | 04 | 3 | PIPE-03, PIPE-04 | state transitions | `cd /home/katya/dimplom/core-backend && go test ./internal/channelresults ./internal/processing -run 'TestIndependentChannelCompletion|TestRetryBudget|TestFatalVsTemporaryError|TestMandatoryChannelExhaustionFailsExamination' -count=1` | ❌ W0 | ⬜ pending |
| 2-04-02 | 04 | 3 | PIPE-04 | HTTP status projection | `cd /home/katya/dimplom/core-backend && go test ./internal/http -run 'TestProcessingStatusEndpoint|TestProcessingStatusEndpointReturnsTerminalError' -count=1` | ❌ W0 | ⬜ pending |
| 2-05-01 | 05 | 4 | RSLT-01 | frontend progress polling | `cd /home/katya/dimplom/frontend && npm run lint && npx tsc --noEmit` | ❌ W0 | ⬜ pending |
| 2-05-02 | 05 | 4 | RSLT-01 | UI integration | `cd /home/katya/dimplom/frontend && npm run lint && npx tsc --noEmit && npm run build` | ❌ W0 | ⬜ pending |
| 2-06-01 | 06 | 4 | QUAL-03 | end-to-end stack | `cd /home/katya/dimplom && docker compose up -d --build && docker compose ps && cd /home/katya/dimplom/core-backend && go test ./... -count=1 && cd /home/katya/dimplom/frontend && npm run lint && npm run build && npx tsc --noEmit` | ✅ | ✅ green |
| 2-06-02 | 06 | 4 | QUAL-03 | implementation log | `cd /home/katya/dimplom && rg -n "transactional outbox|channel queues|processing-status|stub worker" docs/02_implementation.md` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `/home/katya/dimplom/core-backend/internal/processing/` — orchestration tests for outbox creation, relay publish, retry budget, and terminal failure projection
- [ ] `/home/katya/dimplom/core-backend/internal/channelresults/` — unified result ingestion tests for independent channel completion and error classification
- [ ] `/home/katya/dimplom/core-backend/internal/http/processing_status_test.go` — API coverage for per-channel progress DTO and terminal error state
- [ ] `/home/katya/dimplom/core-backend/internal/rabbitmq/` test seam or fake publisher/consumer harness — otherwise AMQP topology logic will stay unverified
- [ ] `/home/katya/dimplom/frontend/app/(app)/operator/examinations/[id]/processing/` polling and rendering tests or at minimum lint/type-safe integration coverage
- [ ] Compose service definitions for `text`, `acoustic`, and `paralinguistic` worker containers with health/readiness behavior — required for `QUAL-03`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Operator sees live per-channel progress until terminal completion or error | RSLT-01 | Requires browser polling, navigation, and multi-service runtime coordination across backend, broker, and worker containers | Start local stack, create and finish an examination with uploaded answers, open the operator processing page, and confirm `text`, `acoustic`, and `paralinguistic` statuses move from queued/processing into success or terminal error without page reload hacks |
| Local stack is reproducible with all mandatory services | QUAL-03 | Compose smoke proves services start, but not that they interoperate through one realistic user flow | Run `docker compose up -d --build`, verify all containers are healthy, then finish an examination and confirm worker logs, backend progress endpoint, and UI status all reflect the same pipeline run |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 180s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
