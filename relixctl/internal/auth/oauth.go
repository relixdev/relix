package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// BrowserLogin starts a local HTTP server, opens the browser to authURL, waits
// for the OAuth callback with an authorization code, exchanges it for a JWT
// token via tokenURL, and returns the token.
func BrowserLogin(clientID, authURL, tokenURL string) (string, error) {
	return BrowserLoginWithOpen(clientID, authURL, tokenURL, openBrowser)
}

// BrowserLoginWithOpen is the testable variant that accepts a custom function
// to open the browser URL.
func BrowserLoginWithOpen(clientID, authURL, tokenURL string, openFunc func(string) error) (string, error) {
	// Start local server on a random port.
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", fmt.Errorf("oauth: start callback server: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			errCh <- fmt.Errorf("oauth: callback missing code")
			return
		}
		fmt.Fprintln(w, "Authentication successful. You may close this tab.")
		codeCh <- code
	})

	srv := &http.Server{Handler: mux}
	go func() {
		_ = srv.Serve(listener)
	}()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	// Build the auth URL with redirect_uri and client_id.
	authRedirect, err := buildAuthURL(authURL, clientID, redirectURI)
	if err != nil {
		return "", fmt.Errorf("oauth: build auth URL: %w", err)
	}

	// Open browser (or call the provided open function).
	if err := openFunc(authRedirect); err != nil {
		return "", fmt.Errorf("oauth: open browser: %w", err)
	}

	// Wait for callback or error.
	var code string
	select {
	case code = <-codeCh:
	case err = <-errCh:
		return "", err
	case <-time.After(5 * time.Minute):
		return "", fmt.Errorf("oauth: timed out waiting for callback")
	}

	// Exchange code for token.
	return exchangeCode(tokenURL, clientID, code, redirectURI)
}

func buildAuthURL(authURL, clientID, redirectURI string) (string, error) {
	u, err := url.Parse(authURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func exchangeCode(tokenURL, clientID, code, redirectURI string) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", clientID)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode())) //nolint:noctx
	if err != nil {
		return "", fmt.Errorf("oauth: token exchange: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("oauth: token exchange returned status %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("oauth: decode token response: %w", err)
	}
	if result.Token == "" {
		return "", fmt.Errorf("oauth: empty token in response")
	}
	return result.Token, nil
}

// openBrowser opens the given URL in the default browser.
func openBrowser(u string) error {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	return exec.Command(cmd, u).Start()
}
