---
phase: 01-trusted-access-and-intake
plan: 02
title: Backend RBAC Route Segmentation
type: execute
wave: 2
depends_on:
  - 01-PLAN.md
files_modified:
  - /home/vadim/diplom/core-backend/internal/http/auth_middleware.go
  - /home/vadim/diplom/core-backend/internal/http/router.go
  - /home/vadim/diplom/core-backend/internal/http/rbac_test.go
  - /home/vadim/diplom/docs/01_contract.md
  - /home/vadim/diplom/docs/02_implementation.md
autonomous: true
requirements_addressed:
  - ACCS-04
must_haves:
  truths:
    - "Backend rejects admin/operator routes for the wrong role with 403."
    - "Protected routes are segmented by role in the Go router."
  artifacts:
    - path: /home/vadim/diplom/core-backend/internal/http/auth_middleware.go
      provides: RequireRoles middleware
    - path: /home/vadim/diplom/core-backend/internal/http/router.go
      provides: role-scoped route groups
  key_links:
    - from: /home/vadim/diplom/core-backend/internal/http/router.go
      to: /home/vadim/diplom/core-backend/internal/http/auth_middleware.go
      via: RequireRoles middleware attachment
---

# Objective

Enforce admin/operator boundaries on the backend with centralized middleware and role-scoped route groups.

<tasks>

<task id="1-02-01" type="auto">
  <name>Task 1: Add RequireRoles middleware and split route groups</name>
  <files>
    /home/vadim/diplom/core-backend/internal/http/auth_middleware.go
    /home/vadim/diplom/core-backend/internal/http/router.go
    /home/vadim/diplom/core-backend/internal/http/rbac_test.go
  </files>
  <read_first>
    /home/vadim/diplom/AGENTS.md
    /home/vadim/diplom/docs/00_project.md
    /home/vadim/diplom/docs/01_contract.md
    /home/vadim/diplom/core-backend/internal/http/router.go
  </read_first>
  <action>
    Extend `core-backend/internal/http/auth_middleware.go` with `RequireRoles(allowed ...string)` and restructure `core-backend/internal/http/router.go` into explicit authenticated subgroups: shared routes keep `GET /me`, admin-only routes contain `/users` and `/questionnaires`, and operator/admin routes contain `/specialists`, `/examinations`, `/answers`, and `GET /specialists/{id}/examinations`. Add `core-backend/internal/http/rbac_test.go` proving wrong-role `403` and unauthenticated `401`.
  </action>
  <verify>
    <automated>cd /home/vadim/diplom/core-backend && go test ./internal/http -run 'TestRequireRoles|TestAdminRoutes|TestOperatorRoutes' -count=1</automated>
  </verify>
  <acceptance_criteria>
    `rg -n "RequireRoles\\(\"admin\"\\)|RequireRoles\\(\"operator\", \"admin\"\\)" /home/vadim/diplom/core-backend/internal/http/router.go`
  </acceptance_criteria>
  <done>
    Backend route groups are role-scoped, negative-path tests exist and pass, and no protected route depends on frontend-only role checks for authorization.
  </done>
</task>

<task id="1-02-02" type="auto">
  <name>Task 2: Document RBAC contracts and implementation state</name>
  <files>
    /home/vadim/diplom/docs/01_contract.md
    /home/vadim/diplom/docs/02_implementation.md
  </files>
  <read_first>
    /home/vadim/diplom/AGENTS.md
    /home/vadim/diplom/docs/01_contract.md
    /home/vadim/diplom/docs/02_implementation.md
  </read_first>
  <action>
    Update `docs/01_contract.md` to annotate admin-only and operator/admin route groups with required roles and `403` semantics, then append a concise implementation note to `docs/02_implementation.md` describing backend-enforced RBAC. Treat the documentation update required by `AGENTS.md` as part of completion.
  </action>
  <verify>
    <automated>cd /home/vadim/diplom/core-backend && go test ./internal/http -run 'TestRequireRoles|TestAdminRoutes|TestOperatorRoutes' -count=1</automated>
  </verify>
  <acceptance_criteria>
    `rg -n "403|admin only|operator" /home/vadim/diplom/docs/01_contract.md`
  </acceptance_criteria>
  <done>
    RBAC route requirements are documented in `docs/01_contract.md` and implementation status is logged in `docs/02_implementation.md` per `AGENTS.md`.
  </done>
</task>

</tasks>
