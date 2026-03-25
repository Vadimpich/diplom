# Мультимодальная система оценки психоэмоционального состояния специалистов

## What This Is

Self-hosted система для проведения обследований специалистов критических областей на основе записи речевых ответов и их мультимодального анализа. Продукт используется внутри организации оператором и администратором, чтобы проводить обследования, получать интерпретируемый результат и передавать агрегированный профиль состояния во внешнюю систему поддержки принятия решений о допуске к профессиональной деятельности.

По состоянию после архивации milestone `v1.1` система уже покрывает полный опорный контур из дипломного scope и имеет отдельные operator/admin интерфейсы, административные control surfaces и уплотнённый рабочий UI. При этом `v1.1` закрыт с принятыми audit gaps: пользовательский слой в коде существенно продвинут, но formal verification evidence для фаз `8-12` и повторная live-проверка части post-closure runtime fixes остаются отдельным техдолгом.

## Core Value

Система должна давать оператору надёжный, интерпретируемый и воспроизводимый результат обследования специалиста, основанный на полном мультимодальном анализе речевых ответов, а не на ручной субъективной оценке.

## Current Milestone: None

**Last archived milestone:** `v1.1 UI & Admin Completion` on 2026-03-25  
**Archive status:** accepted with audit gaps

## Requirements

### Validated

- ✓ Пользователь может пройти защищённый auth flow с refresh/logout и server-enforced RBAC boundary — v1.0
- ✓ Оператор может создать обследование, загрузить ответы, завершить его идемпотентно и увидеть историю с актуальными статусами — v1.0
- ✓ Система выполняет асинхронную мультимодальную обработку через RabbitMQ для обязательных каналов `text`, `acoustic` и `paralinguistic` — v1.0
- ✓ Aggregator формирует один канонический baseline-aware профиль только после успешного завершения всех обязательных каналов — v1.0
- ✓ Система передаёт нормализованный профиль в WiMi/KЭСМИ integration boundary и честно показывает `analysis_not_implemented_yet`, пока реальная decision-модель отсутствует — v1.0
- ✓ Frontend показывает processing progress, result surface, history re-entry, baseline deviation, channel contributions и integration diagnostics — v1.0
- ✓ Система ведёт audit trail, readiness/metrics surfaces и end-to-end correlation/tracing на production-like уровне — v1.0
- ✓ Все milestone v1 требования закрыты и подтверждены audit-пакетом `30/30` — v1.0

### Active

- [ ] Закрыть verification debt milestone `v1.1`: добавить `VERIFICATION.md` для фаз `8-12` и повторно прогнать milestone audit на актуальном коде
- [ ] Повторно проверить live runtime пути operator auth/examination save/processing/result после post-closure fixes
- [ ] Определить scope следующего milestone поверх уже shipped платформы и UI/admin baseline

### Out of Scope

- Публичная регистрация и self-service пользовательские сценарии — система предназначена для внутреннего контура организации
- Публичный доступ к объектному хранилищу или передача бинарных аудиофайлов через RabbitMQ — противоречит архитектурным ограничениям проекта
- Замена операторского решения полностью автоматическим решением системы — итоговое управленческое решение остаётся за оператором и внешней системой поддержки принятия решений
- Массовый consumer/mobile-first интерфейс — приоритетом являются рабочие станции операторов и административная панель во внутреннем контуре

## Context

- Основной источник истины по продукту и архитектуре: `docs/00_project.md`
- Контракты API и интеграций фиксируются в `docs/01_contract.md`, а история реализации — в `docs/02_implementation.md`
- Текущий shipped baseline включает `frontend/` на Next.js App Router, `core-backend/` на Go, PostgreSQL, RabbitMQ, MinIO, Python ML services, `ml-baseline`, WiMi/KЭСМИ runtime и Prometheus-backed observability surfaces
- Milestone `v1.0` закрыт с passed audit: `30/30 requirements`, `7/7 phases`, `6/6 integration`, `6/6 flows`
- Milestone `v1.1` архивирован 2026-03-25 по прямому решению пользователя, но с сохранённым `gaps_found` audit и accepted verification debt
- Архитектура дипломного проекта по-прежнему предполагает отдельные ML-сервисы, baseline service, integration boundary с КЭСМИ и идемпотентные source-of-truth переходы workflow
- На текущем этапе всё ещё допускается отсутствие финальных ML- и KЭСМИ-моделей: это уже не архитектурный blocker, а следующее предметное расширение поверх стабилизированной orchestration-платформы

## Current State

- Milestone `v1.0 MVP` shipped and archived on 2026-03-24
- Milestone `v1.1 UI & Admin Completion` archived on 2026-03-25 with accepted audit gaps
- Latest archived UI/admin milestone delivered 5 phases, 24 plans, and a denser operator/admin control surface on top of the `v1.0` platform
- Approximate codebase size across Go, TypeScript/TSX, and Python sources: `28,816` lines in the current working tree estimate
- Current planning state has no active milestone; the next step is either verification-debt closure or next-milestone definition

## Next Milestone Goals

- Решить, превращается ли archived `v1.1` verification debt в отдельный cleanup milestone или принимается как исторически зафиксированный риск
- При необходимости формально довести UI/admin milestone до auditable состояния через `VERIFICATION.md` и live revalidation
- После этого определить следующий product milestone поверх уже shipped orchestration и UI/admin foundation

## Constraints

- **Architecture**: PostgreSQL остаётся источником истины по состояниям обследований и задач — это прямо задано в `docs/00_project.md`
- **Workflow**: Все три аналитических канала обязательны для успешного завершения обследования — финальное успешное состояние невозможно при провале хотя бы одного канала
- **Integration**: RabbitMQ используется только для метаданных и ссылок на S3-объекты, без передачи бинарных файлов — это ограничение нельзя нарушать
- **Security**: Система работает во внутреннем контуре, не должна публиковать внутренние сервисы, логировать чувствительные данные или отдавать наружу сырой stack trace
- **Documentation**: Любое изменение контрактов должно обновлять `docs/01_contract.md`, а заметные изменения реализации должны фиксироваться в `docs/02_implementation.md`
- **Deployment**: Решение должно оставаться self-hosted и воспроизводимо запускаться локально через контейнерный стек

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Развивать проект как brownfield, а не как greenfield | В репозитории уже реализованы базовые operator/admin сценарии и инфраструктурный каркас | ✓ Good |
| Использовать `docs/00_project.md` как основной продуктовый источник для инициализации PROJECT.md | Пользователь явно указал, что полное описание проекта и ответы на ключевые вопросы уже содержатся в ТЗ | ✓ Good |
| Зафиксировать существующий функционал как Validated, а оставшийся путь до целевой архитектуры как Active | Это позволяет планировать workstream по gap between current implementation and target architecture без потери уже построенного | ✓ Good |
| Развивать ML и KЭСМИ integration contract-first, даже при отсутствии финальных моделей | Это позволяет завершить архитектурные и workflow-фазы заранее и отложить только feature mapping и model-specific verification | ✓ Good |
| Инициализировать workflow с commit tracking и полным набором quality agents | Проект high-stakes, многосервисный и интеграционный; дешёвые shortcuts здесь повышают риск неверного плана | ✓ Good |
| Сфокусировать milestone `v1.1` на frontend/admin completion, а не на ML или decision expansion | Пользовательский слой уже является главным видимым gap после закрытия v1 orchestration-platform | ✓ Good |
| Ограничить backend-изменения точечными API-доработками под UI/admin нужды | Это сохраняет архитектурную стабильность и не раздувает milestone за пределы UI-first scope | ✓ Good |
| Архивировать `v1.1` с сохранённым failed audit по прямому решению пользователя | Пользователь предпочёл зафиксировать milestone как завершённый организационно, не скрывая verification debt | ⚠ Revisit |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `$gsd-transition`):
1. Requirements invalidated? -> Move to Out of Scope with reason
2. Requirements validated? -> Move to Validated with phase reference
3. New requirements emerged? -> Add to Active
4. Decisions to log? -> Add to Key Decisions
5. "What This Is" still accurate? -> Update if drifted

**After each milestone** (via `$gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check -> still the right priority?
3. Audit Out of Scope -> reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-03-25 after archiving milestone v1.1 UI & Admin Completion*
