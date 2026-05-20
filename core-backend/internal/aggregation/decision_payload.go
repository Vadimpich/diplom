package aggregation

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"

	"diplom/internal/processing"
)

var lowQualityFlags = map[string]struct{}{
	"too_short_speech":     {},
	"low_speech_ratio":     {},
	"too_short_text":       {},
	"low_volume":           {},
	"f0_unavailable":       {},
	"audio_too_short":      {},
	"emotion_model_failed": {},
}

var decisionChannelScoreKeys = map[string][]string{
	processing.ChannelAcoustic: {
		"acoustic_stress_score",
		"voice_stability_score",
		"intensity_variability_score",
	},
	processing.ChannelText: {
		"text_negativity_score",
		"text_anxiety_score",
		"text_confidence_score",
		"text_coherence_score",
		"text_evasion_score",
	},
	processing.ChannelParalinguistic: {
		"hesitation_score",
		"speech_disorganization_score",
	},
}

func BuildDecisionPayload(
	profile AggregatedProfile,
	channels ChannelPayloadMap,
	baselineMetricScores map[string]BaselineMetricScore,
) (DecisionPayload, error) {
	if channels == nil {
		return DecisionPayload{}, fmt.Errorf("aggregation_failed: missing channel payloads")
	}
	for _, channel := range processing.MandatoryChannels {
		payload, ok := channels[channel]
		if !ok {
			return DecisionPayload{}, fmt.Errorf("aggregation_failed: missing mandatory channel %s", channel)
		}
		if payload.Status != "done" {
			return DecisionPayload{}, fmt.Errorf("aggregation_failed: channel %s status=%s", channel, payload.Status)
		}
	}

	dataReliability := round4(profile.BaselineSnapshot.Personal.DataReliability)
	if dataReliability <= 0 {
		dataReliability = estimateDataReliability(channels)
	}

	decisionChannels := make(map[string]DecisionChannel, len(processing.MandatoryChannels))
	allFlags := make([]string, 0)
	for _, channel := range processing.MandatoryChannels {
		payload := channels[channel]
		scores := make(map[string]float64, len(decisionChannelScoreKeys[channel]))
		for _, key := range decisionChannelScoreKeys[channel] {
			if value, ok := floatFromAny(payload.Scores[key]); ok {
				scores[key] = round4(value)
			}
		}
		flags := append([]string(nil), payload.QualityFlags...)
		sort.Strings(flags)
		allFlags = append(allFlags, flags...)
		decisionChannels[channel] = DecisionChannel{
			Scores:       scores,
			QualityFlags: flags,
		}
	}

	baselineSource := profile.BaselineSnapshot.Personal.BaselineSource
	if baselineSource == "" {
		baselineSource = "general"
	}
	baselinePreviewFactor := baselinePreviewFactor(profile)
	baselineDeviationIndex := normalizeBaselineDeviationScore(profile.BaselineSnapshot.Personal.Delta)
	if baselineSource == "general" {
		baselineDeviationIndex = normalizeBaselineDeviationScore(profile.BaselineSnapshot.General.Delta)
	}
	if !profile.BaselineSnapshot.Personal.BaselineAvailable {
		baselineDeviationIndex = round4(baselineDeviationIndex * baselinePreviewFactor)
	}

	zScores := make(map[string]float64, len(baselineMetricScores))
	for key, score := range baselineMetricScores {
		value := score.ZScore
		if !profile.BaselineSnapshot.Personal.BaselineAvailable {
			value = value * baselinePreviewFactor
		}
		zScores[key] = round4(value)
	}

	derived := DecisionDerivedIndicators{
		SemanticStressIndex: round4(clamp01(
			0.42*scoreValue(decisionChannels, processing.ChannelText, "text_negativity_score") +
				0.28*scoreValue(decisionChannels, processing.ChannelText, "text_anxiety_score") +
				0.18*(1-scoreValue(decisionChannels, processing.ChannelText, "text_confidence_score")) +
				0.07*scoreValue(decisionChannels, processing.ChannelText, "text_evasion_score") +
				0.05*(1-scoreValue(decisionChannels, processing.ChannelText, "text_coherence_score")) +
				math.Max(0.0, scoreValue(decisionChannels, processing.ChannelText, "text_negativity_score")-0.35)*0.18 +
				math.Max(0.0, scoreValue(decisionChannels, processing.ChannelText, "text_anxiety_score")-0.25)*0.12,
		)),
		AcousticActivationIndex: round4(clamp01(
			0.45*scoreValue(decisionChannels, processing.ChannelAcoustic, "acoustic_stress_score") +
				0.30*scoreValue(decisionChannels, processing.ChannelAcoustic, "intensity_variability_score") +
				0.25*(1-scoreValue(decisionChannels, processing.ChannelAcoustic, "voice_stability_score")),
		)),
		SpeechDisorganizationIndex: round4(clamp01(
			0.60*scoreValue(decisionChannels, processing.ChannelParalinguistic, "speech_disorganization_score") +
				0.40*scoreValue(decisionChannels, processing.ChannelParalinguistic, "hesitation_score"),
		)),
		BaselineShiftIndex: round4(clamp01(0.70*baselineDeviationIndex + 0.30*normalizeZScores(zScores))),
	}

	evidence := buildDecisionEvidence(profile, decisionChannels, dataReliability, baselineDeviationIndex, allFlags)

	return DecisionPayload{
		ExaminationID:   profile.ExaminationID,
		SpecialistID:    profile.SpecialistID,
		DataReliability: dataReliability,
		Channels:        decisionChannels,
		Baseline: DecisionBaseline{
			Source:                 baselineSource,
			Available:              profile.BaselineSnapshot.Personal.BaselineAvailable,
			BaselineDeviationIndex: baselineDeviationIndex,
			SignificantDeviations:  append([]string(nil), profile.BaselineSnapshot.Personal.SignificantDeviations...),
			ZScores:                zScores,
		},
		DerivedIndicators: derived,
		Evidence:          evidence,
	}, nil
}

func normalizeBaselineDeviationScore(value float64) float64 {
	return round4(clamp01(value / 4.0))
}

func baselinePreviewFactor(profile AggregatedProfile) float64 {
	if profile.BaselineSnapshot.Personal.BaselineAvailable {
		return 1
	}
	return 0
}

func estimateDataReliability(channels ChannelPayloadMap) float64 {
	reliability := 1.0
	for _, payload := range channels {
		for _, flag := range payload.QualityFlags {
			reliability -= 0.08
			if isCriticalQualityFlag(flag) {
				reliability -= 0.10
			} else if _, ok := lowQualityFlags[flag]; ok {
				reliability -= 0.04
			}
		}
	}
	return round4(clamp01(reliability))
}

func normalizeZScores(values map[string]float64) float64 {
	if len(values) == 0 {
		return 0
	}
	maxAbs := 0.0
	for _, value := range values {
		maxAbs = math.Max(maxAbs, math.Abs(value))
	}
	return clamp01(maxAbs / 3.0)
}

func scoreValue(channels map[string]DecisionChannel, channel, key string) float64 {
	if item, ok := channels[channel]; ok {
		return item.Scores[key]
	}
	return 0
}

func buildDecisionEvidence(
	profile AggregatedProfile,
	channels map[string]DecisionChannel,
	dataReliability float64,
	baselineDeviationIndex float64,
	allFlags []string,
) []string {
	items := make([]string, 0, 6)
	if dataReliability < 0.7 {
		items = append(items, fmt.Sprintf("Надежность данных снижена: %.2f.", dataReliability))
	}
	if len(profile.BaselineSnapshot.Personal.SignificantDeviations) > 0 {
		items = append(items, fmt.Sprintf(
			"Зафиксированы значимые baseline-отклонения: %s.",
			strings.Join(profile.BaselineSnapshot.Personal.SignificantDeviations, ", "),
		))
	}
	if baselineDeviationIndex >= 0.75 {
		items = append(items, fmt.Sprintf("Сдвиг относительно baseline выражен: %.2f.", baselineDeviationIndex))
	}
	if semantic := scoreValue(channels, processing.ChannelText, "text_anxiety_score"); semantic >= 0.6 {
		items = append(items, fmt.Sprintf("Текстовый канал показывает повышенную тревожность: %.2f.", semantic))
	}
	if activation := scoreValue(channels, processing.ChannelAcoustic, "acoustic_stress_score"); activation >= 0.6 {
		items = append(items, fmt.Sprintf("Акустический канал показывает повышенное напряжение: %.2f.", activation))
	}
	if speech := scoreValue(channels, processing.ChannelParalinguistic, "speech_disorganization_score"); speech >= 0.6 {
		items = append(items, fmt.Sprintf("Паралингвистический канал показывает дезорганизацию речи: %.2f.", speech))
	}
	if len(allFlags) > 0 {
		sort.Strings(allFlags)
		unique := slices.Compact(allFlags)
		items = append(items, fmt.Sprintf("Quality flags: %s.", strings.Join(unique, ", ")))
	}
	if len(items) == 0 {
		items = append(items, "Агрегатор передает explainable channel scores и baseline deviations без итогового решения.")
	}
	return items
}
