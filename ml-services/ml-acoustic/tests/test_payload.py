import app.main as main


def command() -> main.ProcessingCommandEnvelope:
    return main.ProcessingCommandEnvelope.model_validate(
        {
            "message_version": 1,
            "message_id": "msg-1",
            "correlation_id": "exam-100-acoustic-v1",
            "examination_id": 100,
            "specialist_id": 55,
            "channel": "acoustic",
            "attempt": 1,
            "max_attempts": 3,
            "requested_at": "2026-03-23T18:10:12Z",
            "answers": [],
        }
    )


def test_build_payload_returns_canonical_result_without_proxy_fields(monkeypatch) -> None:
    def fake_analyze_audio_bytes(_: bytes) -> dict:
        return {
            "channel": "acoustic",
            "status": "done",
            "features": {
                "duration_ms": 1_000,
                "sample_rate": 16_000,
                "rms_energy_mean": 0.1,
            },
            "scores": {
                "acoustic_stress_score": 0.4,
                "voice_stability_score": 0.7,
                "intensity_variability_score": 0.2,
            },
            "emotion_probs": {
                "neutral": 0.1,
                "happiness": 0.2,
                "sadness": 0.1,
                "anger": 0.5,
                "fear": 0.0,
                "other": 0.1,
            },
            "dominant_emotion": "anger",
            "quality_flags": [],
            "evidence": [],
            "model_version": "xbgoose-hubert-large-dusha-v1",
        }

    monkeypatch.setattr(main, "analyze_audio_bytes", fake_analyze_audio_bytes)

    payload = main.build_payload(
        command(),
        [{"answer_id": 1, "question_id": 2, "audio_s3_key": "a.webm", "bytes": 100, "content": b"audio"}],
    )

    assert payload["channel"] == "acoustic"
    assert payload["examination_id"] == 100
    assert payload["answer_id"] == 1
    assert payload["emotion_probs"]["anger"] == 0.5
    assert payload["dominant_emotion"] == "anger"
    assert payload["processing_time_ms"] >= 0
    assert payload["error"] is None
    assert "audio_energy_proxy" not in payload
    assert "audio_average_bytes" not in payload
