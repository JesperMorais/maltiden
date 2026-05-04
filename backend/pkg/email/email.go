// Package email provides minimal email sending abstractions.
//
// Two implementations are provided:
//   - LogSender: logs messages via slog. The default; useful for local dev and
//     beta where email isn't critical (operators can read logs to find reset
//     links).
//   - ResendSender: posts to the Resend HTTP API when RESEND_API_KEY is set.
package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// EmailSender is the minimal interface used by services that need to send
// transactional email.
type EmailSender interface {
	Send(to, subject, body string) error
}

// LogSender logs would-be emails to the standard slog logger. It never fails.
type LogSender struct {
	From string
}

// NewLogSender returns a LogSender with the given from address.
func NewLogSender(from string) *LogSender {
	return &LogSender{From: from}
}

// Send logs the email contents at INFO level.
func (s *LogSender) Send(to, subject, body string) error {
	slog.Info("email.send (log-only)",
		"from", s.From,
		"to", to,
		"subject", subject,
		"body", body,
	)
	return nil
}

// ResendSender posts emails to the Resend HTTP API.
//
// The actual HTTP integration is intentionally minimal — if the request fails,
// the caller decides whether to surface that failure. Most callers should
// swallow it (so a transient email outage doesn't leak whether an account
// exists during password reset, for example).
type ResendSender struct {
	APIKey string
	From   string
	Client *http.Client
}

// NewResendSender returns a ResendSender configured with the given API key
// and from address.
func NewResendSender(apiKey, from string) *ResendSender {
	return &ResendSender{
		APIKey: apiKey,
		From:   from,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
}

// Send posts a single email to the Resend API.
func (s *ResendSender) Send(to, subject, body string) error {
	payload := resendPayload{
		From:    s.From,
		To:      []string{to},
		Subject: subject,
		Text:    body,
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal resend payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("build resend request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("resend request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend api status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
