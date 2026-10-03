package main

import (
	"errors"
	"net/http"
	"sort"
)

func (a *App) followRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/users/{username}/follow", a.requireUser(a.handleFollow))
	mux.HandleFunc("DELETE /api/users/{username}/follow", a.requireUser(a.handleUnfollow))
	mux.HandleFunc("POST /api/users/{username}/block", a.requireUser(a.handleBlock))
	mux.HandleFunc("DELETE /api/users/{username}/block", a.requireUser(a.handleUnblock))
	mux.HandleFunc("GET /api/users/{username}/followers", a.publicUser(a.handlePublicFollowers))
	mux.HandleFunc("GET /api/users/{username}/following", a.publicUser(a.handlePublicFollowing))

	mux.HandleFunc("GET /api/me/following", a.requireUser(a.handleMyFollowing))
	mux.HandleFunc("GET /api/me/followers", a.requireUser(a.handleMyFollowers))
	mux.HandleFunc("DELETE /api/me/followers/{username}", a.requireUser(a.handleRemoveFollower))
	mux.HandleFunc("GET /api/me/blocked", a.requireUser(a.handleMyBlocked))
}

// targetUser looks up the {username} in the path, writing a 404 if there's no such user.
func (a *App) targetUser(w http.ResponseWriter, r *http.Request) (*User, bool) {
	var u *User
	if name, err := normalizeUsername(r.PathValue("username")); err == nil {
		if u, err = a.store.UserByUsername(r.Context(), name); err != nil {
			serverError(w, err)
			return nil, false
		}
	}
	if u == nil {
		writeError(w, http.StatusNotFound, "profile not found")
		return nil, false
	}
	return u, true
}

// handleFollow follows a public profile. You need a public profile yourself, so every
// follower and following list only ever contains people who chose to be visible.
func (a *App) handleFollow(w http.ResponseWriter, r *http.Request, me *User) {
	if !a.allow(w, a.lim.follow, me.ID) {
		return
	}
	if !me.ProfilePublic {
		writeJSON(w, http.StatusForbidden, map[string]string{
			"error": "make your profile public in Settings to follow people",
			"code":  "profile_private",
		})
		return
	}
	target, ok := a.targetUser(w, r)
	if !ok {
		return
	}
	if target.ID == me.ID {
		writeError(w, http.StatusBadRequest, "you can't follow yourself")
		return
	}
	notFound := func() { writeError(w, http.StatusNotFound, "profile not found") }
	if !target.ProfilePublic {
		notFound()
		return
	}
	switch err := a.store.Follow(r.Context(), me.ID, target.ID); {
	case errors.Is(err, ErrBlocked):
		notFound() // same as a private profile, so a block isn't revealed
	case errors.Is(err, ErrFollowingTooMany):
		writeError(w, http.StatusConflict, "you're following the maximum of 1,000 people")
	case err != nil:
		serverError(w, err)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func (a *App) handleUnfollow(w http.ResponseWriter, r *http.Request, me *User) {
	target, ok := a.targetUser(w, r)
	if !ok {
		return
	}
	if _, err := a.store.Unfollow(r.Context(), me.ID, target.ID); err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleBlock(w http.ResponseWriter, r *http.Request, me *User) {
	if !a.allow(w, a.lim.follow, me.ID) {
		return
	}
	target, ok := a.targetUser(w, r)
	if !ok {
		return
	}
	if target.ID == me.ID {
		writeError(w, http.StatusBadRequest, "you can't block yourself")
		return
	}
	if err := a.store.Block(r.Context(), me.ID, target.ID); err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleUnblock(w http.ResponseWriter, r *http.Request, me *User) {
	target, ok := a.targetUser(w, r)
	if !ok {
		return
	}
	if _, err := a.store.Unblock(r.Context(), me.ID, target.ID); err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handlePublicFollowers(w http.ResponseWriter, r *http.Request, u *User) {
	list, err := a.store.Followers(r.Context(), u.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *App) handlePublicFollowing(w http.ResponseWriter, r *http.Request, u *User) {
	list, err := a.store.Following(r.Context(), u.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// FriendStats is someone you follow, with the numbers for the friends leaderboard.
type FriendStats struct {
	FollowUser
	WeekMinutes   int `json:"weekMinutes"`
	TotalMinutes  int `json:"totalMinutes"`
	CurrentStreak int `json:"currentStreak"`
}

// handleMyFollowing lists who you follow, ranked by study time this week.
func (a *App) handleMyFollowing(w http.ResponseWriter, r *http.Request, me *User) {
	today, ok := todayParam(w, r)
	if !ok {
		return
	}
	following, err := a.store.Following(r.Context(), me.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	out := make([]FriendStats, 0, len(following))
	for _, f := range following {
		totals, err := a.store.DailyTotals(r.Context(), f.ID)
		if err != nil {
			serverError(w, err)
			return
		}
		st := BuildStats(totals, today)
		out = append(out, FriendStats{FollowUser: f, WeekMinutes: st.WeekMinutes, TotalMinutes: st.TotalMinutes, CurrentStreak: st.CurrentStreak})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].WeekMinutes != out[j].WeekMinutes {
			return out[i].WeekMinutes > out[j].WeekMinutes
		}
		return out[i].Username < out[j].Username
	})
	writeJSON(w, http.StatusOK, out)
}

// handleMyFollowers lists your followers and whether you follow each one back.
func (a *App) handleMyFollowers(w http.ResponseWriter, r *http.Request, me *User) {
	followers, err := a.store.Followers(r.Context(), me.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	type follower struct {
		FollowUser
		FollowingBack bool `json:"followingBack"`
	}
	out := make([]follower, 0, len(followers))
	for _, f := range followers {
		back, err := a.store.IsFollowing(r.Context(), me.ID, f.ID)
		if err != nil {
			serverError(w, err)
			return
		}
		out = append(out, follower{FollowUser: f, FollowingBack: back})
	}
	writeJSON(w, http.StatusOK, out)
}

// handleRemoveFollower makes someone stop following you (they can follow again; block to prevent that).
func (a *App) handleRemoveFollower(w http.ResponseWriter, r *http.Request, me *User) {
	follower, ok := a.targetUser(w, r)
	if !ok {
		return
	}
	if _, err := a.store.Unfollow(r.Context(), follower.ID, me.ID); err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleMyBlocked(w http.ResponseWriter, r *http.Request, me *User) {
	list, err := a.store.Blocked(r.Context(), me.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}
