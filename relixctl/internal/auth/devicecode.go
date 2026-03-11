package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// deviceCodeResponse is the JSON response from the device code endpoint.
type deviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	Interval        int    `json:"interval"`
}

// DeviceCodeLogin performs the OAuth 2.0 device code flow.
// It POSTs to deviceCodeURL to obtain a device code, prints the user code and
// verification URI, then polls tokenURL every interval seconds until the user
// authorises the device or an error occurs.
func DeviceCodeLogin(clientID, deviceCodeURL, tokenURL string) (string, error) {
	// Step 1: request device and user codes.
	form := url.Values{}
	form.Set("client_id", clientID)

	resp, err := http.Post(deviceCodeURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode())) //nolint:noctx
	if err != nil {
		return "", fmt.Errorf("device code: request codes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("device code: request returned status %d", resp.StatusCode)
	}

	var dcr deviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&dcr); err != nil {
		return "", fmt.Errorf("device code: decode response: %w", err)
	}

	// Step 2: instruct the user.
	fmt.Printf("Go to %s and enter code: %s\n", dcr.VerificationURI, dcr.UserCode)

	// Step 3: poll for token.
	interval := dcr.Interval
	if interval <= 0 {
		interval = 5
	}
	pollInterval := time.Duration(interval) * time.Second

	for {
		time.Sleep(pollInterval)

		token, pending, err := pollForToken(tokenURL, clientID, dcr.DeviceCode)
		if err != nil {
			return "", err
		}
		if pending {
			continue
		}
		return token, nil
	}
}

// pollForToken polls the token endpoint once. Returns (token, false, nil) on
// success, ("", true, nil) when authorization is pending, or ("", false, err)
// on a non-recoverable error.
func pollForToken(tokenURL, clientID, deviceCode string) (string, bool, error) {
	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	form.Set("client_id", clientID)
	form.Set("device_code", deviceCode)

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode())) //nolint:noctx
	if err != nil {
		return "", false, fmt.Errorf("device code: poll token: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Token string `json:"token"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", false, fmt.Errorf("device code: decode poll response: %w", err)
	}

	if result.Error == "authorization_pending" || result.Error == "slow_down" {
		return "", true, nil
	}
	if result.Error != "" {
		return "", false, fmt.Errorf("device code: token error: %s", result.Error)
	}
	if result.Token == "" {
		return "", false, fmt.Errorf("device code: empty token in response")
	}
	return result.Token, false, nil
}
