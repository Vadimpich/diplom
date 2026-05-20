from app.analysis import AnalysisConfig, SpeechSegment, build_result_from_segments, failed_result


def test_build_result_from_vad_segments_calculates_pause_features() -> None:
    result = build_result_from_segments(
        total_audio_duration_ms=10_000,
        segments=[
            SpeechSegment(start_ms=500, end_ms=2_500),
            SpeechSegment(start_ms=3_100, end_ms=5_000),
            SpeechSegment(start_ms=6_200, end_ms=8_000),
        ],
        word_count=30,
        config=AnalysisConfig(long_pause_threshold_ms=1_000),
    )

    assert result["model_version"] == "paralinguistic-vad-v1"
    assert result["features"]["speech_duration_ms"] == 5_700
    assert result["features"]["silence_duration_ms"] == 4_300
    assert result["features"]["speech_ratio"] == 0.57
    assert result["features"]["response_delay_ms"] == 500
    assert result["features"]["pause_count"] == 2
    assert result["features"]["long_pause_count"] == 1
    assert result["features"]["mean_pause_ms"] == 900
    assert result["features"]["max_pause_ms"] == 1_200
    assert result["features"]["speech_segment_count"] == 3
    assert result["features"]["speech_rate_wpm"] == 315.79
    assert 0 < result["scores"]["hesitation_score"] < 1
    assert result["quality_flags"] == []
    assert any("Длинных пауз: 1" in item for item in result["evidence"])


def test_build_result_flags_no_speech_and_low_ratio() -> None:
    result = build_result_from_segments(
        total_audio_duration_ms=5_000,
        segments=[],
        config=AnalysisConfig(low_speech_ratio_threshold=0.2),
    )

    assert result["features"]["speech_duration_ms"] == 0
    assert result["features"]["response_delay_ms"] == 5_000
    assert result["features"]["speech_rate_wpm"] is None
    assert "no_speech_detected" in result["quality_flags"]
    assert "low_speech_ratio" in result["quality_flags"]


def test_build_result_flags_too_short_speech() -> None:
    result = build_result_from_segments(
        total_audio_duration_ms=3_000,
        segments=[SpeechSegment(start_ms=100, end_ms=500)],
        config=AnalysisConfig(min_speech_duration_ms=700),
    )

    assert result["features"]["speech_duration_ms"] == 400
    assert "too_short_speech" in result["quality_flags"]


def test_failed_result_returns_vad_failed_quality_flag() -> None:
    result = failed_result("decode failed")

    assert result["status"] == "done"
    assert result["quality_flags"] == ["vad_failed"]
    assert result["features"]["total_audio_duration_ms"] == 0
