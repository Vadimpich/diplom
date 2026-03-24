package answers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"diplom/internal/examinations"
	"diplom/internal/repository"
)

func TestCreateAnswerPersistsQuestionLink(t *testing.T) {
	repo := &answerRepoStub{
		exam: examinations.Examination{ID: 10, SpecialistID: 7, Status: examinations.StatusCollectingAnswers},
		question: examinations.ExaminationQuestion{ID: 22, ExaminationID: 10, SpecialistID: 7},
		createResult: Answer{ID: 55, ExaminationID: 10, ExaminationQuestionID: 22, SpecialistID: 7},
	}
	storage := &answerStorageStub{}
	service := NewService(repo, storage)

	answer, err := service.Create(context.Background(), CreateInput{
		ExaminationID:         10,
		ExaminationQuestionID: 22,
		SpecialistID:          7,
		CreatedByUserID:       3,
		Text:                  "test",
		FileName:              "audio.webm",
		ContentType:           "audio/webm",
		Size:                  4,
		Content:               bytes.NewReader([]byte("test")),
	})
	if err != nil {
		t.Fatalf("create answer: %v", err)
	}
	if answer.ExaminationQuestionID != 22 {
		t.Fatalf("expected examination_question_id 22, got %d", answer.ExaminationQuestionID)
	}
	if repo.createParams.ExaminationQuestionID != 22 || repo.createParams.SpecialistID != 7 {
		t.Fatal("expected repository create params to preserve question and specialist linkage")
	}
}

func TestCreateAnswerRejectsMismatchedQuestion(t *testing.T) {
	repo := &answerRepoStub{
		exam: examinations.Examination{ID: 10, SpecialistID: 7, Status: examinations.StatusCollectingAnswers},
		question: examinations.ExaminationQuestion{ID: 22, ExaminationID: 11, SpecialistID: 7},
	}
	service := NewService(repo, &answerStorageStub{})

	_, err := service.Create(context.Background(), CreateInput{
		ExaminationID:         10,
		ExaminationQuestionID: 22,
		SpecialistID:          7,
		CreatedByUserID:       3,
		Text:                  "test",
		FileName:              "audio.webm",
		ContentType:           "audio/webm",
		Size:                  4,
		Content:               bytes.NewReader([]byte("test")),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateAnswerRejectsDuplicateQuestionAnswer(t *testing.T) {
	repo := &answerRepoStub{
		exam:       examinations.Examination{ID: 10, SpecialistID: 7, Status: examinations.StatusCollectingAnswers},
		question:   examinations.ExaminationQuestion{ID: 22, ExaminationID: 10, SpecialistID: 7},
		createErr:  repository.ErrConflict,
	}
	service := NewService(repo, &answerStorageStub{})

	_, err := service.Create(context.Background(), CreateInput{
		ExaminationID:         10,
		ExaminationQuestionID: 22,
		SpecialistID:          7,
		CreatedByUserID:       3,
		Text:                  "test",
		FileName:              "audio.webm",
		ContentType:           "audio/webm",
		Size:                  4,
		Content:               bytes.NewReader([]byte("test")),
	})
	if !errors.Is(err, repository.ErrConflict) {
		t.Fatalf("expected repository.ErrConflict, got %v", err)
	}
}

type answerRepoStub struct {
	exam        examinations.Examination
	question    examinations.ExaminationQuestion
	createParams CreateParams
	createResult Answer
	createErr    error
}

func (s *answerRepoStub) ReserveID(context.Context) (int64, error) {
	return 55, nil
}

func (s *answerRepoStub) Create(_ context.Context, params CreateParams) (Answer, error) {
	s.createParams = params
	if s.createErr != nil {
		return Answer{}, s.createErr
	}
	result := s.createResult
	result.CreatedAt = time.Unix(10, 0).UTC()
	result.AudioS3Key = params.AudioS3Key
	result.CreatedByUserID = params.CreatedByUserID
	result.Text = params.Text
	return result, nil
}

func (s *answerRepoStub) GetExaminationByID(context.Context, int64) (examinations.Examination, error) {
	return s.exam, nil
}

func (s *answerRepoStub) GetExaminationQuestionByID(context.Context, int64) (examinations.ExaminationQuestion, error) {
	return s.question, nil
}

type answerStorageStub struct{}

func (s *answerStorageStub) Upload(context.Context, string, string, int64, io.Reader) error {
	return nil
}

func (s *answerStorageStub) Delete(context.Context, string) error {
	return nil
}
