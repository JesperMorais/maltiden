package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/pkg/middleware"

	"github.com/getsentry/sentry-go"
)

type AccountHandler struct {
	accountService *services.AccountService
}

func NewAccountHandler(accountService *services.AccountService) *AccountHandler {
	return &AccountHandler{accountService: accountService}
}

type deleteAccountRequest struct {
	Password string `json:"password"`
}

// Delete handles DELETE /account. Body: {"password": "..."}. 204 on success.
func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req deleteAccountRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}
	if req.Password == "" {
		WriteError(w, http.StatusBadRequest, "password_required")
		return
	}

	err := h.accountService.DeleteAccount(userID, req.Password)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, domain.ErrInvalidCredentials):
		WriteError(w, http.StatusForbidden, "invalid_password")
	case errors.Is(err, domain.ErrNotFound):
		WriteError(w, http.StatusNotFound, "not_found")
	default:
		sentry.CaptureException(err)
		log.Printf("ERROR [DeleteAccount] %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error")
	}
}

// Export handles GET /account/export and returns the user's data as a JSON
// file download.
func (h *AccountHandler) Export(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	data, err := h.accountService.Export(userID)
	if errors.Is(err, domain.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "not_found")
		return
	}
	if err != nil {
		sentry.CaptureException(err)
		log.Printf("ERROR [ExportAccount] %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	filename := "maltiden-export-" + time.Now().UTC().Format("2006-01-02") + ".json"
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Cache-Control", "no-store")
	WriteJSON(w, http.StatusOK, data)
}
