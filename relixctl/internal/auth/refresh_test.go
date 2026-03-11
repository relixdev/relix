package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/relixdev/relix/relixctl/internal/auth"
)

func TestTokenRefresher_RefreshIfNeeded_NearExpiry(t *testing.T) {
	const newToken = "refreshed.jwt.token"
	refreshCalled := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		refreshCalled = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"token": newToken})
	}))
	defer srv.Close()

	// Token expires in 2 minutes — within the 5-minute refresh window.
	expiry := time.Now().Add(2 * time.Minute)
	tr := auth.NewTokenRefresher("old.jwt.token", srv.URL, expiry)

	token, err := tr.RefreshIfNeeded()
	if err != nil {
		t.Fatalf("RefreshIfNeeded returned error: %v", err)
	}
	if !refreshCalled {
		t.Error("expected refresh to be called for near-expiry token")
	}
	if token != newToken {
		t.Errorf("got token %q, want %q", token, newToken)
	}
}

func TestTokenRefresher_RefreshIfNeeded_FarFromExpiry(t *testing.T) {
	refreshCalled := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refreshCalled = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "new.token"})
	}))
	defer srv.Close()

	// Token expires in 1 hour — well outside the 5-minute refresh window.
	expiry := time.Now().Add(1 * time.Hour)
	tr := auth.NewTokenRefresher("current.jwt.token", srv.URL, expiry)

	token, err := tr.RefreshIfNeeded()
	if err != nil {
		t.Fatalf("RefreshIfNeeded returned error: %v", err)
	}
	if refreshCalled {
		t.Error("expected no refresh for token far from expiry")
	}
	if token != "current.jwt.token" {
		t.Errorf("got token %q, want %q", token, "current.jwt.token")
	}
}

func TestTokenRefresher_RefreshIfNeeded_Expired(t *testing.T) {
	const newToken = "refreshed.expired.token"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"token": newToken})
	}))
	defer srv.Close()

	// Token already expired.
	expiry := time.Now().Add(-1 * time.Minute)
	tr := auth.NewTokenRefresher("expired.jwt.token", srv.URL, expiry)

	token, err := tr.RefreshIfNeeded()
	if err != nil {
		t.Fatalf("RefreshIfNeeded returned error: %v", err)
	}
	if token != newToken {
		t.Errorf("got token %q, want %q", token, newToken)
	}
}
