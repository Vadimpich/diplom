package questionnaires

import (
	"context"
	"testing"
	"time"

	"diplom/internal/audit"
)

func TestQuestionnaireMutationWritesAuditEvent(t *testing.T) {
	repo := &questionnaireRepoStub{
		createResult: Questionnaire{
			ID:        7,
			Title:     "Q1",
			IsActive:  true,
			Questions: []Question{{ID: 1, Text: "How?", Position: 1}},
			CreatedAt: time.Unix(10, 0).UTC(),
			UpdatedAt: time.Unix(10, 0).UTC(),
		},
		updateResult: Questionnaire{
			ID:        7,
			Title:     "Q1 updated",
			IsActive:  false,
			Questions: []Question{{ID: 1, Text: "How?", Position: 1}},
			CreatedAt: time.Unix(10, 0).UTC(),
			UpdatedAt: time.Unix(20, 0).UTC(),
		},
	}
	auditRepo := &questionnaireAuditRepoStub{}
	service := NewService(repo, audit.NewService(auditRepo))

	if _, err := service.Create(context.Background(), CreateInput{
		Title:     " Q1 ",
		IsActive:  true,
		Questions: []QuestionInput{{Text: "How?"}},
	}); err != nil {
		t.Fatalf("create questionnaire: %v", err)
	}
	if _, err := service.Update(context.Background(), UpdateInput{
		ID:        7,
		Title:     " Q1 updated ",
		IsActive:  false,
		Questions: []QuestionInput{{Text: "How?"}},
	}); err != nil {
		t.Fatalf("update questionnaire: %v", err)
	}
	if len(auditRepo.events) != 2 {
		t.Fatalf("expected two audit events, got %d", len(auditRepo.events))
	}
	if auditRepo.events[0].Type != audit.EventTypeQuestionnaireCreated || auditRepo.events[1].Type != audit.EventTypeQuestionnaireUpdated {
		t.Fatalf("unexpected audit sequence: %#v", auditRepo.events)
	}
}

type questionnaireRepoStub struct {
	createResult Questionnaire
	updateResult Questionnaire
}

func (s *questionnaireRepoStub) List(context.Context) ([]Questionnaire, error) {
	return nil, nil
}

func (s *questionnaireRepoStub) GetByID(context.Context, int64) (Questionnaire, error) {
	return Questionnaire{}, nil
}

func (s *questionnaireRepoStub) Create(context.Context, CreateInput) (Questionnaire, error) {
	return s.createResult, nil
}

func (s *questionnaireRepoStub) Update(context.Context, UpdateInput) (Questionnaire, error) {
	return s.updateResult, nil
}

type questionnaireAuditRepoStub struct {
	events []audit.Event
}

func (s *questionnaireAuditRepoStub) Append(_ context.Context, event audit.Event) error {
	s.events = append(s.events, event)
	return nil
}

func (s *questionnaireAuditRepoStub) List(context.Context, audit.ListFilter) ([]audit.Event, error) {
	return append([]audit.Event(nil), s.events...), nil
}
