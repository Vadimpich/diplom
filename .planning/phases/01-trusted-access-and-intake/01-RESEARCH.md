# Phase 1: Trusted Access And Intake - Research

**Researched:** 2026-03-20
**Domain:** Backend auth/session lifecycle, server-side RBAC, examination intake integrity
**Confidence:** MEDIUM

## User Constraints

No phase-specific `CONTEXT.md` exists yet. Planning must therefore treat the following as fixed constraints from `docs/00_project.md`, `.planning/ROADMAP.md`, `.planning/REQUIREMENTS.md`, and `.planning/PROJECT.md`:

### Locked Decisions

- Phase 1 scope is `ACCS-01`, `ACCS-02`, `ACCS-03`, `ACCS-04`, `EXAM-01`, `EXAM-02`, `EXAM-03`, `EXAM-04`.
- PostgreSQL remains the source of truth for auth/session state and examination workflow state.
- Backend must enforce role boundaries; frontend-only hiding is insufficient.
- Refresh flow and logout with refresh-token revocation are required in this phase.
- Examination completion must be idempotent and must not allow duplicate processing starts.
- Each answer must be linked to the examination, question, and specialist.
- RabbitMQ carries metadata only; binary audio stays in S3-compatible storage.
- Any contract change must update `docs/01_contract.md`.

### Claude's Discretion

- Choose the concrete refresh-session schema and token rotation mechanics.
- Choose the exact RBAC middleware/policy structure in the existing `chi` router.
- Choose the intake persistence pattern that makes answer linkage and finish idempotency reliable.
- Choose the minimum viable test harness additions for this phase.

### Deferred Ideas (OUT OF SCOPE)

- RabbitMQ publication and async ML orchestration.
- Aggregation, baseline, KESMI delivery, and final result rendering.
- Broader observability and audit implementation beyond what Phase 1 needs to avoid blocking later phases.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| ACCS-01 | Пользователь может получить access token по логину и паролю через защищённый auth flow | Short-lived access JWT, password hash verification, server-managed cookie/session boundary |
| ACCS-02 | Пользователь может обновить access token через refresh flow без повторного ввода логина и пароля | Rotating refresh-session table with hashed tokens, `/auth/refresh` contract, cookie transport |
| ACCS-03 | Пользователь может завершить сессию через logout с отзывом refresh token | Session revocation fields and `/auth/logout` invalidation flow |
| ACCS-04 | Backend ограничивает доступ к admin и operator возможностям по ролям на стороне сервера, а не только во frontend | Route-group RBAC middleware and negative-path API tests |
| EXAM-01 | Оператор может создать обследование для выбранного специалиста и опросника | Keep `POST /examinations`, add exam-question snapshot at creation |
| EXAM-02 | Оператор может загружать отдельные аудиоответы с привязкой к обследованию, вопросу и специалисту | Add question linkage in DB/API, preferably via exam-scoped question snapshot |
| EXAM-03 | Оператор может завершить обследование, и backend идемпотентно переводит его в обработку только один раз | Transactional finish with row lock and one-time processing-launch fence |
| EXAM-04 | Система хранит и отображает историю обследований специалиста с актуальными статусами workflow | Keep PostgreSQL-authoritative status reads; do not derive workflow state from frontend cookies or local draft state |
</phase_requirements>

## Summary

The repo already has a working brownfield baseline for login, CRUD, examination creation, and answer upload, but it does not yet satisfy the trust model required by Phase 1. The backend currently authenticates any JWT and then exposes admin, operator, and intake endpoints inside the same authenticated route group. The frontend tries to compensate with role cookies, but that is not a security boundary. Current session handling also has no refresh or logout lifecycle, and the access token is stored in a JavaScript-readable cookie.

The intake path also stops short of the product requirement. `POST /answers` stores answers only by `examination_id`, while `docs/00_project.md` requires every answer to be linked to the examination, question, and specialist. That gap matters immediately because questionnaire edits currently recreate question rows, so planning cannot safely point historical answers at mutable questionnaire rows. Phase 1 should therefore harden auth and intake together: introduce a DB-backed rotating refresh-session model, move authorization to backend route groups, snapshot assigned questions when an examination is created, and make `finish` transactional so later queue publication can be added without duplicate launches.

**Primary recommendation:** Plan Phase 1 around four backend-first changes: rotating refresh sessions in PostgreSQL, route-level RBAC middleware, examination-question snapshotting for answer linkage, and a transactional finish fence that guarantees one processing start.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/go-chi/chi/v5` | `v5.2.5` (verified 2026-02-05) | HTTP routing and middleware groups | Fits the existing router and makes role-scoped groups straightforward |
| `github.com/golang-jwt/jwt/v5` | `v5.3.1` (verified 2026-01-28) | Short-lived access JWT issuance and parsing | Already in use; sufficient for access tokens while refresh stays DB-backed |
| `github.com/jackc/pgx/v5` | `v5.8.0` (verified 2025-12-26) | PostgreSQL access, transactions, row locking | Required for authoritative session revocation and idempotent workflow transitions |
| `golang.org/x/crypto/bcrypt` | `v0.39.0` (repo version) | Password hashing | Already wired into auth service; acceptable per product doc guidance |
| Go stdlib `crypto/rand` + `crypto/sha256` | builtin | Opaque refresh token generation and hashing | Avoids adding a session library when the session model is domain-specific |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Go stdlib `net/http/httptest` | builtin | HTTP middleware and handler tests | RBAC, login, refresh, logout, and finish endpoint tests |
| Go stdlib `testing` | builtin | Unit and repository-facing tests | Service and middleware behavior |
| Next.js App Router | repo uses `15.2.4`; npm latest is `15.5.14` on 2026-03-19 | Frontend auth boundary and protected layouts | Use existing app; do not upgrade as part of Phase 1 unless blocked by auth delivery mechanics |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| DB-backed opaque refresh sessions | Self-contained long-lived JWT refresh tokens | Easier to issue, harder to revoke safely; conflicts with logout/revocation requirement |
| Backend-only direct API auth with JS-readable cookies | Next.js BFF/route handlers managing HttpOnly cookies on frontend origin | BFF is more secure and SSR-friendly, but increases frontend touch scope |
| Linking answers directly to `questionnaire_questions` | `examination_questions` snapshot table | Snapshotting is safer because questionnaire edits currently recreate rows |
| Frontend button-disabling for dedupe | PostgreSQL transaction + unique fence row | UI-only dedupe is not a correctness guarantee |

**Installation:**

```bash
# No new runtime dependency is strictly required for the backend design.
# Version verification commands used during research:
cd /home/vadim/diplom/core-backend
go list -m -json github.com/go-chi/chi/v5@latest
go list -m -json github.com/golang-jwt/jwt/v5@latest
go list -m -json github.com/jackc/pgx/v5@latest

cd /home/vadim/diplom/frontend
npm view next version
```

**Version verification:** Research used live registry queries on 2026-03-20. The repo is behind latest on `chi`, `jwt/v5`, `pgx`, and `next`, but Phase 1 does not require a dependency upgrade if current APIs are sufficient.

## Architecture Patterns

### Recommended Project Structure
```text
core-backend/
├── internal/auth/                 # login, refresh, logout, refresh-session domain
├── internal/http/                 # auth handlers, auth middleware, role middleware
├── internal/examinations/         # exam creation/finish transaction rules
├── internal/answers/              # answer upload with exam-question linkage
├── db/queries/                    # SQL for sessions, exam question snapshots, idempotent finish
└── migrations/                    # schema changes for Phase 1
frontend/
├── app/api/auth/                  # preferred BFF boundary if Phase 1 removes JS-readable cookies
└── app/(app)/...                  # route guards become UX-only, not security controls
```

### Pattern 1: Access JWT + Rotating Refresh Session
**What:** Keep access tokens short-lived and stateless; keep refresh state server-side in PostgreSQL with one row per refresh session and hashed token material.
**When to use:** All authenticated user sessions.
**Example:**
```go
// Source: OWASP Session Management Cheat Sheet + local auth stack
type RefreshSession struct {
    ID            uuid.UUID
    UserID        int64
    TokenHash     []byte
    ExpiresAt     time.Time
    RevokedAt     *time.Time
    ReplacedByID  *uuid.UUID
    CreatedByIP   string
    UserAgent     string
}

// Login:
// 1. verify password
// 2. issue short-lived access JWT
// 3. generate opaque refresh token with crypto/rand
// 4. store SHA-256 hash in PostgreSQL
// 5. return/set refresh token through HttpOnly cookie transport
```

### Pattern 2: Route-Group RBAC in `chi`
**What:** Split authenticated routes into role-scoped groups and add middleware that checks `claims.RoleSlug`.
**When to use:** All admin-only and operator-only endpoints.
**Example:**
```go
// Source: existing router structure in core-backend/internal/http/router.go
router.Group(func(private chi.Router) {
    private.Use(AuthMiddleware(tokens))

    private.Get("/me", authHandler.Me)

    private.Group(func(admin chi.Router) {
        admin.Use(RequireRoles("admin"))
        admin.Get("/users", authHandler.ListUsers)
        admin.Post("/users", authHandler.CreateUser)
        admin.Put("/users/{id}", authHandler.UpdateUser)
        admin.Get("/questionnaires", questionnairesHandler.List)
        admin.Post("/questionnaires", questionnairesHandler.Create)
        admin.Put("/questionnaires/{id}", questionnairesHandler.Update)
    })

    private.Group(func(operator chi.Router) {
        operator.Use(RequireRoles("operator", "admin"))
        operator.Post("/specialists", specialistsHandler.Create)
        operator.Post("/examinations", examinationsHandler.Create)
        operator.Post("/answers", answersHandler.Create)
        operator.Post("/examinations/{id}/finish", examinationsHandler.Finish)
    })
})
```

### Pattern 3: Examination-Question Snapshot
**What:** Materialize the exact ordered questions assigned to an examination at creation time, then attach answers to the snapshot row, not to mutable questionnaire rows.
**When to use:** Every examination created with a questionnaire.
**Example:**
```sql
-- Source: inference from current questionnaire rewrite behavior + product requirement
CREATE TABLE examination_questions (
    id BIGSERIAL PRIMARY KEY,
    examination_id BIGINT NOT NULL REFERENCES examinations(id) ON DELETE CASCADE,
    specialist_id BIGINT NOT NULL REFERENCES specialists(id) ON DELETE RESTRICT,
    questionnaire_id BIGINT NOT NULL REFERENCES questionnaires(id) ON DELETE RESTRICT,
    source_question_id BIGINT NULL REFERENCES questions(id) ON DELETE SET NULL,
    position INTEGER NOT NULL,
    question_text TEXT NOT NULL,
    UNIQUE (examination_id, position)
);
```

### Pattern 4: Transactional Finish Fence
**What:** `finish` must lock the examination row, validate status and answer completeness, and create exactly one durable "processing launch" marker before moving the status.
**When to use:** `POST /examinations/{id}/finish`.
**Example:**
```sql
-- Source: PostgreSQL explicit locking + ON CONFLICT docs
BEGIN;

SELECT id, status
FROM examinations
WHERE id = $1
FOR UPDATE;

-- validate collecting_answers and mandatory answer count here

INSERT INTO examination_processing_launches (examination_id, launched_at)
VALUES ($1, NOW())
ON CONFLICT (examination_id) DO NOTHING;

UPDATE examinations
SET status = 'ready_for_processing',
    finished_at = COALESCE(finished_at, NOW()),
    updated_at = NOW()
WHERE id = $1;

COMMIT;
```

### Anti-Patterns to Avoid
- **Client-side RBAC as the authority:** The current role cookie and route guard are bypassable and must become UX-only.
- **Long-lived refresh JWT without server state:** It makes logout and revocation weak and complicates compromised-session handling.
- **Linking answers to mutable questionnaire rows only:** Current questionnaire updates recreate rows, so historical linkage can break.
- **Read-then-update finish outside a transaction:** It is enough for today's synchronous status flip, but it will duplicate launches once Phase 2 adds RabbitMQ publication.
- **Bootstrap user mutation as normal admin flow:** `EnsureInitialUser` currently upserts live accounts on startup; Phase 1 planning should avoid depending on that for steady-state user management.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Session revocation | "JWT only" logout that deletes a browser cookie | PostgreSQL refresh-session table with hashed token and revocation columns | Real logout requires server-side invalidation |
| Role enforcement | Frontend route hiding | Backend role middleware on `chi` groups | Security boundary must be server-side |
| Finish dedupe | Disabled submit button or local draft flag | Transaction + row lock + unique launch fence | Dedupe must survive retries and concurrent requests |
| Answer/question linkage | Array index inferred from frontend state | Persisted `examination_questions` snapshot and FK from answers | Historical integrity survives questionnaire edits |
| Session storage | JavaScript-readable persistent auth cookie | HttpOnly cookie transport, ideally on same origin/BFF boundary | Reduces XSS exposure and stale role-cookie drift |

**Key insight:** For this phase, correctness belongs in PostgreSQL constraints/transactions and backend middleware, not in browser state.

## Common Pitfalls

### Pitfall 1: Secure backend auth but insecure browser storage
**What goes wrong:** Access or refresh token is still stored in a JavaScript-readable cookie or local storage.
**Why it happens:** The repo already uses `document.cookie` and direct browser fetches.
**How to avoid:** Plan a cookie boundary deliberately. Preferred: Next.js BFF/route handlers on the frontend origin set `HttpOnly` cookies. Minimum acceptable fallback: keep refresh in `HttpOnly` cookie and never persist access token in JS-readable storage.
**Warning signs:** `document.cookie` writes for auth tokens; frontend role cookie used for authorization decisions.

### Pitfall 2: RBAC only at the handler call site
**What goes wrong:** Some endpoints get checks, others remain exposed because middleware is not centralized.
**Why it happens:** The current router mounts all private routes in one group.
**How to avoid:** Restructure routes by role and make negative-path tests mandatory.
**Warning signs:** New authenticated endpoint added without a role middleware line.

### Pitfall 3: Historical answers tied to mutable questionnaire rows
**What goes wrong:** Editing a questionnaire invalidates the original question linkage for older examinations.
**Why it happens:** Current questionnaire updates recreate question/link rows wholesale.
**How to avoid:** Snapshot questions into `examination_questions` when the examination is created.
**Warning signs:** FK from `answers` points directly to `questions` or `questionnaire_questions` without snapshotting.

### Pitfall 4: Finish is idempotent only by status, not by side effect
**What goes wrong:** Repeated finish requests can become duplicate downstream launches once async publication is added.
**Why it happens:** Current `Finish` does a read followed by a plain status update with no durable launch fence.
**How to avoid:** Add a unique processing-launch record inside the same transaction as the status transition.
**Warning signs:** `finish` implementation has no transaction, no `FOR UPDATE`, and no unique write besides the status field itself.

### Pitfall 5: Missing completeness rule before finish
**What goes wrong:** An examination reaches `ready_for_processing` without one answer per assigned question.
**Why it happens:** Current answers have no question linkage and `finish` does not validate completeness.
**How to avoid:** Count persisted answers against `examination_questions` before allowing finish.
**Warning signs:** `finish` depends only on status and ignores answer count.

## Code Examples

Verified patterns from official and project sources:

### Role Middleware
```go
// Source: local router/auth middleware + chi route groups
func RequireRoles(allowed ...string) func(http.Handler) http.Handler {
    allowedSet := make(map[string]struct{}, len(allowed))
    for _, role := range allowed {
        allowedSet[role] = struct{}{}
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims, ok := ClaimsFromContext(r.Context())
            if !ok {
                writeError(w, http.StatusUnauthorized, "authentication required")
                return
            }
            if _, ok := allowedSet[claims.RoleSlug]; !ok {
                writeError(w, http.StatusForbidden, "insufficient permissions")
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### Refresh Rotation Skeleton
```go
// Source: OWASP session guidance + repo auth service style
func (s *Service) Refresh(ctx context.Context, rawToken string) (LoginResult, string, error) {
    tokenHash := sha256.Sum256([]byte(rawToken))

    session, err := s.repo.GetRefreshSessionByHash(ctx, tokenHash[:])
    if err != nil || session.RevokedAt != nil || time.Now().After(session.ExpiresAt) {
        return LoginResult{}, "", ErrInvalidToken
    }

    nextToken := generateOpaqueToken()
    nextHash := sha256.Sum256([]byte(nextToken))

    err = s.repo.RotateRefreshSession(ctx, RotateRefreshSessionParams{
        SessionID:      session.ID,
        NextTokenHash:  nextHash[:],
        ReplacedAt:     time.Now(),
    })
    if err != nil {
        return LoginResult{}, "", err
    }

    user, err := s.repo.GetUserByID(ctx, session.UserID)
    if err != nil || !user.IsActive {
        return LoginResult{}, "", ErrInactiveUser
    }

    access, ttl, err := s.tokens.Issue(toUser(user))
    if err != nil {
        return LoginResult{}, "", err
    }

    return LoginResult{AccessToken: access, TokenType: "Bearer", ExpiresIn: ttl, User: toUser(user)}, nextToken, nil
}
```

### Finish Completeness Check
```go
// Source: inference from docs/00_project.md answer-link requirement
requiredCount, err := repo.CountExaminationQuestions(ctx, examID)
if err != nil {
    return Examination{}, err
}

actualCount, err := repo.CountDistinctAnsweredQuestions(ctx, examID)
if err != nil {
    return Examination{}, err
}

if actualCount != requiredCount {
    return Examination{}, ErrIncompleteAnswers
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Single access JWT stored in JS-readable cookie | Short-lived access JWT plus server-revocable refresh session | Current best practice reflected in OWASP guidance and project spec | Enables logout, rotation, and lower XSS exposure |
| UI-only route hiding | Backend route-group RBAC | Required by current project requirements | Direct API calls outside role boundary are rejected |
| Question inferred from frontend order only | Persisted exam-question snapshot | Required now because questionnaire rows are mutable in this repo | Historical answer integrity is preserved |
| Idempotency as "same status returned" | Transactional side-effect fence with unique row | Needed before Phase 2 async publication | Prevents duplicate processing starts |

**Deprecated/outdated:**
- Browser-managed auth tokens in `document.cookie` for privileged flows: outdated for this phase's trust goal.
- Direct answer storage without question linkage: incompatible with `docs/00_project.md` data requirements.

## Open Questions

1. **Where should HttpOnly auth cookies live: frontend origin or API origin?**
   - What we know: current app uses direct browser calls to `NEXT_PUBLIC_API_URL`, and current SSR route guards read frontend cookies.
   - What's unclear: whether Phase 1 should introduce a Next.js BFF boundary now or accept a smaller backend-only auth change.
   - Recommendation: prefer Next.js route handlers/BFF so cookies stay on the frontend origin and SSR route guards remain trustworthy. If that is too large for Phase 1, keep the compromise explicit in the plan.

2. **Should operators be allowed to replace an already uploaded answer for the same question?**
   - What we know: the UI supports re-recording before upload, but the persisted model does not define post-upload replacement semantics.
   - What's unclear: whether the business rule is one immutable answer per question or replace-latest.
   - Recommendation: plan one persisted answer per `examination_question` for Phase 1 with a unique constraint; treat replacement as a future explicit feature if needed.

3. **How much of audit logging is needed inside Phase 1 implementation?**
   - What we know: the product spec requires login/logout and examination lifecycle events in audit logs, but full audit scope is Phase 5.
   - What's unclear: whether to add minimal audit hooks now or only leave extension points.
   - Recommendation: at least design auth/session tables and finish flow so audit insertion can be added without rewiring business logic.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + `net/http/httptest` |
| Config file | none |
| Quick run command | `cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/auth ./internal/examinations ./internal/answers -count=1` |
| Full suite command | `cd /home/vadim/diplom/core-backend && go test ./... -count=1 && cd /home/vadim/diplom/frontend && npm run lint && npx tsc --noEmit` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ACCS-01 | Valid login returns access token and authenticated user payload | HTTP + service | `cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/auth -run 'TestLogin|TestMe' -count=1` | ❌ Wave 0 |
| ACCS-02 | Refresh rotates refresh session and returns new access token | service + repository | `cd /home/vadim/diplom/core-backend && go test ./internal/auth -run TestRefreshRotation -count=1` | ❌ Wave 0 |
| ACCS-03 | Logout revokes refresh session and blocks reuse | service + HTTP | `cd /home/vadim/diplom/core-backend && go test ./internal/auth ./internal/http -run TestLogoutRevokesSession -count=1` | ❌ Wave 0 |
| ACCS-04 | Admin/operator endpoints reject wrong role | middleware + HTTP | `cd /home/vadim/diplom/core-backend && go test ./internal/http -run TestRequireRoles -count=1` | ❌ Wave 0 |
| EXAM-01 | Examination creation snapshots assigned questions | service + repository | `cd /home/vadim/diplom/core-backend && go test ./internal/examinations -run TestCreateExaminationSnapshotsQuestions -count=1` | ❌ Wave 0 |
| EXAM-02 | Answer upload persists one answer per assigned question with linkage | service + repository | `cd /home/vadim/diplom/core-backend && go test ./internal/answers -run TestCreateAnswerPersistsQuestionLink -count=1` | ❌ Wave 0 |
| EXAM-03 | Finish is idempotent and emits one processing-launch fence | service + repository | `cd /home/vadim/diplom/core-backend && go test ./internal/examinations -run TestFinishIsIdempotent -count=1` | ❌ Wave 0 |
| EXAM-04 | Specialist history returns authoritative workflow statuses | HTTP + repository | `cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/examinations -run TestListBySpecialistReturnsCurrentStatuses -count=1` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/auth ./internal/examinations ./internal/answers -count=1`
- **Per wave merge:** `cd /home/vadim/diplom/core-backend && go test ./... -count=1`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `/home/vadim/diplom/core-backend/internal/http/auth_handler_test.go` — login, refresh, logout, `/me`
- [ ] `/home/vadim/diplom/core-backend/internal/http/rbac_test.go` — wrong-role rejection and allowed-role success
- [ ] `/home/vadim/diplom/core-backend/internal/auth/service_test.go` — refresh rotation, revocation, inactive-user checks
- [ ] `/home/vadim/diplom/core-backend/internal/examinations/service_test.go` — create snapshot, finish completeness, idempotent finish
- [ ] `/home/vadim/diplom/core-backend/internal/answers/service_test.go` — question linkage and duplicate-answer handling
- [ ] Minimal repository integration harness for transaction-sensitive behavior — otherwise finish/idempotency tests will be too mock-heavy

## Sources

### Primary (HIGH confidence)
- Local product spec: `/home/vadim/diplom/docs/00_project.md` — auth, refresh, RBAC, answer linkage, workflow, and testing requirements
- Local contract: `/home/vadim/diplom/docs/01_contract.md` — current API/schema shape and current gaps
- Local implementation map: `/home/vadim/diplom/.planning/codebase/ARCHITECTURE.md`, `/home/vadim/diplom/.planning/codebase/CONCERNS.md`, `/home/vadim/diplom/.planning/codebase/TESTING.md`
- Current code: `/home/vadim/diplom/core-backend/internal/http/router.go`, `/home/vadim/diplom/core-backend/internal/auth/service.go`, `/home/vadim/diplom/core-backend/internal/examinations/service.go`, `/home/vadim/diplom/core-backend/internal/answers/service.go`
- OWASP Session Management Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
- MDN `Set-Cookie`: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie
- PostgreSQL explicit locking: https://www.postgresql.org/docs/current/explicit-locking.html
- PostgreSQL `INSERT ... ON CONFLICT`: https://www.postgresql.org/docs/current/sql-insert.html

### Secondary (MEDIUM confidence)
- Live module/version verification via `go list -m -json ...@latest` for `chi`, `jwt/v5`, `pgx` on 2026-03-20
- Live npm registry verification via `npm view next version`, `npm view @tanstack/react-query version`, `npm view zod version` on 2026-03-20

### Tertiary (LOW confidence)
- Inference: `examination_questions` snapshot table is not described explicitly in product docs; it is recommended because current questionnaire mutation makes direct historical question references unsafe

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - recommendations mostly extend the existing stack and were version-verified live
- Architecture: MEDIUM - core auth/RBAC direction is clear, but cookie-boundary choice still needs a planning decision
- Pitfalls: HIGH - directly supported by current codebase concerns and product requirements

**Research date:** 2026-03-20
**Valid until:** 2026-04-19
