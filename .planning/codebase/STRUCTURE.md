# Codebase Structure

**Analysis Date:** 2026-03-20

## Directory Layout

```text
[project-root]/
├── `core-backend/`      # Go API, domain services, repositories, migrations
├── `frontend/`          # Next.js operator/admin application
├── `docs/`              # Project brief, contracts, implementation log
├── `.planning/codebase/`# Generated mapping documents for future phases
├── `docker-compose.yml` # Local multi-service orchestration
├── `README.md`          # Local stack bootstrap notes
└── `AGENTS.md`          # Agent workflow and documentation rules
```

## Directory Purposes

**`core-backend/`:**
- Purpose: Contains the only backend application in the repository.
- Contains: Go module source, Dockerfile, SQL migrations, raw SQL queries, generated `sqlc` code.
- Key files: `core-backend/cmd/api/main.go`, `core-backend/internal/app/app.go`, `core-backend/internal/http/router.go`, `core-backend/go.mod`

**`core-backend/cmd/api/`:**
- Purpose: Backend executable entrypoint.
- Contains: The `main` package only.
- Key files: `core-backend/cmd/api/main.go`

**`core-backend/internal/app/`:**
- Purpose: Composition root for the backend.
- Contains: App bootstrap and lifecycle management.
- Key files: `core-backend/internal/app/app.go`

**`core-backend/internal/http/`:**
- Purpose: HTTP transport layer.
- Contains: Router, middleware, handlers, helper functions, health endpoint, CORS tests.
- Key files: `core-backend/internal/http/router.go`, `core-backend/internal/http/auth_handler.go`, `core-backend/internal/http/examinations_handler.go`, `core-backend/internal/http/cors_test.go`

**`core-backend/internal/auth/`, `core-backend/internal/specialists/`, `core-backend/internal/examinations/`, `core-backend/internal/questionnaires/`, `core-backend/internal/answers/`:**
- Purpose: Domain modules grouped by bounded feature area.
- Contains: Domain models, service logic, repository implementations.
- Key files: `core-backend/internal/auth/service.go`, `core-backend/internal/specialists/service.go`, `core-backend/internal/examinations/service.go`, `core-backend/internal/questionnaires/repository.go`, `core-backend/internal/answers/service.go`

**`core-backend/internal/config/`:**
- Purpose: Runtime config parsing.
- Contains: Environment-backed config loading.
- Key files: `core-backend/internal/config/config.go`

**`core-backend/internal/postgres/`:**
- Purpose: PostgreSQL pool bootstrap and generated query access.
- Contains: Client wrapper around `pgxpool`.
- Key files: `core-backend/internal/postgres/client.go`

**`core-backend/internal/storage/`:**
- Purpose: Object storage adapter.
- Contains: MinIO/S3 client wrapper.
- Key files: `core-backend/internal/storage/s3.go`

**`core-backend/internal/migrations/`:**
- Purpose: Startup migration runner.
- Contains: Runtime migration application code.
- Key files: `core-backend/internal/migrations/runner.go`

**`core-backend/db/queries/`:**
- Purpose: Source SQL for `sqlc`.
- Contains: Per-domain `.sql` query files.
- Key files: `core-backend/db/queries/auth.sql`, `core-backend/db/queries/examinations.sql`, `core-backend/db/queries/questionnaires.sql`

**`core-backend/db/sqlc/`:**
- Purpose: Generated Go data-access code.
- Contains: `sqlc` models and query wrappers.
- Key files: `core-backend/db/sqlc/db.go`, `core-backend/db/sqlc/models.go`, `core-backend/db/sqlc/questionnaires.sql.go`

**`core-backend/migrations/`:**
- Purpose: Database schema history.
- Contains: Ordered `*.up.sql` and `*.down.sql` files.
- Key files: `core-backend/migrations/000001_init_domain.up.sql`, `core-backend/migrations/000002_questionnaires_and_api_support.up.sql`

**`frontend/`:**
- Purpose: Contains the only web client in the repository.
- Contains: Next.js App Router app, components, hooks, API client, Tailwind config, Dockerfile.
- Key files: `frontend/package.json`, `frontend/app/layout.tsx`, `frontend/lib/api/client.ts`, `frontend/next.config.ts`

**`frontend/app/`:**
- Purpose: Route tree and layouts for App Router.
- Contains: Global layout, route groups `(auth)` and `(app)`, operator pages, admin pages.
- Key files: `frontend/app/layout.tsx`, `frontend/app/page.tsx`, `frontend/app/(auth)/login/page.tsx`, `frontend/app/(app)/operator/layout.tsx`, `frontend/app/(app)/admin/layout.tsx`

**`frontend/components/`:**
- Purpose: Reusable UI and application-level client components.
- Contains: `layout/`, `providers/`, `operator/`, `ui/`.
- Key files: `frontend/components/layout/app-shell.tsx`, `frontend/components/layout/route-guard.tsx`, `frontend/components/providers/app-providers.tsx`

**`frontend/hooks/`:**
- Purpose: Shared React Query hooks.
- Contains: Thin wrappers over `apiClient`.
- Key files: `frontend/hooks/use-current-user.ts`

**`frontend/lib/`:**
- Purpose: Shared frontend infrastructure and utilities.
- Contains: API client, DTO types, auth cookie helpers, constants, draft persistence, formatting helpers.
- Key files: `frontend/lib/api/client.ts`, `frontend/lib/api/types.ts`, `frontend/lib/auth.ts`, `frontend/lib/examination-drafts.ts`

**`docs/`:**
- Purpose: Human-maintained project documentation that agents must respect.
- Contains: Project brief, contracts, implementation log.
- Key files: `docs/00_project.md`, `docs/01_contract.md`, `docs/02_implementation.md`

## Key File Locations

**Entry Points:**
- `core-backend/cmd/api/main.go`: Backend process entry.
- `frontend/app/layout.tsx`: Frontend root layout.
- `frontend/app/page.tsx`: Frontend root redirect.
- `docker-compose.yml`: Local environment entry for all services.

**Configuration:**
- `core-backend/internal/config/config.go`: Backend env parsing.
- `core-backend/db/sqlc.yaml`: `sqlc` generation config.
- `frontend/tsconfig.json`: Frontend TypeScript and path alias config.
- `frontend/tailwind.config.ts`: Tailwind setup.
- `frontend/eslint.config.mjs`: Frontend lint config.

**Core Logic:**
- `core-backend/internal/app/app.go`: Backend composition root.
- `core-backend/internal/http/router.go`: Endpoint registration and middleware.
- `core-backend/internal/examinations/service.go`: Examination lifecycle rules.
- `core-backend/internal/answers/service.go`: Audio upload orchestration.
- `frontend/lib/api/client.ts`: Shared HTTP contract boundary.

**Testing:**
- `core-backend/internal/http/cors_test.go`: Current backend test coverage sample.

## Naming Conventions

**Files:**
- Go backend files are snake_case by feature or responsibility: `auth_handler.go`, `service.go`, `repository.go`.
- Next.js route files follow App Router conventions: `page.tsx`, `layout.tsx`.
- Shared frontend component files are kebab-case: `app-shell.tsx`, `route-guard.tsx`, `empty-state.tsx`.

**Directories:**
- Backend domain directories are plural nouns matching resource areas: `core-backend/internal/specialists/`, `core-backend/internal/examinations/`, `core-backend/internal/questionnaires/`.
- Frontend route directories mirror URL structure: `frontend/app/(app)/operator/specialists/[id]/`, `frontend/app/(app)/admin/users/new/`.
- Frontend reusable code is grouped by role, not by route: `frontend/components/ui/`, `frontend/components/layout/`, `frontend/components/operator/`.

## Where to Add New Code

**New Backend API Endpoint:**
- Primary code: add a handler method under the matching module in `core-backend/internal/http/`; wire the route in `core-backend/internal/http/router.go`.
- Domain logic: add or extend the service in the relevant module under `core-backend/internal/<domain>/service.go`.
- Persistence: add SQL in `core-backend/db/queries/*.sql`, regenerate `core-backend/db/sqlc/`, and update the repository in `core-backend/internal/<domain>/repository.go`.
- Schema changes: add a new ordered file in `core-backend/migrations/`.

**New Frontend Screen:**
- Route implementation: create a new `page.tsx` under the correct App Router subtree in `frontend/app/`.
- Shared layout or navigation changes: update `frontend/app/(app)/operator/layout.tsx`, `frontend/app/(app)/admin/layout.tsx`, or `frontend/components/layout/app-shell.tsx`.
- Remote data access: add typed client methods in `frontend/lib/api/client.ts` and DTOs in `frontend/lib/api/types.ts`.

**New Shared Frontend Component:**
- Implementation: put generic UI into `frontend/components/ui/`.
- Feature-specific reusable UI: put operator-specific pieces into `frontend/components/operator/`.
- Cross-route chrome or guards: put them into `frontend/components/layout/`.

**New Frontend Utility or Hook:**
- Hook: add to `frontend/hooks/` if it wraps remote state or repeated React behavior.
- Utility/helper: add to `frontend/lib/`.
- Session or contract-related helper: prefer `frontend/lib/auth.ts` or `frontend/lib/api/`.

**New Backend Integration Adapter:**
- Infrastructure client: place under `core-backend/internal/storage/` for storage-like adapters or create a new `core-backend/internal/<integration>/` module if it has its own domain logic.
- Wiring: construct it centrally in `core-backend/internal/app/app.go`.

## Special Directories

**`core-backend/db/sqlc/`:**
- Purpose: Generated database access layer.
- Generated: Yes.
- Committed: Yes.

**`frontend/.next/`, `frontend/.next.17026/`, `frontend/.next.bak/`:**
- Purpose: Next.js build artifacts and cached outputs.
- Generated: Yes.
- Committed: Not intended for source edits; treat as build output.

**`.planning/codebase/`:**
- Purpose: Repository mapping artifacts consumed by planning/execution workflows.
- Generated: Yes, by mapping tasks.
- Committed: Yes, intended as working documentation.

**`docs/`:**
- Purpose: Manual coordination layer across agents.
- Generated: No.
- Committed: Yes.

## Practical Placement Rules

- Put backend business rules in `core-backend/internal/<domain>/service.go`, not in `core-backend/internal/http/*.go`.
- Put new SQL in `core-backend/db/queries/`, not inline in handlers.
- Treat `core-backend/db/sqlc/` as generated output; regenerate instead of hand-authoring large changes there.
- Put new operator pages under `frontend/app/(app)/operator/` and new admin pages under `frontend/app/(app)/admin/`.
- Put contract-facing frontend fetch logic only in `frontend/lib/api/client.ts`; route files should call client methods or hooks.
- Keep auth and role-routing behavior aligned with `frontend/app/page.tsx`, `frontend/app/(app)/operator/layout.tsx`, and `frontend/app/(app)/admin/layout.tsx`.

---

*Structure analysis: 2026-03-20*
