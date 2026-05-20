# Контракты системы

## HTTP API

### Examination status vocabulary

Единый статусный словарь обследования на Phase 4:

- `created` — обследование создано, но оператор ещё не начал сбор ответов;
- `collecting_answers` — идёт запись и загрузка ответов;
- `ready_for_processing` — ответы собраны и обследование готово к асинхронной обработке;
- `processing` — core backend ожидает результаты обязательных каналов `text`, `acoustic`, `paralinguistic`;
- `aggregating` — все обязательные каналы успешно завершились, core backend формирует канонический агрегированный профиль и обогащает его baseline-данными;
- `aggregated` — агрегированный профиль и baseline snapshot успешно сохранены, но доставка решения во внешний decision layer ещё не завершена;
- `decision_pending` — decision snapshot уже зафиксирован в PostgreSQL и ожидает первой или повторной доставки во внешний decision engine;
- `completed` — decision delivery завершён terminal outcome, а оператор видит нормализованный `decision_result`, даже если recommendation пока недоступна;
- `failed` — хотя бы один обязательный канал или post-processing этап завершился terminal failure.

Правила:
- PostgreSQL остаётся source of truth для coarse workflow status;
- `terminal=true` в `GET /examinations/{id}/processing-status` означает только `completed` или `failed`;
- факт успешного завершения всех каналов сам по себе больше не означает terminal success;
- Phase 4 decision delivery не раскрывает raw WiMi payload наружу; operator-facing API получает только нормализованные backend-owned DTO.

### GET /health

Назначение:
- cheap liveness-check core backend;
- проверка только того, что HTTP-процесс жив и способен отвечать.

Ответ `200 OK`:

```json
{
  "status": "ok",
  "service": "core-backend"
}
```

Технические детали:
- `Content-Type: application/json`;
- endpoint без аутентификации;
- endpoint не должен делать dependency fan-in;
- проверка PostgreSQL, RabbitMQ, MinIO, baseline и WiMi вынесена в `GET /ready`;
- Prometheus exposition вынесена в `GET /metrics`.

## Operational Trustworthiness Contract

### audit_event

Назначение:
- append-only audit trail для критичных действий и source-of-truth переходов;
- единая схема для auth, admin mutations, workflow переходов и decision delivery.

Контракт:

```json
{
  "id": 901,
  "event_type": "processing.launch",
  "event_key": "processing-launch:examination:101",
  "outcome": "succeeded",
  "happened_at": "2026-03-23T18:10:12Z",
  "request_id": "req-4b09d7d9b1b8",
  "trace_id": "8ec8c1b6409f4a6cb80cfcb4f74aa98c",
  "traceparent": "00-8ec8c1b6409f4a6cb80cfcb4f74aa98c-5d7c1f97db7840b3-01",
  "correlation_id": "exam-101-text-v1",
  "actor": {
    "user_id": 1,
    "login": "operator",
    "role_slug": "operator",
    "ip": "127.0.0.1",
    "user_agent": "Mozilla/5.0"
  },
  "resource": {
    "kind": "examination",
    "id": 101
  },
  "domain_refs": {
    "specialist_id": 55,
    "questionnaire_id": 7,
    "channel": "text",
    "decision_snapshot_id": 44
  },
  "payload": {
    "status_from": "ready_for_processing",
    "status_to": "processing"
  }
}
```

Поля:
- `event_type`: стабильный namespaced тип события;
- `event_key`: optional stable deduplication key для идемпотентных transition fences;
- `outcome`: одно из `succeeded`, `failed`, `rejected`;
- `request_id`: request-local идентификатор HTTP boundary;
- `trace_id`: canonical distributed trace identifier;
- `traceparent`: transport-safe W3C trace context, если событие пришло из traced path;
- `correlation_id`: business/debug correlation, уже используемый в processing и decision flow;
- `actor`: nullable для system-originated событий;
- `resource.kind`: одно из `auth_session`, `user`, `questionnaire`, `examination`, `channel_result`, `decision_snapshot`;
- `domain_refs`: domain-specific координаты без raw payload leakage;
- `payload`: low-sensitivity JSON для transition facts и diagnostics, без аудио, access tokens и stack trace.

Обязательные `event_type` для Phase 5:
- `auth.login`
- `auth.login_failed`
- `admin.user_created`
- `admin.user_updated`
- `admin.settings_updated`
- `admin.questionnaire_created`
- `admin.questionnaire_updated`
- `examination.created`
- `examination.started`
- `examination.finished`
- `processing.launch`
- `processing.result_received`
- `aggregation.completed`
- `decision.completed`
- `decision.failed`

Правила:
- audit storage append-only; update/delete audit rows запрещены;
- событие должно писаться из source-of-truth backend boundary, а не из frontend, worker-а или raw log sink;
- для fenced idempotent transitions допустим ровно один semantic audit event на один `event_key`;
- failed login обязан попадать в audit trail даже без успешной доменной транзакции;
- raw credentials, refresh tokens, raw WiMi payload и raw channel payload в `payload` запрещены.

### GET /audit/events

Назначение:
- admin-only read surface для UI-аудита без прямого доступа к внутреннему storage package.

Аутентификация:
- `Authorization: Bearer <jwt>`;
- доступно только роли `admin`.

Query parameters:
- `event_type`: optional точный фильтр по `audit_event.event_type`;
- `resource_kind`: optional фильтр по связанному объекту (`user`, `questionnaire`, `examination`, `channel_result`, `decision_snapshot`, `auth_session`);
- `resource_id`: optional `bigint` идентификатор связанного объекта; допускается только вместе с `resource_kind`;
- `from`: optional lower bound для `happened_at` в RFC3339;
- `to`: optional upper bound для `happened_at` в RFC3339;
- `limit`: optional количество записей, `1..200`, default `100`.

Ответ `200 OK`:

```json
{
  "items": [
    {
      "id": 901,
      "event_type": "decision.completed",
      "event_key": "decision:44:completed",
      "outcome": "succeeded",
      "happened_at": "2026-03-24T16:30:00Z",
      "request_id": "req-4b09d7d9b1b8",
      "trace_id": "8ec8c1b6409f4a6cb80cfcb4f74aa98c",
      "traceparent": "00-8ec8c1b6409f4a6cb80cfcb4f74aa98c-5d7c1f97db7840b3-01",
      "correlation_id": "exam-101-kesmi-1",
      "actor": {
        "user_id": 1,
        "login": "admin",
        "role_slug": "admin",
        "ip": "127.0.0.1",
        "user_agent": "Mozilla/5.0"
      },
      "resource": {
        "kind": "decision_snapshot",
        "id": 44
      },
      "domain_refs": {
        "examination_id": 101,
        "decision_snapshot_id": 44
      },
      "payload": {
        "recommendation": "unavailable"
      }
    }
  ]
}
```

Правила:
- сортировка по `happened_at DESC`;
- endpoint возвращает тот же canonical `audit_event`, который пишет backend, без отдельной read-model;
- `resource_id` без `resource_kind` возвращает `400 Bad Request`;
- `from > to`, невалидный RFC3339 и `limit` вне диапазона тоже возвращают `400 Bad Request`.

### GET /settings

Назначение:
- чтение mutable subset system settings из persisted core-owned source of truth.

Аутентификация:
- `Authorization: Bearer <jwt>`;
- доступно только роли `admin`.

Ответ `200 OK`:

```json
{
  "audio_retention_ttl_days": 30,
  "processing_max_attempts": 3,
  "kesmi_max_retries": 2,
  "created_at": "2026-03-24T11:00:00Z",
  "updated_at": "2026-03-24T11:30:00Z"
}
```

Правила:
- endpoint не раскрывает secrets, hostnames, bootstrap credentials и env-only runtime topology;
- значения относятся только к mutable subset, который core backend действительно хранит и читает;
- `processing_max_attempts` применяется к будущим processing launches;
- `kesmi_max_retries` применяется к новым decision snapshots;
- `audio_retention_ttl_days` хранится как persisted retention policy для audio artifacts и не является proxy для правки `.env`.

### PUT /settings

Назначение:
- обновление mutable subset system settings.

Аутентификация:
- `Authorization: Bearer <jwt>`;
- доступно только роли `admin`.

Запрос:

```json
{
  "audio_retention_ttl_days": 45,
  "processing_max_attempts": 4,
  "kesmi_max_retries": 3
}
```

Правила валидации:
- `audio_retention_ttl_days`: `1..365`;
- `processing_max_attempts`: `1..10`;
- `kesmi_max_retries`: `1..10`.

Ответ `200 OK`: обновлённый объект `GET /settings`.

Ошибки:
- `400 Bad Request` если payload невалидный;
- `401 Unauthorized` если токен отсутствует или невалиден;
- `403 Forbidden` если роль не `admin`.

### Runtime health, readiness, metrics

Общие правила:
- `GET /health` отвечает только за cheap liveness конкретного процесса;
- `GET /ready` проверяет зависимости, без которых сервис не может безопасно принимать traffic/work;
- `GET /metrics` отдаёт Prometheus exposition format;
- health/readiness endpoints не требуют аутентификации;
- label values в метриках должны быть низкокардинальными: status, route template, channel, dependency, outcome, service.
- examination IDs, specialist IDs, user IDs, `correlation_id`, `request_id`, `trace_id` и free-form error text в labels запрещены.

#### core-backend

`GET /health`

```json
{
  "status": "ok",
  "service": "core-backend"
}
```

Правила:
- endpoint не делает dependency fan-in;
- допускается только in-process sanity check.

`GET /ready`

```json
{
  "status": "ready",
  "service": "core-backend",
  "dependencies": {
    "postgres": "up",
    "rabbitmq": "up",
    "minio": "up",
    "baseline": "up",
    "kesmi": "up"
  }
}
```

Правила:
- `503 Service Unavailable`, если хотя бы одна dependency из shipped workflow недоступна;
- readiness учитывает `postgres`, `rabbitmq`, `minio`, `ml-baseline`, `wimi`;
- `/health` и `/ready` не должны менять состояние системы.

`GET /metrics`

Пример обязательных metric families:
- `diplom_http_requests_total{route,method,status_class}`
- `diplom_http_request_duration_seconds_bucket{route,method}`
- `diplom_processing_messages_total{channel,status}`
- `diplom_decision_attempts_total{outcome}`
- `diplom_dependency_up{dependency}`

#### frontend

`GET /api/health`

```json
{
  "status": "ok",
  "service": "frontend"
}
```

`GET /api/ready`

```json
{
  "status": "ready",
  "service": "frontend",
  "dependencies": {
    "core_backend": "up"
  }
}
```

`GET /api/metrics`

Пример metric families:
- `diplom_frontend_requests_total{route,method,status_class}`
- `diplom_frontend_request_duration_seconds_bucket{route,method}`
- `diplom_frontend_dependency_up{dependency}`

Правила:
- frontend readiness зависит только от достижимости `core-backend`, а не от PostgreSQL/RabbitMQ напрямую.

#### text-worker / acoustic-worker / paralinguistic-worker

`GET /health`

```json
{
  "status": "ok",
  "service": "text-worker",
  "channel": "text"
}
```

`GET /ready`

```json
{
  "status": "ready",
  "service": "text-worker",
  "channel": "text",
  "dependencies": {
    "rabbitmq": "up",
    "minio": "up"
  }
}
```

`GET /metrics`

Пример metric families:
- `diplom_worker_messages_total{channel,status}`
- `diplom_worker_message_duration_seconds_bucket{channel,status}`
- `diplom_worker_dependency_up{channel,dependency}`

Правила:
- worker readiness degraded, если consumer не готов, RabbitMQ недоступен или MinIO недоступен;
- `/health` может оставаться `200` во время reconnect loop, если процесс жив;
- worker logs обязаны включать `channel`, `correlation_id`, `request_id` и `trace_id`, когда они доступны.

#### ml-baseline

`GET /health`

```json
{
  "status": "ok",
  "service": "ml-baseline"
}
```

`GET /ready`

```json
{
  "status": "ready",
  "service": "ml-baseline"
}
```

`GET /metrics`

Пример metric families:
- `diplom_baseline_requests_total{route,method,status_class}`
- `diplom_baseline_request_duration_seconds_bucket{route,method}`

Правила:
- `ml-baseline` не invent-ит PostgreSQL или RabbitMQ dependency checks;
- readiness зависит только от собственного process/runtime health и возможности обрабатывать HTTP requests.

### Correlation and trace propagation

Общие правила:
- `correlation_id` сохраняется как business/debug identifier и не заменяется trace IDs;
- canonical distributed trace transport: W3C `traceparent`, optional `tracestate`;
- каждый inbound HTTP request в `core-backend` получает `request_id` и trace context;
- если клиент не прислал `traceparent`, `core-backend` генерирует новый trace;
- audit events, structured logs и outbound calls используют один и тот же `request_id`/`trace_id` pair для текущего request path.

HTTP:
- inbound: `traceparent`, `tracestate`, `X-Request-Id` принимаются и нормализуются;
- outbound from `core-backend`: `traceparent`, optional `tracestate`, `X-Request-Id`, `X-Correlation-Id` отправляются в `ml-baseline` и WiMi.

RabbitMQ processing envelopes:
- `ProcessingCommandEnvelope` и `ChannelResultEnvelope` расширяются полями `request_id`, `traceparent`, optional `tracestate`;
- RabbitMQ message properties должны дублировать `message_id`/`correlation_id` и carry trace context, чтобы worker мог восстановить trace even if payload logger is incomplete;
- workers обязаны сохранять пришедший `correlation_id`, продолжать `traceparent` и возвращать те же correlation fields в result envelope.

KESMI / WiMi boundary:
- `core-backend` остаётся единственным caller WiMi;
- outbound WiMi request обязан нести `traceparent` и `X-Correlation-Id`;
- raw WiMi response не становится trace carrier для внешнего frontend API.

## Auth API

### POST /auth/login

Назначение:
- вход пользователя по логину и паролю;
- выпуск short-lived JWT access token;
- выпуск opaque `refresh_token` для доверенной frontend BFF boundary.

Транспортная модель:
- core backend НЕ устанавливает browser cookies самостоятельно;
- frontend BFF или route handlers на origin фронтенда принимают `refresh_token` из ответа и кладут его в `HttpOnly` cookie;
- browser не должен хранить `access_token` и `refresh_token` в JavaScript-readable storage.
- frontend использует BFF endpoints `/api/auth/login`, `/api/auth/refresh`, `/api/auth/logout` и `/api/auth/session` как единственную browser-cookie boundary.

Запрос:

```json
{
  "login": "operator",
  "password": "secret"
}
```

Ответ `200 OK`:

```json
{
  "access_token": "<jwt>",
  "refresh_token": "<opaque-refresh-token>",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_expires_in": 604800,
  "user": {
    "id": 1,
    "login": "operator",
    "role": {
      "id": 2,
      "slug": "operator",
      "name": "Operator"
    },
    "is_active": true,
    "last_login_at": "2026-03-25T08:15:00Z",
    "created_at": "2026-03-19T12:00:00Z",
    "updated_at": "2026-03-19T12:00:00Z"
  }
}
```

Поле `refresh_expires_in`:
- срок жизни refresh-session в секундах;
- используется frontend BFF для корректной установки `HttpOnly` refresh-cookie независимо от short-lived access token TTL.

Поле `user.last_login_at`:
- nullable RFC3339 timestamp последнего успешного логина;
- обновляется только после успешного `POST /auth/login`;
- чтение этого поля через `/me`, `GET /users` и `GET /users/{id}` не создаёт новых auth/admin mutation событий.

Ошибки:
- `401 Unauthorized` при неверных учётных данных;
- `403 Forbidden` если пользователь неактивен.

### POST /auth/refresh

Назначение:
- обмен активной refresh-session на новый short-lived access token;
- ротация refresh session в PostgreSQL с отзывом предыдущей сессии.

Транспортная модель:
- endpoint предназначен для trusted frontend BFF boundary;
- BFF читает `HttpOnly` cookie на своём origin, отправляет `refresh_token` в core backend и получает новый `refresh_token` для обновления cookie.

Запрос:

```json
{
  "refresh_token": "<opaque-refresh-token>"
}
```

Ответ `200 OK`:

```json
{
  "access_token": "<jwt>",
  "refresh_token": "<rotated-opaque-refresh-token>",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_expires_in": 604800,
  "user": {
    "id": 1,
    "login": "operator",
    "role": {
      "id": 2,
      "slug": "operator",
      "name": "Operator"
    },
    "is_active": true,
    "created_at": "2026-03-19T12:00:00Z",
    "updated_at": "2026-03-19T12:00:00Z"
  }
}
```

Ошибки:
- `400 Bad Request` если payload невалидный;
- `401 Unauthorized` если refresh token отсутствует, истёк, отозван или уже был использован;
- `403 Forbidden` если пользователь деактивирован.

### POST /auth/logout

Назначение:
- завершение refresh-session;
- server-side отзыв refresh token без прямого управления browser cookies со стороны core backend.

Запрос:

```json
{
  "refresh_token": "<opaque-refresh-token>"
}
```

Ответ `204 No Content`.

Технические детали:
- endpoint идемпотентен для trusted BFF: если сессия уже отозвана или отсутствует, backend не раскрывает это наружу;
- после успешного вызова BFF должен удалить свой `HttpOnly` cookie.

Ошибки:
- `400 Bad Request` если payload невалидный.

### Frontend BFF Auth API

Назначение:
- Next.js route handlers на frontend origin проксируют auth transport к core backend;
- только BFF layer устанавливает и очищает browser cookies.

Endpoints:
- `POST /api/auth/login` -> прокси к `POST /auth/login`, устанавливает browser cookies после успешного входа;
- `POST /api/auth/refresh` -> прокси к `POST /auth/refresh`, ротирует browser cookies;
- `POST /api/auth/logout` -> прокси к `POST /auth/logout`, очищает browser cookies;
- `GET /api/auth/session` -> возвращает текущего пользователя, при необходимости обновляя access token через refresh cookie.

Технические правила:
- `diplom_refresh_token` хранится только как `HttpOnly` cookie и не должен читаться браузерным JavaScript;
- BFF может выставлять вспомогательные cookies для переходного UI-слоя, но transport refresh-session остаётся централизованным в `/api/auth/*`;
- frontend-клиент не должен писать auth cookies через `document.cookie`.
- protected layouts и route guards должны опираться на `/api/auth/session` как на авторитетный источник session/role state; role-cookie допустим только как UX hint для первичного redirect.
- browser-side API client должен уметь один раз автоматически выполнить `POST /api/auth/refresh` после `401 Unauthorized` и повторить исходный запрос с новым access token, если refresh-session ещё валидна.

### CORS / preflight

Технические правила:
- core backend обрабатывает CORS на уровне общего HTTP middleware до роутинга endpoint;
- cross-origin запросы разрешены только для origin из runtime-параметра `CORS_ALLOWED_ORIGINS`;
- значение по умолчанию для локальной разработки: `http://localhost:3000,http://127.0.0.1:3000`;
- для разрешённого origin backend возвращает `Access-Control-Allow-Origin` со значением исходного `Origin`;
- разрешённые методы для CORS: `GET, POST, PUT, DELETE, OPTIONS`;
- разрешённые заголовки для CORS: `Authorization, Content-Type`;
- preflight `OPTIONS` для поддерживаемых frontend-запросов завершается ответом `204 No Content` и не должен доходить до прикладных handler-ов;
- для origin вне `CORS_ALLOWED_ORIGINS` CORS-заголовки не добавляются.

### GET /me

Назначение:
- получение текущего пользователя из access token.

Аутентификация:
- `Authorization: Bearer <jwt>`.

Роли и авторизация:
- `/users`, `POST /questionnaires`, `GET /questionnaires/{id}` и `PUT /questionnaires/{id}` доступны только роли `admin`; для роли `operator` backend возвращает `403 Forbidden`;
- `GET /questionnaires` доступен ролям `operator` и `admin` как read-only список для operator examination flow и admin configuration UI;
- `/specialists`, `/examinations`, `/answers` и `GET /specialists/{id}/examinations` доступны ролям `operator` и `admin`;
- `/me` доступен любой аутентифицированной роли.

Технические детали:
- `/me` использует только access token;
- refresh token в этот endpoint не передаётся.

Ответ `200 OK`:

```json
{
  "id": 1,
  "login": "operator",
  "role": {
    "id": 2,
    "slug": "operator",
    "name": "Operator"
  },
  "is_active": true,
  "created_at": "2026-03-19T12:00:00Z",
  "updated_at": "2026-03-19T12:00:00Z"
}
```

Ошибки:
- `401 Unauthorized` если токен отсутствует или невалиден;
- `403 Forbidden` если пользователь деактивирован.

## Examination Result API

### GET /examinations/{id}/processing-status

Назначение:
- отдать backend-authoritative состояние Phase 2/3 pipeline;
- показать прогресс обязательных каналов и переход в `aggregating`/`aggregated`;
- не опираться на broker management state или клиентские вычисления.

Аутентификация:
- `Authorization: Bearer <jwt>`.

Доступ:
- роли `operator` и `admin`.

Ответ `200 OK`:

```json
{
  "examination_id": 101,
  "status": "aggregating",
  "message_version": 1,
  "channels_total": 3,
  "channels_completed": 3,
  "terminal": false,
  "started_at": "2026-03-22T10:00:00Z",
  "updated_at": "2026-03-22T10:02:00Z",
  "finished_at": null,
  "failed_at": null,
  "channels": [
    {
      "channel": "text",
      "status": "succeeded",
      "attempt_count": 1,
      "max_attempts": 3,
      "message_version": 1,
      "queued_at": "2026-03-22T10:00:00Z",
      "started_at": "2026-03-22T10:00:02Z",
      "finished_at": "2026-03-22T10:00:08Z",
      "last_error_code": null,
      "last_error_message": null,
      "broker_message_id": "msg-text-101-1",
      "broker_correlation_id": "exam-101-text-v1"
    }
  ]
}
```

Правила:
- допустимые `status`: `ready_for_processing`, `processing`, `aggregating`, `aggregated`, `failed`;
- `terminal=false` для `ready_for_processing`, `processing`, `aggregating`;
- `terminal=true` только для `aggregated` и `failed`;
- `channels_completed` считает только каналы со статусом `succeeded`;
- `finished_at` фиксирует момент terminal success `aggregated`;
- `failed_at` фиксирует terminal failure;
- endpoint не раскрывает raw worker payload.

Ошибки:
- `401 Unauthorized` если токен отсутствует или невалиден;
- `403 Forbidden` если роль не имеет доступа;
- `409 Conflict` если обследование ещё не вошло в processing pipeline.

### GET /examinations/{id}/result

Назначение:
- вернуть один канонический агрегированный профиль обследования;
- предоставить frontend и следующим фазам стабильный контракт, не зависящий от raw channel payload;
- быть единственным HTTP-источником итогового результата на Phase 4, включая decision projection.

Аутентификация:
- `Authorization: Bearer <jwt>`.

Доступ:
- роли `operator` и `admin`.

Ответ `200 OK`:

```json
{
  "schema_version": 1,
  "aggregation_version": "agg-v1",
  "examination_id": 101,
  "specialist_id": 55,
  "status": "completed",
  "generated_at": "2026-03-22T10:02:10Z",
  "summary": {
    "overall_score": 0.58,
    "overall_band": "elevated",
    "primary_metric_key": "overall_deviation_index",
    "neutral_recommendation_placeholder": "phase3_pending_external_decision"
  },
  "metrics": [
    {
      "key": "overall_deviation_index",
      "label": "Сводный прокси-индекс",
      "value": 0.58,
      "scale": "0..1",
      "direction": "higher_means_more_deviation"
    }
  ],
  "channel_contributions": [
    {
      "channel": "text",
      "metric_key": "overall_deviation_index",
      "weight": 0.33,
      "contribution": 0.17,
      "evidence_keys": [
        "text_risk_signal"
      ]
    }
  ],
  "channel_reports": [
    {
      "channel": "text",
      "model_version": "rubert-go-emotions-v1",
      "quality_flags": [],
      "evidence": [
        "Нейтральный и достаточно связный ответ без выраженной тревожной окраски."
      ],
      "scores": [
        {
          "key": "text_negativity_score",
          "label": "Негативная окраска текста",
          "value": 0.22
        },
        {
          "key": "text_confidence_score",
          "label": "Уверенность ответа",
          "value": 0.78
        }
      ]
    }
  ],
  "explanations": [
    {
      "position": 1,
      "kind": "summary",
      "text": "Повышение индекса в основном связано с score-метриками acoustic и paralinguistic каналов."
    }
  ],
  "baseline_snapshot": {
    "algorithm_version": "baseline-v1",
    "refreshed_at": "2026-03-22T10:02:09Z",
    "general": {
      "delta": 0.21,
      "band": "mild",
      "baseline_available": true,
      "baseline_source": "general",
      "reference_population_version": "general-v1"
    },
    "personal": {
      "delta": 0.37,
      "band": "moderate",
      "baseline_available": false,
      "baseline_source": "general",
      "baseline_exam_count": 4,
      "update_eligible": false,
      "data_reliability": 0.91,
      "significant_deviations": [
        "overall_deviation_index"
      ]
    }
  },
  "decision": {
    "state": "succeeded",
    "recommendation": "unavailable",
    "message": "analysis_not_implemented_yet",
    "correlation_id": "exam-101-kesmi-1",
    "attempt_count": 1,
    "max_attempts": 2,
    "last_attempt_at": "2026-03-22T10:02:11Z",
    "diagnostics": {
      "error_class": null,
      "error_code": null,
      "error_message": null,
      "http_status": 200,
      "retryable": false
    },
    "raw_response_available": true
  }
}
```

Правила:
- профиль один на обследование;
- контракт versioned через `schema_version`, а реализация агрегации versioned через `aggregation_version`;
- названия метрик должны быть нейтральными и score-oriented до появления финальной ML-семантики;
- `channel_reports` предназначен для operator-facing подробного отчёта и возвращает только explainable score-метрики каналов, без сырых `features`;
- `channel_reports[*].scores[*].value` нормирован в том виде, в каком score был сохранён worker/aggregator слоем; визуальная цветовая интерпретация делается на frontend;
- поле `neutral_recommendation_placeholder` резервирует место под future decision delivery, но не содержит KЭСМИ контракта и не заменяет финальную рекомендацию;
- `status` для этого endpoint допускает `aggregated`, `decision_pending`, `completed`;
- поле `decision` принадлежит backend domain и не совпадает с raw WiMi contract;
- recommendation до появления финальной модели фиксируется как `unavailable`, а `message` фиксируется как `analysis_not_implemented_yet`;
- при terminal transport-провале backend возвращает `decision.state="transport_exhausted"`;
- при terminal business-провале backend возвращает `decision.state="business_error"`;
- raw payload stub-воркеров и raw WiMi payload не являются частью business-контракта и наружу не отдаётся, кроме признака `raw_response_available`.

Ошибки:
- `401 Unauthorized` если токен отсутствует или невалиден;
- `403 Forbidden` если роль не имеет доступа;
- `404 Not Found` если агрегированный профиль ещё не сохранён.

### GET /specialists/{id}/result-history

Назначение:
- отдать историю агрегированных обследований специалиста для трендов и baseline dynamics;
- использовать только канонические агрегированные snapshot-данные.

Аутентификация:
- `Authorization: Bearer <jwt>`.

Доступ:
- роли `operator` и `admin`.

Ответ `200 OK`:

```json
{
  "specialist_id": 55,
  "items": [
    {
      "examination_id": 101,
      "generated_at": "2026-03-22T10:02:10Z",
      "status": "aggregated",
      "summary": {
        "overall_score": 0.58,
        "overall_band": "elevated"
      },
      "baseline_snapshot": {
        "algorithm_version": "baseline-v1",
        "refreshed_at": "2026-03-22T10:02:09Z",
        "general_delta": 0.21,
        "personal_delta": 0.37,
        "baseline_exam_count": 4
      },
      "key_metrics": [
        {
          "key": "overall_deviation_index",
          "label": "Сводный прокси-индекс",
          "value": 0.58,
          "previous_value": 0.46,
          "delta_from_previous": 0.12
        }
      ]
    }
  ]
}
```

Правила:
- endpoint возвращает только обследования со статусом `aggregated`;
- `key_metrics` предназначены для динамики и не заменяют полный result DTO;
- история строится по сохранённым aggregated snapshots, а не по повторному вычислению raw channel payload.

Ошибки:
- `401 Unauthorized` если токен отсутствует или невалиден;
- `403 Forbidden` если роль не имеет доступа;
- `404 Not Found` если специалист не существует.

## Baseline Service Contract

## Decision Delivery Contract

### decision_input

Назначение:
- зафиксировать backend-owned payload, который строится из канонического aggregated profile перед доставкой в WiMi;
- отделить внутренний доменный контракт проекта от raw `POST /ModelCalc` payload.

Контракт:

```json
{
  "schema_version": 1,
  "payload_version": "decision-input-v2",
  "aggregation_version": "agg-v1",
  "examination_id": 101,
  "specialist_id": 55,
  "generated_at": "2026-03-22T10:02:10Z",
  "data_reliability": 0.91,
  "channels": {
    "acoustic": {
      "scores": {
        "acoustic_stress_score": 0.46,
        "voice_stability_score": 0.62,
        "intensity_variability_score": 0.31
      },
      "quality_flags": []
    },
    "text": {
      "scores": {
        "text_negativity_score": 0.28,
        "text_anxiety_score": 0.34,
        "text_confidence_score": 0.72,
        "text_coherence_score": 0.82,
        "text_evasion_score": 0.18
      },
      "quality_flags": []
    },
    "paralinguistic": {
      "scores": {
        "hesitation_score": 0.33,
        "speech_disorganization_score": 0.22
      },
      "quality_flags": []
    }
  },
  "baseline": {
    "source": "general",
    "available": false,
    "baseline_deviation_index": 0.37,
    "significant_deviations": [
      "overall_deviation_index"
    ],
    "z_scores": {
      "overall_deviation_index": 1.5,
      "text_risk_signal": 2.1
    }
  },
  "derived_indicators": {
    "semantic_stress_index": 0.32,
    "acoustic_activation_index": 0.41,
    "speech_disorganization_index": 0.28,
    "baseline_shift_index": 0.39
  },
  "evidence": [
    "Агрегатор передает explainable channel scores и baseline deviations без итогового решения."
  ],
  "service_metadata": {
    "target_system": "kesmi",
    "delivery_mode": "canonical_kesmi_payload",
    "message": "analysis_not_implemented_yet"
  }
}
```

Правила:
- `decision_input` строится из successful mandatory channel payloads, baseline snapshot и backend-derived indicators;
- mandatory channels: `acoustic`, `text`, `paralinguistic`;
- `channels.*.scores` передают explainable channel-level параметры и не заменяются одним общим score;
- `baseline.source` принимает `general` или `personal`, а `baseline.z_scores` содержит per-parameter deviations для КЭСМИ;
- `data_reliability < 1` означает, что анализ выполнен на данных с quality limitations и не эквивалентен “низкому риску”;
- `derived_indicators` вычисляются backend-агрегатором как промежуточные explainable индексы (`semantic_stress_index`, `acoustic_activation_index`, `speech_disorganization_index`, `baseline_shift_index`);
- payload versioned через `payload_version`;
- integration layer маппит этот объект в WiMi/KЭСМИ `incommingParameters: [{id, value}]` по ID параметров конкретной decision-модели, не перекладывая final decision logic в backend.

### KESMI /ModelCalc bridge

`core-backend` отправляет в WiMi:

```json
{
  "modelID": "specialists-model-v2-decision",
  "incommingParameters": [
    { "id": "i1", "value": 0.28 },
    { "id": "i2", "value": 0.34 },
    { "id": "i3", "value": 0.72 },
    { "id": "i4", "value": 0.82 },
    { "id": "i5", "value": 0.18 },
    { "id": "i6", "value": 0.46 },
    { "id": "i7", "value": 0.62 },
    { "id": "i8", "value": 0.31 },
    { "id": "i9", "value": 0.33 },
    { "id": "i10", "value": 0.22 },
    { "id": "i11", "value": 0.0 },
    { "id": "i12", "value": 0.37 },
    { "id": "i13", "value": 1.5 },
    { "id": "i14", "value": 2.1 },
    { "id": "i15", "value": 1.7 },
    { "id": "i16", "value": 1.2 },
    { "id": "i17", "value": -0.4 },
    { "id": "i18", "value": 0.91 },
    { "id": "i19", "value": 0.32 },
    { "id": "i20", "value": 0.41 },
    { "id": "i21", "value": 0.28 },
    { "id": "i22", "value": 0.39 }
  ],
  "outputParameters": ["p32", "p33"],
  "service": {
    "outputFields": [
      "requiredExploredParameters"
    ]
  }
}
```

Правила:
- raw transport body к WiMi остаётся обёрткой вокруг `decision_input`;
- `modelID` должен совпадать с ID, под которым decision-модель загружена в WiMi через `POST /Models`;
- входные параметры передаются по opaque model parameter IDs из XML-модели, а не по человекочитаемым backend field names;
- для текущей модели `specialists_model_v2_decision.xml` backend обязан заполнять все входы `i1..i22`;
- `outputParameters` запрашивают только `final_decision (p32)` и `final_summary (p33)`;
- `service.outputFields` ограничивается только `requiredExploredParameters`, так как `notRequiredExploredParameters` в текущей WiMi-модели не даёт полезного вывода, а диагностические поля `algorithm`, `timing` и подобные остаются отключёнными;
- финальные экспертные правила и recommendation остаются внутри КЭСМИ/WiMi.

### decision_result

Назначение:
- нормализовать итог decision delivery для frontend и operator workflow;
- скрыть raw WiMi response и оставить только backend-owned terminal semantics.

Контракт:

```json
{
  "state": "pending",
  "recommendation": "unavailable",
  "message": "analysis_not_implemented_yet",
  "decision_code": "monitoring",
  "risk_class": "medium",
  "patterns": ["emotional_cross", "contradictory_profile"],
  "correlation_id": "exam-101-kesmi-1",
  "attempt_count": 1,
  "max_attempts": 2,
  "last_attempt_at": "2026-03-22T10:02:11Z",
  "diagnostics": {
    "error_class": "transport",
    "error_code": "pool_busy",
    "error_message": "model pool is busy",
    "http_status": 503,
    "retryable": true
  },
  "raw_response_available": false
}
```

Допустимые значения:
- `state`: `pending`, `succeeded`, `transport_exhausted`, `business_error`;
- `recommendation`: `unavailable`, `allowed`, `risk`, `denied`.

Нормализация ответа WiMi:
- backend читает из `requiredExploredParameters` как минимум `p32=final_decision` и `p33=final_summary`;
- `decision_code` возвращает сырой код WiMi без переименования: `allow`, `monitoring`, `extended_check`, `no_access`;
- `risk_class` извлекается из `final_summary` (`low`, `attention`, `medium`, `high`, `critical`);
- `patterns` извлекается из хвоста `patterns=...` в `final_summary`; `patterns=none` нормализуется в пустой массив;
- coarse `recommendation` сохраняется для совместимости frontend/workflow и маппится так:
  - `allow` -> `allowed`
  - `monitoring` -> `risk`
  - `extended_check` -> `risk`
  - `no_access` -> `denied`
- `message` в terminal successful outcome содержит сырой `final_summary` WiMi, например `risk=medium; decision=monitoring; patterns=emotional_cross;contradictory_profile;`;
- если WiMi вернул `200 OK`, но не прислал корректные `p32/p33`, backend завершает decision delivery как `business_error` с `diagnostics.error_code=invalid_kesmi_response`.

Правила:
- `pending` отражает `decision_pending` в coarse workflow и не является terminal operator outcome;
- `succeeded`, `transport_exhausted` и `business_error` проецируются только при `examinations.status='completed'`;
- `diagnostics` допускает transport и business metadata, но не делает raw external payload частью API;
- `correlation_id` должен сохраняться на каждом attempt и возвращаться оператору для трассировки;
- `raw_response_available=true` означает, что backend сохранил raw response в internal persistence, но не раскрывает его через public API.

Phase 3 baseline остаётся отдельным Python compute-only service boundary.

Правила:
- service не читает PostgreSQL core backend напрямую;
- service не пишет состояние обследования или baseline в PostgreSQL;
- core backend формирует request из канонического aggregated metric vector и history snapshot, затем транзакционно сохраняет response;
- контракт versioned и intentionally narrow.

### POST /baseline/calculate

Назначение:
- вычислить отклонение от общей и персональной baseline-нормы;
- вернуть метаданные алгоритма и признак, допустимо ли обновлять персональную baseline текущим обследованием.

Запрос:

```json
{
  "schema_version": 1,
  "algorithm_version": "baseline-v1",
  "specialist_id": 55,
  "examination_id": 101,
  "generated_at": "2026-03-22T10:02:08Z",
  "metrics": [
    {
      "key": "overall_deviation_index",
      "value": 0.58
    },
    {
      "key": "speech_stability_score",
      "value": 0.41
    }
  ],
  "history": {
    "baseline_exam_count": 4,
    "metric_vectors": [
      {
        "examination_id": 91,
        "generated_at": "2026-03-20T10:02:08Z",
        "metrics": [
          {
            "key": "overall_deviation_index",
            "value": 0.46
          }
        ]
      }
    ]
  },
  "existing_baseline": {
    "baseline_available": false,
    "metrics": {}
  },
  "context": {
    "all_channels_done": true,
    "critical_quality_flags": [],
    "data_reliability": 0.91,
    "overall_band": "mild"
  },
  "general_reference_population_version": "general-v1"
}
```

Ответ `200 OK`:

```json
{
  "schema_version": 1,
  "algorithm_version": "baseline-v1",
  "refreshed_at": "2026-03-22T10:02:09Z",
  "general_deviation": {
    "score": 0.21,
    "band": "mild",
    "baseline_available": true,
    "baseline_source": "general",
    "significant_deviations": [],
    "metric_scores": {
      "overall_deviation_index": {
        "baseline_available": true,
        "baseline_source": "general",
        "baseline_mean": 0.26,
        "baseline_std": 0.08,
        "sample_count": 0,
        "method": "general_reference",
        "delta": 0.08,
        "z_score": 1.0,
        "deviation_level": "mild"
      }
    }
  },
  "personal_deviation": {
    "score": 0.37,
    "band": "moderate",
    "baseline_available": false,
    "baseline_source": "general",
    "significant_deviations": [
      "overall_deviation_index"
    ],
    "metric_scores": {
      "overall_deviation_index": {
        "baseline_available": false,
        "baseline_source": "general",
        "baseline_mean": 0.26,
        "baseline_std": 0.08,
        "sample_count": 0,
        "method": "general_reference",
        "delta": 0.12,
        "z_score": 1.5,
        "deviation_level": "mild"
      }
    }
  },
  "update_eligibility": {
    "eligible": true,
    "reason": "accepted",
    "baseline_exam_count_after_update": 5,
    "data_reliability": 0.91
  },
  "next_baseline": {
    "exam_count": 5,
    "baseline_available": true,
    "metrics": {
      "overall_deviation_index": {
        "baseline_mean": 0.48,
        "baseline_std": 0.05,
        "sample_count": 5,
        "last_updated_at": "2026-03-22T10:02:09Z",
        "method": "ewma_bootstrap"
      }
    },
    "refreshed_at": "2026-03-22T10:02:09Z"
  }
}
```

Ошибки:
- `400 Bad Request` если request schema невалидна;
- `422 Unprocessable Entity` если обязательные metric keys отсутствуют;
- `503 Service Unavailable` если baseline service временно недоступен.

Правила:
- `general_deviation` и `personal_deviation` обязательны даже если history пустая;
- `metric_scores` внутри обоих deviation-блоков содержат per-metric baseline reference, `delta`, `z_score` и `deviation_level` для канонических score-метрик;
- `algorithm_version` и `refreshed_at` обязательны для persistence snapshot в core backend;
- `existing_baseline` передаётся только как backend-owned persisted state; baseline service не получает raw DB access;
- `context` передаёт eligibility-факторы уровня workflow: готовность каналов, критичные `quality_flags`, `data_reliability`, coarse `overall_band`;
- `update_eligibility` описывает только предварительную допустимость baseline refresh на уровне compute-сервиса; финальное persisted обновление personal baseline происходит в core backend только после завершения decision-stage;
- `next_baseline` обязателен и содержит кандидатный snapshot personal baseline на основе EWMA mean/std и `sample_count`, без немедленной записи в PostgreSQL;
- personal baseline считается доступным только после `BASELINE_MIN_PERSONAL_EXAMS` качественных обследований, до этого `personal_deviation` использует `baseline_source="general"`;
- до созревания personal baseline (`baseline_available=false`) general fallback не должен доминировать над реальными каналами: в decision payload для КЭСМИ backend передаёт `baseline_deviation_index=0`, `baseline_shift_index=0` и обнуляет baseline `z_scores`, оставляя baseline только как operator-facing справочный слой;
- history для bootstrap personal baseline формируется только из baseline snapshot-ов, которые прошли финальный eligibility gate и были приняты в persisted baseline history.
- клинические выводы и KЭСМИ-поля в ответ baseline service не включаются.

### POST /users

Назначение:
- создание нового пользователя;
- минимальное управление пользователями в рамках синхронного этапа.

Аутентификация:
- `Authorization: Bearer <jwt>`.

Запрос:

```json
{
  "login": "operator2",
  "password": "secret",
  "role": "operator"
}
```

Поля:
- `login`: обязательное, уникальное;
- `password`: обязательное;
- `role`: необязательное, по умолчанию `operator`, допустимые значения на текущем этапе: `admin`, `operator`.

Ответ `201 Created`:

```json
{
  "id": 2,
  "login": "operator2",
  "role": {
    "id": 2,
    "slug": "operator",
    "name": "Operator"
  },
  "is_active": true,
  "created_at": "2026-03-19T12:10:00Z",
  "updated_at": "2026-03-19T12:10:00Z"
}
```

Ошибки:
- `400 Bad Request` если payload невалидный;
- `401 Unauthorized` если токен отсутствует или невалиден;
- `404 Not Found` если указана несуществующая роль;
- `409 Conflict` если `login` уже занят.

### GET /users

Назначение:
- получение списка пользователей для административного интерфейса.

Аутентификация:
- `Authorization: Bearer <jwt>`.

Ответ `200 OK`:

```json
{
  "items": [
    {
      "id": 1,
      "login": "operator",
      "role": {
        "id": 2,
        "slug": "operator",
        "name": "Operator"
      },
      "is_active": true,
      "last_login_at": "2026-03-25T08:15:00Z",
      "created_at": "2026-03-19T12:00:00Z",
      "updated_at": "2026-03-19T12:00:00Z"
    }
  ]
}
```

Дополнительно:
- `last_login_at` может быть `null`, если пользователь ещё ни разу не входил после появления этой метрики.

### GET /users/{id}

Назначение:
- получение карточки пользователя.

Аутентификация:
- `Authorization: Bearer <jwt>`.

Ответ `200 OK`: объект пользователя в формате `/me`.

Ошибки:
- `404 Not Found` если пользователь не найден.

### PUT /users/{id}

Назначение:
- обновление логина, роли, статуса активности пользователя и, при необходимости, его пароля.

Аутентификация:
- `Authorization: Bearer <jwt>`.

Запрос:

```json
{
  "login": "operator2",
  "password": "new-secret",
  "role": "operator",
  "is_active": true
}
```

Поля:
- `login`: обязательное, уникальное;
- `password`: необязательное; если передано непустым, заменяет текущий пароль пользователя;
- `role`: обязательное, допустимые значения текущего этапа: `admin`, `operator`;
- `is_active`: обязательное.

Ответ `200 OK`: обновлённый объект пользователя.

Ошибки:
- `400 Bad Request` если payload невалидный;
- `404 Not Found` если пользователь или роль не найдены;
- `409 Conflict` если `login` уже занят.

## Specialists API

Объект `Specialist`:

```json
{
  "id": 10,
  "full_name": "Иванов Иван Иванович",
  "personnel_number": "A-123",
  "examinations_count": 6,
  "last_examination_id": 42,
  "last_examination_at": "2026-03-25T07:40:00Z",
  "last_examination_status": "completed",
  "last_overall_score": 0.74,
  "last_overall_band": "elevated",
  "baseline_exam_count": 4,
  "baseline_refreshed_at": "2026-03-24T12:00:00Z",
  "created_at": "2026-03-19T12:00:00Z",
  "updated_at": "2026-03-19T12:00:00Z"
}
```

Поля registry summary:
- `examinations_count`: общее число обследований специалиста;
- `last_examination_id`: идентификатор последнего обследования или `null`, если обследований ещё не было;
- `last_examination_at`: timestamp последнего обследования (`finished_at`, иначе `started_at`, иначе `created_at`) или `null`;
- `last_examination_status`: текущий coarse status последнего обследования или `null`;
- `last_overall_score`: итоговый агрегированный score последнего обследования, если aggregated profile уже сохранён;
- `last_overall_band`: итоговая textual band последнего обследования, если aggregated profile уже сохранён;
- `baseline_exam_count`: текущее количество обследований, вошедших в persisted baseline state;
- `baseline_refreshed_at`: время последнего обновления baseline state или `null`.

Все endpoints раздела защищены Bearer JWT.

### POST /specialists

Запрос:

```json
{
  "full_name": "Иванов Иван Иванович",
  "personnel_number": "A-123"
}
```

Ответ `201 Created`: объект `Specialist`.

Ошибки:
- `400 Bad Request` если `full_name` пустой;
- `409 Conflict` при конфликте `personnel_number`.

### GET /specialists

Ответ `200 OK`:

```json
{
  "items": [
    {
      "id": 10,
      "full_name": "Иванов Иван Иванович",
      "personnel_number": "A-123",
      "examinations_count": 6,
      "last_examination_id": 42,
      "last_examination_at": "2026-03-25T07:40:00Z",
      "last_examination_status": "completed",
      "last_overall_score": 0.74,
      "last_overall_band": "elevated",
      "baseline_exam_count": 4,
      "baseline_refreshed_at": "2026-03-24T12:00:00Z",
      "created_at": "2026-03-19T12:00:00Z",
      "updated_at": "2026-03-19T12:00:00Z"
    }
  ]
}
```

### GET /specialists/{id}

Ответ `200 OK`: объект `Specialist`.

Ошибки:
- `404 Not Found` если специалист не найден.

### PUT /specialists/{id}

Запрос:

```json
{
  "full_name": "Иванов Иван Иванович",
  "personnel_number": "A-124"
}
```

Ответ `200 OK`: обновлённый объект `Specialist`.

Ошибки:
- `400 Bad Request` если payload невалидный;
- `404 Not Found` если специалист не найден;
- `409 Conflict` при конфликте `personnel_number`.

### DELETE /specialists/{id}

Ответ `204 No Content`.

Ошибки:
- `404 Not Found` если специалист не найден;
- `409 Conflict` если специалист уже связан с обследованиями.

## Examinations API

Объект `Examination`:

```json
{
  "id": 100,
  "specialist_id": 10,
  "created_by_user_id": 1,
  "questionnaire_id": 5,
  "status": "created",
  "created_at": "2026-03-19T12:00:00Z",
  "started_at": null,
  "finished_at": null,
  "updated_at": "2026-03-19T12:00:00Z"
}
```

Все endpoints раздела защищены Bearer JWT.

### POST /examinations

Запрос:

```json
{
  "specialist_id": 10,
  "questionnaire_id": 5
}
```

Поля:
- `specialist_id`: обязательное;
- `questionnaire_id`: необязательное. Если передано, должно ссылаться на существующий опросник. Поле добавлено минимально для operator flow, чтобы обследование могло быть связано с выбранным опросом.

Ответ `201 Created`: объект `Examination` в статусе `created`.

Ошибки:
- `400 Bad Request` если `specialist_id` невалиден;
- `404 Not Found` если специалист или опросник не найдены.

### GET /examinations

Назначение:
- получение общего списка обследований.

Ответ `200 OK`:

```json
{
  "items": [
    {
      "id": 100,
      "specialist_id": 10,
      "created_by_user_id": 1,
      "questionnaire_id": 5,
      "status": "created",
      "created_at": "2026-03-19T12:00:00Z",
      "started_at": null,
      "finished_at": null,
      "updated_at": "2026-03-19T12:00:00Z"
    }
  ]
}
```

### GET /examinations/{id}

Назначение:
- получение одного обследования по идентификатору.

Ответ `200 OK`: объект `Examination` с `questions` snapshot-массивом для operator recording flow.

```json
{
  "id": 100,
  "specialist_id": 10,
  "created_by_user_id": 1,
  "questionnaire_id": 5,
  "status": "collecting_answers",
  "questions": [
    {
      "id": 9001,
      "examination_id": 100,
      "specialist_id": 10,
      "questionnaire_id": 5,
      "source_question_id": 11,
      "position": 1,
      "text": "Как вы себя чувствуете сегодня?"
    }
  ],
  "created_at": "2026-03-19T12:00:00Z",
  "started_at": "2026-03-19T12:01:00Z",
  "finished_at": null,
  "updated_at": "2026-03-19T12:01:00Z"
}
```

Правила:
- `questions` содержит snapshot-вопросы из `examination_questions`, а не live-версию из `questionnaires`;
- порядок элементов в `questions` совпадает с `position`;
- operator UI обязан использовать именно `questions[*].id` как `examination_question_id` при `POST /answers`.

Ошибки:
- `404 Not Found` если обследование не найдено.

### GET /specialists/{id}/examinations

Назначение:
- получение истории обследований конкретного специалиста.

Ответ `200 OK`:

```json
{
  "items": [
    {
      "id": 100,
      "specialist_id": 10,
      "created_by_user_id": 1,
      "questionnaire_id": 5,
      "status": "ready_for_processing",
      "created_at": "2026-03-19T12:00:00Z",
      "started_at": "2026-03-19T12:01:00Z",
      "finished_at": "2026-03-19T12:05:00Z",
      "updated_at": "2026-03-19T12:05:00Z"
    }
  ]
}
```

Технические детали:
- список должен отражать авторитетные backend-статусы `created`, `collecting_answers`, `ready_for_processing`, `processing`, `failed`;
- UI не должен вычислять эти статусы из local draft state, cookie-кэша или optimistic флагов.

Ошибки:
- `404 Not Found` если специалист не найден.

### POST /examinations/{id}/start

Назначение:
- перевод обследования в статус `collecting_answers`.

Ответ `200 OK`: объект `Examination`.

Правила:
- переход допустим из `created`;
- если обследование уже в `collecting_answers`, endpoint идемпотентно возвращает текущее состояние;
- любые другие переходы дают ошибку.

Ошибки:
- `404 Not Found` если обследование не найдено;
- `409 Conflict` при недопустимом переходе статуса.

### POST /examinations/{id}/finish

Назначение:
- перевод обследования в статус `ready_for_processing` и атомарная фиксация намерения на запуск асинхронной обработки.

Ответ `200 OK`: объект `Examination`.

Правила:
- переход допустим из `collecting_answers`;
- если обследование уже в `ready_for_processing`, endpoint идемпотентно возвращает текущее состояние;
- backend переводит обследование в `ready_for_processing` только если число сохранённых ответов совпадает с числом snapshot-вопросов `examination_questions`;
- `POST /examinations/{id}/finish` в той же PostgreSQL-транзакции обязан:
  - создать или переиспользовать launch fence;
  - создать по одной записи `examination_channel_runs` для обязательных каналов `text`, `acoustic`, `paralinguistic`;
  - создать по одной pending-записи в `processing_outbox` для публикации команд в RabbitMQ;
- processing launch защищён server-side fence в БД, поэтому повторный `finish` не должен создавать дублирующий запуск или повторные channel-run/outbox записи;
- любые другие переходы дают ошибку.

Ошибки:
- `404 Not Found` если обследование не найдено;
- `409 Conflict` при недопустимом переходе статуса или неполном наборе ответов.

### GET /examinations/{id}/processing-status

Назначение:
- получение backend-authoritative состояния конвейера обработки по обследованию;
- отображение progress/failure по каналам без обращения к RabbitMQ из UI.

Ответ `200 OK`:

```json
{
  "examination_id": 100,
  "status": "processing",
  "message_version": 1,
  "channels_total": 3,
  "channels_completed": 1,
  "terminal": false,
  "started_at": "2026-03-21T09:00:00Z",
  "updated_at": "2026-03-21T09:01:00Z",
  "finished_at": null,
  "failed_at": null,
  "channels": [
    {
      "channel": "text",
      "status": "succeeded",
      "attempt_count": 1,
      "max_attempts": 3,
      "message_version": 1,
      "queued_at": "2026-03-21T09:00:01Z",
      "started_at": "2026-03-21T09:00:03Z",
      "finished_at": "2026-03-21T09:00:10Z",
      "last_error_code": null,
      "last_error_message": null,
      "broker_message_id": "msg-100-text-1",
      "broker_correlation_id": "exam-100-text-v1"
    },
    {
      "channel": "acoustic",
      "status": "processing",
      "attempt_count": 1,
      "max_attempts": 3,
      "message_version": 1,
      "queued_at": "2026-03-21T09:00:01Z",
      "started_at": "2026-03-21T09:00:05Z",
      "finished_at": null,
      "last_error_code": null,
      "last_error_message": null,
      "broker_message_id": "msg-100-acoustic-1",
      "broker_correlation_id": "exam-100-acoustic-v1"
    },
    {
      "channel": "paralinguistic",
      "status": "queued",
      "attempt_count": 0,
      "max_attempts": 3,
      "message_version": 1,
      "queued_at": "2026-03-21T09:00:01Z",
      "started_at": null,
      "finished_at": null,
      "last_error_code": null,
      "last_error_message": null,
      "broker_message_id": null,
      "broker_correlation_id": "exam-100-paralinguistic-v1"
    }
  ]
}
```

Правила:
- endpoint возвращает состояние, вычисленное из PostgreSQL (`examinations`, `examination_channel_runs`, `processing_outbox`, `channel_results`), а не из состояния очередей RabbitMQ;
- общее поле `status` допускает значения `ready_for_processing`, `processing`, `aggregating`, `aggregated`, `decision_pending`, `completed`, `failed`;
- детальное состояние каналов живёт только в `channels[*]` и не должно дублироваться в основном объекте `Examination`;
- после завершения обязательных каналов endpoint продолжает оставаться доступным и для post-processing состояний `aggregated`, `decision_pending`, `completed`;
- `terminal=true` означает только terminal coarse outcome `completed` или `failed`;
- `last_error_message` предназначено для операторской диагностики и не должно содержать stack trace или чувствительные данные.

Ошибки:
- `404 Not Found` если обследование не найдено;
- `409 Conflict` если обследование ещё не было переведено в `ready_for_processing` и не входило в processing/post-processing pipeline.

## Asynchronous Processing Contract

### Mandatory channels

Обязательные каналы Phase 2:
- `text`
- `acoustic`
- `paralinguistic`

### Versioning

- `message_version` обязателен во всех processing command/result envelopes;
- начальная версия контрактов Phase 2: `1`;
- несовместимые изменения envelope shape требуют увеличения `message_version` и обновления этого документа до изменения кода.

### AMQP topology

Транспорт:
- RabbitMQ `3.13.x`;
- topology должна быть объявлена core backend и/или инициализацией инфраструктуры до запуска workers;
- PostgreSQL остаётся source of truth для progress, retry ledger и terminal failures.

Exchanges:
- `processing.commands` (`topic`, durable) — публикация команд на обработку;
- `processing.results` (`topic`, durable) — публикация унифицированных результатов каналов.

Routing keys:
- команды:
  - `processing.command.text`
  - `processing.command.acoustic`
  - `processing.command.paralinguistic`
- результаты:
  - `processing.result`

Queues:
- `qq.processing.text` — binding `processing.command.text`
- `qq.processing.acoustic` — binding `processing.command.acoustic`
- `qq.processing.paralinguistic` — binding `processing.command.paralinguistic`
- `qq.processing.results` — binding `processing.result`

Retry/DLX assumptions для RabbitMQ `3.13.x`:
- очереди команд должны быть durable; предпочтительный тип для bounded retries: quorum queue;
- так как в RabbitMQ `3.13.x` нет безопасного default `delivery-limit`, система обязана явно задавать retry policy и DLX behavior для очередей команд;
- core backend relay объявляет `processing.commands.dlx` и задаёт для каждой command queue аргументы `x-queue-type=quorum`, `x-dead-letter-exchange=processing.commands.dlx`, `x-delivery-limit=<PROCESSING_OUTBOX_MAX_ATTEMPTS>`;
- повторные доставки ограничиваются `max_attempts`, хранимым в PostgreSQL и отражаемым в DTO;
- DLX/poison-message semantics используются только как транспортный механизм, но не как источник истины для UI;
- результаты всех каналов публикуются в единый exchange/queue `processing.results` / `qq.processing.results` с единой envelope shape.

Local Compose topology Phase 2:
- сервисы `text-worker`, `acoustic-worker`, `paralinguistic-worker` стартуют в одном `docker-compose.yml` вместе с `frontend`, `core-backend`, `postgres`, `rabbitmq`, `minio`;
- каждый worker читает только свою очередь через env `WORKER_QUEUE_NAME` и bind-ит её к `WORKER_COMMAND_EXCHANGE` + своему routing key;
- все workers публикуют результат только в `PROCESSING_RESULT_EXCHANGE=processing.results` с routing key `PROCESSING_RESULT_ROUTING_KEY=processing.result`;
- workers используют только S3 references из command envelope и runtime env `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY_ID`, `MINIO_SECRET_ACCESS_KEY`, `MINIO_USE_SSL`;
- локальный compose не меняет версию broker: используется `rabbitmq:3.13-management-alpine`, поэтому retry/DLX policy должна оставаться явной и совместимой с RabbitMQ `3.13.x`.

Worker runtime env:
- общие: `RABBITMQ_HOST`, `RABBITMQ_PORT`, `RABBITMQ_VHOST`, `RABBITMQ_USER`, `RABBITMQ_PASSWORD`, `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY_ID`, `MINIO_SECRET_ACCESS_KEY`, `MINIO_USE_SSL`, `WORKER_PREFETCH_COUNT`;
- `text-worker`: `TEXT_WORKER_QUEUE_NAME=qq.processing.text`, `TEXT_WORKER_COMMAND_ROUTING_KEY=processing.command.text`;
- `acoustic-worker`: `ACOUSTIC_WORKER_QUEUE_NAME=qq.processing.acoustic`, `ACOUSTIC_WORKER_COMMAND_ROUTING_KEY=processing.command.acoustic`;
- `paralinguistic-worker`: `PARALINGUISTIC_WORKER_QUEUE_NAME=qq.processing.paralinguistic`, `PARALINGUISTIC_WORKER_COMMAND_ROUTING_KEY=processing.command.paralinguistic`.

### Processing command envelope v1

Назначение:
- одна публикация на один `examination_channel_run`;
- payload должен ссылаться на S3-объекты, а не содержать бинарное аудио.

```json
{
  "message_version": 1,
  "message_id": "6f8664a1-12f8-4fd4-817d-3784a55296f7",
  "correlation_id": "exam-100-text-v1",
  "examination_id": 100,
  "specialist_id": 10,
  "channel": "text",
  "attempt": 1,
  "max_attempts": 3,
  "requested_at": "2026-03-21T09:00:01Z",
  "answers": [
    {
      "answer_id": 500,
      "question_id": 9001,
      "audio_s3_bucket": "diplom-audio",
      "audio_s3_key": "examinations/100/answers/500/audio.webm",
      "answer_text": ""
    }
  ]
}
```

Поля:
- `message_version`: обязательное целое;
- `message_id`: уникальный идентификатор сообщения для transport-level traceability;
- `correlation_id`: общий идентификатор попытки обработки обследования/канала;
- `examination_id`, `specialist_id`: обязательные доменные идентификаторы;
- `channel`: одно из обязательных значений `text`, `acoustic`, `paralinguistic`;
- `attempt`: номер попытки публикации/обработки, начиная с `1`;
- `max_attempts`: максимальное число попыток для данного канала;
- `requested_at`: время формирования команды в core backend;
- `answers[*].audio_s3_bucket` и `answers[*].audio_s3_key`: обязательные ссылки на объект в S3;
- `answers[*].answer_text`: строковое поле общего контракта; в текущем audio-only operator flow может быть пустой строкой до появления отдельной транскрипции;
- передача бинарных audio bytes в RabbitMQ запрещена.

### Channel result envelope v1

Назначение:
- единый формат для `text`, `acoustic`, `paralinguistic`;
- core backend принимает result envelope channel-neutral способом и сам обновляет PostgreSQL state machine.

```json
{
  "message_version": 1,
  "message_id": "544147f4-5d1f-4c45-8b1f-baa5a638969c",
  "correlation_id": "exam-100-text-v1",
  "examination_id": 100,
  "channel": "text",
  "attempt": 1,
  "status": "succeeded",
  "completed_at": "2026-03-21T09:00:10Z",
  "model_version": "rubert-go-emotions-v1",
  "error_code": null,
  "error_message": null,
  "payload": {
    "channel": "text",
    "status": "done",
    "examination_id": 100,
    "answer_id": null,
    "features": {},
    "scores": {},
    "quality_flags": [],
    "evidence": [],
    "model_version": "rubert-go-emotions-v1",
    "processing_time_ms": 120,
    "error": null
  }
}
```

Поля:
- `status`: одно из `succeeded`, `temporary_error`, `fatal_error`;
- `payload`: канонический channel result payload для `succeeded`; при transport/contract ошибках может быть `{}`;
- `error_code` и `error_message` обязательны при `temporary_error` и `fatal_error`, должны быть пустыми при `succeeded`;
- `completed_at` обязателен для всех terminal result-сообщений;
- worker не должен публиковать разные envelope shapes для разных каналов.
- worker обязан заполнять `model_version` и использовать `temporary_error` для transport/S3 availability failures, `fatal_error` для невалидного command payload или отсутствующего S3 object reference.

#### Canonical channel payload v1

Все успешные channel results сохраняются в PostgreSQL в едином payload-формате:

```json
{
  "channel": "acoustic",
  "status": "done",
  "examination_id": 100,
  "answer_id": null,
  "features": {},
  "scores": {},
  "quality_flags": [],
  "evidence": [],
  "model_version": "acoustic-librosa-v1",
  "processing_time_ms": 90,
  "error": null
}
```

Правила:
- `channel`: `text`, `acoustic` или `paralinguistic`;
- `status`: `done` или `failed` внутри payload; transport-level статус остаётся в result envelope;
- `examination_id`: id обследования из command envelope;
- `answer_id`: id ответа, если payload относится к одному ответу; `null`, если worker агрегировал несколько audio answers в один channel result;
- `features`: измеренные признаки канала, без итоговых решений;
- `scores`: нормированные/производные оценки канала `0..1`, без решения о допуске;
- `quality_flags`: технические флаги качества данных/обработки;
- `evidence`: человекочитаемые факты для оператора и отладки;
- `model_version`: версия модели/алгоритма канала, обязательна;
- `processing_time_ms`: wall-clock время обработки worker payload;
- `error`: `null` для `done`; строка/объект ошибки для `failed`.

Core backend валидирует наличие этих полей перед сохранением successful channel result. Старые stub/proxy поля (`summary`, `text_total_characters`, `audio_energy_proxy`, `speech_rate_proxy` и аналоги) больше не являются контрактом.

#### Paralinguistic payload v1

`paralinguistic-worker` больше не возвращает stub payload. Worker декодирует каждый audio object через local `ffmpeg` в mono PCM 16 kHz, выделяет speech-сегменты lightweight WebRTC VAD и возвращает измеримые признаки речевого поведения. Ошибки декодирования/VAD не должны падать из worker process: если S3 object доступен, но аудио не анализируется, result envelope остаётся `status=succeeded`, а payload получает `quality_flags=["vad_failed"]` и evidence с причиной. Transport/S3 ошибки по-прежнему идут через `temporary_error`/`fatal_error` envelope.

Минимальный payload:

```json
{
  "channel": "paralinguistic",
  "status": "done",
  "examination_id": 100,
  "answer_id": null,
  "features": {
    "total_audio_duration_ms": 10000,
    "speech_duration_ms": 5700,
    "silence_duration_ms": 4300,
    "speech_ratio": 0.57,
    "response_delay_ms": 500,
    "pause_count": 2,
    "long_pause_count": 1,
    "mean_pause_ms": 900,
    "max_pause_ms": 1200,
    "speech_segment_count": 3,
    "speech_rate_wpm": null
  },
  "scores": {
    "hesitation_score": 0.31,
    "speech_disorganization_score": 0.2
  },
  "quality_flags": [],
  "evidence": [
    "Речь занимает 57.0% суммарного аудио.",
    "Обнаружено пауз между речевыми сегментами: 2.",
    "Длинных пауз: 1."
  ],
  "model_version": "paralinguistic-vad-v1",
  "processing_time_ms": 70,
  "error": null,
  "answers": [
    {
      "answer_id": 500,
      "question_id": 9001,
      "audio_s3_key": "examinations/100/answers/500/audio.webm",
      "bytes": 524288,
      "result": {
        "channel": "paralinguistic",
        "status": "done",
        "features": {},
        "scores": {},
        "quality_flags": [],
        "evidence": [],
        "model_version": "paralinguistic-vad-v1"
      }
    }
  ]
}
```

Поля:
- `features.total_audio_duration_ms`: суммарная длительность обработанных аудиоответов после декодирования;
- `features.speech_duration_ms`: суммарная длительность VAD speech-сегментов;
- `features.silence_duration_ms`: `total_audio_duration_ms - speech_duration_ms`;
- `features.speech_ratio`: доля речи в аудио, `0..1`;
- `features.response_delay_ms`: средняя задержка от начала ответа до первого speech-сегмента по обработанным ответам;
- `features.pause_count`: количество пауз между speech-сегментами, паузы между разными audio objects не склеиваются;
- `features.long_pause_count`: количество пауз с длительностью не меньше `PARALINGUISTIC_LONG_PAUSE_THRESHOLD_MS`, default `1200`;
- `features.mean_pause_ms`, `features.max_pause_ms`: средняя и максимальная длительность пауз;
- `features.speech_segment_count`: количество speech-сегментов после VAD merge;
- `features.speech_rate_wpm`: слов в минуту, если worker получил `answer_text`/transcript; иначе `null`;
- `scores.hesitation_score`: нормированный `0..1` индекс на основе pause ratio, long pauses и response delay;
- `scores.speech_disorganization_score`: нормированный `0..1` индекс фрагментации речи и низкой доли речи;
- `quality_flags`: набор из `no_speech_detected`, `too_short_speech`, `low_speech_ratio`, `vad_failed`;
- `evidence`: человекочитаемые факты для operator-facing explanation;
- `answers`: per-answer diagnostic breakdown без audio bytes и без transcript leakage.

Runtime config:
- `PARALINGUISTIC_LONG_PAUSE_THRESHOLD_MS`, default `1200`;
- `PARALINGUISTIC_MIN_SPEECH_DURATION_MS`, default `700`;
- `PARALINGUISTIC_LOW_SPEECH_RATIO_THRESHOLD`, default `0.15`;
- `PARALINGUISTIC_VAD_AGGRESSIVENESS`, default `2`, допустимые значения WebRTC VAD `0..3`.

Compatibility proxy-поля `speech_rate_proxy` и `prosody_variation_proxy` удалены. Aggregator читает `scores.hesitation_score` и `scores.speech_disorganization_score`.

#### Acoustic payload v1

`acoustic-worker` больше не возвращает stub payload. Worker декодирует каждый audio object через local `ffmpeg` в mono float32 PCM 22.05 kHz и извлекает explainable audio features через `librosa`/`numpy` без обучения большой модели. Ошибки декодирования или feature extraction не должны падать из worker process: если S3 object доступен, result envelope остаётся `status=succeeded`, а payload получает `quality_flags=["feature_extraction_failed"]` и evidence с причиной. Transport/S3 ошибки по-прежнему идут через `temporary_error`/`fatal_error` envelope.

Минимальный payload:

```json
{
  "channel": "acoustic",
  "status": "done",
  "examination_id": 100,
  "answer_id": null,
  "features": {
    "duration_ms": 2000,
    "sample_rate": 16000,
    "rms_energy_mean": 0.12
  },
  "scores": {
    "acoustic_stress_score": 0.41,
    "voice_stability_score": 0.63,
    "intensity_variability_score": 0.18
  },
  "emotion_probs": {
    "neutral": 0.24,
    "happiness": 0.08,
    "sadness": 0.16,
    "anger": 0.44,
    "fear": 0.0,
    "other": 0.08
  },
  "dominant_emotion": "anger",
  "quality_flags": [],
  "evidence": [
    "Доминирующая эмоция: anger (44.0%).",
    "Средняя энергия RMS: 0.12.",
    "Индекс акустического напряжения: 0.41."
  ],
  "model_version": "xbgoose-hubert-large-dusha-v1",
  "processing_time_ms": 90,
  "error": null,
  "answers": [
    {
      "answer_id": 500,
      "question_id": 9001,
      "audio_s3_key": "examinations/100/answers/500/audio.webm",
      "bytes": 524288,
      "result": {
        "channel": "acoustic",
        "status": "done",
        "features": {},
        "scores": {},
        "emotion_probs": {},
        "dominant_emotion": "neutral",
        "quality_flags": [],
        "evidence": [],
        "model_version": "xbgoose-hubert-large-dusha-v1"
      }
    }
  ]
}
```

Поля:
- `features.duration_ms`: длительность декодированного аудио после нормализации в mono `16 kHz`;
- `features.sample_rate`: sample rate, на котором модель получает аудио; в Phase 4 acoustic worker это `16000`;
- `features.rms_energy_mean`: вспомогательная low-level метрика для quality gating и operator-facing evidence;
- `emotion_probs`: вероятности классов SER-модели, нормализованные к ключам `neutral`, `happiness`, `sadness`, `anger`, `fear`, `other`; если исходный label модели не попадает в известный mapping, он агрегируется в `other`;
- `dominant_emotion`: класс с максимальной вероятностью после нормализации label mapping;
- `scores.voice_stability_score`: нормированный `0..1`, где `1` означает более стабильный голос; в MVP считается эвристически из SER negative signal, neutral share и intensity variability;
- `scores.acoustic_stress_score`: основной acoustic risk/stress score `0..1`, вычисляемый из SER emotion probabilities c небольшим вкладом intensity variability;
- `scores.intensity_variability_score`: вспомогательный совместимый score на основе RMS variability; сохраняется в payload для текущего aggregator/KЭСМИ mapping и не является основным acoustic inference-каналом;
- `quality_flags`: набор из `audio_too_short`, `low_volume`, `model_load_failed`, `inference_failed`, `unsupported_audio_format`;
- `evidence`: человекочитаемые факты для operator-facing explanation;
- `answers`: per-answer diagnostic breakdown без audio bytes.

Runtime env:
- `ACOUSTIC_MODEL_PATH`: локальная директория модели, default `/app/models/acoustic-emotion`;
- `ACOUSTIC_MODEL_ID`: fallback Hugging Face model id, default `xbgoose/hubert-large-speech-emotion-recognition-russian-dusha-finetuned`;
- `ACOUSTIC_DEVICE`: `cpu` или `cuda`;
- `HF_HUB_OFFLINE`: если `1`, acoustic worker обязан грузить модель только из `ACOUSTIC_MODEL_PATH`.

Compatibility proxy-поля `audio_total_bytes`, `audio_average_bytes` и `audio_energy_proxy` удалены. Aggregator читает `scores.acoustic_stress_score`, `scores.voice_stability_score` и `scores.intensity_variability_score`.

#### Text analysis payload v1

`text-worker` выполняет локальную STT-транскрибацию через `faster-whisper`, а затем анализирует transcript внутри того же worker-а. Для emotion probabilities используется локальная RuBERT-compatible multi-label emotion-модель, default `seara/rubert-base-cased-russian-emotion-detection-ru-go-emotions`; модель грузится лениво при первом непустом transcript. Выходы исходной go-emotions taxonomy нормализуются к backend-owned ключам `joy`, `sadness`, `anger`, `fear`, `surprise`, `neutral`, чтобы downstream aggregator, КЭСМИ и operator UI не зависели от конкретной модели. Health/readiness worker-а не требуют загрузки STT или emotion-модели. Если STT или emotion-модель не сработали, worker не должен падать: result envelope остаётся `status=succeeded`, payload получает quality flag (`stt_failed` или `emotion_model_failed`) и сохраняет explainable heuristic features/scores там, где это возможно.

Минимальный payload:

```json
{
  "channel": "text",
  "status": "done",
  "examination_id": 100,
  "answer_id": null,
  "stt": {
    "transcript": "текст ответа",
    "language": "ru",
    "segments": [
      {
        "start_ms": 0,
        "end_ms": 1200,
        "text": "текст ответа",
        "answer_id": 500,
        "question_id": 9001
      }
    ],
    "duration_ms": 1200,
    "word_count": 2,
    "model_version": "faster-whisper-medium"
  },
  "features": {
    "transcript_length_chars": 11,
    "word_count": 2,
    "sentence_count": 1,
    "avg_sentence_length": 2.0,
    "uncertainty_marker_count": 0,
    "negation_count": 0,
    "short_answer_flag": true
  },
  "scores": {
    "text_negativity_score": 0.1,
    "text_anxiety_score": 0.2,
    "text_confidence_score": 0.8,
    "text_coherence_score": 0.7,
    "text_evasion_score": 0.45
  },
  "emotion_probs": {
    "joy": 0.05,
    "sadness": 0.1,
    "anger": 0.0,
    "fear": 0.2,
    "surprise": 0.0,
    "neutral": 0.65
  },
  "quality_flags": [],
  "evidence": [
    "STT total transcript length: 11 characters.",
    "STT total word count: 2.",
    "Короткий ответ: мало слов или символов для устойчивого текстового анализа."
  ],
  "model_version": "rubert-go-emotions-v1",
  "emotion_model_version": "emotion-rubert-base-cased-russian-emotion-detection-ru-go-emotions",
  "processing_time_ms": 120,
  "error": null,
  "answers": [
    {
      "answer_id": 500,
      "question_id": 9001,
      "audio_s3_key": "examinations/100/answers/500/audio.webm",
      "bytes": 524288,
      "stt": {
        "transcript": "текст ответа",
        "language": "ru",
        "segments": [
          {
            "start_ms": 0,
            "end_ms": 1200,
            "text": "текст ответа"
          }
        ],
        "duration_ms": 1200,
        "word_count": 2,
        "model_version": "faster-whisper-medium"
      },
      "quality_flags": [],
      "evidence": []
    }
  ]
}
```

Поля:
- `stt.transcript`: агрегированный transcript по всем audio answers текущего обследования;
- `stt.language`: наиболее частый язык, который вернул `faster-whisper`; может быть `null`;
- `stt.segments`: STT-сегменты в миллисекундах, обогащённые `answer_id` и `question_id`;
- `stt.duration_ms`: суммарная длительность STT-аудио, если модель вернула duration;
- `stt.word_count`: количество слов в transcript;
- `stt.model_version`: фактическая версия/источник STT-модели;
- `features.transcript_length_chars`, `features.word_count`, `features.sentence_count`, `features.avg_sentence_length`: базовые признаки объёма и структуры transcript;
- `features.uncertainty_marker_count`: количество простых русскоязычных маркеров неопределённости (`не знаю`, `возможно`, `кажется`, `затрудняюсь` и т.п.);
- `features.negation_count`: количество простых отрицательных маркеров;
- `features.distress_marker_count`: количество прямых текстовых маркеров неблагополучия/напряжения (`переживаю`, `не готов`, `плохо`, `болит`, `не спал`, `ужас` и т.п.);
- `features.short_answer_flag`: `true`, если transcript слишком короткий для устойчивого текстового анализа;
- `scores.text_negativity_score`: эвристический `0..1` score на основе `sadness`, `anger`, `fear`, усиленный при накоплении нескольких негативных эмоций и прямых distress-маркеров в transcript;
- `scores.text_anxiety_score`: эвристический `0..1` score на основе `fear`, distress-маркеров, uncertainty markers и negations;
- `scores.text_confidence_score`: эвристический `0..1`, где `1` означает более уверенный/определённый ответ; снижается не только от uncertainty/evasion, но и от выраженного негативного или тревожного текстового профиля;
- `scores.text_coherence_score`: эвристический `0..1` score связности по длине/структуре/уклончивости;
- `scores.text_evasion_score`: эвристический `0..1` score уклончивости по короткому ответу, uncertainty markers и negations;
- `emotion_probs`: вероятности эмоций из emotion-модели, нормализованные к ключам `joy`, `sadness`, `anger`, `fear`, `surprise`, `neutral`; если модель не поддерживает часть классов, они возвращаются как `0.0`;
- `quality_flags`: `stt_failed`, `emotion_model_failed`, `empty_transcript` при пустом STT, `too_short_text` при коротком transcript;
- `evidence`: человекочитаемые факты: короткий ответ, число неопределённых формулировок, отрицательные маркеры, высокая тревожная/негативная окраска;
- `emotion_model_version`: фактический источник emotion-модели;
- `answers[*].stt`: per-answer STT breakdown; transcript хранится в channel result payload и не создаёт отдельного постоянного storage-контракта.

Runtime config:
- `STT_MODEL_PATH`: путь к локальной модели внутри контейнера, default `/app/models/stt`;
- `STT_MODEL_SIZE`: fallback model size/name для download/dev-запуска, default `medium`, допускается `large-v3`;
- `STT_DEVICE`: `cpu` или `cuda`;
- `STT_COMPUTE_TYPE`: например `int8` для CPU или `float16` для CUDA;
- `TEXT_EMOTION_MODEL_PATH`: путь к локальной emotion-модели внутри контейнера, default `/app/models/text-emotion`;
- `TEXT_EMOTION_MODEL_NAME`: fallback Hugging Face model id для dev-download, default `seara/rubert-base-cased-russian-emotion-detection-ru-go-emotions`;
- `TEXT_EMOTION_DEVICE`: `cpu` или `cuda`;
- `HF_HUB_OFFLINE=1`: запрещает download и требует наличие локальной модели.

Docker runtime:
- host-папка `./models` монтируется в `text-worker` как `/app/models:ro`;
- локальная offline STT-модель должна лежать в `./models/stt` на хосте и быть доступна как `/app/models/stt` внутри контейнера;
- локальная offline emotion-модель должна лежать в `./models/text-emotion` на хосте и быть доступна как `/app/models/text-emotion` внутри контейнера;
- локальная STT-модель считается доступной только если в `STT_MODEL_PATH` есть `config.json`; пустая директория `models/stt` не блокирует dev-download;
- локальная emotion-модель считается доступной только если в `TEXT_EMOTION_MODEL_PATH` есть `config.json`; пустая директория `models/text-emotion` не блокирует dev-download;
- если локальные модели отсутствуют/неполные и `HF_HUB_OFFLINE` выключен, worker использует `STT_MODEL_SIZE` и `TEXT_EMOTION_MODEL_NAME` для dev-download.

Compatibility proxy-поля `text_total_characters`, `text_non_empty_answers`, `audio_total_bytes` удалены. Aggregator читает `scores.text_negativity_score`, `scores.text_anxiety_score`, `scores.text_confidence_score`, `scores.text_coherence_score` и `scores.text_evasion_score`.

## Answers API

Объект `Answer`:

```json
{
  "id": 500,
  "examination_id": 100,
  "created_by_user_id": 1,
  "text": "",
  "audio_s3_key": "examinations/100/answers/500/audio.webm",
  "created_at": "2026-03-19T12:05:00Z"
}
```

Примечание:
- operator flow на текущем этапе audio-only, поэтому `text` в объекте ответа сохраняется как пустая строка и зарезервирован под будущую транскрипцию/текстовый слой.

### POST /answers

Назначение:
- загрузка аудиоответа в S3-совместимое хранилище;
- сохранение аудиозаписи и `audio_s3_key` в PostgreSQL.

Аутентификация:
- `Authorization: Bearer <jwt>`.

Тип запроса:
- `multipart/form-data`.

Поля формы:
- `examination_id`: integer, обязательное;
- `examination_question_id`: integer, обязательное;
- `specialist_id`: integer, обязательное;
- `audio`: binary file, обязательное.

Ответ `201 Created`: объект `Answer`.

Правила:
- файл не сохраняется в PostgreSQL;
- в БД хранится `audio_s3_key`, а поле `answer_text` сохраняется как пустая строка до появления отдельной транскрипции;
- ключ объекта детерминирован: `examinations/{examination_id}/answers/{answer_id}/audio{ext}`;
- ответ можно сохранять только для обследования в статусе `collecting_answers`;
- `examination_question_id` должен ссылаться на snapshot-вопрос из `examination_questions`, принадлежащий тому же `examination_id` и `specialist_id`;
- повторный ответ на один и тот же snapshot-вопрос должен отклоняться как конфликт.

Ошибки:
- `400 Bad Request` при невалидном multipart payload;
- `404 Not Found` если обследование не найдено;
- `409 Conflict` если обследование не находится в статусе `collecting_answers` или если на этот snapshot-вопрос уже сохранён ответ.

## Questionnaires API

Объект `Questionnaire`:

```json
{
  "id": 5,
  "title": "Предсменный опрос",
  "description": "Базовый набор вопросов",
  "is_active": true,
  "questions": [
    {
      "id": 11,
      "text": "Как вы себя чувствуете сегодня?",
      "position": 1
    },
    {
      "id": 12,
      "text": "Что вызвало наибольшее напряжение за последние сутки?",
      "position": 2
    }
  ],
  "created_at": "2026-03-19T12:00:00Z",
  "updated_at": "2026-03-19T12:00:00Z"
}
```

Все endpoints раздела защищены Bearer JWT.

### GET /questionnaires

Назначение:
- получение read-only списка опросников для operator/admin frontend-сценариев.

Аутентификация:
- `Authorization: Bearer <jwt>`.

Роли и авторизация:
- доступно ролям `operator` и `admin`;
- endpoint предназначен для выбора опросника в operator flow и для чтения списка в admin UI;
- endpoint не расширяет mutation-права оператора: создание и изменение опросников остаются только у `admin`.

Ответ `200 OK`:

```json
{
  "items": [
    {
      "id": 5,
      "title": "Предсменный опрос",
      "description": "Базовый набор вопросов",
      "is_active": true,
      "usage_count": 18,
      "last_used_at": "2026-03-25T06:20:00Z",
      "last_edited_at": "2026-03-25T05:55:00Z",
      "last_editor": {
        "id": 1,
        "login": "admin"
      },
      "questions": [
        {
          "id": 11,
          "text": "Как вы себя чувствуете сегодня?",
          "position": 1
        }
      ],
      "created_at": "2026-03-19T12:00:00Z",
      "updated_at": "2026-03-19T12:00:00Z"
    }
  ]
}
```

Дополнительно:
- `usage_count` — реальное число обследований, созданных с этим `questionnaire_id`;
- `last_used_at` — время последнего использования опросника в обследовании или `null`, если он ещё не использовался;
- `last_edited_at` — persisted timestamp последнего административного изменения метаданных/состава вопросов;
- `last_editor` — `{ id, login }` пользователя, выполнившего последнее изменение, либо `null` для исторических записей без зафиксированного редактора.

### POST /questionnaires

Роли и авторизация:
- доступно только роли `admin`.

Запрос:

```json
{
  "title": "Предсменный опрос",
  "description": "Базовый набор вопросов",
  "is_active": true,
  "questions": [
    {
      "text": "Как вы себя чувствуете сегодня?"
    },
    {
      "text": "Что вызвало наибольшее напряжение за последние сутки?"
    }
  ]
}
```

Поля:
- `title`: обязательное;
- `description`: необязательное;
- `is_active`: обязательное;
- `questions`: обязательный непустой массив;
- порядок элементов в `questions` определяет `position` в ответе и в БД.

Ответ `201 Created`: объект `Questionnaire`.

Ошибки:
- `400 Bad Request` если payload невалидный.

### GET /questionnaires/{id}

Роли и авторизация:
- доступно только роли `admin`.

Ответ `200 OK`: объект `Questionnaire` того же enriched формата, что и в `GET /questionnaires`.

Ошибки:
- `404 Not Found` если опросник не найден.

### PUT /questionnaires/{id}

Назначение:
- полное обновление метаданных и состава вопросов опросника.

Запрос: такой же, как `POST /questionnaires`.

Правила:
- состав `questions` заменяется целиком;
- `position` в ответе пересчитывается по порядку элементов входного массива.

Ответ `200 OK`: обновлённый объект `Questionnaire`.

Ошибки:
- `400 Bad Request` если payload невалидный;
- `404 Not Found` если опросник не найден.
- `403 Forbidden` если роль не `admin`.

## Infrastructure Config Contract

Обязательные переменные окружения:
- `DATABASE_URL`
- `JWT_ACCESS_SECRET`
- `MINIO_ENDPOINT`
- `MINIO_ACCESS_KEY_ID` или `MINIO_ROOT_USER`
- `MINIO_SECRET_ACCESS_KEY` или `MINIO_ROOT_PASSWORD`
- `MINIO_BUCKET`

Поддерживаемые переменные окружения:
- `HTTP_ADDR`, по умолчанию `:8080`
- `MIGRATIONS_DIR`, по умолчанию `migrations`
- `RABBITMQ_URL`, по умолчанию `amqp://guest:guest@localhost:5672/`
- `PROCESSING_OUTBOX_POLL_INTERVAL`, по умолчанию `1s`
- `PROCESSING_OUTBOX_MAX_ATTEMPTS`, по умолчанию `3`
- `AUDIO_RETENTION_TTL_DAYS`, по умолчанию `30`
- `JWT_ISSUER`, по умолчанию `core-backend`
- `JWT_ACCESS_TTL`, по умолчанию `15m`
- `MINIO_USE_SSL`, по умолчанию `false`
- `MAX_UPLOAD_SIZE_BYTES`, по умолчанию `26214400`
- `INITIAL_USER_LOGIN`
- `INITIAL_USER_PASSWORD`
- `INITIAL_USER_ROLE`, по умолчанию `operator`

Bootstrap initial user:
- если заданы одновременно `INITIAL_USER_LOGIN` и `INITIAL_USER_PASSWORD`, backend при старте создаёт или обновляет начального пользователя;
- migration `000014_seed_basic_questionnaire` при инициализации пустой БД создаёт стартовый опубликованный опросник `Базовый опрос` с 3 вопросами:
  - `Расскажите кратко о своем текущем состоянии.`
  - `Как вы спали и отдыхали в последние сутки?`
  - `Есть ли что-то, что мешает вам сосредоточиться на работе?`
- `INITIAL_USER_ROLE` должен ссылаться на существующую роль (`admin` или `operator`).

Migrations bootstrap:
- при старте backend автоматически применяет все `*.up.sql` миграции из `MIGRATIONS_DIR`;
- список применённых миграций хранится в таблице `schema_migrations`.

## Database Schema Contract

### roles

- `id BIGSERIAL PRIMARY KEY`
- `slug TEXT NOT NULL UNIQUE`
- `name TEXT NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

Начальные записи:
- `admin`
- `operator`

### users

- `id BIGSERIAL PRIMARY KEY`
- `login TEXT NOT NULL UNIQUE`
- `password_hash TEXT NOT NULL`
- `role_id BIGINT NOT NULL REFERENCES roles(id)`
- `is_active BOOLEAN NOT NULL DEFAULT TRUE`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

### specialists

- `id BIGSERIAL PRIMARY KEY`
- `full_name TEXT NOT NULL`
- `personnel_number TEXT UNIQUE`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

### examinations

- `id BIGSERIAL PRIMARY KEY`
- `specialist_id BIGINT NOT NULL REFERENCES specialists(id) ON DELETE RESTRICT`
- `created_by_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT`
- `questionnaire_id BIGINT NULL REFERENCES questionnaires(id) ON DELETE RESTRICT`
- `status TEXT NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `started_at TIMESTAMPTZ NULL`
- `finished_at TIMESTAMPTZ NULL`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

Допустимые значения `status`:
- `created`
- `collecting_answers`
- `ready_for_processing`
- `processing`
- `failed`

### answers

- `id BIGSERIAL PRIMARY KEY`
- `examination_id BIGINT NOT NULL REFERENCES examinations(id) ON DELETE CASCADE`
- `created_by_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT`
- `answer_text TEXT NOT NULL`
- `audio_s3_key TEXT NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

### examination_processing_launches

- `examination_id BIGINT PRIMARY KEY REFERENCES examinations(id) ON DELETE CASCADE`
- `launched_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

Назначение:
- launch fence для идемпотентного старта Phase 2 processing pipeline;
- не заменяется outbox-моделью, а используется как единственный guard от повторного fan-out при повторных `finish`.

### examination_channel_runs

- `id BIGSERIAL PRIMARY KEY`
- `examination_id BIGINT NOT NULL REFERENCES examinations(id) ON DELETE CASCADE`
- `channel TEXT NOT NULL`
- `status TEXT NOT NULL`
- `attempt_count INTEGER NOT NULL DEFAULT 0`
- `max_attempts INTEGER NOT NULL DEFAULT 3`
- `message_version INTEGER NOT NULL`
- `last_error_code TEXT NULL`
- `last_error_message TEXT NULL`
- `broker_message_id TEXT NULL`
- `broker_correlation_id TEXT NULL`
- `queued_at TIMESTAMPTZ NULL`
- `started_at TIMESTAMPTZ NULL`
- `finished_at TIMESTAMPTZ NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `UNIQUE (examination_id, channel)`

Допустимые значения `channel`:
- `text`
- `acoustic`
- `paralinguistic`

Допустимые значения `status`:
- `pending`
- `queued`
- `processing`
- `succeeded`
- `retry_scheduled`
- `failed_temporary`
- `failed_fatal`
- `exhausted`

### processing_outbox

- `id BIGSERIAL PRIMARY KEY`
- `examination_id BIGINT NOT NULL REFERENCES examinations(id) ON DELETE CASCADE`
- `channel_run_id BIGINT NOT NULL REFERENCES examination_channel_runs(id) ON DELETE CASCADE`
- `channel TEXT NOT NULL`
- `exchange_name TEXT NOT NULL`
- `routing_key TEXT NOT NULL`
- `status TEXT NOT NULL`
- `attempt_count INTEGER NOT NULL DEFAULT 0`
- `max_attempts INTEGER NOT NULL DEFAULT 3`
- `message_version INTEGER NOT NULL`
- `payload JSONB NOT NULL`
- `broker_message_id TEXT NULL`
- `broker_correlation_id TEXT NULL`
- `last_error_code TEXT NULL`
- `last_error_message TEXT NULL`
- `published_at TIMESTAMPTZ NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `UNIQUE (channel_run_id)`

Допустимые значения `status`:
- `pending`
- `published`
- `failed`

### channel_results

- `id BIGSERIAL PRIMARY KEY`
- `examination_id BIGINT NOT NULL REFERENCES examinations(id) ON DELETE CASCADE`
- `channel_run_id BIGINT NOT NULL REFERENCES examination_channel_runs(id) ON DELETE CASCADE`
- `channel TEXT NOT NULL`
- `message_version INTEGER NOT NULL`
- `attempt INTEGER NOT NULL`
- `status TEXT NOT NULL`
- `model_version TEXT NOT NULL`
- `payload JSONB NOT NULL`
- `error_code TEXT NULL`
- `error_message TEXT NULL`
- `broker_message_id TEXT NULL`
- `broker_correlation_id TEXT NULL`
- `completed_at TIMESTAMPTZ NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `UNIQUE (channel_run_id, attempt)`

Допустимые значения `status`:
- `succeeded`
- `temporary_error`
- `fatal_error`

### questionnaires

- `id BIGSERIAL PRIMARY KEY`
- `title TEXT NOT NULL`
- `description TEXT NULL`
- `is_active BOOLEAN NOT NULL DEFAULT TRUE`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

### questions

- `id BIGSERIAL PRIMARY KEY`
- `text TEXT NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

### questionnaire_questions

- `questionnaire_id BIGINT NOT NULL REFERENCES questionnaires(id) ON DELETE CASCADE`
- `question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE`
- `position INTEGER NOT NULL CHECK (position > 0)`
- `PRIMARY KEY (questionnaire_id, question_id)`
- `UNIQUE (questionnaire_id, position)`

## Workflow Status Model

В рамках текущего API-контракта разрешены статусы обследования:
- `created`
- `collecting_answers`
- `ready_for_processing`
- `processing`
- `aggregating`
- `aggregated`
- `decision_pending`
- `completed`
- `failed`

Правила модели статусов:
- `processing`, `aggregating`, `aggregated`, `decision_pending`, `completed` и `failed` используются как coarse-grained examination runtime states в основном объекте `Examination`;
- детальный прогресс по каналам, попыткам и ошибкам публикуется только через `GET /examinations/{id}/processing-status`;
- `completed` означает terminal projection decision delivery, а детализация результата должна читаться через `GET /examinations/{id}/result`.

## Frontend runtime configuration

### `NEXT_PUBLIC_API_URL`

- Public base URL core backend для browser-side frontend requests.
- В Docker Compose должен указывать на host-reachable адрес `http://localhost:18080`, потому что это значение попадает в клиентский bundle и используется браузером, а не контейнерной сетью.
- Для локального запуска вне Compose по умолчанию также используется `http://localhost:18080`.

### `INTERNAL_API_BASE_URL`

- Server-side base URL core backend для Next.js route handlers и readiness probe внутри frontend runtime.
- В Docker Compose должен указывать на внутренний адрес `http://core-backend:8080`.
- Вне Compose по умолчанию может совпадать с `http://localhost:18080`.

### `FRONTEND_HOST_PORT`

- Внешний порт публикации frontend в локальном Docker Compose.
- Значение по умолчанию: `3000`.

## KESMI Runtime Contract

Phase 4 runtime использует обязательный internal-only compose service `wimi`.

Обязательные env-переменные backend:

- `KESMI_BASE_URL=http://wimi:8081`
- `KESMI_MODEL_ID=specialists-model-v2-decision`
- `KESMI_TIMEOUT_MS=1000ms`
- `KESMI_MAX_RETRIES=2`
- `KESMI_RETRY_BACKOFF_MS=500ms`
- `WIMI_AUTOLOAD_MODEL_PATH=/opt/wimi-models/specialists_model_v2_decision.xml`

Правила:
- WiMi в текущем локальном compose остаётся internal-only сервисом и не публикуется наружу; ручные smoke/debug проверки выполняются изнутри compose-сети через `core-backend` или другой контейнер;
- только `core-backend` обращается к WiMi;
- decision-модель `wimi-server/models/specialists_model_v2_decision.xml` автозагружается в контейнере `wimi` при старте через `POST /Models` под `KESMI_MODEL_ID`;
- smoke-проверка runtime выполняется изнутри compose-сети через `docker compose exec core-backend ... http://wimi:8081/Models`.
