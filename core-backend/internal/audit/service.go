package audit

import (
	"context"
	"encoding/json"
	"time"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Append(ctx context.Context, event Event) error {
	if s == nil || s.repo == nil {
		return nil
	}
	if event.HappenedAt.IsZero() {
		event.HappenedAt = s.now()
	}
	if len(event.Payload) == 0 {
		event.Payload = json.RawMessage(`{}`)
	}
	return s.repo.Append(ctx, event)
}

func (s *Service) AppendFromContext(ctx context.Context, event Event) error {
	meta := MetadataFromContext(ctx)
	if event.RequestID == "" {
		event.RequestID = meta.RequestID
	}
	if event.TraceID == "" {
		event.TraceID = meta.TraceID
	}
	if event.TraceParent == "" {
		event.TraceParent = meta.TraceParent
	}
	if event.TraceState == "" {
		event.TraceState = meta.TraceState
	}
	if event.CorrelationID == "" {
		event.CorrelationID = meta.CorrelationID
	}
	if event.Actor.UserID == nil && event.Actor.Login == "" && event.Actor.RoleSlug == "" && event.Actor.IP == "" && event.Actor.UserAgent == "" {
		event.Actor = meta.Actor
	}
	return s.Append(ctx, event)
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]Event, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	return s.repo.List(ctx, filter)
}
