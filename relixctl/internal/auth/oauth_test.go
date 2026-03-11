package auth_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/relixdev/relix/relixctl/internal/auth"
)

func TestBrowserLogin(t *testing.T) {
	const fakeCode = "testcode123"
	const fakeToken = "jwt.token.value"

	// Mock auth server handles both the authorize redirect and token exchange.
	mux := http.NewServeMux()

	// Token endpoint: exchange code for JWT.
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		if r.FormValue("code") != fakeCode {
			http.Error(w, "bad code", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"token": fakeToken})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	tokenURL := srv.URL + "/token"
	// authURL is irrelevant since we won't actually open a browser; we
	// just need BrowserLogin to start its local server and we'll POST the
	// callback ourselves.
	authURL := srv.URL + "/authorize"

	// Run BrowserLogin with a hook that simulates the browser redirect instead
	// of actually opening a browser. We use the testable variant that accepts
	// an openFunc.
	var callbackURL string
	captureOpen := func(u string) error {
		// Extract the redirect_uri from the auth URL and POST to it with a code.
		parsed, err := url.Parse(u)
		if err != nil {
			return err
		}
		redirectURI := parsed.Query().Get("redirect_uri")
		if redirectURI == "" {
			return fmt.Errorf("no redirect_uri in auth URL %q", u)
		}
		callbackURL = redirectURI + "?code=" + fakeCode

		// Fire the callback asynchronously so BrowserLogin can proceed.
		go func() {
			time.Sleep(20 * time.Millisecond)
			resp, err := http.Get(callbackURL) //nolint:noctx
			if err == nil {
				resp.Body.Close()
			}
		}()
		return nil
	}

	token, err := auth.BrowserLoginWithOpen("client-id", authURL, tokenURL, captureOpen)
	if err != nil {
		t.Fatalf("BrowserLogin returned error: %v", err)
	}
	if token != fakeToken {
		t.Errorf("got token %q, want %q", token, fakeToken)
	}
	if !strings.Contains(callbackURL, "localhost") {
		t.Errorf("redirect_uri should be localhost, got %q", callbackURL)
	}
}

func TestBrowserLogin_OpenError(t *testing.T) {
	failOpen := func(_ string) error {
		return fmt.Errorf("browser unavailable")
	}
	_, err := auth.BrowserLoginWithOpen("client-id", "http://example.com/auth", "http://example.com/token", failOpen)
	if err == nil {
		t.Fatal("expected error when openFunc fails")
	}
}
