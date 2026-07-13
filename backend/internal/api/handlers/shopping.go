package handlers

import (
	"context"
	"log"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/pkg/middleware"
	"net/http"
)

type ShoppingHandler struct {
	shoppingService *services.ShoppingService
	menuStorage     domain.MenuRepository
	events          *services.EventService
}

func NewShoppingHandler(shoppingService *services.ShoppingService, menuStorage domain.MenuRepository) *ShoppingHandler {
	return &ShoppingHandler{
		shoppingService: shoppingService,
		menuStorage:     menuStorage,
	}
}

// SetEventService wires in an optional, nil-safe analytics dependency.
func (h *ShoppingHandler) SetEventService(events *services.EventService) {
	h.events = events
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

	// IDOR protection: verify menu belongs to user's household
	menuHouseholdID, err := h.menuStorage.GetHouseholdIDByMenuID(menuID)
	if err != nil {
		log.Printf("ERROR [GetShoppingList] %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	if menuHouseholdID == "" {
		WriteError(w, http.StatusNotFound, "menu_not_found")
		return
	}
	if menuHouseholdID != householdID {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	list, err := h.shoppingService.GetShoppingList(menuID)
	if err != nil {
		log.Printf("ERROR [GetShoppingList] %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	if list == nil {
		WriteError(w, http.StatusNotFound, "menu_not_found")
		return
	}

	if h.events != nil {
		h.events.Track(context.Background(), domain.EventShoppingListViewed, householdID, "")
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

	// IDOR protection: verify menu belongs to user's household
	menuHouseholdID, err := h.menuStorage.GetHouseholdIDByMenuID(menuID)
	if err != nil {
		log.Printf("ERROR [UpdateShoppingItem] %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	if menuHouseholdID == "" {
		WriteError(w, http.StatusNotFound, "menu_not_found")
		return
	}
	if menuHouseholdID != householdID {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req domain.UpdateShoppingItemRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	if err := h.shoppingService.UpdateItemChecked(menuID, itemID, req.Checked); err != nil {
		log.Printf("ERROR [UpdateShoppingItem] %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
