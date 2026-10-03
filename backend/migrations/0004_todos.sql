-- Personal to-do lists. Private to their owner; never shown on public profiles.
CREATE TABLE todos (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    text       TEXT NOT NULL CHECK (length(text) BETWEEN 1 AND 200),
    done       INTEGER NOT NULL DEFAULT 0,
    done_at    INTEGER NOT NULL DEFAULT 0, -- unix ms, 0 while not done
    created_at INTEGER NOT NULL            -- unix ms
);
CREATE INDEX todos_user ON todos (user_id, done, created_at);
