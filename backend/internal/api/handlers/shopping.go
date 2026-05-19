package handlers

import (
	"database/sql"
	"errors"
	"log/slog"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/pkg/middleware"
	"net/http"
)

type ShoppingHandler struct {
	shoppingService *services.ShoppingService
	menuStorage     domain.MenuRepository
}

func NewShoppingHandler(shoppingService *services.ShoppingService, menuStorage domain.MenuRepository) *ShoppingHandler {
	return &ShoppingHandler{
		shoppingService: shoppingService,
		menuStorage:     menuStorage,
	}
}

func (h *ShoppingHandler) GetShoppingList(w http.ResponseWriter, r *http.Request) {
	// Verify authenticated user has householdID
	householdID := middleware.GetHouseholdID(r)
	if householdID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	menuID := r.URL.Query().Get("menuId")
	if !ValidateID(w, menuID, "menu_id") {
		return
	}

	list, err := h.shoppingService.GetShoppingList(menuID, householdID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMenuNotFound):
			WriteError(w, http.StatusNotFound, "menu_not_found")
		case errors.Is(err, domain.ErrForbidden):
			WriteError(w, http.StatusForbidden, "forbidden")
		default:
			slog.Error("GetShoppingList failed", "error", err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	if list == nil {
		WriteError(w, http.StatusNotFound, "menu_not_found")
		return
	}

	WriteJSON(w, http.StatusOK, list)
}

func (h *ShoppingHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	// Verify authenticated user has householdID
	householdID := middleware.GetHouseholdID(r)
	if householdID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Extract and validate item ID from path
	itemID := r.PathValue("id")
	if !ValidateItemID(w, itemID, "item_id") {
		return
	}

	menuID := r.URL.Query().Get("menuId")
	if !ValidateID(w, menuID, "menu_id") {
		return
	}

	var req domain.UpdateShoppingItemRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	if err := h.shoppingService.UpdateItemChecked(menuID, itemID, householdID, req.Checked); err != nil {
		switch {
		case errors.Is(err, domain.ErrMenuNotFound):
			WriteError(w, http.StatusNotFound, "menu_not_found")
		case errors.Is(err, domain.ErrForbidden):
			WriteError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, sql.ErrNoRows):
			// Custom-item not found OR cross-tenant probe — return 404 either way.
			slog.Info("UpdateShoppingItem not found", "item_id", itemID)
			WriteError(w, http.StatusNotFound, "not_found")
		default:
			slog.Error("UpdateShoppingItem failed", "error", err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *ShoppingHandler) AddCustomItem(w http.ResponseWriter, r *http.Request) {
	householdID := middleware.GetHouseholdID(r)
	if householdID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	menuID := r.URL.Query().Get("menuId")
	if !ValidateID(w, menuID, "menu_id") {
		return
	}

	var req domain.CreateCustomItemRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	item, err := h.shoppingService.CreateCustomItem(menuID, householdID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMenuNotFound):
			WriteError(w, http.StatusNotFound, "menu_not_found")
		case errors.Is(err, domain.ErrForbidden):
			WriteError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, domain.ErrNameRequired):
			WriteError(w, http.StatusBadRequest, "name_required")
		case errors.Is(err, domain.ErrNameTooLong):
			WriteError(w, http.StatusBadRequest, "name_too_long")
		case errors.Is(err, domain.ErrUnitTooLong):
			WriteError(w, http.StatusBadRequest, "unit_too_long")
		case errors.Is(err, domain.ErrInvalidAmount):
			WriteError(w, http.StatusBadRequest, "invalid_amount")
		case errors.Is(err, domain.ErrAmountTooLarge):
			WriteError(w, http.StatusBadRequest, "amount_too_large")
		case errors.Is(err, domain.ErrTooManyItems):
			WriteError(w, http.StatusBadRequest, "too_many_items")
		default:
			slog.Error("AddCustomItem failed", "error", err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	WriteJSON(w, http.StatusCreated, item)
}

func (h *ShoppingHandler) DeleteCustomItem(w http.ResponseWriter, r *http.Request) {
	householdID := middleware.GetHouseholdID(r)
	if householdID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	itemID := r.PathValue("id")
	if !ValidateItemID(w, itemID, "item_id") {
		return
	}

	if err := h.shoppingService.DeleteCustomItem(itemID, householdID); err != nil {
		// Not found OR cross-tenant probe — return 404 either way.
		// Logged at INFO level (not ERROR) to avoid log spam from probing.
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("DeleteCustomItem not found", "item_id", itemID)
			WriteError(w, http.StatusNotFound, "not_found")
			return
		}
		slog.Error("DeleteCustomItem failed", "error", err)
		WriteError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
