package examinations

import (
	"context"
	"errors"
	"testing"
	"time"

	"diplom/internal/audit"
)

func TestCreateExaminationSnapshotsQuestions(t *testing.T) {
	questionnaireID := int64(12)
	repo := &examRepoStub{
		createResult: Examination{
			ID:              99,
			SpecialistID:    3,
			CreatedByUserID: 7,
			QuestionnaireID: &questionnaireID,
			Status:          StatusCreated,
			CreatedAt:       time.Unix(10, 0).UTC(),
			UpdatedAt:       time.Unix(10, 0).UTC(),
		},
	}
	service := NewService(repo)

	exam, err := service.Create(context.Background(), CreateInput{
		SpecialistID:    3,
		CreatedByUserID: 7,
		QuestionnaireID: &questionnaireID,
	})
	if err != nil {
		t.Fatalf("create examination: %v", err)
	}
	if exam.ID != 99 {
		t.Fatalf("expected examination id 99, got %d", exam.ID)
	}
	if repo.createInput == nil || repo.createInput.QuestionnaireID == nil || *repo.createInput.QuestionnaireID != questionnaireID {
		t.Fatal("expected repository create call to receive questionnaire_id for snapshotting")
	}
}

func TestCreateExaminationRejectsInvalidQuestionnaire(t *testing.T) {
	service := NewService(&examRepoStub{})

	if _, err := service.Create(context.Background(), CreateInput{
		SpecialistID:    3,
		CreatedByUserID: 7,
	}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for missing questionnaire_id, got %v", err)
	}

	invalidQuestionnaireID := int64(0)
	if _, err := service.Create(context.Background(), CreateInput{
		SpecialistID:    3,
		CreatedByUserID: 7,
		QuestionnaireID: &invalidQuestionnaireID,
	}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for invalid questionnaire_id, got %v", err)
	}
}

func TestExaminationFinishWritesSingleAuditEvent(t *testing.T) {
	questionnaireID := int64(12)
	repo := &examRepoStub{
		finishResult: Examination{
			ID:              99,
			SpecialistID:    3,
			CreatedByUserID: 7,
			QuestionnaireID: &questionnaireID,
			Status:          StatusReadyForProcessing,
		},
	}
	auditRepo := &examAuditRepoStub{}
	service := NewService(repo, audit.NewService(auditRepo))

	if _, err := service.Finish(context.Background(), 99); err != nil {
		t.Fatalf("finish examination: %v", err)
	}
	if len(auditRepo.events) != 1 {
		t.Fatalf("expected one audit event, got %d", len(auditRepo.events))
	}
	if auditRepo.events[0].Type != audit.EventTypeExaminationFinished {
		t.Fatalf("expected examination.finished event, got %q", auditRepo.events[0].Type)
	}
}

func TestGetByIDReturnsSnapshotQuestions(t *testing.T) {
	repo := &examRepoStub{
		getByIDResult: Examination{
			ID:           99,
			SpecialistID: 3,
			Status:       StatusCollectingAnswers,
		},
		listQuestionsResult: []ExaminationQuestion{
			{ID: 501, ExaminationID: 99, SpecialistID: 3, Position: 1, QuestionText: "Первый вопрос"},
			{ID: 502, ExaminationID: 99, SpecialistID: 3, Position: 2, QuestionText: "Второй вопрос"},
		},
	}
	service := NewService(repo)

	exam, err := service.GetByID(context.Background(), 99)
	if err != nil {
		t.Fatalf("get examination: %v", err)
	}
	if len(exam.Questions) != 2 {
		t.Fatalf("expected 2 snapshot questions, got %d", len(exam.Questions))
	}
	if exam.Questions[0].ID != 501 || exam.Questions[1].ID != 502 {
		t.Fatalf("unexpected question ids: %+v", exam.Questions)
	}
}

type examRepoStub struct {
	createInput  *CreateInput
	createResult Examination
	createErr    error
	finishResult Examination
	getByIDResult Examination
	listQuestionsResult []ExaminationQuestion
}

func (s *examRepoStub) Create(_ context.Context, input CreateInput) (Examination, error) {
	copy := input
	s.createInput = &copy
	if s.createErr != nil {
		return Examination{}, s.createErr
	}
	return s.createResult, nil
}

func (s *examRepoStub) List(context.Context) ([]Examination, error) {
	return nil, nil
}

func (s *examRepoStub) GetByID(context.Context, int64) (Examination, error) {
	return s.getByIDResult, nil
}

func (s *examRepoStub) ListQuestionsByExaminationID(context.Context, int64) ([]ExaminationQuestion, error) {
	return s.listQuestionsResult, nil
}

func (s *examRepoStub) ListBySpecialistID(context.Context, int64) ([]Examination, error) {
	return nil, nil
}

func (s *examRepoStub) UpdateStatus(context.Context, int64, string) (Examination, error) {
	return Examination{}, nil
}

func (s *examRepoStub) Finish(context.Context, int64) (Examination, error) {
	return s.finishResult, nil
}

type examAuditRepoStub struct {
	events []audit.Event
}

func (s *examAuditRepoStub) Append(_ context.Context, event audit.Event) error {
	s.events = append(s.events, event)
	return nil
}

func (s *examAuditRepoStub) List(context.Context, audit.ListFilter) ([]audit.Event, error) {
	return append([]audit.Event(nil), s.events...), nil
}
