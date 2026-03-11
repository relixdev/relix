package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/relixdev/relix/relay/internal/hub"
	"github.com/relixdev/relix/relay/internal/server"
)

func newPairTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	h := hub.New()
	srv := server.New(testSecret, h)
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)
	return ts
}

func pairRequest(t *testing.T, ts *httptest.Server, code string) *http.Response {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"code": code})
	resp, err := http.Post(ts.URL+"/pair", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /pair: %v", err)
	}
	return resp
}

func TestPairEndpointExists(t *testing.T) {
	ts := newPairTestServer(t)

	// Send a request with no code — expect 400 (Bad Request), not 404 from the mux.
	body, _ := json.Marshal(map[string]string{})
	resp, err := http.Post(ts.URL+"/pair", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /pair: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		t.Errorf("/pair endpoint not registered (got 404)")
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("/pair with empty code: want 400, got %d", resp.StatusCode)
	}
}

func TestPairEndpointRejectsGet(t *testing.T) {
	ts := newPairTestServer(t)

	resp, err := http.Get(ts.URL + "/pair")
	if err != nil {
		t.Fatalf("GET /pair: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("GET /pair: want 405, got %d", resp.StatusCode)
	}
}

func TestPairEndpointRateLimitsRapidRequests(t *testing.T) {
	h := hub.New()
	// Use a rate limiter with maxAttempts=3 so we don't need 6 requests.
	srv := server.NewWithOptions(testSecret, h, server.Options{
		PairRateMaxAttempts: 3,
		PairRateWindow:      3 * time.Second,
	})
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)

	// Register a pairing code so requests don't 404 before hitting the rate limit.
	h.Pairing().RegisterPairing("RL-CODE", "user1")

	// First 3 requests should not be rate limited.
	// Request 1 consumes the code (404 on subsequent ones, but not 429).
	for i := 0; i < 3; i++ {
		body, _ := json.Marshal(map[string]string{"code": "RL-CODE"})
		resp, err := http.Post(ts.URL+"/pair", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("request %d: %v", i+1, err)
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusTooManyRequests {
			t.Errorf("request %d: got 429 before limit reached", i+1)
		}
	}

	// 4th request should be rate limited (429), not 404.
	body, _ := json.Marshal(map[string]string{"code": "RL-CODE"})
	resp, err := http.Post(ts.URL+"/pair", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("4th request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("4th request: want 429 (rate limited), got %d", resp.StatusCode)
	}
}

func TestPairEndpointRequiresCode(t *testing.T) {
	ts := newPairTestServer(t)

	// Send empty body.
	resp, err := http.Post(ts.URL+"/pair", "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		t.Fatalf("POST /pair: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("empty code: want 400, got %d", resp.StatusCode)
	}
}

func TestPairEndpointUnknownCodeReturns404(t *testing.T) {
	ts := newPairTestServer(t)

	resp := pairRequest(t, ts, "NONEXISTENT")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("unknown code: want 404, got %d", resp.StatusCode)
	}
}

func TestPairEndpointValidCodeReturns200(t *testing.T) {
	h := hub.New()
	// Register a pairing code directly on the hub's pairing store.
	h.Pairing().RegisterPairing("VALID-CODE", "user42")

	srv := server.New(testSecret, h)
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)

	resp := pairRequest(t, ts, "VALID-CODE")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("valid code: want 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result["user_id"] != "user42" {
		t.Errorf("want user_id=user42, got %q", result["user_id"])
	}
}
