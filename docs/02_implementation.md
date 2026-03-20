# Журнал реализации

## 2026-03-19

- В core backend добавлен минимальный CORS middleware на уровне общего chi-router: preflight `OPTIONS` больше не падает `405`, разрешённые origin берутся из `CORS_ALLOWED_ORIGINS` с локальными значениями по умолчанию `http://localhost:3000,http://127.0.0.1:3000`.
- Созданы обязательные документы `docs/01_contract.md` и `docs/02_implementation.md` в соответствии с требованиями `AGENTS.md`.
- Обновлены конфигурации субагентов в `.codex/agents/*.toml`: добавлены явные требования читать `docs/00_project.md` и `docs/01_contract.md`, синхронизироваться через документы, фиксировать изменения контрактов в `docs/01_contract.md` и добавлять записи в `docs/02_implementation.md`.
- Добавлен локальный инфраструктурный каркас: `docker-compose.yml`, `.env.example`, `.dockerignore`, `README.md`; backend Dockerfile переведён на сборку из текущего Go-каркаса с healthcheck по `/health`.
- Добавлен минимальный backend-каркас на Go: `cmd/api`, `internal/*`, конфигурация из env, HTTP-сервер на `chi`, health-check `GET /health` с проверкой PostgreSQL через `pgx`, стартовая миграция `users`, базовые `db/` и `Dockerfile`.
- Реализован базовый синхронный домен core backend: роли и пользователи с JWT login `/auth/login` и `/me`, CRUD специалистов, обследования со статусами `created` -> `collecting_answers` -> `ready_for_processing`, загрузка ответов в MinIO/S3 через `POST /answers`, полная схема БД и `sqlc`-repository слой.
- Доведён runtime домена до рабочего состояния: добавлен `POST /users`, автоприменение SQL-миграций при старте backend, bootstrap начального пользователя через env и совместимая загрузка S3-конфига для MinIO в локальном Docker Compose.
- Минимально завершён stage 3 backend API для frontend-сценариев: добавлены `GET/PUT /users`, `GET /examinations`, `GET /examinations/{id}`, `GET /specialists/{id}/examinations`, CRUD API опросников, sqlc-репозиторий для `questionnaires/questions/questionnaire_questions` и минимальная привязка `examinations.questionnaire_id` для operator flow.
- Реализован stage 3 frontend в `frontend/` на Next.js App Router + TypeScript: auth/login, TanStack Query и API-клиент, server-side protected layouts для `/operator/*` и `/admin/*`, operator flow со специалистами, обследованиями, MediaRecorder, processing/results shells, а также admin UI для пользователей, опросников, monitoring и settings placeholders по актуальным контрактам `docs/01_contract.md`.
- Реализован stage 3 frontend в `frontend/` на Next.js App Router: отдельные контуры `/operator/*` и `/admin/*` с раздельными layout, JWT login и route guards, базовый API-клиент, TanStack Query, Tailwind-стили и совместимая shadcn/ui-база компонентов.
- Подключён приоритетный operator flow: dashboard, CRUD специалистов, создание и прохождение обследования с выбором опросника, `MediaRecorder`-запись и загрузка ответов в `POST /answers`, экраны processing/results как честные UI-shell, история обследований по `GET /examinations` и `GET /specialists/{id}/examinations`.
- Подключён admin flow по актуальным контрактам: список/создание/редактирование пользователей, список/создание/редактирование опросников, monitoring по `GET /health`, placeholders для settings; frontend проверен командами `eslint`, `tsc --noEmit` и `next build`.
- Добавлен `frontend` сервис в корневой `docker-compose.yml` с production Dockerfile, `NEXT_PUBLIC_API_URL=http://core-backend:8080` внутри сети Compose и минимальными env/contract обновлениями для локального запуска.

## 2026-03-20

- Сгенерирована актуальная карта кодовой базы в `.planning/codebase/`: `STACK.md`, `INTEGRATIONS.md`, `ARCHITECTURE.md`, `STRUCTURE.md`, `CONVENTIONS.md`, `TESTING.md`, `CONCERNS.md` на основе текущего состояния `core-backend/`, `frontend/`, `docs/` и инфраструктурных файлов для последующего планирования фаз и онбординга.
- Инициализирован GSD planning-контур проекта для brownfield-репозитория: добавлены `.planning/PROJECT.md`, `.planning/config.json`, `.planning/REQUIREMENTS.md`, `.planning/ROADMAP.md`, `.planning/STATE.md`; зафиксированы существующие validated capabilities, целевые v1 requirements и roadmap из 5 фаз для доведения реализации до архитектуры из `docs/00_project.md`.
- Созданы planning-артефакты `.planning/ROADMAP.md` и `.planning/STATE.md`, а также обновлена traceability-матрица в `.planning/REQUIREMENTS.md`: roadmap сжат до 5 фаз по coarse granularity и сфокусирован на brownfield-gap к целевой архитектуре из `docs/00_project.md`.
