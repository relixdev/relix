package user

import "context"

// Store defines the interface for user persistence operations.
type Store interface {
	// CreateUser persists a new user and returns it with a generated ID.
	CreateUser(ctx context.Context, u *User) (*User, error)

	// GetUserByID retrieves a user by their unique ID.
	GetUserByID(ctx context.Context, id string) (*User, error)

	// GetUserByEmail retrieves a user by email address.
	GetUserByEmail(ctx context.Context, email string) (*User, error)

	// GetUserByGitHubID retrieves a user by their GitHub user ID.
	GetUserByGitHubID(ctx context.Context, githubID string) (*User, error)

	// GetUserByStripeCustomerID retrieves a user by their Stripe customer ID.
	GetUserByStripeCustomerID(ctx context.Context, customerID string) (*User, error)

	// UpdateUser persists changes to an existing user.
	UpdateUser(ctx context.Context, u *User) (*User, error)

	// UpsertSubscription creates or replaces the subscription record for a user.
	UpsertSubscription(ctx context.Context, sub *Subscription) error

	// GetSubscription retrieves the current subscription for a user.
	GetSubscription(ctx context.Context, userID string) (*Subscription, error)
}
