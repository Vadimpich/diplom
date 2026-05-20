import os
import tempfile
from dataclasses import dataclass
from pathlib import Path
from typing import Any


@dataclass(frozen=True)
class STTConfig:
    model_path: str
    model_size: str
    device: str
    compute_type: str


class STTError(Exception):
    pass


class FasterWhisperSTT:
    def __init__(self, config: STTConfig) -> None:
        self.config = config
        self._model: Any | None = None
        self._model_version = f"faster-whisper-{config.model_size}"

    @property
    def model_version(self) -> str:
        return self._model_version

    def transcribe_bytes(self, audio: bytes, *, suffix: str = ".webm") -> dict[str, Any]:
        if not audio:
            raise STTError("empty audio object")

        model = self._load_model()
        with tempfile.NamedTemporaryFile(suffix=suffix) as tmp:
            tmp.write(audio)
            tmp.flush()
            try:
                segments_iter, info = model.transcribe(
                    tmp.name,
                    beam_size=5,
                    vad_filter=True,
                    word_timestamps=False,
                )
                segments = list(segments_iter)
            except Exception as exc:
                raise STTError(f"transcription failed: {exc}") from exc

        transcript = " ".join(segment.text.strip() for segment in segments if segment.text.strip()).strip()
        normalized_segments = [
            {
                "start_ms": int(round(float(segment.start) * 1000)),
                "end_ms": int(round(float(segment.end) * 1000)),
                "text": segment.text.strip(),
            }
            for segment in segments
        ]
        language = getattr(info, "language", None) or None
        duration = getattr(info, "duration", None)
        return {
            "transcript": transcript,
            "language": language,
            "segments": normalized_segments,
            "duration_ms": int(round(float(duration) * 1000)) if duration is not None else None,
            "word_count": count_words(transcript),
            "model_version": self.model_version,
        }

    def _load_model(self) -> Any:
        if self._model is not None:
            return self._model

        model_name_or_path = self._resolve_model_name_or_path()
        try:
            from faster_whisper import WhisperModel

            self._model = WhisperModel(
                model_name_or_path,
                device=self.config.device,
                compute_type=self.config.compute_type,
                local_files_only=offline_mode_enabled(),
            )
        except Exception as exc:
            raise STTError(f"model load failed: {exc}") from exc
        return self._model

    def _resolve_model_name_or_path(self) -> str:
        if self.config.model_path:
            path = Path(self.config.model_path)
            if local_model_available(path):
                self._model_version = f"faster-whisper-local-{path.name}"
                return str(path)
            if offline_mode_enabled():
                raise STTError(f"STT_MODEL_PATH is missing or incomplete in offline mode: {self.config.model_path}")
        return self.config.model_size


def load_stt_config() -> STTConfig:
    return STTConfig(
        model_path=os.getenv("STT_MODEL_PATH", "models/stt"),
        model_size=os.getenv("STT_MODEL_SIZE", "medium"),
        device=os.getenv("STT_DEVICE", "cpu"),
        compute_type=os.getenv("STT_COMPUTE_TYPE", "int8"),
    )


def offline_mode_enabled() -> bool:
    return os.getenv("HF_HUB_OFFLINE", "").strip().lower() in {"1", "true", "yes", "on"}


def local_model_available(path: Path) -> bool:
    return path.is_dir() and (path / "config.json").is_file()


def count_words(text: str) -> int:
    return len([part for part in text.replace("\n", " ").split(" ") if part.strip()])


def failed_stt_result(message: str, *, model_version: str) -> dict[str, Any]:
    return {
        "transcript": "",
        "language": None,
        "segments": [],
        "duration_ms": None,
        "word_count": 0,
        "model_version": model_version,
        "error": message,
    }
