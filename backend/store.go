package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const dateLayout = "2006-01-02"

// Session is a single block of study time logged on a given local date.
type Session struct {
	ID        string    `json:"id"`
	Date      string    `json:"date"`    // YYYY-MM-DD in the user's local time
	Minutes   int       `json:"minutes"` // length of the session
	Kind      string    `json:"kind"`    // "pomodoro" or "manual"
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// Store keeps sessions in memory and persists them to a JSON file.
type Store struct {
	mu       sync.RWMutex
	path     string
	sessions []Session
}

func OpenStore(path string) (*Store, error) {
	s := &Store{path: path}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &s.sessions); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// save writes the file atomically. Caller must hold the write lock.
func (s *Store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.sessions, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) Add(sess Session) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess.ID = newID()
	sess.CreatedAt = time.Now().UTC()
	s.sessions = append(s.sessions, sess)
	if err := s.save(); err != nil {
		s.sessions = s.sessions[:len(s.sessions)-1]
		return Session{}, err
	}
	return sess, nil
}

func (s *Store) Delete(id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, sess := range s.sessions {
		if sess.ID == id {
			s.sessions = append(s.sessions[:i], s.sessions[i+1:]...)
			return true, s.save()
		}
	}
	return false, nil
}

// OnDate returns the sessions logged on a date, oldest first.
func (s *Store) OnDate(date string) []Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Session{}
	for _, sess := range s.sessions {
		if sess.Date == date {
			out = append(out, sess)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

// DailyTotals sums minutes and session counts per date.
func (s *Store) DailyTotals() map[string]DayTotal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	totals := make(map[string]DayTotal)
	for _, sess := range s.sessions {
		t := totals[sess.Date]
		t.Minutes += sess.Minutes
		t.Sessions++
		totals[sess.Date] = t
	}
	return totals
}

type DayTotal struct {
	Minutes  int
	Sessions int
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
