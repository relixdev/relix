package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/relixdev/relix/relay/internal/server"
)

func TestHealthEndpoint(t *testing.T) {
	srv := server.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("want status=ok, got %q", body["status"])
	}
}

func TestWSEndpointUpgrades(t *testing.T) {
	srv := server.New()
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// A plain HTTP GET to /ws should return 400 (bad WebSocket handshake),
	// proving the route exists and attempts an upgrade.
	resp, err := http.Get(ts.URL + "/ws")
	if err != nil {
		t.Fatalf("GET /ws: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		t.Errorf("/ws route not registered (got 404)")
	}
}
