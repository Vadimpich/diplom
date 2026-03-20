---
phase: 1
slug: trusted-access-and-intake
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-20
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` + `net/http/httptest` |
| **Config file** | none |
| **Quick run command** | `cd /home/katya/dimplom/core-backend && go test ./internal/http ./internal/auth ./internal/examinations ./internal/answers -count=1` |
| **Full suite command** | `cd /home/katya/dimplom/core-backend && go test ./... -count=1 && cd /home/katya/dimplom/frontend && npm run lint && npx tsc --noEmit` |
| **Estimated runtime** | ~45 seconds |

---

## Sampling Rate

- **After every task commit:** Run `cd /home/katya/dimplom/core-backend && go test ./internal/http ./internal/auth ./internal/examinations ./internal/answers -count=1`
- **After every plan wave:** Run `cd /home/katya/dimplom/core-backend && go test ./... -count=1 && cd /home/katya/dimplom/frontend && npm run lint && npx tsc --noEmit`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 45 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 1-01-01 | 01 | 1 | ACCS-01 | HTTP + service | `cd /home/katya/dimplom/core-backend && go test ./internal/http ./internal/auth -run 'TestLogin|TestMe' -count=1` | ❌ W0 | ⬜ pending |
| 1-01-02 | 01 | 1 | ACCS-02 | service + repository | `cd /home/katya/dimplom/core-backend && go test ./internal/auth -run TestRefreshRotation -count=1` | ❌ W0 | ⬜ pending |
| 1-01-03 | 01 | 1 | ACCS-03 | service + HTTP | `cd /home/katya/dimplom/core-backend && go test ./internal/auth ./internal/http -run TestLogoutRevokesSession -count=1` | ❌ W0 | ⬜ pending |
| 1-02-01 | 02 | 1 | ACCS-04 | middleware + HTTP | `cd /home/katya/dimplom/core-backend && go test ./internal/http -run TestRequireRoles -count=1` | ❌ W0 | ⬜ pending |
| 1-03-01 | 03 | 2 | EXAM-01 | service + repository | `cd /home/katya/dimplom/core-backend && go test ./internal/examinations -run TestCreateExaminationSnapshotsQuestions -count=1` | ❌ W0 | ⬜ pending |
| 1-03-02 | 03 | 2 | EXAM-02 | service + repository | `cd /home/katya/dimplom/core-backend && go test ./internal/answers -run TestCreateAnswerPersistsQuestionLink -count=1` | ❌ W0 | ⬜ pending |
| 1-04-01 | 04 | 2 | EXAM-03 | service + repository | `cd /home/katya/dimplom/core-backend && go test ./internal/examinations -run TestFinishIsIdempotent -count=1` | ❌ W0 | ⬜ pending |
| 1-04-02 | 04 | 2 | EXAM-04 | HTTP + repository | `cd /home/katya/dimplom/core-backend && go test ./internal/http ./internal/examinations -run TestListBySpecialistReturnsCurrentStatuses -count=1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `/home/katya/dimplom/core-backend/internal/http/auth_handler_test.go` — login, refresh, logout, `/me`
- [ ] `/home/katya/dimplom/core-backend/internal/http/rbac_test.go` — wrong-role rejection and allowed-role success
- [ ] `/home/katya/dimplom/core-backend/internal/auth/service_test.go` — refresh rotation, revocation, inactive-user checks
- [ ] `/home/katya/dimplom/core-backend/internal/examinations/service_test.go` — create snapshot, finish completeness, idempotent finish
- [ ] `/home/katya/dimplom/core-backend/internal/answers/service_test.go` — question linkage and duplicate-answer handling
- [ ] Minimal repository integration harness for transaction-sensitive behavior — otherwise finish/idempotency tests will be too mock-heavy

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Frontend session redirect after refresh/logout boundary changes | ACCS-02, ACCS-03 | Depends on cookie transport and browser navigation behavior across Next.js layouts | Log in as `admin` and `operator`, refresh the browser, confirm protected routes still resolve correctly, then logout and verify both `/admin/*` and `/operator/*` redirect to login |
| Examination intake happy path with assigned questionnaire | EXAM-01, EXAM-02, EXAM-03, EXAM-04 | End-to-end operator flow crosses browser recording, upload, backend persistence, and status rendering | Create specialist, create examination with questionnaire, upload one answer per assigned question, finish once, re-submit finish, then reopen specialist history and confirm statuses remain authoritative |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 45s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
