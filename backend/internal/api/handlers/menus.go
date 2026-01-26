package handlers

import (
	"encoding/json"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/pkg/middleware"
	"net/http"
)

type MenuHandler struct {
	menuService *services.MenuService
}

func NewMenuHandler(menuService *services.MenuService) *MenuHandler {
	return &MenuHandler{menuService: menuService}
}

func (h *MenuHandler) Generate(w http.ResponseWriter, r *http.Request) {
	// Get household ID from auth context
	householdID := middleware.GetHouseholdID(r.Context())
	if householdID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req domain.GenerateMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
		return
	}

	menu, err := h.menuService.Generate(householdID, req)
	if err != nil {
		http.Error(w, `{"error":"internal_error"}`, http.StatusInternalServerError)
		return
	}

	if menu == nil {
		http.Error(w, `{"error":"no_recipes_available"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(menu)
}

func (h *MenuHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	// Get household ID from auth context
	householdID := middleware.GetHouseholdID(r.Context())
	if householdID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	menu, err := h.menuService.GetCurrent(householdID)
	if err != nil {
		http.Error(w, `{"error":"internal_error"}`, http.StatusInternalServerError)
		return
	}

	if menu == nil {
		http.Error(w, `{"error":"no_active_menu"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menu)
}
