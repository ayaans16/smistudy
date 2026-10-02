package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	oauthCookie     = "smistudy_oauth"
	googleUserInfo  = "https://openidconnect.googleapis.com/v1/userinfo"
	oauthCookiePath = "/api/auth/google"
)

func newGoogleConfig(cfg Config) *oauth2.Config {
	if !cfg.GoogleEnabled() {
		return nil
	}
	return &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  cfg.PublicURL + "/api/auth/google/callback",
		Scopes:       []string{"openid", "email", "profile"},
	}
}

// handleGoogleStart redirects to Google with a random state (CSRF) and a PKCE challenge,
// remembering both in a short-lived HttpOnly cookie.
func (a *App) handleGoogleStart(w http.ResponseWriter, r *http.Request) {
	if a.google == nil {
		writeError(w, http.StatusNotFound, "Google sign-in isn't enabled")
		return
	}
	if !a.allow(w, a.lim.oauth, a.clientIP(r)) {
		return
	}
	state := randomToken()
	verifier := oauth2.GenerateVerifier()
	http.SetCookie(w, &http.Cookie{
		Name: oauthCookie, Value: state + "." + verifier, Path: oauthCookiePath, MaxAge: 600,
		HttpOnly: true, Secure: a.cfg.SecureCookies(), SameSite: http.SameSiteLaxMode,
	})
	url := a.google.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("prompt", "select_account"))
	http.Redirect(w, r, url, http.StatusFound)
}

func (a *App) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	fail := func(reason string, err error) {
		if err != nil {
			log.Printf("google sign-in: %s: %v", reason, err)
		}
		http.Redirect(w, r, a.cfg.PublicURL+"/login?error=google", http.StatusFound)
	}
	if a.google == nil {
		fail("disabled", nil)
		return
	}
	if !a.allow(w, a.lim.oauth, a.clientIP(r)) {
		return
	}
	c, err := r.Cookie(oauthCookie)
	http.SetCookie(w, &http.Cookie{Name: oauthCookie, Path: oauthCookiePath, MaxAge: -1,
		HttpOnly: true, Secure: a.cfg.SecureCookies(), SameSite: http.SameSiteLaxMode})
	if err != nil {
		fail("missing state cookie", nil)
		return
	}
	state, verifier, _ := strings.Cut(c.Value, ".")
	got := r.URL.Query().Get("state")
	if state == "" || subtle.ConstantTimeCompare([]byte(state), []byte(got)) != 1 {
		fail("state mismatch", nil)
		return
	}
	if r.URL.Query().Get("error") != "" {
		fail("user cancelled", nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	tok, err := a.google.Exchange(ctx, r.URL.Query().Get("code"), oauth2.VerifierOption(verifier))
	if err != nil {
		fail("exchange", err)
		return
	}
	info, err := fetchGoogleUser(ctx, a.google.Client(ctx, tok))
	if err != nil {
		fail("userinfo", err)
		return
	}
	if info.Sub == "" || !info.EmailVerified {
		fail("unverified Google email", nil)
		return
	}
	email, err := normalizeEmail(info.Email)
	if err != nil {
		fail("bad email", err)
		return
	}

	u, err := a.findOrCreateGoogleUser(r.Context(), info, email)
	if err != nil {
		fail("account", err)
		return
	}
	token, expires, err := a.store.CreateAuthSession(r.Context(), u.ID, r.UserAgent())
	if err != nil {
		fail("session", err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: a.cookieName(), Value: token, Path: "/", Expires: expires,
		HttpOnly: true, Secure: a.cfg.SecureCookies(), SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, a.cfg.PublicURL+"/", http.StatusFound)
}

type googleUser struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
}

func fetchGoogleUser(ctx context.Context, client *http.Client) (googleUser, error) {
	var info googleUser
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfo, nil)
	if err != nil {
		return info, err
	}
	res, err := client.Do(req)
	if err != nil {
		return info, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return info, fmt.Errorf("userinfo: %s", res.Status)
	}
	err = json.NewDecoder(io.LimitReader(res.Body, 64<<10)).Decode(&info)
	return info, err
}

// findOrCreateGoogleUser signs in a linked account, links Google to an existing account
// with the same (Google-verified) email, or creates a new account.
func (a *App) findOrCreateGoogleUser(ctx context.Context, info googleUser, email string) (*User, error) {
	if u, err := a.store.UserByGoogleSub(ctx, info.Sub); err != nil || u != nil {
		return u, err
	}
	if u, err := a.store.UserByEmail(ctx, email); err != nil {
		return nil, err
	} else if u != nil {
		if u.GoogleSub != "" {
			return nil, errors.New("email belongs to an account linked to a different Google account")
		}
		if err := a.store.LinkGoogle(ctx, u, info.Sub); err != nil {
			return nil, err
		}
		return a.store.UserByID(ctx, u.ID)
	}

	base := usernameFromEmail(email)
	for attempt := 0; attempt < 8; attempt++ {
		name := base
		if attempt > 0 {
			n, _ := rand.Int(rand.Reader, big.NewInt(9000))
			name = fmt.Sprintf("%s%d", base[:min(len(base), 16)], n.Int64()+1000)
		}
		u := &User{
			Email: email, EmailVerified: true, GoogleSub: info.Sub,
			Username: name, DisplayName: cleanText(info.Name, 50),
		}
		if u.DisplayName == "" {
			u.DisplayName = name
		}
		switch err := a.store.CreateUser(ctx, u); {
		case err == nil:
			return u, nil
		case errors.Is(err, ErrUsernameTaken):
			continue
		default:
			return nil, err
		}
	}
	return nil, errors.New("couldn't pick a free username")
}

// usernameFromEmail turns "Ayaan.S+x@gmail.com" into "ayaans".
func usernameFromEmail(email string) string {
	local, _, _ := strings.Cut(email, "@")
	local, _, _ = strings.Cut(local, "+")
	var b strings.Builder
	for _, r := range strings.ToLower(local) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	name := b.String()
	if len(name) > 20 {
		name = name[:20]
	}
	if len(name) < 3 || reservedUsernames[name] {
		name = "studier"
	}
	return name
}
