# Architecture

**Analysis Date:** 2026-03-20

## Pattern Overview

**Overall:** Modular monorepo with two deployable applications and a layered backend.

**Key Characteristics:**
- `frontend/` and `core-backend/` are developed in one repository but deployed as separate services from `docker-compose.yml`.
- `core-backend/` uses a clear HTTP handler -> domain service -> repository/storage dependency direction wired in `core-backend/internal/app/app.go`.
- `frontend/` is route-driven with Next.js App Router in `frontend/app/`, and all server communication goes through the shared client in `frontend/lib/api/client.ts`.

## Layers

**Frontend App Layer:**
- Purpose: Render operator and admin interfaces, enforce route-level access, and orchestrate browser-side API calls.
- Location: `frontend/app/`, `frontend/components/`, `frontend/hooks/`, `frontend/lib/`
- Contains: Next.js layouts/pages, shared UI, route guards, query hooks, cookie session helpers, typed API client.
- Depends on: `frontend/lib/api/client.ts`, `frontend/lib/auth.ts`, `@tanstack/react-query`, `next/headers`, browser cookies.
- Used by: End users through the `frontend` container in `docker-compose.yml`.

**Frontend API Boundary:**
- Purpose: Keep backend contract usage centralized.
- Location: `frontend/lib/api/client.ts`, `frontend/lib/api/types.ts`
- Contains: Request wrapper, auth header injection, typed request methods, shared DTO definitions.
- Depends on: `NEXT_PUBLIC_API_URL`, browser `fetch`, cookies named in `frontend/lib/constants.ts`.
- Used by: Route components such as `frontend/app/(auth)/login/page.tsx`, `frontend/app/(app)/operator/examinations/new/page.tsx`, and `frontend/app/(app)/admin/users/page.tsx`.

**HTTP Transport Layer:**
- Purpose: Expose backend endpoints, apply middleware, decode/encode transport payloads, and translate domain errors to HTTP responses.
- Location: `core-backend/internal/http/`
- Contains: Router setup, middleware, handler structs per domain, health endpoint helpers.
- Depends on: Domain services from `core-backend/internal/auth`, `core-backend/internal/specialists`, `core-backend/internal/examinations`, `core-backend/internal/questionnaires`, `core-backend/internal/answers`.
- Used by: `core-backend/internal/app/app.go` through `core-backend/internal/http/router.go`.

**Domain Service Layer:**
- Purpose: Validate input, enforce state transitions, and keep domain rules outside handlers and repositories.
- Location: `core-backend/internal/auth/service.go`, `core-backend/internal/specialists/service.go`, `core-backend/internal/examinations/service.go`, `core-backend/internal/questionnaires/service.go`, `core-backend/internal/answers/service.go`
- Contains: Input structs, domain models, validation, status machine checks, orchestration between repository and storage.
- Depends on: Domain-specific repository interfaces and, for answers, a storage interface.
- Used by: HTTP handlers only.

**Persistence Layer:**
- Purpose: Persist structured state in PostgreSQL and hide SQL details from services.
- Location: `core-backend/internal/*/repository.go`, `core-backend/internal/postgres/client.go`, `core-backend/db/queries/`, `core-backend/db/sqlc/`
- Contains: Repository implementations, SQL queries, generated `sqlc` code, pool bootstrap.
- Depends on: `pgx`, `pgxpool`, generated queries in `core-backend/db/sqlc/`.
- Used by: Domain services created in `core-backend/internal/app/app.go`.

**Object Storage Layer:**
- Purpose: Store uploaded answer audio outside PostgreSQL.
- Location: `core-backend/internal/storage/s3.go`
- Contains: MinIO/S3 client bootstrap, bucket initialization, upload/delete operations.
- Depends on: `MINIO_*` runtime config and `minio-go`.
- Used by: `core-backend/internal/answers/service.go`.

**Bootstrap and Runtime Layer:**
- Purpose: Assemble infrastructure, apply migrations, create services, and run the HTTP server.
- Location: `core-backend/cmd/api/main.go`, `core-backend/internal/app/app.go`, `core-backend/internal/config/config.go`, `core-backend/internal/migrations/runner.go`
- Contains: Config loading, DB connection, migration runner, initial user bootstrap, server lifecycle.
- Depends on: Environment variables, PostgreSQL, MinIO/S3, JWT secrets.
- Used by: The `core-backend` process and container.

## Data Flow

**Login Flow:**

1. `frontend/app/(auth)/login/page.tsx` submits credentials through `apiClient.login()` in `frontend/lib/api/client.ts`.
2. `core-backend/internal/http/auth_handler.go` passes the payload to `core-backend/internal/auth/service.go`.
3. `core-backend/internal/auth/service.go` loads the user from `core-backend/internal/auth/repository.go`, verifies the bcrypt hash, and issues JWT via `core-backend/internal/auth/jwt.go`.
4. The frontend stores token and role in cookies through `frontend/lib/auth.ts`, then redirects by role.

**Protected App Navigation Flow:**

1. `frontend/app/page.tsx` reads the role cookie server-side and redirects into `/admin/*` or `/operator/*`.
2. `frontend/app/(app)/admin/layout.tsx` and `frontend/app/(app)/operator/layout.tsx` enforce the presence of a token before rendering.
3. `frontend/components/layout/route-guard.tsx` calls `useCurrentUser()` from `frontend/hooks/use-current-user.ts`.
4. `apiClient.me()` requests `GET /me`, and the backend validates JWT in `core-backend/internal/http/auth_middleware.go`.

**Examination Creation and Answer Upload Flow:**

1. `frontend/app/(app)/operator/examinations/new/page.tsx` loads specialists and questionnaires from `/specialists` and `/questionnaires`.
2. The same page creates an examination through `POST /examinations`, handled in `core-backend/internal/http/examinations_handler.go`.
3. `core-backend/internal/examinations/service.go` validates IDs and persists the initial `created` status via `core-backend/internal/examinations/repository.go`.
4. Answer uploads go to `POST /answers` through `frontend/lib/api/client.ts` as `FormData`.
5. `core-backend/internal/answers/service.go` verifies that the examination is in `collecting_answers`, uploads audio with `core-backend/internal/storage/s3.go`, then stores answer metadata in PostgreSQL.

**Schema Evolution Flow:**

1. SQL migrations live in `core-backend/migrations/`.
2. `core-backend/internal/migrations/runner.go` applies every `*.up.sql` file on startup and records versions in `schema_migrations`.
3. Query definitions in `core-backend/db/queries/` are reflected into generated code in `core-backend/db/sqlc/`.
4. Repository implementations call generated query methods instead of embedding raw SQL in handlers or services.

## State Management

**Frontend State Management:**
- Server-side route gating is done with `cookies()` in `frontend/app/page.tsx`, `frontend/app/(app)/admin/layout.tsx`, and `frontend/app/(app)/operator/layout.tsx`.
- Client-side remote state is handled with TanStack Query in `frontend/components/providers/app-providers.tsx`.
- Session state is browser-cookie based in `frontend/lib/auth.ts`; there is no global Zustand/Redux store.
- Draft examination state is persisted separately in `frontend/lib/examination-drafts.ts`.

**Backend State Management:**
- PostgreSQL is the source of truth for users, specialists, examinations, answers, questionnaires, and roles via `core-backend/migrations/000001_init_domain.up.sql` and `core-backend/migrations/000002_questionnaires_and_api_support.up.sql`.
- Audio binaries are externalized to MinIO/S3 and referenced by `answers.audio_s3_key`.
- Examination lifecycle is explicit in `core-backend/internal/examinations/service.go`: `created` -> `collecting_answers` -> `ready_for_processing`.

## Key Abstractions

**Service + Repository Pairing:**
- Purpose: Keep domain rules testable and transport-agnostic.
- Examples: `core-backend/internal/specialists/service.go` + `core-backend/internal/specialists/repository.go`, `core-backend/internal/questionnaires/service.go` + `core-backend/internal/questionnaires/repository.go`
- Pattern: Services depend on narrow interfaces; repositories own database details.

**Typed Shared API Client:**
- Purpose: Prevent route components from duplicating fetch logic and DTO shapes.
- Examples: `frontend/lib/api/client.ts`, `frontend/lib/api/types.ts`
- Pattern: One `request<T>()` helper centralizes headers, auth, JSON parsing, and error normalization.

**Layout-Based Role Segmentation:**
- Purpose: Separate operator and admin UX without separate frontend apps.
- Examples: `frontend/app/(app)/operator/layout.tsx`, `frontend/app/(app)/admin/layout.tsx`, `frontend/components/layout/app-shell.tsx`
- Pattern: Route groups provide isolated navigation shells with the same provider tree.

## Entry Points

**Backend API Process:**
- Location: `core-backend/cmd/api/main.go`
- Triggers: Container start or local `go run`.
- Responsibilities: Load config, build `App`, and run the HTTP server until shutdown.

**Backend Composition Root:**
- Location: `core-backend/internal/app/app.go`
- Triggers: Called from `core-backend/cmd/api/main.go`.
- Responsibilities: Connect PostgreSQL, apply migrations, bootstrap initial user, initialize S3, construct services, mount router.

**Frontend Root Layout:**
- Location: `frontend/app/layout.tsx`
- Triggers: Every Next.js page render.
- Responsibilities: Load global CSS and mount `QueryClientProvider` through `frontend/components/providers/app-providers.tsx`.

**Frontend Root Redirect:**
- Location: `frontend/app/page.tsx`
- Triggers: Requests to `/`.
- Responsibilities: Redirect based on the role cookie.

## Error Handling

**Strategy:** Fail fast at bootstrap, validate at the service layer, and convert domain/infrastructure errors to transport-safe HTTP responses.

**Patterns:**
- Startup failures call `log.Fatalf` in `core-backend/cmd/api/main.go`.
- Handlers use helpers such as `decodeJSON`, `writeJSON`, `writeError`, and `mapDomainError` in `core-backend/internal/http/`.
- Domain services return sentinel errors like `ErrInvalidInput`, `ErrInvalidTransition`, and `ErrInactiveUser`.
- `frontend/lib/api/client.ts` converts non-2xx responses into `ApiError` for UI-level rendering.

## Cross-Cutting Concerns

**Logging:** Request logging is middleware-local in `core-backend/internal/http/router.go`; bootstrap logging is in `core-backend/cmd/api/main.go`.

**Validation:** Backend validation is service-local in files such as `core-backend/internal/questionnaires/service.go`; frontend form validation is schema-based in pages like `frontend/app/(auth)/login/page.tsx` with Zod and React Hook Form.

**Authentication:** JWT issuance and parsing live in `core-backend/internal/auth/jwt.go`; backend route protection is in `core-backend/internal/http/auth_middleware.go`; frontend route protection combines cookie redirects with `frontend/components/layout/route-guard.tsx`.

**CORS:** Applied globally before routing in `core-backend/internal/http/router.go` through `core-backend/internal/http/cors.go`.

**Infrastructure Wiring:** Local orchestration is centralized in `docker-compose.yml`, which starts `postgres`, `rabbitmq`, `minio`, `core-backend`, and `frontend`.

---

*Architecture analysis: 2026-03-20*
