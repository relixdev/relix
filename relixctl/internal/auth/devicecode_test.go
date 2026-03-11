package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/relixdev/relix/relixctl/internal/auth"
)

func TestDeviceCodeLogin(t *testing.T) {
	const fakeDeviceCode = "device_code_abc"
	const fakeUserCode = "USER-CODE"
	const fakeVerificationURI = "https://relix.sh/activate"
	const fakeToken = "jwt.device.token"
	pollCount := 0

	mux := http.NewServeMux()

	// Device code endpoint.
	mux.HandleFunc("/device", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"device_code":      fakeDeviceCode,
			"user_code":        fakeUserCode,
			"verification_uri": fakeVerificationURI,
			"interval":         0, // 0 seconds so tests run fast
		})
	})

	// Token poll endpoint — succeed on second poll.
	mux.HandleFunc("/token_poll", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		pollCount++
		if pollCount < 2 {
			// Return authorization_pending on first poll.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "authorization_pending"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"token": fakeToken})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	token, err := auth.DeviceCodeLogin("client-id", srv.URL+"/device", srv.URL+"/token_poll")
	if err != nil {
		t.Fatalf("DeviceCodeLogin returned error: %v", err)
	}
	if token != fakeToken {
		t.Errorf("got token %q, want %q", token, fakeToken)
	}
	if pollCount < 2 {
		t.Errorf("expected at least 2 polls, got %d", pollCount)
	}
}

func TestDeviceCodeLogin_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := auth.DeviceCodeLogin("client-id", srv.URL+"/device", srv.URL+"/token")
	if err == nil {
		t.Fatal("expected error on server failure")
	}
}
