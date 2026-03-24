# Technology Stack

**Analysis Date:** 2026-03-20

## Languages

**Primary:**
- Go 1.24 - `core-backend/` backend runtime declared in `core-backend/go.mod` and built in `core-backend/Dockerfile`
- TypeScript 5.9 - `frontend/` web application declared in `frontend/package.json`

**Secondary:**
- SQL (PostgreSQL dialect) - schema and query definitions in `core-backend/migrations/*.sql` and `core-backend/db/queries/*.sql`
- CSS - Tailwind-driven styles in `frontend/app/globals.css` and `frontend/tailwind.config.ts`
- YAML - local infrastructure and code generation config in `docker-compose.yml` and `core-backend/db/sqlc.yaml`

## Runtime

**Environment:**
- Go runtime on Linux container from `golang:1.24` build stage and `alpine:3.20` runtime in `core-backend/Dockerfile`
- Node.js 22 on Alpine in `frontend/Dockerfile`
- Browser runtime APIs for recording in `frontend/components/operator/media-recorder-card.tsx`

**Package Manager:**
- Go modules for `core-backend/` via `core-backend/go.mod` and `core-backend/go.sum`
- npm lockfile is present and used for container builds in `frontend/package-lock.json` and `frontend/Dockerfile`
- `pnpm` lockfile is also committed in `frontend/pnpm-lock.yaml`, but current container build path does not use it
- Lockfile: present for both backend and frontend

## Frameworks

**Core:**
- Chi v5.2.3 - HTTP router and middleware in `core-backend/internal/http/router.go`
- Next.js v15.2.4 - frontend framework in `frontend/package.json` and `frontend/next.config.ts`
- React v19.2.4 - UI runtime in `frontend/package.json`

**Testing:**
- Go `testing` package - backend test coverage currently visible in `core-backend/internal/http/cors_test.go`
- No dedicated frontend test runner is detected in `frontend/package.json`

**Build/Dev:**
- sqlc config v2 with generated code using `pgx/v5` in `core-backend/db/sqlc.yaml`
- Tailwind CSS v3.4.19 with PostCSS + Autoprefixer in `frontend/tailwind.config.ts` and `frontend/postcss.config.mjs`
- ESLint 9 with Next flat config in `frontend/eslint.config.mjs`
- Docker Compose - local multi-service runtime in `docker-compose.yml`

## Key Dependencies

**Critical:**
- `github.com/jackc/pgx/v5` v5.7.4 - PostgreSQL client and pool used by `core-backend/internal/postgres/client.go`
- `github.com/go-chi/chi/v5` v5.2.3 - request routing, middleware, and URL params in `core-backend/internal/http/*.go`
- `github.com/golang-jwt/jwt/v5` v5.2.2 - access token issuance and parsing in `core-backend/internal/auth/jwt.go`
- `github.com/minio/minio-go/v7` v7.0.95 - S3-compatible object storage client in `core-backend/internal/storage/s3.go`
- `next` v15.2.4 - application shell and server rendering in `frontend/`
- `@tanstack/react-query` v5.91.2 - frontend data fetching and mutation state across pages in `frontend/app/(app)/**/*.tsx`
- `react-hook-form` v7.71.2 and `zod` v3.25.76 - form state and validation in `frontend/app/(app)/**/page.tsx`

**Infrastructure:**
- `golang.org/x/crypto` v0.39.0 - bcrypt password hashing in `core-backend/internal/auth/service.go`
- `@radix-ui/react-slot` v1.2.4 - UI composition helper in `frontend/components/ui/`
- `lucide-react` v0.511.0 - icon set used in `frontend/`
- `clsx` and `tailwind-merge` - class composition in `frontend/lib/utils.ts`

## Configuration

**Environment:**
- Backend runtime config is loaded from env in `core-backend/internal/config/config.go`
- Local example variables are documented in `.env.example`
- Frontend API base URL comes from `NEXT_PUBLIC_API_URL` in `frontend/lib/api/client.ts`
- Compose assembles `DATABASE_URL` from `POSTGRES_*` values in `docker-compose.yml`

**Build:**
- Backend container build in `core-backend/Dockerfile`
- Frontend standalone Next build in `frontend/Dockerfile`
- Frontend runtime/build config in `frontend/next.config.ts`, `frontend/tsconfig.json`, `frontend/tailwind.config.ts`, `frontend/postcss.config.mjs`, and `frontend/eslint.config.mjs`
- SQL code generation config in `core-backend/db/sqlc.yaml`

## Platform Requirements

**Development:**
- Docker and Docker Compose v2 are required for the documented local stack in `README.md`
- No repo-level `.nvmrc`, `.python-version`, `pyproject.toml`, or `requirements.txt` are detected

**Production:**
- Current deployment target is containerized self-hosted services via `docker-compose.yml`
- Frontend is built as a standalone Next server in `frontend/Dockerfile`
- Backend is a single statically linked Go HTTP service in `core-backend/Dockerfile`

---

*Stack analysis: 2026-03-20*
