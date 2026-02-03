package handlers

import (
	"encoding/json"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/pkg/middleware"
	"net/http"
	"strings"
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
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	household, err := h.householdService.GetMyHousehold(userID)
	if err != nil {
		http.Error(w, `{"error":"internal_server_error"}`, http.StatusInternalServerError)
		return
	}

	if household == nil {
		http.Error(w, `{"error":"household_not_found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(household)
}

func (h *HouseholdHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	householdID := middleware.GetHouseholdID(r.Context())
	if householdID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	resp, err := h.householdService.CreateInvite(householdID)
	if err != nil {
		http.Error(w, `{"error":"internal_server_error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *HouseholdHandler) JoinHousehold(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req domain.JoinHouseholdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.householdService.JoinHousehold(userID, req)
	if err != nil {
		switch err.Error() {
		case "invalid_code":
			http.Error(w, `{"error":"invalid_code"}`, http.StatusBadRequest)
		case "already_member":
			http.Error(w, `{"error":"already_member"}`, http.StatusConflict)
		default:
			http.Error(w, `{"error":"internal_server_error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *HouseholdHandler) GetMemberStatuses(w http.ResponseWriter, r *http.Request) {
	householdID := middleware.GetHouseholdID(r.Context())
	if householdID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	resp, err := h.householdService.GetMemberStatuses(householdID)
	if err != nil {
		http.Error(w, `{"error":"internal_server_error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *HouseholdHandler) UpdateMemberStatus(w http.ResponseWriter, r *http.Request) {
	householdID := middleware.GetHouseholdID(r.Context())
	if householdID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Extract member ID from path: /households/members/{id}/status
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 5 {
		http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
		return
	}
	memberID := parts[len(parts)-2] // second to last is the {id}

	var req domain.UpdateMemberStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
		return
	}

	if req.IsEatingToday == nil && req.WantsLunchBox == nil {
		http.Error(w, `{"error":"no_fields_to_update"}`, http.StatusBadRequest)
		return
	}

	err := h.householdService.UpdateMemberStatus(householdID, memberID, req)
	if err != nil {
		switch err.Error() {
		case "not_found":
			http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
		default:
			http.Error(w, `{"error":"internal_server_error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (h *HouseholdHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	householdID := middleware.GetHouseholdID(r.Context())
	if userID == "" || householdID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Extract member ID from path: /households/members/{id}
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
		return
	}
	targetID := parts[len(parts)-1]

	err := h.householdService.RemoveMember(householdID, userID, targetID)
	if err != nil {
		switch err.Error() {
		case "cannot_remove":
			http.Error(w, `{"error":"cannot_remove"}`, http.StatusForbidden)
		case "forbidden":
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		case "not_found":
			http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
		default:
			http.Error(w, `{"error":"internal_server_error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
