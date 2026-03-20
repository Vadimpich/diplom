package examinations

import (
	"context"
	"errors"
	"testing"
	"time"
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

type examRepoStub struct {
	createInput  *CreateInput
	createResult Examination
	createErr    error
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
	return Examination{}, nil
}

func (s *examRepoStub) ListBySpecialistID(context.Context, int64) ([]Examination, error) {
	return nil, nil
}

func (s *examRepoStub) UpdateStatus(context.Context, int64, string) (Examination, error) {
	return Examination{}, nil
}

func (s *examRepoStub) Finish(context.Context, int64) (Examination, error) {
	return Examination{}, nil
}
