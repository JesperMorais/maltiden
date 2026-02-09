package handlers

import (
	"encoding/json"
	"errors"
	"log"
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
	// Parse JSON body
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	// Call service
	resp, err := h.authService.Register(req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrWeakPassword):
			WriteError(w, http.StatusBadRequest, "weak_password")
		case errors.Is(err, domain.ErrDuplicateEmail):
			WriteError(w, http.StatusConflict, "email_already_exists")
		default:
			log.Printf("ERROR [Register] %v", err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	// Return JSON
	WriteJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Parse JSON body
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	// Call service
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

	// Return JSON
	WriteJSON(w, http.StatusOK, resp)
}
