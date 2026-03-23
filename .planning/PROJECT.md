# Мультимодальная система оценки психоэмоционального состояния специалистов

## What This Is

Self-hosted система для проведения обследований специалистов критических областей на основе записи речевых ответов и их мультимодального анализа. Продукт используется внутри организации оператором и администратором, чтобы проводить обследования, получать интерпретируемый результат и передавать агрегированный профиль состояния во внешнюю систему поддержки принятия решений о допуске к профессиональной деятельности.

По состоянию на milestone `v1.0` система уже покрывает полный опорный контур из дипломного scope: защищённый operator/admin доступ, intake обследований, асинхронную обработку через обязательные каналы `text` / `acoustic` / `paralinguistic`, baseline-aware aggregation, delivery в WiMi/KЭСМИ boundary, operator-facing result surface, audit trail, readiness/metrics и verification evidence для milestone closure.

## Core Value

Система должна давать оператору надёжный, интерпретируемый и воспроизводимый результат обследования специалиста, основанный на полном мультимодальном анализе речевых ответов, а не на ручной субъективной оценке.

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

- [ ] Встроить реальную decision-модель WiMi и утвердить окончательный feature mapping вместо pre-model `analysis_not_implemented_yet`
- [ ] Довести ML-модели каналов от stub/runtime-ready реализации до финального предметного качества и зафиксированных feature schemas
- [ ] Добавить operator/admin возможности следующего слоя: TTL management, audit log browsing и техническую статистику без прямого обращения к инфраструктуре
- [ ] Рассмотреть confidence/explainability extension для итогового decision flow после стабилизации реальных моделей
- [ ] Подготовить следующий milestone с новыми requirement-границами вместо повторного использования v1 backlog

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
- Архитектура дипломного проекта по-прежнему предполагает отдельные ML-сервисы, baseline service, integration boundary с КЭСМИ и идемпотентные source-of-truth переходы workflow
- На текущем этапе всё ещё допускается отсутствие финальных ML- и KЭСМИ-моделей: это уже не архитектурный blocker, а следующее предметное расширение поверх стабилизированной orchestration-платформы

## Current State

- Milestone `v1.0 MVP` shipped and archived on 2026-03-24
- 7 phases completed, 36 plans completed, 25 milestone tasks recorded by archival workflow
- Approximate codebase size in shipped stack: `80,821` lines across Go, TypeScript/TSX, and Python sources
- Git planning range used for milestone traceability: `1f61843` -> `2b2ff47`

## Next Milestone Goals

- Подготовить новый milestone вокруг реальных ML-моделей и live WiMi decision mapping
- Формализовать новые requirements вместо продолжения работы поверх archived v1 requirement matrix
- Решить, какие deferred items остаются tech debt, а какие становятся активными deliverables следующего milestone

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

---
*Last updated: 2026-03-24 after v1.0 milestone completion*
