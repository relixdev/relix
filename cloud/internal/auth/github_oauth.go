package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// GitHubUser holds the fields we care about from the GitHub user API.
type GitHubUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

// GitHubOAuth handles the GitHub OAuth2 code-exchange flow.
type GitHubOAuth struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

// NewGitHubOAuth creates a GitHubOAuth provider with the given credentials.
func NewGitHubOAuth(clientID, clientSecret string) *GitHubOAuth {
	return &GitHubOAuth{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{},
	}
}

// ExchangeCode trades a GitHub OAuth authorization code for an access token,
// then fetches and returns the authenticated GitHub user profile.
func (g *GitHubOAuth) ExchangeCode(ctx context.Context, code string) (*GitHubUser, error) {
	accessToken, err := g.exchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return g.fetchUser(ctx, accessToken)
}

func (g *GitHubOAuth) exchangeCode(ctx context.Context, code string) (string, error) {
	body := url.Values{
		"client_id":     {g.clientID},
		"client_secret": {g.clientSecret},
		"code":          {code},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://github.com/login/oauth/access_token",
		strings.NewReader(body.Encode()))
	if err != nil {
		return "", fmt.Errorf("github_oauth: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("github_oauth: exchange request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("github_oauth: decode response: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("github_oauth: %s", result.Error)
	}
	return result.AccessToken, nil
}

func (g *GitHubOAuth) fetchUser(ctx context.Context, accessToken string) (*GitHubUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("github_oauth: build user request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github_oauth: user request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github_oauth: user API returned %d: %s", resp.StatusCode, b)
	}

	var u GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("github_oauth: decode user: %w", err)
	}
	return &u, nil
}
