-- Personal reward goals: "after N hours of studying, I get <reward>". Private to their owner.
CREATE TABLE reward_goals (
    id             TEXT PRIMARY KEY,
    user_id        TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    reward         TEXT NOT NULL CHECK (length(reward) BETWEEN 1 AND 100),
    target_minutes INTEGER NOT NULL CHECK (target_minutes BETWEEN 30 AND 60000),
    start_date     TEXT NOT NULL CHECK (start_date GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]'),
    created_at     INTEGER NOT NULL, -- unix ms
    claimed_at     INTEGER NOT NULL DEFAULT 0 -- unix ms, 0 until the reward is claimed
);
CREATE INDEX reward_goals_user ON reward_goals (user_id, claimed_at);
