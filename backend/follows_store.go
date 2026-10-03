package main

import (
	"context"
	"errors"
	"time"
)

// maxFollowing caps how many people one user can follow (anti-spam).
const maxFollowing = 1000

var (
	ErrBlocked          = errors.New("one of these users has blocked the other")
	ErrFollowingTooMany = errors.New("following too many people")
)

// FollowUser is a user as shown in follower/following/blocked lists.
type FollowUser struct {
	ID          string    `json:"-"`
	Username    string    `json:"username"`
	DisplayName string    `json:"displayName"`
	Since       time.Time `json:"since"` // when the follow (or block) happened
}

// Follow makes followerID follow followeeID. Following someone twice is a no-op.
func (s *Store) Follow(ctx context.Context, followerID, followeeID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var blocked int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM blocks WHERE (blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)`,
		followerID, followeeID, followeeID, followerID).Scan(&blocked); err != nil {
		return err
	}
	if blocked > 0 {
		return ErrBlocked
	}
	var n int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM follows WHERE follower_id = ?`, followerID).Scan(&n); err != nil {
		return err
	}
	if n >= maxFollowing {
		return ErrFollowingTooMany
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT OR IGNORE INTO follows (follower_id, followee_id, created_at) VALUES (?, ?, ?)`,
		followerID, followeeID, time.Now().UnixMilli()); err != nil {
		return err
	}
	return tx.Commit()
}

// Unfollow removes a follow; it reports whether there was one.
func (s *Store) Unfollow(ctx context.Context, followerID, followeeID string) (bool, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM follows WHERE follower_id = ? AND followee_id = ?`, followerID, followeeID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (s *Store) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM follows WHERE follower_id = ? AND followee_id = ?`, followerID, followeeID).Scan(&n)
	return n > 0, err
}

// Block blocks a user and removes any follows between the two.
func (s *Store) Block(ctx context.Context, blockerID, blockedID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx,
		`INSERT OR IGNORE INTO blocks (blocker_id, blocked_id, created_at) VALUES (?, ?, ?)`,
		blockerID, blockedID, time.Now().UnixMilli()); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM follows WHERE (follower_id = ? AND followee_id = ?) OR (follower_id = ? AND followee_id = ?)`,
		blockerID, blockedID, blockedID, blockerID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) Unblock(ctx context.Context, blockerID, blockedID string) (bool, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM blocks WHERE blocker_id = ? AND blocked_id = ?`, blockerID, blockedID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// Lists only ever include users whose profiles are public: going private hides you
// from every follower and following list.

// Followers lists the public users who follow userID, newest first.
func (s *Store) Followers(ctx context.Context, userID string) ([]FollowUser, error) {
	return s.followList(ctx,
		`SELECT u.id, u.username, u.display_name, f.created_at FROM follows f JOIN users u ON u.id = f.follower_id
		 WHERE f.followee_id = ? AND u.profile_public = 1 ORDER BY f.created_at DESC`, userID)
}

// Following lists the public users userID follows, newest first.
func (s *Store) Following(ctx context.Context, userID string) ([]FollowUser, error) {
	return s.followList(ctx,
		`SELECT u.id, u.username, u.display_name, f.created_at FROM follows f JOIN users u ON u.id = f.followee_id
		 WHERE f.follower_id = ? AND u.profile_public = 1 ORDER BY f.created_at DESC`, userID)
}

// Blocked lists everyone userID has blocked (private to userID).
func (s *Store) Blocked(ctx context.Context, userID string) ([]FollowUser, error) {
	return s.followList(ctx,
		`SELECT u.id, u.username, u.display_name, b.created_at FROM blocks b JOIN users u ON u.id = b.blocked_id
		 WHERE b.blocker_id = ? ORDER BY b.created_at DESC`, userID)
}

func (s *Store) followList(ctx context.Context, query, userID string) ([]FollowUser, error) {
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []FollowUser{}
	for rows.Next() {
		var f FollowUser
		var since int64
		if err := rows.Scan(&f.ID, &f.Username, &f.DisplayName, &since); err != nil {
			return nil, err
		}
		f.Since = time.UnixMilli(since).UTC()
		out = append(out, f)
	}
	return out, rows.Err()
}

// FollowCounts counts a user's public followers and the public users they follow.
func (s *Store) FollowCounts(ctx context.Context, userID string) (followers, following int, err error) {
	err = s.db.QueryRowContext(ctx,
		`SELECT
			(SELECT COUNT(*) FROM follows f JOIN users u ON u.id = f.follower_id WHERE f.followee_id = ? AND u.profile_public = 1),
			(SELECT COUNT(*) FROM follows f JOIN users u ON u.id = f.followee_id WHERE f.follower_id = ? AND u.profile_public = 1)`,
		userID, userID).Scan(&followers, &following)
	return followers, following, err
}
