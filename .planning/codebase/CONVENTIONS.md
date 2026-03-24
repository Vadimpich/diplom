# Coding Conventions

**Analysis Date:** 2026-03-20

## Naming Patterns

**Files:**
- Go backend uses lowercase snake_case-style filenames by responsibility: `core-backend/internal/http/auth_handler.go`, `core-backend/internal/http/auth_middleware.go`, `core-backend/internal/http/cors_test.go`, `core-backend/internal/specialists/service.go`.
- Frontend route files follow Next.js App Router conventions: `frontend/app/(auth)/login/page.tsx`, `frontend/app/(app)/admin/layout.tsx`, `frontend/app/layout.tsx`.
- Frontend shared components use lowercase kebab-case filenames: `frontend/components/layout/route-guard.tsx`, `frontend/components/operator/media-recorder-card.tsx`, `frontend/components/ui/page-header.tsx`.
- Frontend utility and API modules use short lowercase filenames: `frontend/lib/api/client.ts`, `frontend/lib/api/types.ts`, `frontend/lib/auth.ts`, `frontend/hooks/use-current-user.ts`.

**Functions:**
- Go exported constructors and service methods use PascalCase: `NewRouter`, `NewService`, `CreateUser`, `EnsureInitialUser` in `core-backend/internal/http/router.go` and `core-backend/internal/auth/service.go`.
- Go internal helpers use lowerCamelCase: `writeJSON`, `mapDomainError`, `normalizeCreate`, `parseInt64Param` in `core-backend/internal/http/helpers.go` and `core-backend/internal/specialists/service.go`.
- React components use PascalCase for component names and lowerCamelCase for helpers: `LoginPage` in `frontend/app/(auth)/login/page.tsx`, `MediaRecorderCard` and `startRecording` in `frontend/components/operator/media-recorder-card.tsx`.
- Custom hooks use the `use*` prefix: `useCurrentUser` in `frontend/hooks/use-current-user.ts`.

**Variables:**
- Frontend local variables use lowerCamelCase: `searchParams`, `selectedQuestionnaireId`, `createMutation`, `audioBlob` in `frontend/app/(app)/operator/examinations/new/page.tsx` and `frontend/components/operator/media-recorder-card.tsx`.
- Go local variables use short lowerCamelCase names: `cfg`, `db`, `queries`, `tokenManager`, `request`, `item` in `core-backend/internal/app/app.go` and `core-backend/internal/http/specialists_handler.go`.
- Request/response DTO variables are consistently named `request`, `result`, `items`, `item`, `user` in handlers such as `core-backend/internal/http/auth_handler.go` and `core-backend/internal/http/specialists_handler.go`.

**Types:**
- Go domain structs and interfaces use PascalCase nouns: `User`, `Role`, `StoredUser`, `Repository`, `TokenManager` in `core-backend/internal/auth/service.go`.
- Request DTOs in Go handlers are private lowerCamelCase types with `Request`/`Response` suffixes: `loginRequest`, `createUserRequest`, `usersResponse` in `core-backend/internal/http/auth_handler.go`.
- TypeScript API shapes use `interface` and PascalCase names: `User`, `Questionnaire`, `HealthResponse`, `ApiErrorShape` in `frontend/lib/api/types.ts`.
- TypeScript string union types are used for contract enums: `RoleSlug`, `ExaminationStatus` in `frontend/lib/api/types.ts`.

## Code Style

**Formatting:**
- Frontend formatting is enforced indirectly through ESLint only; no dedicated Prettier or Biome config is present in the repository root or `frontend/`.
- Frontend code uses 2-space indentation, semicolons, and double quotes, as shown in `frontend/app/(auth)/login/page.tsx` and `frontend/components/ui/button.tsx`.
- Go code follows standard `gofmt` formatting with tabs, grouped imports, and no custom formatter config, as shown in `core-backend/cmd/api/main.go` and `core-backend/internal/http/router.go`.

**Linting:**
- Frontend uses ESLint flat config in `frontend/eslint.config.mjs`.
- Extend only `next/core-web-vitals` and `next/typescript` from `frontend/eslint.config.mjs`; no project-specific custom rules are added.
- Ignored frontend paths are `.next/**`, `.next.*/**`, and `node_modules/**` via `frontend/eslint.config.mjs`.
- No `golangci-lint` config or `Makefile` wrapper is present for `core-backend/`; quality checks rely on standard Go tooling and code review.

## Import Organization

**Order:**
1. Standard library or framework packages first.
2. Third-party dependencies second.
3. Internal alias imports last.

**Observed patterns:**
- Go files separate stdlib, third-party, and local module imports with blank lines: `core-backend/internal/http/router.go`, `core-backend/internal/http/helpers.go`, `core-backend/internal/auth/service.go`.
- Frontend files usually place framework packages first, then third-party libraries, then `@/` aliases: `frontend/app/(auth)/login/page.tsx`, `frontend/components/operator/media-recorder-card.tsx`.
- Type-only imports are used where helpful but not universally; examples include `import type { ReactNode }` in `frontend/app/(app)/admin/layout.tsx` and `import type { Answer }` in `frontend/components/operator/media-recorder-card.tsx`.

**Path Aliases:**
- Frontend uses the `@/*` alias mapped to project root in `frontend/tsconfig.json`.
- Prefer `@/lib/...`, `@/components/...`, and `@/hooks/...` for non-relative imports, matching `frontend/lib/api/client.ts` and `frontend/components/layout/route-guard.tsx`.
- Go backend uses the module path `diplom/internal/...` for internal packages, matching `core-backend/cmd/api/main.go` and `core-backend/internal/app/app.go`.

## Error Handling

**Patterns:**
- Go service packages define sentinel errors near the top of the file and return them for validation failures: `ErrInvalidCredentials`, `ErrInactiveUser`, `ErrInvalidInput` in `core-backend/internal/auth/service.go`; `ErrInvalidInput` in `core-backend/internal/specialists/service.go`.
- Go HTTP handlers do not format domain errors inline. They call `mapDomainError` and return a JSON error envelope through `writeError`, as in `core-backend/internal/http/auth_handler.go` and `core-backend/internal/http/specialists_handler.go`.
- JSON decoding in handlers uses `decodeJSON`, which enables `DisallowUnknownFields`, so request contracts are intentionally strict in `core-backend/internal/http/helpers.go`.
- Frontend API requests throw a typed `ApiError` with HTTP status and backend message fallback handling in `frontend/lib/api/client.ts`.
- Frontend pages surface async failures through alert components instead of suppressing them, for example `(loginMutation.error as ApiError).message` in `frontend/app/(auth)/login/page.tsx` and `frontend/app/(app)/operator/examinations/new/page.tsx`.

## Logging

**Framework:** `log` package in Go backend; no dedicated frontend logging abstraction detected.

**Patterns:**
- Startup and fatal runtime failures use `log.Fatalf` in `core-backend/cmd/api/main.go`.
- Request-level logging is middleware-based through `requestLogger` in `core-backend/internal/http/router.go`.
- Logs are structured as message plus key-value-like fragments inside a plain string, for example `method=%s path=%s duration=%s` in `core-backend/internal/http/router.go`.
- Frontend code does not use `console.*` in the inspected application files; user-visible alerts are preferred over client-side logging.

## Comments

**When to Comment:**
- Inline comments are rare. The codebase relies on explicit naming and small functions instead of descriptive comments.
- Generated artifacts are identified by location rather than comment banners; `core-backend/db/sqlc/*.go` is generated from `core-backend/db/sqlc.yaml`.

**JSDoc/TSDoc:**
- Not detected in inspected frontend files such as `frontend/lib/api/client.ts`, `frontend/components/ui/button.tsx`, and `frontend/hooks/use-current-user.ts`.
- Go doc comments for exported identifiers are also not routinely used in inspected backend files.

## Function Design

**Size:**
- Backend handlers stay narrow: parse input, call service, map error, write response. Follow the pattern in `core-backend/internal/http/auth_handler.go` and `core-backend/internal/http/specialists_handler.go`.
- Backend business logic lives in services with small normalization helpers such as `normalizeCreate` and `normalizeUpdate` in `core-backend/internal/specialists/service.go`.
- Frontend route components can be moderately large when they own page-level query/mutation wiring, as in `frontend/app/(app)/operator/examinations/new/page.tsx`.

**Parameters:**
- Go service methods take explicit input structs for non-trivial operations: `LoginInput`, `CreateUserInput`, `UpdateInput` in `core-backend/internal/auth/service.go` and `core-backend/internal/specialists/service.go`.
- React components prefer inline prop object typing for small components and exported interfaces for reusable primitives, compare `MediaRecorderCard` in `frontend/components/operator/media-recorder-card.tsx` with `ButtonProps` in `frontend/components/ui/button.tsx`.

**Return Values:**
- Go functions return `(value, error)` consistently.
- Go handlers write directly to `http.ResponseWriter`; they do not return errors.
- Frontend API functions return typed promises from `request<T>()` in `frontend/lib/api/client.ts`.
- React hooks return TanStack Query result objects directly, as in `frontend/hooks/use-current-user.ts`.

## Module Design

**Exports:**
- Backend packages expose a small surface centered on `Service`, constructor functions, and DTOs/interfaces: `core-backend/internal/auth/service.go`, `core-backend/internal/specialists/service.go`.
- Frontend utility modules export named functions and types instead of default exports: `frontend/lib/auth.ts`, `frontend/lib/utils.ts`, `frontend/lib/api/client.ts`.
- App Router pages and layouts use default exports, following Next.js conventions in `frontend/app/(auth)/login/page.tsx` and `frontend/app/(app)/admin/layout.tsx`.

**Barrel Files:**
- Not detected. Imports point directly to concrete modules such as `frontend/components/ui/button.tsx` and `frontend/lib/api/types.ts`.

## Practical Guidance

- Keep frontend contract types in `frontend/lib/api/types.ts` using snake_case field names that match backend JSON, for example `is_active`, `created_at`, `questionnaire_id`.
- Add new frontend API calls to `frontend/lib/api/client.ts` rather than issuing raw `fetch` from pages or components.
- Add backend validation in service packages first, then map those errors once in `core-backend/internal/http/helpers.go`.
- For new Go handlers, follow the existing sequence: parse params/body, call service, `mapDomainError`, `writeJSON` or `writeError`.
- For new frontend data fetching, prefer `useQuery`/`useMutation` with stable array `queryKey`s, matching `frontend/hooks/use-current-user.ts` and `frontend/app/(app)/operator/examinations/new/page.tsx`.

---

*Convention analysis: 2026-03-20*
