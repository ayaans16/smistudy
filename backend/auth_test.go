package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

// captureMailer records emails so tests can follow verification/reset links.
type captureMailer struct {
	mu   sync.Mutex
	sent []string
}

func (m *captureMailer) Send(_ context.Context, to, subject, text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, to+"\n"+subject+"\n"+text)
	return nil
}

var tokenInLink = regexp.MustCompile(`token=([A-Za-z0-9_-]+)`)

// lastToken waits for the async email to the address and returns its link token.
func (m *captureMailer) lastToken(t *testing.T, to string) string {
	t.Helper()
	for i := 0; i < 100; i++ {
		m.mu.Lock()
		for j := len(m.sent) - 1; j >= 0; j-- {
			if strings.HasPrefix(m.sent[j], to+"\n") {
				if match := tokenInLink.FindStringSubmatch(m.sent[j]); match != nil {
					m.mu.Unlock()
					return match[1]
				}
			}
		}
		m.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("no email with a token sent to %s", to)
	return ""
}

type testEnv struct {
	srv   *httptest.Server
	store *Store
	mail  *captureMailer
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	store := openTestStore(t)
	mail := &captureMailer{}
	cfg := Config{PublicURL: "http://smistudy.test"}
	srv := httptest.NewServer(NewApp(cfg, store, mail).Handler())
	t.Cleanup(srv.Close)
	return &testEnv{srv: srv, store: store, mail: mail}
}

// client is a browser-like client with its own cookie jar.
type client struct {
	t   *testing.T
	env *testEnv
	hc  *http.Client
}

func (e *testEnv) client(t *testing.T) *client {
	jar, _ := cookiejar.New(nil)
	return &client{t: t, env: e, hc: &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}}
}

func (c *client) do(method, path, body string, headers ...string) (int, map[string]any) {
	c.t.Helper()
	req, _ := http.NewRequest(method, c.env.srv.URL+path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	res, err := c.hc.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer res.Body.Close()
	var out map[string]any
	json.NewDecoder(res.Body).Decode(&out)
	return res.StatusCode, out
}

// doList is do for endpoints that return a JSON array.
func (c *client) doList(method, path string) (int, []map[string]any) {
	c.t.Helper()
	req, _ := http.NewRequest(method, c.env.srv.URL+path, nil)
	res, err := c.hc.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer res.Body.Close()
	var out []map[string]any
	json.NewDecoder(res.Body).Decode(&out)
	return res.StatusCode, out
}

// signUp registers, follows the emailed verification link, and ends up signed in.
func (c *client) signUp(email, username, password string) {
	c.t.Helper()
	code, body := c.do("POST", "/api/auth/signup",
		`{"email":"`+email+`","username":"`+username+`","password":"`+password+`","acceptTerms":true}`)
	if code != http.StatusAccepted {
		c.t.Fatalf("signup %s: %d %v", email, code, body)
	}
	token := c.env.mail.lastToken(c.t, email)
	if code, body := c.do("POST", "/api/auth/verify", `{"token":"`+token+`"}`); code != http.StatusOK {
		c.t.Fatalf("verify: %d %v", code, body)
	}
}

func TestSignupVerifyLoginFlow(t *testing.T) {
	env := newTestEnv(t)
	c := env.client(t)

	if code, _ := c.do("GET", "/api/me", ""); code != http.StatusUnauthorized {
		t.Fatalf("anonymous /api/me = %d, want 401", code)
	}

	// Unverified accounts can't log in yet.
	c.do("POST", "/api/auth/signup", `{"email":"Ana@Example.com","username":"ana","password":"correct horse battery","acceptTerms":true}`)
	if code, body := c.do("POST", "/api/auth/login", `{"email":"ana@example.com","password":"correct horse battery"}`); code != http.StatusForbidden || body["code"] != "email_unverified" {
		t.Fatalf("unverified login = %d %v", code, body)
	}

	token := env.mail.lastToken(t, "ana@example.com")
	if code, body := c.do("POST", "/api/auth/verify", `{"token":"`+token+`"}`); code != http.StatusOK || body["username"] != "ana" {
		t.Fatalf("verify = %d %v", code, body)
	}
	if code, _ := c.do("POST", "/api/auth/verify", `{"token":"`+token+`"}`); code != http.StatusBadRequest {
		t.Error("verification token worked twice")
	}

	if code, _ := c.do("POST", "/api/sessions", `{"date":"2026-10-02","minutes":25,"kind":"pomodoro"}`); code != http.StatusCreated {
		t.Fatalf("log session = %d", code)
	}

	c.do("POST", "/api/auth/logout", "")
	if code, _ := c.do("GET", "/api/me", ""); code != http.StatusUnauthorized {
		t.Error("still signed in after logout")
	}

	if code, _ := c.do("POST", "/api/auth/login", `{"email":"ana@example.com","password":"wrong password!"}`); code != http.StatusUnauthorized {
		t.Errorf("wrong password = %d", code)
	}
	if code, body := c.do("POST", "/api/auth/login", `{"email":"ANA@example.com","password":"correct horse battery"}`); code != http.StatusOK {
		t.Fatalf("login = %d %v", code, body)
	}
	if code, body := c.do("GET", "/api/stats?today=2026-10-02", ""); code != http.StatusOK || body["todayMinutes"] != 25.0 {
		t.Errorf("stats after login = %d %v", code, body)
	}
}

func TestUsersCannotSeeOrTouchEachOthersData(t *testing.T) {
	env := newTestEnv(t)
	ana, bea := env.client(t), env.client(t)
	ana.signUp("ana@example.com", "ana", "correct horse battery")
	bea.signUp("bea@example.com", "bea", "another long password")

	_, created := ana.do("POST", "/api/sessions", `{"date":"2026-10-02","minutes":50,"kind":"manual","note":"secret"}`)
	id := created["id"].(string)

	if _, body := bea.do("GET", "/api/stats?today=2026-10-02", ""); body["totalMinutes"] != 0.0 {
		t.Errorf("bea sees ana's minutes: %v", body)
	}
	if code, _ := bea.do("DELETE", "/api/sessions/"+id, ""); code != http.StatusNotFound {
		t.Errorf("bea deleting ana's session = %d, want 404", code)
	}
	if _, body := ana.do("GET", "/api/stats?today=2026-10-02", ""); body["totalMinutes"] != 50.0 {
		t.Errorf("ana's session was affected: %v", body)
	}
}

func TestSignupDoesNotRevealExistingEmails(t *testing.T) {
	env := newTestEnv(t)
	env.client(t).signUp("ana@example.com", "ana", "correct horse battery")

	code, body := env.client(t).do("POST", "/api/auth/signup",
		`{"email":"ana@example.com","username":"someoneelse","password":"a different password","acceptTerms":true}`)
	if code != http.StatusAccepted || body["status"] != "check_email" {
		t.Errorf("signup with existing email = %d %v, want the same 202 as a new email", code, body)
	}
	if code, _ := env.client(t).do("POST", "/api/auth/forgot", `{"email":"nobody@example.com"}`); code != http.StatusAccepted {
		t.Errorf("forgot for unknown email = %d, want 202", code)
	}
}

func TestCrossSiteWritesAreRejected(t *testing.T) {
	env := newTestEnv(t)
	c := env.client(t)
	c.signUp("ana@example.com", "ana", "correct horse battery")

	code, _ := c.do("POST", "/api/sessions", `{"date":"2026-10-02","minutes":25,"kind":"manual"}`,
		"Sec-Fetch-Site", "cross-site", "Origin", "https://evil.example")
	if code != http.StatusForbidden {
		t.Errorf("cross-site POST = %d, want 403", code)
	}
	code, _ = c.do("POST", "/api/sessions", `{"date":"2026-10-02","minutes":25,"kind":"manual"}`,
		"Sec-Fetch-Site", "same-origin")
	if code != http.StatusCreated {
		t.Errorf("same-origin POST = %d, want 201", code)
	}
}

func TestAccountLocksAfterRepeatedFailures(t *testing.T) {
	env := newTestEnv(t)
	env.client(t).signUp("ana@example.com", "ana", "correct horse battery")
	u, _ := env.store.UserByEmail(context.Background(), "ana@example.com")

	// Simulate 5 failures directly (the per-account rate limit would otherwise kick in first).
	for i := 0; i < 5; i++ {
		u, _ = env.store.UserByEmail(context.Background(), "ana@example.com")
		env.store.RecordFailedLogin(context.Background(), u)
	}
	code, _ := env.client(t).do("POST", "/api/auth/login", `{"email":"ana@example.com","password":"correct horse battery"}`)
	if code != http.StatusTooManyRequests {
		t.Errorf("login while locked = %d, want 429 even with the right password", code)
	}
}

func TestLoginIsRateLimitedPerAccount(t *testing.T) {
	env := newTestEnv(t)
	env.client(t).signUp("ana@example.com", "ana", "correct horse battery")
	c := env.client(t)
	var last int
	for i := 0; i < 8; i++ {
		last, _ = c.do("POST", "/api/auth/login", `{"email":"ana@example.com","password":"guess `+string(rune('a'+i))+`xxxxxxxx"}`)
	}
	if last != http.StatusTooManyRequests {
		t.Errorf("8 rapid guesses: last status %d, want 429", last)
	}
}

func TestPasswordResetSignsOutOtherDevices(t *testing.T) {
	env := newTestEnv(t)
	laptop := env.client(t)
	laptop.signUp("ana@example.com", "ana", "correct horse battery")

	phone := env.client(t)
	phone.do("POST", "/api/auth/forgot", `{"email":"ana@example.com"}`)
	token := env.mail.lastToken(t, "ana@example.com")
	if code, body := phone.do("POST", "/api/auth/reset", `{"token":"`+token+`","password":"brand new password"}`); code != http.StatusOK {
		t.Fatalf("reset = %d %v", code, body)
	}
	if code, _ := laptop.do("GET", "/api/me", ""); code != http.StatusUnauthorized {
		t.Error("old session survived a password reset")
	}
	if code, _ := phone.do("GET", "/api/me", ""); code != http.StatusOK {
		t.Error("reset should sign in the device that used the link")
	}
	if code, _ := env.client(t).do("POST", "/api/auth/login", `{"email":"ana@example.com","password":"correct horse battery"}`); code != http.StatusUnauthorized {
		t.Error("old password still works")
	}
}

func TestInputValidation(t *testing.T) {
	env := newTestEnv(t)
	c := env.client(t)
	cases := []struct{ name, body string }{
		{"short password", `{"email":"a@example.com","username":"abc","password":"short","acceptTerms":true}`},
		{"bad email", `{"email":"not-an-email","username":"abc","password":"long enough pw","acceptTerms":true}`},
		{"sql-ish username", `{"email":"a@example.com","username":"x' OR 1=1--","password":"long enough pw","acceptTerms":true}`},
		{"reserved username", `{"email":"a@example.com","username":"admin","password":"long enough pw","acceptTerms":true}`},
		{"unknown field", `{"email":"a@example.com","username":"abc","password":"long enough pw","admin":true,"acceptTerms":true}`},
	}
	for _, tc := range cases {
		if code, _ := c.do("POST", "/api/auth/signup", tc.body); code != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400", tc.name, code)
		}
	}
}

func TestDeleteAccountRemovesData(t *testing.T) {
	env := newTestEnv(t)
	c := env.client(t)
	c.signUp("ana@example.com", "ana", "correct horse battery")
	c.do("POST", "/api/sessions", `{"date":"2026-10-02","minutes":25,"kind":"manual"}`)

	if code, _ := c.do("DELETE", "/api/me", `{"password":"wrong password!!"}`); code != http.StatusForbidden {
		t.Errorf("delete with wrong password = %d", code)
	}
	if code, _ := c.do("DELETE", "/api/me", `{"password":"correct horse battery"}`); code != http.StatusNoContent {
		t.Fatalf("delete = %d", code)
	}
	if u, _ := env.store.UserByEmail(context.Background(), "ana@example.com"); u != nil {
		t.Error("user still exists")
	}
	var n int
	env.store.db.QueryRow(`SELECT COUNT(*) FROM study_sessions`).Scan(&n)
	if n != 0 {
		t.Errorf("%d study sessions left after account deletion", n)
	}
}

func TestPasswordHashing(t *testing.T) {
	h, err := HashPassword("correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(h, "$argon2id$") || strings.Contains(h, "correct") {
		t.Errorf("unexpected hash %q", h)
	}
	if ok, _ := VerifyPassword("correct horse battery", h); !ok {
		t.Error("right password rejected")
	}
	if ok, _ := VerifyPassword("correct horse batterY", h); ok {
		t.Error("wrong password accepted")
	}
}

func TestUsernameFromEmail(t *testing.T) {
	cases := map[string]string{
		"Ayaan.S+study@gmail.com": "ayaans",
		"ab@x.com":                "studier",
		"admin@x.com":             "studier",
	}
	for email, want := range cases {
		if got := usernameFromEmail(strings.ToLower(email)); got != want {
			t.Errorf("usernameFromEmail(%q) = %q, want %q", email, got, want)
		}
	}
}

func TestPublicProfilesAreOptIn(t *testing.T) {
	env := newTestEnv(t)
	ana := env.client(t)
	ana.signUp("ana@example.com", "ana", "correct horse battery")
	ana.do("POST", "/api/sessions", `{"date":"2026-10-02","minutes":30,"kind":"manual","note":"private note"}`)
	visitor := env.client(t)

	for _, path := range []string{"/api/users/ana", "/api/users/ana/contributions", "/api/users/nobody"} {
		if code, _ := visitor.do("GET", path, ""); code != http.StatusNotFound {
			t.Errorf("GET %s while private = %d, want 404", path, code)
		}
	}

	ana.do("PATCH", "/api/me", `{"profilePublic":true}`)
	code, body := visitor.do("GET", "/api/users/ANA?today=2026-10-02", "")
	if code != http.StatusOK || body["username"] != "ana" {
		t.Fatalf("public profile = %d %v", code, body)
	}
	if _, leaked := body["email"]; leaked {
		t.Error("public profile exposes email")
	}
	if code, body := visitor.do("GET", "/api/users/ana/contributions?today=2026-10-02", ""); code != http.StatusOK || body["totalMinutes"] != 30.0 {
		t.Errorf("public contributions = %d %v", code, body)
	}
	// Day-by-day session details (with notes) stay private.
	if code, _ := visitor.do("GET", "/api/sessions?date=2026-10-02", ""); code != http.StatusUnauthorized {
		t.Errorf("anonymous session list = %d, want 401", code)
	}
}

func TestSignupRequiresAgreeingToTerms(t *testing.T) {
	env := newTestEnv(t)
	code, body := env.client(t).do("POST", "/api/auth/signup",
		`{"email":"ana@example.com","username":"ana","password":"correct horse battery"}`)
	if code != http.StatusBadRequest {
		t.Errorf("signup without acceptTerms = %d %v, want 400", code, body)
	}
	env.client(t).signUp("bea@example.com", "bea", "correct horse battery")
	u, _ := env.store.UserByEmail(context.Background(), "bea@example.com")
	if u.TermsVersion != TermsVersion {
		t.Errorf("terms version = %q, want %q", u.TermsVersion, TermsVersion)
	}
}

func TestExportContainsOnlyYourData(t *testing.T) {
	env := newTestEnv(t)
	ana, bea := env.client(t), env.client(t)
	ana.signUp("ana@example.com", "ana", "correct horse battery")
	bea.signUp("bea@example.com", "bea", "another long password")
	ana.do("POST", "/api/sessions", `{"date":"2026-10-02","minutes":30,"kind":"manual","note":"orgo"}`)
	bea.do("POST", "/api/sessions", `{"date":"2026-10-02","minutes":45,"kind":"manual","note":"bea's secret"}`)

	code, body := ana.do("GET", "/api/me/export", "")
	if code != http.StatusOK {
		t.Fatalf("export = %d", code)
	}
	account := body["account"].(map[string]any)
	sessions := body["studySessions"].([]any)
	if account["email"] != "ana@example.com" || len(sessions) != 1 || sessions[0].(map[string]any)["note"] != "orgo" {
		t.Errorf("export = %v", body)
	}
	raw, _ := json.Marshal(body)
	if strings.Contains(string(raw), "argon2") || strings.Contains(string(raw), "bea") {
		t.Errorf("export leaks a password hash or another user's data: %s", raw)
	}
	if code, _ := env.client(t).do("GET", "/api/me/export", ""); code != http.StatusUnauthorized {
		t.Errorf("anonymous export = %d, want 401", code)
	}
}
