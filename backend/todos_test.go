package main

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestTodoLifecycle(t *testing.T) {
	env := newTestEnv(t)
	c := env.client(t)
	c.signUp("ana@example.com", "ana", "correct horse battery")

	_, first := c.do("POST", "/api/todos", `{"text":"  read chapter 4  "}`)
	_, second := c.do("POST", "/api/todos", `{"text":"flashcards"}`)
	if first["text"] != "read chapter 4" || first["done"] != false {
		t.Fatalf("created = %v", first)
	}
	id := first["id"].(string)

	code, done := c.do("PATCH", "/api/todos/"+id, `{"done":true}`)
	if code != http.StatusOK || done["done"] != true || done["doneAt"] == nil {
		t.Fatalf("mark done = %d %v", code, done)
	}
	// Completed items sink below open ones.
	_, list := c.doList("GET", "/api/todos")
	if len(list) != 2 || list[0]["id"] != second["id"] || list[1]["id"] != id {
		t.Errorf("order = %v", list)
	}

	if _, renamed := c.do("PATCH", "/api/todos/"+id, `{"text":"read chapter 5"}`); renamed["text"] != "read chapter 5" || renamed["done"] != true {
		t.Errorf("rename should keep done state: %v", renamed)
	}
	if _, undone := c.do("PATCH", "/api/todos/"+id, `{"done":false}`); undone["done"] != false || undone["doneAt"] != nil {
		t.Errorf("unchecking = %v", undone)
	}

	c.do("PATCH", "/api/todos/"+id, `{"done":true}`)
	if _, res := c.do("POST", "/api/todos/clear-done", ""); res["cleared"] != 1.0 {
		t.Errorf("clear-done = %v", res)
	}
	if code, _ := c.do("DELETE", "/api/todos/"+second["id"].(string), ""); code != http.StatusNoContent {
		t.Errorf("delete = %d", code)
	}
	if _, list := c.doList("GET", "/api/todos"); len(list) != 0 {
		t.Errorf("list should be empty: %v", list)
	}
}

func TestTodosArePrivate(t *testing.T) {
	env := newTestEnv(t)
	ana, bea := env.client(t), env.client(t)
	ana.signUp("ana@example.com", "ana", "correct horse battery")
	bea.signUp("bea@example.com", "bea", "another long password")
	_, todo := ana.do("POST", "/api/todos", `{"text":"ana's secret plan"}`)
	id := todo["id"].(string)

	if _, list := bea.doList("GET", "/api/todos"); len(list) != 0 {
		t.Errorf("bea sees ana's to-dos: %v", list)
	}
	if code, _ := bea.do("PATCH", "/api/todos/"+id, `{"done":true}`); code != http.StatusNotFound {
		t.Errorf("bea editing ana's to-do = %d, want 404", code)
	}
	if code, _ := bea.do("DELETE", "/api/todos/"+id, ""); code != http.StatusNotFound {
		t.Errorf("bea deleting ana's to-do = %d, want 404", code)
	}
	bea.do("POST", "/api/todos/clear-done", "")
	if _, list := ana.doList("GET", "/api/todos"); len(list) != 1 || list[0]["done"] != false {
		t.Errorf("ana's to-do was affected: %v", list)
	}
	if code, _ := env.client(t).do("GET", "/api/todos", ""); code != http.StatusUnauthorized {
		t.Errorf("anonymous list = %d, want 401", code)
	}
	// Never exposed on public profiles.
	ana.do("PATCH", "/api/me", `{"profilePublic":true}`)
	_, _, card := getCard(t, env, "/api/users/ana/card.svg")
	_, profile := env.client(t).do("GET", "/api/users/ana", "")
	if strings.Contains(card, "secret plan") || profile["todos"] != nil {
		t.Error("to-dos leaked onto the public profile")
	}
}

func TestTodoValidationAndLimit(t *testing.T) {
	env := newTestEnv(t)
	c := env.client(t)
	c.signUp("ana@example.com", "ana", "correct horse battery")

	for _, body := range []string{`{"text":""}`, `{"text":"   "}`, `{"text":"x","done":true}`} {
		if code, _ := c.do("POST", "/api/todos", body); code != http.StatusBadRequest {
			t.Errorf("POST %s = %d, want 400", body, code)
		}
	}
	_, long := c.do("POST", "/api/todos", `{"text":"`+strings.Repeat("a", 300)+`"}`)
	if n := len([]rune(long["text"].(string))); n != 200 {
		t.Errorf("long text stored with %d chars, want 200", n)
	}

	u, _ := env.store.UserByEmail(context.Background(), "ana@example.com")
	for i := 1; i < maxTodos; i++ {
		if _, err := env.store.AddTodo(context.Background(), u.ID, "item"); err != nil {
			t.Fatal(err)
		}
	}
	if code, _ := c.do("POST", "/api/todos", `{"text":"one too many"}`); code != http.StatusConflict {
		t.Errorf("todo #%d = %d, want 409", maxTodos+1, code)
	}
}
