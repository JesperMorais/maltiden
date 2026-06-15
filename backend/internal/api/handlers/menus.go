package handlers

import (
	"errors"
	"log"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/pkg/middleware"
	"net/http"

	"github.com/getsentry/sentry-go"
)

type MenuHandler struct {
	menuService *services.MenuService
}

func NewMenuHandler(menuService *services.MenuService) *MenuHandler {
	return &MenuHandler{menuService: menuService}
}

func (h *MenuHandler) Generate(w http.ResponseWriter, r *http.Request) {
	// Get household ID from auth context
	householdID := middleware.GetHouseholdID(r)
	if householdID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.GenerateMenuRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	menu, err := h.menuService.Generate(householdID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidDays):
			WriteError(w, http.StatusBadRequest, "invalid_days")
		case errors.Is(err, domain.ErrInvalidServings):
			WriteError(w, http.StatusBadRequest, "invalid_servings")
		default:
			sentry.CaptureException(err)
			log.Printf("ERROR [GenerateMenu] %v", err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	if menu == nil {
		WriteError(w, http.StatusBadRequest, "no_recipes_available")
		return
	}

	WriteJSON(w, http.StatusCreated, menu)
}

func (h *MenuHandler) UpdateCurrent(w http.ResponseWriter, r *http.Request) {
	householdID := middleware.GetHouseholdID(r)
	if householdID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.UpdateMenuRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	menu, err := h.menuService.UpdateCurrent(householdID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			WriteError(w, http.StatusNotFound, "no_active_menu")
		case errors.Is(err, domain.ErrInvalidDays):
			WriteError(w, http.StatusBadRequest, "invalid_days")
		default:
			sentry.CaptureException(err)
			log.Printf("ERROR [UpdateCurrentMenu] %v", err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	WriteJSON(w, http.StatusOK, menu)
}

func (h *MenuHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	householdID := middleware.GetHouseholdID(r)
	if householdID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	prefs, err := h.menuService.GetPreferences(householdID)
	if err != nil {
		sentry.CaptureException(err)
		log.Printf("ERROR [GetMenuPreferences] %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	WriteJSON(w, http.StatusOK, prefs)
}

func (h *MenuHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	householdID := middleware.GetHouseholdID(r)
	if householdID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.UpdateMenuPreferencesRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	prefs, err := h.menuService.UpdatePreferences(householdID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidDays):
			WriteError(w, http.StatusBadRequest, "invalid_days")
		case errors.Is(err, domain.ErrInvalidServings):
			WriteError(w, http.StatusBadRequest, "invalid_servings")
		case errors.Is(err, domain.ErrInvalidVegetarianDays):
			WriteError(w, http.StatusBadRequest, "invalid_vegetarian_days")
		case errors.Is(err, domain.ErrTooManyExcludedTags):
			WriteError(w, http.StatusBadRequest, "too_many_excluded_tags")
		case errors.Is(err, domain.ErrInvalidExcludedTag):
			WriteError(w, http.StatusBadRequest, "invalid_excluded_tag")
		case errors.Is(err, domain.ErrTooManyDislikedIngredients):
			WriteError(w, http.StatusBadRequest, "too_many_disliked_ingredients")
		case errors.Is(err, domain.ErrInvalidDislikedIngredient):
			WriteError(w, http.StatusBadRequest, "invalid_disliked_ingredient")
		default:
			sentry.CaptureException(err)
			log.Printf("ERROR [UpdateMenuPreferences] %v", err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	WriteJSON(w, http.StatusOK, prefs)
}

func (h *MenuHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	// Get household ID from auth context
	householdID := middleware.GetHouseholdID(r)
	if householdID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	menu, err := h.menuService.GetCurrent(householdID)
	if err != nil {
		sentry.CaptureException(err)
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
