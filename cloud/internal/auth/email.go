package auth

import (
	"context"
	"fmt"

	"github.com/relixdev/relix/cloud/internal/user"
	"golang.org/x/crypto/bcrypt"
)

// EmailAuth provides email/password registration and login against a user.Store.
type EmailAuth struct {
	store user.Store
}

// NewEmailAuth creates an EmailAuth backed by the given store.
func NewEmailAuth(store user.Store) *EmailAuth {
	return &EmailAuth{store: store}
}

// Register creates a new user with a hashed password stored in a separate
// mechanism. For now we store the bcrypt hash as GitHubID (re-used field) to
// avoid schema changes; a real DB implementation would have a password_hash column.
//
// Returns the newly created user.
func (e *EmailAuth) Register(ctx context.Context, email, password string) (*user.User, error) {
	if email == "" || password == "" {
		return nil, fmt.Errorf("email_auth: email and password are required")
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("email_auth: password must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("email_auth: hash password: %w", err)
	}

	u := &user.User{
		Email:    email,
		GitHubID: "pw:" + string(hash), // prefix distinguishes from real GitHub IDs
		Tier:     user.TierFree,
	}
	return e.store.CreateUser(ctx, u)
}

// Login authenticates a user by email and password, returning the user on success.
func (e *EmailAuth) Login(ctx context.Context, email, password string) (*user.User, error) {
	u, err := e.store.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("email_auth: invalid credentials")
	}

	if len(u.GitHubID) < 3 || u.GitHubID[:3] != "pw:" {
		return nil, fmt.Errorf("email_auth: account uses OAuth login")
	}
	hash := u.GitHubID[3:]

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return nil, fmt.Errorf("email_auth: invalid credentials")
	}
	return u, nil
}
