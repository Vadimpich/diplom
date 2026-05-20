import math
import os
import subprocess
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Callable

import numpy as np


DEFAULT_MODEL_ID = "xbgoose/hubert-large-speech-emotion-recognition-russian-dusha-finetuned"
MODEL_VERSION = "xbgoose-hubert-large-dusha-v1"
SAMPLE_RATE = 16_000
EMOTION_KEYS = ("neutral", "happiness", "sadness", "anger", "fear", "other")


@dataclass(frozen=True)
class AnalysisConfig:
    model_path: str = "/app/models/acoustic-emotion"
    model_id: str = DEFAULT_MODEL_ID
    device: str = "cpu"
    min_audio_duration_ms: int = 1_000
    low_volume_rms_threshold: float = 0.005


class AudioDecodeError(Exception):
    pass


class AcousticInferenceError(Exception):
    def __init__(self, flag: str, message: str) -> None:
        super().__init__(message)
        self.flag = flag
        self.message = message


class AcousticSERClassifier:
    def __init__(
        self,
        config: AnalysisConfig,
        *,
        pipeline_factory: Callable[..., Any] | None = None,
    ) -> None:
        self.config = config
        self._pipeline_factory = pipeline_factory
        self._pipeline: Any | None = None
        self._model_version = MODEL_VERSION

    @property
    def model_version(self) -> str:
        return self._model_version

    def predict(self, samples: np.ndarray, *, sample_rate: int) -> dict[str, float]:
        classifier = self._load()
        try:
            predictions = classifier({"array": samples.astype(np.float32), "sampling_rate": sample_rate}, top_k=None)
        except Exception as exc:
            raise AcousticInferenceError("inference_failed", f"emotion inference failed: {exc}") from exc

        if isinstance(predictions, dict):
            predictions = [predictions]

        probs = {key: 0.0 for key in EMOTION_KEYS}
        for item in predictions:
            normalized_label = normalize_emotion_label(str(item.get("label", "")))
            probability = float(item.get("score", 0.0))
            if normalized_label in probs:
                probs[normalized_label] += probability

        total = sum(probs.values())
        if total > 0:
            probs = {key: float(value / total) for key, value in probs.items()}
        return probs

    def _load(self) -> Any:
        if self._pipeline is not None:
            return self._pipeline

        model_name_or_path = self._resolve_model_name_or_path()
        try:
            pipeline_factory = self._pipeline_factory
            if pipeline_factory is None:
                from transformers import pipeline

                pipeline_factory = pipeline
            self._pipeline = pipeline_factory(
                "audio-classification",
                model=model_name_or_path,
                device=device_index(self.config.device),
            )
        except Exception as exc:
            raise AcousticInferenceError("model_load_failed", f"emotion model load failed: {exc}") from exc
        return self._pipeline

    def _resolve_model_name_or_path(self) -> str:
        if self.config.model_path:
            path = Path(self.config.model_path)
            if local_model_available(path):
                self._model_version = f"{MODEL_VERSION}-local-{path.name}"
                return str(path)
            if offline_mode_enabled():
                raise AcousticInferenceError(
                    "model_load_failed",
                    f"ACOUSTIC_MODEL_PATH is missing or incomplete in offline mode: {path}",
                )
        return self.config.model_id


def analyze_audio_bytes(
    audio: bytes,
    *,
    config: AnalysisConfig | None = None,
    classifier: AcousticSERClassifier | None = None,
) -> dict[str, Any]:
    active_config = config or load_analysis_config()
    active_classifier = classifier or AcousticSERClassifier(active_config)
    try:
        samples = decode_audio_to_float32(audio)
        return analyze_samples(samples, sample_rate=SAMPLE_RATE, config=active_config, classifier=active_classifier)
    except AudioDecodeError as exc:
        return failed_result("unsupported_audio_format", str(exc), model_version=active_classifier.model_version)
    except AcousticInferenceError as exc:
        return failed_result(exc.flag, exc.message, model_version=active_classifier.model_version)


def decode_audio_to_float32(audio: bytes) -> np.ndarray:
    if not audio:
        raise AudioDecodeError("empty audio object")

    command = [
        "ffmpeg",
        "-hide_banner",
        "-loglevel",
        "error",
        "-i",
        "pipe:0",
        "-ac",
        "1",
        "-ar",
        str(SAMPLE_RATE),
        "-f",
        "f32le",
        "pipe:1",
    ]
    try:
        completed = subprocess.run(
            command,
            input=audio,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
            timeout=30,
        )
    except FileNotFoundError as exc:
        raise AudioDecodeError("ffmpeg is not installed") from exc
    except subprocess.TimeoutExpired as exc:
        raise AudioDecodeError("ffmpeg decode timeout") from exc

    if completed.returncode != 0:
        error = completed.stderr.decode("utf-8", errors="replace").strip()
        raise AudioDecodeError(error or "ffmpeg decode failed")
    if not completed.stdout:
        raise AudioDecodeError("decoded audio is empty")

    samples = np.frombuffer(completed.stdout, dtype=np.float32)
    if samples.size == 0:
        raise AudioDecodeError("decoded audio has no samples")
    return sanitize_samples(samples)


def analyze_samples(
    samples: np.ndarray,
    *,
    sample_rate: int = SAMPLE_RATE,
    config: AnalysisConfig | None = None,
    classifier: AcousticSERClassifier | None = None,
) -> dict[str, Any]:
    active_config = config or load_analysis_config()
    active_classifier = classifier or AcousticSERClassifier(active_config)
    y = sanitize_samples(samples)
    duration_ms = int(round(len(y) / sample_rate * 1000))
    rms_mean, rms_std = rms_statistics(y)
    intensity_variability = safe_ratio(rms_std, rms_mean)

    quality_flags = quality_flags_for(
        duration_ms=duration_ms,
        rms_energy_mean=rms_mean,
        config=active_config,
    )
    emotion_probs = active_classifier.predict(y, sample_rate=sample_rate)
    dominant_emotion, dominant_probability = dominant_emotion_for(emotion_probs)

    features = {
        "duration_ms": duration_ms,
        "sample_rate": sample_rate,
        "rms_energy_mean": round6(rms_mean),
    }
    scores = calculate_scores(emotion_probs, intensity_variability=intensity_variability)
    evidence = build_evidence(
        features,
        scores,
        emotion_probs,
        dominant_emotion=dominant_emotion,
        dominant_probability=dominant_probability,
        quality_flags=quality_flags,
    )

    return {
        "channel": "acoustic",
        "status": "done",
        "features": features,
        "scores": scores,
        "emotion_probs": round_probabilities(emotion_probs),
        "dominant_emotion": dominant_emotion,
        "quality_flags": quality_flags,
        "evidence": evidence,
        "model_version": active_classifier.model_version,
    }


def load_analysis_config() -> AnalysisConfig:
    return AnalysisConfig(
        model_path=os.getenv("ACOUSTIC_MODEL_PATH", "/app/models/acoustic-emotion"),
        model_id=os.getenv("ACOUSTIC_MODEL_ID", DEFAULT_MODEL_ID),
        device=os.getenv("ACOUSTIC_DEVICE", "cpu"),
    )


def local_model_available(path: Path) -> bool:
    return path.is_dir() and (path / "config.json").is_file()


def offline_mode_enabled() -> bool:
    return os.getenv("HF_HUB_OFFLINE", "").strip().lower() in {"1", "true", "yes", "on"}


def device_index(device: str) -> int:
    return 0 if device.strip().lower() == "cuda" else -1


def normalize_emotion_label(label: str) -> str:
    normalized = label.strip().lower().replace("-", "_")
    mapping = {
        "neutral": "neutral",
        "нейтральный": "neutral",
        "angry": "anger",
        "anger": "anger",
        "злость": "anger",
        "гнев": "anger",
        "positive": "happiness",
        "joy": "happiness",
        "happy": "happiness",
        "happiness": "happiness",
        "радость": "happiness",
        "sad": "sadness",
        "sadness": "sadness",
        "грусть": "sadness",
        "fear": "fear",
        "страх": "fear",
        "other": "other",
    }
    return mapping.get(normalized, "other")


def rms_statistics(samples: np.ndarray, *, frame_length: int = 400, hop_length: int = 160) -> tuple[float, float]:
    if samples.size == 0:
        return 0.0, 0.0

    if samples.size < frame_length:
        rms = float(np.sqrt(np.mean(np.square(samples), dtype=np.float64)))
        return rms, 0.0

    frame_values: list[float] = []
    for start in range(0, samples.size - frame_length + 1, hop_length):
        frame = samples[start : start + frame_length]
        frame_values.append(float(np.sqrt(np.mean(np.square(frame), dtype=np.float64))))
    values = np.asarray(frame_values, dtype=np.float64)
    return float(np.mean(values)), float(np.std(values))


def dominant_emotion_for(emotion_probs: dict[str, float]) -> tuple[str, float]:
    if not emotion_probs:
        return "neutral", 0.0
    emotion, probability = max(emotion_probs.items(), key=lambda item: item[1])
    return emotion, float(probability)


def calculate_scores(emotion_probs: dict[str, float], *, intensity_variability: float) -> dict[str, float]:
    intensity_variability_score = clamp01(intensity_variability / 1.25)
    negative_signal = clamp01(
        emotion_probs.get("anger", 0.0) * 1.0
        + emotion_probs.get("sadness", 0.0) * 0.8
        + emotion_probs.get("fear", 0.0) * 0.9
        + emotion_probs.get("other", 0.0) * 0.5
    )
    positive_buffer = emotion_probs.get("happiness", 0.0) * 0.7 + emotion_probs.get("neutral", 0.0) * 0.45
    acoustic_stress_score = clamp01(
        0.70 * negative_signal
        + 0.15 * intensity_variability_score
        + 0.15 * (1.0 - positive_buffer)
    )
    voice_stability_score = 1.0 - clamp01(
        0.55 * intensity_variability_score
        + 0.30 * negative_signal
        + 0.15 * (1.0 - emotion_probs.get("neutral", 0.0))
    )
    return {
        "acoustic_stress_score": round4(acoustic_stress_score),
        "voice_stability_score": round4(voice_stability_score),
        "intensity_variability_score": round4(intensity_variability_score),
    }


def quality_flags_for(*, duration_ms: int, rms_energy_mean: float, config: AnalysisConfig) -> list[str]:
    flags: list[str] = []
    if duration_ms < config.min_audio_duration_ms:
        flags.append("audio_too_short")
    if rms_energy_mean < config.low_volume_rms_threshold:
        flags.append("low_volume")
    return flags


def build_evidence(
    features: dict[str, Any],
    scores: dict[str, float],
    emotion_probs: dict[str, float],
    *,
    dominant_emotion: str,
    dominant_probability: float,
    quality_flags: list[str],
) -> list[str]:
    evidence = [
        f"Доминирующая эмоция: {dominant_emotion} ({round(dominant_probability * 100, 1)}%).",
        f"Средняя энергия RMS: {features['rms_energy_mean']}.",
        f"Индекс акустического напряжения: {scores['acoustic_stress_score']}.",
        f"Индекс стабильности голоса: {scores['voice_stability_score']}.",
    ]
    if emotion_probs.get("anger", 0.0) + emotion_probs.get("sadness", 0.0) + emotion_probs.get("fear", 0.0) >= 0.45:
        evidence.append("Модель зафиксировала выраженную долю негативных эмоциональных классов.")
    if quality_flags:
        evidence.append(f"Флаги качества: {', '.join(quality_flags)}.")
    return evidence


def failed_result(flag: str, message: str, *, model_version: str = MODEL_VERSION) -> dict[str, Any]:
    return {
        "channel": "acoustic",
        "status": "done",
        "features": {
            "duration_ms": 0,
            "sample_rate": SAMPLE_RATE,
            "rms_energy_mean": 0.0,
        },
        "scores": {
            "acoustic_stress_score": 0.0,
            "voice_stability_score": 0.0,
            "intensity_variability_score": 0.0,
        },
        "emotion_probs": {key: 0.0 for key in EMOTION_KEYS},
        "dominant_emotion": "neutral",
        "quality_flags": [flag],
        "evidence": [f"Акустический анализ не выполнен: {message}."],
        "model_version": model_version,
    }


def sanitize_samples(samples: np.ndarray) -> np.ndarray:
    y = np.asarray(samples, dtype=np.float32)
    y = y[np.isfinite(y)]
    if y.size == 0:
        raise AudioDecodeError("audio contains no finite samples")
    return np.clip(y, -1.0, 1.0)


def safe_ratio(value: float, total: float) -> float:
    if total <= 0:
        return 0.0
    return float(value) / float(total)


def clamp01(value: float) -> float:
    if math.isnan(value):
        return 0.0
    return max(0.0, min(1.0, float(value)))


def round4(value: float) -> float:
    return round(float(value), 4)


def round6(value: float) -> float:
    return round(float(value), 6)


def round_probabilities(probabilities: dict[str, float]) -> dict[str, float]:
    return {key: round4(probabilities.get(key, 0.0)) for key in EMOTION_KEYS}
