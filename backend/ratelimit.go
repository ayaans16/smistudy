package main

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Limiter is a set of per-key token buckets (key = client IP, email, ...).
// It's in-memory, so each API process enforces its own limits; per-account
// login lockout lives in the database so it holds across processes.
type Limiter struct {
	mu      sync.Mutex
	every   rate.Limit
	burst   int
	buckets map[string]*bucket
}

type bucket struct {
	lim  *rate.Limiter
	seen time.Time
}

// NewLimiter allows `burst` events immediately, refilling one every `per`.
func NewLimiter(per time.Duration, burst int) *Limiter {
	l := &Limiter{every: rate.Every(per), burst: burst, buckets: make(map[string]*bucket)}
	go l.janitor()
	return l
}

// Allow reports whether an event for key may happen now, and if not, how long to wait.
func (l *Limiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{lim: rate.NewLimiter(l.every, l.burst)}
		l.buckets[key] = b
	}
	b.seen = time.Now()
	l.mu.Unlock()

	r := b.lim.Reserve()
	if d := r.Delay(); d > 0 {
		r.Cancel()
		return false, d
	}
	return true, 0
}

// janitor drops buckets idle long enough to have fully refilled.
func (l *Limiter) janitor() {
	idle := time.Duration(float64(time.Second)/float64(l.every))*time.Duration(l.burst) + time.Minute
	for range time.Tick(time.Minute) {
		cutoff := time.Now().Add(-idle)
		l.mu.Lock()
		for k, b := range l.buckets {
			if b.seen.Before(cutoff) {
				delete(l.buckets, k)
			}
		}
		l.mu.Unlock()
	}
}

type limits struct {
	api, signup, login, loginEmail, emailSend, token, oauth, follow *Limiter
}

func newLimits() *limits {
	return &limits{
		api:        NewLimiter(50*time.Millisecond, 60), // 20 req/s per IP, bursts of 60
		signup:     NewLimiter(12*time.Minute, 5),       // 5/hour per IP
		login:      NewLimiter(6*time.Second, 10),       // 10/min per IP
		loginEmail: NewLimiter(12*time.Second, 5),       // 5/min per account
		emailSend:  NewLimiter(20*time.Minute, 3),       // 3 emails/hour per address
		token:      NewLimiter(3*time.Minute, 20),       // 20 token checks/hour per IP
		oauth:      NewLimiter(3*time.Second, 20),       // 20/min per IP
		follow:     NewLimiter(2*time.Second, 30),       // 30 follows/blocks per minute per user
	}
}
