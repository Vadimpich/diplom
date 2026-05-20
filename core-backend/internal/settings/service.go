package settings

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"diplom/internal/audit"
)

var ErrInvalidInput = errors.New("settings: invalid input")

type RuntimeSettings struct {
	AudioRetentionTTLDays int32     `json:"audio_retention_ttl_days"`
	ProcessingMaxAttempts int32     `json:"processing_max_attempts"`
	KESMIMaxRetries       int32     `json:"kesmi_max_retries"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type Defaults struct {
	AudioRetentionTTLDays int32
	ProcessingMaxAttempts int32
	KESMIMaxRetries       int32
}

type UpdateInput struct {
	AudioRetentionTTLDays int32 `json:"audio_retention_ttl_days"`
	ProcessingMaxAttempts int32 `json:"processing_max_attempts"`
	KESMIMaxRetries       int32 `json:"kesmi_max_retries"`
}

type Repository interface {
	Get(context.Context) (RuntimeSettings, error)
	Ensure(context.Context, Defaults) (RuntimeSettings, error)
	Update(context.Context, UpdateInput) (RuntimeSettings, error)
}

type Service struct {
	repo     Repository
	defaults Defaults
	auditor  *audit.Service
}

func NewService(repo Repository, defaults Defaults, auditors ...*audit.Service) *Service {
	var auditor *audit.Service
	if len(auditors) > 0 {
		auditor = auditors[0]
	}
	return &Service{
		repo:     repo,
		defaults: defaults,
		auditor:  auditor,
	}
}

func (s *Service) EnsureDefaults(ctx context.Context) (RuntimeSettings, error) {
	if s == nil || s.repo == nil {
		return RuntimeSettings{}, nil
	}
	if err := validateDefaults(s.defaults); err != nil {
		return RuntimeSettings{}, err
	}
	return s.repo.Ensure(ctx, s.defaults)
}

func (s *Service) Get(ctx context.Context) (RuntimeSettings, error) {
	if s == nil || s.repo == nil {
		return RuntimeSettings{}, nil
	}
	return s.repo.Get(ctx)
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (RuntimeSettings, error) {
	if s == nil || s.repo == nil {
		return RuntimeSettings{}, nil
	}
	if err := validateUpdate(input); err != nil {
		return RuntimeSettings{}, err
	}
	item, err := s.repo.Update(ctx, input)
	if err != nil {
		return RuntimeSettings{}, err
	}
	s.appendAudit(ctx, item)
	return item, nil
}

func (s *Service) ProcessingMaxAttempts(ctx context.Context) (int32, error) {
	item, err := s.Get(ctx)
	if err != nil {
		return 0, err
	}
	return item.ProcessingMaxAttempts, nil
}

func (s *Service) KESMIMaxRetries(ctx context.Context) (int32, error) {
	item, err := s.Get(ctx)
	if err != nil {
		return 0, err
	}
	return item.KESMIMaxRetries, nil
}

func validateDefaults(input Defaults) error {
	return validateUpdate(UpdateInput{
		AudioRetentionTTLDays: input.AudioRetentionTTLDays,
		ProcessingMaxAttempts: input.ProcessingMaxAttempts,
		KESMIMaxRetries:       input.KESMIMaxRetries,
	})
}

func validateUpdate(input UpdateInput) error {
	switch {
	case input.AudioRetentionTTLDays < 1 || input.AudioRetentionTTLDays > 365:
		return ErrInvalidInput
	case input.ProcessingMaxAttempts < 1 || input.ProcessingMaxAttempts > 10:
		return ErrInvalidInput
	case input.KESMIMaxRetries < 1 || input.KESMIMaxRetries > 10:
		return ErrInvalidInput
	default:
		return nil
	}
}

func (s *Service) appendAudit(ctx context.Context, item RuntimeSettings) {
	if s.auditor == nil {
		return
	}
	payload, _ := json.Marshal(map[string]int32{
		"audio_retention_ttl_days": item.AudioRetentionTTLDays,
		"processing_max_attempts":  item.ProcessingMaxAttempts,
		"kesmi_max_retries":        item.KESMIMaxRetries,
	})
	_ = s.auditor.AppendFromContext(ctx, audit.Event{
		Type:    audit.EventTypeAdminSettingsUpdated,
		Outcome: audit.OutcomeSucceeded,
		Resource: audit.ResourceRef{
			Kind: "system_settings",
			ID:   1,
		},
		Payload: payload,
	})
}
