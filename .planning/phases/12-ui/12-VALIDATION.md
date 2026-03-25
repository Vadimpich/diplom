# Phase 12 Validation

## Purpose

Канонический regression-набор для shipped Phase 12 UI pass. Этот файл фиксирует:

- какие планы закрыты в рамках Phase 12;
- какие пользовательские истины считаются shipped;
- какие автоматические команды нужно прогонять повторно для frontend и для backend-поддержки из `12-02`;
- какие runtime/manual проверки остаются вне автоматического покрытия.

## Completed slices

- `12-01` — компактный login и плотный contour chrome без developer-facing copy.
- `12-02` — bounded backend/contract support для плотных user/questionnaire/specialist registries.
- `12-03` — workload-first operator dashboard, specialist registry и examinations journal.
- `12-04` — компактный admin home, плотные registries и честный monitoring screen.
- `12-05` — плотные operator detail/intake/processing/results surfaces.
- `12-06` — компактные admin edit surfaces для users, questionnaires и settings.

## Shipped user-visible truths

- `/login`, `/operator/*` и `/admin/*` читаются как плотные рабочие поверхности, а не презентационные экраны.
- Операторские overview/detail/intake/results surfaces используют только контрактные данные и не показывают raw debug wording в основном UI.
- Административные home/registry/editor surfaces используют реальные summary и metadata из существующих запросов и bounded DTO enrichments `12-02`.
- Реестры пользователей, опросников и специалистов остаются на существующих endpoints; Phase 12 не добавляла новые summary APIs.
- Backend support из `12-02` ограничен UI-critical metadata: `last_login_at`, questionnaire usage/editor facts и derived specialist summary.

## Automated regression commands

Команды выполняются последовательно из корня репозитория.

### Backend contract support from 12-02

```bash
cd /home/vadim/diplom/core-backend && go test ./internal/http ./internal/auth ./internal/questionnaires ./internal/specialists
```

Покрывает:
- `last_login_at` и auth-adjacent registry support;
- questionnaire registry metadata и repository mapping;
- specialist summary read-model wiring;
- HTTP handler regression вокруг enriched DTO paths.

### Frontend regression

```bash
cd /home/vadim/diplom/frontend && npm run lint
cd /home/vadim/diplom/frontend && npm run build
cd /home/vadim/diplom/frontend && npx tsc --noEmit
```

Порядок важен:
- `next build` запускается перед `tsc`, потому что в текущем workspace часть generated `.next/types` появляется только после build.

### Plan-level artifact checks

```bash
rg -n "Phase 12|12-" /home/vadim/diplom/docs/02_implementation.md
test -f /home/vadim/diplom/.planning/phases/12-ui/12-VALIDATION.md
```

## Current execution record

- `2026-03-25` backend: `go test ./internal/http ./internal/auth ./internal/questionnaires ./internal/specialists` — passed (`ok`, cached).
- `2026-03-25` frontend lint: `npm run lint` — passed.
- `2026-03-25` frontend build: `npm run build` — passed; production build completed for operator/admin/app/api routes.
- `2026-03-25` frontend typecheck: `npx tsc --noEmit` — passed after the build-generated `.next/types` refresh.
- `2026-03-25` implementation log grep: `rg -n "Phase 12|12-" docs/02_implementation.md` — passed; `12-01` through `12-06` entries are present.
- `2026-03-25` validation artifact existence: `test -f .planning/phases/12-ui/12-VALIDATION.md` — passed.

## Runtime and manual checks not covered here

- Browser-side visual review of `/login`, `/operator`, `/operator/specialists`, `/operator/history`, `/admin`, `/admin/users`, `/admin/questionnaires`, `/admin/monitoring`, `/admin/settings`.
- Full-stack compose smoke with live auth, backend, RabbitMQ, MinIO, ML services and WiMi.
- Manual UX confirmation of unsaved-change warning, toast flows, and dense registry readability on real data volumes.
