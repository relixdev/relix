package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/relixdev/relix/relay/internal/auth"
)

const testSecret = "super-secret-key"

func makeToken(t *testing.T, userID, role string, expiry time.Time) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  expiry.Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func TestValidAgentToken(t *testing.T) {
	token := makeToken(t, "user-123", "agent", time.Now().Add(time.Hour))

	claims, err := auth.ValidateToken(token, testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("want UserID=user-123, got %q", claims.UserID)
	}
	if claims.Role != "agent" {
		t.Errorf("want Role=agent, got %q", claims.Role)
	}
}

func TestValidMobileToken(t *testing.T) {
	token := makeToken(t, "user-456", "mobile", time.Now().Add(time.Hour))

	claims, err := auth.ValidateToken(token, testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Role != "mobile" {
		t.Errorf("want Role=mobile, got %q", claims.Role)
	}
}

func TestExpiredTokenFails(t *testing.T) {
	token := makeToken(t, "user-123", "agent", time.Now().Add(-time.Hour))

	_, err := auth.ValidateToken(token, testSecret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestWrongSecretFails(t *testing.T) {
	token := makeToken(t, "user-123", "agent", time.Now().Add(time.Hour))

	_, err := auth.ValidateToken(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestMissingRoleFails(t *testing.T) {
	claims := jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(time.Hour).Unix(),
		// "role" intentionally omitted
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tok.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	_, err = auth.ValidateToken(token, testSecret)
	if err == nil {
		t.Fatal("expected error for missing role, got nil")
	}
}

func TestInvalidRoleFails(t *testing.T) {
	token := makeToken(t, "user-123", "admin", time.Now().Add(time.Hour))

	_, err := auth.ValidateToken(token, testSecret)
	if err == nil {
		t.Fatal("expected error for invalid role 'admin', got nil")
	}
}
