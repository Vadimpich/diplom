package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	ErrInactiveUser       = errors.New("auth: inactive user")
	ErrInvalidInput       = errors.New("auth: invalid input")
	ErrInvalidSession     = errors.New("auth: invalid session")
)

const defaultRefreshSessionTTL = 7 * 24 * time.Hour

type Role struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type User struct {
	ID       int64     `json:"id"`
	Login    string    `json:"login"`
	Role     Role      `json:"role"`
	IsActive bool      `json:"is_active"`
	Created  time.Time `json:"created_at"`
	Updated  time.Time `json:"updated_at"`
}

type LoginInput struct {
	Login     string
	Password  string
	IP        string
	UserAgent string
}

type LoginResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	User         User   `json:"user"`
}

type RefreshInput struct {
	RefreshToken string
	IP           string
	UserAgent    string
}

type LogoutInput struct {
	RefreshToken string
}

type CreateUserInput struct {
	Login    string
	Password string
	RoleSlug string
}

type UpdateUserInput struct {
	ID       int64
	Login    string
	RoleSlug string
	IsActive bool
}

type Repository interface {
	GetUserByLogin(context.Context, string) (StoredUser, error)
	GetUserByID(context.Context, int64) (StoredUser, error)
	ListUsers(context.Context) ([]StoredUser, error)
	GetRoleBySlug(context.Context, string) (Role, error)
	CreateUser(context.Context, CreateUserParams) (StoredUser, error)
	UpdateUser(context.Context, UpdateUserParams) (StoredUser, error)
	UpsertUser(context.Context, UpsertUserParams) (StoredUser, error)
	CreateRefreshSession(context.Context, CreateRefreshSessionParams) (RefreshSession, error)
	GetRefreshSessionByHash(context.Context, []byte) (RefreshSession, error)
	RotateRefreshSession(context.Context, RotateRefreshSessionParams) (RefreshSession, error)
	TouchRefreshSession(context.Context, int64, time.Time) error
	RevokeRefreshSession(context.Context, RevokeRefreshSessionParams) error
}

type TokenManager interface {
	Issue(User) (string, int64, error)
	Parse(string) (Claims, error)
}

type StoredUser struct {
	ID           int64
	Login        string
	PasswordHash string
	Role         Role
	IsActive     bool
	Created      time.Time
	Updated      time.Time
}

type UpsertUserParams struct {
	Login        string
	PasswordHash string
	RoleID       int64
}

type CreateUserParams struct {
	Login        string
	PasswordHash string
	RoleID       int64
}

type UpdateUserParams struct {
	ID       int64
	Login    string
	RoleID   int64
	IsActive bool
}

type BootstrapConfig struct {
	Login    string
	Password string
	RoleSlug string
}

type RefreshSession struct {
	ID                  int64
	UserID              int64
	TokenHash           []byte
	ExpiresAt           time.Time
	RevokedAt           *time.Time
	ReplacedBySessionID *int64
	CreatedByIP         *string
	UserAgent           *string
	LastUsedAt          *time.Time
	CreatedAt           time.Time
}

type CreateRefreshSessionParams struct {
	UserID      int64
	TokenHash   []byte
	ExpiresAt   time.Time
	CreatedByIP string
	UserAgent   string
	LastUsedAt  time.Time
}

type RotateRefreshSessionParams struct {
	SessionID        int64
	CurrentTokenHash []byte
	NewTokenHash     []byte
	ExpiresAt        time.Time
	CreatedByIP      string
	UserAgent        string
	Now              time.Time
}

type RevokeRefreshSessionParams struct {
	SessionID int64
	RevokedAt time.Time
}

type Service struct {
	repo       Repository
	tokens     TokenManager
	now        func() time.Time
	refreshTTL time.Duration
}

func NewService(repo Repository, tokens TokenManager) *Service {
	return &Service{
		repo:       repo,
		tokens:     tokens,
		now:        time.Now,
		refreshTTL: defaultRefreshSessionTTL,
	}
}

func (s *Service) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	user, err := s.repo.GetUserByLogin(ctx, input.Login)
	if err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	if !user.IsActive {
		return LoginResult{}, ErrInactiveUser
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	view := toUser(user)

	token, expiresIn, err := s.tokens.Issue(view)
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue token: %w", err)
	}

	refreshToken, tokenHash, err := NewRefreshToken()
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue refresh token: %w", err)
	}

	now := s.now()
	if _, err := s.repo.CreateRefreshSession(ctx, CreateRefreshSessionParams{
		UserID:      user.ID,
		TokenHash:   tokenHash,
		ExpiresAt:   now.Add(s.refreshTTL),
		CreatedByIP: strings.TrimSpace(input.IP),
		UserAgent:   strings.TrimSpace(input.UserAgent),
		LastUsedAt:  now,
	}); err != nil {
		return LoginResult{}, fmt.Errorf("create refresh session: %w", err)
	}

	return LoginResult{
		AccessToken:  token,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		User:         view,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, input RefreshInput) (LoginResult, error) {
	input.RefreshToken = strings.TrimSpace(input.RefreshToken)
	if input.RefreshToken == "" {
		return LoginResult{}, ErrInvalidInput
	}

	now := s.now()
	currentHash := HashRefreshToken(input.RefreshToken)
	session, err := s.repo.GetRefreshSessionByHash(ctx, currentHash)
	if err != nil {
		return LoginResult{}, ErrInvalidSession
	}
	if session.RevokedAt != nil || now.After(session.ExpiresAt) {
		return LoginResult{}, ErrInvalidSession
	}

	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return LoginResult{}, ErrInvalidSession
	}
	if !user.IsActive {
		return LoginResult{}, ErrInactiveUser
	}

	view := toUser(user)
	accessToken, expiresIn, err := s.tokens.Issue(view)
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue token: %w", err)
	}

	refreshToken, nextTokenHash, err := NewRefreshToken()
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue refresh token: %w", err)
	}

	if _, err := s.repo.RotateRefreshSession(ctx, RotateRefreshSessionParams{
		SessionID:        session.ID,
		CurrentTokenHash: currentHash,
		NewTokenHash:     nextTokenHash,
		ExpiresAt:        now.Add(s.refreshTTL),
		CreatedByIP:      strings.TrimSpace(input.IP),
		UserAgent:        strings.TrimSpace(input.UserAgent),
		Now:              now,
	}); err != nil {
		return LoginResult{}, ErrInvalidSession
	}

	return LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		User:         view,
	}, nil
}

func (s *Service) Logout(ctx context.Context, input LogoutInput) error {
	input.RefreshToken = strings.TrimSpace(input.RefreshToken)
	if input.RefreshToken == "" {
		return ErrInvalidInput
	}

	session, err := s.repo.GetRefreshSessionByHash(ctx, HashRefreshToken(input.RefreshToken))
	if err != nil {
		return nil
	}

	now := s.now()
	if err := s.repo.TouchRefreshSession(ctx, session.ID, now); err != nil {
		return err
	}
	if session.RevokedAt != nil {
		return nil
	}

	return s.repo.RevokeRefreshSession(ctx, RevokeRefreshSessionParams{
		SessionID: session.ID,
		RevokedAt: now,
	})
}

func (s *Service) Me(ctx context.Context, userID int64) (User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return User{}, err
	}

	if !user.IsActive {
		return User{}, ErrInactiveUser
	}

	return toUser(user), nil
}

func (s *Service) CreateUser(ctx context.Context, input CreateUserInput) (User, error) {
	input.Login = strings.TrimSpace(input.Login)
	input.Password = strings.TrimSpace(input.Password)
	input.RoleSlug = strings.TrimSpace(input.RoleSlug)
	if input.Login == "" || input.Password == "" {
		return User{}, ErrInvalidInput
	}
	if input.RoleSlug == "" {
		input.RoleSlug = "operator"
	}

	role, err := s.repo.GetRoleBySlug(ctx, input.RoleSlug)
	if err != nil {
		return User{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, CreateUserParams{
		Login:        input.Login,
		PasswordHash: string(passwordHash),
		RoleID:       role.ID,
	})
	if err != nil {
		return User{}, err
	}

	return toUser(user), nil
}

func (s *Service) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]User, 0, len(rows))
	for _, row := range rows {
		result = append(result, toUser(row))
	}

	return result, nil
}

func (s *Service) GetUserByID(ctx context.Context, id int64) (User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return User{}, err
	}

	return toUser(user), nil
}

func (s *Service) UpdateUser(ctx context.Context, input UpdateUserInput) (User, error) {
	input.Login = strings.TrimSpace(input.Login)
	input.RoleSlug = strings.TrimSpace(input.RoleSlug)
	if input.ID <= 0 || input.Login == "" || input.RoleSlug == "" {
		return User{}, ErrInvalidInput
	}

	role, err := s.repo.GetRoleBySlug(ctx, input.RoleSlug)
	if err != nil {
		return User{}, err
	}

	user, err := s.repo.UpdateUser(ctx, UpdateUserParams{
		ID:       input.ID,
		Login:    input.Login,
		RoleID:   role.ID,
		IsActive: input.IsActive,
	})
	if err != nil {
		return User{}, err
	}

	return toUser(user), nil
}

func (s *Service) EnsureInitialUser(ctx context.Context, cfg BootstrapConfig) error {
	if cfg.Login == "" || cfg.Password == "" {
		return nil
	}

	roleSlug := cfg.RoleSlug
	if roleSlug == "" {
		roleSlug = "operator"
	}

	role, err := s.repo.GetRoleBySlug(ctx, roleSlug)
	if err != nil {
		return fmt.Errorf("get role: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(cfg.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	_, err = s.repo.UpsertUser(ctx, UpsertUserParams{
		Login:        cfg.Login,
		PasswordHash: string(passwordHash),
		RoleID:       role.ID,
	})
	if err != nil {
		return fmt.Errorf("upsert initial user: %w", err)
	}

	return nil
}

func toUser(user StoredUser) User {
	return User{
		ID:       user.ID,
		Login:    user.Login,
		Role:     user.Role,
		IsActive: user.IsActive,
		Created:  user.Created,
		Updated:  user.Updated,
	}
}
