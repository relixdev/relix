package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// RoleAgent is the JWT role for relixctl agents.
	RoleAgent = "agent"
	// RoleMobile is the JWT role for mobile clients.
	RoleMobile = "mobile"

	tokenTTL = 24 * time.Hour
)

// Claims are the custom JWT claims issued by Relix Cloud.
type Claims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

// TokenService issues and validates JWTs.
type TokenService struct {
	secret []byte
}

// NewTokenService creates a TokenService with the given HMAC secret.
func NewTokenService(secret string) *TokenService {
	return &TokenService{secret: []byte(secret)}
}

// IssueToken creates a signed JWT for the given user ID and role.
// role should be RoleAgent or RoleMobile.
func (s *TokenService) IssueToken(userID, role string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
		Role: role,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("auth: sign token: %w", err)
	}
	return signed, nil
}

// ValidateToken parses and validates a JWT, returning its claims.
func (s *TokenService) ValidateToken(tokenStr string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("auth: unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("auth: parse token: %w", err)
	}
	claims, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, fmt.Errorf("auth: invalid token claims")
	}
	return claims, nil
}
