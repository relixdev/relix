CREATE TABLE users (
    id          TEXT PRIMARY KEY,
    email       TEXT UNIQUE,
    github_id   TEXT UNIQUE,
    tier        TEXT NOT NULL DEFAULT 'free',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
