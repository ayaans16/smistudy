package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	Addr      string
	DBPath    string
	LegacyDB  string
	PublicURL string // e.g. https://smistudy.ca — used for links, OAuth redirects and CSRF checks

	// Header holding the real client IP, set by the proxy in front (e.g. CF-Connecting-IP).
	// Empty means use the TCP peer address. Only set this if the API is unreachable except
	// through that proxy, or clients could spoof it to dodge per-IP rate limits.
	ClientIPHeader string

	GoogleClientID     string
	GoogleClientSecret string

	ResendAPIKey string // empty = print emails to the log instead of sending
	MailFrom     string
}

func LoadConfig() (Config, error) {
	c := Config{
		Addr:               envOr("SMISTUDY_ADDR", "127.0.0.1:8080"),
		DBPath:             envOr("SMISTUDY_DB", "data/smistudy.db"),
		LegacyDB:           envOr("SMISTUDY_DATA", "data/sessions.json"),
		PublicURL:          strings.TrimRight(envOr("SMISTUDY_PUBLIC_URL", "http://localhost:3000"), "/"),
		ClientIPHeader:     os.Getenv("SMISTUDY_CLIENT_IP_HEADER"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		ResendAPIKey:       os.Getenv("RESEND_API_KEY"),
		MailFrom:           envOr("SMISTUDY_MAIL_FROM", "smistudy <noreply@smistudy.ca>"),
	}
	u, err := url.Parse(c.PublicURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.Path != "" {
		return c, fmt.Errorf("SMISTUDY_PUBLIC_URL must be an origin like https://smistudy.ca, got %q", c.PublicURL)
	}
	return c, nil
}

// SecureCookies is true in production; browsers only send Secure cookies over HTTPS.
func (c Config) SecureCookies() bool { return strings.HasPrefix(c.PublicURL, "https://") }

func (c Config) GoogleEnabled() bool { return c.GoogleClientID != "" && c.GoogleClientSecret != "" }

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
