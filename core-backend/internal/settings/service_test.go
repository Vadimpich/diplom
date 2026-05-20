package settings

import (
	"context"
	"testing"
	"time"

	"diplom/internal/audit"
)

func TestServiceUpdateValidatesInput(t *testing.T) {
	service := NewService(&settingsRepoStub{}, Defaults{
		AudioRetentionTTLDays: 30,
		ProcessingMaxAttempts: 3,
		KESMIMaxRetries:       2,
	})

	_, err := service.Update(context.Background(), UpdateInput{
		AudioRetentionTTLDays: 0,
		ProcessingMaxAttempts: 3,
		KESMIMaxRetries:       2,
	})
	if err != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestServiceUpdateAppendsAuditEvent(t *testing.T) {
	auditRepo := &auditRepoStub{}
	service := NewService(&settingsRepoStub{
		item: RuntimeSettings{
			AudioRetentionTTLDays: 45,
			ProcessingMaxAttempts: 4,
			KESMIMaxRetries:       3,
			CreatedAt:             time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2026, 3, 24, 10, 5, 0, 0, time.UTC),
		},
	}, Defaults{
		AudioRetentionTTLDays: 30,
		ProcessingMaxAttempts: 3,
		KESMIMaxRetries:       2,
	}, audit.NewService(auditRepo))

	_, err := service.Update(context.Background(), UpdateInput{
		AudioRetentionTTLDays: 45,
		ProcessingMaxAttempts: 4,
		KESMIMaxRetries:       3,
	})
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if len(auditRepo.events) != 1 {
		t.Fatalf("expected one audit event, got %d", len(auditRepo.events))
	}
	if auditRepo.events[0].Type != audit.EventTypeAdminSettingsUpdated {
		t.Fatalf("expected admin.settings_updated, got %q", auditRepo.events[0].Type)
	}
}

type settingsRepoStub struct {
	item RuntimeSettings
}

func (s *settingsRepoStub) Get(context.Context) (RuntimeSettings, error) {
	return s.item, nil
}

func (s *settingsRepoStub) Ensure(context.Context, Defaults) (RuntimeSettings, error) {
	return s.item, nil
}

func (s *settingsRepoStub) Update(_ context.Context, input UpdateInput) (RuntimeSettings, error) {
	s.item.AudioRetentionTTLDays = input.AudioRetentionTTLDays
	s.item.ProcessingMaxAttempts = input.ProcessingMaxAttempts
	s.item.KESMIMaxRetries = input.KESMIMaxRetries
	return s.item, nil
}

type auditRepoStub struct {
	events []audit.Event
}

func (s *auditRepoStub) Append(_ context.Context, event audit.Event) error {
	s.events = append(s.events, event)
	return nil
}

func (s *auditRepoStub) List(context.Context, audit.ListFilter) ([]audit.Event, error) {
	return nil, nil
}
