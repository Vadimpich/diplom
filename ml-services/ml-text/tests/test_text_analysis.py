from app.text_analysis import (
    EmotionClassifier,
    EmotionModelConfig,
    TextAnalysisError,
    TextAnalyzer,
    calculate_scores,
    combine_multilabel_probability,
    extract_text_features,
    is_multilabel_model,
    normalize_emotion_label,
)


class FakeClassifier:
    model_version = "emotion-fake"

    def predict(self, text: str) -> dict[str, float]:
        return {
            "joy": 0.05,
            "sadness": 0.2,
            "anger": 0.1,
            "fear": 0.5,
            "surprise": 0.0,
            "neutral": 0.15,
        }


class FailingClassifier:
    model_version = "emotion-failing"

    def predict(self, text: str) -> dict[str, float]:
        raise TextAnalysisError("offline")


def test_extract_text_features_counts_explainable_markers() -> None:
    features = extract_text_features("Я не знаю. Возможно, всё нормально, но я не уверен.")

    assert features["transcript_length_chars"] > 0
    assert features["word_count"] == 10
    assert features["sentence_count"] == 2
    assert features["uncertainty_marker_count"] == 3
    assert features["negation_count"] == 2
    assert features["distress_marker_count"] == 0
    assert features["short_answer_flag"] is False


def test_analyzer_combines_emotion_model_and_heuristics() -> None:
    analyzer = TextAnalyzer(EmotionModelConfig("", "fake", "cpu"), classifier=FakeClassifier())

    result = analyzer.analyze("Я не знаю. Возможно, сейчас тревожно.", stt_word_count=6)

    assert result["emotion_probs"]["fear"] == 0.5
    assert result["scores"]["text_anxiety_score"] > 0.3
    assert result["scores"]["text_confidence_score"] < 1.0
    assert result["quality_flags"] == []
    assert result["model_version"] == "rubert-go-emotions-v1"


def test_empty_transcript_returns_quality_flags_without_model_call() -> None:
    analyzer = TextAnalyzer(EmotionModelConfig("", "fake", "cpu"), classifier=FakeClassifier())

    result = analyzer.analyze("", stt_word_count=0)

    assert "empty_transcript" in result["quality_flags"]
    assert "too_short_text" in result["quality_flags"]
    assert result["emotion_probs"]["fear"] == 0.0


def test_emotion_model_failure_keeps_heuristic_result() -> None:
    analyzer = TextAnalyzer(EmotionModelConfig("", "fake", "cpu"), classifier=FailingClassifier())

    result = analyzer.analyze("Текст достаточно длинный для анализа.", stt_word_count=5)

    assert "emotion_model_failed" in result["quality_flags"]
    assert result["features"]["word_count"] == 5
    assert "scores" in result


def test_missing_local_emotion_model_falls_back_to_model_name(tmp_path, monkeypatch) -> None:
    monkeypatch.delenv("HF_HUB_OFFLINE", raising=False)
    model_dir = tmp_path / "text-emotion"
    model_dir.mkdir()
    classifier = EmotionClassifier(EmotionModelConfig(str(model_dir), "remote-model", "cpu"))

    assert classifier._resolve_model_name_or_path() == "remote-model"


def test_offline_emotion_model_requires_complete_local_model(tmp_path, monkeypatch) -> None:
    monkeypatch.setenv("HF_HUB_OFFLINE", "1")
    model_dir = tmp_path / "text-emotion"
    model_dir.mkdir()
    classifier = EmotionClassifier(EmotionModelConfig(str(model_dir), "remote-model", "cpu"))

    try:
        classifier._resolve_model_name_or_path()
    except TextAnalysisError as exc:
        assert "missing or incomplete" in str(exc)
    else:
        raise AssertionError("expected TextAnalysisError")


def test_calculate_scores_are_normalized() -> None:
    features = {
        "transcript_length_chars": 100,
        "word_count": 10,
        "sentence_count": 2,
        "avg_sentence_length": 5.0,
        "uncertainty_marker_count": 2,
        "negation_count": 1,
        "distress_marker_count": 1,
        "short_answer_flag": False,
    }
    scores = calculate_scores(features, {"sadness": 0.5, "anger": 0.2, "fear": 0.4})

    assert all(0.0 <= value <= 1.0 for value in scores.values())
    assert scores["text_negativity_score"] > 0.0


def test_distress_markers_make_negative_text_more_reactive() -> None:
    features = extract_text_features(
        "Мне плохо, я не готов сегодня работать, болит голова, почти не спал и очень переживаю."
    )
    scores = calculate_scores(
        features,
        {
            "joy": 0.12,
            "sadness": 0.41,
            "anger": 0.24,
            "fear": 0.22,
            "surprise": 0.09,
            "neutral": 0.03,
        },
    )

    assert features["distress_marker_count"] >= 4
    assert scores["text_negativity_score"] >= 0.55
    assert scores["text_anxiety_score"] >= 0.3
    assert scores["text_confidence_score"] <= 0.8


def test_go_emotions_labels_map_to_existing_contract_keys() -> None:
    assert normalize_emotion_label("nervousness") == "fear"
    assert normalize_emotion_label("annoyance") == "anger"
    assert normalize_emotion_label("gratitude") == "joy"
    assert normalize_emotion_label("disappointment") == "sadness"
    assert normalize_emotion_label("realization") == "surprise"


def test_multilabel_model_detection_works_for_go_emotions_shape() -> None:
    class Config:
        problem_type = "multi_label_classification"
        id2label = {0: "joy", 1: "nervousness"}

    class Model:
        config = Config()

    assert is_multilabel_model(Model()) is True


def test_multilabel_bucket_combination_uses_noisy_or() -> None:
    combined = combine_multilabel_probability(0.7, 0.6)

    assert round(combined, 4) == 0.88
    assert 0.0 <= combined <= 1.0


def test_classifier_predict_caps_collapsed_bucket_probabilities(monkeypatch) -> None:
    classifier = EmotionClassifier(EmotionModelConfig("", "fake-model", "cpu"))

    class FakeTokenizer:
        def __call__(self, text: str, return_tensors: str, truncation: bool, max_length: int) -> dict:
            import torch

            return {"input_ids": torch.tensor([[1, 2, 3]]), "attention_mask": torch.tensor([[1, 1, 1]])}

    class FakeConfig:
        problem_type = "multi_label_classification"
        id2label = {0: "joy", 1: "gratitude", 2: "optimism"}

    class FakeOutput:
        def __init__(self, logits) -> None:
            self.logits = logits

    class FakeModel:
        config = FakeConfig()
        device = "cpu"

        def __call__(self, **kwargs):
            import math
            import torch

            def logit(p: float) -> float:
                return math.log(p / (1 - p))

            return FakeOutput(torch.tensor([[logit(0.7), logit(0.6), logit(0.5)]], dtype=torch.float32))

    monkeypatch.setattr(classifier, "_load", lambda: (FakeTokenizer(), FakeModel()))
    classifier._id2label = {0: "joy", 1: "gratitude", 2: "optimism"}

    probs = classifier.predict("спокойный и уверенный ответ")

    assert 0.0 <= probs["joy"] <= 1.0
    assert round(probs["joy"], 4) == 0.94
