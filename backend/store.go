package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const dateLayout = "2006-01-02"

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Session is a single block of study time logged on a given local date.
type Session struct {
	ID        string    `json:"id"`
	Date      string    `json:"date"`    // YYYY-MM-DD in the user's local time
	Minutes   int       `json:"minutes"` // length of the session
	Kind      string    `json:"kind"`    // "pomodoro" or "manual"
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type DayTotal struct {
	Minutes  int
	Sessions int
}

// Store is the SQLite-backed data layer. Every query uses bound parameters;
// no user input is ever concatenated into SQL.
type Store struct {
	db *sql.DB
}

func OpenStore(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	// busy_timeout goes first so every later step (including switching to WAL) waits for
	// other processes instead of failing. _txlock=immediate takes the write lock when a
	// transaction starts, so concurrent writers queue rather than deadlock on upgrade.
	dsn := "file:" + path +
		"?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)" +
		"&_txlock=immediate"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	// Switching a brand-new database to WAL ignores busy_timeout when another process is
	// doing the same, so retry briefly on "busy" while starting up.
	for attempt := 0; ; attempt++ {
		err = s.migrate(context.Background())
		if err == nil || attempt == 50 || !strings.Contains(err.Error(), "SQLITE_BUSY") {
			break
		}
		time.Sleep(time.Duration(20+attempt*10) * time.Millisecond)
	}
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

// migrate applies migrations/NNNN_*.sql in order, each exactly once.
func (s *Store) migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY)`); err != nil {
		return err
	}
	names, err := fs.Glob(migrationFiles, "migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)
	for _, name := range names {
		if err := s.applyMigration(ctx, name); err != nil {
			return err
		}
	}
	return nil
}

// applyMigration runs one migration unless it's already recorded. The check happens
// inside the write-locked transaction, so when several processes start at once only
// the first applies it and the rest see it as done.
func (s *Store) applyMigration(ctx context.Context, name string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var done int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, name).Scan(&done); err != nil {
		return err
	}
	if done > 0 {
		return nil
	}
	body, err := migrationFiles.ReadFile(name)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, string(body)); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (name) VALUES (?)`, name); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) Add(ctx context.Context, userID string, sess Session) (Session, error) {
	sess.ID = newID()
	sess.CreatedAt = time.Now().UTC().Truncate(time.Millisecond)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO study_sessions (id, user_id, date, minutes, kind, note, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sess.ID, userID, sess.Date, sess.Minutes, sess.Kind, sess.Note, sess.CreatedAt.UnixMilli())
	if err != nil {
		return Session{}, err
	}
	return sess, nil
}

// Delete removes a session only if it belongs to userID.
func (s *Store) Delete(ctx context.Context, userID, id string) (bool, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM study_sessions WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// OnDate returns the sessions logged on a date, oldest first.
func (s *Store) OnDate(ctx context.Context, userID, date string) ([]Session, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, date, minutes, kind, note, created_at FROM study_sessions
		 WHERE user_id = ? AND date = ? ORDER BY created_at`, userID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Session{}
	for rows.Next() {
		var sess Session
		var created int64
		if err := rows.Scan(&sess.ID, &sess.Date, &sess.Minutes, &sess.Kind, &sess.Note, &created); err != nil {
			return nil, err
		}
		sess.CreatedAt = time.UnixMilli(created).UTC()
		out = append(out, sess)
	}
	return out, rows.Err()
}

// DailyTotals sums minutes and session counts per date.
func (s *Store) DailyTotals(ctx context.Context, userID string) (map[string]DayTotal, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT date, SUM(minutes), COUNT(*) FROM study_sessions WHERE user_id = ? GROUP BY date`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	totals := make(map[string]DayTotal)
	for rows.Next() {
		var date string
		var t DayTotal
		if err := rows.Scan(&date, &t.Minutes, &t.Sessions); err != nil {
			return nil, err
		}
		totals[date] = t
	}
	return totals, rows.Err()
}

// ImportLegacyJSON loads sessions from the old JSON-file store, if present and the
// database has no sessions yet, then renames the file so it isn't imported twice.
func (s *Store) ImportLegacyJSON(ctx context.Context, path string) (int, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var legacy []Session
	if err := json.Unmarshal(data, &legacy); err != nil {
		return 0, fmt.Errorf("parse %s: %w", path, err)
	}
	// Check-and-insert under one write lock, so concurrent startups import only once.
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var existing int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM study_sessions`).Scan(&existing); err != nil {
		return 0, err
	}
	if existing > 0 {
		return 0, nil
	}
	for _, sess := range legacy {
		if sess.ID == "" {
			sess.ID = newID()
		}
		if sess.Kind != "pomodoro" {
			sess.Kind = "manual"
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO study_sessions (id, date, minutes, kind, note, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
			sess.ID, sess.Date, sess.Minutes, sess.Kind, strings.TrimSpace(sess.Note), sess.CreatedAt.UnixMilli()); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return len(legacy), os.Rename(path, path+".imported")
}

// ClaimLegacy gives every ownerless session (from the single-user era) to userID.
func (s *Store) ClaimLegacy(ctx context.Context, userID string) (int64, error) {
	res, err := s.db.ExecContext(ctx, `UPDATE study_sessions SET user_id = ? WHERE user_id IS NULL`, userID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Backup writes a transactionally consistent snapshot of the database to path.
func (s *Store) Backup(ctx context.Context, path string) error {
	_, err := s.db.ExecContext(ctx, `VACUUM INTO ?`, path)
	return err
}

func newID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
