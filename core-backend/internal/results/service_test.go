package results

import (
	"context"
	"testing"
)

func TestGetExaminationResultReturnsDecisionView(t *testing.T) {
	repo := resultsRepoStub{
		result: ExaminationResultResponse{
			ExaminationID: 101,
			Status:        "completed",
			Decision: DecisionResultView{
				State:                "succeeded",
				Recommendation:       "unavailable",
				Message:              "analysis_not_implemented_yet",
				CorrelationID:        "exam-101-kesmi-1",
				AttemptCount:         1,
				MaxAttempts:          2,
				RawResponseAvailable: true,
			},
		},
	}
	service := NewService(repo)

	result, err := service.GetExaminationResult(context.Background(), 101)
	if err != nil {
		t.Fatalf("get examination result: %v", err)
	}
	if result.Decision.Message != "analysis_not_implemented_yet" {
		t.Fatalf("expected placeholder message, got %q", result.Decision.Message)
	}
	if result.Decision.CorrelationID == "" {
		t.Fatal("expected correlation_id in decision result")
	}
}

type resultsRepoStub struct {
	result ExaminationResultResponse
}

func (r resultsRepoStub) GetExaminationResult(context.Context, int64) (ExaminationResultResponse, error) {
	return r.result, nil
}

func (r resultsRepoStub) GetSpecialistHistory(context.Context, int64) (SpecialistHistoryResponse, error) {
	return SpecialistHistoryResponse{}, nil
}
