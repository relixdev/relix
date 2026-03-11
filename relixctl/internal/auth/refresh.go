package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const refreshWindow = 5 * time.Minute

// TokenRefresher holds a JWT token and refreshes it when it is close to expiry.
type TokenRefresher struct {
	mu         sync.Mutex
	token      string
	refreshURL string
	expiry     time.Time
}

// NewTokenRefresher creates a TokenRefresher with the given initial token,
// refresh URL, and token expiry time.
func NewTokenRefresher(token, refreshURL string, expiry time.Time) *TokenRefresher {
	return &TokenRefresher{
		token:      token,
		refreshURL: refreshURL,
		expiry:     expiry,
	}
}

// Token returns the current token.
func (tr *TokenRefresher) Token() string {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	return tr.token
}

// RefreshIfNeeded returns the current token. If the token expires within
// refreshWindow (5 minutes), it first POSTs to the refresh URL to obtain a new
// token and updates the internal state.
func (tr *TokenRefresher) RefreshIfNeeded() (string, error) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if time.Until(tr.expiry) > refreshWindow {
		return tr.token, nil
	}

	newToken, err := tr.doRefresh()
	if err != nil {
		return "", err
	}
	tr.token = newToken
	// Advance expiry — server should return a new expiry; for now assume 1 hour.
	tr.expiry = time.Now().Add(1 * time.Hour)
	return tr.token, nil
}

// StartAutoRefresh runs a goroutine that periodically calls RefreshIfNeeded.
// The goroutine stops when stopCh is closed. The provided onUpdate callback
// is called with the new token after each successful refresh.
func (tr *TokenRefresher) StartAutoRefresh(stopCh <-chan struct{}, onUpdate func(string)) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				token, err := tr.RefreshIfNeeded()
				if err == nil && onUpdate != nil {
					onUpdate(token)
				}
			}
		}
	}()
}

func (tr *TokenRefresher) doRefresh() (string, error) {
	body := strings.NewReader(`{"grant_type":"refresh_token"}`)
	req, err := http.NewRequest(http.MethodPost, tr.refreshURL, body)
	if err != nil {
		return "", fmt.Errorf("token refresh: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tr.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token refresh: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token refresh: server returned %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("token refresh: decode response: %w", err)
	}
	if result.Token == "" {
		return "", fmt.Errorf("token refresh: empty token in response")
	}
	return result.Token, nil
}
