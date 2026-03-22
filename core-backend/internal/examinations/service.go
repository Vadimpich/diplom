package examinations

import (
	"context"
	"errors"
	"time"
)

const (
	StatusCreated            = "created"
	StatusCollectingAnswers  = "collecting_answers"
	StatusReadyForProcessing = "ready_for_processing"
	StatusProcessing         = "processing"
	StatusAggregating        = "aggregating"
	StatusAggregated         = "aggregated"
	StatusFailed             = "failed"
)

var (
	ErrInvalidInput      = errors.New("examinations: invalid input")
	ErrInvalidTransition = errors.New("examinations: invalid transition")
	ErrAnswersIncomplete = errors.New("examinations: answers incomplete")
)

type Examination struct {
	ID              int64      `json:"id"`
	SpecialistID    int64      `json:"specialist_id"`
	CreatedByUserID int64      `json:"created_by_user_id"`
	QuestionnaireID *int64     `json:"questionnaire_id,omitempty"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ExaminationQuestion struct {
	ID               int64
	ExaminationID    int64
	SpecialistID     int64
	QuestionnaireID  int64
	SourceQuestionID *int64
	Position         int32
	QuestionText     string
}

type CreateInput struct {
	SpecialistID    int64
	CreatedByUserID int64
	QuestionnaireID *int64
}

type Repository interface {
	Create(context.Context, CreateInput) (Examination, error)
	List(context.Context) ([]Examination, error)
	GetByID(context.Context, int64) (Examination, error)
	ListBySpecialistID(context.Context, int64) ([]Examination, error)
	UpdateStatus(context.Context, int64, string) (Examination, error)
	Finish(context.Context, int64) (Examination, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Examination, error) {
	if input.SpecialistID <= 0 || input.CreatedByUserID <= 0 {
		return Examination{}, ErrInvalidInput
	}
	if input.QuestionnaireID == nil || *input.QuestionnaireID <= 0 {
		return Examination{}, ErrInvalidInput
	}
	return s.repo.Create(ctx, input)
}

func (s *Service) List(ctx context.Context) ([]Examination, error) {
	return s.repo.List(ctx)
}

func (s *Service) Start(ctx context.Context, id int64) (Examination, error) {
	exam, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Examination{}, err
	}
	if exam.Status == StatusCollectingAnswers {
		return exam, nil
	}
	if exam.Status != StatusCreated {
		return Examination{}, ErrInvalidTransition
	}
	return s.repo.UpdateStatus(ctx, id, StatusCollectingAnswers)
}

func (s *Service) GetByID(ctx context.Context, id int64) (Examination, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListBySpecialistID(ctx context.Context, specialistID int64) ([]Examination, error) {
	if specialistID <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.ListBySpecialistID(ctx, specialistID)
}

func (s *Service) Finish(ctx context.Context, id int64) (Examination, error) {
	return s.repo.Finish(ctx, id)
}
