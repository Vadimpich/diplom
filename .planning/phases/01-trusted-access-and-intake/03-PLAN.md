---
phase: 01-trusted-access-and-intake
plan: 03
title: Frontend BFF Cookie Transport
type: execute
wave: 2
depends_on:
  - 01-PLAN.md
files_modified:
  - /home/vadim/diplom/frontend/app/api/auth/login/route.ts
  - /home/vadim/diplom/frontend/app/api/auth/refresh/route.ts
  - /home/vadim/diplom/frontend/app/api/auth/logout/route.ts
  - /home/vadim/diplom/frontend/app/api/auth/session/route.ts
  - /home/vadim/diplom/frontend/lib/api/client.ts
  - /home/vadim/diplom/frontend/lib/auth.ts
  - /home/vadim/diplom/frontend/lib/constants.ts
  - /home/vadim/diplom/docs/01_contract.md
  - /home/vadim/diplom/docs/02_implementation.md
autonomous: true
requirements_addressed:
  - ACCS-02
  - ACCS-03
must_haves:
  truths:
    - "Frontend no longer depends on JavaScript-readable auth cookies for secure session state."
    - "The Next.js BFF is the single browser cookie authority for refresh-session transport."
    - "BFF routes proxy login, refresh, logout, and session bootstrap to core-backend."
  artifacts:
    - path: /home/vadim/diplom/frontend/app/api/auth/login/route.ts
      provides: frontend BFF login bridge
    - path: /home/vadim/diplom/frontend/app/api/auth/session/route.ts
      provides: frontend session bootstrap route
    - path: /home/vadim/diplom/frontend/lib/api/client.ts
      provides: BFF-backed auth client transport
  key_links:
    - from: /home/vadim/diplom/frontend/lib/api/client.ts
      to: /home/vadim/diplom/frontend/app/api/auth/session/route.ts
      via: BFF session bootstrap fetch
---

# Objective

Establish the Next.js BFF as the only browser cookie authority so auth transport is coherent before broader client-session UX migration.

<tasks>

<task id="1-03-01" type="auto">
  <name>Task 1: Add Next.js auth BFF routes and browser cookie transport</name>
  <files>
    /home/vadim/diplom/frontend/app/api/auth/login/route.ts
    /home/vadim/diplom/frontend/app/api/auth/refresh/route.ts
    /home/vadim/diplom/frontend/app/api/auth/logout/route.ts
    /home/vadim/diplom/frontend/app/api/auth/session/route.ts
    /home/vadim/diplom/frontend/lib/api/client.ts
    /home/vadim/diplom/frontend/lib/auth.ts
    /home/vadim/diplom/frontend/lib/constants.ts
  </files>
  <read_first>
    /home/vadim/diplom/AGENTS.md
    /home/vadim/diplom/.planning/phases/01-trusted-access-and-intake/01-RESEARCH.md
    /home/vadim/diplom/frontend/lib/auth.ts
    /home/vadim/diplom/frontend/lib/api/client.ts
  </read_first>
  <action>
    Implement Next.js route handlers `app/api/auth/login/route.ts`, `refresh/route.ts`, `logout/route.ts`, and `session/route.ts` that proxy to the backend auth contract and are the only layer that sets or clears browser `HttpOnly` refresh cookies. Update `frontend/lib/api/client.ts`, `frontend/lib/auth.ts`, and `frontend/lib/constants.ts` so login, refresh, logout, and session bootstrap call frontend `/api/auth/*` endpoints instead of writing or reading JavaScript-readable auth cookies. Keep this plan scoped to transport and BFF ownership only; protected-layout and route-guard UX migration moves to a separate follow-up plan.
  </action>
  <verify>
    <automated>cd /home/vadim/diplom/frontend && npm run lint && npx tsc --noEmit</automated>
  </verify>
  <acceptance_criteria>
    `rg -n "app/api/auth/login/route.ts|app/api/auth/refresh/route.ts|app/api/auth/logout/route.ts|app/api/auth/session/route.ts" /home/vadim/diplom/.planning/phases/01-trusted-access-and-intake/03-PLAN.md`
  </acceptance_criteria>
  <done>
    BFF routes own browser cookie transport, auth helpers use `/api/auth/*`, and no browser auth flow writes refresh state through `document.cookie`.
  </done>
</task>

<task id="1-03-02" type="auto">
  <name>Task 2: Document frontend auth migration</name>
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
    Update `docs/01_contract.md` to document the frontend BFF boundary as the single browser cookie authority that fronts `/auth/login`, `/auth/refresh`, and `/auth/logout`. Append the migration note to `docs/02_implementation.md`. Treat the documentation duties required by `AGENTS.md` as part of completion.
  </action>
  <verify>
    <automated>cd /home/vadim/diplom/frontend && npm run lint && npx tsc --noEmit</automated>
  </verify>
  <acceptance_criteria>
    `rg -n "BFF|/api/auth/login|/api/auth/refresh|/api/auth/logout|/api/auth/session" /home/vadim/diplom/docs/01_contract.md`
  </acceptance_criteria>
  <done>
    Frontend auth migration is recorded in `docs/01_contract.md` and `docs/02_implementation.md` per `AGENTS.md`.
  </done>
</task>

</tasks>
