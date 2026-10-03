package main

import (
	"net/http"
	"time"
)

// Public profiles are opt-in. A private profile and a missing user get the same 404,
// so these routes can't be used to discover which usernames exist.
func (a *App) profileRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/users/{username}", a.publicUser(func(w http.ResponseWriter, r *http.Request, u *User) {
		today, ok := todayParam(w, r)
		if !ok {
			return
		}
		totals, ok := a.totals(w, r, u.ID)
		if !ok {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"username":    u.Username,
			"displayName": u.DisplayName,
			"joinedAt":    u.CreatedAt.Format(time.RFC3339),
			"stats":       BuildStats(totals, today),
		})
	}))
	mux.HandleFunc("GET /api/users/{username}/contributions", a.publicUser(a.handleContributions))
	mux.HandleFunc("GET /api/users/{username}/years", a.publicUser(a.handleYears))
	mux.HandleFunc("GET /api/users/{username}/card.svg", a.handleProfileCard)
}

func (a *App) publicUser(next func(http.ResponseWriter, *http.Request, *User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name, err := normalizeUsername(r.PathValue("username"))
		var u *User
		if err == nil {
			if u, err = a.store.UserByUsername(r.Context(), name); err != nil {
				serverError(w, err)
				return
			}
		}
		if u == nil || !u.ProfilePublic {
			writeError(w, http.StatusNotFound, "profile not found")
			return
		}
		next(w, r, u)
	}
}
