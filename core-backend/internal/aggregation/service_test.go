package aggregation

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"dimplom/internal/examinations"
	"dimplom/internal/processing"
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
	service := NewService(repo, nil, "general-v1", BaselineAlgorithmVersion)

	if err := service.TryAggregate(context.Background(), 101); err != nil {
		t.Fatalf("try aggregate: %v", err)
	}
	if repo.saved {
		t.Fatal("expected aggregation to stay idle until all mandatory channels succeed")
	}
}

func TestAggregationStoresVersionedProfile(t *testing.T) {
	repo := newReadyRepositoryStub()
	service := NewService(repo, nil, "general-v1", BaselineAlgorithmVersion)

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
	service := NewService(repo, nil, "general-v1", BaselineAlgorithmVersion)

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

type repositoryStub struct {
	readiness Readiness
	results   []PersistedChannelResult
	saved     bool
	persisted PersistInput
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
				ModelVersion:  "text-stub-0.1.0",
				Payload:       mustJSON(`{"text_total_characters":320,"text_non_empty_answers":4,"audio_total_bytes":16000}`),
				CompletedAt:   time.Unix(1_742_550_008, 0).UTC(),
			},
			{
				ExaminationID: 101,
				SpecialistID:  55,
				Channel:       processing.ChannelAcoustic,
				ModelVersion:  "acoustic-stub-0.1.0",
				Payload:       mustJSON(`{"audio_energy_proxy":14000,"audio_average_bytes":12000}`),
				CompletedAt:   time.Unix(1_742_550_009, 0).UTC(),
			},
			{
				ExaminationID: 101,
				SpecialistID:  55,
				Channel:       processing.ChannelParalinguistic,
				ModelVersion:  "paralinguistic-stub-0.1.0",
				Payload:       mustJSON(`{"speech_rate_proxy":170,"prosody_variation_proxy":250}`),
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

func (r *repositoryStub) FinalizeAggregatedProfile(context.Context, FinalizeInput) error {
	return nil
}

func mustJSON(value string) json.RawMessage {
	return json.RawMessage(value)
}
