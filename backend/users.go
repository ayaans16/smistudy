package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

var (
	ErrEmailTaken    = errors.New("email already registered")
	ErrUsernameTaken = errors.New("username taken")
)

type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	PasswordHash  string    `json:"-"`
	GoogleSub     string    `json:"-"`
	Username      string    `json:"username"`
	DisplayName   string    `json:"displayName"`
	ProfilePublic bool      `json:"profilePublic"`
	FailedLogins  int       `json:"-"`
	LockedUntil   time.Time `json:"-"`
	CreatedAt     time.Time `json:"createdAt"`
	TermsVersion  string    `json:"termsVersion"`
}

// TermsVersion is the "last updated" date of the current Terms and Privacy Policy.
// Bump it (and the date on the pages) whenever either document changes materially.
const TermsVersion = "2026-10-03"

const userColumns = `id, email, email_verified, COALESCE(password_hash, ''), COALESCE(google_sub, ''),
	username, display_name, profile_public, failed_logins, locked_until, created_at, terms_version`

type scanner interface{ Scan(...any) error }

func scanUser(row scanner) (*User, error) {
	var u User
	var locked, created int64
	err := row.Scan(&u.ID, &u.Email, &u.EmailVerified, &u.PasswordHash, &u.GoogleSub,
		&u.Username, &u.DisplayName, &u.ProfilePublic, &u.FailedLogins, &locked, &created, &u.TermsVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.LockedUntil = time.UnixMilli(locked)
	u.CreatedAt = time.UnixMilli(created).UTC()
	return &u, nil
}

// userBy looks a user up by one of a fixed set of columns (never user input).
func (s *Store) userBy(ctx context.Context, column, value string) (*User, error) {
	switch column {
	case "id", "email", "username", "google_sub":
	default:
		panic("userBy: unexpected column " + column)
	}
	return scanUser(s.db.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE `+column+` = ?`, value))
}

func (s *Store) UserByID(ctx context.Context, id string) (*User, error) {
	return s.userBy(ctx, "id", id)
}
func (s *Store) UserByEmail(ctx context.Context, email string) (*User, error) {
	return s.userBy(ctx, "email", email)
}
func (s *Store) UserByUsername(ctx context.Context, username string) (*User, error) {
	return s.userBy(ctx, "username", username)
}
func (s *Store) UserByGoogleSub(ctx context.Context, sub string) (*User, error) {
	return s.userBy(ctx, "google_sub", sub)
}

func (s *Store) CreateUser(ctx context.Context, u *User) error {
	u.ID = newID()
	u.CreatedAt = time.Now().UTC().Truncate(time.Millisecond)
	u.TermsVersion = TermsVersion // callers only create users who have agreed to the current terms
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO users (id, email, email_verified, password_hash, google_sub, username, display_name, created_at,
		                    terms_accepted_at, terms_version)
		 VALUES (?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?, ?)`,
		u.ID, u.Email, u.EmailVerified, u.PasswordHash, u.GoogleSub, u.Username, u.DisplayName, u.CreatedAt.UnixMilli(),
		u.CreatedAt.UnixMilli(), u.TermsVersion)
	return uniqueErr(err)
}

type ProfileUpdate struct {
	Username      *string `json:"username"`
	DisplayName   *string `json:"displayName"`
	ProfilePublic *bool   `json:"profilePublic"`
}

func (s *Store) UpdateProfile(ctx context.Context, userID string, p ProfileUpdate) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET
			username       = COALESCE(?, username),
			display_name   = COALESCE(?, display_name),
			profile_public = COALESCE(?, profile_public)
		 WHERE id = ?`,
		p.Username, p.DisplayName, p.ProfilePublic, userID)
	return uniqueErr(err)
}

func (s *Store) SetPassword(ctx context.Context, userID, hash string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ?, failed_logins = 0, locked_until = 0 WHERE id = ?`, hash, userID)
	return err
}

func (s *Store) MarkEmailVerified(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET email_verified = 1 WHERE id = ?`, userID)
	return err
}

// LinkGoogle attaches a Google account. If the email was never verified, whoever set
// the existing password may not own the inbox, so that password and its sessions are
// dropped (prevents pre-registering someone else's email to hijack their account).
func (s *Store) LinkGoogle(ctx context.Context, u *User, sub string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if !u.EmailVerified {
		if _, err := tx.ExecContext(ctx, `UPDATE users SET password_hash = NULL WHERE id = ?`, u.ID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM auth_sessions WHERE user_id = ?`, u.ID); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET google_sub = ?, email_verified = 1 WHERE id = ?`, sub, u.ID); err != nil {
		return uniqueErr(err)
	}
	return tx.Commit()
}

func (s *Store) DeleteUser(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, userID)
	return err
}

// RecordFailedLogin counts a bad password and locks the account with exponential
// backoff after 5 consecutive failures (1 min, 2 min, 4 min, ... up to 1 hour).
func (s *Store) RecordFailedLogin(ctx context.Context, u *User) error {
	fails := u.FailedLogins + 1
	var locked int64
	if fails >= 5 {
		backoff := time.Minute << min(fails-5, 6)
		locked = time.Now().Add(min(backoff, time.Hour)).UnixMilli()
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET failed_logins = ?, locked_until = ? WHERE id = ?`, fails, locked, u.ID)
	return err
}

func (s *Store) ResetFailedLogins(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET failed_logins = 0, locked_until = 0 WHERE id = ?`, userID)
	return err
}

// --- login sessions ---

const sessionTTL = 30 * 24 * time.Hour

// CreateAuthSession returns the raw token for the cookie; only its hash is stored.
func (s *Store) CreateAuthSession(ctx context.Context, userID, userAgent string) (string, time.Time, error) {
	token := randomToken()
	now := time.Now()
	expires := now.Add(sessionTTL)
	if len(userAgent) > 200 {
		userAgent = userAgent[:200]
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO auth_sessions (token_hash, user_id, created_at, expires_at, user_agent) VALUES (?, ?, ?, ?, ?)`,
		hashToken(token), userID, now.UnixMilli(), expires.UnixMilli(), userAgent)
	return token, expires, err
}

// UserForSession resolves a cookie token, sliding its expiry forward at most once a day.
func (s *Store) UserForSession(ctx context.Context, token string) (*User, error) {
	h := hashToken(token)
	var userID string
	var expires int64
	err := s.db.QueryRowContext(ctx,
		`SELECT user_id, expires_at FROM auth_sessions WHERE token_hash = ?`, h).Scan(&userID, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if now.UnixMilli() >= expires {
		_, err := s.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE token_hash = ?`, h)
		return nil, err
	}
	if time.UnixMilli(expires).Sub(now) < sessionTTL-24*time.Hour {
		if _, err := s.db.ExecContext(ctx, `UPDATE auth_sessions SET expires_at = ? WHERE token_hash = ?`,
			now.Add(sessionTTL).UnixMilli(), h); err != nil {
			return nil, err
		}
	}
	return s.UserByID(ctx, userID)
}

func (s *Store) DeleteAuthSession(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE token_hash = ?`, hashToken(token))
	return err
}

// DeleteOtherSessions signs a user out everywhere except (optionally) the current session.
func (s *Store) DeleteOtherSessions(ctx context.Context, userID, keepToken string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM auth_sessions WHERE user_id = ? AND token_hash != ?`, userID, hashToken(keepToken))
	return err
}

// --- email verification / password reset tokens ---

var tokenTTL = map[string]time.Duration{"verify": 24 * time.Hour, "reset": time.Hour}

// CreateToken issues a fresh single-use token, replacing any earlier one for the same purpose.
func (s *Store) CreateToken(ctx context.Context, userID, purpose string) (string, error) {
	token := randomToken()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM auth_tokens WHERE user_id = ? AND purpose = ?`, userID, purpose); err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO auth_tokens (token_hash, user_id, purpose, expires_at) VALUES (?, ?, ?, ?)`,
		hashToken(token), userID, purpose, time.Now().Add(tokenTTL[purpose]).UnixMilli()); err != nil {
		return "", err
	}
	return token, tx.Commit()
}

// ConsumeToken returns the token's user and deletes it; expired or unknown tokens return "".
func (s *Store) ConsumeToken(ctx context.Context, token, purpose string) (string, error) {
	var userID string
	var expires int64
	err := s.db.QueryRowContext(ctx,
		`DELETE FROM auth_tokens WHERE token_hash = ? AND purpose = ? RETURNING user_id, expires_at`,
		hashToken(token), purpose).Scan(&userID, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if time.Now().UnixMilli() >= expires {
		return "", nil
	}
	return userID, nil
}

// PurgeExpired removes stale sessions and tokens.
func (s *Store) PurgeExpired(ctx context.Context) error {
	now := time.Now().UnixMilli()
	if _, err := s.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE expires_at <= ?`, now); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM auth_tokens WHERE expires_at <= ?`, now)
	return err
}

func randomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// uniqueErr maps SQLite unique-constraint failures to friendly errors.
func uniqueErr(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "UNIQUE") && strings.Contains(msg, "users.email"):
		return ErrEmailTaken
	case strings.Contains(msg, "UNIQUE") && strings.Contains(msg, "users.username"):
		return ErrUsernameTaken
	}
	return err
}
