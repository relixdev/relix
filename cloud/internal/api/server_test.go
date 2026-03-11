package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/relixdev/relix/cloud/internal/api"
	"github.com/relixdev/relix/cloud/internal/auth"
	"github.com/relixdev/relix/cloud/internal/billing"
	"github.com/relixdev/relix/cloud/internal/machine"
	"github.com/relixdev/relix/cloud/internal/push"
	"github.com/relixdev/relix/cloud/internal/user"
)

const testSecret = "test-secret-minimum-32-chars-long!!"

func newTestServer(t *testing.T) (*api.Server, *user.MemoryStore, *auth.TokenService) {
	t.Helper()
	store := user.NewMemoryStore()
	tokens := auth.NewTokenService(testSecret)
	reg := machine.NewRegistry(store)
	srv := api.New(api.Config{
		Tokens:      tokens,
		EmailAuth:   auth.NewEmailAuth(store),
		GitHubOAuth: auth.NewGitHubOAuth("", ""),
		UserStore:   store,
		Registry:    reg,
		Stripe:      billing.NewStubStripe(),
		Push:        push.NewAPNs(),
	})
	return srv, store, tokens
}

func TestHealthCheck(t *testing.T) {
	srv, _, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("health = %d, want 200", w.Code)
	}
}

func TestEmailRegisterAndLogin(t *testing.T) {
	srv, _, tokens := newTestServer(t)

	body := `{"email":"alice@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/email/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("register = %d, want 201: %s", w.Code, w.Body)
	}

	var resp struct {
		Token string     `json:"token"`
		User  *user.User `json:"user"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Token == "" {
		t.Error("expected token in response")
	}

	// Validate the token
	claims, err := tokens.ValidateToken(resp.Token)
	if err != nil {
		t.Fatalf("invalid token: %v", err)
	}
	if claims.Subject != resp.User.ID {
		t.Errorf("token sub = %q, want %q", claims.Subject, resp.User.ID)
	}

	// Login
	req2 := httptest.NewRequest(http.MethodPost, "/auth/email/login", bytes.NewBufferString(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("login = %d, want 200: %s", w2.Code, w2.Body)
	}
}

func TestEmailLogin_WrongPassword(t *testing.T) {
	srv, _, _ := newTestServer(t)

	srv.ServeHTTP(httptest.NewRecorder(),
		postJSON("/auth/email/register", `{"email":"b@example.com","password":"correctpass"}`))

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, postJSON("/auth/email/login", `{"email":"b@example.com","password":"wrongpass"}`))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("wrong password = %d, want 401", w.Code)
	}
}

func TestMachineRoutes(t *testing.T) {
	srv, store, tokens := newTestServer(t)
	ctx := httptest.NewRequest(http.MethodGet, "/", nil).Context()

	u, _ := store.CreateUser(ctx, &user.User{Email: "c@example.com", Tier: user.TierFree})
	tok, _ := tokens.IssueToken(u.ID, auth.RoleMobile)

	// List empty
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/machines", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list machines = %d: %s", w.Code, w.Body)
	}

	// Register machine
	w2 := httptest.NewRecorder()
	req2 := postJSON("/machines", `{"name":"laptop","public_key":"pk123"}`)
	req2.Header.Set("Authorization", "Bearer "+tok)
	srv.ServeHTTP(w2, req2)
	if w2.Code != http.StatusCreated {
		t.Fatalf("register machine = %d: %s", w2.Code, w2.Body)
	}

	var m user.Machine
	json.NewDecoder(w2.Body).Decode(&m)
	if m.ID == "" {
		t.Error("expected machine ID")
	}

	// Delete machine
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodDelete, "/machines/"+m.ID, nil)
	req3.Header.Set("Authorization", "Bearer "+tok)
	srv.ServeHTTP(w3, req3)
	if w3.Code != http.StatusNoContent {
		t.Errorf("delete machine = %d: %s", w3.Code, w3.Body)
	}
}

func TestMachineRoutes_Unauthorized(t *testing.T) {
	srv, _, _ := newTestServer(t)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/machines", nil))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("no auth = %d, want 401", w.Code)
	}
}

func TestBillingGetPlan(t *testing.T) {
	srv, store, tokens := newTestServer(t)
	ctx := httptest.NewRequest(http.MethodGet, "/", nil).Context()

	u, _ := store.CreateUser(ctx, &user.User{Email: "d@example.com", Tier: user.TierPlus})
	tok, _ := tokens.IssueToken(u.ID, auth.RoleMobile)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/billing/plan", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("get plan = %d: %s", w.Code, w.Body)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["tier"] != string(user.TierPlus) {
		t.Errorf("tier = %v, want plus", resp["tier"])
	}
}

func postJSON(path, body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}
