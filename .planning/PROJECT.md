# Мультимодальная система оценки психоэмоционального состояния специалистов

## What This Is

Self-hosted система для проведения обследований специалистов критических областей на основе записи речевых ответов и их мультимодального анализа. Продукт используется внутри организации оператором и администратором, чтобы проводить обследования, получать интерпретируемый результат и передавать агрегированный профиль состояния во внешнюю систему поддержки принятия решений о допуске к профессиональной деятельности.

По состоянию после архивации milestone `v1.1` система уже покрывает полный опорный контур из дипломного scope и имеет отдельные operator/admin интерфейсы, административные control surfaces и уплотнённый рабочий UI. При этом `v1.1` закрыт с принятыми audit gaps: пользовательский слой в коде существенно продвинут, но formal verification evidence для фаз `8-12` и повторная live-проверка части post-closure runtime fixes остаются отдельным техдолгом.

## Core Value

Система должна давать оператору надёжный, интерпретируемый и воспроизводимый результат обследования специалиста, основанный на полном мультимодальном анализе речевых ответов, а не на ручной субъективной оценке.

## Current Milestone: v1.2 Operator UI

**Goal:** полностью перестроить operator UX как рабочий инструмент с search-first, list-first и action-first взаимодействием, без dashboard-паттернов и лишнего текста.

**Target features:**
- Полностью пересобранный login screen с одной центрированной формой без информационных панелей.
- Workflow-first operator home, список специалистов и история обследований без карточного dashboard UI.
- Упрощённый examination screen, в котором оператор видит только вопрос, запись и прогресс.

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

- [ ] Пересобрать operator UX как рабочий инструмент, а не dashboard
- [ ] Убрать dashboard-карточки, лишние тексты и конкурирующие блоки со всех ключевых operator-экранов из milestone scope
- [ ] Свести operator взаимодействие к search-first, list-first и action-first flow без изменения admin интерфейса

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
- Новый milestone `v1.2` не добавляет новые product capabilities, а целенаправленно заменяет концептуально неверную interaction model operator UI
- Архитектура дипломного проекта по-прежнему предполагает отдельные ML-сервисы, baseline service, integration boundary с КЭСМИ и идемпотентные source-of-truth переходы workflow
- На текущем этапе всё ещё допускается отсутствие финальных ML- и KЭСМИ-моделей: это уже не архитектурный blocker, а следующее предметное расширение поверх стабилизированной orchestration-платформы

## Current State

- Milestone `v1.0 MVP` shipped and archived on 2026-03-24
- Milestone `v1.1 UI & Admin Completion` archived on 2026-03-25 with accepted audit gaps
- Milestone `v1.2 Operator UI` starts from a shipped platform where operator/admin surfaces already exist, but operator interaction model is considered conceptually wrong and subject to replacement
- Approximate codebase size across Go, TypeScript/TSX, and Python sources: `28,816` lines in the current working tree estimate
- Current planning state is reset for a new operator-only milestone starting at Phase 13

## Next Milestone Goals

- Перестроить operator UI вокруг линейного рабочего процесса `найти специалиста -> открыть -> провести обследование`
- Убрать dashboard-паттерны, визуальный шум и весь текст, не влияющий на действие оператора
- Сохранить существующие возможности системы, меняя только interaction model и operator-facing structure

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
| Перестраивать `v1.2` вокруг operator workflow, а не вокруг визуального polish | Пользователь явно потребовал заменить interaction model и запретил dashboard-style улучшательства | ✓ Good |
| Не трогать admin interface в `v1.2` | Scope milestone жёстко ограничен operator UX restructuring | ✓ Good |

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
*Last updated: 2026-03-25 after starting milestone v1.2 Operator UI*
