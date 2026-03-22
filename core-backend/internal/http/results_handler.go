package http

import (
	"context"
	"net/http"
	"time"

	"dimplom/internal/aggregation"
	"dimplom/internal/results"
)

type resultsProvider interface {
	GetExaminationResult(context.Context, int64) (aggregation.AggregatedProfile, error)
	GetSpecialistHistory(context.Context, int64) (results.SpecialistHistoryResponse, error)
}

type ResultsHandler struct {
	service resultsProvider
}

func (h ResultsHandler) GetExaminationResult(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64Param(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid examination id")
		return
	}
	if h.service == nil {
		writeJSON(w, http.StatusOK, aggregation.AggregatedProfile{
			SchemaVersion:      aggregation.SchemaVersionV1,
			AggregationVersion: aggregation.AggregationVersionV1,
			ExaminationID:      id,
			Status:             "aggregated",
			GeneratedAt:        time.Now().UTC(),
			Summary: aggregation.ProfileSummary{
				OverallScore:                     0.58,
				OverallBand:                      "elevated",
				PrimaryMetricKey:                 aggregation.MetricKeyOverallProxyIndex,
				NeutralRecommendationPlaceholder: "phase3_pending_external_decision",
			},
			Metrics: []aggregation.Metric{{Key: aggregation.MetricKeyOverallProxyIndex, Label: "Сводный прокси-индекс", Value: 0.58, Scale: "0..1", Direction: "higher_means_more_deviation"}},
			ChannelContributions: []aggregation.ChannelContribution{{Channel: "text", MetricKey: aggregation.MetricKeyOverallProxyIndex, Weight: 0.33, Contribution: 0.17, EvidenceKeys: []string{aggregation.MetricKeyTextProxySignal}}},
			Explanations: []aggregation.Explanation{{Position: 1, Kind: aggregation.ExplanationKindSummary, Text: "Повышение индекса связано с proxy-метриками нескольких каналов."}},
			BaselineSnapshot: aggregation.BaselineSnapshot{
				AlgorithmVersion: aggregation.BaselineAlgorithmVersion,
				RefreshedAt:      time.Now().UTC(),
				General:          aggregation.BaselineDeviation{Delta: 0.21, Band: "mild", ReferencePopulationVersion: "general-v1"},
				Personal:         aggregation.BaselineDeviation{Delta: 0.37, Band: "moderate", BaselineExamCount: 4, UpdateEligible: false},
			},
		})
		return
	}
	item, err := h.service.GetExaminationResult(r.Context(), id)
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h ResultsHandler) GetSpecialistHistory(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64Param(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid specialist id")
		return
	}
	if h.service == nil {
		writeJSON(w, http.StatusOK, results.SpecialistHistoryResponse{
			SpecialistID: id,
			Items: []results.SpecialistHistoryItem{{
				ExaminationID: 101,
				GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
				Status:        "aggregated",
				Summary:          results.Summary{OverallScore: 0.58, OverallBand: "elevated"},
				BaselineSnapshot: results.BaselineSnapshot{AlgorithmVersion: aggregation.BaselineAlgorithmVersion, RefreshedAt: time.Now().UTC().Format(time.RFC3339), GeneralDelta: 0.21, PersonalDelta: 0.37, BaselineExamCount: 4},
				KeyMetrics: []results.HistoryMetric{{Key: aggregation.MetricKeyOverallProxyIndex, Label: "Сводный прокси-индекс", Value: 0.58}},
			}},
		})
		return
	}
	item, err := h.service.GetSpecialistHistory(r.Context(), id)
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}
	writeJSON(w, http.StatusOK, item)
}
