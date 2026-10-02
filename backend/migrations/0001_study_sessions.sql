CREATE TABLE study_sessions (
    id         TEXT PRIMARY KEY,
    date       TEXT NOT NULL CHECK (date GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]'),
    minutes    INTEGER NOT NULL CHECK (minutes BETWEEN 1 AND 1440),
    kind       TEXT NOT NULL CHECK (kind IN ('pomodoro', 'manual')),
    note       TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL -- unix milliseconds
);

CREATE INDEX study_sessions_date ON study_sessions (date);
