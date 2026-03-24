# Phase 10: Admin Control Surfaces - Research

**Researched:** 2026-03-24
**Domain:** Admin settings, runtime monitoring, and audit visibility
**Confidence:** HIGH

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| STNG-01 | Администратор может изменять через UI политики TTL хранения аудио и параметры retry/processing, разрешённые текущей архитектурой | Требуется новый admin settings contract + minimal backend persistence/source-of-truth; текущий код этого не даёт |
| STNG-02 | Экран системных настроек показывает текущие значения, валидирует ввод и даёт явный success/error feedback после сохранения | Реализуется через typed GET/PUT settings API и form validation на frontend |
| MONR-01 | Администратор может просматривать в UI базовый статус системы по health/readiness и ключевым техническим индикаторам без прямого доступа к инфраструктуре | Основные runtime endpoints уже есть: `GET /health`, `GET /ready`, `GET /metrics`; frontend уже имеет BFF probes |
| AUDT-01 | Администратор может просматривать и фильтровать audit log по типу события, периоду и связанным доменным объектам | Audit storage и list-query уже существуют в backend internals, но отсутствует admin HTTP surface и UI |
</phase_requirements>

## Summary

Phase 10 должен оставаться frontend-first только для monitoring. Для `MONR-01` уже есть почти весь runtime plumbing: core backend публикует `GET /health`, `GET /ready`, `GET /metrics`, а frontend уже проксирует readiness и dependency gauge через `/api/ready` и `/api/metrics`. Текущий admin monitoring screen использует только `GET /health`, поэтому Phase 10 здесь в основном про UI composition и безопасное использование уже существующих endpoints.

Для settings и audit текущий код недостаточен. Settings page остаётся явным placeholder, а backend config живёт в env/process memory без admin mutation API. Audit subsystem уже пишет и умеет читать события из PostgreSQL, но router не публикует admin endpoint для списка audit events, а имеющийся filter model не покрывает все нужные доменные object filters. Значит минимальные backend additions нужны ровно в двух местах: typed system settings API и admin audit-read API.

**Primary recommendation:** Разбить Phase 10 на 4 плана: monitoring UI на существующих endpoints, audit read API, audit UI, system settings contract+backend+UI.

## Project Constraints

### From AGENTS.md / project docs
- `docs/00_project.md` и `docs/01_contract.md` обязательны как sources of truth.
- `docs/00_project.md` нельзя редактировать агентом.
- Любое изменение HTTP/API или payload contract требует обновления `docs/01_contract.md`.
- После реализации задачи должна добавляться краткая запись в `docs/02_implementation.md`.
- Менять только то, что требуется задачей; не делать попутный refactor.
- Критичная логика должна иметь логирование, error handling и понятное fallback-поведение.
- Не хардкодить секреты и не логировать чувствительные данные.
- Audit и чувствительные admin actions должны оставаться согласованными с documented workflow.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard Here |
|---------|---------|---------|-------------------|
| Next.js | 15.5.14 | Admin UI and BFF routes | Уже используется во всём frontend контуре |
| React | 19.2.4 | UI runtime | Базовый runtime проекта |
| @tanstack/react-query | 5.91.2 | Query/polling state | Уже используется на admin страницах |
| React Hook Form | 7.71.2 | Settings/audit filter forms | Подходит для validated admin forms |
| Zod | 3.25.76 | Client/server form schemas | Уже выбранный validation layer |
| chi | 5.2.3 | Go HTTP router | Текущий backend router |
| pgx/sqlc | pgx 5.7.4 / sqlc-generated repo | PostgreSQL access | Audit storage и остальная data layer уже на этом стеке |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| sonner | 2.0.7 | Success/error feedback | Для save feedback на settings |
| lucide-react | 0.511.0 | Admin status icons | Для monitoring/audit presentation |

## Existing Support vs Gaps

### Already Available
| Surface | Evidence | Reuse Plan | Confidence |
|--------|----------|------------|------------|
| Core liveness | `GET /health` in [health.go](/home/vadim/diplom/core-backend/internal/http/health.go#L33) and routed in [router.go](/home/vadim/diplom/core-backend/internal/http/router.go#L77) | Reuse as lightweight heartbeat card | HIGH |
| Core readiness | `GET /ready` checks `postgres`, `rabbitmq`, `minio`, `baseline`, `kesmi` in [app.go](/home/vadim/diplom/core-backend/internal/app/app.go#L118) and [health.go](/home/vadim/diplom/core-backend/internal/http/health.go#L44) | Reuse for dependency matrix | HIGH |
| Metrics exposition | `GET /metrics` in [health.go](/home/vadim/diplom/core-backend/internal/http/health.go#L79) | Frontend can derive a few key indicators without infra access | HIGH |
| Frontend BFF readiness probe | [core-readiness.ts](/home/vadim/diplom/frontend/lib/server/core-readiness.ts#L15), `/api/ready`, `/api/metrics` | Use instead of browser-to-core direct probing | HIGH |
| Audit persistence | `audit_logs` table and SQL repo list path in [000009_audit_observability.up.sql](/home/vadim/diplom/core-backend/migrations/000009_audit_observability.up.sql#L1) and [repository.go](/home/vadim/diplom/core-backend/internal/audit/repository.go#L68) | Build admin read API on top of existing package | HIGH |

### Missing or Incomplete
| Surface | Current State | Minimal Addition Required | Confidence |
|--------|---------------|---------------------------|------------|
| Monitoring page | Current page only polls `apiClient.health` in [monitoring/page.tsx](/home/vadim/diplom/frontend/app/(app)/admin/monitoring/page.tsx#L10) | Expand UI to show readiness + selected technical indicators | HIGH |
| Settings UI | Explicit placeholder in [settings/page.tsx](/home/vadim/diplom/frontend/app/(app)/admin/settings/page.tsx#L15) | Real form, validation, load/save feedback | HIGH |
| Settings backend | No router/admin endpoint for settings in [router.go](/home/vadim/diplom/core-backend/internal/http/router.go#L89) | Add admin `GET/PUT` settings API | HIGH |
| Settings source of truth | Current values live in env-loaded config only in [config.go](/home/vadim/diplom/core-backend/internal/config/config.go#L43) | Add persisted settings row/table or equivalent core-owned store | HIGH |
| Audit API | Audit package can list, but no HTTP handler/route exposes it | Add admin read-only endpoint | HIGH |
| Audit filters | Current `ListFilter` supports event type, period, actor user, examination, resource kind, but not `resource_id`, `specialist_id`, `questionnaire_id` | Extend filter DTO/query for linked object filtering | HIGH |
| Audit UI | No `/admin/audit` route exists | Add dedicated screen and nav entry | HIGH |

## Recommended Phase Decomposition

### Plan 10-01: Monitoring Surface on Existing Runtime Endpoints
**Scope**
- Expand `/admin/monitoring` to show:
  - core liveness from `GET /health`
  - readiness status and dependency map from `/api/ready`
  - a compact “key indicators” block from `/api/metrics`
- Keep browser traffic on frontend BFF routes, not direct core calls.

**Why first**
- Zero or near-zero backend change.
- Immediately closes most of `MONR-01`.

### Plan 10-02: Admin Audit Read API
**Scope**
- Add admin-only `GET /audit/events` backend route.
- Expose stable DTO based on existing `audit.Event`.
- Support filters required by roadmap:
  - `event_type`
  - `from`, `to`
  - linked object filters via `resource_kind` + `resource_id`
  - likely keep `examination_id` as convenience shortcut
  - `limit`
- Reuse existing audit repository instead of inventing a second read model.

**Why separate**
- Small, isolated backend task.
- Unblocks audit UI without coupling to settings work.

### Plan 10-03: Audit UI and Admin Navigation
**Scope**
- Add `/admin/audit` screen with filter form and table/list.
- Show event type, outcome, actor, happened_at, resource, domain refs, correlation/request IDs where useful.
- Add explicit loading/empty/error states and low-noise defaults.
- Add nav/index entry without overloading Phase 9 admin landing page.

**Why separate**
- Frontend-heavy work after API exists.
- Keeps audit usability and state handling reviewable on its own.

### Plan 10-04: System Settings Contract, Backend, and UI
**Scope**
- Introduce typed admin settings contract for the subset truly mutable in current architecture.
- Add backend persistence/source-of-truth plus admin `GET/PUT`.
- Implement settings screen with RHF+Zod validation and toast/inline feedback.
- Emit audit event for settings mutation.

**Recommended editable subset**
- `audio_retention_ttl_days`
- `processing_outbox_max_attempts`
- `kesmi_max_retries`

**Deliberately not in MVP unless user expands scope**
- raw env/network endpoints
- secrets
- transport timeouts / poll intervals that imply hot-reload or infrastructure-level reconfiguration

## Architecture Patterns

### Pattern 1: Frontend BFF for Runtime Probes
**What:** Browser reads `/api/ready` and `/api/metrics`; frontend server route talks to core.
**Why:** Keeps internal base URL and runtime topology out of the browser and matches current project pattern.
**Use for:** monitoring cards and dependency indicators.

### Pattern 2: Thin Admin HTTP Layer over Existing Domain Service
**What:** Publish read-only audit endpoint via chi handler that delegates to `internal/audit.Service`.
**Why:** Audit storage logic already exists; Phase 10 should not create parallel repositories.

### Pattern 3: Persisted Settings, Not Process Env Mutation
**What:** Admin-edited settings live in PostgreSQL `system_settings` source-of-truth and are read by backend services.
**Why:** Current config is env-loaded once at process start; mutating env from UI is not a real runtime control surface.

### Anti-Patterns to Avoid
- **Editing `.env` semantics from UI:** This creates fake control surfaces because current process config is loaded at startup.
- **Direct browser access to core `/ready` or `/metrics`:** Use existing frontend BFF routes and keep internal topology private.
- **New custom audit storage/view model:** Existing `audit_logs` table and service already cover the needed domain.
- **Trying to expose full Prometheus text in the admin UI:** Render selected indicators instead of dumping raw exposition.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Audit persistence | Separate admin audit table/view store | Existing `audit_logs` + `internal/audit` | Already append-only, indexed, and queryable |
| Monitoring transport | Client-side direct infra probing | Existing `/api/ready` and `/api/metrics` BFF routes | Safer topology boundary |
| Form validation | Ad hoc inline checks | RHF + Zod | Already standard in frontend |
| Runtime metrics UI | Full Prometheus browser parser | Server-side selected metric extraction or compact proxy DTO | Simpler, lower risk, easier UX |

## Common Pitfalls

### Pitfall 1: Treating env config as admin-editable runtime state
**What goes wrong:** UI “saves” values that do not affect the running system.
**Why it happens:** Current config is loaded once from environment on startup.
**How to avoid:** Persist only the subset the backend explicitly reads from a mutable store.
**Warning signs:** PUT succeeds but worker/core behaviour never changes.

### Pitfall 2: Audit filter is too narrow for “linked domain object”
**What goes wrong:** UI can filter only by `examination_id`, which fails the broader requirement.
**Why it happens:** Existing repository filter lacks `resource_id`, `specialist_id`, `questionnaire_id`.
**How to avoid:** Expand filter contract deliberately before building the page.
**Warning signs:** User cannot find questionnaire/user-related events without manual scanning.

### Pitfall 3: Monitoring UI leaks raw internal telemetry
**What goes wrong:** Admin page becomes a raw text dump or exposes high-cardinality identifiers.
**Why it happens:** Reusing `/metrics` blindly.
**How to avoid:** Show curated indicators and preserve low-cardinality presentation.
**Warning signs:** Page contains unbounded labels or unreadable metric blocks.

## Code Examples

### Existing readiness proxy pattern
```ts
// Source: frontend/lib/server/core-readiness.ts
const response = await fetch(coreUrl("/ready"), { cache: "no-store" });
if (!response.ok) {
  return { dependency: "down", serviceStatus: "degraded", statusCode: 503, dependencyGauge: 0 };
}
return { dependency: "up", serviceStatus: "ready", statusCode: 200, dependencyGauge: 1 };
```

### Existing audit list backend pattern
```go
// Source: core-backend/internal/audit/repository.go
rows, err := r.queries.ListAuditLogs(ctx, sqlcdb.ListAuditLogsParams{
    EventType:     textPtrArg(filter.EventType),
    ResourceKind:  textPtrArg(filter.ResourceKind),
    ActorUserID:   int8PtrArg(filter.ActorUserID),
    ExaminationID: int8PtrArg(filter.ExaminationID),
    FromAt:        pgtype.Timestamptz{Time: fromAt.UTC(), Valid: true},
    ToAt:          pgtype.Timestamptz{Time: toAt.UTC(), Valid: true},
    LimitCount:    limit,
})
```

## State of the Art

| Old Approach | Current Recommended Approach | Impact |
|--------------|------------------------------|--------|
| Placeholder admin settings | Typed admin settings API backed by mutable core-owned storage | Real `STNG-01`/`STNG-02` instead of stub |
| Health-only monitoring page | Health + readiness + curated indicators | Real `MONR-01` without infra access |
| Internal-only audit package | Admin read API + dedicated UI | Real `AUDT-01` |

## Open Questions

1. **Which settings are genuinely mutable at runtime without process restart?**
   - What we know: `KESMI_MAX_RETRIES`, `PROCESSING_OUTBOX_MAX_ATTEMPTS`, and future audio TTL policy are the clearest candidates.
   - What's unclear: whether poll interval/timeouts should be included in scope.
   - Recommendation: keep Phase 10 to the three-field subset above unless product explicitly asks for live reload semantics.

2. **Does “linked domain object” need only examination filters or broader resource targeting?**
   - What we know: roadmap wording is broader than current repository filter.
   - What's unclear: whether specialist/questionnaire/user lookup UX is required in one pass.
   - Recommendation: support `resource_kind` + `resource_id` in the API so the planner does not paint itself into an examination-only corner.

## Environment Availability

Step 2.6: SKIPPED (phase is code/config/UI work; no new external tool dependency beyond the existing project stack was identified)

## Sources

### Primary (HIGH confidence)
- [docs/00_project.md](/home/vadim/diplom/docs/00_project.md) - admin settings, audit, storage expectations
- [docs/01_contract.md](/home/vadim/diplom/docs/01_contract.md) - runtime health/readiness/metrics and audit contract
- [docs/02_implementation.md](/home/vadim/diplom/docs/02_implementation.md) - current shipped status for placeholders and prior audit/runtime work
- [ROADMAP.md](/home/vadim/diplom/.planning/ROADMAP.md) - Phase 10 goal and success criteria
- [REQUIREMENTS.md](/home/vadim/diplom/.planning/REQUIREMENTS.md) - requirement IDs in scope
- [monitoring/page.tsx](/home/vadim/diplom/frontend/app/(app)/admin/monitoring/page.tsx) - current monitoring surface
- [settings/page.tsx](/home/vadim/diplom/frontend/app/(app)/admin/settings/page.tsx) - current settings placeholder
- [core-readiness.ts](/home/vadim/diplom/frontend/lib/server/core-readiness.ts) - existing readiness probe
- [router.go](/home/vadim/diplom/core-backend/internal/http/router.go) - published HTTP surfaces
- [health.go](/home/vadim/diplom/core-backend/internal/http/health.go) - health/ready/metrics handlers
- [config.go](/home/vadim/diplom/core-backend/internal/config/config.go) - current env-based settings model
- [app.go](/home/vadim/diplom/core-backend/internal/app/app.go) - actual readiness dependencies wired into runtime
- [repository.go](/home/vadim/diplom/core-backend/internal/audit/repository.go) - existing audit read path

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - versions taken from repository manifests
- Architecture: HIGH - conclusions derived from current router, handlers, and project docs
- Pitfalls: HIGH - directly grounded in current placeholder/env-only/audit-query gaps

**Research date:** 2026-03-24
**Valid until:** 2026-04-23
