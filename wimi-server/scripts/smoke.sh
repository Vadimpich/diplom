#!/usr/bin/env sh
set -eu

docker compose up -d --build wimi core-backend
docker compose exec core-backend sh -lc 'wget -qO- http://wimi:8081/Models >/dev/null'

for _ in $(seq 1 30); do
    if [ "$(docker inspect --format '{{.State.Health.Status}}' diplom-core-backend-1 2>/dev/null || true)" = "healthy" ]; then
        break
    fi
    sleep 1
done

curl --retry 10 --retry-delay 1 --retry-connrefused -fsS http://localhost:8080/health >/dev/null
