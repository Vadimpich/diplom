package aggregation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"diplom/internal/examinations"
	"diplom/internal/processing"
)

var ErrProfileNotReady = errors.New("aggregation: profile not ready")

type BaselineCalculator interface {
	Calculate(context.Context, BaselineRequest) (BaselineResponse, error)
}

type DecisionStarter interface {
	CreatePendingDecision(context.Context, AggregatedProfile) error
}

type Repository interface {
	GetReadiness(context.Context, int64) (Readiness, error)
	LoadSucceededResults(context.Context, int64) ([]PersistedChannelResult, error)
	SaveAggregatingProfile(context.Context, PersistInput) (bool, error)
	LoadBaselineHistory(context.Context, int64, int64) (BaselineHistory, error)
	LoadExistingBaseline(context.Context, int64) (ExistingBaselinePayload, error)
	FinalizeAggregatedProfile(context.Context, FinalizeInput) error
}

type Service struct {
	repo                     Repository
	baseline                 BaselineCalculator
	decision                 DecisionStarter
	generalReferenceVersion  string
	baselineAlgorithmVersion string
	nowFunc                  func() time.Time
}

type Readiness struct {
	ExaminationID     int64
	SpecialistID      int64
	Status            string
	ChannelsSucceeded int
	AlreadyAggregated bool
}

type PersistedChannelResult struct {
	ExaminationID int64
	SpecialistID  int64
	Channel       string
	ModelVersion  string
	Payload       json.RawMessage
	CompletedAt   time.Time
}

type PersistInput struct {
	Profile AggregatedProfile
}

type FinalizeInput struct {
	Profile  AggregatedProfile
	Baseline BaselineResponse
}

func NewService(
	repo Repository,
	baseline BaselineCalculator,
	decision DecisionStarter,
	generalReferenceVersion string,
	baselineAlgorithmVersion string,
) *Service {
	return &Service{
		repo:                     repo,
		baseline:                 baseline,
		decision:                 decision,
		generalReferenceVersion:  generalReferenceVersion,
		baselineAlgorithmVersion: baselineAlgorithmVersion,
		nowFunc:                  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) TryAggregate(ctx context.Context, examinationID int64) error {
	readiness, err := s.repo.GetReadiness(ctx, examinationID)
	if err != nil {
		return err
	}
	if readiness.AlreadyAggregated || readiness.Status == examinations.StatusAggregated {
		return nil
	}
	if readiness.Status == examinations.StatusFailed || readiness.ChannelsSucceeded < len(processing.MandatoryChannels) {
		return nil
	}

	results, err := s.repo.LoadSucceededResults(ctx, examinationID)
	if err != nil {
		return err
	}
	profile, err := s.buildProfile(readiness, results)
	if err != nil {
		return err
	}
	inserted, err := s.repo.SaveAggregatingProfile(ctx, PersistInput{Profile: profile})
	if err != nil {
		return err
	}
	if !inserted && s.baseline == nil {
		return nil
	}
	if s.baseline == nil {
		return nil
	}

	history, err := s.repo.LoadBaselineHistory(ctx, readiness.SpecialistID, examinationID)
	if err != nil {
		return err
	}
	existingBaseline, err := s.repo.LoadExistingBaseline(ctx, readiness.SpecialistID)
	if err != nil {
		return err
	}
	request := BaselineRequest{
		SchemaVersion:                     SchemaVersionV1,
		AlgorithmVersion:                  s.baselineAlgorithmVersion,
		SpecialistID:                      readiness.SpecialistID,
		ExaminationID:                     examinationID,
		GeneratedAt:                       profile.GeneratedAt,
		Metrics:                           baselineMetricsFromProfile(profile),
		History:                           history,
		ExistingBaseline:                  existingBaseline,
		Context:                           baselineContextFromResults(results, profile),
		GeneralReferencePopulationVersion: s.generalReferenceVersion,
	}
	response, err := s.baseline.Calculate(ctx, request)
	if err != nil {
		return err
	}

	if err := s.repo.FinalizeAggregatedProfile(ctx, FinalizeInput{
		Profile:  profile,
		Baseline: response,
	}); err != nil {
		return err
	}
	if s.decision != nil {
		return s.decision.CreatePendingDecision(ctx, profile)
	}
	return nil
}

func (s *Service) buildProfile(readiness Readiness, results []PersistedChannelResult) (AggregatedProfile, error) {
	if len(results) < len(processing.MandatoryChannels) {
		return AggregatedProfile{}, ErrProfileNotReady
	}

	byChannel := make(map[string]map[string]any, len(results))
	generatedAt := time.Time{}
	for _, result := range results {
		var payload map[string]any
		if err := json.Unmarshal(result.Payload, &payload); err != nil {
			return AggregatedProfile{}, fmt.Errorf("decode %s payload: %w", result.Channel, err)
		}
		byChannel[result.Channel] = payload
		if generatedAt.IsZero() || result.CompletedAt.After(generatedAt) {
			generatedAt = result.CompletedAt.UTC()
		}
	}
	if generatedAt.IsZero() {
		generatedAt = s.nowFunc()
	}

	textScore := textRiskSignal(byChannel[processing.ChannelText])
	acousticScore := acousticStressSignal(byChannel[processing.ChannelAcoustic])
	paralinguisticScore := paralinguisticBehaviorSignal(byChannel[processing.ChannelParalinguistic])
	stabilityScore := speechStabilityScore(
		byChannel[processing.ChannelAcoustic],
		byChannel[processing.ChannelParalinguistic],
	)
	overall := round4(textScore*0.40 + acousticScore*0.40 + paralinguisticScore*0.20)

	metrics := []Metric{
		{
			Key:       MetricKeyOverallDeviationIndex,
			Label:     "Сводный индекс отклонения",
			Value:     overall,
			Scale:     "0..1",
			Direction: "higher_means_more_deviation",
		},
		{
			Key:       MetricKeyTextRiskSignal,
			Label:     "Текстовый риск-сигнал",
			Value:     round4(textScore),
			Scale:     "0..1",
			Direction: "higher_means_more_deviation",
		},
		{
			Key:       MetricKeyAcousticStressSignal,
			Label:     "Акустический сигнал напряжения",
			Value:     round4(acousticScore),
			Scale:     "0..1",
			Direction: "higher_means_more_deviation",
		},
		{
			Key:       MetricKeyParalinguisticBehavior,
			Label:     "Паралингвистический поведенческий сигнал",
			Value:     round4(paralinguisticScore),
			Scale:     "0..1",
			Direction: "higher_means_more_deviation",
		},
		{
			Key:       MetricKeySpeechStabilityScore,
			Label:     "Устойчивость речи",
			Value:     round4(stabilityScore),
			Scale:     "0..1",
			Direction: "higher_means_less_deviation",
		},
	}

	contributions := []ChannelContribution{
		{
			Channel:      processing.ChannelText,
			MetricKey:    MetricKeyOverallDeviationIndex,
			Weight:       0.40,
			Contribution: round4(textScore * 0.40),
			EvidenceKeys: []string{MetricKeyTextRiskSignal},
		},
		{
			Channel:      processing.ChannelAcoustic,
			MetricKey:    MetricKeyOverallDeviationIndex,
			Weight:       0.40,
			Contribution: round4(acousticScore * 0.40),
			EvidenceKeys: []string{MetricKeyAcousticStressSignal},
		},
		{
			Channel:      processing.ChannelParalinguistic,
			MetricKey:    MetricKeyOverallDeviationIndex,
			Weight:       0.20,
			Contribution: round4(paralinguisticScore * 0.20),
			EvidenceKeys: []string{MetricKeyParalinguisticBehavior, MetricKeySpeechStabilityScore},
		},
	}
	sort.Slice(contributions, func(i, j int) bool {
		return contributions[i].Contribution > contributions[j].Contribution
	})

	primary := contributions[0]
	secondary := contributions[1]
	explanations := []Explanation{
		{
			Position: 1,
			Kind:     ExplanationKindSummary,
			Text: fmt.Sprintf(
				"Повышение индекса в основном связано с измеримыми score-ами %s и %s каналов.",
				primary.Channel,
				secondary.Channel,
			),
		},
		{
			Position: 2,
			Kind:     ExplanationKindSummary,
			Text: fmt.Sprintf(
				"Текущая версия агрегирует реальные channel scores и не принимает итоговое решение о допуске.",
			),
		},
	}

	return AggregatedProfile{
		SchemaVersion:      SchemaVersionV1,
		AggregationVersion: AggregationVersionV1,
		ExaminationID:      readiness.ExaminationID,
		SpecialistID:       readiness.SpecialistID,
		Status:             examinations.StatusAggregating,
		GeneratedAt:        generatedAt,
		Summary: ProfileSummary{
			OverallScore:                     overall,
			OverallBand:                      bandForScore(overall),
			PrimaryMetricKey:                 MetricKeyOverallDeviationIndex,
			NeutralRecommendationPlaceholder: "phase3_pending_external_decision",
		},
		Metrics:              metrics,
		ChannelContributions: contributions,
		Explanations:         explanations,
		BaselineSnapshot: BaselineSnapshot{
			AlgorithmVersion: s.baselineAlgorithmVersion,
		},
	}, nil
}

func baselineMetricsFromProfile(profile AggregatedProfile) []BaselineMetricValue {
	values := make([]BaselineMetricValue, 0, len(profile.Metrics))
	for _, metric := range profile.Metrics {
		values = append(values, BaselineMetricValue{
			Key:   metric.Key,
			Value: metric.Value,
		})
	}
	return values
}

func baselineContextFromResults(results []PersistedChannelResult, profile AggregatedProfile) BaselineContext {
	flags := collectCriticalQualityFlags(results)
	reliability := dataReliabilityFromResults(results)
	return BaselineContext{
		AllChannelsDone:      len(results) >= len(processing.MandatoryChannels),
		CriticalQualityFlags: flags,
		DataReliability:      reliability,
		OverallBand:          profile.Summary.OverallBand,
	}
}

func collectCriticalQualityFlags(results []PersistedChannelResult) []string {
	flags := map[string]struct{}{}
	for _, result := range results {
		var payload map[string]any
		if err := json.Unmarshal(result.Payload, &payload); err != nil {
			continue
		}
		for _, raw := range stringSlice(objectValue(payload, "quality_flags")) {
			if isCriticalQualityFlag(raw) {
				flags[raw] = struct{}{}
			}
		}
	}
	items := make([]string, 0, len(flags))
	for flag := range flags {
		items = append(items, flag)
	}
	sort.Strings(items)
	return items
}

func dataReliabilityFromResults(results []PersistedChannelResult) float64 {
	totalFlags := 0
	criticalFlags := 0
	for _, result := range results {
		var payload map[string]any
		if err := json.Unmarshal(result.Payload, &payload); err != nil {
			return 0
		}
		flags := stringSlice(objectValue(payload, "quality_flags"))
		totalFlags += len(flags)
		for _, flag := range flags {
			if isCriticalQualityFlag(flag) {
				criticalFlags++
			}
		}
	}
	reliability := 1.0 - float64(totalFlags)*0.08 - float64(criticalFlags)*0.18
	return round4(clamp01(reliability))
}

func textRiskSignal(payload map[string]any) float64 {
	scores := objectFromPayload(payload, "scores")
	base := clamp01(
		0.35*valueFromPayload(scores, "text_negativity_score") +
			0.25*valueFromPayload(scores, "text_anxiety_score") +
			0.14*valueFromPayload(scores, "text_evasion_score") +
			0.18*(1-valueFromPayload(scores, "text_confidence_score")) +
			0.08*(1-valueFromPayload(scores, "text_coherence_score")),
	)
	boost := 0.0
	boost += math.Max(0.0, valueFromPayload(scores, "text_negativity_score")-0.35) * 0.45
	boost += math.Max(0.0, valueFromPayload(scores, "text_anxiety_score")-0.25) * 0.35
	return clamp01(base + boost)
}

func acousticStressSignal(payload map[string]any) float64 {
	scores := objectFromPayload(payload, "scores")
	return clamp01(
		0.60*valueFromPayload(scores, "acoustic_stress_score") +
			0.25*(1-valueFromPayload(scores, "voice_stability_score")) +
			0.15*valueFromPayload(scores, "intensity_variability_score"),
	)
}

func paralinguisticBehaviorSignal(payload map[string]any) float64 {
	scores := objectFromPayload(payload, "scores")
	return clamp01(
		0.35*valueFromPayload(scores, "hesitation_score") +
			0.65*valueFromPayload(scores, "speech_disorganization_score"),
	)
}

func speechStabilityScore(acousticPayload map[string]any, paralinguisticPayload map[string]any) float64 {
	acousticScores := objectFromPayload(acousticPayload, "scores")
	paralinguisticScores := objectFromPayload(paralinguisticPayload, "scores")
	voiceStability := valueFromPayload(acousticScores, "voice_stability_score")
	behaviorStability := 1 - valueFromPayload(paralinguisticScores, "speech_disorganization_score")
	if voiceStability == 0 {
		return clamp01(behaviorStability)
	}
	return clamp01(0.55*voiceStability + 0.45*behaviorStability)
}

func bandForScore(score float64) string {
	switch {
	case score >= 0.75:
		return "high"
	case score >= 0.5:
		return "elevated"
	case score >= 0.25:
		return "mild"
	default:
		return "stable"
	}
}

func objectFromPayload(payload map[string]any, key string) map[string]any {
	if payload == nil {
		return nil
	}
	value, ok := payload[key].(map[string]any)
	if !ok {
		return nil
	}
	return value
}

func objectValue(payload map[string]any, key string) any {
	if payload == nil {
		return nil
	}
	return payload[key]
}

func stringSlice(value any) []string {
	rawItems, ok := value.([]any)
	if !ok {
		return nil
	}
	items := make([]string, 0, len(rawItems))
	for _, raw := range rawItems {
		text, ok := raw.(string)
		if ok && text != "" {
			items = append(items, text)
		}
	}
	return items
}

func valueFromPayload(payload map[string]any, key string) float64 {
	if payload == nil {
		return 0
	}
	switch value := payload[key].(type) {
	case float64:
		return value
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case json.Number:
		number, _ := value.Float64()
		return number
	default:
		return 0
	}
}

func floatFromAny(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case json.Number:
		number, err := typed.Float64()
		return number, err == nil
	default:
		return 0, false
	}
}

func isCriticalQualityFlag(flag string) bool {
	switch flag {
	case "stt_failed", "empty_transcript", "feature_extraction_failed", "vad_failed", "no_speech_detected":
		return true
	default:
		return false
	}
}

func ratio(value float64, max float64) float64 {
	if max <= 0 {
		return 0
	}
	return clamp01(value / max)
}

func clamp01(value float64) float64 {
	return math.Max(0, math.Min(1, value))
}

func round4(value float64) float64 {
	return math.Round(value*10000) / 10000
}

func explanationSummary(explanations []Explanation) string {
	if len(explanations) == 0 {
		return ""
	}
	return strings.TrimSpace(explanations[0].Text)
}
