---
phase: 01-trusted-access-and-intake
plan: 08
title: Protected Layout Session UX Migration
type: execute
wave: 3
depends_on:
  - 03-PLAN.md
files_modified:
  - /home/katya/dimplom/frontend/app/(auth)/login/page.tsx
  - /home/katya/dimplom/frontend/app/(app)/admin/layout.tsx
  - /home/katya/dimplom/frontend/app/(app)/operator/layout.tsx
  - /home/katya/dimplom/frontend/app/page.tsx
  - /home/katya/dimplom/frontend/components/layout/route-guard.tsx
  - /home/katya/dimplom/frontend/components/layout/app-shell.tsx
  - /home/katya/dimplom/frontend/hooks/use-current-user.ts
  - /home/katya/dimplom/docs/01_contract.md
  - /home/katya/dimplom/docs/02_implementation.md
autonomous: true
requirements_addressed:
  - ACCS-02
  - ACCS-03
must_haves:
  truths:
    - "Protected layouts bootstrap session through the BFF session endpoint rather than stale role cookies."
    - "Login and logout redirects are consistent across admin and operator shells."
  artifacts:
    - path: /home/katya/dimplom/frontend/hooks/use-current-user.ts
      provides: client session bootstrap
    - path: /home/katya/dimplom/frontend/components/layout/route-guard.tsx
      provides: UX-only route protection aligned with BFF session state
    - path: /home/katya/dimplom/frontend/app/(app)/admin/layout.tsx
      provides: admin shell session gate
    - path: /home/katya/dimplom/frontend/app/(app)/operator/layout.tsx
      provides: operator shell session gate
  key_links:
    - from: /home/katya/dimplom/frontend/hooks/use-current-user.ts
      to: /home/katya/dimplom/frontend/app/api/auth/session/route.ts
      via: session bootstrap fetch
    - from: /home/katya/dimplom/frontend/components/layout/route-guard.tsx
      to: /home/katya/dimplom/frontend/hooks/use-current-user.ts
      via: protected route session check
---

# Objective

Finish the brownfield client migration so protected pages, login redirects, and logout UX all follow BFF-backed session state.

<tasks>

<task id="1-08-01" type="auto">
  <name>Task 1: Rewire protected layouts and route guard to BFF session state</name>
  <files>
    /home/katya/dimplom/frontend/app/(auth)/login/page.tsx
    /home/katya/dimplom/frontend/app/(app)/admin/layout.tsx
    /home/katya/dimplom/frontend/app/(app)/operator/layout.tsx
    /home/katya/dimplom/frontend/app/page.tsx
    /home/katya/dimplom/frontend/components/layout/route-guard.tsx
    /home/katya/dimplom/frontend/components/layout/app-shell.tsx
    /home/katya/dimplom/frontend/hooks/use-current-user.ts
  </files>
  <read_first>
    /home/katya/dimplom/AGENTS.md
    /home/katya/dimplom/.planning/phases/01-trusted-access-and-intake/03-PLAN.md
    /home/katya/dimplom/frontend/app/(auth)/login/page.tsx
    /home/katya/dimplom/frontend/app/(app)/admin/layout.tsx
    /home/katya/dimplom/frontend/app/(app)/operator/layout.tsx
    /home/katya/dimplom/frontend/components/layout/route-guard.tsx
    /home/katya/dimplom/frontend/hooks/use-current-user.ts
  </read_first>
  <action>
    Update the brownfield login page, root redirect, admin/operator layouts, `route-guard.tsx`, `app-shell.tsx`, and `use-current-user.ts` so protected-route and redirect behavior bootstrap from `/api/auth/session` and react to logout through the BFF routes introduced in Plan 03. Remove reliance on stale role-cookie checks as the authoritative auth source; any client role cache becomes UX-only and derived from the BFF session response.
  </action>
  <verify>
    <automated>cd /home/katya/dimplom/frontend && npm run lint && npx tsc --noEmit</automated>
  </verify>
  <acceptance_criteria>
    `rg -n "/api/auth/session|useCurrentUser|route-guard|/login" /home/katya/dimplom/frontend/app /home/katya/dimplom/frontend/components /home/katya/dimplom/frontend/hooks`
  </acceptance_criteria>
  <done>
    Protected layouts and guards bootstrap from the BFF session endpoint, login redirects are based on session state, and logout UX no longer depends on stale browser cookie assumptions.
  </done>
</task>

<task id="1-08-02" type="auto">
  <name>Task 2: Document protected-layout session migration</name>
  <files>
    /home/katya/dimplom/docs/01_contract.md
    /home/katya/dimplom/docs/02_implementation.md
  </files>
  <read_first>
    /home/katya/dimplom/AGENTS.md
    /home/katya/dimplom/docs/01_contract.md
    /home/katya/dimplom/docs/02_implementation.md
  </read_first>
  <action>
    Update `docs/01_contract.md` to document protected-layout/session-bootstrap expectations around `/api/auth/session`, and append the migration note to `docs/02_implementation.md`. Treat the documentation duties required by `AGENTS.md` as part of completion.
  </action>
  <verify>
    <automated>cd /home/katya/dimplom/frontend && npm run lint && npx tsc --noEmit</automated>
  </verify>
  <acceptance_criteria>
    `rg -n "/api/auth/session|protected routes|session bootstrap" /home/katya/dimplom/docs/01_contract.md`
  </acceptance_criteria>
  <done>
    Session-bootstrap UX semantics are recorded in `docs/01_contract.md` and `docs/02_implementation.md` per `AGENTS.md`.
  </done>
</task>

</tasks>
