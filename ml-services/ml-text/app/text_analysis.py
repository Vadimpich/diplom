import math
import os
import re
from dataclasses import dataclass
from pathlib import Path
from typing import Any


TEXT_MODEL_VERSION = "rubert-go-emotions-v1"
DEFAULT_EMOTION_MODEL_NAME = "seara/rubert-base-cased-russian-emotion-detection-ru-go-emotions"
EMOTION_KEYS = ("joy", "sadness", "anger", "fear", "surprise", "neutral")
UNCERTAINTY_MARKERS = (
    "не знаю",
    "я не знаю",
    "затрудняюсь",
    "затрудняюсь ответить",
    "сложно сказать",
    "трудно сказать",
    "не могу сказать",
    "не могу точно сказать",
    "не могу ответить",
    "не уверен",
    "не уверена",
    "не уверен что",
    "не уверена что",
    "не уверен в ответе",
    "не уверена в ответе",
    "сомневаюсь",
    "есть сомнения",
    "не до конца уверен",
    "не до конца уверена",
    "не совсем уверен",
    "не совсем уверена",
    "может быть",
    "возможно",
    "наверное",
    "кажется",
    "мне кажется",
    "вроде",
    "вроде бы",
    "скорее всего",
    "как будто",
    "не могу сформулировать",
    "сложно сформулировать",
    "трудно сформулировать",
    "не понимаю вопрос",
    "не понял вопрос",
    "не поняла вопрос",
)
NEGATION_MARKERS = ("не", "нет", "никогда", "ничего", "никто", "никак", "нигде", "без")
DISTRESS_MARKERS = (
    # тревога / эмоциональное напряжение
    "переживаю",
    "сильно переживаю",
    "тревожно",
    "мне тревожно",
    "тревога",
    "чувствую тревогу",
    "волнуюсь",
    "сильно волнуюсь",
    "напряжен",
    "напряжена",
    "напряжён",
    "напряжённая",
    "нервничаю",
    "сильно нервничаю",
    "не готов",
    "не готова",
    "не чувствую себя готовым",
    "не чувствую себя готовой",

    # плохое самочувствие
    "плохо себя чувствую",
    "очень плохо себя чувствую",
    "не очень себя чувствую",
    "самочувствие плохое",
    "самочувствие не очень",
    "мне плохо",
    "будет плохо",
    "не комфортно",
    "некомфортно",
    "мне некомфортно",
    "чувствую дискомфорт",
    "слабость",
    "чувствую слабость",
    "слабость в теле",
    "голова болит",
    "болит голова",
    "болит живот",
    "тошнит",
    "меня тошнит",
    "головокружение",
    "кружится голова",
    "дрожит",
    "меня трясет",
    "меня трясёт",
    "трясутся руки",
    "дрожат руки",

    # сон / усталость
    "почти не спал",
    "почти не спала",
    "не спал",
    "не спала",
    "не выспался",
    "не выспалась",
    "плохо спал",
    "плохо спала",
    "мало спал",
    "мало спала",
    "хочется спать",
    "очень хочется спать",
    "сонный",
    "сонная",
    "сильно устал",
    "сильно устала",
    "очень устал",
    "очень устала",
    "нет сил",
    "совсем нет сил",
    "измотан",
    "измотана",

    # болезнь / физический дискомфорт
    "простудился",
    "простудилась",
    "заболел",
    "заболела",
    "кажется заболел",
    "кажется заболела",
    "температура",
    "есть температура",
    "поднялась температура",
    "кашель",
    "сильный кашель",
    "насморк",
    "ломит тело",
    "холодно",
    "мне холодно",
    "замерз",
    "замёрз",
    "замерзла",
    "замёрзла",

    # сильная негативная оценка состояния
    "ужас",
    "ужасно",
    "очень тяжело",
    "мне тяжело",
    "тяжело держаться",
    "не могу сосредоточиться",
    "трудно сосредоточиться",
    "сложно сосредоточиться",
    "путаюсь",
    "я путаюсь",
)
SENTENCE_SPLIT_RE = re.compile(r"[.!?…]+")
WORD_RE = re.compile(r"[A-Za-zА-Яа-яЁё0-9-]+")


@dataclass(frozen=True)
class EmotionModelConfig:
    model_path: str
    model_name: str
    device: str


class TextAnalysisError(Exception):
    pass


class TextAnalyzer:
    def __init__(self, config: EmotionModelConfig, classifier: Any | None = None) -> None:
        self.config = config
        self.classifier = classifier or EmotionClassifier(config)

    def analyze(self, transcript: str, *, stt_word_count: int | None = None) -> dict[str, Any]:
        features = extract_text_features(transcript, stt_word_count=stt_word_count)
        quality_flags = text_quality_flags(features)
        evidence = build_feature_evidence(features)

        emotion_probs: dict[str, float] = {key: 0.0 for key in EMOTION_KEYS}
        emotion_model_version = self.classifier.model_version
        if transcript.strip():
            try:
                emotion_probs.update(self.classifier.predict(transcript))
                emotion_model_version = self.classifier.model_version
            except TextAnalysisError as exc:
                quality_flags.append("emotion_model_failed")
                evidence.append(f"Emotion model failed: {exc}.")

        scores = calculate_scores(features, emotion_probs)
        evidence.extend(build_score_evidence(features, scores, emotion_probs))

        return {
            "features": features,
            "scores": scores,
            "emotion_probs": round_probabilities(emotion_probs),
            "quality_flags": sorted(set(quality_flags)),
            "evidence": evidence,
            "model_version": TEXT_MODEL_VERSION,
            "emotion_model_version": emotion_model_version,
        }


class EmotionClassifier:
    def __init__(self, config: EmotionModelConfig) -> None:
        self.config = config
        self._tokenizer: Any | None = None
        self._model: Any | None = None
        self._id2label: dict[int, str] = {}
        self._model_version = f"emotion-{config.model_name.split('/')[-1]}"

    @property
    def model_version(self) -> str:
        return self._model_version

    def predict(self, text: str) -> dict[str, float]:
        if not text.strip():
            return {key: 0.0 for key in EMOTION_KEYS}

        tokenizer, model = self._load()
        try:
            import torch

            inputs = tokenizer(text, return_tensors="pt", truncation=True, max_length=512)
            inputs = {key: value.to(model.device) for key, value in inputs.items()}
            with torch.no_grad():
                logits = model(**inputs).logits[0]
                if is_multilabel_model(model):
                    probs = torch.sigmoid(logits).detach().cpu().tolist()
                else:
                    probs = torch.softmax(logits, dim=-1).detach().cpu().tolist()
        except Exception as exc:
            raise TextAnalysisError(f"emotion inference failed: {exc}") from exc

        result = {key: 0.0 for key in EMOTION_KEYS}
        for index, probability in enumerate(probs):
            label = normalize_emotion_label(self._id2label.get(index, f"label_{index}"))
            if label in result:
                result[label] = combine_multilabel_probability(result[label], float(probability))
        return result

    def _load(self) -> tuple[Any, Any]:
        if self._tokenizer is not None and self._model is not None:
            return self._tokenizer, self._model

        model_name_or_path = self._resolve_model_name_or_path()
        try:
            from transformers import AutoModelForSequenceClassification, AutoTokenizer

            self._tokenizer = AutoTokenizer.from_pretrained(
                model_name_or_path,
                local_files_only=offline_mode_enabled(),
            )
            self._model = AutoModelForSequenceClassification.from_pretrained(
                model_name_or_path,
                local_files_only=offline_mode_enabled(),
            )
            if self.config.device == "cuda":
                self._model = self._model.to("cuda")
            else:
                self._model = self._model.to("cpu")
            self._model.eval()
            self._id2label = {int(key): value for key, value in self._model.config.id2label.items()}
        except Exception as exc:
            raise TextAnalysisError(f"emotion model load failed: {exc}") from exc
        return self._tokenizer, self._model

    def _resolve_model_name_or_path(self) -> str:
        if self.config.model_path:
            path = Path(self.config.model_path)
            if local_model_available(path):
                self._model_version = f"emotion-local-{path.name}"
                return str(path)
            if offline_mode_enabled():
                raise TextAnalysisError(f"TEXT_EMOTION_MODEL_PATH is missing or incomplete in offline mode: {path}")
        return self.config.model_name


def load_text_analysis_config() -> EmotionModelConfig:
    return EmotionModelConfig(
        model_path=os.getenv("TEXT_EMOTION_MODEL_PATH", "/app/models/text-emotion"),
        model_name=os.getenv("TEXT_EMOTION_MODEL_NAME", DEFAULT_EMOTION_MODEL_NAME),
        device=os.getenv("TEXT_EMOTION_DEVICE", os.getenv("STT_DEVICE", "cpu")),
    )


def extract_text_features(transcript: str, *, stt_word_count: int | None = None) -> dict[str, Any]:
    normalized = normalize_text(transcript)
    words = WORD_RE.findall(normalized)
    sentences = [part.strip() for part in SENTENCE_SPLIT_RE.split(normalized) if part.strip()]
    word_count = stt_word_count if stt_word_count is not None else len(words)
    sentence_count = len(sentences)
    avg_sentence_length = round(word_count / sentence_count, 2) if sentence_count > 0 else 0.0
    return {
        "transcript_length_chars": len(normalized),
        "word_count": int(word_count),
        "sentence_count": sentence_count,
        "avg_sentence_length": avg_sentence_length,
        "uncertainty_marker_count": count_phrase_markers(normalized, UNCERTAINTY_MARKERS),
        "negation_count": count_negations(normalized),
        "distress_marker_count": count_phrase_markers(normalized, DISTRESS_MARKERS),
        "short_answer_flag": word_count < 4 or len(normalized) < 20,
    }


def calculate_scores(features: dict[str, Any], emotion_probs: dict[str, float]) -> dict[str, float]:
    negative_emotion_share = clamp01(
        emotion_probs.get("sadness", 0.0) + emotion_probs.get("anger", 0.0) + emotion_probs.get("fear", 0.0)
    )
    base_negativity = clamp01(
        emotion_probs.get("sadness", 0.0) * 0.38
        + emotion_probs.get("anger", 0.0) * 0.32
        + emotion_probs.get("fear", 0.0) * 0.30
    )
    uncertainty_ratio = ratio(features["uncertainty_marker_count"], max(features["word_count"], 1), scale=0.15)
    negation_ratio = ratio(features["negation_count"], max(features["word_count"], 1), scale=0.12)
    distress_ratio = ratio(features["distress_marker_count"], max(features["word_count"], 1), scale=0.08)
    short_answer = 1.0 if features["short_answer_flag"] else 0.0

    negativity = clamp01(
        base_negativity * 1.15
        + max(0.0, negative_emotion_share-0.45)*0.55
        + distress_ratio*0.30
        + negation_ratio*0.08
    )
    anxiety = clamp01(
        emotion_probs.get("fear", 0.0) * 0.45
        + max(0.0, negative_emotion_share-0.55)*0.20
        + distress_ratio * 0.30
        + uncertainty_ratio * 0.15
        + negation_ratio * 0.10
    )
    evasion = clamp01(short_answer * 0.45 + uncertainty_ratio * 0.40 + negation_ratio * 0.15)
    coherence = clamp01(1.0 - evasion * 0.45 - short_answer * 0.25 - max(0.0, features["avg_sentence_length"] - 28) / 80)
    confidence = clamp01(
        1.0
        - uncertainty_ratio * 0.25
        - evasion * 0.18
        - negativity * 0.34
        - anxiety * 0.22
        - distress_ratio * 0.16
    )
    return {
        "text_negativity_score": round(negativity, 4),
        "text_anxiety_score": round(anxiety, 4),
        "text_confidence_score": round(confidence, 4),
        "text_coherence_score": round(coherence, 4),
        "text_evasion_score": round(evasion, 4),
    }


def text_quality_flags(features: dict[str, Any]) -> list[str]:
    flags: list[str] = []
    if features["transcript_length_chars"] == 0:
        flags.append("empty_transcript")
    if features["word_count"] < 4 or features["transcript_length_chars"] < 20:
        flags.append("too_short_text")
    return flags


def build_feature_evidence(features: dict[str, Any]) -> list[str]:
    evidence = [
        f"Transcript length: {features['transcript_length_chars']} characters.",
        f"Word count: {features['word_count']}.",
        f"Sentence count: {features['sentence_count']}.",
    ]
    if features["short_answer_flag"]:
        evidence.append("Короткий ответ: мало слов или символов для устойчивого текстового анализа.")
    if features["uncertainty_marker_count"] > 0:
        evidence.append(f"Неопределённых формулировок: {features['uncertainty_marker_count']}.")
    if features["negation_count"] > 0:
        evidence.append(f"Отрицательных маркеров: {features['negation_count']}.")
    if features["distress_marker_count"] > 0:
        evidence.append(f"Маркеров неблагополучия/напряжения: {features['distress_marker_count']}.")
    return evidence


def build_score_evidence(features: dict[str, Any], scores: dict[str, float], emotion_probs: dict[str, float]) -> list[str]:
    evidence: list[str] = []
    negative_emotion_share = emotion_probs.get("sadness", 0.0) + emotion_probs.get("anger", 0.0) + emotion_probs.get("fear", 0.0)
    if negative_emotion_share >= 0.5 or scores["text_negativity_score"] >= 0.35:
        evidence.append("Высокая доля тревожной/негативной эмоциональной окраски в тексте.")
    if scores["text_anxiety_score"] >= 0.35:
        evidence.append("Повышенный текстовый anxiety score из-за fear/uncertainty markers.")
    if scores["text_evasion_score"] >= 0.45:
        evidence.append("Есть признаки уклончивости: короткий ответ, неопределённые формулировки или отрицания.")
    if features["sentence_count"] <= 1 and not features["short_answer_flag"]:
        evidence.append("Ответ представлен одной длинной фразой, coherence score снижен осторожно.")
    return evidence


def normalize_emotion_label(label: str) -> str:
    normalized = label.strip().lower()
    mapping = {
        "радость": "joy",
        "joy": "joy",
        "happy": "joy",
        "happiness": "joy",
        "amusement": "joy",
        "approval": "joy",
        "caring": "joy",
        "excitement": "joy",
        "gratitude": "joy",
        "love": "joy",
        "optimism": "joy",
        "pride": "joy",
        "relief": "joy",
        "грусть": "sadness",
        "печаль": "sadness",
        "sadness": "sadness",
        "sad": "sadness",
        "disappointment": "sadness",
        "grief": "sadness",
        "remorse": "sadness",
        "embarrassment": "sadness",
        "злость": "anger",
        "гнев": "anger",
        "anger": "anger",
        "angry": "anger",
        "annoyance": "anger",
        "disapproval": "anger",
        "disgust": "anger",
        "страх": "fear",
        "fear": "fear",
        "fearful": "fear",
        "nervousness": "fear",
        "confusion": "fear",
        "удивление": "surprise",
        "surprise": "surprise",
        "realization": "surprise",
        "neutral": "neutral",
        "нейтрально": "neutral",
        "нейтральный": "neutral",
    }
    return mapping.get(normalized, normalized)


def count_phrase_markers(text: str, markers: tuple[str, ...]) -> int:
    lowered = text.lower()
    return sum(lowered.count(marker) for marker in markers)


def count_negations(text: str) -> int:
    lowered_words = [word.lower() for word in WORD_RE.findall(text)]
    return sum(1 for word in lowered_words if word in NEGATION_MARKERS)


def normalize_text(text: str) -> str:
    return re.sub(r"\s+", " ", text).strip()


def ratio(value: int | float, denominator: int | float, *, scale: float) -> float:
    if denominator <= 0:
        return 0.0
    return clamp01((float(value) / float(denominator)) / scale)


def clamp01(value: float) -> float:
    if math.isnan(value) or value < 0:
        return 0.0
    if value > 1:
        return 1.0
    return value


def round_probabilities(probabilities: dict[str, float]) -> dict[str, float]:
    return {key: round(float(probabilities.get(key, 0.0)), 6) for key in EMOTION_KEYS}


def combine_multilabel_probability(current: float, new_value: float) -> float:
    current = clamp01(current)
    new_value = clamp01(new_value)
    return clamp01(1.0 - (1.0 - current) * (1.0 - new_value))


def is_multilabel_model(model: Any) -> bool:
    problem_type = str(getattr(model.config, "problem_type", "") or "").lower()
    if problem_type == "multi_label_classification":
        return True
    id2label = getattr(model.config, "id2label", {}) or {}
    labels = [str(value).lower() for value in id2label.values()]
    go_emotions_markers = {"nervousness", "annoyance", "approval", "gratitude", "remorse"}
    return any(label in go_emotions_markers for label in labels)


def local_model_available(path: Path) -> bool:
    return path.is_dir() and (path / "config.json").is_file()


def offline_mode_enabled() -> bool:
    return os.getenv("HF_HUB_OFFLINE", "").strip().lower() in {"1", "true", "yes", "on"}
