package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/relixdev/relix/cloud/internal/idgen"
)

// PostgresStore implements Store using PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore returns a PostgresStore backed by the given database connection.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) CreateUser(ctx context.Context, u *User) (*User, error) {
	created := &User{
		ID:        idgen.New("usr"),
		Email:     u.Email,
		GitHubID:  u.GitHubID,
		Tier:      u.Tier,
		CreatedAt: time.Now().UTC(),
	}
	if created.Tier == "" {
		created.Tier = TierFree
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO users (id, email, github_id, tier, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		created.ID,
		nullString(created.Email),
		nullString(created.GitHubID),
		string(created.Tier),
		created.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("postgres: create user: %w", err)
	}
	return created, nil
}

func (s *PostgresStore) GetUserByID(ctx context.Context, id string) (*User, error) {
	return s.scanUser(s.db.QueryRowContext(ctx,
		`SELECT id, email, github_id, tier, created_at FROM users WHERE id = $1`, id))
}

func (s *PostgresStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.scanUser(s.db.QueryRowContext(ctx,
		`SELECT id, email, github_id, tier, created_at FROM users WHERE email = $1`, email))
}

func (s *PostgresStore) GetUserByGitHubID(ctx context.Context, githubID string) (*User, error) {
	return s.scanUser(s.db.QueryRowContext(ctx,
		`SELECT id, email, github_id, tier, created_at FROM users WHERE github_id = $1`, githubID))
}

func (s *PostgresStore) UpdateUser(ctx context.Context, u *User) (*User, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE users SET email = $1, github_id = $2, tier = $3 WHERE id = $4`,
		nullString(u.Email),
		nullString(u.GitHubID),
		string(u.Tier),
		u.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("postgres: update user: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("postgres: update user rows affected: %w", err)
	}
	if rows == 0 {
		return nil, fmt.Errorf("user %q not found", u.ID)
	}
	return s.GetUserByID(ctx, u.ID)
}

func (s *PostgresStore) UpsertSubscription(ctx context.Context, sub *Subscription) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO subscriptions (user_id, tier, expires_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id) DO UPDATE SET tier = $2, expires_at = $3`,
		sub.UserID,
		string(sub.Tier),
		sub.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("postgres: upsert subscription: %w", err)
	}
	// Also update the user's tier field for consistency with MemoryStore behavior.
	_, err = s.db.ExecContext(ctx,
		`UPDATE users SET tier = $1 WHERE id = $2`,
		string(sub.Tier),
		sub.UserID,
	)
	if err != nil {
		return fmt.Errorf("postgres: update user tier: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetSubscription(ctx context.Context, userID string) (*Subscription, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT user_id, tier, expires_at FROM subscriptions WHERE user_id = $1`, userID)

	var sub Subscription
	var tier string
	if err := row.Scan(&sub.UserID, &tier, &sub.ExpiresAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("subscription for user %q not found", userID)
		}
		return nil, fmt.Errorf("postgres: get subscription: %w", err)
	}
	sub.Tier = Tier(tier)
	return &sub, nil
}

func (s *PostgresStore) scanUser(row *sql.Row) (*User, error) {
	var u User
	var email, githubID sql.NullString
	var tier string

	if err := row.Scan(&u.ID, &email, &githubID, &tier, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("postgres: scan user: %w", err)
	}
	u.Email = email.String
	u.GitHubID = githubID.String
	u.Tier = Tier(tier)
	return &u, nil
}

// nullString converts an empty string to a sql.NullString with Valid=false.
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
