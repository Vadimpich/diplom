package http

import (
	"bytes"
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"diplom/internal/auth"
	"diplom/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin(t *testing.T) {
	fixture := newAuthHandlerForTest(t)

	body := bytes.NewBufferString(`{"login":"operator","password":"secret"}`)
	req := httptest.NewRequest(nethttp.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "handler-test")
	rec := httptest.NewRecorder()

	fixture.handler.Login(rec, req)

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["access_token"] == "" {
		t.Fatal("expected access_token in response")
	}
	if payload["refresh_token"] == "" {
		t.Fatal("expected refresh_token in response")
	}
}

func TestLogoutRevokesSession(t *testing.T) {
	fixture := newAuthHandlerForTest(t)
	refreshToken := fixture.repo.seedRefreshToken

	body := bytes.NewBufferString(`{"refresh_token":"` + refreshToken + `"}`)
	req := httptest.NewRequest(nethttp.MethodPost, "/auth/logout", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	fixture.handler.Logout(rec, req)

	if rec.Code != nethttp.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}

	session := fixture.repo.refreshSessions[string(auth.HashRefreshToken(refreshToken))]
	if session.RevokedAt == nil {
		t.Fatal("expected refresh session to be revoked")
	}
}

func TestMe(t *testing.T) {
	fixture := newAuthHandlerForTest(t)
	jwtManager := auth.NewJWTManager("test-suite", "access-secret", 15*time.Minute)
	token, _, err := jwtManager.Issue(auth.User{
		ID:    1,
		Login: "operator",
		Role: auth.Role{
			ID:   2,
			Slug: "operator",
			Name: "Operator",
		},
		IsActive: true,
		Created:  fixture.repo.usersByID[1].Created,
		Updated:  fixture.repo.usersByID[1].Updated,
	})
	if err != nil {
		t.Fatalf("issue jwt: %v", err)
	}

	router := NewRouter(Dependencies{
		AuthService:    fixture.service,
		AuthTokens:     jwtManager,
		AllowedOrigins: []string{"http://localhost:3000"},
	})

	req := httptest.NewRequest(nethttp.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

type authHandlerFixture struct {
	handler AuthHandler
	service *auth.Service
	repo    *httpAuthRepoStub
}

func newAuthHandlerForTest(t *testing.T) authHandlerFixture {
	t.Helper()

	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	repo := newHTTPAuthRepoStub(t, now)
	service := auth.NewService(repo, &httpTokenManager{
		token:     "access-1",
		expiresIn: 900,
	})

	return authHandlerFixture{
		handler: AuthHandler{service: service},
		service: service,
		repo:    repo,
	}
}

type httpTokenManager struct {
	token     string
	expiresIn int64
}

func (m *httpTokenManager) Issue(auth.User) (string, int64, error) {
	return m.token, m.expiresIn, nil
}

func (m *httpTokenManager) Parse(string) (auth.Claims, error) {
	return auth.Claims{}, auth.ErrInvalidToken
}

type httpAuthRepoStub struct {
	usersByLogin     map[string]auth.StoredUser
	usersByID        map[int64]auth.StoredUser
	refreshSessions  map[string]auth.RefreshSession
	nextSessionID    int64
	seedRefreshToken string
}

func newHTTPAuthRepoStub(t *testing.T, now time.Time) *httpAuthRepoStub {
	t.Helper()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := auth.StoredUser{
		ID:           1,
		Login:        "operator",
		PasswordHash: string(passwordHash),
		IsActive:     true,
		Created:      now,
		Updated:      now,
		Role: auth.Role{
			ID:   2,
			Slug: "operator",
			Name: "Operator",
		},
	}

	refreshToken := "handler-refresh-token"
	session := auth.RefreshSession{
		ID:        10,
		UserID:    user.ID,
		TokenHash: auth.HashRefreshToken(refreshToken),
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
	}

	return &httpAuthRepoStub{
		usersByLogin: map[string]auth.StoredUser{user.Login: user},
		usersByID:    map[int64]auth.StoredUser{user.ID: user},
		refreshSessions: map[string]auth.RefreshSession{
			string(session.TokenHash): session,
		},
		nextSessionID:    11,
		seedRefreshToken: refreshToken,
	}
}

func (r *httpAuthRepoStub) GetUserByLogin(_ context.Context, login string) (auth.StoredUser, error) {
	user, ok := r.usersByLogin[login]
	if !ok {
		return auth.StoredUser{}, repository.ErrNotFound
	}
	return user, nil
}

func (r *httpAuthRepoStub) GetUserByID(_ context.Context, id int64) (auth.StoredUser, error) {
	user, ok := r.usersByID[id]
	if !ok {
		return auth.StoredUser{}, repository.ErrNotFound
	}
	return user, nil
}

func (r *httpAuthRepoStub) ListUsers(context.Context) ([]auth.StoredUser, error) {
	return nil, nil
}

func (r *httpAuthRepoStub) GetRoleBySlug(context.Context, string) (auth.Role, error) {
	return auth.Role{}, nil
}

func (r *httpAuthRepoStub) CreateUser(context.Context, auth.CreateUserParams) (auth.StoredUser, error) {
	return auth.StoredUser{}, nil
}

func (r *httpAuthRepoStub) UpdateUser(context.Context, auth.UpdateUserParams) (auth.StoredUser, error) {
	return auth.StoredUser{}, nil
}

func (r *httpAuthRepoStub) UpsertUser(context.Context, auth.UpsertUserParams) (auth.StoredUser, error) {
	return auth.StoredUser{}, nil
}

func (r *httpAuthRepoStub) CreateRefreshSession(_ context.Context, params auth.CreateRefreshSessionParams) (auth.RefreshSession, error) {
	session := auth.RefreshSession{
		ID:         r.nextSessionID,
		UserID:     params.UserID,
		TokenHash:  params.TokenHash,
		ExpiresAt:  params.ExpiresAt,
		CreatedAt:  params.LastUsedAt,
		LastUsedAt: &params.LastUsedAt,
	}
	r.refreshSessions[string(params.TokenHash)] = session
	r.nextSessionID++
	return session, nil
}

func (r *httpAuthRepoStub) GetRefreshSessionByHash(_ context.Context, tokenHash []byte) (auth.RefreshSession, error) {
	session, ok := r.refreshSessions[string(tokenHash)]
	if !ok {
		return auth.RefreshSession{}, repository.ErrNotFound
	}
	return session, nil
}

func (r *httpAuthRepoStub) RotateRefreshSession(_ context.Context, params auth.RotateRefreshSessionParams) (auth.RefreshSession, error) {
	current, ok := r.refreshSessions[string(params.CurrentTokenHash)]
	if !ok {
		return auth.RefreshSession{}, repository.ErrNotFound
	}
	next := auth.RefreshSession{
		ID:         r.nextSessionID,
		UserID:     current.UserID,
		TokenHash:  params.NewTokenHash,
		ExpiresAt:  params.ExpiresAt,
		CreatedAt:  params.Now,
		LastUsedAt: &params.Now,
	}
	replacedBy := next.ID
	current.RevokedAt = &params.Now
	current.ReplacedBySessionID = &replacedBy
	r.refreshSessions[string(params.CurrentTokenHash)] = current
	r.refreshSessions[string(params.NewTokenHash)] = next
	r.nextSessionID++
	return next, nil
}

func (r *httpAuthRepoStub) TouchRefreshSession(_ context.Context, sessionID int64, touchedAt time.Time) error {
	for key, session := range r.refreshSessions {
		if session.ID == sessionID {
			session.LastUsedAt = &touchedAt
			r.refreshSessions[key] = session
			return nil
		}
	}
	return repository.ErrNotFound
}

func (r *httpAuthRepoStub) RevokeRefreshSession(_ context.Context, params auth.RevokeRefreshSessionParams) error {
	for key, session := range r.refreshSessions {
		if session.ID == params.SessionID {
			session.RevokedAt = &params.RevokedAt
			r.refreshSessions[key] = session
			return nil
		}
	}
	return repository.ErrNotFound
}
