package questionnaires

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"diplom/internal/audit"
)

var ErrInvalidInput = errors.New("questionnaires: invalid input")

type Question struct {
	ID       int64  `json:"id"`
	Text     string `json:"text"`
	Position int32  `json:"position"`
}

type Questionnaire struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	IsActive    bool       `json:"is_active"`
	Questions   []Question `json:"questions"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type QuestionInput struct {
	Text string
}

type CreateInput struct {
	Title       string
	Description *string
	IsActive    bool
	Questions   []QuestionInput
}

type UpdateInput struct {
	ID          int64
	Title       string
	Description *string
	IsActive    bool
	Questions   []QuestionInput
}

type Repository interface {
	List(context.Context) ([]Questionnaire, error)
	GetByID(context.Context, int64) (Questionnaire, error)
	Create(context.Context, CreateInput) (Questionnaire, error)
	Update(context.Context, UpdateInput) (Questionnaire, error)
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

func (s *Service) List(ctx context.Context) ([]Questionnaire, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int64) (Questionnaire, error) {
	if id <= 0 {
		return Questionnaire{}, ErrInvalidInput
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Questionnaire, error) {
	input = normalizeCreate(input)
	if err := validate(input.Title, input.Questions); err != nil {
		return Questionnaire{}, err
	}
	item, err := s.repo.Create(ctx, input)
	if err != nil {
		return Questionnaire{}, err
	}
	s.appendAudit(ctx, audit.Event{
		Type:    audit.EventTypeQuestionnaireCreated,
		Key:     "questionnaire-created:" + strings.TrimSpace(item.Title),
		Outcome: audit.OutcomeSucceeded,
		Resource: audit.ResourceRef{
			Kind: "questionnaire",
			ID:   item.ID,
		},
		DomainRefs: audit.DomainRefs{QuestionnaireID: &item.ID},
		Payload:    questionnairePayload(item),
	})
	return item, nil
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (Questionnaire, error) {
	input = normalizeUpdate(input)
	if input.ID <= 0 {
		return Questionnaire{}, ErrInvalidInput
	}
	if err := validate(input.Title, input.Questions); err != nil {
		return Questionnaire{}, err
	}
	item, err := s.repo.Update(ctx, input)
	if err != nil {
		return Questionnaire{}, err
	}
	s.appendAudit(ctx, audit.Event{
		Type:    audit.EventTypeQuestionnaireUpdated,
		Key:     "questionnaire-updated:" + strings.TrimSpace(item.Title),
		Outcome: audit.OutcomeSucceeded,
		Resource: audit.ResourceRef{
			Kind: "questionnaire",
			ID:   item.ID,
		},
		DomainRefs: audit.DomainRefs{QuestionnaireID: &item.ID},
		Payload:    questionnairePayload(item),
	})
	return item, nil
}

func validate(title string, questions []QuestionInput) error {
	if title == "" || len(questions) == 0 {
		return ErrInvalidInput
	}
	return nil
}

func normalizeCreate(input CreateInput) CreateInput {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = normalizeDescription(input.Description)
	input.Questions = normalizeQuestions(input.Questions)
	return input
}

func normalizeUpdate(input UpdateInput) UpdateInput {
	return UpdateInput{
		ID:          input.ID,
		Title:       strings.TrimSpace(input.Title),
		Description: normalizeDescription(input.Description),
		IsActive:    input.IsActive,
		Questions:   normalizeQuestions(input.Questions),
	}
}

func normalizeDescription(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeQuestions(values []QuestionInput) []QuestionInput {
	result := make([]QuestionInput, 0, len(values))
	for _, value := range values {
		text := strings.TrimSpace(value.Text)
		if text == "" {
			continue
		}
		result = append(result, QuestionInput{Text: text})
	}
	return result
}

func (s *Service) appendAudit(ctx context.Context, event audit.Event) {
	if s.auditor == nil {
		return
	}
	_ = s.auditor.AppendFromContext(ctx, event)
}

func questionnairePayload(item Questionnaire) json.RawMessage {
	data, _ := json.Marshal(map[string]any{
		"title":          item.Title,
		"is_active":      item.IsActive,
		"question_count": len(item.Questions),
	})
	return data
}
