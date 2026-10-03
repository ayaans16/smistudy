-- Record when each user agreed to the Terms of Service and Privacy Policy, and which version.
ALTER TABLE users ADD COLUMN terms_accepted_at INTEGER NOT NULL DEFAULT 0; -- unix ms, 0 = before this was tracked
ALTER TABLE users ADD COLUMN terms_version TEXT NOT NULL DEFAULT '';
