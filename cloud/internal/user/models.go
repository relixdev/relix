package user

import "time"

// Tier represents the subscription tier for a user.
type Tier string

const (
	TierFree Tier = "free"
	TierPlus Tier = "plus"
	TierPro  Tier = "pro"
	TierTeam Tier = "team"
)

// User represents an authenticated user of the Relix platform.
type User struct {
	ID        string
	Email     string
	GitHubID  string
	Tier      Tier
	CreatedAt time.Time
}

// Machine represents a registered developer machine belonging to a user.
type Machine struct {
	ID        string
	UserID    string
	Name      string
	PublicKey string
	CreatedAt time.Time
}

// Subscription tracks billing state for a user.
type Subscription struct {
	UserID    string
	Tier      Tier
	ExpiresAt time.Time
}
