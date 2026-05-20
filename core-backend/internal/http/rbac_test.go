package http

import (
	"context"
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"diplom/internal/auth"
	"diplom/internal/questionnaires"
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

func TestOperatorCanListQuestionnairesButCannotMutateThem(t *testing.T) {
	router := NewRouter(Dependencies{
		AuthService:    &auth.Service{},
		AuthTokens:     staticTokenManager{claims: auth.Claims{UserID: 1, RoleSlug: "operator"}},
		AllowedOrigins: []string{"http://localhost:3000"},
		Questionnaires: questionnaires.NewService(questionnaireRepoStub{}),
	})

	t.Run("list", func(t *testing.T) {
		req := httptest.NewRequest(nethttp.MethodGet, "/questionnaires", nil)
		req.Header.Set("Authorization", "Bearer operator-token")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != nethttp.StatusOK {
			t.Fatalf("expected status %d, got %d", nethttp.StatusOK, rec.Code)
		}
	})

	t.Run("create", func(t *testing.T) {
		req := httptest.NewRequest(nethttp.MethodPost, "/questionnaires", strings.NewReader(`{"title":"Q1","is_active":true,"questions":[{"text":"How?"}]}`))
		req.Header.Set("Authorization", "Bearer operator-token")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != nethttp.StatusForbidden {
			t.Fatalf("expected status %d, got %d", nethttp.StatusForbidden, rec.Code)
		}
	})

	t.Run("update", func(t *testing.T) {
		req := httptest.NewRequest(nethttp.MethodPut, "/questionnaires/7", strings.NewReader(`{"title":"Q1","is_active":true,"questions":[{"text":"How?"}]}`))
		req.Header.Set("Authorization", "Bearer operator-token")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != nethttp.StatusForbidden {
			t.Fatalf("expected status %d, got %d", nethttp.StatusForbidden, rec.Code)
		}
	})
}

func TestAdminQuestionnaireRoutesRemainReachable(t *testing.T) {
	router := NewRouter(Dependencies{
		AuthService:    &auth.Service{},
		AuthTokens:     staticTokenManager{claims: auth.Claims{UserID: 1, RoleSlug: "admin"}},
		AllowedOrigins: []string{"http://localhost:3000"},
		Questionnaires: questionnaires.NewService(questionnaireRepoStub{}),
	})

	tests := []struct {
		name   string
		method string
		target string
		body   io.Reader
		status int
	}{
		{name: "list", method: nethttp.MethodGet, target: "/questionnaires", status: nethttp.StatusOK},
		{name: "detail", method: nethttp.MethodGet, target: "/questionnaires/7", status: nethttp.StatusOK},
		{name: "create", method: nethttp.MethodPost, target: "/questionnaires", body: strings.NewReader(`{"title":"Q1","is_active":true,"questions":[{"text":"How?"}]}`), status: nethttp.StatusCreated},
		{name: "update", method: nethttp.MethodPut, target: "/questionnaires/7", body: strings.NewReader(`{"title":"Q1","is_active":true,"questions":[{"text":"How?"}]}`), status: nethttp.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, tt.body)
			req.Header.Set("Authorization", "Bearer admin-token")
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.status {
				t.Fatalf("expected status %d, got %d", tt.status, rec.Code)
			}
		})
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

type questionnaireRepoStub struct{}

func (questionnaireRepoStub) List(context.Context) ([]questionnaires.Questionnaire, error) {
	return []questionnaires.Questionnaire{
		{
			ID:       7,
			Title:    "Предсменный опрос",
			IsActive: true,
			Questions: []questionnaires.Question{
				{ID: 11, Text: "Как вы себя чувствуете?", Position: 1},
			},
			CreatedAt: time.Unix(10, 0).UTC(),
			UpdatedAt: time.Unix(20, 0).UTC(),
		},
	}, nil
}

func (questionnaireRepoStub) GetByID(context.Context, int64) (questionnaires.Questionnaire, error) {
	return questionnaires.Questionnaire{
		ID:       7,
		Title:    "Предсменный опрос",
		IsActive: true,
		Questions: []questionnaires.Question{
			{ID: 11, Text: "Как вы себя чувствуете?", Position: 1},
		},
		CreatedAt: time.Unix(10, 0).UTC(),
		UpdatedAt: time.Unix(20, 0).UTC(),
	}, nil
}

func (questionnaireRepoStub) Create(context.Context, questionnaires.CreateInput) (questionnaires.Questionnaire, error) {
	return questionnaires.Questionnaire{
		ID:       8,
		Title:    "Созданный опросник",
		IsActive: true,
		Questions: []questionnaires.Question{
			{ID: 12, Text: "Новый вопрос", Position: 1},
		},
		CreatedAt: time.Unix(30, 0).UTC(),
		UpdatedAt: time.Unix(30, 0).UTC(),
	}, nil
}

func (questionnaireRepoStub) Update(context.Context, questionnaires.UpdateInput) (questionnaires.Questionnaire, error) {
	return questionnaires.Questionnaire{
		ID:       7,
		Title:    "Обновлённый опросник",
		IsActive: true,
		Questions: []questionnaires.Question{
			{ID: 11, Text: "Как вы себя чувствуете?", Position: 1},
		},
		CreatedAt: time.Unix(10, 0).UTC(),
		UpdatedAt: time.Unix(40, 0).UTC(),
	}, nil
}
