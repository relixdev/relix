CREATE TABLE subscriptions (
    user_id    TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    tier       TEXT NOT NULL DEFAULT 'free',
    expires_at TIMESTAMPTZ NOT NULL
);
