#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MODELS_DIR="${ROOT_DIR}/models"

STT_MODEL_ID="${STT_MODEL_ID:-Systran/faster-whisper-medium}"
TEXT_EMOTION_MODEL_ID="${TEXT_EMOTION_MODEL_ID:-seara/rubert-base-cased-russian-emotion-detection-ru-go-emotions}"
ACOUSTIC_MODEL_ID="${ACOUSTIC_MODEL_ID:-xbgoose/hubert-large-speech-emotion-recognition-russian-dusha-finetuned}"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

download_model() {
  local model_id="$1"
  local target_dir="$2"
  echo "Downloading ${model_id} -> ${target_dir}"
  hf download "${model_id}" \
    --local-dir "${target_dir}"
}

require_cmd hf

mkdir -p "${MODELS_DIR}/stt" "${MODELS_DIR}/text-emotion" "${MODELS_DIR}/acoustic-emotion"

download_model "${STT_MODEL_ID}" "${MODELS_DIR}/stt"
download_model "${TEXT_EMOTION_MODEL_ID}" "${MODELS_DIR}/text-emotion"
download_model "${ACOUSTIC_MODEL_ID}" "${MODELS_DIR}/acoustic-emotion"

echo
echo "Models are ready in ${MODELS_DIR}:"
find "${MODELS_DIR}" -maxdepth 2 -name 'config.json' -o -name 'model.bin' | sort
