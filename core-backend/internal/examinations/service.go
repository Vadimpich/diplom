package examinations

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"diplom/internal/audit"
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
	repo    Repository
	auditor *audit.Service
}

func NewService(repo Repository, auditors ...*audit.Service) *Service {
	var auditor *audit.Service
	if len(auditors) > 0 {
		auditor = auditors[0]
	}
	return &Service{repo: repo, auditor: auditor}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Examination, error) {
	if input.SpecialistID <= 0 || input.CreatedByUserID <= 0 {
		return Examination{}, ErrInvalidInput
	}
	if input.QuestionnaireID == nil || *input.QuestionnaireID <= 0 {
		return Examination{}, ErrInvalidInput
	}
	exam, err := s.repo.Create(ctx, input)
	if err != nil {
		return Examination{}, err
	}
	s.appendAudit(ctx, audit.Event{
		Type:    audit.EventTypeExaminationCreated,
		Key:     eventKey("created", exam.ID),
		Outcome: audit.OutcomeSucceeded,
		Resource: audit.ResourceRef{
			Kind: "examination",
			ID:   exam.ID,
		},
		DomainRefs: audit.DomainRefs{
			ExaminationID:   &exam.ID,
			SpecialistID:    &exam.SpecialistID,
			QuestionnaireID: exam.QuestionnaireID,
		},
		Payload: examinationPayload("", exam.Status),
	})
	return exam, nil
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
	updated, err := s.repo.UpdateStatus(ctx, id, StatusCollectingAnswers)
	if err != nil {
		return Examination{}, err
	}
	s.appendAudit(ctx, audit.Event{
		Type:    audit.EventTypeExaminationStarted,
		Key:     eventKey("started", updated.ID),
		Outcome: audit.OutcomeSucceeded,
		Resource: audit.ResourceRef{
			Kind: "examination",
			ID:   updated.ID,
		},
		DomainRefs: audit.DomainRefs{
			ExaminationID:   &updated.ID,
			SpecialistID:    &updated.SpecialistID,
			QuestionnaireID: updated.QuestionnaireID,
		},
		Payload: examinationPayload(exam.Status, updated.Status),
	})
	return updated, nil
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
	exam, err := s.repo.Finish(ctx, id)
	if err != nil {
		return Examination{}, err
	}
	s.appendAudit(ctx, audit.Event{
		Type:    audit.EventTypeExaminationFinished,
		Key:     eventKey("finished", exam.ID),
		Outcome: audit.OutcomeSucceeded,
		Resource: audit.ResourceRef{
			Kind: "examination",
			ID:   exam.ID,
		},
		DomainRefs: audit.DomainRefs{
			ExaminationID:   &exam.ID,
			SpecialistID:    &exam.SpecialistID,
			QuestionnaireID: exam.QuestionnaireID,
		},
		Payload: examinationPayload(StatusCollectingAnswers, exam.Status),
	})
	return exam, nil
}

func (s *Service) appendAudit(ctx context.Context, event audit.Event) {
	if s.auditor == nil {
		return
	}
	_ = s.auditor.AppendFromContext(ctx, event)
}

func eventKey(prefix string, examinationID int64) string {
	return prefix + ":" + "examination:" + strconv.FormatInt(examinationID, 10)
}

func examinationPayload(from, to string) json.RawMessage {
	data, _ := json.Marshal(map[string]any{
		"status_from": from,
		"status_to":   to,
	})
	return data
}
