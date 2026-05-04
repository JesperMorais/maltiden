package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"maltiden/internal/domain"
	"maltiden/pkg/email"
	"maltiden/pkg/utils"
	"net/mail"
	"os"
	"strings"
	"time"
)

// PasswordResetService handles the forgot/reset password flow.
//
// Tokens are 32 random bytes encoded as hex (64 hex chars), single-use, and
// expire one hour after creation. Successfully resetting a password also
// increments the user's token_version, invalidating any existing JWTs.
type PasswordResetService struct {
	db          *sql.DB
	userStorage domain.UserRepository
	emailer     email.EmailSender
	frontendURL string
	tokenTTL    time.Duration
}

// NewPasswordResetService constructs a PasswordResetService.
func NewPasswordResetService(db *sql.DB, userStorage domain.UserRepository, emailer email.EmailSender) *PasswordResetService {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	return &PasswordResetService{
		db:          db,
		userStorage: userStorage,
		emailer:     emailer,
		frontendURL: frontendURL,
		tokenTTL:    time.Hour,
	}
}

// RequestReset begins the password reset flow for the given email.
//
// To avoid leaking which emails are registered, this returns nil for unknown
// emails (and also for invalid emails) — only logging the no-op internally.
func (s *PasswordResetService) RequestReset(emailAddr string) error {
	emailAddr = strings.TrimSpace(emailAddr)

	// Validate email format silently — never leak.
	if _, err := mail.ParseAddress(emailAddr); err != nil {
		slog.Info("password_reset.request: invalid email format (no-op)", "email", emailAddr)
		return nil
	}

	user, err := s.userStorage.GetByEmail(emailAddr)
	if err != nil {
		// Real DB error — surface it; the handler still returns 200 to the
		// client but logs server-side.
		return err
	}
	if user == nil {
		slog.Info("password_reset.request: unknown email (no-op)", "email", emailAddr)
		return nil
	}

	token, err := generateResetToken()
	if err != nil {
		return err
	}

	now := time.Now()
	prt := &domain.PasswordResetToken{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: now.Add(s.tokenTTL),
		CreatedAt: now,
	}
	if err := s.userStorage.CreatePasswordResetToken(prt); err != nil {
		return err
	}

	resetURL := s.frontendURL + "/reset-password?token=" + token
	subject := "Återställ ditt lösenord på Måltiden"
	body := "Hej!\n\nKlicka på länken nedan för att återställa ditt lösenord. Länken är giltig i 1 timme.\n\n" +
		resetURL +
		"\n\nOm du inte begärt en återställning kan du ignorera detta mejl.\n\nVänliga hälsningar,\nMåltiden"

	if err := s.emailer.Send(user.Email, subject, body); err != nil {
		// Do not surface email errors — log only. We don't want a transient
		// email outage to leak account existence.
		slog.Error("password_reset.request: email send failed", "error", err, "email", user.Email)
	}
	return nil
}

// ResetPassword consumes a reset token and updates the user's password.
//
// The token must exist, not be used, and not be expired. The new password
// must satisfy the same minimum length as registration (>= 8 chars).
// On success the user's token_version is incremented, invalidating any
// existing JWTs.
func (s *PasswordResetService) ResetPassword(token, newPassword string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.ErrInvalidResetToken
	}
	if len(newPassword) < 8 {
		return domain.ErrWeakPassword
	}

	prt, err := s.userStorage.GetPasswordResetToken(token)
	if err != nil {
		return err
	}
	if prt == nil {
		return domain.ErrInvalidResetToken
	}
	if prt.UsedAt != nil {
		return domain.ErrUsedResetToken
	}
	if time.Now().After(prt.ExpiresAt) {
		return domain.ErrExpiredResetToken
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Wrap the three writes in a transaction so a mid-flow crash cannot leave
	// the system in a partially-updated state (e.g., password changed but token
	// still usable, or token consumed but old JWTs still valid).
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.userStorage.UpdatePasswordTx(tx, prt.UserID, hash); err != nil {
		return err
	}
	if err := s.userStorage.MarkPasswordResetTokenUsedTx(tx, token); err != nil {
		return err
	}
	// Invalidate all existing JWTs for this user.
	if err := s.userStorage.IncrementTokenVersionTx(tx, prt.UserID); err != nil {
		return err
	}
	return tx.Commit()
}

// generateResetToken returns 32 cryptographically-random bytes encoded as hex.
func generateResetToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
