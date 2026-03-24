---
phase: 1
slug: trusted-access-and-intake
status: draft
nyquist_compliant: true
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
| **Quick run command** | `cd /home/vadim/diplom/core-backend && go test ./internal/http -run 'TestRequireRoles|TestMe|TestListBySpecialistReturnsCurrentStatuses' -count=1 && go test ./internal/examinations -run 'TestCreateExaminationRejectsMissingQuestionnaire|TestFinishRequiresAllAnswers' -count=1 && go test ./internal/answers -run 'TestCreateAnswerRejectsMismatchedQuestion' -count=1` |
| **Full suite command** | `cd /home/vadim/diplom/core-backend && go test ./... -count=1 && cd /home/vadim/diplom/frontend && npm run lint && npx tsc --noEmit` |
| **Estimated runtime** | ~25 seconds |

---

## Sampling Rate

- **After every task commit:** Run `cd /home/vadim/diplom/core-backend && go test ./internal/http -run 'TestRequireRoles|TestMe|TestListBySpecialistReturnsCurrentStatuses' -count=1 && go test ./internal/examinations -run 'TestCreateExaminationRejectsMissingQuestionnaire|TestFinishRequiresAllAnswers' -count=1 && go test ./internal/answers -run 'TestCreateAnswerRejectsMismatchedQuestion' -count=1`
- **After every plan wave:** Run `cd /home/vadim/diplom/core-backend && go test ./... -count=1 && cd /home/vadim/diplom/frontend && npm run lint && npx tsc --noEmit`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 25 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 1-01-01 | 01 | 1 | ACCS-01, ACCS-02, ACCS-03 | HTTP + service | `cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/auth -run 'TestLogin|TestRefreshRotation|TestLogoutRevokesSession|TestMe' -count=1` | ❌ W0 | ⬜ pending |
| 1-01-02 | 01 | 1 | ACCS-01, ACCS-02, ACCS-03 | docs + tests | `cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/auth -run 'TestLogin|TestRefreshRotation|TestLogoutRevokesSession|TestMe' -count=1` | ❌ W0 | ⬜ pending |
| 1-02-01 | 02 | 2 | ACCS-04 | middleware + HTTP | `cd /home/vadim/diplom/core-backend && go test ./internal/http -run 'TestRequireRoles|TestAdminRoutes|TestOperatorRoutes' -count=1` | ❌ W0 | ⬜ pending |
| 1-02-02 | 02 | 2 | ACCS-04 | docs + tests | `cd /home/vadim/diplom/core-backend && go test ./internal/http -run 'TestRequireRoles|TestAdminRoutes|TestOperatorRoutes' -count=1` | ❌ W0 | ⬜ pending |
| 1-03-01 | 03 | 2 | ACCS-02, ACCS-03 | frontend BFF | `cd /home/vadim/diplom/frontend && npm run lint && npx tsc --noEmit` | ❌ W0 | ⬜ pending |
| 1-03-02 | 03 | 2 | ACCS-02, ACCS-03 | docs | `cd /home/vadim/diplom/frontend && npm run lint && npx tsc --noEmit` | ❌ W0 | ⬜ pending |
| 1-04-01 | 04 | 3 | EXAM-01 | service + repository | `cd /home/vadim/diplom/core-backend && go test ./internal/examinations -run 'TestCreateExaminationSnapshotsQuestions|TestCreateExaminationRejectsMissingQuestionnaire|TestCreateExaminationRejectsInvalidQuestionnaire' -count=1` | ❌ W0 | ⬜ pending |
| 1-04-02 | 04 | 3 | EXAM-01 | docs | `cd /home/vadim/diplom/core-backend && go test ./internal/examinations -run 'TestCreateExaminationSnapshotsQuestions|TestCreateExaminationRejectsMissingQuestionnaire|TestCreateExaminationRejectsInvalidQuestionnaire' -count=1` | ❌ W0 | ⬜ pending |
| 1-05-01 | 05 | 4 | EXAM-02 | service + repository | `cd /home/vadim/diplom/core-backend && go test ./internal/answers -run 'TestCreateAnswerPersistsQuestionLink|TestCreateAnswerRejectsMismatchedQuestion|TestCreateAnswerRejectsDuplicateQuestionAnswer' -count=1` | ❌ W0 | ⬜ pending |
| 1-05-02 | 05 | 4 | EXAM-02 | docs | `cd /home/vadim/diplom/core-backend && go test ./internal/answers -run 'TestCreateAnswerPersistsQuestionLink|TestCreateAnswerRejectsMismatchedQuestion|TestCreateAnswerRejectsDuplicateQuestionAnswer' -count=1` | ❌ W0 | ⬜ pending |
| 1-06-01 | 06 | 5 | EXAM-03 | service + HTTP | `cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/examinations -run 'TestFinishIsIdempotent|TestFinishRequiresAllAnswers|TestListBySpecialistReturnsCurrentStatuses' -count=1` | ❌ W0 | ⬜ pending |
| 1-06-02 | 06 | 5 | EXAM-03 | docs | `cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/examinations -run 'TestFinishIsIdempotent|TestFinishRequiresAllAnswers|TestListBySpecialistReturnsCurrentStatuses' -count=1` | ❌ W0 | ⬜ pending |
| 1-07-01 | 07 | 6 | EXAM-04 | backend + frontend | `cd /home/vadim/diplom/core-backend && go test ./internal/http -run TestListBySpecialistReturnsCurrentStatuses -count=1 && cd /home/vadim/diplom/frontend && npm run lint && npx tsc --noEmit` | ❌ W0 | ⬜ pending |
| 1-07-02 | 07 | 6 | EXAM-04 | docs | `cd /home/vadim/diplom/core-backend && go test ./internal/http -run TestListBySpecialistReturnsCurrentStatuses -count=1 && cd /home/vadim/diplom/frontend && npm run lint && npx tsc --noEmit` | ❌ W0 | ⬜ pending |
| 1-08-01 | 08 | 3 | ACCS-02, ACCS-03 | frontend session UX | `cd /home/vadim/diplom/frontend && npm run lint && npx tsc --noEmit` | ❌ W0 | ⬜ pending |
| 1-08-02 | 08 | 3 | ACCS-02, ACCS-03 | docs | `cd /home/vadim/diplom/frontend && npm run lint && npx tsc --noEmit` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `/home/vadim/diplom/core-backend/internal/http/auth_handler_test.go` — login, refresh, logout, `/me`
- [ ] `/home/vadim/diplom/core-backend/internal/http/rbac_test.go` — wrong-role rejection and allowed-role success
- [ ] `/home/vadim/diplom/core-backend/internal/auth/service_test.go` — refresh rotation, revocation, inactive-user checks
- [ ] `/home/vadim/diplom/core-backend/internal/examinations/service_test.go` — create snapshot, missing-questionnaire rejection, finish completeness, idempotent finish
- [ ] `/home/vadim/diplom/core-backend/internal/answers/service_test.go` — question linkage and duplicate-answer handling
- [ ] `/home/vadim/diplom/core-backend/internal/http/examinations_handler_test.go` — authoritative specialist history endpoint after finish
- [ ] `/home/vadim/diplom/frontend/app/api/auth/login/route.ts` and related BFF route handlers — auth transport boundary
- [ ] `/home/vadim/diplom/frontend/hooks/use-current-user.ts` and protected layouts — session bootstrap UX
- [ ] Minimal repository integration harness for transaction-sensitive behavior — otherwise finish/idempotency tests will be too mock-heavy

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Frontend session redirect after refresh/logout boundary changes | ACCS-02, ACCS-03 | Depends on cookie transport and browser navigation behavior across Next.js layouts and BFF handlers | Log in as `admin` and `operator`, refresh the browser, confirm protected routes still resolve correctly through `/api/auth/session`, then logout and verify both `/admin/*` and `/operator/*` redirect to login |
| Examination intake happy path with assigned questionnaire | EXAM-01, EXAM-02, EXAM-03, EXAM-04 | End-to-end operator flow crosses browser recording, upload, backend persistence, and status rendering | Create specialist, create examination with questionnaire, upload one answer per assigned question, finish once, re-submit finish, then reopen specialist history and confirm statuses remain authoritative |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
