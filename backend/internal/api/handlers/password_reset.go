package handlers

import (
	"errors"
	"log"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"net/http"

	"github.com/getsentry/sentry-go"
)

// PasswordResetHandler exposes HTTP endpoints for the forgot/reset password flow.
type PasswordResetHandler struct {
	service *services.PasswordResetService
}

// NewPasswordResetHandler constructs a PasswordResetHandler.
func NewPasswordResetHandler(service *services.PasswordResetService) *PasswordResetHandler {
	return &PasswordResetHandler{service: service}
}

// ForgotPassword handles POST /auth/forgot-password.
//
// Always returns 200 with {"ok": true} regardless of whether the email exists,
// to avoid leaking which accounts are registered.
func (h *PasswordResetHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req domain.ForgotPasswordRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	if err := h.service.RequestReset(req.Email); err != nil {
		// Log internally — never surface to client.
		sentry.CaptureException(err)
		log.Printf("ERROR [ForgotPassword] %v", err)
	}

	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ResetPassword handles POST /auth/reset-password.
//
// Returns 200 on success, 400 on invalid/expired/used token or weak password.
func (h *PasswordResetHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req domain.ResetPasswordRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	if err := h.service.ResetPassword(req.Token, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidResetToken):
			WriteError(w, http.StatusBadRequest, "invalid_reset_token")
		case errors.Is(err, domain.ErrExpiredResetToken):
			WriteError(w, http.StatusBadRequest, "expired_reset_token")
		case errors.Is(err, domain.ErrUsedResetToken):
			WriteError(w, http.StatusBadRequest, "used_reset_token")
		case errors.Is(err, domain.ErrWeakPassword):
			WriteError(w, http.StatusBadRequest, "weak_password")
		default:
			sentry.CaptureException(err)
			log.Printf("ERROR [ResetPassword] %v", err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
