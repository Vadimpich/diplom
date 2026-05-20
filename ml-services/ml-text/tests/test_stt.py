from app.stt import FasterWhisperSTT, STTConfig, STTError, count_words, failed_stt_result


def test_count_words_handles_extra_spaces_and_newlines() -> None:
    assert count_words("  раз два\nтри  ") == 3


def test_failed_stt_result_keeps_contract_shape() -> None:
    result = failed_stt_result("model missing", model_version="faster-whisper-medium")

    assert result["transcript"] == ""
    assert result["language"] is None
    assert result["segments"] == []
    assert result["duration_ms"] is None
    assert result["word_count"] == 0
    assert result["model_version"] == "faster-whisper-medium"
    assert result["error"] == "model missing"


def test_missing_local_model_falls_back_to_model_size(tmp_path, monkeypatch) -> None:
    monkeypatch.delenv("HF_HUB_OFFLINE", raising=False)
    model_dir = tmp_path / "stt"
    model_dir.mkdir()
    (model_dir / ".gitkeep").write_text("", encoding="utf-8")
    stt = FasterWhisperSTT(STTConfig(str(model_dir), "medium", "cpu", "int8"))

    assert stt._resolve_model_name_or_path() == "medium"
    assert stt.model_version == "faster-whisper-medium"


def test_offline_mode_requires_complete_local_model(tmp_path, monkeypatch) -> None:
    monkeypatch.setenv("HF_HUB_OFFLINE", "1")
    model_dir = tmp_path / "stt"
    model_dir.mkdir()
    stt = FasterWhisperSTT(STTConfig(str(model_dir), "medium", "cpu", "int8"))

    try:
        stt._resolve_model_name_or_path()
    except STTError as exc:
        assert "missing or incomplete" in str(exc)
    else:
        raise AssertionError("expected STTError")


def test_complete_local_model_uses_model_path(tmp_path, monkeypatch) -> None:
    monkeypatch.delenv("HF_HUB_OFFLINE", raising=False)
    model_dir = tmp_path / "stt"
    model_dir.mkdir()
    (model_dir / "config.json").write_text("{}", encoding="utf-8")
    stt = FasterWhisperSTT(STTConfig(str(model_dir), "medium", "cpu", "int8"))

    assert stt._resolve_model_name_or_path() == str(model_dir)
    assert stt.model_version == "faster-whisper-local-stt"
