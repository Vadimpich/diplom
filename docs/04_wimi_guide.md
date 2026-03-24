# WiMi / КЭСМИ Guide

## Назначение

WiMi Server в папке `wimi-server` является серверным API-слоем над движком Разуматора и может выступать внешней СППР для нашего проекта. Для нас это не часть доменной логики и не пользовательский сервис, а внешний decision engine, к которому обращается только `core-backend`.

Основной архитектурный вывод:

- наш проект должен интегрироваться с WiMi через отдельный `KESMI Adapter` внутри `core-backend`;
- frontend не должен ходить в WiMi напрямую;
- интеграция должна быть идемпотентной, с разделением transport и business errors;
- Phase 4 должна закрыть переходы `aggregated -> decision_pending -> completed` согласно `docs/00_project.md`.

## Важное ограничение текущего этапа

Сейчас в проекте нет:

- финальной decision-модели внутри WiMi;
- финальных продуктивных моделей для `text`, `acoustic` и `paralinguistic`;
- утверждённого списка входных и выходных параметров decision layer.

Это не блокирует завершение архитектурной части Phase 4. Рабочая стратегия на текущем этапе:

- доделать все интеграционные слои, workflow, persistence, retry/error handling и UI diagnostics заранее;
- использовать внутренние versioned DTO и stub/mock outputs там, где ещё нет реальных моделей;
- отложить на финальный этап только наполнение этих seams реальными ML-моделями, feature mapping и decision-моделью WiMi.

Практический вывод:

- Phase 4 можно реализовать contract-first;
- integration module к WiMi можно сделать без знания финальных переменных модели;
- нужно отдельно стабилизировать внутренние канонические форматы:
  - normalized channel result;
  - aggregated profile;
  - decision input;
  - decision result.

Главный риск при таком подходе не в том, что модели появятся поздно, а в том, что без внутренних стабильных DTO внешние модели начнут диктовать архитектуру проекта. Этого нужно избежать.

---

## Что важно про поставку WiMi

WiMi поставляется как готовый проприетарный бинарь.

- Linux-поставка: `wimi-server/bin/WiMi-0.1.7.deb`
- Windows-поставка: `wimi-server/bin/WiMi Server Windows.zip`

По Linux-пакету подтверждено:

- установка в `/usr/local/bin/WiMi/`;
- основной бинарник: `WiMi`;
- конфиг REST API: `MivREST.ini`;
- запуск через `WiMi -e`;
- рядом лежит локальная API-документация.

Вывод:

- для дипломного проекта разумнее считать WiMi отдельным внутренним сервисом на Linux-хосте;
- модель WiMi должна быть предзагружена и управляться как deployment/admin concern, а не как часть operator workflow.

---

## Конфиг WiMi

Фактический `MivREST.ini` из поставки:

```ini
[listener]
port=8081
minThreads=10
maxThreads=100
cleanupInterval=1000
readTimeout=1000
maxRequestSize=10000000
maxMultiPartSize=10000000

[logging]
fileName=MivarREST.log
maxSize=1000000
maxBackups=10
minLevel=0
bufferSize=100
timestampFormat=dd.MM.yyyy hh:mm:ss.zzz
msgFormat={timestamp} {typeNr} {type} {thread} {file} {function} {msg}

[files]
path=./
encoding=UTF-8
maxAge=90000
cacheTime=60000
cacheSize=1000000
maxCachedFileSize=65536
```

Практически это означает:

- дефолтный порт поставки: `8081`;
- в локальной инструкции встречается `8092`, но это ручное изменение, не дефолт;
- у сервера короткий `readTimeout=1000`, это надо учитывать в reverse proxy и при сетевой диагностике;
- сервер многопоточный и использует локальный файловый кэш;
- размер запроса ограничен примерно `10 MB`.

Для интеграции проекта нужно вынести в env:

- `KESMI_BASE_URL`
- `KESMI_MODEL_ID`
- `KESMI_TIMEOUT_MS`
- `KESMI_MAX_RETRIES`
- `KESMI_RETRY_BACKOFF_MS`

---

## Запуск

### Linux

Из поставки и локальной инструкции:

1. Установить `.deb`.
2. Настроить `MivREST.ini`.
3. Запускать `WiMi -e` или `WiMi.run`.
4. Для постоянной работы оформить как `systemd` service.

Фактический wrapper:

```bash
#!/bin/bash
export LD_LIBRARY_PATH="$LD_LIBRARY_PATH:/usr/local/bin/WiMi/libs"
export LD_PRELOAD="$LD_PRELOAD:/usr/local/bin/WiMi/libs/libjemalloc.so.2"

./WiMi -e
```

Это важно, потому что запуск зависит от bundled libs и не должен сводиться к голому запуску бинаря из любого каталога.

Для контейнеризации в нашем проекте дополнительно подтвердилось:

- одного vendor `.deb` недостаточно;
- в Ubuntu runtime нужен системный пакет `libglib2.0-0`, иначе WiMi падает на `libgthread-2.0.so.0` / `libglib-2.0.so.0`;
- поэтому Docker-образ должен явно доустанавливать этот runtime dependency до запуска `WiMi -e`.

### Windows

Запуск: `WiMi.exe -e`.

Для нашего проекта это вторично; основной вариант интеграции должен ориентироваться на Linux-hosted runtime.

---

## REST API WiMi

Ниже только то, что важно для интеграции.

### `GET /Models`

Использование:

- `GET /Models` — список доступных моделей;
- `GET /Models?modelID=<id>` — статистика по модели.

Полезно для:

- smoke check после развёртывания;
- проверки, что нужная decision-модель загружена;
- контроля `poolSize`.

### `POST /Models`

Назначение:

- загрузка модели в движок.

Вход:

- `modelID`
- `modelPoolSize`
- `modelXML`

Для runtime интеграции проекта не нужен. Это административный/deploy endpoint.

### `DELETE /Models`

Назначение:

- удаление модели.

Важно:

- если модель используется, удаление может быть подтверждено раньше фактического освобождения пула.

Для runtime-интеграции проекта не нужен.

### `POST /ModelCalc`

Это основной endpoint для Phase 4.

Вход:

- `modelID`
- `incommingParameters`
  - массив `{ id, value }`
- `outputParameters`
- `service.outputFields`

Поддерживаемые `outputFields`:

- `algorithm`
- `requiredExploredParameters`
- `requiredNotExploredParameters`
- `notRequiredExploredParameters`
- `timing`

Типовой успешный ответ содержит:

- `requiredExploredParameters`
- `requiredNotExploredParameters`
- `notRequiredExploredParameters`
- `algorithm`
- `timing`

Отдельный важный исход:

- `constraint`

Как использовать в проекте:

- наш `core-backend` строит payload из готового aggregated profile;
- payload маппится в `incommingParameters`;
- через `outputParameters` запрашиваются только финальные decision outputs;
- ответ WiMi нормализуется в нашу внутреннюю decision DTO.

### `POST /ModelsCalcPackage`

Назначение:

- batch-обработка нескольких расчётов.

Для operator flow не нужен. Может пригодиться для:

- оффлайн регрессии;
- пакетной валидации модели;
- тестовых прогонов исторических обследований.

### `POST /ModelsParametersInfo`

Назначение:

- получить перечень параметров модели.

Вход:

- `modelID`
- `allowedFields`
  - `type`
  - `description`
  - `defaultValue`

Это критично для интеграции.

Именно через этот endpoint нужно:

- снять реальный список входных параметров WiMi;
- понять типы параметров;
- проверить соответствие нашего aggregated profile decision-модели;
- построить mapping `our_field -> WiMi parameter ID`.

### `POST /RelationParse`

Назначение:

- парсинг и проверка формул.

Для runtime path проекта не нужен.

### `POST /DecodeMML`

Назначение:

- декодирование модели в XML.

Для runtime path проекта не нужен.

---

## Ошибки WiMi

WiMi использует обычный HTTP и доменные error codes в JSON.

Подтверждены важные группы:

- `510x` — ошибки вычисления модели;
- `520x` — ошибки удаления модели;
- `540x` — ошибки парсинга relation;
- `550x` — ошибки информации о модели;
- `560x` — ошибки загрузки модели;
- `570x` — ошибки decode/package;
- `5800` — ошибки `ModelsParametersInfo`.

Особенно важные для Phase 4:

- `5101 NoModelID`
- `5102 NoModelOnServer`
- `5103 BadIncommingParameters`
- `5104 BadOutputParameters`
- `5105 ElementNotFound`
- `5106 TypesMismatch`
- `5107 InitScript`
- `5108 LoopError`
- `constraint`
- `No model in pool, all model are busy`

Рекомендуемая классификация в нашем адаптере:

### Retryable transport/temporary errors

- timeout;
- network/DNS/TCP ошибки;
- `5xx`;
- `all model are busy`.

Для них:

- допустим retry;
- обследование остаётся в `decision_pending`;
- нужен retry budget и backoff.

### Non-retryable business/contract errors

- `NoModelOnServer`;
- `BadIncommingParameters`;
- `BadOutputParameters`;
- `ElementNotFound`;
- `TypesMismatch`;
- `constraint`.

Для них:

- retry не нужен;
- ошибка должна фиксироваться как terminal integration/business failure;
- оператор должен видеть понятную причину, а не просто "WiMi недоступен".

---

## Что нужно сделать для интеграции в проект

### 1. Поднять реальный экземпляр WiMi

Нужно:

- развернуть `WiMi-0.1.7.deb` на отдельном внутреннем Linux-хосте;
- зафиксировать рабочий порт;
- ограничить сетевой доступ только для `core-backend`;
- проверить `GET /Models`.

### 2. Получить реальную decision-модель

Без конкретной модели полноценной интеграции не будет.

Нужно получить:

- `modelID`;
- список входных параметров;
- список выходных параметров;
- типы параметров;
- семантику recommendation outputs.

### 3. Выполнить parameter discovery

Через `POST /ModelsParametersInfo` нужно зафиксировать:

- все WiMi parameter IDs;
- их типы;
- default values;
- описания, если есть.

### 4. Построить mapping из нашего aggregated profile

Нужно ввести отдельный mapping-слой:

- наш канонический decision field -> `WiMi parameter ID`;
- преобразование значения в ожидаемый тип;
- versioning mapping-конфигурации.

Это обязательно, потому что WiMi использует opaque parameter IDs, а не человекочитаемые доменные имена.

До появления финальной модели discovery и mapping можно оставить как deferred integration step, но все остальные части адаптера должны разрабатываться уже сейчас против внутреннего `decision input` DTO и тестовых fixtures.

### 5. Реализовать adapter в `core-backend`

Рекомендуемая структура:

- `core-backend/internal/kesmi/` или `core-backend/internal/wimi/`

Внутри:

- `client.go`
- `contracts.go`
- `mapper.go`
- `service.go`
- `repository.go`

### 6. Добавить persistence под decision flow

Нужно хранить:

- `examination_id`
- `correlation_id`
- `request_payload_version`
- `external_model_id`
- `attempt_no`
- `transport_status`
- `business_status`
- `recommendation`
- `error_code`
- `error_name`
- `error_description`
- `requested_at`
- `responded_at`

Это требуется ТЗ, потому что решение КЭСМИ должно иметь версию payload и timestamp.

### 7. Расширить workflow и UI

По `docs/00_project.md` нужны статусы:

- `decision_pending`
- `completed`

Оператор должен видеть:

- итоговую recommendation;
- integration diagnostics;
- correlation ID;
- причину integration error, если вызов неуспешен.

Даже до появления реальной decision-модели UI и backend workflow можно довести до готовности на stub данных, если recommendation rendering, diagnostics и статусы будут опираться на наш внутренний `decision result`, а не на сырой WiMi response.

---

## Рекомендуемый runtime path

Для одного обследования:

1. Phase 3 формирует канонический aggregated profile.
2. `core-backend` переводит обследование в интеграционный этап.
3. WiMi adapter маппит профиль в `ModelCalc` request.
4. WiMi возвращает outcome.
5. Адаптер классифицирует ответ:
   - success
   - retryable temporary failure
   - terminal business failure
6. В БД сохраняются решение, диагностика и metadata.
7. Оператор получает recommendation или понятную integration error.

---

## Главный риск интеграции

Главный риск не в HTTP и не в запуске WiMi, а в отсутствии зафиксированного соответствия:

- наш aggregated profile
- конкретные входы decision-модели WiMi
- конкретные выходы recommendation

Поэтому первый обязательный практический шаг Phase 4:

- не писать клиент вслепую,
- а сначала снять параметры реальной модели через `ModelsParametersInfo` и утвердить mapping.

Но это не означает, что сама Phase 4 должна ждать готовой модели. Ожидание должно касаться только:

- финального feature mapping;
- выбора реальных `outputParameters`;
- бизнес-семантики recommendation;
- финального contract verification против живой модели.

---

## Вывод

Для нашего проекта WiMi подходит как внешняя СППР, если использовать его через изолированный backend adapter.

Минимально достаточная схема интеграции:

- WiMi развёрнут отдельно;
- `core-backend` ходит в него через `POST /ModelCalc`;
- contract discovery делается через `POST /ModelsParametersInfo`;
- transport и business errors различаются явно;
- recommendation и diagnostics сохраняются в PostgreSQL и отображаются оператору.

Открытая документация по КЭСМИ полезна только как вспомогательная. Источником истины для реальной Phase 4 интеграции должен стать конкретный экземпляр WiMi и проверка его API на живой decision-модели.
