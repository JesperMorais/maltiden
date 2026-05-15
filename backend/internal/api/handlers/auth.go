package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"net/http"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	resp, err := h.authService.Register(req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrWeakPassword):
			WriteError(w, http.StatusBadRequest, "weak_password")
		case errors.Is(err, domain.ErrDuplicateEmail):
			WriteError(w, http.StatusConflict, "email_already_exists")
		case errors.Is(err, domain.ErrInvalidEmail):
			WriteError(w, http.StatusBadRequest, "invalid_email")
		default:
			log.Printf("ERROR [Register] %v", err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	WriteJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) PasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" {
		WriteError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	slog.Info("password reset requested", "email", body.Email)
	WriteJSON(w, http.StatusOK, map[string]string{"message": "if account exists, reset link sent"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	resp, err := h.authService.Login(req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			WriteError(w, http.StatusUnauthorized, "invalid_credentials")
		} else {
			log.Printf("ERROR [Login] %v", err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	WriteJSON(w, http.StatusOK, resp)
}
