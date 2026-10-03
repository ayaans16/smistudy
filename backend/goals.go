package main

import (
	"errors"
	"math"
	"net/http"
	"time"
)

func (a *App) goalRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/goals", a.requireUser(a.handleListGoals))
	mux.HandleFunc("POST /api/goals", a.requireUser(a.handleAddGoal))
	mux.HandleFunc("POST /api/goals/{id}/claim", a.requireUser(a.handleClaimGoal))
	mux.HandleFunc("DELETE /api/goals/{id}", a.requireUser(a.handleDeleteGoal))
}

// goalsWithProgress loads a user's goals and fills in how far along each one is.
func (a *App) goalsWithProgress(r *http.Request, userID string) ([]Goal, error) {
	goals, err := a.store.Goals(r.Context(), userID)
	if err != nil {
		return nil, err
	}
	totals, err := a.store.DailyTotals(r.Context(), userID)
	if err != nil {
		return nil, err
	}
	for i := range goals {
		goals[i].Minutes, goals[i].ReachedOn = goalProgress(totals, goals[i].StartDate, goals[i].TargetMinutes)
	}
	return goals, nil
}

func (a *App) handleListGoals(w http.ResponseWriter, r *http.Request, u *User) {
	goals, err := a.goalsWithProgress(r, u.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, goals)
}

func (a *App) handleAddGoal(w http.ResponseWriter, r *http.Request, u *User) {
	var in struct {
		Reward      string
		TargetHours float64
		StartDate   string
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	reward := cleanText(in.Reward, 100)
	if reward == "" {
		writeError(w, http.StatusBadRequest, "what's the reward? give it a name")
		return
	}
	if math.IsNaN(in.TargetHours) || in.TargetHours < 0.5 || in.TargetHours > 1000 {
		writeError(w, http.StatusBadRequest, "the goal must be between 0.5 and 1000 hours")
		return
	}
	start, err := time.Parse(dateLayout, in.StartDate)
	if err != nil || start.Year() < 2000 || start.After(time.Now().AddDate(0, 0, 1)) {
		writeError(w, http.StatusBadRequest, "startDate must be a date (YYYY-MM-DD) that isn't in the future")
		return
	}
	g := &Goal{Reward: reward, TargetMinutes: int(math.Round(in.TargetHours * 60)), StartDate: in.StartDate}
	switch err := a.store.AddGoal(r.Context(), u.ID, g); {
	case errors.Is(err, ErrTooManyGoals):
		writeError(w, http.StatusConflict, "you have 50 goals already, delete an old one first")
		return
	case err != nil:
		serverError(w, err)
		return
	}
	totals, err := a.store.DailyTotals(r.Context(), u.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	g.Minutes, g.ReachedOn = goalProgress(totals, g.StartDate, g.TargetMinutes)
	writeJSON(w, http.StatusCreated, g)
}

// handleClaimGoal marks a reached goal's reward as claimed.
func (a *App) handleClaimGoal(w http.ResponseWriter, r *http.Request, u *User) {
	goals, err := a.goalsWithProgress(r, u.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	id := r.PathValue("id")
	for _, g := range goals {
		if g.ID != id {
			continue
		}
		if g.ReachedOn == "" {
			writeError(w, http.StatusConflict, "keep going, you haven't reached this goal yet")
			return
		}
		if _, err := a.store.ClaimGoal(r.Context(), u.ID, id); err != nil {
			serverError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeError(w, http.StatusNotFound, "goal not found")
}

func (a *App) handleDeleteGoal(w http.ResponseWriter, r *http.Request, u *User) {
	found, err := a.store.DeleteGoal(r.Context(), u.ID, r.PathValue("id"))
	switch {
	case err != nil:
		serverError(w, err)
	case !found:
		writeError(w, http.StatusNotFound, "goal not found")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
