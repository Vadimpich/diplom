# diplom self-hosted stack

Self-hosted стек мультимодальной системы оценки психоэмоционального состояния специалистов.

В состав локального контура входят:

- `frontend` — Next.js UI для оператора и администратора;
- `core-backend` — Go orchestration/API слой;
- `text-worker`, `acoustic-worker`, `paralinguistic-worker` — ML-каналы;
- `ml-baseline` — baseline/deviation compute service;
- `wimi` — internal-only decision engine runtime;
- `postgres`, `rabbitmq`, `minio` — инфраструктура хранения и очередей.

## Что нужно для развёртывания

- Docker Engine
- Docker Compose v2
- Git
- для первого online-setup: доступ в интернет для загрузки ML-моделей

Рекомендуемо для комфортного локального запуска:

- `8+ CPU`
- `16 GB RAM`
- `15–25 GB` свободного места под образы, модели и volumes

## Профили окружения

В репозитории есть готовые шаблоны env:

- `ops/env/local-dev.env.example`  
  Локальная разработка на одной машине, разрешён online fallback для моделей.

- `ops/env/offline-demo.env.example`  
  Полностью офлайн/air-gapped demo-режим. Все модели должны уже лежать в `./models`.

- `ops/env/lan-demo.env.example`  
  Демо с доступом с другого устройства в одной сети. Нужно заменить `NEXT_PUBLIC_API_URL` на IP хоста.

Быстрый выбор профиля:

```bash
cp ops/env/local-dev.env.example .env
```

или:

```bash
cp ops/env/offline-demo.env.example .env
```

или:

```bash
cp ops/env/lan-demo.env.example .env
```

После этого отредактируйте секреты и сетевые значения в `.env`.

## Модели

Для полного запуска нужны три локальные модели:

1. **STT**
   - model id: `Systran/faster-whisper-medium`
   - target dir: `models/stt`

2. **Текстовая emotion-модель**
   - model id: `seara/rubert-base-cased-russian-emotion-detection-ru-go-emotions`
   - target dir: `models/text-emotion`

3. **Акустическая SER-модель**
   - model id: `xbgoose/hubert-large-speech-emotion-recognition-russian-dusha-finetuned`
   - target dir: `models/acoustic-emotion`

Decision-модель WiMi уже хранится в репозитории:

- `wimi-server/models/specialists_model_v2_decision.xml`

### Вариант A: перенести готовые модели с другой машины

Если на исходной машине `models/` уже заполнена и проверена, просто перенесите каталог целиком:

```bash
rsync -av models/ user@target:/path/to/diplom/models/
```

или любым другим способом копирования.

### Вариант B: скачать модели на новой машине

Подготовьте `huggingface_hub` CLI:

```bash
python3 -m venv .venv-models
source .venv-models/bin/activate
pip install -U pip huggingface_hub
```

Актуальный CLI команды `huggingface_hub`:

```bash
hf --help
```

Затем скачайте все модели одним скриптом:

```bash
chmod +x ops/setup/download-models.sh
./ops/setup/download-models.sh
```

По умолчанию скрипт скачивает:

- `Systran/faster-whisper-medium`
- `seara/rubert-base-cased-russian-emotion-detection-ru-go-emotions`
- `xbgoose/hubert-large-speech-emotion-recognition-russian-dusha-finetuned`

Если нужен ручной режим, эквивалентные команды такие:

```bash
hf download Systran/faster-whisper-medium \
  --local-dir models/stt

hf download seara/rubert-base-cased-russian-emotion-detection-ru-go-emotions \
  --local-dir models/text-emotion

hf download xbgoose/hubert-large-speech-emotion-recognition-russian-dusha-finetuned \
  --local-dir models/acoustic-emotion
```

Проверка, что модели на месте:

```bash
find models -maxdepth 2 \( -name config.json -o -name model.bin \) | sort
```

Если вы используете offline-профиль, убедитесь, что в `.env` стоит:

```bash
HF_HUB_OFFLINE=1
```

## Полноценное развёртывание на новой машине

1. Клонировать репозиторий:

```bash
git clone <repo-url> diplom
cd diplom
```

2. Выбрать env-профиль:

```bash
cp ops/env/offline-demo.env.example .env
```

3. Подготовить модели:

- либо перенести каталог `models/` с уже настроенной машины;
- либо скачать модели через `./ops/setup/download-models.sh`.

4. Поднять весь стек:

```bash
docker compose up -d --build
```

5. Проверить состояние контейнеров:

```bash
docker compose ps
```

6. Проверить базовые endpoints:

```bash
curl -fsS http://localhost:3000/api/health
curl -fsS http://localhost:3000/api/ready
curl -fsS http://localhost:18080/health
curl -fsS http://localhost:18080/ready
```

7. Проверить загрузку WiMi-модели изнутри compose-сети:

```bash
./wimi-server/scripts/smoke.sh
```

## Доступные порты

- PostgreSQL: `5432`
- RabbitMQ AMQP: `5672`
- RabbitMQ management: `15672`
- MinIO API: `9000`
- MinIO console: `9001`
- core backend: `18080` -> container `8080`
- frontend: `3000`

`ml-baseline` и `wimi` остаются internal-only сервисами Compose.

## Важные env-переменные

Наиболее важные группы:

- `POSTGRES_*`
- `RABBITMQ_*`
- `MINIO_*`
- `HTTP_ADDR`, `CORE_BACKEND_*`
- `NEXT_PUBLIC_API_URL`, `INTERNAL_API_BASE_URL`
- `STT_*`
- `TEXT_EMOTION_*`
- `ACOUSTIC_*`
- `PARALINGUISTIC_*`
- `BASELINE_*`
- `KESMI_*`
- `JWT_*`

Ключевые runtime-значения:

- `KESMI_BASE_URL=http://wimi:8081`
- `WIMI_AUTOLOAD_MODEL_PATH=/opt/wimi-models/specialists_model_v2_decision.xml`
- `NEXT_PUBLIC_API_URL` должен быть доступен браузеру
- `INTERNAL_API_BASE_URL` внутри compose должен оставаться `http://core-backend:8080`

## Проверка observability runtime

Минимальный smoke:

```bash
docker compose up -d --build frontend core-backend text-worker acoustic-worker paralinguistic-worker ml-baseline wimi
curl -fsS http://localhost:3000/api/health
curl -fsS http://localhost:3000/api/ready
curl -fsS http://localhost:3000/api/metrics
curl -fsS http://localhost:18080/health
curl -fsS http://localhost:18080/ready
curl -fsS http://localhost:18080/metrics
```

## Репликация ML-воркеров

Для умеренной параллельной нагрузки можно масштабировать именно ML-воркеры через обычный Docker Compose.  
Swarm или Kubernetes для этого не нужны.

Подходящий сценарий:

- несколько операторов одновременно запускают обследования;
- `text-worker`, `acoustic-worker`, `paralinguistic-worker` обрабатывают задачи из своих RabbitMQ-очередей параллельно;
- `core-backend`, `frontend`, `postgres`, `rabbitmq`, `minio`, `ml-baseline`, `wimi` остаются в одной реплике.

### Что можно масштабировать

Без изменения архитектуры безопасно масштабировать:

- `text-worker`
- `acoustic-worker`
- `paralinguistic-worker`

Пример запуска с тремя репликами каждого канала:

```bash
docker compose up -d \
  --scale text-worker=3 \
  --scale acoustic-worker=3 \
  --scale paralinguistic-worker=3 \
  text-worker acoustic-worker paralinguistic-worker
```

Или вместе с основным стеком:

```bash
docker compose up -d --build \
  --scale text-worker=3 \
  --scale acoustic-worker=3 \
  --scale paralinguistic-worker=3
```

### Что не стоит масштабировать без доработок

Пока лучше оставлять в одной реплике:

- `core-backend`
- `frontend`
- `ml-baseline`
- `wimi`
- `postgres`
- `rabbitmq`
- `minio`

Причина: текущая схема безопасно поддерживает многопотребительскую обработку именно на уровне ML-очередей.  
Горизонтальное масштабирование `core-backend` потребует отдельной доработки outbox/publisher-логики.

### Что важно понимать

- Реплики ускоряют обработку нескольких обследований одновременно.
- Реплики не ускоряют один конкретный канал внутри одного обследования, потому что одно сообщение канала обрабатывается одним worker-экземпляром целиком.
- На практике первым bottleneck обычно становится `text-worker`, потому что он включает STT и text emotion inference.

### Как проверить, что реплики реально работают

Посмотреть поднятые контейнеры:

```bash
docker compose ps
```

Проверить число consumers в RabbitMQ:

```bash
docker compose exec -T rabbitmq rabbitmqctl list_consumers \
  --no-table-headers queue_name consumer_tag ack_required prefetch_count | sort
```

Для трёх реплик каждого ML-канала в выводе должно быть:

- `3` consumers на `qq.processing.text`
- `3` consumers на `qq.processing.acoustic`
- `3` consumers на `qq.processing.paralinguistic`

Также можно смотреть логи воркеров:

```bash
docker compose logs -f text-worker acoustic-worker paralinguistic-worker
```

В логах появятся события:

- `processing_started`
- `audio_objects_loaded`
- `processing_succeeded`
- `result_published`

По ним можно увидеть, какой именно контейнер взял конкретное обследование и сколько времени заняла обработка.

## Полная локальная проверка

```bash
docker compose up -d --build
docker compose ps
cd /home/vadim/diplom/core-backend && go test ./... -count=1
cd /home/vadim/diplom/frontend && npm run lint && npm run typecheck && npm test && npm run build
cd /home/vadim/diplom/ml-services/ml-text && docker run --rm -v "$PWD":/app -w /app python:3.12-slim sh -c 'pip install --no-cache-dir -r requirements.txt >/tmp/pip.log && PYTHONPATH=. pytest -q'
cd /home/vadim/diplom/ml-services/ml-acoustic && docker run --rm -v "$PWD":/app -w /app python:3.12-slim sh -c 'pip install --no-cache-dir -r requirements.txt >/tmp/pip.log && PYTHONPATH=. pytest -q'
cd /home/vadim/diplom/ml-services/ml-paralinguistic && docker run --rm -v "$PWD":/app -w /app python:3.12-slim sh -c 'pip install --no-cache-dir -r requirements.txt >/tmp/pip.log && PYTHONPATH=. pytest -q'
cd /home/vadim/diplom/ml-services/ml-baseline && docker run --rm -v "$PWD":/app -w /app python:3.12-slim sh -c 'pip install --no-cache-dir -r requirements.txt >/tmp/pip.log && PYTHONPATH=. pytest -q'
```

## Примечания

- `wimi` не публикуется наружу и доступен только внутри compose-сети.
- Если volumes RabbitMQ/PostgreSQL остались от старых несовместимых прогонов, перед чистым стартом можно выполнить:

```bash
docker compose down -v
```

- Если браузер открывает frontend с другой машины, не забудьте выставить корректный `NEXT_PUBLIC_API_URL` в `.env` до сборки контейнера frontend.
