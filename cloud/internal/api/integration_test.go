package api_test

import (
	"bytes"
	"context"
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

// TestIntegration_RegisterMachines_HitLimit_Upgrade exercises the full flow:
// register user → add machines → hit free limit → upgrade tier → add more.
func TestIntegration_RegisterMachines_HitLimit_Upgrade(t *testing.T) {
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

	// Step 1: Register a user via email.
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, postJSON("/auth/email/register",
		`{"email":"integration@example.com","password":"securepass123"}`))
	if w.Code != http.StatusCreated {
		t.Fatalf("register user = %d: %s", w.Code, w.Body)
	}

	var regResp struct {
		Token string     `json:"token"`
		User  *user.User `json:"user"`
	}
	json.NewDecoder(w.Body).Decode(&regResp)
	tok := regResp.Token
	userID := regResp.User.ID

	// Step 2: Add 3 machines (free tier limit).
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]string{
			"name":       "machine",
			"public_key": "pk",
		})
		req := httptest.NewRequest(http.MethodPost, "/machines", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		srv.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("register machine %d = %d: %s", i, w.Code, w.Body)
		}
	}

	// Step 3: 4th machine must fail (free limit is 3).
	{
		w := httptest.NewRecorder()
		body := `{"name":"extra","public_key":"pk"}`
		req := httptest.NewRequest(http.MethodPost, "/machines", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		srv.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Errorf("4th machine on free tier = %d, want 403", w.Code)
		}
	}

	// Step 4: Upgrade user to Plus (simulated — in production this is a Stripe webhook).
	ctx := context.Background()
	u, _ := store.GetUserByID(ctx, userID)
	u.Tier = user.TierPlus
	store.UpdateUser(ctx, u)

	// Step 5: After upgrade, adding more machines should succeed.
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		body := `{"name":"server","public_key":"pk"}`
		req := httptest.NewRequest(http.MethodPost, "/machines", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		srv.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Errorf("machine after upgrade %d = %d: %s", i, w.Code, w.Body)
		}
	}

	// Step 6: Verify billing plan reflects Plus tier.
	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/billing/plan", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		srv.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("get plan = %d: %s", w.Code, w.Body)
		}
		var plan map[string]any
		json.NewDecoder(w.Body).Decode(&plan)
		if plan["tier"] != string(user.TierPlus) {
			t.Errorf("plan tier = %v, want plus", plan["tier"])
		}
	}

	// Step 7: List machines shows all 6.
	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/machines", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		srv.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("list machines = %d: %s", w.Code, w.Body)
		}
		var resp map[string]any
		json.NewDecoder(w.Body).Decode(&resp)
		machines, _ := resp["machines"].([]any)
		if len(machines) != 6 {
			t.Errorf("machine count = %d, want 6", len(machines))
		}
	}
}
