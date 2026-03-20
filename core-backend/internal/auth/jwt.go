package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("auth: invalid token")

type Claims struct {
	UserID   int64
	Login    string
	RoleSlug string
}

type JWTManager struct {
	issuer    string
	secret    []byte
	accessTTL time.Duration
	now       func() time.Time
}

type accessClaims struct {
	UserID   int64  `json:"uid"`
	Login    string `json:"login"`
	RoleSlug string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTManager(issuer, secret string, accessTTL time.Duration) *JWTManager {
	return &JWTManager{
		issuer:    issuer,
		secret:    []byte(secret),
		accessTTL: accessTTL,
		now:       time.Now,
	}
}

func (m *JWTManager) Issue(user User) (string, int64, error) {
	now := m.now()
	expiresAt := now.Add(m.accessTTL)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims{
		UserID:   user.ID,
		Login:    user.Login,
		RoleSlug: user.Role.Slug,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	})

	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", 0, err
	}

	return signed, int64(m.accessTTL.Seconds()), nil
}

func (m *JWTManager) Parse(token string) (Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &accessClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer))
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(*accessClaims)
	if !ok || !parsed.Valid {
		return Claims{}, ErrInvalidToken
	}

	return Claims{
		UserID:   claims.UserID,
		Login:    claims.Login,
		RoleSlug: claims.RoleSlug,
	}, nil
}

func NewRefreshToken() (string, []byte, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, fmt.Errorf("read random bytes: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	return token, sum[:], nil
}

func HashRefreshToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
