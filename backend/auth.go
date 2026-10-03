package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

func (a *App) authRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/auth/providers", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"google": a.google != nil})
	})
	mux.HandleFunc("POST /api/auth/signup", a.handleSignup)
	mux.HandleFunc("POST /api/auth/login", a.handleLogin)
	mux.HandleFunc("POST /api/auth/logout", a.handleLogout)
	mux.HandleFunc("POST /api/auth/verify", a.handleVerify)
	mux.HandleFunc("POST /api/auth/resend-verification", a.handleResendVerification)
	mux.HandleFunc("POST /api/auth/forgot", a.handleForgot)
	mux.HandleFunc("POST /api/auth/reset", a.handleReset)
	mux.HandleFunc("GET /api/auth/google/start", a.handleGoogleStart)
	mux.HandleFunc("GET /api/auth/google/callback", a.handleGoogleCallback)

	mux.HandleFunc("GET /api/me", a.requireUser(a.handleMe))
	mux.HandleFunc("PATCH /api/me", a.requireUser(a.handleUpdateMe))
	mux.HandleFunc("POST /api/me/password", a.requireUser(a.handleChangePassword))
	mux.HandleFunc("DELETE /api/me", a.requireUser(a.handleDeleteMe))
	mux.HandleFunc("GET /api/me/export", a.requireUser(a.handleExport))
}

type meResponse struct {
	*User
	HasPassword bool `json:"hasPassword"`
	HasGoogle   bool `json:"hasGoogle"`
}

func me(u *User) meResponse {
	return meResponse{User: u, HasPassword: u.PasswordHash != "", HasGoogle: u.GoogleSub != ""}
}

// --- sign up / verify ---

func (a *App) handleSignup(w http.ResponseWriter, r *http.Request) {
	if !a.allow(w, a.lim.signup, a.clientIP(r)) {
		return
	}
	var in struct {
		Email, Password, Username string
		AcceptTerms               bool
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if !in.AcceptTerms {
		writeError(w, http.StatusBadRequest, "please agree to the Terms of Service and Privacy Policy")
		return
	}
	email, err := normalizeEmail(in.Email)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	username, err := normalizeUsername(in.Username)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validatePassword(in.Password, email, username); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	// Usernames are public (profile URLs), so saying one is taken leaks nothing.
	if taken, err := a.store.UserByUsername(r.Context(), username); err != nil {
		serverError(w, err)
		return
	} else if taken != nil {
		writeError(w, http.StatusConflict, "that username is taken")
		return
	}

	// Emails are private: an existing account gets a heads-up email, and the
	// response is identical either way so signup can't be used to probe emails.
	accepted := func() { writeJSON(w, http.StatusAccepted, map[string]string{"status": "check_email"}) }
	existing, err := a.store.UserByEmail(r.Context(), email)
	if err != nil {
		serverError(w, err)
		return
	}
	if existing != nil {
		a.emailAccountExists(email)
		accepted()
		return
	}

	hash, err := HashPassword(in.Password)
	if err != nil {
		serverError(w, err)
		return
	}
	u := &User{Email: email, PasswordHash: hash, Username: username, DisplayName: username}
	switch err := a.store.CreateUser(r.Context(), u); {
	case errors.Is(err, ErrUsernameTaken):
		writeError(w, http.StatusConflict, "that username is taken")
		return
	case errors.Is(err, ErrEmailTaken):
		a.emailAccountExists(email)
		accepted()
		return
	case err != nil:
		serverError(w, err)
		return
	}
	if err := a.emailVerification(r, u); err != nil {
		serverError(w, err)
		return
	}
	accepted()
}

func (a *App) handleVerify(w http.ResponseWriter, r *http.Request) {
	if !a.allow(w, a.lim.token, a.clientIP(r)) {
		return
	}
	var in struct{ Token string }
	if !decodeJSON(w, r, &in) {
		return
	}
	userID, err := a.store.ConsumeToken(r.Context(), in.Token, "verify")
	if err != nil {
		serverError(w, err)
		return
	}
	if userID == "" {
		writeError(w, http.StatusBadRequest, "this link is invalid or has expired")
		return
	}
	if err := a.store.MarkEmailVerified(r.Context(), userID); err != nil {
		serverError(w, err)
		return
	}
	a.startSession(w, r, userID)
}

func (a *App) handleResendVerification(w http.ResponseWriter, r *http.Request) {
	if !a.allow(w, a.lim.login, a.clientIP(r)) {
		return
	}
	var in struct{ Email string }
	if !decodeJSON(w, r, &in) {
		return
	}
	if email, err := normalizeEmail(in.Email); err == nil {
		if u, err := a.store.UserByEmail(r.Context(), email); err != nil {
			serverError(w, err)
			return
		} else if u != nil && !u.EmailVerified {
			if err := a.emailVerification(r, u); err != nil {
				serverError(w, err)
				return
			}
		}
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "check_email"})
}

// --- log in / out ---

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	if !a.allow(w, a.lim.login, a.clientIP(r)) {
		return
	}
	var in struct{ Email, Password string }
	if !decodeJSON(w, r, &in) {
		return
	}
	email, _ := normalizeEmail(in.Email)
	if email != "" && !a.allow(w, a.lim.loginEmail, email) {
		return
	}
	u, err := a.store.UserByEmail(r.Context(), email)
	if err != nil {
		serverError(w, err)
		return
	}
	wrong := func() { writeError(w, http.StatusUnauthorized, "incorrect email or password") }
	if u == nil || u.PasswordHash == "" {
		VerifyPassword(in.Password, dummyHash) // keep timing the same as a real check
		wrong()
		return
	}
	if wait := time.Until(u.LockedUntil); wait > 0 {
		writeError(w, http.StatusTooManyRequests,
			fmt.Sprintf("too many failed attempts — try again in %d min, or reset your password", int(wait.Minutes())+1))
		return
	}
	ok, err := VerifyPassword(in.Password, u.PasswordHash)
	if err != nil {
		serverError(w, err)
		return
	}
	if !ok {
		if err := a.store.RecordFailedLogin(r.Context(), u); err != nil {
			serverError(w, err)
			return
		}
		wrong()
		return
	}
	if err := a.store.ResetFailedLogins(r.Context(), u.ID); err != nil {
		serverError(w, err)
		return
	}
	if !u.EmailVerified {
		if err := a.emailVerification(r, u); err != nil {
			serverError(w, err)
			return
		}
		writeJSON(w, http.StatusForbidden, map[string]string{
			"error": "please confirm your email first — we've sent you a new link",
			"code":  "email_unverified",
		})
		return
	}
	a.startSession(w, r, u.ID)
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(a.cookieName()); err == nil {
		if err := a.store.DeleteAuthSession(r.Context(), c.Value); err != nil {
			serverError(w, err)
			return
		}
	}
	a.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// --- password reset ---

func (a *App) handleForgot(w http.ResponseWriter, r *http.Request) {
	if !a.allow(w, a.lim.login, a.clientIP(r)) {
		return
	}
	var in struct{ Email string }
	if !decodeJSON(w, r, &in) {
		return
	}
	if email, err := normalizeEmail(in.Email); err == nil {
		u, err := a.store.UserByEmail(r.Context(), email)
		if err != nil {
			serverError(w, err)
			return
		}
		if u != nil {
			if ok, _ := a.lim.emailSend.Allow(u.Email); ok {
				token, err := a.store.CreateToken(r.Context(), u.ID, "reset")
				if err != nil {
					serverError(w, err)
					return
				}
				sendAsync(a.mail, u.Email, "Reset your smistudy password", fmt.Sprintf(
					"Hi %s,\n\nReset your password here (link expires in 1 hour):\n%s/reset?token=%s\n\n"+
						"If you didn't ask for this, you can ignore this email.", u.Username, a.cfg.PublicURL, token))
			}
		}
	}
	// Same response whether or not the account exists.
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "check_email"})
}

func (a *App) handleReset(w http.ResponseWriter, r *http.Request) {
	if !a.allow(w, a.lim.token, a.clientIP(r)) {
		return
	}
	var in struct{ Token, Password string }
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := validatePassword(in.Password, "", ""); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	userID, err := a.store.ConsumeToken(r.Context(), in.Token, "reset")
	if err != nil {
		serverError(w, err)
		return
	}
	if userID == "" {
		writeError(w, http.StatusBadRequest, "this link is invalid or has expired")
		return
	}
	hash, err := HashPassword(in.Password)
	if err != nil {
		serverError(w, err)
		return
	}
	ctx := r.Context()
	// Using the link proves inbox ownership; sign out every other device.
	for _, step := range []func() error{
		func() error { return a.store.SetPassword(ctx, userID, hash) },
		func() error { return a.store.MarkEmailVerified(ctx, userID) },
		func() error { return a.store.DeleteOtherSessions(ctx, userID, "") },
	} {
		if err := step(); err != nil {
			serverError(w, err)
			return
		}
	}
	a.startSession(w, r, userID)
}

// --- account ---

func (a *App) handleMe(w http.ResponseWriter, r *http.Request, u *User) {
	writeJSON(w, http.StatusOK, me(u))
}

func (a *App) handleUpdateMe(w http.ResponseWriter, r *http.Request, u *User) {
	var in ProfileUpdate
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Username != nil {
		name, err := normalizeUsername(*in.Username)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		in.Username = &name
	}
	if in.DisplayName != nil {
		name := cleanText(*in.DisplayName, 50)
		in.DisplayName = &name
	}
	switch err := a.store.UpdateProfile(r.Context(), u.ID, in); {
	case errors.Is(err, ErrUsernameTaken):
		writeError(w, http.StatusConflict, "that username is taken")
		return
	case err != nil:
		serverError(w, err)
		return
	}
	updated, err := a.store.UserByID(r.Context(), u.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, me(updated))
}

func (a *App) handleChangePassword(w http.ResponseWriter, r *http.Request, u *User) {
	if !a.allow(w, a.lim.loginEmail, u.Email) {
		return
	}
	var in struct{ Current, New string }
	if !decodeJSON(w, r, &in) {
		return
	}
	if u.PasswordHash != "" {
		ok, err := VerifyPassword(in.Current, u.PasswordHash)
		if err != nil {
			serverError(w, err)
			return
		}
		if !ok {
			writeError(w, http.StatusForbidden, "current password is incorrect")
			return
		}
	}
	if err := validatePassword(in.New, u.Email, u.Username); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	hash, err := HashPassword(in.New)
	if err != nil {
		serverError(w, err)
		return
	}
	if err := a.store.SetPassword(r.Context(), u.ID, hash); err != nil {
		serverError(w, err)
		return
	}
	c, _ := r.Cookie(a.cookieName())
	if err := a.store.DeleteOtherSessions(r.Context(), u.ID, c.Value); err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleExport returns everything stored about the user as a JSON download
// (right of access under PIPEDA; data portability under Quebec's Law 25).
func (a *App) handleExport(w http.ResponseWriter, r *http.Request, u *User) {
	sessions, err := a.store.AllSessions(r.Context(), u.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="smistudy-data-`+u.Username+`.json"`)
	writeJSON(w, http.StatusOK, map[string]any{
		"exportedAt": time.Now().UTC().Format(time.RFC3339),
		"account": map[string]any{
			"email":         u.Email,
			"emailVerified": u.EmailVerified,
			"username":      u.Username,
			"displayName":   u.DisplayName,
			"profilePublic": u.ProfilePublic,
			"signInMethods": signInMethods(u),
			"createdAt":     u.CreatedAt.Format(time.RFC3339),
			"termsVersion":  u.TermsVersion,
		},
		"studySessions": sessions,
	})
}

func signInMethods(u *User) []string {
	methods := []string{}
	if u.PasswordHash != "" {
		methods = append(methods, "password")
	}
	if u.GoogleSub != "" {
		methods = append(methods, "google")
	}
	return methods
}

// handleDeleteMe needs the password, or (for Google-only accounts) the username typed out.
func (a *App) handleDeleteMe(w http.ResponseWriter, r *http.Request, u *User) {
	if !a.allow(w, a.lim.loginEmail, u.Email) {
		return
	}
	var in struct{ Password, Confirm string }
	if !decodeJSON(w, r, &in) {
		return
	}
	if u.PasswordHash != "" {
		ok, err := VerifyPassword(in.Password, u.PasswordHash)
		if err != nil {
			serverError(w, err)
			return
		}
		if !ok {
			writeError(w, http.StatusForbidden, "password is incorrect")
			return
		}
	} else if in.Confirm != u.Username {
		writeError(w, http.StatusForbidden, "type your username to confirm")
		return
	}
	if err := a.store.DeleteUser(r.Context(), u.ID); err != nil {
		serverError(w, err)
		return
	}
	a.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

func (a *App) emailVerification(r *http.Request, u *User) error {
	if ok, _ := a.lim.emailSend.Allow(u.Email); !ok {
		return nil // already sent a few recently; don't let anyone spam an inbox
	}
	token, err := a.store.CreateToken(r.Context(), u.ID, "verify")
	if err != nil {
		return err
	}
	sendAsync(a.mail, u.Email, "Confirm your smistudy email", fmt.Sprintf(
		"Hi %s, welcome to smistudy!\n\nConfirm your email to finish signing up (link expires in 24 hours):\n%s/verify?token=%s\n\n"+
			"If you didn't sign up, you can ignore this email.", u.Username, a.cfg.PublicURL, token))
	return nil
}

func (a *App) emailAccountExists(email string) {
	if ok, _ := a.lim.emailSend.Allow(email); !ok {
		return
	}
	sendAsync(a.mail, email, "You already have a smistudy account", fmt.Sprintf(
		"Someone (hopefully you) tried to sign up for smistudy with this email, but it already has an account.\n\n"+
			"Log in: %[1]s/login\nForgot your password? %[1]s/forgot\n\nIf this wasn't you, you can ignore this email.", a.cfg.PublicURL))
}

func (a *App) startSession(w http.ResponseWriter, r *http.Request, userID string) {
	token, expires, err := a.store.CreateAuthSession(r.Context(), userID, r.UserAgent())
	if err != nil {
		serverError(w, err)
		return
	}
	u, err := a.store.UserByID(r.Context(), userID)
	if err != nil || u == nil {
		serverError(w, fmt.Errorf("load user after login: %v", err))
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: a.cookieName(), Value: token, Path: "/", Expires: expires,
		HttpOnly: true, Secure: a.cfg.SecureCookies(), SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, me(u))
}

func (a *App) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: a.cookieName(), Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: a.cfg.SecureCookies(), SameSite: http.SameSiteLaxMode,
	})
}

// cookieName uses the __Host- prefix over HTTPS, which browsers only accept for
// Secure, host-only, Path=/ cookies — so subdomains can't plant or overwrite it.
func (a *App) cookieName() string {
	if a.cfg.SecureCookies() {
		return "__Host-smistudy_session"
	}
	return "smistudy_session"
}
