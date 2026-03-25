package specialists

import (
	"context"
	"testing"
	"time"
)

func TestSpecialistReadModelsExposeRegistrySummary(t *testing.T) {
	lastExamAt := time.Unix(20, 0).UTC()
	baselineRefreshedAt := time.Unix(30, 0).UTC()
	repo := &specialistRepoStub{
		listResult: []Specialist{
			{
				ID:                    10,
				FullName:              "Иванов Иван",
				ExaminationsCount:     6,
				LastExaminationID:     int64Ptr(21),
				LastExaminationAt:     &lastExamAt,
				LastExaminationStatus: stringPtr("completed"),
				LastOverallScore:      float64Ptr(0.74),
				LastOverallBand:       stringPtr("elevated"),
				BaselineExamCount:     4,
				BaselineRefreshedAt:   &baselineRefreshedAt,
			},
		},
		getByIDResult: Specialist{
			ID:                    10,
			FullName:              "Иванов Иван",
			ExaminationsCount:     6,
			LastExaminationID:     int64Ptr(21),
			LastExaminationAt:     &lastExamAt,
			LastExaminationStatus: stringPtr("completed"),
			LastOverallScore:      float64Ptr(0.74),
			LastOverallBand:       stringPtr("elevated"),
			BaselineExamCount:     4,
			BaselineRefreshedAt:   &baselineRefreshedAt,
		},
	}

	service := NewService(repo)

	items, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("list specialists: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one specialist, got %d", len(items))
	}
	if items[0].ExaminationsCount != 6 {
		t.Fatalf("expected 6 examinations, got %d", items[0].ExaminationsCount)
	}
	if items[0].LastExaminationStatus == nil || *items[0].LastExaminationStatus != "completed" {
		t.Fatalf("expected completed status, got %#v", items[0].LastExaminationStatus)
	}

	item, err := service.GetByID(context.Background(), 10)
	if err != nil {
		t.Fatalf("get specialist: %v", err)
	}
	if item.LastOverallBand == nil || *item.LastOverallBand != "elevated" {
		t.Fatalf("expected elevated band, got %#v", item.LastOverallBand)
	}
	if item.BaselineRefreshedAt == nil || !item.BaselineRefreshedAt.Equal(baselineRefreshedAt) {
		t.Fatalf("expected baseline refreshed at %s, got %#v", baselineRefreshedAt.Format(time.RFC3339), item.BaselineRefreshedAt)
	}
}

type specialistRepoStub struct {
	listResult    []Specialist
	getByIDResult Specialist
}

func (s *specialistRepoStub) Create(context.Context, CreateInput) (Specialist, error) {
	return Specialist{}, nil
}

func (s *specialistRepoStub) List(context.Context) ([]Specialist, error) {
	return s.listResult, nil
}

func (s *specialistRepoStub) GetByID(context.Context, int64) (Specialist, error) {
	return s.getByIDResult, nil
}

func (s *specialistRepoStub) Update(context.Context, UpdateInput) (Specialist, error) {
	return Specialist{}, nil
}

func (s *specialistRepoStub) Delete(context.Context, int64) error {
	return nil
}

func int64Ptr(value int64) *int64 {
	return &value
}

func float64Ptr(value float64) *float64 {
	return &value
}

func stringPtr(value string) *string {
	return &value
}
