# Testing Patterns

**Analysis Date:** 2026-03-20

## Test Framework

**Runner:**
- Go standard library `testing` package.
- Config: Not applicable. No `go test` wrapper config, `Makefile`, `gotestsum`, or third-party test framework config is present in `core-backend/`.

**Assertion Library:**
- Standard `testing` assertions with `t.Fatal` and `t.Fatalf`, as shown in `core-backend/internal/http/cors_test.go`.

**Run Commands:**
```bash
cd /home/vadim/diplom/core-backend && go test ./...     # Run all backend tests
cd /home/vadim/diplom/frontend && npm run lint          # Current frontend static check
cd /home/vadim/diplom/frontend && npx tsc --noEmit      # Type-check frontend
```

## Test File Organization

**Location:**
- Backend tests are co-located with the package under test. Current example: `core-backend/internal/http/cors_test.go`.
- No project-owned frontend test files are present under `frontend/`; only application source files and generated/dependency files exist.

**Naming:**
- Backend tests use Go’s `_test.go` suffix, for example `core-backend/internal/http/cors_test.go`.
- No `.test.ts`, `.spec.ts`, or `__tests__/` directories are present outside dependencies.

**Structure:**
```text
core-backend/internal/<package>/<name>_test.go
```

## Test Structure

**Suite Organization:**
```go
func TestCORSMiddlewarePreflight(t *testing.T) {
	middleware := CORSMiddleware(CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
	})

	called := false
	handler := middleware(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		called = true
		w.WriteHeader(nethttp.StatusOK)
	}))

	req := httptest.NewRequest(nethttp.MethodOptions, "/auth/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if called {
		t.Fatal("expected preflight request to be handled before next handler")
	}
	if rec.Code != nethttp.StatusNoContent {
		t.Fatalf("expected status %d, got %d", nethttp.StatusNoContent, rec.Code)
	}
}
```

**Patterns:**
- Tests are function-based, not suite-based, with one behavior per `Test...` function in `core-backend/internal/http/cors_test.go`.
- Setup is done inline inside each test instead of shared fixtures or helper packages.
- HTTP behavior is tested through `net/http/httptest` request and recorder objects.
- Assertions are direct conditional checks with explicit failure messages.

## Mocking

**Framework:** Standard library only.

**Patterns:**
```go
called := false
handler := middleware(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
	called = true
	w.WriteHeader(nethttp.StatusOK)
}))
```

**What to Mock:**
- Use lightweight in-memory closures for HTTP middleware and handler boundaries, as in `core-backend/internal/http/cors_test.go`.
- Prefer standard-library test doubles like `httptest.NewRequest` and `httptest.NewRecorder` when verifying HTTP concerns.

**What NOT to Mock:**
- Do not invent custom mocking frameworks unless the repository adopts one.
- Do not add frontend mocks until a real frontend test runner exists in `frontend/package.json` or dedicated config files.

## Fixtures and Factories

**Test Data:**
```go
req := httptest.NewRequest(nethttp.MethodOptions, "/auth/login", nil)
req.Header.Set("Origin", "http://localhost:3000")
req.Header.Set("Access-Control-Request-Method", nethttp.MethodPost)
req.Header.Set("Access-Control-Request-Headers", "content-type")
```

**Location:**
- No shared fixture or factory directories are present.
- Test data is currently embedded inline in `core-backend/internal/http/cors_test.go`.

## Coverage

**Requirements:** None enforced.

**View Coverage:**
```bash
cd /home/vadim/diplom/core-backend && go test ./... -cover
```

## Test Types

**Unit Tests:**
- Present only in a very limited form for backend HTTP middleware and router behavior in `core-backend/internal/http/cors_test.go`.
- Current style isolates a small behavior slice without spinning up external infrastructure.

**Integration Tests:**
- Not detected.
- Database, S3, auth flow, and frontend-to-backend integration are currently untested at repository level despite concrete code in `core-backend/internal/app/app.go`, `core-backend/internal/auth/service.go`, and `frontend/lib/api/client.ts`.

**E2E Tests:**
- Not used.
- No Playwright, Cypress, or browser automation config is present in `frontend/` or the repository root.

## Common Patterns

**Async Testing:**
```go
rec := httptest.NewRecorder()
router.ServeHTTP(rec, req)
if rec.Code != nethttp.StatusNoContent {
	t.Fatalf("expected status %d, got %d", nethttp.StatusNoContent, rec.Code)
}
```
- Current backend tests do not use goroutines, channels, or explicit async waiting; request handling is exercised synchronously through the handler interface.

**Error Testing:**
```go
if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
	t.Fatalf("expected allow-origin header to match request origin, got %q", got)
}
```
- Assertions verify exact status codes and header values instead of only checking non-nil errors.
- Failure messages are specific and include expected/actual values, matching `core-backend/internal/http/cors_test.go`.

## Static Verification Pattern

- Frontend quality verification currently depends on static checks, not executable tests.
- `frontend/package.json` exposes only `npm run lint`; no `test` or `coverage` script exists.
- `docs/02_implementation.md` records that frontend was validated with `eslint`, `tsc --noEmit`, and `next build`, which reflects current practice for changes under `frontend/`.
- When planning new work, treat `npm run lint`, `npx tsc --noEmit`, and `next build` in `frontend/` plus `go test ./...` in `core-backend/` as the minimum verification baseline.

## Gaps That Affect Planning

- There is exactly one first-party test file: `core-backend/internal/http/cors_test.go`.
- No test harness exists for `core-backend/internal/auth/service.go`, `core-backend/internal/specialists/service.go`, `core-backend/internal/http/*_handler.go`, or any SQL/storage integration path.
- No frontend component, hook, or route tests exist for files such as `frontend/app/(auth)/login/page.tsx`, `frontend/components/layout/route-guard.tsx`, or `frontend/components/operator/media-recorder-card.tsx`.
- New phases that touch auth, forms, or workflow transitions should budget time to add the missing tests rather than assuming coverage already exists.

---

*Testing analysis: 2026-03-20*
