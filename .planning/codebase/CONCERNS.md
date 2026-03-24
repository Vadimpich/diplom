# Codebase Concerns

**Analysis Date:** 2026-03-20

## Tech Debt

**Authorization and role enforcement are split between frontend cookies and backend JWT parsing:**
- Issue: `core-backend` treats every authenticated user as fully privileged. `AuthMiddleware` only parses the token and stores claims in context, while `NewRouter` mounts `/users`, `/questionnaires`, `/specialists`, `/examinations`, and `/answers` under the same authenticated group with no role checks. The frontend tries to compensate with `AUTH_ROLE_COOKIE` checks in layouts and `RouteGuard`.
- Files: `core-backend/internal/http/auth_middleware.go`, `core-backend/internal/http/router.go`, `frontend/lib/auth.ts`, `frontend/components/layout/route-guard.tsx`, `frontend/app/(app)/admin/layout.tsx`, `frontend/app/(app)/operator/layout.tsx`
- Impact: privilege boundaries are enforced in UI only, which is easy to bypass with direct API calls and creates long-term coupling between security rules and browser state.
- Fix approach: add backend authorization middleware/policies keyed off JWT claims, then reduce frontend role logic to UX routing only.

**Bootstrap user management mutates live accounts on startup:**
- Issue: `EnsureInitialUser` silently returns when bootstrap env is absent, but when it is present it calls `UpsertUser`, which overwrites `password_hash`, `role_id`, sets `is_active = TRUE`, and updates the user on every start.
- Files: `core-backend/internal/auth/service.go`, `core-backend/db/queries/auth.sql`, `core-backend/internal/app/app.go`
- Impact: restarting with changed env can unexpectedly reactivate a disabled user, rotate a password, or escalate/downgrade a role without audit.
- Fix approach: limit bootstrap to first-run creation, or gate updates behind an explicit maintenance command.

**Questionnaire updates recreate all question rows instead of preserving identity:**
- Issue: questionnaire edits delete all links, remove orphan questions, and insert brand new question records for the full set.
- Files: `core-backend/internal/questionnaires/repository.go`, `core-backend/migrations/000002_questionnaires_and_api_support.up.sql`
- Impact: question IDs are unstable, historical references cannot safely point to a specific question revision, and future analytics tied to question identity become hard to implement.
- Fix approach: introduce stable question/version semantics and update only changed links/content.

**Frontend workspace contains generated and package-manager-specific artifacts:**
- Issue: the working tree contains `frontend/node_modules`, multiple `.next*` directories, `frontend/tsconfig.tsbuildinfo`, and both `frontend/package-lock.json` and `frontend/pnpm-lock.yaml`.
- Files: `frontend/node_modules`, `frontend/.next`, `frontend/.next.17026`, `frontend/.next.bak`, `frontend/tsconfig.tsbuildinfo`, `frontend/package-lock.json`, `frontend/pnpm-lock.yaml`
- Impact: repository hygiene is weak, installs are less reproducible, and CI/local runs can diverge depending on which lockfile or cached artifact is used.
- Fix approach: standardize on one package manager, ignore build artifacts, and keep generated directories out of the project tree.

## Known Bugs

**Authenticated operators can call admin APIs directly:**
- Symptoms: any valid JWT can create users, update users, list users, and manage questionnaires because only authentication is enforced server-side.
- Files: `core-backend/internal/http/router.go`, `core-backend/internal/http/auth_middleware.go`, `core-backend/internal/http/auth_handler.go`, `core-backend/internal/http/questionnaires_handler.go`
- Trigger: login as an `operator` and call `/users`, `/users/{id}`, `/questionnaires`, or `/questionnaires/{id}` directly.
- Workaround: none in backend; only the frontend UI tries to hide admin screens.

**Answers are not linked to questionnaire questions:**
- Symptoms: the system stores free-form answers per examination, but there is no `question_id`, `questionnaire_question_id`, or answer position in persistence or API payloads.
- Files: `core-backend/migrations/000001_init_domain.up.sql`, `core-backend/db/queries/answers.sql`, `core-backend/internal/answers/service.go`, `core-backend/internal/http/answers_handler.go`, `frontend/lib/api/client.ts`
- Trigger: run an examination with multiple questions and submit several answers.
- Workaround: none; downstream code cannot reconstruct which answer belongs to which question from persisted data alone.

**Server-side routing can trust stale or forged role cookies until `/me` resolves:**
- Symptoms: admin/operator redirects in Next.js layouts use `AUTH_ROLE_COOKIE`, while the backend is authoritative only after `/me` is fetched client-side.
- Files: `frontend/lib/auth.ts`, `frontend/lib/constants.ts`, `frontend/app/(app)/admin/layout.tsx`, `frontend/app/(app)/operator/layout.tsx`, `frontend/components/layout/route-guard.tsx`
- Trigger: manually alter `diplom_user_role` in the browser or carry an outdated role cookie after role changes.
- Workaround: reload and wait for client-side guard to correct navigation; this does not fix backend authorization.

## Security Considerations

**Access token is stored in a JavaScript-readable cookie:**
- Risk: `persistSession` writes `diplom_access_token` with `document.cookie` and no `HttpOnly` or `Secure` flag.
- Files: `frontend/lib/auth.ts`, `frontend/lib/constants.ts`
- Current mitigation: `SameSite=Lax` is set.
- Recommendations: move token issuance/storage to server-managed `HttpOnly; Secure` cookies and stop exposing raw JWTs to browser JavaScript.

**Privilege checks are missing on sensitive endpoints:**
- Risk: user-management and questionnaire-management endpoints are protected only by JWT presence, not by role.
- Files: `core-backend/internal/http/router.go`, `core-backend/internal/http/auth_middleware.go`
- Current mitigation: frontend route separation in `frontend/app/(app)/admin/*` and `frontend/app/(app)/operator/*`.
- Recommendations: enforce role-based authorization in backend handlers or dedicated middleware, and test both positive and negative access paths.

**Health endpoint can report green while critical dependencies for real workflows are unavailable:**
- Risk: `/health` checks only PostgreSQL even though startup and answer upload depend on MinIO/S3, and the documented target architecture expects more services.
- Files: `core-backend/internal/http/health.go`, `core-backend/internal/app/app.go`, `core-backend/internal/storage/s3.go`, `README.md`
- Current mitigation: none beyond database ping.
- Recommendations: split liveness/readiness and include storage and future queue dependencies in readiness checks.

## Performance Bottlenecks

**Answer uploads are synchronous through the core backend:**
- Problem: `POST /answers` fully parses multipart input and uploads the file to S3 inline before writing the answer row.
- Files: `core-backend/internal/http/answers_handler.go`, `core-backend/internal/answers/service.go`, `core-backend/internal/storage/s3.go`
- Cause: the backend acts as the upload proxy and completion of the request depends on object storage latency.
- Improvement path: use pre-signed uploads or asynchronous ingestion with explicit completion records and retries.

**Questionnaire edits rewrite the full question set:**
- Problem: every update performs delete-all-links, delete-orphans, and insert-all-questions even for small edits.
- Files: `core-backend/internal/questionnaires/repository.go`
- Cause: update logic is modeled as full replacement.
- Improvement path: diff question sets, preserve row identity, and update only changed records.

## Fragile Areas

**Examination answer capture is incomplete at the data-model level:**
- Files: `core-backend/migrations/000001_init_domain.up.sql`, `core-backend/db/queries/answers.sql`, `frontend/components/operator/media-recorder-card.tsx`, `frontend/lib/api/client.ts`
- Why fragile: the workflow is question-driven in UI, but persistence is examination-driven only. Any later addition of per-question analytics, retries, or reordering will require contract and schema changes across frontend and backend.
- Safe modification: add explicit question linkage before extending result processing or historical reporting.
- Test coverage: no tests cover answer sequencing, duplicate submission, or reconstruction of a full examination transcript.

**Auth flow depends on duplicated state in JWT and cookies:**
- Files: `frontend/lib/auth.ts`, `frontend/components/layout/route-guard.tsx`, `frontend/app/(app)/admin/layout.tsx`, `frontend/app/(app)/operator/layout.tsx`, `core-backend/internal/auth/jwt.go`
- Why fragile: role and auth state are stored in separate browser cookies and revalidated later by `/me`. State drift causes redirect glitches and hides the real missing backend authorization problem.
- Safe modification: centralize auth state on server-validated session cookies and keep role derivation from one source.
- Test coverage: no frontend or integration tests cover login persistence, role change propagation, or unauthorized navigation.

**Application startup is tightly coupled to all infra being available:**
- Files: `core-backend/internal/app/app.go`, `core-backend/internal/storage/s3.go`, `core-backend/internal/migrations/runner.go`, `core-backend/cmd/api/main.go`
- Why fragile: startup applies migrations, bootstraps auth, initializes S3, and only then starts HTTP. A temporary storage issue prevents even non-upload endpoints from coming up.
- Safe modification: separate boot phases, degrade non-critical integrations when possible, and expose readiness independently.
- Test coverage: no startup integration tests exercise partial dependency failure.

## Scaling Limits

**Current processing path is single-service and synchronous:**
- Current capacity: one core backend instance handles auth, CRUD, examination state changes, and audio upload in-request.
- Limit: throughput is bounded by HTTP worker availability and object-storage latency; there is no queue-backed buffering or worker pool for later analysis stages.
- Scaling path: introduce queue-backed ingestion and dedicated workers before adding ML processing loads.
- Files: `core-backend/internal/http/router.go`, `core-backend/internal/http/answers_handler.go`, `core-backend/internal/answers/service.go`, `README.md`

**Examination results domain stops at `ready_for_processing`:**
- Current capacity: examination status model ends at `created`, `collecting_answers`, and `ready_for_processing`.
- Limit: there is no persisted state for queued, processing, failed, completed, or delivered results, so operational recovery and retry flows do not exist yet.
- Scaling path: extend status model and orchestration records before introducing RabbitMQ/ML services.
- Files: `core-backend/internal/examinations/service.go`, `core-backend/migrations/000001_init_domain.up.sql`, `frontend/app/(app)/operator/examinations/[id]/results/page.tsx`

## Dependencies at Risk

**Frontend dependency resolution is ambiguous because two lockfiles exist:**
- Risk: `npm` and `pnpm` can resolve different transitive trees for the same `frontend/package.json`.
- Impact: CI, Docker, and local machines can build against different dependency graphs and produce hard-to-reproduce failures.
- Migration plan: keep one lockfile, remove the other, and document the chosen package manager in `README.md`.
- Files: `frontend/package.json`, `frontend/package-lock.json`, `frontend/pnpm-lock.yaml`

## Missing Critical Features

**Asynchronous orchestration and ML pipeline are not implemented in code:**
- Problem: the target architecture in `docs/00_project.md` requires RabbitMQ, ML services, aggregator, baseline, and external integration layers, but the codebase contains only synchronous core CRUD/upload flow and placeholder UI for results.
- Blocks: real processing, retries, idempotent task handling, aggregated result delivery, and operator-facing final decisions.
- Files: `core-backend/internal`, `frontend/app/(app)/operator/examinations/[id]/results/page.tsx`, `frontend/app/(app)/operator/specialists/[id]/page.tsx`, `README.md`

**Operator result screens are placeholders, not working result views:**
- Problem: frontend explicitly states that baseline, explanations, and final decision await backend contracts.
- Blocks: acceptance of the main business scenario described in `docs/00_project.md`.
- Files: `frontend/app/(app)/operator/examinations/[id]/results/page.tsx`, `frontend/app/(app)/operator/specialists/[id]/page.tsx`, `frontend/app/(app)/operator/page.tsx`

## Test Coverage Gaps

**Business-critical backend flows are effectively untested:**
- What's not tested: login lifecycle, role restrictions, user CRUD, specialist CRUD, examination transitions, answer upload, S3 failure handling, and bootstrap behavior.
- Files: `core-backend/internal/auth`, `core-backend/internal/specialists`, `core-backend/internal/examinations`, `core-backend/internal/answers`, `core-backend/internal/app/app.go`
- Risk: regressions in auth, data integrity, and upload flows will reach runtime unnoticed.
- Priority: High

**Frontend operator/admin flows have no automated coverage:**
- What's not tested: login, route protection, operator examination flow, admin forms, monitoring page, and error states.
- Files: `frontend/app/(auth)/login/page.tsx`, `frontend/app/(app)/operator`, `frontend/app/(app)/admin`, `frontend/components/layout/route-guard.tsx`
- Risk: UI regressions and contract mismatches will only be found manually.
- Priority: High

**Existing tests cover only CORS preflight behavior:**
- What's not tested: everything beyond `OPTIONS` handling on login routes.
- Files: `core-backend/internal/http/cors_test.go`
- Risk: the presence of a small test file can create false confidence while the primary domain remains uncovered.
- Priority: Medium

---

*Concerns audit: 2026-03-20*
