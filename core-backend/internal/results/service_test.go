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
				Recommendation:       "risk",
				Message:              "risk=medium; decision=monitoring; patterns=emotional_cross;contradictory_profile;",
				DecisionCode:         "monitoring",
				RiskClass:            "medium",
				Patterns:             []string{"emotional_cross", "contradictory_profile"},
				CorrelationID:        "exam-101-kesmi-1",
				AttemptCount:         1,
				MaxAttempts:          2,
				RawResponseAvailable: true,
			},
			ChannelReports: []ChannelReport{{
				Channel:      "text",
				ModelVersion: "rubert-cedr-v1",
				Scores: []ChannelReportScore{{
					Key:   "text_negativity_score",
					Label: "Негативная окраска текста",
					Value: 0.31,
				}},
			}},
		},
	}
	service := NewService(repo)

	result, err := service.GetExaminationResult(context.Background(), 101)
	if err != nil {
		t.Fatalf("get examination result: %v", err)
	}
	if result.Decision.DecisionCode != "monitoring" || result.Decision.RiskClass != "medium" {
		t.Fatalf("expected structured decision mapping, got %#v", result.Decision)
	}
	if len(result.ChannelReports) != 1 || result.ChannelReports[0].Channel != "text" {
		t.Fatalf("expected channel reports in result, got %#v", result.ChannelReports)
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
