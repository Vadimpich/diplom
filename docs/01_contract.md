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
- `/users` и `/questionnaires` доступны только роли `admin`; для роли `operator` backend возвращает `403 Forbidden`;
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
    "primary_metric_key": "overall_proxy_index",
    "neutral_recommendation_placeholder": "phase3_pending_external_decision"
  },
  "metrics": [
    {
      "key": "overall_proxy_index",
      "label": "Сводный прокси-индекс",
      "value": 0.58,
      "scale": "0..1",
      "direction": "higher_means_more_deviation"
    }
  ],
  "channel_contributions": [
    {
      "channel": "text",
      "metric_key": "overall_proxy_index",
      "weight": 0.33,
      "contribution": 0.17,
      "evidence_keys": [
        "text_proxy_signal"
      ]
    }
  ],
  "explanations": [
    {
      "position": 1,
      "kind": "summary",
      "text": "Повышение индекса в основном связано с proxy-метриками acoustic и paralinguistic каналов."
    }
  ],
  "baseline_snapshot": {
    "algorithm_version": "baseline-v1",
    "refreshed_at": "2026-03-22T10:02:09Z",
    "general": {
      "delta": 0.21,
      "band": "mild",
      "reference_population_version": "general-v1"
    },
    "personal": {
      "delta": 0.37,
      "band": "moderate",
      "baseline_exam_count": 4,
      "update_eligible": false
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
- названия метрик должны быть нейтральными и proxy-oriented до появления финальной ML-семантики;
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
          "key": "overall_proxy_index",
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
  "payload_version": "decision-input-v1",
  "aggregation_version": "agg-v1",
  "examination_id": 101,
  "specialist_id": 55,
  "generated_at": "2026-03-22T10:02:10Z",
  "summary": {
    "overall_score": 0.58,
    "overall_band": "elevated",
    "primary_metric_key": "overall_proxy_index"
  },
  "metrics": [
    {
      "key": "overall_proxy_index",
      "label": "Сводный прокси-индекс",
      "value": 0.58,
      "scale": "0..1",
      "direction": "higher_means_more_deviation"
    }
  ],
  "channel_contributions": [
    {
      "channel": "text",
      "metric_key": "overall_proxy_index",
      "weight": 0.33,
      "contribution": 0.17,
      "evidence_keys": [
        "text_proxy_signal"
      ]
    }
  ],
  "baseline_snapshot": {
    "algorithm_version": "baseline-v1",
    "general": {
      "delta": 0.21,
      "band": "mild"
    },
    "personal": {
      "delta": 0.37,
      "band": "moderate",
      "baseline_exam_count": 4,
      "update_eligible": false
    }
  },
  "service_metadata": {
    "target_system": "kesmi",
    "delivery_mode": "placeholder",
    "message": "analysis_not_implemented_yet"
  }
}
```

Правила:
- `decision_input` строится только из backend-owned aggregated profile и baseline snapshot;
- payload versioned через `payload_version`;
- список полей в `decision_input` стабилизируется раньше, чем появятся реальные WiMi model parameters;
- последующий mapping `decision_input -> incommingParameters` живёт в integration layer и не меняет operator-facing contract.

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
      "key": "overall_proxy_index",
      "value": 0.58
    },
    {
      "key": "speech_stability_proxy",
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
            "key": "overall_proxy_index",
            "value": 0.46
          }
        ]
      }
    ]
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
    "metric_scores": {
      "overall_proxy_index": {
        "delta": 0.08,
        "robust_z": 0.54,
        "band": "low"
      }
    }
  },
  "personal_deviation": {
    "score": 0.37,
    "band": "moderate",
    "metric_scores": {
      "overall_proxy_index": {
        "delta": 0.12,
        "robust_z": 1.91,
        "band": "mild"
      }
    }
  },
  "update_eligibility": {
    "eligible": false,
    "reason": "outlier_detected",
    "baseline_exam_count_after_update": 4
  },
  "next_baseline": {
    "exam_count": 4,
    "centers": {
      "overall_proxy_index": 0.46
    },
    "scales": {
      "overall_proxy_index": 0.03
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
- `metric_scores` внутри обоих deviation-блоков содержат per-metric delta, robust z-score и severity band для канонических proxy-метрик;
- `algorithm_version` и `refreshed_at` обязательны для persistence snapshot в core backend;
- `update_eligibility` обязателен для outlier-gated baseline refresh;
- `next_baseline` обязателен и содержит кандидатный snapshot baseline после применения bounded-history и outlier gate;
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
      "created_at": "2026-03-19T12:00:00Z",
      "updated_at": "2026-03-19T12:00:00Z"
    }
  ]
}
```

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
- обновление логина, роли и статуса активности пользователя.

Аутентификация:
- `Authorization: Bearer <jwt>`.

Запрос:

```json
{
  "login": "operator2",
  "role": "operator",
  "is_active": true
}
```

Поля:
- `login`: обязательное, уникальное;
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
  "created_at": "2026-03-19T12:00:00Z",
  "updated_at": "2026-03-19T12:00:00Z"
}
```

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

Ответ `200 OK`: объект `Examination`.

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
- общее поле `status` для обследования в Phase 2 допускает значения `ready_for_processing`, `processing`, `failed`;
- детальное состояние каналов живёт только в `channels[*]` и не должно дублироваться в основном объекте `Examination`;
- `terminal=true` означает, что все обязательные каналы достигли финального состояния (`succeeded` либо терминальная ошибка с переводом обследования в `failed`);
- `last_error_message` предназначено для операторской диагностики и не должно содержать stack trace или чувствительные данные.

Ошибки:
- `404 Not Found` если обследование не найдено;
- `409 Conflict` если обследование ещё не было переведено в `ready_for_processing`.

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
      "answer_text": "Ответ обследуемого"
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
- `answers[*].answer_text`: текстовая транскрипция/ответ, доступная всем каналам как часть общего контракта;
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
  "model_version": "text-stub-0.1.0",
  "error_code": null,
  "error_message": null,
  "payload": {
    "summary": "stub result"
  }
}
```

Поля:
- `status`: одно из `succeeded`, `temporary_error`, `fatal_error`;
- `payload`: channel-specific JSON object, но envelope shape одинакова для всех каналов;
- `error_code` и `error_message` обязательны при `temporary_error` и `fatal_error`, должны быть пустыми при `succeeded`;
- `completed_at` обязателен для всех terminal result-сообщений;
- worker не должен публиковать разные envelope shapes для разных каналов.
- stub workers текущей фазы обязаны заполнять `model_version` и использовать `temporary_error` для transport/S3 availability failures, `fatal_error` для невалидного payload или отсутствующего S3 object reference.

## Answers API

Объект `Answer`:

```json
{
  "id": 500,
  "examination_id": 100,
  "created_by_user_id": 1,
  "text": "Ответ обследуемого",
  "audio_s3_key": "examinations/100/answers/500/audio.webm",
  "created_at": "2026-03-19T12:05:00Z"
}
```

### POST /answers

Назначение:
- загрузка аудиоответа в S3-совместимое хранилище;
- сохранение текстового ответа и `audio_s3_key` в PostgreSQL.

Аутентификация:
- `Authorization: Bearer <jwt>`.

Тип запроса:
- `multipart/form-data`.

Поля формы:
- `examination_id`: integer, обязательное;
- `examination_question_id`: integer, обязательное;
- `specialist_id`: integer, обязательное;
- `text`: string, обязательное;
- `audio`: binary file, обязательное.

Ответ `201 Created`: объект `Answer`.

Правила:
- файл не сохраняется в PostgreSQL;
- в БД хранится только `audio_s3_key`;
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
- получение списка опросников для operator/admin frontend-сценариев.

Ответ `200 OK`:

```json
{
  "items": [
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
        }
      ],
      "created_at": "2026-03-19T12:00:00Z",
      "updated_at": "2026-03-19T12:00:00Z"
    }
  ]
}
```

### POST /questionnaires

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

Ответ `200 OK`: объект `Questionnaire`.

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
- `JWT_ISSUER`, по умолчанию `core-backend`
- `JWT_ACCESS_TTL`, по умолчанию `15m`
- `MINIO_USE_SSL`, по умолчанию `false`
- `MAX_UPLOAD_SIZE_BYTES`, по умолчанию `26214400`
- `INITIAL_USER_LOGIN`
- `INITIAL_USER_PASSWORD`
- `INITIAL_USER_ROLE`, по умолчанию `operator`

Bootstrap initial user:
- если заданы одновременно `INITIAL_USER_LOGIN` и `INITIAL_USER_PASSWORD`, backend при старте создаёт или обновляет начального пользователя;
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

- Базовый URL core backend для frontend-клиента.
- В Docker Compose должен указывать на внутренний адрес `http://core-backend:8080`.
- Для локального запуска вне Compose по умолчанию используется `http://localhost:8080`.

### `FRONTEND_HOST_PORT`

- Внешний порт публикации frontend в локальном Docker Compose.
- Значение по умолчанию: `3000`.

## KESMI Runtime Contract

Phase 4 runtime использует обязательный internal-only compose service `wimi`.

Обязательные env-переменные backend:

- `KESMI_BASE_URL=http://wimi:8081`
- `KESMI_MODEL_ID=stub-decision-model`
- `KESMI_TIMEOUT_MS=1000ms`
- `KESMI_MAX_RETRIES=2`
- `KESMI_RETRY_BACKOFF_MS=500ms`

Правила:
- WiMi не должен публиковаться наружу через host-port и не требует `WIMI_HOST_PORT`;
- только `core-backend` обращается к WiMi;
- smoke-проверка runtime выполняется изнутри compose-сети через `docker compose exec core-backend ... http://wimi:8081/Models`.
