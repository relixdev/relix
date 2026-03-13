package user

import (
	"database/sql"
	"fmt"
)

// Migrate runs the schema migrations needed for the PostgresStore.
// It is idempotent — safe to call on every startup.
func Migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id          TEXT PRIMARY KEY,
			email       TEXT UNIQUE,
			github_id   TEXT UNIQUE,
			tier        TEXT NOT NULL DEFAULT 'free',
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS machines (
			id          TEXT PRIMARY KEY,
			user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name        TEXT NOT NULL,
			public_key  TEXT NOT NULL,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS device_tokens (
			id           TEXT PRIMARY KEY,
			user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			device_token TEXT NOT NULL,
			platform     TEXT NOT NULL,
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(user_id, device_token)
		)`,
		`CREATE TABLE IF NOT EXISTS subscriptions (
			user_id    TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			tier       TEXT NOT NULL DEFAULT 'free',
			expires_at TIMESTAMPTZ NOT NULL
		)`,
	}
	for i, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
	}
	return nil
}
