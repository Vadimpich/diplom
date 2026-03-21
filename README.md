# Dimplom local stack

Локальный self-hosted стек для Phase 2:

- PostgreSQL
- RabbitMQ с management UI
- MinIO
- core backend
- frontend
- `text`, `acoustic`, `paralinguistic` stub workers

## Требования

- Docker
- Docker Compose v2

## Подготовка

1. Скопируйте пример переменных окружения:

```bash
cp .env.example .env
```

2. При необходимости скорректируйте значения в `.env`.

Основные группы переменных:

- `POSTGRES_*`
- `RABBITMQ_*`
- `RABBITMQ_URL`
- `MINIO_*`
- `HTTP_ADDR`
- `JWT_*`
- `PROCESSING_*`
- `*_WORKER_*`

`DATABASE_URL` и `RABBITMQ_URL` для core backend собираются через `docker-compose.yml`/`.env`, поэтому backend и workers используют один и тот же локальный broker topology без ручного bootstrap.

## Запуск

Поднять весь стек:

```bash
docker compose up -d --build
```

Проверить, что все сервисы поднялись:

```bash
docker compose ps
```

Если нужно поднять только инфраструктуру без frontend/backend/worker-сервисов:

```bash
docker compose up -d postgres rabbitmq minio
```

## Доступные порты

- PostgreSQL: `5432`
- RabbitMQ AMQP: `5672`
- RabbitMQ management: `15672`
- MinIO API: `9000`
- MinIO console: `9001`
- core backend: `8080`
- frontend: `3000`

## Проверка Phase 2

Полная локальная проверка Phase 2:

```bash
docker compose up -d --build
docker compose ps
cd /home/katya/dimplom/core-backend && go test ./... -count=1
cd /home/katya/dimplom/frontend && npm run lint && npm run build && npx tsc --noEmit
```

Примечание по frontend: `next build` генерирует актуальные `.next/types` для App Router. Если рабочее дерево уже содержит устаревшие `.next` артефакты, запускайте `npm run build` перед отдельным `npx tsc --noEmit`.
Если локальный RabbitMQ volume остался от ранних Phase 2 итераций с другой topology, один раз выполните `docker compose down -v` перед `up -d --build`.

## Примечание

В compose уже зафиксированы queue/exchange env для `processing.commands`, `processing.results`, `qq.processing.text`, `qq.processing.acoustic`, `qq.processing.paralinguistic`, общего result routing key и `PROCESSING_OUTBOX_MAX_ATTEMPTS=3`. Backend relay и worker-сервисы должны использовать один и тот же retry/DLX baseline локального async pipeline.
