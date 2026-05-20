import app.main as main


class FakeSTT:
    model_version = "faster-whisper-test"

    def transcribe_bytes(self, audio: bytes, *, suffix: str = ".webm") -> dict:
        return {
            "transcript": "проверка связи",
            "language": "ru",
            "segments": [{"start_ms": 0, "end_ms": 1200, "text": "проверка связи"}],
            "duration_ms": 1200,
            "word_count": 2,
            "model_version": self.model_version,
        }


class FakeTextAnalyzer:
    def analyze(self, transcript: str, *, stt_word_count: int | None = None) -> dict:
        return {
            "features": {
                "transcript_length_chars": len(transcript),
                "word_count": stt_word_count or 0,
                "sentence_count": 1,
                "avg_sentence_length": float(stt_word_count or 0),
                "uncertainty_marker_count": 0,
                "negation_count": 0,
                "distress_marker_count": 0,
                "short_answer_flag": False,
            },
            "scores": {
                "text_negativity_score": 0.1,
                "text_anxiety_score": 0.2,
                "text_confidence_score": 0.8,
                "text_coherence_score": 0.9,
                "text_evasion_score": 0.0,
            },
            "emotion_probs": {"joy": 0.1, "sadness": 0.1, "anger": 0.0, "fear": 0.2, "surprise": 0.0, "neutral": 0.6},
            "quality_flags": [],
            "evidence": ["Text analysis done."],
            "model_version": "rubert-go-emotions-v1",
            "emotion_model_version": "emotion-test",
        }


def test_build_payload_includes_canonical_text_result(monkeypatch) -> None:
    monkeypatch.setattr(main.state, "stt", FakeSTT())
    monkeypatch.setattr(main.state, "text_analyzer", FakeTextAnalyzer())
    command = main.ProcessingCommandEnvelope.model_validate(
        {
            "message_version": 1,
            "message_id": "msg-1",
            "correlation_id": "exam-100-text-v1",
            "examination_id": 100,
            "specialist_id": 55,
            "channel": "text",
            "attempt": 1,
            "max_attempts": 3,
            "requested_at": "2026-03-23T18:10:12Z",
            "answers": [],
        }
    )

    payload = main.build_payload(
        command,
        [
            {
                "answer_id": 1,
                "question_id": 2,
                "audio_s3_key": "answer.webm",
                "bytes": 100,
                "content": b"audio",
            }
        ],
    )

    assert payload["stt"]["transcript"] == "проверка связи"
    assert payload["stt"]["language"] == "ru"
    assert payload["stt"]["word_count"] == 2
    assert payload["stt"]["segments"][0]["answer_id"] == 1
    assert payload["features"]["word_count"] == 2
    assert payload["scores"]["text_confidence_score"] == 0.8
    assert payload["emotion_probs"]["neutral"] == 0.6
    assert payload["quality_flags"] == []
    assert payload["examination_id"] == 100
    assert payload["answer_id"] == 1
    assert payload["processing_time_ms"] >= 0
    assert payload["error"] is None
    assert "text_total_characters" not in payload
    assert "text_non_empty_answers" not in payload


def test_build_payload_returns_stt_failed_quality_flag(monkeypatch) -> None:
    class FailingSTT:
        model_version = "faster-whisper-test"

        def transcribe_bytes(self, audio: bytes, *, suffix: str = ".webm") -> dict:
            raise main.STTError("offline")

    monkeypatch.setattr(main.state, "stt", FailingSTT())
    monkeypatch.setattr(main.state, "text_analyzer", FakeTextAnalyzer())
    command = main.ProcessingCommandEnvelope.model_validate(
        {
            "message_version": 1,
            "message_id": "msg-1",
            "correlation_id": "exam-100-text-v1",
            "examination_id": 100,
            "specialist_id": 55,
            "channel": "text",
            "attempt": 1,
            "max_attempts": 3,
            "requested_at": "2026-03-23T18:10:12Z",
            "answers": [],
        }
    )

    payload = main.build_payload(
        command,
        [
            {
                "answer_id": 1,
                "question_id": 2,
                "audio_s3_key": "answer.webm",
                "bytes": 100,
                "content": b"audio",
            }
        ],
    )

    assert "stt_failed" in payload["quality_flags"]
    assert payload["stt"]["transcript"] == ""
    assert payload["error"] is None


def test_build_payload_logs_stt_transcripts_when_debug_enabled(monkeypatch) -> None:
    monkeypatch.setattr(main.state, "stt", FakeSTT())
    monkeypatch.setattr(main.state, "text_analyzer", FakeTextAnalyzer())
    monkeypatch.setenv("TEXT_DEBUG_LOG_STT_TRANSCRIPTS", "1")
    monkeypatch.setenv("TEXT_DEBUG_LOG_STT_MAX_CHARS", "20")

    log_lines: list[str] = []

    class Logger:
        def info(self, message: str, *args) -> None:
            log_lines.append(message % args)

    monkeypatch.setattr(main.state, "logger", Logger())
    command = main.ProcessingCommandEnvelope.model_validate(
        {
            "message_version": 1,
            "message_id": "msg-1",
            "correlation_id": "exam-100-text-v1",
            "examination_id": 100,
            "specialist_id": 55,
            "channel": "text",
            "attempt": 1,
            "max_attempts": 3,
            "requested_at": "2026-03-23T18:10:12Z",
            "answers": [],
        }
    )

    main.build_payload(
        command,
        [
            {
                "answer_id": 1,
                "question_id": 2,
                "audio_s3_key": "answer.webm",
                "bytes": 100,
                "content": b"audio",
            }
        ],
    )

    assert any("stt_debug answer" in line for line in log_lines)
    assert any("stt_debug aggregate" in line for line in log_lines)
