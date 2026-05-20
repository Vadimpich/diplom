#!/usr/bin/env sh
set -eu

export LD_LIBRARY_PATH="${LD_LIBRARY_PATH:+$LD_LIBRARY_PATH:}/usr/local/bin/WiMi/libs"
export LD_PRELOAD="${LD_PRELOAD:+$LD_PRELOAD:}/usr/local/bin/WiMi/libs/libjemalloc.so.2"

wait_for_wimi() {
  retries="${1:-60}"
  while [ "$retries" -gt 0 ]; do
    if wget -qO- http://127.0.0.1:8081/Models >/dev/null 2>&1; then
      return 0
    fi
    retries=$((retries - 1))
    sleep 1
  done
  return 1
}

escape_json() {
  sed 's/\\/\\\\/g; s/"/\\"/g'
}

autoload_model() {
  model_path="${WIMI_AUTOLOAD_MODEL_PATH:-}"
  model_id="${WIMI_AUTOLOAD_MODEL_ID:-${KESMI_MODEL_ID:-}}"

  if [ -z "$model_path" ] || [ -z "$model_id" ]; then
    return 0
  fi
  if [ ! -f "$model_path" ]; then
    echo "WiMi autoload: model file not found: $model_path" >&2
    return 1
  fi
  if wget -qO- http://127.0.0.1:8081/Models | grep -F "\"$model_id\"" >/dev/null 2>&1; then
    echo "WiMi autoload: model already loaded: $model_id"
    return 0
  fi

  model_xml="$(tr -d '\r\n' < "$model_path" | escape_json)"
  payload="$(printf '{"modelID":"%s","modelPoolSize":5,"modelXML":"%s"}' "$model_id" "$model_xml")"

  echo "WiMi autoload: loading model $model_id from $model_path"
  wget -qO- \
    --header="Content-Type: application/json" \
    --post-data="$payload" \
    http://127.0.0.1:8081/Models >/dev/null
}

cd /usr/local/bin/WiMi
./WiMi -e &
pid="$!"

trap 'kill "$pid" 2>/dev/null || true; wait "$pid" 2>/dev/null || true' INT TERM EXIT

wait_for_wimi
autoload_model

wait "$pid"
