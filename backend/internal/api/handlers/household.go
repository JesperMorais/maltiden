package handlers

import (
	"encoding/json"
	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/middleware"
	"net/http"
)

type HouseholdHandler struct {
	householdStorage *sqlite.HouseholdStorage
}

func NewHouseholdHandler(householdStorage *sqlite.HouseholdStorage) *HouseholdHandler {
	return &HouseholdHandler{householdStorage: householdStorage}
}

func (h *HouseholdHandler) GetMyHousehold(w http.ResponseWriter, r *http.Request) {
	// Get user ID from auth middleware context
	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Get household
	household, err := h.householdStorage.GetByUserID(userID)
	if err != nil {
		http.Error(w, `{"error":"internal_server_error"}`, http.StatusInternalServerError)
		return
	}

	if household == nil {
		http.Error(w, `{"error":"household_not_found"}`, http.StatusNotFound)
		return
	}

	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(household)
}
