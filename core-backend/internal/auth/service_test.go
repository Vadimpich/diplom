package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"diplom/internal/audit"
	"diplom/internal/repository"
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

func TestLoginWritesAuditEvent(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	repo := newAuthRepoStub(now)
	auditRepo := &authAuditRepoStub{}
	service := NewService(repo, &stubTokenManager{token: "access-1", expiresIn: 900}, audit.NewService(auditRepo))
	service.now = func() time.Time { return now }

	if _, err := service.Login(context.Background(), LoginInput{
		Login:     "operator",
		Password:  "secret",
		IP:        "127.0.0.1",
		UserAgent: "test-agent",
	}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}
	if len(auditRepo.events) != 1 {
		t.Fatalf("expected one audit event, got %d", len(auditRepo.events))
	}
	if auditRepo.events[0].Type != audit.EventTypeAuthLogin {
		t.Fatalf("expected auth.login event, got %q", auditRepo.events[0].Type)
	}
	loggedInUser := repo.usersByID[1]
	if loggedInUser.LastLoginAt == nil || !loggedInUser.LastLoginAt.Equal(now) {
		t.Fatalf("expected last login to be updated to %s, got %#v", now.Format(time.RFC3339), loggedInUser.LastLoginAt)
	}
}

func TestFailedLoginWritesAuditEvent(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	repo := newAuthRepoStub(now)
	auditRepo := &authAuditRepoStub{}
	service := NewService(repo, &stubTokenManager{token: "access-1", expiresIn: 900}, audit.NewService(auditRepo))

	_, err := service.Login(context.Background(), LoginInput{
		Login:     "operator",
		Password:  "wrong",
		IP:        "127.0.0.1",
		UserAgent: "test-agent",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	if len(auditRepo.events) != 1 {
		t.Fatalf("expected one audit event, got %d", len(auditRepo.events))
	}
	if auditRepo.events[0].Type != audit.EventTypeAuthLoginFailed {
		t.Fatalf("expected auth.login_failed event, got %q", auditRepo.events[0].Type)
	}
}

func TestGetUserByIDDoesNotWriteUpdateAuditEvent(t *testing.T) {
	now := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	repo := newAuthRepoStub(now)
	auditRepo := &authAuditRepoStub{}
	service := NewService(repo, &stubTokenManager{}, audit.NewService(auditRepo))

	user, err := service.GetUserByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("get user returned error: %v", err)
	}

	if user.ID != 1 {
		t.Fatalf("expected user ID 1, got %d", user.ID)
	}
	if len(auditRepo.events) != 0 {
		t.Fatalf("expected no audit events for read path, got %d", len(auditRepo.events))
	}
}

func TestCreateUserAndLoginAuditBehaviorRemainsIntact(t *testing.T) {
	now := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	repo := newAuthRepoStub(now)
	auditRepo := &authAuditRepoStub{}
	service := NewService(repo, &stubTokenManager{token: "access-1", expiresIn: 900}, audit.NewService(auditRepo))

	createdUser, err := service.CreateUser(context.Background(), CreateUserInput{
		Login:    "admin-2",
		Password: "secret-2",
		RoleSlug: "admin",
	})
	if err != nil {
		t.Fatalf("create user returned error: %v", err)
	}

	if createdUser.Login != "admin-2" {
		t.Fatalf("expected created login admin-2, got %q", createdUser.Login)
	}

	if _, err := service.Login(context.Background(), LoginInput{
		Login:     "operator",
		Password:  "secret",
		IP:        "127.0.0.1",
		UserAgent: "test-agent",
	}); err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	if len(auditRepo.events) != 2 {
		t.Fatalf("expected two audit events, got %d", len(auditRepo.events))
	}
	if auditRepo.events[0].Type != audit.EventTypeAdminUserCreated {
		t.Fatalf("expected first event admin.user_created, got %q", auditRepo.events[0].Type)
	}
	if auditRepo.events[1].Type != audit.EventTypeAuthLogin {
		t.Fatalf("expected second event auth.login, got %q", auditRepo.events[1].Type)
	}
}

func TestUpdateUserIsOnlyPathThatWritesUserUpdatedAuditEvent(t *testing.T) {
	now := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	repo := newAuthRepoStub(now)
	auditRepo := &authAuditRepoStub{}
	service := NewService(repo, &stubTokenManager{}, audit.NewService(auditRepo))

	updatedUser, err := service.UpdateUser(context.Background(), UpdateUserInput{
		ID:       1,
		Login:    "operator-2",
		RoleSlug: "admin",
		IsActive: false,
	})
	if err != nil {
		t.Fatalf("update user returned error: %v", err)
	}

	if updatedUser.Login != "operator-2" {
		t.Fatalf("expected updated login operator-2, got %q", updatedUser.Login)
	}
	if len(auditRepo.events) != 1 {
		t.Fatalf("expected one audit event, got %d", len(auditRepo.events))
	}
	if auditRepo.events[0].Type != audit.EventTypeAdminUserUpdated {
		t.Fatalf("expected admin.user_updated event, got %q", auditRepo.events[0].Type)
	}
}

func TestListUsersIncludesLastLoginTimestamp(t *testing.T) {
	now := time.Date(2026, 3, 25, 8, 0, 0, 0, time.UTC)
	repo := newAuthRepoStub(now)
	lastLogin := now.Add(-2 * time.Hour)
	user := repo.usersByID[1]
	user.LastLoginAt = &lastLogin
	repo.usersByID[user.ID] = user
	repo.usersByLogin[user.Login] = user

	service := NewService(repo, &stubTokenManager{})
	items, err := service.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("list users returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one user, got %d", len(items))
	}
	if items[0].LastLoginAt == nil || !items[0].LastLoginAt.Equal(lastLogin) {
		t.Fatalf("expected last login %s, got %#v", lastLogin.Format(time.RFC3339), items[0].LastLoginAt)
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
	rolesBySlug      map[string]Role
	refreshSessions  map[string]RefreshSession
	nextSessionID    int64
	nextUserID       int64
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
		rolesBySlug: map[string]Role{
			"admin": {
				ID:   1,
				Slug: "admin",
				Name: "Admin",
			},
			"operator": user.Role,
		},
		refreshSessions: map[string]RefreshSession{
			hashKey(session.TokenHash): session,
		},
		nextSessionID:    11,
		nextUserID:       2,
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
	items := make([]StoredUser, 0, len(r.usersByID))
	for _, user := range r.usersByID {
		items = append(items, user)
	}
	return items, nil
}

func (r *authRepoStub) GetRoleBySlug(_ context.Context, slug string) (Role, error) {
	role, ok := r.rolesBySlug[slug]
	if !ok {
		return Role{}, repository.ErrNotFound
	}
	return role, nil
}

func (r *authRepoStub) CreateUser(_ context.Context, params CreateUserParams) (StoredUser, error) {
	role, ok := roleByID(r.rolesBySlug, params.RoleID)
	if !ok {
		return StoredUser{}, repository.ErrNotFound
	}

	user := StoredUser{
		ID:           r.nextUserID,
		Login:        params.Login,
		PasswordHash: params.PasswordHash,
		Role:         role,
		IsActive:     true,
		Created:      time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC),
		Updated:      time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC),
	}
	r.usersByID[user.ID] = user
	r.usersByLogin[user.Login] = user
	r.nextUserID++
	return user, nil
}

func (r *authRepoStub) UpdateUser(_ context.Context, params UpdateUserParams) (StoredUser, error) {
	user, ok := r.usersByID[params.ID]
	if !ok {
		return StoredUser{}, repository.ErrNotFound
	}

	role, ok := roleByID(r.rolesBySlug, params.RoleID)
	if !ok {
		return StoredUser{}, repository.ErrNotFound
	}

	delete(r.usersByLogin, user.Login)
	user.Login = params.Login
	user.Role = role
	user.IsActive = params.IsActive
	user.Updated = time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	r.usersByID[user.ID] = user
	r.usersByLogin[user.Login] = user
	return user, nil
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

func (r *authRepoStub) UpdateUserLastLogin(_ context.Context, userID int64, loggedAt time.Time) error {
	user, ok := r.usersByID[userID]
	if !ok {
		return repository.ErrNotFound
	}
	user.LastLoginAt = &loggedAt
	r.usersByID[userID] = user
	r.usersByLogin[user.Login] = user
	return nil
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

func roleByID(roles map[string]Role, id int64) (Role, bool) {
	for _, role := range roles {
		if role.ID == id {
			return role, true
		}
	}
	return Role{}, false
}

type authAuditRepoStub struct {
	events []audit.Event
}

func (s *authAuditRepoStub) Append(_ context.Context, event audit.Event) error {
	s.events = append(s.events, event)
	return nil
}

func (s *authAuditRepoStub) List(context.Context, audit.ListFilter) ([]audit.Event, error) {
	return append([]audit.Event(nil), s.events...), nil
}
