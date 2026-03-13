package user

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set, skipping PostgreSQL tests")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Run migrations inline for test isolation.
	for _, stmt := range []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY, email TEXT UNIQUE, github_id TEXT UNIQUE,
			tier TEXT NOT NULL DEFAULT 'free', created_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS subscriptions (
			user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			tier TEXT NOT NULL DEFAULT 'free', expires_at TIMESTAMPTZ NOT NULL)`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}
	// Clean slate for each test run.
	t.Cleanup(func() {
		db.Exec("DELETE FROM subscriptions")
		db.Exec("DELETE FROM users")
	})
	return db
}

func TestPostgresStore_CreateAndGet(t *testing.T) {
	db := openTestDB(t)
	store := NewPostgresStore(db)
	ctx := context.Background()

	created, err := store.CreateUser(ctx, &User{Email: "test@example.com", GitHubID: "gh123"})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if created.Tier != TierFree {
		t.Fatalf("expected tier %q, got %q", TierFree, created.Tier)
	}

	// GetUserByID
	got, err := store.GetUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if got.Email != "test@example.com" {
		t.Fatalf("expected email %q, got %q", "test@example.com", got.Email)
	}

	// GetUserByEmail
	got, err = store.GetUserByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("expected ID %q, got %q", created.ID, got.ID)
	}

	// GetUserByGitHubID
	got, err = store.GetUserByGitHubID(ctx, "gh123")
	if err != nil {
		t.Fatalf("GetUserByGitHubID: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("expected ID %q, got %q", created.ID, got.ID)
	}
}

func TestPostgresStore_UpdateUser(t *testing.T) {
	db := openTestDB(t)
	store := NewPostgresStore(db)
	ctx := context.Background()

	created, err := store.CreateUser(ctx, &User{Email: "orig@example.com"})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	created.Email = "updated@example.com"
	created.Tier = TierPlus
	updated, err := store.UpdateUser(ctx, created)
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if updated.Email != "updated@example.com" {
		t.Fatalf("expected email %q, got %q", "updated@example.com", updated.Email)
	}
	if updated.Tier != TierPlus {
		t.Fatalf("expected tier %q, got %q", TierPlus, updated.Tier)
	}

	// Old email should not resolve.
	_, err = store.GetUserByEmail(ctx, "orig@example.com")
	if err == nil {
		t.Fatal("expected error for old email, got nil")
	}
}

func TestPostgresStore_DuplicateEmail(t *testing.T) {
	db := openTestDB(t)
	store := NewPostgresStore(db)
	ctx := context.Background()

	_, err := store.CreateUser(ctx, &User{Email: "dup@example.com"})
	if err != nil {
		t.Fatalf("first CreateUser: %v", err)
	}
	_, err = store.CreateUser(ctx, &User{Email: "dup@example.com"})
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
}

func TestPostgresStore_Subscription(t *testing.T) {
	db := openTestDB(t)
	store := NewPostgresStore(db)
	ctx := context.Background()

	created, err := store.CreateUser(ctx, &User{Email: "sub@example.com"})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	expires := time.Now().Add(30 * 24 * time.Hour).UTC().Truncate(time.Microsecond)
	err = store.UpsertSubscription(ctx, &Subscription{
		UserID:    created.ID,
		Tier:      TierPro,
		ExpiresAt: expires,
	})
	if err != nil {
		t.Fatalf("UpsertSubscription: %v", err)
	}

	sub, err := store.GetSubscription(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetSubscription: %v", err)
	}
	if sub.Tier != TierPro {
		t.Fatalf("expected tier %q, got %q", TierPro, sub.Tier)
	}

	// Verify user tier was also updated.
	u, err := store.GetUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if u.Tier != TierPro {
		t.Fatalf("expected user tier %q after subscription, got %q", TierPro, u.Tier)
	}
}

func TestPostgresStore_NotFound(t *testing.T) {
	db := openTestDB(t)
	store := NewPostgresStore(db)
	ctx := context.Background()

	_, err := store.GetUserByID(ctx, "usr_nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
	_, err = store.GetUserByEmail(ctx, "nobody@example.com")
	if err == nil {
		t.Fatal("expected error for nonexistent email")
	}
	_, err = store.GetUserByGitHubID(ctx, "gh_none")
	if err == nil {
		t.Fatal("expected error for nonexistent github ID")
	}
	_, err = store.GetSubscription(ctx, "usr_nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent subscription")
	}
}
