package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"dimplom/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

func TestRefreshRotation(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	repo := newAuthRepoStub(now)
	tokens := &stubTokenManager{token: "access-2", expiresIn: 900}
	service := NewService(repo, tokens)
	service.now = func() time.Time { return now }
	service.refreshTTL = 24 * time.Hour

	result, err := service.Refresh(context.Background(), RefreshInput{
		RefreshToken: repo.seedRefreshToken,
		IP:           "127.0.0.1",
		UserAgent:    "test-agent",
	})
	if err != nil {
		t.Fatalf("refresh returned error: %v", err)
	}

	if result.AccessToken != "access-2" {
		t.Fatalf("expected access token to be re-issued, got %q", result.AccessToken)
	}
	if result.RefreshToken == "" || result.RefreshToken == repo.seedRefreshToken {
		t.Fatal("expected rotated refresh token to differ from the original")
	}

	previous := repo.refreshSessions[hashKey(HashRefreshToken(repo.seedRefreshToken))]
	if previous.RevokedAt == nil {
		t.Fatal("expected previous refresh session to be revoked")
	}
	if previous.ReplacedBySessionID == nil {
		t.Fatal("expected previous session to point at replacement session")
	}
	if len(repo.refreshSessions) != 2 {
		t.Fatalf("expected exactly two refresh sessions after rotation, got %d", len(repo.refreshSessions))
	}
}

func TestRefreshRejectsRevokedSession(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	repo := newAuthRepoStub(now)
	session := repo.refreshSessions[hashKey(HashRefreshToken(repo.seedRefreshToken))]
	session.RevokedAt = &now
	repo.refreshSessions[hashKey(HashRefreshToken(repo.seedRefreshToken))] = session

	service := NewService(repo, &stubTokenManager{token: "access-2", expiresIn: 900})
	service.now = func() time.Time { return now }

	_, err := service.Refresh(context.Background(), RefreshInput{RefreshToken: repo.seedRefreshToken})
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected ErrInvalidSession, got %v", err)
	}
}

func TestLoginRejectsInactiveUser(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	repo := newAuthRepoStub(now)
	user := repo.usersByLogin["operator"]
	user.IsActive = false
	repo.usersByLogin["operator"] = user
	repo.usersByID[user.ID] = user

	service := NewService(repo, &stubTokenManager{token: "access-1", expiresIn: 900})

	_, err := service.Login(context.Background(), LoginInput{
		Login:    "operator",
		Password: "secret",
	})
	if !errors.Is(err, ErrInactiveUser) {
		t.Fatalf("expected ErrInactiveUser, got %v", err)
	}
}

type stubTokenManager struct {
	token     string
	expiresIn int64
	parse     func(string) (Claims, error)
}

func (s *stubTokenManager) Issue(User) (string, int64, error) {
	return s.token, s.expiresIn, nil
}

func (s *stubTokenManager) Parse(token string) (Claims, error) {
	if s.parse != nil {
		return s.parse(token)
	}
	return Claims{}, ErrInvalidToken
}

type authRepoStub struct {
	usersByLogin     map[string]StoredUser
	usersByID        map[int64]StoredUser
	refreshSessions  map[string]RefreshSession
	nextSessionID    int64
	seedRefreshToken string
}

func newAuthRepoStub(now time.Time) *authRepoStub {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	user := StoredUser{
		ID:           1,
		Login:        "operator",
		PasswordHash: string(passwordHash),
		IsActive:     true,
		Created:      now,
		Updated:      now,
		Role: Role{
			ID:   2,
			Slug: "operator",
			Name: "Operator",
		},
	}

	seedToken := "seed-refresh-token"
	session := RefreshSession{
		ID:        10,
		UserID:    user.ID,
		TokenHash: HashRefreshToken(seedToken),
		ExpiresAt: now.Add(12 * time.Hour),
		CreatedAt: now,
	}

	return &authRepoStub{
		usersByLogin: map[string]StoredUser{user.Login: user},
		usersByID:    map[int64]StoredUser{user.ID: user},
		refreshSessions: map[string]RefreshSession{
			hashKey(session.TokenHash): session,
		},
		nextSessionID:    11,
		seedRefreshToken: seedToken,
	}
}

func (r *authRepoStub) GetUserByLogin(_ context.Context, login string) (StoredUser, error) {
	user, ok := r.usersByLogin[login]
	if !ok {
		return StoredUser{}, repository.ErrNotFound
	}
	return user, nil
}

func (r *authRepoStub) GetUserByID(_ context.Context, id int64) (StoredUser, error) {
	user, ok := r.usersByID[id]
	if !ok {
		return StoredUser{}, repository.ErrNotFound
	}
	return user, nil
}

func (r *authRepoStub) ListUsers(context.Context) ([]StoredUser, error) {
	return nil, nil
}

func (r *authRepoStub) GetRoleBySlug(context.Context, string) (Role, error) {
	return Role{}, nil
}

func (r *authRepoStub) CreateUser(context.Context, CreateUserParams) (StoredUser, error) {
	return StoredUser{}, nil
}

func (r *authRepoStub) UpdateUser(context.Context, UpdateUserParams) (StoredUser, error) {
	return StoredUser{}, nil
}

func (r *authRepoStub) UpsertUser(context.Context, UpsertUserParams) (StoredUser, error) {
	return StoredUser{}, nil
}

func (r *authRepoStub) CreateRefreshSession(_ context.Context, params CreateRefreshSessionParams) (RefreshSession, error) {
	session := RefreshSession{
		ID:          r.nextSessionID,
		UserID:      params.UserID,
		TokenHash:   params.TokenHash,
		ExpiresAt:   params.ExpiresAt,
		CreatedAt:   params.LastUsedAt,
		LastUsedAt:  &params.LastUsedAt,
		CreatedByIP: strPtr(params.CreatedByIP),
		UserAgent:   strPtr(params.UserAgent),
	}
	r.refreshSessions[hashKey(params.TokenHash)] = session
	r.nextSessionID++
	return session, nil
}

func (r *authRepoStub) GetRefreshSessionByHash(_ context.Context, tokenHash []byte) (RefreshSession, error) {
	session, ok := r.refreshSessions[hashKey(tokenHash)]
	if !ok {
		return RefreshSession{}, repository.ErrNotFound
	}
	return session, nil
}

func (r *authRepoStub) RotateRefreshSession(_ context.Context, params RotateRefreshSessionParams) (RefreshSession, error) {
	current, ok := r.refreshSessions[hashKey(params.CurrentTokenHash)]
	if !ok || current.ID != params.SessionID || current.RevokedAt != nil || params.Now.After(current.ExpiresAt) {
		return RefreshSession{}, repository.ErrNotFound
	}

	next := RefreshSession{
		ID:          r.nextSessionID,
		UserID:      current.UserID,
		TokenHash:   params.NewTokenHash,
		ExpiresAt:   params.ExpiresAt,
		CreatedAt:   params.Now,
		LastUsedAt:  &params.Now,
		CreatedByIP: strPtr(params.CreatedByIP),
		UserAgent:   strPtr(params.UserAgent),
	}
	replacedBy := next.ID
	current.RevokedAt = &params.Now
	current.ReplacedBySessionID = &replacedBy
	current.LastUsedAt = &params.Now
	r.refreshSessions[hashKey(params.CurrentTokenHash)] = current
	r.refreshSessions[hashKey(params.NewTokenHash)] = next
	r.nextSessionID++
	return next, nil
}

func (r *authRepoStub) TouchRefreshSession(_ context.Context, sessionID int64, touchedAt time.Time) error {
	for key, session := range r.refreshSessions {
		if session.ID == sessionID {
			session.LastUsedAt = &touchedAt
			r.refreshSessions[key] = session
			return nil
		}
	}
	return repository.ErrNotFound
}

func (r *authRepoStub) RevokeRefreshSession(_ context.Context, params RevokeRefreshSessionParams) error {
	for key, session := range r.refreshSessions {
		if session.ID == params.SessionID {
			session.RevokedAt = &params.RevokedAt
			r.refreshSessions[key] = session
			return nil
		}
	}
	return repository.ErrNotFound
}

func hashKey(hash []byte) string {
	return string(hash)
}

func strPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
