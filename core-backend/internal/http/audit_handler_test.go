package http

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"diplom/internal/audit"
	"diplom/internal/auth"
)

func TestAdminAuditEventsRouteRequiresAdminRole(t *testing.T) {
	router := NewRouter(Dependencies{
		Audit:          audit.NewService(&auditRepoStub{}),
		AuthService:    &auth.Service{},
		AuthTokens:     staticTokenManager{claims: auth.Claims{UserID: 2, RoleSlug: "operator"}},
		AllowedOrigins: []string{"http://localhost:3000"},
	})

	req := httptest.NewRequest(nethttp.MethodGet, "/audit/events", nil)
	req.Header.Set("Authorization", "Bearer operator-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusForbidden {
		t.Fatalf("expected status %d, got %d", nethttp.StatusForbidden, rec.Code)
	}
}

func TestAdminAuditEventsRejectsResourceIDWithoutKind(t *testing.T) {
	router := NewRouter(Dependencies{
		Audit:          audit.NewService(&auditRepoStub{}),
		AuthService:    &auth.Service{},
		AuthTokens:     staticTokenManager{claims: auth.Claims{UserID: 1, RoleSlug: "admin"}},
		AllowedOrigins: []string{"http://localhost:3000"},
	})

	req := httptest.NewRequest(nethttp.MethodGet, "/audit/events?resource_id=55", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", nethttp.StatusBadRequest, rec.Code)
	}
}

func TestAdminAuditEventsListsFilteredResults(t *testing.T) {
	repo := &auditRepoStub{
		listResult: []audit.Event{
			{
				ID:         901,
				Type:       audit.EventTypeDecisionCompleted,
				Outcome:    audit.OutcomeSucceeded,
				HappenedAt: time.Date(2026, 3, 24, 16, 30, 0, 0, time.UTC),
				Resource: audit.ResourceRef{
					Kind: "decision_snapshot",
					ID:   44,
				},
				DomainRefs: audit.DomainRefs{
					ExaminationID: int64PtrAudit(101),
				},
			},
		},
	}

	router := NewRouter(Dependencies{
		Audit:          audit.NewService(repo),
		AuthService:    &auth.Service{},
		AuthTokens:     staticTokenManager{claims: auth.Claims{UserID: 1, RoleSlug: "admin"}},
		AllowedOrigins: []string{"http://localhost:3000"},
	})

	req := httptest.NewRequest(
		nethttp.MethodGet,
		"/audit/events?event_type=decision.completed&resource_kind=decision_snapshot&resource_id=44&from=2026-03-24T16:00:00Z&to=2026-03-24T17:00:00Z&limit=25",
		nil,
	)
	req.Header.Set("Authorization", "Bearer admin-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("expected status %d, got %d", nethttp.StatusOK, rec.Code)
	}
	if repo.lastFilter.EventType == nil || *repo.lastFilter.EventType != audit.EventTypeDecisionCompleted {
		t.Fatalf("expected event_type filter to be forwarded, got %+v", repo.lastFilter.EventType)
	}
	if repo.lastFilter.ResourceKind == nil || *repo.lastFilter.ResourceKind != "decision_snapshot" {
		t.Fatalf("expected resource_kind filter to be forwarded, got %+v", repo.lastFilter.ResourceKind)
	}
	if repo.lastFilter.ResourceID == nil || *repo.lastFilter.ResourceID != 44 {
		t.Fatalf("expected resource_id filter to be forwarded, got %+v", repo.lastFilter.ResourceID)
	}
	if repo.lastFilter.Limit != 25 {
		t.Fatalf("expected limit 25, got %d", repo.lastFilter.Limit)
	}

	var payload auditEventsResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected one audit event, got %d", len(payload.Items))
	}
	if payload.Items[0].EventType != audit.EventTypeDecisionCompleted {
		t.Fatalf("expected decision.completed event, got %q", payload.Items[0].EventType)
	}
	if payload.Items[0].Resource == nil || payload.Items[0].Resource.Kind != "decision_snapshot" {
		t.Fatalf("expected decision_snapshot resource, got %#v", payload.Items[0].Resource)
	}
}

type auditRepoStub struct {
	listResult []audit.Event
	lastFilter audit.ListFilter
}

func (s *auditRepoStub) Append(context.Context, audit.Event) error {
	return nil
}

func (s *auditRepoStub) List(_ context.Context, filter audit.ListFilter) ([]audit.Event, error) {
	s.lastFilter = filter
	return append([]audit.Event(nil), s.listResult...), nil
}

func int64PtrAudit(value int64) *int64 {
	return &value
}
