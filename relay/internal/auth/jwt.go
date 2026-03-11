package auth

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// Role represents an authenticated client's role.
type Role string

const (
	RoleAgent  Role = "agent"
	RoleMobile Role = "mobile"
)

// Claims holds the validated fields extracted from a JWT.
type Claims struct {
	UserID string
	Role   Role
}

// ValidateToken parses and validates a signed JWT string against secret.
// It returns Claims on success, or an error if the token is invalid, expired,
// signed with the wrong secret, missing a role, or has an unrecognised role.
func ValidateToken(tokenString, secret string) (Claims, error) {
	tok, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return Claims{}, err
	}

	mapClaims, ok := tok.Claims.(jwt.MapClaims)
	if !ok || !tok.Valid {
		return Claims{}, errors.New("invalid token claims")
	}

	userID, _ := mapClaims["sub"].(string)

	roleStr, ok := mapClaims["role"].(string)
	if !ok || roleStr == "" {
		return Claims{}, errors.New("token missing role claim")
	}

	role := Role(roleStr)
	if role != RoleAgent && role != RoleMobile {
		return Claims{}, fmt.Errorf("invalid role %q: must be agent or mobile", roleStr)
	}

	return Claims{UserID: userID, Role: role}, nil
}
