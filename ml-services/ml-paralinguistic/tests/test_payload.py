import app.main as main


def command() -> main.ProcessingCommandEnvelope:
    return main.ProcessingCommandEnvelope.model_validate(
        {
            "message_version": 1,
            "message_id": "msg-1",
            "correlation_id": "exam-100-paralinguistic-v1",
            "examination_id": 100,
            "specialist_id": 55,
            "channel": "paralinguistic",
            "attempt": 1,
            "max_attempts": 3,
            "requested_at": "2026-03-23T18:10:12Z",
            "answers": [],
        }
    )


def test_build_payload_returns_canonical_result_without_proxy_fields(monkeypatch) -> None:
    def fake_analyze_audio_bytes(*args, **kwargs) -> dict:
        return {
            "channel": "paralinguistic",
            "status": "done",
            "features": {
                "total_audio_duration_ms": 1_000,
                "speech_duration_ms": 600,
                "silence_duration_ms": 400,
                "speech_ratio": 0.6,
                "response_delay_ms": 100,
                "pause_count": 1,
                "long_pause_count": 0,
                "mean_pause_ms": 200,
                "max_pause_ms": 200,
                "speech_segment_count": 2,
                "speech_rate_wpm": None,
            },
            "scores": {"hesitation_score": 0.2, "speech_disorganization_score": 0.1},
            "quality_flags": [],
            "evidence": [],
            "model_version": "paralinguistic-vad-v1",
        }

    monkeypatch.setattr(main, "analyze_audio_bytes", fake_analyze_audio_bytes)

    payload = main.build_payload(
        command(),
        [{"answer_id": 1, "question_id": 2, "audio_s3_key": "a.webm", "bytes": 100, "content": b"audio", "answer_text": ""}],
    )

    assert payload["channel"] == "paralinguistic"
    assert payload["examination_id"] == 100
    assert payload["answer_id"] == 1
    assert payload["processing_time_ms"] >= 0
    assert payload["error"] is None
    assert "speech_rate_proxy" not in payload
    assert "prosody_variation_proxy" not in payload
