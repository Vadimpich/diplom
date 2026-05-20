package aggregation

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"diplom/internal/examinations"
	"diplom/internal/processing"
)

func TestAggregationReadyOnlyAfterAllChannelsSucceeded(t *testing.T) {
	repo := &repositoryStub{
		readiness: Readiness{
			ExaminationID:     101,
			SpecialistID:      55,
			Status:            examinations.StatusProcessing,
			ChannelsSucceeded: 2,
		},
	}
	service := NewService(repo, nil, nil, "general-v1", BaselineAlgorithmVersion)

	if err := service.TryAggregate(context.Background(), 101); err != nil {
		t.Fatalf("try aggregate: %v", err)
	}
	if repo.saved {
		t.Fatal("expected aggregation to stay idle until all mandatory channels succeed")
	}
}

func TestAggregationStoresVersionedProfile(t *testing.T) {
	repo := newReadyRepositoryStub()
	service := NewService(repo, nil, nil, "general-v1", BaselineAlgorithmVersion)

	if err := service.TryAggregate(context.Background(), 101); err != nil {
		t.Fatalf("try aggregate: %v", err)
	}
	if !repo.saved {
		t.Fatal("expected canonical profile to be persisted")
	}
	if repo.persisted.Profile.SchemaVersion != SchemaVersionV1 {
		t.Fatalf("expected schema_version=%d, got %d", SchemaVersionV1, repo.persisted.Profile.SchemaVersion)
	}
	if repo.persisted.Profile.AggregationVersion != AggregationVersionV1 {
		t.Fatalf("expected aggregation_version=%q, got %q", AggregationVersionV1, repo.persisted.Profile.AggregationVersion)
	}
	if repo.persisted.Profile.Status != examinations.StatusAggregating {
		t.Fatalf("expected status=%q, got %q", examinations.StatusAggregating, repo.persisted.Profile.Status)
	}
	if len(repo.persisted.Profile.Metrics) < 4 {
		t.Fatalf("expected canonical metrics, got %d", len(repo.persisted.Profile.Metrics))
	}
}

func TestAggregationIncludesContributionsAndExplanations(t *testing.T) {
	repo := newReadyRepositoryStub()
	service := NewService(repo, nil, nil, "general-v1", BaselineAlgorithmVersion)

	if err := service.TryAggregate(context.Background(), 101); err != nil {
		t.Fatalf("try aggregate: %v", err)
	}
	profile := repo.persisted.Profile
	if len(profile.ChannelContributions) != len(processing.MandatoryChannels) {
		t.Fatalf("expected %d channel contributions, got %d", len(processing.MandatoryChannels), len(profile.ChannelContributions))
	}
	if len(profile.Explanations) == 0 {
		t.Fatal("expected deterministic explanation bullets")
	}
	if profile.Explanations[0].Kind != ExplanationKindSummary {
		t.Fatalf("expected summary explanation, got %q", profile.Explanations[0].Kind)
	}
	if profile.ChannelContributions[0].Contribution < profile.ChannelContributions[1].Contribution {
		t.Fatal("expected channel contributions to be sorted deterministically")
	}
}

func TestAggregationStartsDecisionDeliveryAfterFinalize(t *testing.T) {
	repo := newReadyRepositoryStub()
	baseline := baselineStub{
		response: BaselineResponse{
			SchemaVersion:     SchemaVersionV1,
			AlgorithmVersion:  BaselineAlgorithmVersion,
			RefreshedAt:       time.Unix(1_742_550_011, 0).UTC(),
			GeneralDeviation:  BaselineScore{Score: 0.21, Band: "mild"},
			PersonalDeviation: BaselineScore{Score: 0.37, Band: "moderate"},
			UpdateEligibility: BaselineUpdateEligibility{
				Eligible:                     false,
				Reason:                       "frozen",
				BaselineExamCountAfterUpdate: 4,
			},
		},
	}
	decision := &decisionStarterStub{}
	service := NewService(repo, baseline, decision, "general-v1", BaselineAlgorithmVersion)

	if err := service.TryAggregate(context.Background(), 101); err != nil {
		t.Fatalf("try aggregate: %v", err)
	}
	if !decision.called {
		t.Fatal("expected aggregation finalization to start decision delivery")
	}
	if decision.profile.ExaminationID != 101 {
		t.Fatalf("expected decision starter to receive examination_id=101, got %d", decision.profile.ExaminationID)
	}
}

func TestBuildDecisionPayloadHappyPath(t *testing.T) {
	profile := newReadyAggregatedProfileForDecision()
	payload, err := BuildDecisionPayload(profile, sampleChannelPayloads(), sampleBaselineMetricScores())
	if err != nil {
		t.Fatalf("build decision payload: %v", err)
	}
	if payload.DataReliability != 0.91 {
		t.Fatalf("expected data reliability 0.91, got %.2f", payload.DataReliability)
	}
	if payload.Baseline.Source != "general" {
		t.Fatalf("expected baseline source=general, got %q", payload.Baseline.Source)
	}
	if payload.DerivedIndicators.SemanticStressIndex <= 0 {
		t.Fatal("expected semantic stress index to be derived")
	}
}

func TestBuildDecisionPayloadFailsWhenMandatoryChannelMissing(t *testing.T) {
	profile := newReadyAggregatedProfileForDecision()
	payloads := sampleChannelPayloads()
	delete(payloads, processing.ChannelText)

	_, err := BuildDecisionPayload(profile, payloads, sampleBaselineMetricScores())
	if err == nil {
		t.Fatal("expected missing mandatory channel to fail aggregation payload build")
	}
}

func TestBuildDecisionPayloadLowersReliabilityOnQualityFlags(t *testing.T) {
	profile := newReadyAggregatedProfileForDecision()
	profile.BaselineSnapshot.Personal.DataReliability = 0
	payloads := sampleChannelPayloads()
	payloads[processing.ChannelText] = processing.CanonicalChannelPayload{
		Channel:      processing.ChannelText,
		Status:       "done",
		Scores:       payloads[processing.ChannelText].Scores,
		QualityFlags: []string{"too_short_text", "empty_transcript"},
	}

	payload, err := BuildDecisionPayload(profile, payloads, sampleBaselineMetricScores())
	if err != nil {
		t.Fatalf("build decision payload: %v", err)
	}
	if payload.DataReliability >= 0.8 {
		t.Fatalf("expected reduced reliability for low-quality data, got %.2f", payload.DataReliability)
	}
}

func TestBuildDecisionPayloadFallsBackToGeneralBaseline(t *testing.T) {
	profile := newReadyAggregatedProfileForDecision()
	profile.BaselineSnapshot.Personal.BaselineAvailable = false
	profile.BaselineSnapshot.Personal.BaselineSource = "general"
	profile.BaselineSnapshot.Personal.BaselineExamCount = 2

	payload, err := BuildDecisionPayload(profile, sampleChannelPayloads(), sampleBaselineMetricScores())
	if err != nil {
		t.Fatalf("build decision payload: %v", err)
	}
	if payload.Baseline.Available {
		t.Fatal("expected personal baseline to be unavailable")
	}
	if payload.Baseline.Source != "general" {
		t.Fatalf("expected general fallback source, got %q", payload.Baseline.Source)
	}
	if payload.Baseline.BaselineDeviationIndex != 0 {
		t.Fatalf("expected immature general baseline to stay out of decision payload, got %.4f", payload.Baseline.BaselineDeviationIndex)
	}
	if payload.Baseline.ZScores[MetricKeyOverallDeviationIndex] != 0 {
		t.Fatalf("expected immature baseline z-score to be suppressed, got %.4f", payload.Baseline.ZScores[MetricKeyOverallDeviationIndex])
	}
	if payload.DerivedIndicators.BaselineShiftIndex != 0 {
		t.Fatalf("expected immature baseline shift to be suppressed, got %.4f", payload.DerivedIndicators.BaselineShiftIndex)
	}
}

func TestBuildDecisionPayloadIncludesSignificantBaselineDeviation(t *testing.T) {
	profile := newReadyAggregatedProfileForDecision()
	profile.BaselineSnapshot.Personal.BaselineAvailable = true
	profile.BaselineSnapshot.Personal.BaselineSource = "personal"
	profile.BaselineSnapshot.Personal.SignificantDeviations = []string{MetricKeyOverallDeviationIndex, MetricKeyTextRiskSignal}

	payload, err := BuildDecisionPayload(profile, sampleChannelPayloads(), sampleBaselineMetricScores())
	if err != nil {
		t.Fatalf("build decision payload: %v", err)
	}
	if len(payload.Baseline.SignificantDeviations) != 2 {
		t.Fatalf("expected significant deviations to propagate, got %v", payload.Baseline.SignificantDeviations)
	}
	if payload.DerivedIndicators.BaselineShiftIndex <= 0 {
		t.Fatal("expected positive baseline shift index")
	}
}

func TestBuildDecisionPayloadKeepsPersonalBaselineStrength(t *testing.T) {
	profile := newReadyAggregatedProfileForDecision()
	profile.BaselineSnapshot.Personal.BaselineAvailable = true
	profile.BaselineSnapshot.Personal.BaselineSource = "personal"
	profile.BaselineSnapshot.Personal.Delta = 2.4
	profile.BaselineSnapshot.Personal.BaselineExamCount = 5

	payload, err := BuildDecisionPayload(profile, sampleChannelPayloads(), sampleBaselineMetricScores())
	if err != nil {
		t.Fatalf("build decision payload: %v", err)
	}
	if payload.Baseline.BaselineDeviationIndex < 0.5 {
		t.Fatalf("expected mature personal baseline to keep strong signal, got %.4f", payload.Baseline.BaselineDeviationIndex)
	}
	if payload.Baseline.ZScores[MetricKeyOverallDeviationIndex] != 1.5 {
		t.Fatalf("expected z-scores to remain unchanged for personal baseline, got %.4f", payload.Baseline.ZScores[MetricKeyOverallDeviationIndex])
	}
}

type repositoryStub struct {
	readiness Readiness
	results   []PersistedChannelResult
	saved     bool
	persisted PersistInput
}

type baselineStub struct {
	response BaselineResponse
}

func (b baselineStub) Calculate(context.Context, BaselineRequest) (BaselineResponse, error) {
	return b.response, nil
}

type decisionStarterStub struct {
	called  bool
	profile AggregatedProfile
}

func (d *decisionStarterStub) CreatePendingDecision(_ context.Context, profile AggregatedProfile) error {
	d.called = true
	d.profile = profile
	return nil
}

func newReadyRepositoryStub() *repositoryStub {
	return &repositoryStub{
		readiness: Readiness{
			ExaminationID:     101,
			SpecialistID:      55,
			Status:            examinations.StatusProcessing,
			ChannelsSucceeded: len(processing.MandatoryChannels),
		},
		results: []PersistedChannelResult{
			{
				ExaminationID: 101,
				SpecialistID:  55,
				Channel:       processing.ChannelText,
				ModelVersion:  "rubert-cedr-v1",
				Payload:       mustJSON(`{"channel":"text","status":"done","examination_id":101,"answer_id":null,"features":{"word_count":42},"scores":{"text_negativity_score":0.28,"text_anxiety_score":0.34,"text_confidence_score":0.72,"text_coherence_score":0.82,"text_evasion_score":0.18},"quality_flags":[],"evidence":["text"],"model_version":"rubert-cedr-v1","processing_time_ms":120,"error":null}`),
				CompletedAt:   time.Unix(1_742_550_008, 0).UTC(),
			},
			{
				ExaminationID: 101,
				SpecialistID:  55,
				Channel:       processing.ChannelAcoustic,
				ModelVersion:  "acoustic-librosa-v1",
				Payload:       mustJSON(`{"channel":"acoustic","status":"done","examination_id":101,"answer_id":null,"features":{"rms_energy_mean":0.14},"scores":{"acoustic_stress_score":0.46,"voice_stability_score":0.62,"intensity_variability_score":0.31},"quality_flags":[],"evidence":["acoustic"],"model_version":"acoustic-librosa-v1","processing_time_ms":90,"error":null}`),
				CompletedAt:   time.Unix(1_742_550_009, 0).UTC(),
			},
			{
				ExaminationID: 101,
				SpecialistID:  55,
				Channel:       processing.ChannelParalinguistic,
				ModelVersion:  "paralinguistic-vad-v1",
				Payload:       mustJSON(`{"channel":"paralinguistic","status":"done","examination_id":101,"answer_id":null,"features":{"speech_ratio":0.57},"scores":{"hesitation_score":0.33,"speech_disorganization_score":0.22},"quality_flags":[],"evidence":["paralinguistic"],"model_version":"paralinguistic-vad-v1","processing_time_ms":70,"error":null}`),
				CompletedAt:   time.Unix(1_742_550_010, 0).UTC(),
			},
		},
	}
}

func (r *repositoryStub) GetReadiness(context.Context, int64) (Readiness, error) {
	return r.readiness, nil
}

func (r *repositoryStub) LoadSucceededResults(context.Context, int64) ([]PersistedChannelResult, error) {
	return r.results, nil
}

func (r *repositoryStub) SaveAggregatingProfile(_ context.Context, input PersistInput) (bool, error) {
	r.saved = true
	r.persisted = input
	return true, nil
}

func (r *repositoryStub) LoadBaselineHistory(context.Context, int64, int64) (BaselineHistory, error) {
	return BaselineHistory{}, nil
}

func (r *repositoryStub) LoadExistingBaseline(context.Context, int64) (ExistingBaselinePayload, error) {
	return ExistingBaselinePayload{}, nil
}

func (r *repositoryStub) FinalizeAggregatedProfile(context.Context, FinalizeInput) error {
	return nil
}

func mustJSON(value string) json.RawMessage {
	return json.RawMessage(value)
}

func newReadyAggregatedProfileForDecision() AggregatedProfile {
	return AggregatedProfile{
		AggregationVersion: AggregationVersionV1,
		ExaminationID:      101,
		SpecialistID:       55,
		GeneratedAt:        time.Unix(1_742_550_010, 0).UTC(),
		Summary: ProfileSummary{
			OverallScore:     0.39,
			OverallBand:      "mild",
			PrimaryMetricKey: MetricKeyOverallDeviationIndex,
		},
		BaselineSnapshot: BaselineSnapshot{
			AlgorithmVersion: BaselineAlgorithmVersion,
			General: BaselineDeviation{
				Delta:             0.21,
				Band:              "mild",
				BaselineAvailable: true,
				BaselineSource:    "general",
			},
			Personal: BaselineDeviation{
				Delta:                 0.37,
				Band:                  "moderate",
				BaselineAvailable:     false,
				BaselineSource:        "general",
				DataReliability:       0.91,
				SignificantDeviations: []string{MetricKeyOverallDeviationIndex},
			},
		},
	}
}

func sampleChannelPayloads() ChannelPayloadMap {
	return ChannelPayloadMap{
		processing.ChannelText: {
			Channel: processing.ChannelText,
			Status:  "done",
			Scores: map[string]any{
				"text_negativity_score": 0.28,
				"text_anxiety_score":    0.34,
				"text_confidence_score": 0.72,
				"text_coherence_score":  0.82,
				"text_evasion_score":    0.18,
			},
			QualityFlags: []string{},
		},
		processing.ChannelAcoustic: {
			Channel: processing.ChannelAcoustic,
			Status:  "done",
			Scores: map[string]any{
				"acoustic_stress_score":       0.46,
				"voice_stability_score":       0.62,
				"intensity_variability_score": 0.31,
			},
			QualityFlags: []string{},
		},
		processing.ChannelParalinguistic: {
			Channel: processing.ChannelParalinguistic,
			Status:  "done",
			Scores: map[string]any{
				"hesitation_score":             0.33,
				"speech_disorganization_score": 0.22,
			},
			QualityFlags: []string{},
		},
	}
}

func sampleBaselineMetricScores() map[string]BaselineMetricScore {
	return map[string]BaselineMetricScore{
		MetricKeyOverallDeviationIndex: {
			ZScore:         1.5,
			DeviationLevel: "mild",
		},
		MetricKeyTextRiskSignal: {
			ZScore:         2.1,
			DeviationLevel: "moderate",
		},
	}
}
