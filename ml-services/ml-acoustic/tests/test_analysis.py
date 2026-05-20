from pathlib import Path

import numpy as np

from app.analysis import (
    AcousticSERClassifier,
    AnalysisConfig,
    analyze_audio_bytes,
    analyze_samples,
    failed_result,
)


class FakeClassifier:
    model_version = "xbgoose-hubert-large-dusha-v1-test"

    def __init__(self, probs: dict[str, float] | None = None) -> None:
        self.probs = probs or {
            "neutral": 0.20,
            "happiness": 0.05,
            "sadness": 0.15,
            "anger": 0.50,
            "fear": 0.0,
            "other": 0.10,
        }

    def predict(self, samples: np.ndarray, *, sample_rate: int) -> dict[str, float]:
        assert sample_rate == 16_000
        assert samples.size > 0
        return self.probs


def test_analyze_samples_runs_ser_inference_and_returns_canonical_payload() -> None:
    sample_rate = 16_000
    duration_seconds = 2.0
    t = np.linspace(0, duration_seconds, int(sample_rate * duration_seconds), endpoint=False)
    samples = (0.2 * np.sin(2 * np.pi * 180 * t)).astype(np.float32)

    result = analyze_samples(samples, sample_rate=sample_rate, classifier=FakeClassifier())

    assert result["model_version"] == "xbgoose-hubert-large-dusha-v1-test"
    assert result["features"]["duration_ms"] == 2_000
    assert result["features"]["sample_rate"] == 16_000
    assert result["features"]["rms_energy_mean"] > 0
    assert result["dominant_emotion"] == "anger"
    assert result["emotion_probs"]["anger"] == 0.5
    assert 0 <= result["scores"]["acoustic_stress_score"] <= 1
    assert 0 <= result["scores"]["voice_stability_score"] <= 1
    assert 0 <= result["scores"]["intensity_variability_score"] <= 1
    assert result["quality_flags"] == []


def test_analyze_samples_flags_short_low_volume_audio() -> None:
    samples = np.zeros(1_600, dtype=np.float32)

    result = analyze_samples(samples, sample_rate=16_000, config=AnalysisConfig(min_audio_duration_ms=1_000), classifier=FakeClassifier())

    assert "audio_too_short" in result["quality_flags"]
    assert "low_volume" in result["quality_flags"]


def test_local_model_path_is_preferred_when_present(tmp_path: Path) -> None:
    model_dir = tmp_path / "acoustic-emotion"
    model_dir.mkdir()
    (model_dir / "config.json").write_text("{}", encoding="utf-8")
    captured: dict[str, object] = {}

    def fake_pipeline(*args, **kwargs):
        captured["args"] = args
        captured["kwargs"] = kwargs

        def run(_: dict, top_k=None):
            return [{"label": "neutral", "score": 1.0}]

        return run

    classifier = AcousticSERClassifier(
        AnalysisConfig(model_path=str(model_dir), model_id="remote/model", device="cpu"),
        pipeline_factory=fake_pipeline,
    )

    classifier.predict(np.ones(16_000, dtype=np.float32), sample_rate=16_000)

    assert captured["kwargs"]["model"] == str(model_dir)
    assert classifier.model_version.endswith(f"local-{model_dir.name}")


def test_model_id_fallback_is_used_when_local_model_missing(monkeypatch) -> None:
    monkeypatch.delenv("HF_HUB_OFFLINE", raising=False)
    captured: dict[str, object] = {}

    def fake_pipeline(*args, **kwargs):
        captured["kwargs"] = kwargs

        def run(_: dict, top_k=None):
            return [{"label": "positive", "score": 1.0}]

        return run

    classifier = AcousticSERClassifier(
        AnalysisConfig(model_path="/tmp/missing-acoustic-model", model_id="remote/model", device="cpu"),
        pipeline_factory=fake_pipeline,
    )

    classifier.predict(np.ones(16_000, dtype=np.float32), sample_rate=16_000)

    assert captured["kwargs"]["model"] == "remote/model"


def test_bad_audio_returns_quality_flag_instead_of_panic() -> None:
    result = analyze_audio_bytes(b"not-an-audio-file", classifier=FakeClassifier())

    assert result["status"] == "done"
    assert result["quality_flags"] == ["unsupported_audio_format"]
    assert result["dominant_emotion"] == "neutral"


def test_failed_result_returns_requested_flag() -> None:
    result = failed_result("inference_failed", "decode failed")

    assert result["status"] == "done"
    assert result["quality_flags"] == ["inference_failed"]
    assert result["emotion_probs"]["neutral"] == 0.0
