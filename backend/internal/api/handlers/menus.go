package handlers

import (
	"encoding/json"
	"log"
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
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.GenerateMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	menu, err := h.menuService.Generate(householdID, req)
	if err != nil {
		log.Printf("ERROR [GenerateMenu] %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	if menu == nil {
		WriteError(w, http.StatusBadRequest, "no_recipes_available")
		return
	}

	WriteJSON(w, http.StatusCreated, menu)
}

func (h *MenuHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	// Get household ID from auth context
	householdID := middleware.GetHouseholdID(r.Context())
	if householdID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	menu, err := h.menuService.GetCurrent(householdID)
	if err != nil {
		log.Printf("ERROR [GetCurrentMenu] %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	if menu == nil {
		WriteError(w, http.StatusNotFound, "no_active_menu")
		return
	}

	WriteJSON(w, http.StatusOK, menu)
}
