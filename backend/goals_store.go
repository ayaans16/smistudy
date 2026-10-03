package main

import (
	"context"
	"errors"
	"time"
)

// maxGoals caps how many reward goals a user can have (claimed ones included).
const maxGoals = 50

var ErrTooManyGoals = errors.New("too many goals")

// Goal is a stored reward goal; progress fields are filled in from study data.
type Goal struct {
	ID            string     `json:"id"`
	Reward        string     `json:"reward"`
	TargetMinutes int        `json:"targetMinutes"`
	StartDate     string     `json:"startDate"`
	CreatedAt     time.Time  `json:"createdAt"`
	ClaimedAt     *time.Time `json:"claimedAt,omitempty"`

	Minutes   int    `json:"minutes"`             // studied since StartDate
	ReachedOn string `json:"reachedOn,omitempty"` // date the target was hit
}

// Goals lists a user's goals: active ones oldest first, then claimed ones newest first.
func (s *Store) Goals(ctx context.Context, userID string) ([]Goal, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, reward, target_minutes, start_date, created_at, claimed_at FROM reward_goals
		 WHERE user_id = ? ORDER BY claimed_at != 0, claimed_at DESC, created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Goal{}
	for rows.Next() {
		var g Goal
		var created, claimed int64
		if err := rows.Scan(&g.ID, &g.Reward, &g.TargetMinutes, &g.StartDate, &created, &claimed); err != nil {
			return nil, err
		}
		g.CreatedAt = time.UnixMilli(created).UTC()
		if claimed != 0 {
			c := time.UnixMilli(claimed).UTC()
			g.ClaimedAt = &c
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// AddGoal stores a goal unless the user already has maxGoals (checked in the same
// write-locked transaction so parallel requests can't overshoot).
func (s *Store) AddGoal(ctx context.Context, userID string, g *Goal) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var n int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM reward_goals WHERE user_id = ?`, userID).Scan(&n); err != nil {
		return err
	}
	if n >= maxGoals {
		return ErrTooManyGoals
	}
	g.ID = newID()
	g.CreatedAt = time.Now().UTC().Truncate(time.Millisecond)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO reward_goals (id, user_id, reward, target_minutes, start_date, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		g.ID, userID, g.Reward, g.TargetMinutes, g.StartDate, g.CreatedAt.UnixMilli()); err != nil {
		return err
	}
	return tx.Commit()
}

// ClaimGoal marks a goal as claimed; it reports false if userID has no such unclaimed goal.
func (s *Store) ClaimGoal(ctx context.Context, userID, id string) (bool, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE reward_goals SET claimed_at = ? WHERE id = ? AND user_id = ? AND claimed_at = 0`,
		time.Now().UnixMilli(), id, userID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (s *Store) DeleteGoal(ctx context.Context, userID, id string) (bool, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM reward_goals WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}
