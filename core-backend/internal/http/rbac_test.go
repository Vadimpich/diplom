package http

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"dimplom/internal/auth"
)

func TestRequireRoles(t *testing.T) {
	handler := RequireRoles("admin")(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.WriteHeader(nethttp.StatusNoContent)
	}))

	req := httptest.NewRequest(nethttp.MethodGet, "/admin/users", nil)
	req = req.WithContext(context.WithValue(req.Context(), claimsContextKey, auth.Claims{UserID: 1, RoleSlug: "operator"}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusForbidden {
		t.Fatalf("expected status %d, got %d", nethttp.StatusForbidden, rec.Code)
	}
}

func TestAdminRoutes(t *testing.T) {
	router := NewRouter(Dependencies{
		AuthService:    &auth.Service{},
		AuthTokens:     staticTokenManager{claims: auth.Claims{UserID: 1, RoleSlug: "operator"}},
		AllowedOrigins: []string{"http://localhost:3000"},
	})

	req := httptest.NewRequest(nethttp.MethodGet, "/users", nil)
	req.Header.Set("Authorization", "Bearer operator-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusForbidden {
		t.Fatalf("expected status %d, got %d", nethttp.StatusForbidden, rec.Code)
	}
}

func TestOperatorRoutes(t *testing.T) {
	router := NewRouter(Dependencies{
		AuthService:    &auth.Service{},
		AuthTokens:     staticTokenManager{err: auth.ErrInvalidToken},
		AllowedOrigins: []string{"http://localhost:3000"},
	})

	req := httptest.NewRequest(nethttp.MethodGet, "/specialists", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", nethttp.StatusUnauthorized, rec.Code)
	}
}

type staticTokenManager struct {
	claims auth.Claims
	err    error
}

func (s staticTokenManager) Issue(auth.User) (string, int64, error) {
	return "", 0, nil
}

func (s staticTokenManager) Parse(string) (auth.Claims, error) {
	if s.err != nil {
		return auth.Claims{}, s.err
	}
	return s.claims, nil
}
