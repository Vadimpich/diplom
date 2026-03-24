package answers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"diplom/internal/examinations"
	"diplom/internal/repository"
)

var (
	ErrInvalidInput        = errors.New("answers: invalid input")
	ErrExaminationNotReady = errors.New("answers: examination is not collecting answers")
)

type Answer struct {
	ID                  int64     `json:"id"`
	ExaminationID       int64     `json:"examination_id"`
	ExaminationQuestionID int64   `json:"examination_question_id"`
	SpecialistID        int64     `json:"specialist_id"`
	CreatedByUserID     int64     `json:"created_by_user_id"`
	Text                string    `json:"text"`
	AudioS3Key          string    `json:"audio_s3_key"`
	CreatedAt           time.Time `json:"created_at"`
}

type CreateInput struct {
	ExaminationID         int64
	ExaminationQuestionID int64
	SpecialistID          int64
	CreatedByUserID       int64
	Text                  string
	FileName              string
	ContentType           string
	Size                  int64
	Content               io.Reader
}

type Repository interface {
	ReserveID(context.Context) (int64, error)
	Create(context.Context, CreateParams) (Answer, error)
	GetExaminationByID(context.Context, int64) (examinations.Examination, error)
	GetExaminationQuestionByID(context.Context, int64) (examinations.ExaminationQuestion, error)
}

type Storage interface {
	Upload(ctx context.Context, key, contentType string, size int64, body io.Reader) error
	Delete(context.Context, string) error
}

type CreateParams struct {
	ID                    int64
	ExaminationID         int64
	ExaminationQuestionID int64
	SpecialistID          int64
	CreatedByUserID       int64
	Text                  string
	AudioS3Key            string
}

type Service struct {
	repo    Repository
	storage Storage
}

func NewService(repo Repository, storage Storage) *Service {
	return &Service{
		repo:    repo,
		storage: storage,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Answer, error) {
	input.Text = strings.TrimSpace(input.Text)
	if input.ExaminationID <= 0 || input.ExaminationQuestionID <= 0 || input.SpecialistID <= 0 || input.CreatedByUserID <= 0 || input.Text == "" || input.Content == nil || input.FileName == "" {
		return Answer{}, ErrInvalidInput
	}

	exam, err := s.repo.GetExaminationByID(ctx, input.ExaminationID)
	if err != nil {
		return Answer{}, err
	}
	if exam.Status != examinations.StatusCollectingAnswers {
		return Answer{}, ErrExaminationNotReady
	}
	question, err := s.repo.GetExaminationQuestionByID(ctx, input.ExaminationQuestionID)
	if err != nil {
		return Answer{}, err
	}
	if question.ExaminationID != input.ExaminationID || question.SpecialistID != input.SpecialistID || exam.SpecialistID != input.SpecialistID {
		return Answer{}, ErrInvalidInput
	}

	answerID, err := s.repo.ReserveID(ctx)
	if err != nil {
		return Answer{}, err
	}

	key := buildAudioKey(input.ExaminationID, answerID, input.FileName)
	if err := s.storage.Upload(ctx, key, input.ContentType, input.Size, input.Content); err != nil {
		return Answer{}, fmt.Errorf("upload audio: %w", err)
	}

	answer, err := s.repo.Create(ctx, CreateParams{
		ID:                    answerID,
		ExaminationID:         input.ExaminationID,
		ExaminationQuestionID: input.ExaminationQuestionID,
		SpecialistID:          input.SpecialistID,
		CreatedByUserID:       input.CreatedByUserID,
		Text:                  input.Text,
		AudioS3Key:            key,
	})
	if err != nil {
		_ = s.storage.Delete(ctx, key)
		if errors.Is(err, repository.ErrConflict) {
			return Answer{}, repository.ErrConflict
		}
		return Answer{}, err
	}

	return answer, nil
}

func buildAudioKey(examinationID, answerID int64, fileName string) string {
	ext := path.Ext(fileName)
	if ext == "" {
		ext = ".bin"
	}
	return fmt.Sprintf("examinations/%d/answers/%d/audio%s", examinationID, answerID, ext)
}
