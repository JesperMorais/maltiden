package handlers

import (
	"errors"
	"github.com/getsentry/sentry-go"
	"log"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/pkg/middleware"
	"net/http"
)

type HouseholdHandler struct {
	householdService *services.HouseholdService
}

func NewHouseholdHandler(householdService *services.HouseholdService) *HouseholdHandler {
	return &HouseholdHandler{householdService: householdService}
}

func (h *HouseholdHandler) GetMyHousehold(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	household, err := h.householdService.GetMyHousehold(userID)
	if err != nil {
		log.Printf("ERROR [GetMyHousehold] %v", err)
		sentry.CaptureException(err)
		WriteError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	if household == nil {
		WriteError(w, http.StatusNotFound, "household_not_found")
		return
	}

	WriteJSON(w, http.StatusOK, household)
}

func (h *HouseholdHandler) UpdateMyHousehold(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	householdID := middleware.GetHouseholdID(r)
	if userID == "" || householdID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.UpdateHouseholdRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	err := h.householdService.UpdateName(householdID, userID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrHouseholdNameRequired):
			WriteError(w, http.StatusBadRequest, "household_name_required")
		case errors.Is(err, domain.ErrHouseholdNameTooLong):
			WriteError(w, http.StatusBadRequest, "household_name_too_long")
		case errors.Is(err, domain.ErrForbidden):
			WriteError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, domain.ErrNotFound):
			WriteError(w, http.StatusNotFound, "not_found")
		default:
			log.Printf("ERROR [UpdateMyHousehold] %v", err)
			sentry.CaptureException(err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *HouseholdHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	householdID := middleware.GetHouseholdID(r)
	if userID == "" || householdID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Verify the user is still an active member with permission to invite
	role, err := h.householdService.GetMemberRole(householdID, userID)
	if err != nil || role == "" || role == "guest" {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	resp, err := h.householdService.CreateInvite(householdID)
	if err != nil {
		log.Printf("ERROR [CreateInvite] %v", err)
		sentry.CaptureException(err)
		WriteError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	WriteJSON(w, http.StatusCreated, resp)
}

func (h *HouseholdHandler) JoinHousehold(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.JoinHouseholdRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	resp, err := h.householdService.JoinHousehold(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCodeRequired):
			WriteError(w, http.StatusBadRequest, "code_required")
		case errors.Is(err, domain.ErrInvalidCode):
			WriteError(w, http.StatusBadRequest, "invalid_code")
		case errors.Is(err, domain.ErrAlreadyMember):
			WriteError(w, http.StatusConflict, "already_member")
		default:
			log.Printf("ERROR [JoinHousehold] %v", err)
			sentry.CaptureException(err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	WriteJSON(w, http.StatusOK, resp)
}

func (h *HouseholdHandler) GetMemberStatuses(w http.ResponseWriter, r *http.Request) {
	householdID := middleware.GetHouseholdID(r)
	if householdID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp, err := h.householdService.GetMemberStatuses(householdID)
	if err != nil {
		log.Printf("ERROR [GetMemberStatuses] %v", err)
		sentry.CaptureException(err)
		WriteError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	WriteJSON(w, http.StatusOK, resp)
}

func (h *HouseholdHandler) UpdateMemberStatus(w http.ResponseWriter, r *http.Request) {
	householdID := middleware.GetHouseholdID(r)
	if householdID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	memberID := r.PathValue("id")
	if !ValidateID(w, memberID, "member_id") {
		return
	}

	var req domain.UpdateMemberStatusRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	if req.IsEatingToday == nil && req.WantsLunchBox == nil {
		WriteError(w, http.StatusBadRequest, "no_fields_to_update")
		return
	}

	err := h.householdService.UpdateMemberStatus(householdID, memberID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			WriteError(w, http.StatusNotFound, "not_found")
		default:
			log.Printf("ERROR [UpdateMemberStatus] %v", err)
			sentry.CaptureException(err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *HouseholdHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	householdID := middleware.GetHouseholdID(r)
	if userID == "" || householdID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetID := r.PathValue("id")
	if !ValidateID(w, targetID, "member_id") {
		return
	}

	err := h.householdService.RemoveMember(householdID, userID, targetID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCannotRemove):
			WriteError(w, http.StatusForbidden, "cannot_remove")
		case errors.Is(err, domain.ErrForbidden):
			WriteError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, domain.ErrNotFound):
			WriteError(w, http.StatusNotFound, "not_found")
		default:
			log.Printf("ERROR [RemoveMember] %v", err)
			sentry.CaptureException(err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
