package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"time"
)

type Mailer interface {
	Send(ctx context.Context, to, subject, text string) error
}

// logMailer is for development: it prints emails (including links) to the log.
type logMailer struct{}

func (logMailer) Send(_ context.Context, to, subject, text string) error {
	log.Printf("email to %s — %s\n%s", to, subject, text)
	return nil
}

// resendMailer sends through the Resend HTTP API (https://resend.com).
type resendMailer struct {
	apiKey, from string
	client       *http.Client
}

func NewMailer(cfg Config) Mailer {
	if cfg.ResendAPIKey == "" {
		log.Print("RESEND_API_KEY not set: emails will be printed to the log, not sent")
		return logMailer{}
	}
	return &resendMailer{apiKey: cfg.ResendAPIKey, from: cfg.MailFrom, client: &http.Client{Timeout: 10 * time.Second}}
}

func (m *resendMailer) Send(ctx context.Context, to, subject, text string) error {
	body, _ := json.Marshal(map[string]any{
		"from":    m.from,
		"to":      []string{to},
		"subject": subject,
		"text":    text,
		"html":    "<div style=\"font-family:sans-serif;white-space:pre-wrap\">" + html.EscapeString(text) + "</div>",
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(res.Body, 500))
		return fmt.Errorf("resend: %s: %s", res.Status, msg)
	}
	return nil
}

// sendAsync sends in the background so response timing doesn't reveal whether
// an email was sent (i.e. whether an account exists).
func sendAsync(m Mailer, to, subject, text string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := m.Send(ctx, to, subject, text); err != nil {
			log.Printf("send email: %v", err)
		}
	}()
}
