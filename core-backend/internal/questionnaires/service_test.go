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
			ID:           7,
			Title:        "Q1",
			IsActive:     true,
			UsageCount:   0,
			LastUsedAt:   nil,
			LastEditedAt: time.Unix(10, 0).UTC(),
			LastEditor:   &QuestionnaireEditor{ID: 1, Login: "admin"},
			Questions:    []Question{{ID: 1, Text: "How?", Position: 1}},
			CreatedAt:    time.Unix(10, 0).UTC(),
			UpdatedAt:    time.Unix(10, 0).UTC(),
		},
		updateResult: Questionnaire{
			ID:           7,
			Title:        "Q1 updated",
			IsActive:     false,
			UsageCount:   4,
			LastUsedAt:   timePtr(time.Unix(15, 0).UTC()),
			LastEditedAt: time.Unix(20, 0).UTC(),
			LastEditor:   &QuestionnaireEditor{ID: 2, Login: "chief-admin"},
			Questions:    []Question{{ID: 1, Text: "How?", Position: 1}},
			CreatedAt:    time.Unix(10, 0).UTC(),
			UpdatedAt:    time.Unix(20, 0).UTC(),
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

func TestQuestionnaireReadModelsExposeUsageAndEditorMetadata(t *testing.T) {
	lastUsedAt := time.Unix(30, 0).UTC()
	lastEditedAt := time.Unix(40, 0).UTC()
	repo := &questionnaireRepoStub{
		listResult: []Questionnaire{
			{
				ID:           5,
				Title:        "Shift survey",
				IsActive:     true,
				UsageCount:   8,
				LastUsedAt:   &lastUsedAt,
				LastEditedAt: lastEditedAt,
				LastEditor:   &QuestionnaireEditor{ID: 3, Login: "admin"},
				Questions:    []Question{{ID: 1, Text: "How?", Position: 1}},
			},
		},
		getByIDResult: Questionnaire{
			ID:           5,
			Title:        "Shift survey",
			IsActive:     true,
			UsageCount:   8,
			LastUsedAt:   &lastUsedAt,
			LastEditedAt: lastEditedAt,
			LastEditor:   &QuestionnaireEditor{ID: 3, Login: "admin"},
			Questions:    []Question{{ID: 1, Text: "How?", Position: 1}},
		},
	}

	service := NewService(repo)

	items, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("list questionnaires: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one questionnaire, got %d", len(items))
	}
	if items[0].UsageCount != 8 {
		t.Fatalf("expected usage count 8, got %d", items[0].UsageCount)
	}
	if items[0].LastEditor == nil || items[0].LastEditor.Login != "admin" {
		t.Fatalf("expected last editor admin, got %#v", items[0].LastEditor)
	}

	item, err := service.GetByID(context.Background(), 5)
	if err != nil {
		t.Fatalf("get questionnaire: %v", err)
	}
	if item.LastUsedAt == nil || !item.LastUsedAt.Equal(lastUsedAt) {
		t.Fatalf("expected last used at %s, got %#v", lastUsedAt.Format(time.RFC3339), item.LastUsedAt)
	}
	if !item.LastEditedAt.Equal(lastEditedAt) {
		t.Fatalf("expected last edited at %s, got %s", lastEditedAt.Format(time.RFC3339), item.LastEditedAt.Format(time.RFC3339))
	}
}

type questionnaireRepoStub struct {
	createResult  Questionnaire
	updateResult  Questionnaire
	listResult    []Questionnaire
	getByIDResult Questionnaire
}

func (s *questionnaireRepoStub) List(context.Context) ([]Questionnaire, error) {
	return s.listResult, nil
}

func (s *questionnaireRepoStub) GetByID(context.Context, int64) (Questionnaire, error) {
	return s.getByIDResult, nil
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

func timePtr(value time.Time) *time.Time {
	return &value
}
