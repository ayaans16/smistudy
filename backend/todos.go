package main

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// maxTodos caps each user's list so the API can't be used to fill the disk.
const maxTodos = 200

var ErrTooManyTodos = errors.New("too many to-dos")

type Todo struct {
	ID        string     `json:"id"`
	Text      string     `json:"text"`
	Done      bool       `json:"done"`
	CreatedAt time.Time  `json:"createdAt"`
	DoneAt    *time.Time `json:"doneAt,omitempty"`
}

// --- store ---

// Todos lists a user's to-dos: open ones oldest first, then completed ones most recent first.
func (s *Store) Todos(ctx context.Context, userID string) ([]Todo, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, text, done, done_at, created_at FROM todos WHERE user_id = ?
		 ORDER BY done, done_at DESC, created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Todo{}
	for rows.Next() {
		t, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func scanTodo(row scanner) (*Todo, error) {
	var t Todo
	var doneAt, created int64
	if err := row.Scan(&t.ID, &t.Text, &t.Done, &doneAt, &created); err != nil {
		return nil, err
	}
	t.CreatedAt = time.UnixMilli(created).UTC()
	if t.Done {
		d := time.UnixMilli(doneAt).UTC()
		t.DoneAt = &d
	}
	return &t, nil
}

// AddTodo inserts a to-do unless the user already has maxTodos. The count and insert
// share one write-locked transaction, so parallel requests can't overshoot the cap.
func (s *Store) AddTodo(ctx context.Context, userID, text string) (*Todo, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var n int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM todos WHERE user_id = ?`, userID).Scan(&n); err != nil {
		return nil, err
	}
	if n >= maxTodos {
		return nil, ErrTooManyTodos
	}
	t := &Todo{ID: newID(), Text: text, CreatedAt: time.Now().UTC().Truncate(time.Millisecond)}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO todos (id, user_id, text, created_at) VALUES (?, ?, ?, ?)`,
		t.ID, userID, t.Text, t.CreatedAt.UnixMilli()); err != nil {
		return nil, err
	}
	return t, tx.Commit()
}

type TodoUpdate struct {
	Text *string `json:"text"`
	Done *bool   `json:"done"`
}

// UpdateTodo changes a to-do owned by userID; it returns nil if there's no such to-do.
func (s *Store) UpdateTodo(ctx context.Context, userID, id string, u TodoUpdate) (*Todo, error) {
	now := time.Now().UnixMilli()
	t, err := scanTodo(s.db.QueryRowContext(ctx,
		`UPDATE todos SET
			text    = COALESCE(?, text),
			done_at = CASE WHEN ? IS NULL THEN done_at WHEN ? THEN (CASE WHEN done THEN done_at ELSE ? END) ELSE 0 END,
			done    = COALESCE(?, done)
		 WHERE id = ? AND user_id = ?
		 RETURNING id, text, done, done_at, created_at`,
		u.Text, u.Done, u.Done, now, u.Done, id, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return t, err
}

func (s *Store) DeleteTodo(ctx context.Context, userID, id string) (bool, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM todos WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (s *Store) ClearDoneTodos(ctx context.Context, userID string) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM todos WHERE user_id = ? AND done = 1`, userID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
