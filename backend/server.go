package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

var errBadFilter = errors.New("filter must be \"last\" or a year")

type App struct {
	cfg    Config
	store  *Store
	mail   Mailer
	lim    *limits
	google *oauth2.Config // nil when Google sign-in isn't configured
}

func NewApp(cfg Config, store *Store, mail Mailer) *App {
	return &App{cfg: cfg, store: store, mail: mail, lim: newLimits(), google: newGoogleConfig(cfg)}
}

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	a.authRoutes(mux)
	a.profileRoutes(mux)
	a.todoRoutes(mux)
	a.goalRoutes(mux)
	a.followRoutes(mux)

	// Study data — every route below acts only on the signed-in user's rows.
	mux.HandleFunc("GET /api/contributions", a.requireUser(a.handleContributions))
	mux.HandleFunc("GET /api/stats", a.requireUser(a.handleStats))
	mux.HandleFunc("GET /api/years", a.requireUser(a.handleYears))
	mux.HandleFunc("GET /api/sessions", a.requireUser(a.handleListSessions))
	mux.HandleFunc("POST /api/sessions", a.requireUser(a.handleAddSession))
	mux.HandleFunc("DELETE /api/sessions/{id}", a.requireUser(a.handleDeleteSession))

	csrf := http.NewCrossOriginProtection()
	if err := csrf.AddTrustedOrigin(a.cfg.PublicURL); err != nil {
		log.Fatalf("csrf origin: %v", err)
	}

	var h http.Handler = mux
	h = csrf.Handler(h)
	h = a.rateLimitAPI(h)
	h = securityHeaders(h)
	h = recoverPanics(h)
	return h
}

// --- study handlers ---

func (a *App) handleContributions(w http.ResponseWriter, r *http.Request, u *User) {
	today, ok := todayParam(w, r)
	if !ok {
		return
	}
	from, to, err := RangeFor(r.URL.Query().Get("filter"), today)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if totals, ok := a.totals(w, r, u.ID); ok {
		writeJSON(w, http.StatusOK, BuildCalendar(totals, from, to))
	}
}

func (a *App) handleStats(w http.ResponseWriter, r *http.Request, u *User) {
	today, ok := todayParam(w, r)
	if !ok {
		return
	}
	if totals, ok := a.totals(w, r, u.ID); ok {
		writeJSON(w, http.StatusOK, BuildStats(totals, today))
	}
}

func (a *App) handleYears(w http.ResponseWriter, r *http.Request, u *User) {
	today, ok := todayParam(w, r)
	if !ok {
		return
	}
	if totals, ok := a.totals(w, r, u.ID); ok {
		writeJSON(w, http.StatusOK, Years(totals, today))
	}
}

func (a *App) handleListSessions(w http.ResponseWriter, r *http.Request, u *User) {
	date := r.URL.Query().Get("date")
	if _, err := time.Parse(dateLayout, date); err != nil {
		writeError(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
		return
	}
	sessions, err := a.store.OnDate(r.Context(), u.ID, date)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (a *App) handleAddSession(w http.ResponseWriter, r *http.Request, u *User) {
	var in struct {
		Date    string `json:"date"`
		Minutes int    `json:"minutes"`
		Kind    string `json:"kind"`
		Note    string `json:"note"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if _, err := time.Parse(dateLayout, in.Date); err != nil {
		writeError(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
		return
	}
	if in.Minutes < 1 || in.Minutes > 24*60 {
		writeError(w, http.StatusBadRequest, "minutes must be between 1 and 1440")
		return
	}
	if in.Kind != "pomodoro" {
		in.Kind = "manual"
	}
	sess, err := a.store.Add(r.Context(), u.ID, Session{
		Date: in.Date, Minutes: in.Minutes, Kind: in.Kind, Note: cleanText(in.Note, 200),
	})
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sess)
}

func (a *App) handleDeleteSession(w http.ResponseWriter, r *http.Request, u *User) {
	found, err := a.store.Delete(r.Context(), u.ID, r.PathValue("id"))
	switch {
	case err != nil:
		serverError(w, err)
	case !found:
		writeError(w, http.StatusNotFound, "session not found")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func (a *App) totals(w http.ResponseWriter, r *http.Request, userID string) (map[string]DayTotal, bool) {
	totals, err := a.store.DailyTotals(r.Context(), userID)
	if err != nil {
		serverError(w, err)
		return nil, false
	}
	return totals, true
}

// todayParam reads the client's local date so day boundaries match the user's timezone.
func todayParam(w http.ResponseWriter, r *http.Request) (time.Time, bool) {
	raw := r.URL.Query().Get("today")
	if raw == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), true
	}
	t, err := time.Parse(dateLayout, raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "today must be YYYY-MM-DD")
		return time.Time{}, false
	}
	return t, true
}

// --- middleware ---

type ctxKey int

const userKey ctxKey = iota

// requireUser resolves the session cookie and rejects anonymous requests.
func (a *App) requireUser(next func(http.ResponseWriter, *http.Request, *User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := a.currentUser(r)
		if err != nil {
			serverError(w, err)
			return
		}
		if u == nil {
			writeError(w, http.StatusUnauthorized, "sign in to continue")
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userKey, u)), u)
	}
}

func (a *App) currentUser(r *http.Request) (*User, error) {
	c, err := r.Cookie(a.cookieName())
	if err != nil || c.Value == "" || len(c.Value) > 100 {
		return nil, nil
	}
	return a.store.UserForSession(r.Context(), c.Value)
}

func (a *App) rateLimitAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/health" && !a.allow(w, a.lim.api, a.clientIP(r)) {
			return
		}
		next.ServeHTTP(w, r)
	})
}

// allow applies a limiter, writing a 429 with Retry-After when it's exhausted.
func (a *App) allow(w http.ResponseWriter, l *Limiter, key string) bool {
	ok, wait := l.Allow(key)
	if !ok {
		w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(wait.Seconds()))))
		writeError(w, http.StatusTooManyRequests, "too many requests — please slow down and try again shortly")
	}
	return ok
}

func (a *App) clientIP(r *http.Request) string {
	if h := a.cfg.ClientIPHeader; h != "" {
		v := strings.TrimSpace(strings.Split(r.Header.Get(h), ",")[0])
		if ip := net.ParseIP(v); ip != nil {
			return ip.String()
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Cache-Control", "no-store")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

func recoverPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				if v == http.ErrAbortHandler {
					panic(v)
				}
				log.Printf("panic: %v\n%s", v, debug.Stack())
				writeError(w, http.StatusInternalServerError, "something went wrong")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// --- JSON helpers ---

// decodeJSON reads one JSON object (max 16 KB) with no unknown fields.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		writeError(w, http.StatusUnsupportedMediaType, "expected application/json")
		return false
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// serverError logs the real cause and returns a generic message to the client.
func serverError(w http.ResponseWriter, err error) {
	log.Printf("internal error: %v", err)
	writeError(w, http.StatusInternalServerError, "something went wrong")
}
