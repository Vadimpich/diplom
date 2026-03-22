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

	"dimplom/internal/examinations"
	"dimplom/internal/processing"
)

var ErrProfileNotReady = errors.New("aggregation: profile not ready")

type BaselineCalculator interface {
	Calculate(context.Context, BaselineRequest) (BaselineResponse, error)
}

type Repository interface {
	GetReadiness(context.Context, int64) (Readiness, error)
	LoadSucceededResults(context.Context, int64) ([]PersistedChannelResult, error)
	SaveAggregatingProfile(context.Context, PersistInput) (bool, error)
	LoadBaselineHistory(context.Context, int64, int64) (BaselineHistory, error)
	FinalizeAggregatedProfile(context.Context, FinalizeInput) error
}

type Service struct {
	repo                       Repository
	baseline                   BaselineCalculator
	generalReferenceVersion    string
	baselineAlgorithmVersion   string
	nowFunc                    func() time.Time
}

type Readiness struct {
	ExaminationID      int64
	SpecialistID       int64
	Status             string
	ChannelsSucceeded  int
	AlreadyAggregated  bool
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
	generalReferenceVersion string,
	baselineAlgorithmVersion string,
) *Service {
	return &Service{
		repo:                     repo,
		baseline:                 baseline,
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
	request := BaselineRequest{
		SchemaVersion:                     SchemaVersionV1,
		AlgorithmVersion:                  s.baselineAlgorithmVersion,
		SpecialistID:                      readiness.SpecialistID,
		ExaminationID:                     examinationID,
		GeneratedAt:                       profile.GeneratedAt,
		Metrics:                           baselineMetricsFromProfile(profile),
		History:                           history,
		GeneralReferencePopulationVersion: s.generalReferenceVersion,
	}
	response, err := s.baseline.Calculate(ctx, request)
	if err != nil {
		return err
	}

	return s.repo.FinalizeAggregatedProfile(ctx, FinalizeInput{
		Profile:  profile,
		Baseline: response,
	})
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

	textScore := clamp01(
		0.35*ratio(valueFromPayload(byChannel[processing.ChannelText], "text_total_characters"), 400) +
			0.35*ratio(valueFromPayload(byChannel[processing.ChannelText], "text_non_empty_answers"), 4) +
			0.30*ratio(valueFromPayload(byChannel[processing.ChannelText], "audio_total_bytes"), 20000),
	)
	acousticScore := clamp01(
		0.55*ratio(valueFromPayload(byChannel[processing.ChannelAcoustic], "audio_energy_proxy"), 18000) +
			0.45*ratio(valueFromPayload(byChannel[processing.ChannelAcoustic], "audio_average_bytes"), 16000),
	)
	paralinguisticScore := clamp01(
		0.5*ratio(valueFromPayload(byChannel[processing.ChannelParalinguistic], "speech_rate_proxy"), 200) +
			0.5*ratio(valueFromPayload(byChannel[processing.ChannelParalinguistic], "prosody_variation_proxy"), 800),
	)
	stabilityScore := clamp01(1 - ratio(valueFromPayload(byChannel[processing.ChannelParalinguistic], "prosody_variation_proxy"), 1000))
	overall := round4(textScore*0.33 + acousticScore*0.34 + paralinguisticScore*0.33)

	metrics := []Metric{
		{
			Key:       MetricKeyOverallProxyIndex,
			Label:     "Сводный прокси-индекс",
			Value:     overall,
			Scale:     "0..1",
			Direction: "higher_means_more_deviation",
		},
		{
			Key:       MetricKeyTextProxySignal,
			Label:     "Текстовый прокси-сигнал",
			Value:     round4(textScore),
			Scale:     "0..1",
			Direction: "higher_means_more_deviation",
		},
		{
			Key:       MetricKeyAcousticProxySignal,
			Label:     "Акустический прокси-сигнал",
			Value:     round4(acousticScore),
			Scale:     "0..1",
			Direction: "higher_means_more_deviation",
		},
		{
			Key:       MetricKeyParalinguisticSignal,
			Label:     "Паралингвистический прокси-сигнал",
			Value:     round4(paralinguisticScore),
			Scale:     "0..1",
			Direction: "higher_means_more_deviation",
		},
		{
			Key:       MetricKeySpeechStabilityProxy,
			Label:     "Прокси устойчивости речи",
			Value:     round4(stabilityScore),
			Scale:     "0..1",
			Direction: "higher_means_less_deviation",
		},
	}

	contributions := []ChannelContribution{
		{
			Channel:      processing.ChannelText,
			MetricKey:    MetricKeyOverallProxyIndex,
			Weight:       0.33,
			Contribution: round4(textScore * 0.33),
			EvidenceKeys: []string{MetricKeyTextProxySignal},
		},
		{
			Channel:      processing.ChannelAcoustic,
			MetricKey:    MetricKeyOverallProxyIndex,
			Weight:       0.34,
			Contribution: round4(acousticScore * 0.34),
			EvidenceKeys: []string{MetricKeyAcousticProxySignal},
		},
		{
			Channel:      processing.ChannelParalinguistic,
			MetricKey:    MetricKeyOverallProxyIndex,
			Weight:       0.33,
			Contribution: round4(paralinguisticScore * 0.33),
			EvidenceKeys: []string{MetricKeyParalinguisticSignal, MetricKeySpeechStabilityProxy},
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
				"Повышение индекса в основном связано с proxy-метриками %s и %s каналов.",
				primary.Channel,
				secondary.Channel,
			),
		},
		{
			Position: 2,
			Kind:     ExplanationKindSummary,
			Text: fmt.Sprintf(
				"Текущая версия использует нейтральную агрегацию stub-сигналов без внешнего decision layer.",
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
			PrimaryMetricKey:                 MetricKeyOverallProxyIndex,
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
