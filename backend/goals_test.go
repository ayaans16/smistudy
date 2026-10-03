package main

import (
	"net/http"
	"testing"
)

func TestRewardGoalFlow(t *testing.T) {
	env := newTestEnv(t)
	c := env.client(t)
	c.signUp("ana@example.com", "ana", "correct horse battery")
	c.do("POST", "/api/sessions", `{"date":"2026-09-30","minutes":600,"kind":"manual"}`) // before the goal

	code, goal := c.do("POST", "/api/goals", `{"reward":"  new smiski blind box ","targetHours":1.5,"startDate":"2026-10-01"}`)
	if code != http.StatusCreated || goal["reward"] != "new smiski blind box" || goal["targetMinutes"] != 90.0 || goal["minutes"] != 0.0 {
		t.Fatalf("create = %d %v", code, goal)
	}
	id := goal["id"].(string)

	if code, _ := c.do("POST", "/api/goals/"+id+"/claim", ""); code != http.StatusConflict {
		t.Errorf("claiming an unreached goal = %d, want 409", code)
	}

	c.do("POST", "/api/sessions", `{"date":"2026-10-01","minutes":50,"kind":"pomodoro"}`)
	c.do("POST", "/api/sessions", `{"date":"2026-10-02","minutes":50,"kind":"pomodoro"}`)
	_, list := c.doList("GET", "/api/goals")
	if len(list) != 1 || list[0]["minutes"] != 100.0 || list[0]["reachedOn"] != "2026-10-02" {
		t.Fatalf("progress = %v", list)
	}

	if code, _ := c.do("POST", "/api/goals/"+id+"/claim", ""); code != http.StatusNoContent {
		t.Errorf("claim = %d", code)
	}
	_, list = c.doList("GET", "/api/goals")
	if list[0]["claimedAt"] == nil {
		t.Error("goal should be claimed")
	}

	_, body := c.do("GET", "/api/me/export", "")
	if goals, _ := body["rewardGoals"].([]any); len(goals) != 1 {
		t.Errorf("export goals = %v", body["rewardGoals"])
	}

	if code, _ := c.do("DELETE", "/api/goals/"+id, ""); code != http.StatusNoContent {
		t.Errorf("delete = %d", code)
	}
}

func TestRewardGoalsArePrivateAndValidated(t *testing.T) {
	env := newTestEnv(t)
	ana, bea := env.client(t), env.client(t)
	ana.signUp("ana@example.com", "ana", "correct horse battery")
	bea.signUp("bea@example.com", "bea", "another long password")
	_, goal := ana.do("POST", "/api/goals", `{"reward":"concert tickets","targetHours":10,"startDate":"2026-10-01"}`)
	id := goal["id"].(string)

	if _, list := bea.doList("GET", "/api/goals"); len(list) != 0 {
		t.Errorf("bea sees ana's goals: %v", list)
	}
	for _, req := range [][2]string{{"POST", "/api/goals/" + id + "/claim"}, {"DELETE", "/api/goals/" + id}} {
		if code, _ := bea.do(req[0], req[1], ""); code != http.StatusNotFound {
			t.Errorf("bea %s %s = %d, want 404", req[0], req[1], code)
		}
	}
	if code, _ := env.client(t).do("GET", "/api/goals", ""); code != http.StatusUnauthorized {
		t.Errorf("anonymous = %d, want 401", code)
	}

	bad := []string{
		`{"reward":"","targetHours":5,"startDate":"2026-10-01"}`,
		`{"reward":"x","targetHours":0.2,"startDate":"2026-10-01"}`,
		`{"reward":"x","targetHours":5000,"startDate":"2026-10-01"}`,
		`{"reward":"x","targetHours":5,"startDate":"tomorrow"}`,
		`{"reward":"x","targetHours":5,"startDate":"2999-01-01"}`,
	}
	for _, body := range bad {
		if code, _ := ana.do("POST", "/api/goals", body); code != http.StatusBadRequest {
			t.Errorf("POST %s = %d, want 400", body, code)
		}
	}
}
