package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var errBadFilter = errors.New("filter must be \"last\" or a year")

func main() {
	addr := flag.String("addr", envOr("SMISTUDY_ADDR", ":8080"), "listen address")
	dataPath := flag.String("data", envOr("SMISTUDY_DATA", "data/sessions.json"), "path to the sessions JSON file")
	flag.Parse()

	store, err := OpenStore(*dataPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}

	log.Printf("smistudy api listening on %s (data: %s)", *addr, *dataPath)
	log.Fatal(http.ListenAndServe(*addr, withCORS(NewServer(store))))
}

func NewServer(store *Store) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// GET /api/contributions?filter=last|2026&today=2026-10-02
	mux.HandleFunc("GET /api/contributions", func(w http.ResponseWriter, r *http.Request) {
		today, ok := todayParam(w, r)
		if !ok {
			return
		}
		from, to, err := RangeFor(r.URL.Query().Get("filter"), today)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, BuildCalendar(store.DailyTotals(), from, to))
	})

	mux.HandleFunc("GET /api/stats", func(w http.ResponseWriter, r *http.Request) {
		today, ok := todayParam(w, r)
		if !ok {
			return
		}
		writeJSON(w, http.StatusOK, BuildStats(store.DailyTotals(), today))
	})

	mux.HandleFunc("GET /api/years", func(w http.ResponseWriter, r *http.Request) {
		today, ok := todayParam(w, r)
		if !ok {
			return
		}
		writeJSON(w, http.StatusOK, Years(store.DailyTotals(), today))
	})

	// GET /api/sessions?date=2026-10-02
	mux.HandleFunc("GET /api/sessions", func(w http.ResponseWriter, r *http.Request) {
		date := r.URL.Query().Get("date")
		if _, err := time.Parse(dateLayout, date); err != nil {
			writeError(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
			return
		}
		writeJSON(w, http.StatusOK, store.OnDate(date))
	})

	mux.HandleFunc("POST /api/sessions", func(w http.ResponseWriter, r *http.Request) {
		var in Session
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
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
		in.Note = strings.TrimSpace(in.Note)
		if len(in.Note) > 200 {
			in.Note = in.Note[:200]
		}
		sess, err := store.Add(in)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not save session")
			return
		}
		writeJSON(w, http.StatusCreated, sess)
	})

	mux.HandleFunc("DELETE /api/sessions/{id}", func(w http.ResponseWriter, r *http.Request) {
		found, err := store.Delete(r.PathValue("id"))
		switch {
		case err != nil:
			writeError(w, http.StatusInternalServerError, "could not delete session")
		case !found:
			writeError(w, http.StatusNotFound, "session not found")
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	return mux
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func withCORS(next http.Handler) http.Handler {
	origin := envOr("SMISTUDY_ORIGIN", "http://localhost:3000")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
