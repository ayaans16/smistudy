package main

import (
	"net/http"
	"testing"
)

// publicUser signs up a user and turns their profile public.
func (e *testEnv) publicUser(t *testing.T, name string) *client {
	t.Helper()
	c := e.client(t)
	c.signUp(name+"@example.com", name, "correct horse battery")
	c.do("PATCH", "/api/me", `{"profilePublic":true}`)
	return c
}

func usernames(list []map[string]any) []string {
	out := []string{}
	for _, u := range list {
		out = append(out, u["username"].(string))
	}
	return out
}

func TestFollowRules(t *testing.T) {
	env := newTestEnv(t)
	ana := env.publicUser(t, "ana")
	env.publicUser(t, "bea")
	cal := env.client(t) // private profile
	cal.signUp("cal@example.com", "cal", "correct horse battery")

	if code, body := cal.do("POST", "/api/users/ana/follow", ""); code != http.StatusForbidden || body["code"] != "profile_private" {
		t.Errorf("private user following = %d %v, want 403 profile_private", code, body)
	}
	if code, _ := ana.do("POST", "/api/users/cal/follow", ""); code != http.StatusNotFound {
		t.Errorf("following a private profile = %d, want 404", code)
	}
	if code, _ := ana.do("POST", "/api/users/nobody/follow", ""); code != http.StatusNotFound {
		t.Errorf("following a missing user = %d, want 404", code)
	}
	if code, _ := ana.do("POST", "/api/users/ana/follow", ""); code != http.StatusBadRequest {
		t.Errorf("following yourself = %d, want 400", code)
	}
	if code, _ := env.client(t).do("POST", "/api/users/ana/follow", ""); code != http.StatusUnauthorized {
		t.Errorf("anonymous follow = %d, want 401", code)
	}

	for i := 0; i < 2; i++ { // following twice is harmless
		if code, _ := ana.do("POST", "/api/users/BEA/follow", ""); code != http.StatusNoContent {
			t.Fatalf("follow = %d", code)
		}
	}
	_, profile := env.client(t).do("GET", "/api/users/bea", "")
	if profile["followers"] != 1.0 || profile["following"] != 0.0 {
		t.Errorf("bea's counts = %v / %v", profile["followers"], profile["following"])
	}
	_, viewed := ana.do("GET", "/api/users/bea", "")
	if viewer, _ := viewed["viewer"].(map[string]any); viewer["following"] != true || viewer["followsYou"] != false {
		t.Errorf("viewer relation = %v", viewed["viewer"])
	}

	ana.do("DELETE", "/api/users/bea/follow", "")
	if _, list := env.client(t).doList("GET", "/api/users/bea/followers"); len(list) != 0 {
		t.Errorf("followers after unfollow = %v", list)
	}
}

func TestGoingPrivateHidesYouFromLists(t *testing.T) {
	env := newTestEnv(t)
	ana := env.publicUser(t, "ana")
	bea := env.publicUser(t, "bea")
	ana.do("POST", "/api/users/bea/follow", "")
	bea.do("POST", "/api/users/ana/follow", "")

	ana.do("PATCH", "/api/me", `{"profilePublic":false}`)
	visitor := env.client(t)
	if _, list := visitor.doList("GET", "/api/users/bea/followers"); len(list) != 0 {
		t.Errorf("private ana still listed as bea's follower: %v", usernames(list))
	}
	if _, list := visitor.doList("GET", "/api/users/bea/following"); len(list) != 0 {
		t.Errorf("private ana still listed in bea's following: %v", usernames(list))
	}
	if _, p := visitor.do("GET", "/api/users/bea", ""); p["followers"] != 0.0 || p["following"] != 0.0 {
		t.Errorf("counts include a private user: %v / %v", p["followers"], p["following"])
	}
	if code, _ := visitor.do("GET", "/api/users/ana/followers", ""); code != http.StatusNotFound {
		t.Errorf("private ana's follower list = %d, want 404", code)
	}

	ana.do("PATCH", "/api/me", `{"profilePublic":true}`)
	if _, list := visitor.doList("GET", "/api/users/bea/followers"); len(list) != 1 {
		t.Errorf("going public again should restore the follow: %v", list)
	}
}

func TestBlocking(t *testing.T) {
	env := newTestEnv(t)
	ana := env.publicUser(t, "ana")
	bea := env.publicUser(t, "bea")
	ana.do("POST", "/api/users/bea/follow", "")
	bea.do("POST", "/api/users/ana/follow", "")

	if code, _ := ana.do("POST", "/api/users/bea/block", ""); code != http.StatusNoContent {
		t.Fatalf("block = %d", code)
	}
	if _, p := env.client(t).do("GET", "/api/users/ana", ""); p["followers"] != 0.0 || p["following"] != 0.0 {
		t.Errorf("block should remove follows both ways: %v / %v", p["followers"], p["following"])
	}
	// Neither can follow the other, and bea sees a plain "not found" rather than "blocked".
	if code, body := bea.do("POST", "/api/users/ana/follow", ""); code != http.StatusNotFound || body["error"] != "profile not found" {
		t.Errorf("blocked user following = %d %v", code, body)
	}
	if code, _ := ana.do("POST", "/api/users/bea/follow", ""); code != http.StatusNotFound {
		t.Errorf("blocker following = %d, want 404", code)
	}
	if _, list := ana.doList("GET", "/api/me/blocked"); len(list) != 1 || list[0]["username"] != "bea" {
		t.Errorf("blocked list = %v", list)
	}
	if _, list := bea.doList("GET", "/api/me/blocked"); len(list) != 0 {
		t.Errorf("bea shouldn't see who blocked her: %v", list)
	}

	ana.do("DELETE", "/api/users/bea/block", "")
	if code, _ := bea.do("POST", "/api/users/ana/follow", ""); code != http.StatusNoContent {
		t.Errorf("follow after unblock = %d", code)
	}
}

func TestRemoveFollowerAndFriendsLeaderboard(t *testing.T) {
	env := newTestEnv(t)
	ana := env.publicUser(t, "ana")
	bea := env.publicUser(t, "bea")
	cal := env.publicUser(t, "cal")
	bea.do("POST", "/api/sessions", `{"date":"2026-10-02","minutes":30,"kind":"manual"}`)
	cal.do("POST", "/api/sessions", `{"date":"2026-10-02","minutes":90,"kind":"manual"}`)
	ana.do("POST", "/api/users/bea/follow", "")
	ana.do("POST", "/api/users/cal/follow", "")
	cal.do("POST", "/api/users/ana/follow", "")

	_, board := ana.doList("GET", "/api/me/following?today=2026-10-02")
	if got := usernames(board); len(got) != 2 || got[0] != "cal" || board[0]["weekMinutes"] != 90.0 {
		t.Errorf("leaderboard = %v", board)
	}

	_, followers := ana.doList("GET", "/api/me/followers")
	if len(followers) != 1 || followers[0]["username"] != "cal" || followers[0]["followingBack"] != true {
		t.Errorf("followers = %v", followers)
	}
	if code, _ := ana.do("DELETE", "/api/me/followers/cal", ""); code != http.StatusNoContent {
		t.Errorf("remove follower = %d", code)
	}
	if _, followers := ana.doList("GET", "/api/me/followers"); len(followers) != 0 {
		t.Errorf("follower not removed: %v", followers)
	}

	_, body := ana.do("GET", "/api/me/export", "")
	if f, _ := body["following"].([]any); len(f) != 2 {
		t.Errorf("export following = %v", body["following"])
	}
}
