# Контракты системы

## HTTP API

### GET /health

Назначение:
- технический health-check core backend;
- проверка доступности HTTP-сервиса и PostgreSQL.

Ответ `200 OK`:

```json
{
  "status": "ok",
  "database": "up"
}
```

Ответ `503 Service Unavailable`:

```json
{
  "status": "degraded",
  "database": "down"
}
```

Технические детали:
- `Content-Type: application/json`;
- endpoint без аутентификации;
- проверяется только PostgreSQL, без RabbitMQ, MinIO и ML.

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
- перевод обследования в статус `ready_for_processing`.

Ответ `200 OK`: объект `Examination`.

Правила:
- переход допустим из `collecting_answers`;
- если обследование уже в `ready_for_processing`, endpoint идемпотентно возвращает текущее состояние;
- любые другие переходы дают ошибку.

Ошибки:
- `404 Not Found` если обследование не найдено;
- `409 Conflict` при недопустимом переходе статуса.

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
- `text`: string, обязательное;
- `audio`: binary file, обязательное.

Ответ `201 Created`: объект `Answer`.

Правила:
- файл не сохраняется в PostgreSQL;
- в БД хранится только `audio_s3_key`;
- ключ объекта детерминирован: `examinations/{examination_id}/answers/{answer_id}/audio{ext}`;
- ответ можно сохранять только для обследования в статусе `collecting_answers`.

Ошибки:
- `400 Bad Request` при невалидном multipart payload;
- `404 Not Found` если обследование не найдено;
- `409 Conflict` если обследование не находится в статусе `collecting_answers`.

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

### answers

- `id BIGSERIAL PRIMARY KEY`
- `examination_id BIGINT NOT NULL REFERENCES examinations(id) ON DELETE CASCADE`
- `created_by_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT`
- `answer_text TEXT NOT NULL`
- `audio_s3_key TEXT NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

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

В рамках текущего синхронного этапа реализованы и разрешены только статусы обследования:
- `created`
- `collecting_answers`
- `ready_for_processing`

Статусы `processing`, `waiting_results`, `aggregating`, `decision_pending`, `completed`, `failed` пока не реализованы в коде и не используются API этого этапа.

## Frontend runtime configuration

### `NEXT_PUBLIC_API_URL`

- Базовый URL core backend для frontend-клиента.
- В Docker Compose должен указывать на внутренний адрес `http://core-backend:8080`.
- Для локального запуска вне Compose по умолчанию используется `http://localhost:8080`.

### `FRONTEND_HOST_PORT`

- Внешний порт публикации frontend в локальном Docker Compose.
- Значение по умолчанию: `3000`.
