import math
import subprocess
from dataclasses import dataclass
from typing import Any

import webrtcvad


MODEL_VERSION = "paralinguistic-vad-v1"
SAMPLE_RATE = 16_000
SAMPLE_WIDTH_BYTES = 2
FRAME_DURATION_MS = 30


@dataclass(frozen=True)
class AnalysisConfig:
    long_pause_threshold_ms: int = 1_200
    min_speech_duration_ms: int = 700
    low_speech_ratio_threshold: float = 0.15
    vad_aggressiveness: int = 2


@dataclass(frozen=True)
class SpeechSegment:
    start_ms: int
    end_ms: int


class AudioDecodeError(Exception):
    pass


class VADAnalysisError(Exception):
    pass


def analyze_audio_bytes(
    audio: bytes,
    *,
    word_count: int | None = None,
    config: AnalysisConfig | None = None,
) -> dict[str, Any]:
    active_config = config or AnalysisConfig()
    try:
        pcm = decode_audio_to_pcm16(audio)
        total_audio_duration_ms = pcm_duration_ms(pcm)
        segments = detect_speech_segments(pcm, config=active_config)
        return build_result_from_segments(
            total_audio_duration_ms=total_audio_duration_ms,
            segments=segments,
            word_count=word_count,
            config=active_config,
        )
    except (AudioDecodeError, VADAnalysisError) as exc:
        return failed_result(str(exc))


def decode_audio_to_pcm16(audio: bytes) -> bytes:
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
        "s16le",
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
    return completed.stdout


def pcm_duration_ms(pcm: bytes) -> int:
    samples = len(pcm) // SAMPLE_WIDTH_BYTES
    return int(round(samples / SAMPLE_RATE * 1000))


def detect_speech_segments(pcm: bytes, *, config: AnalysisConfig) -> list[SpeechSegment]:
    if not pcm:
        return []

    try:
        vad = webrtcvad.Vad(max(0, min(3, config.vad_aggressiveness)))
    except Exception as exc:
        raise VADAnalysisError(f"vad initialization failed: {exc}") from exc
    frame_size = int(SAMPLE_RATE * FRAME_DURATION_MS / 1000) * SAMPLE_WIDTH_BYTES
    if frame_size <= 0:
        raise VADAnalysisError("invalid VAD frame size")

    speech_frames: list[tuple[int, int]] = []
    for offset in range(0, len(pcm) - frame_size + 1, frame_size):
        frame = pcm[offset : offset + frame_size]
        try:
            is_speech = vad.is_speech(frame, SAMPLE_RATE)
        except Exception as exc:
            raise VADAnalysisError(f"vad failed: {exc}") from exc
        if is_speech:
            start_ms = int(round((offset / SAMPLE_WIDTH_BYTES) / SAMPLE_RATE * 1000))
            end_ms = start_ms + FRAME_DURATION_MS
            speech_frames.append((start_ms, end_ms))

    return merge_speech_frames(speech_frames)


def merge_speech_frames(frames: list[tuple[int, int]], *, max_gap_ms: int = FRAME_DURATION_MS) -> list[SpeechSegment]:
    if not frames:
        return []

    merged: list[SpeechSegment] = []
    current_start, current_end = frames[0]
    for start_ms, end_ms in frames[1:]:
        if start_ms - current_end <= max_gap_ms:
            current_end = end_ms
            continue
        merged.append(SpeechSegment(start_ms=current_start, end_ms=current_end))
        current_start, current_end = start_ms, end_ms
    merged.append(SpeechSegment(start_ms=current_start, end_ms=current_end))
    return merged


def build_result_from_segments(
    *,
    total_audio_duration_ms: int,
    segments: list[SpeechSegment],
    word_count: int | None = None,
    config: AnalysisConfig | None = None,
) -> dict[str, Any]:
    active_config = config or AnalysisConfig()
    normalized_segments = normalize_segments(segments, total_audio_duration_ms)
    speech_duration_ms = sum(segment.end_ms - segment.start_ms for segment in normalized_segments)
    silence_duration_ms = max(total_audio_duration_ms - speech_duration_ms, 0)
    speech_ratio = safe_ratio(speech_duration_ms, total_audio_duration_ms)
    pauses = calculate_pauses(normalized_segments)
    long_pauses = [pause for pause in pauses if pause >= active_config.long_pause_threshold_ms]
    response_delay_ms = normalized_segments[0].start_ms if normalized_segments else total_audio_duration_ms
    speech_rate_wpm = calculate_speech_rate_wpm(word_count, speech_duration_ms)
    pause_ratio = safe_ratio(sum(pauses), total_audio_duration_ms)

    features = {
        "total_audio_duration_ms": total_audio_duration_ms,
        "speech_duration_ms": speech_duration_ms,
        "silence_duration_ms": silence_duration_ms,
        "speech_ratio": round4(speech_ratio),
        "response_delay_ms": response_delay_ms,
        "pause_count": len(pauses),
        "long_pause_count": len(long_pauses),
        "mean_pause_ms": round2(sum(pauses) / len(pauses)) if pauses else 0,
        "max_pause_ms": max(pauses) if pauses else 0,
        "speech_segment_count": len(normalized_segments),
        "speech_rate_wpm": speech_rate_wpm,
    }
    scores = {
        "hesitation_score": calculate_hesitation_score(
            pause_ratio=pause_ratio,
            long_pause_count=len(long_pauses),
            response_delay_ms=response_delay_ms,
        ),
        "speech_disorganization_score": calculate_disorganization_score(
            segment_count=len(normalized_segments),
            pause_count=len(pauses),
            speech_ratio=speech_ratio,
            total_audio_duration_ms=total_audio_duration_ms,
        ),
    }
    quality_flags = quality_flags_for(features, active_config)
    evidence = build_evidence(features, scores, quality_flags)
    return {
        "channel": "paralinguistic",
        "status": "done",
        "features": features,
        "scores": scores,
        "quality_flags": quality_flags,
        "evidence": evidence,
        "model_version": MODEL_VERSION,
    }


def normalize_segments(segments: list[SpeechSegment], total_audio_duration_ms: int) -> list[SpeechSegment]:
    normalized: list[SpeechSegment] = []
    for segment in sorted(segments, key=lambda item: item.start_ms):
        start_ms = max(0, min(segment.start_ms, total_audio_duration_ms))
        end_ms = max(start_ms, min(segment.end_ms, total_audio_duration_ms))
        if end_ms > start_ms:
            normalized.append(SpeechSegment(start_ms=start_ms, end_ms=end_ms))
    return normalized


def calculate_pauses(segments: list[SpeechSegment]) -> list[int]:
    pauses: list[int] = []
    for previous, current in zip(segments, segments[1:]):
        pause_ms = current.start_ms - previous.end_ms
        if pause_ms > 0:
            pauses.append(pause_ms)
    return pauses


def calculate_speech_rate_wpm(word_count: int | None, speech_duration_ms: int) -> float | None:
    if word_count is None or word_count <= 0 or speech_duration_ms <= 0:
        return None
    return round2(word_count / (speech_duration_ms / 60_000))


def calculate_hesitation_score(*, pause_ratio: float, long_pause_count: int, response_delay_ms: int) -> float:
    pause_component = clamp01(pause_ratio / 0.8)
    long_pause_component = clamp01(long_pause_count / 6)
    delay_component = clamp01(response_delay_ms / 4_500)
    return round4(0.45 * pause_component + 0.25 * long_pause_component + 0.30 * delay_component)


def calculate_disorganization_score(
    *,
    segment_count: int,
    pause_count: int,
    speech_ratio: float,
    total_audio_duration_ms: int,
) -> float:
    if total_audio_duration_ms <= 0 or segment_count == 0:
        return 0.0
    expected_segments = max(total_audio_duration_ms / 20_000, 1)
    fragmentation_component = clamp01(max(0, segment_count - 1) / (expected_segments * 10))
    pause_component = clamp01(pause_count / (expected_segments * 8))
    low_speech_component = clamp01((0.25 - speech_ratio) / 0.25)
    return round4(0.35 * fragmentation_component + 0.35 * pause_component + 0.30 * low_speech_component)


def quality_flags_for(features: dict[str, Any], config: AnalysisConfig) -> list[str]:
    flags: list[str] = []
    if features["speech_segment_count"] == 0:
        flags.append("no_speech_detected")
    if 0 < features["speech_duration_ms"] < config.min_speech_duration_ms:
        flags.append("too_short_speech")
    if features["speech_ratio"] < config.low_speech_ratio_threshold:
        flags.append("low_speech_ratio")
    return flags


def build_evidence(features: dict[str, Any], scores: dict[str, float], quality_flags: list[str]) -> list[str]:
    evidence = [
        f"Речь занимает {percent(features['speech_ratio'])} аудио.",
        f"Обнаружено пауз между речевыми сегментами: {features['pause_count']}.",
        f"Длинных пауз: {features['long_pause_count']}.",
        f"Задержка до первой речи: {features['response_delay_ms']} мс.",
        f"Индекс нерешительности: {scores['hesitation_score']}.",
    ]
    if quality_flags:
        evidence.append(f"Флаги качества: {', '.join(quality_flags)}.")
    return evidence


def failed_result(message: str) -> dict[str, Any]:
    return {
        "channel": "paralinguistic",
        "status": "done",
        "features": {
            "total_audio_duration_ms": 0,
            "speech_duration_ms": 0,
            "silence_duration_ms": 0,
            "speech_ratio": 0.0,
            "response_delay_ms": 0,
            "pause_count": 0,
            "long_pause_count": 0,
            "mean_pause_ms": 0,
            "max_pause_ms": 0,
            "speech_segment_count": 0,
            "speech_rate_wpm": None,
        },
        "scores": {
            "hesitation_score": 0.0,
            "speech_disorganization_score": 0.0,
        },
        "quality_flags": ["vad_failed"],
        "evidence": [f"VAD-анализ не выполнен: {message}."],
        "model_version": MODEL_VERSION,
    }


def safe_ratio(value: int | float, total: int | float) -> float:
    if total <= 0:
        return 0.0
    return float(value) / float(total)


def clamp01(value: float) -> float:
    if math.isnan(value):
        return 0.0
    return max(0.0, min(1.0, value))


def round2(value: float) -> float:
    return round(value, 2)


def round4(value: float) -> float:
    return round(value, 4)


def percent(value: float) -> str:
    return f"{round(value * 100, 1)}%"
