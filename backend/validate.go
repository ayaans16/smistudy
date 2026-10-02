package main

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9_]{3,20}$`)

// Usernames that would collide with routes or look official.
var reservedUsernames = map[string]bool{
	"admin": true, "api": true, "about": true, "account": true, "forgot": true, "help": true,
	"login": true, "logout": true, "me": true, "reset": true, "root": true, "settings": true,
	"signup": true, "smistudy": true, "static": true, "support": true, "u": true, "verify": true, "www": true,
}

func normalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email || len(email) > 254 {
		return "", errors.New("enter a valid email address")
	}
	return email, nil
}

func normalizeUsername(raw string) (string, error) {
	u := strings.ToLower(strings.TrimSpace(raw))
	if !usernamePattern.MatchString(u) {
		return "", errors.New("username must be 3–20 characters: letters, numbers or _")
	}
	if reservedUsernames[u] {
		return "", errors.New("that username isn't available")
	}
	return u, nil
}

func validatePassword(pw, email, username string) error {
	n := utf8.RuneCountInString(pw)
	switch {
	case n < 10:
		return errors.New("password must be at least 10 characters")
	case n > 128:
		return errors.New("password must be at most 128 characters")
	case strings.EqualFold(pw, email) || strings.EqualFold(pw, username):
		return errors.New("password can't be your email or username")
	}
	return nil
}

// cleanText trims, strips control characters and caps length.
func cleanText(s string, maxRunes int) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, strings.TrimSpace(s))
	if utf8.RuneCountInString(s) > maxRunes {
		s = string([]rune(s)[:maxRunes])
	}
	return s
}
