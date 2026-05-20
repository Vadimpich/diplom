package http

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"diplom/internal/auth"
	"diplom/internal/settings"
)

func TestAdminSettingsRouteRequiresAdminRole(t *testing.T) {
	router := NewRouter(Dependencies{
		Settings:       settings.NewService(&settingsHTTPRepoStub{}, settings.Defaults{AudioRetentionTTLDays: 30, ProcessingMaxAttempts: 3, KESMIMaxRetries: 2}),
		AuthService:    &auth.Service{},
		AuthTokens:     staticTokenManager{claims: auth.Claims{UserID: 2, RoleSlug: "operator"}},
		AllowedOrigins: []string{"http://localhost:3000"},
	})

	req := httptest.NewRequest(nethttp.MethodGet, "/settings", nil)
	req.Header.Set("Authorization", "Bearer operator-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusForbidden {
		t.Fatalf("expected status %d, got %d", nethttp.StatusForbidden, rec.Code)
	}
}

func TestAdminSettingsGet(t *testing.T) {
	router := NewRouter(Dependencies{
		Settings: settings.NewService(&settingsHTTPRepoStub{
			item: settings.RuntimeSettings{
				AudioRetentionTTLDays: 30,
				ProcessingMaxAttempts: 3,
				KESMIMaxRetries:       2,
				CreatedAt:             time.Date(2026, 3, 24, 11, 0, 0, 0, time.UTC),
				UpdatedAt:             time.Date(2026, 3, 24, 11, 30, 0, 0, time.UTC),
			},
		}, settings.Defaults{AudioRetentionTTLDays: 30, ProcessingMaxAttempts: 3, KESMIMaxRetries: 2}),
		AuthService:    &auth.Service{},
		AuthTokens:     staticTokenManager{claims: auth.Claims{UserID: 1, RoleSlug: "admin"}},
		AllowedOrigins: []string{"http://localhost:3000"},
	})

	req := httptest.NewRequest(nethttp.MethodGet, "/settings", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("expected status %d, got %d", nethttp.StatusOK, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"audio_retention_ttl_days":30`) {
		t.Fatalf("expected settings payload, got %s", rec.Body.String())
	}
}

func TestAdminSettingsUpdate(t *testing.T) {
	repo := &settingsHTTPRepoStub{
		item: settings.RuntimeSettings{
			AudioRetentionTTLDays: 30,
			ProcessingMaxAttempts: 3,
			KESMIMaxRetries:       2,
		},
	}
	router := NewRouter(Dependencies{
		Settings:       settings.NewService(repo, settings.Defaults{AudioRetentionTTLDays: 30, ProcessingMaxAttempts: 3, KESMIMaxRetries: 2}),
		AuthService:    &auth.Service{},
		AuthTokens:     staticTokenManager{claims: auth.Claims{UserID: 1, RoleSlug: "admin"}},
		AllowedOrigins: []string{"http://localhost:3000"},
	})

	req := httptest.NewRequest(nethttp.MethodPut, "/settings", strings.NewReader(`{"audio_retention_ttl_days":45,"processing_max_attempts":4,"kesmi_max_retries":3}`))
	req.Header.Set("Authorization", "Bearer admin-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("expected status %d, got %d", nethttp.StatusOK, rec.Code)
	}
	if repo.item.ProcessingMaxAttempts != 4 || repo.item.KESMIMaxRetries != 3 {
		t.Fatalf("expected updated settings in repo, got %+v", repo.item)
	}

	var payload settings.RuntimeSettings
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.AudioRetentionTTLDays != 45 {
		t.Fatalf("expected updated ttl, got %d", payload.AudioRetentionTTLDays)
	}
}

type settingsHTTPRepoStub struct {
	item settings.RuntimeSettings
}

func (s *settingsHTTPRepoStub) Get(context.Context) (settings.RuntimeSettings, error) {
	return s.item, nil
}

func (s *settingsHTTPRepoStub) Ensure(context.Context, settings.Defaults) (settings.RuntimeSettings, error) {
	return s.item, nil
}

func (s *settingsHTTPRepoStub) Update(_ context.Context, input settings.UpdateInput) (settings.RuntimeSettings, error) {
	s.item.AudioRetentionTTLDays = input.AudioRetentionTTLDays
	s.item.ProcessingMaxAttempts = input.ProcessingMaxAttempts
	s.item.KESMIMaxRetries = input.KESMIMaxRetries
	return s.item, nil
}
