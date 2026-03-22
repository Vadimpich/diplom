package results

import (
	"context"

	"dimplom/internal/aggregation"
)

type Repository interface {
	GetExaminationResult(context.Context, int64) (aggregation.AggregatedProfile, error)
	GetSpecialistHistory(context.Context, int64) (SpecialistHistoryResponse, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetExaminationResult(ctx context.Context, examinationID int64) (aggregation.AggregatedProfile, error) {
	return s.repo.GetExaminationResult(ctx, examinationID)
}

func (s *Service) GetSpecialistHistory(ctx context.Context, specialistID int64) (SpecialistHistoryResponse, error) {
	return s.repo.GetSpecialistHistory(ctx, specialistID)
}

type SpecialistHistoryResponse struct {
	SpecialistID int64                `json:"specialist_id"`
	Items        []SpecialistHistoryItem `json:"items"`
}

type SpecialistHistoryItem struct {
	ExaminationID    int64                    `json:"examination_id"`
	GeneratedAt      string                   `json:"generated_at"`
	Status           string                   `json:"status"`
	Summary          Summary                  `json:"summary"`
	BaselineSnapshot BaselineSnapshot         `json:"baseline_snapshot"`
	KeyMetrics       []HistoryMetric          `json:"key_metrics"`
}

type Summary struct {
	OverallScore float64 `json:"overall_score"`
	OverallBand  string  `json:"overall_band"`
}

type BaselineSnapshot struct {
	AlgorithmVersion  string  `json:"algorithm_version"`
	RefreshedAt       string  `json:"refreshed_at"`
	GeneralDelta      float64 `json:"general_delta"`
	PersonalDelta     float64 `json:"personal_delta"`
	BaselineExamCount int     `json:"baseline_exam_count"`
}

type HistoryMetric struct {
	Key               string   `json:"key"`
	Label             string   `json:"label"`
	Value             float64  `json:"value"`
	PreviousValue     *float64 `json:"previous_value"`
	DeltaFromPrevious *float64 `json:"delta_from_previous"`
}
