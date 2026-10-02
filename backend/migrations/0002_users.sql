CREATE TABLE users (
    id             TEXT PRIMARY KEY,
    email          TEXT NOT NULL UNIQUE COLLATE NOCASE,
    email_verified INTEGER NOT NULL DEFAULT 0,
    password_hash  TEXT,                    -- NULL for Google-only accounts
    google_sub     TEXT UNIQUE,             -- Google account id, NULL if not linked
    username       TEXT NOT NULL UNIQUE COLLATE NOCASE,
    display_name   TEXT NOT NULL DEFAULT '',
    profile_public INTEGER NOT NULL DEFAULT 0,
    failed_logins  INTEGER NOT NULL DEFAULT 0,
    locked_until   INTEGER NOT NULL DEFAULT 0, -- unix ms
    created_at     INTEGER NOT NULL
);

-- Login sessions. Only a SHA-256 of the cookie token is stored.
CREATE TABLE auth_sessions (
    token_hash TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    user_agent TEXT NOT NULL DEFAULT ''
);
CREATE INDEX auth_sessions_user ON auth_sessions (user_id);

-- Single-use email verification and password reset tokens (hashed).
CREATE TABLE auth_tokens (
    token_hash TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    purpose    TEXT NOT NULL CHECK (purpose IN ('verify', 'reset')),
    expires_at INTEGER NOT NULL
);
CREATE INDEX auth_tokens_user ON auth_tokens (user_id, purpose);

-- Existing single-user rows keep a NULL owner until claimed (see `claim-legacy`).
ALTER TABLE study_sessions ADD COLUMN user_id TEXT REFERENCES users (id) ON DELETE CASCADE;
CREATE INDEX study_sessions_user_date ON study_sessions (user_id, date);
