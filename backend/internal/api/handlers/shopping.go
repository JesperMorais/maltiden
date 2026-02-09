package handlers

import (
	"encoding/json"
	"log"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/middleware"
	"net/http"
)

type ShoppingHandler struct {
	shoppingService *services.ShoppingService
	menuStorage     *sqlite.MenuStorage
}

func NewShoppingHandler(shoppingService *services.ShoppingService, menuStorage *sqlite.MenuStorage) *ShoppingHandler {
	return &ShoppingHandler{
		shoppingService: shoppingService,
		menuStorage:     menuStorage,
	}
}

func (h *ShoppingHandler) GetShoppingList(w http.ResponseWriter, r *http.Request) {
	// Verify authenticated user has householdID
	householdID := middleware.GetHouseholdID(r.Context())
	if householdID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	menuID := r.URL.Query().Get("menuId")
	if menuID == "" {
		WriteError(w, http.StatusBadRequest, "menu_id_required")
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

	WriteJSON(w, http.StatusOK, list)
}

func (h *ShoppingHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	// Verify authenticated user has householdID
	householdID := middleware.GetHouseholdID(r.Context())
	if householdID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Extract item ID from path using PathValue
	itemID := r.PathValue("id")
	if itemID == "" {
		WriteError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	menuID := r.URL.Query().Get("menuId")
	if menuID == "" {
		WriteError(w, http.StatusBadRequest, "menu_id_required")
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	if err := h.shoppingService.UpdateItemChecked(menuID, itemID, req.Checked); err != nil {
		log.Printf("ERROR [UpdateShoppingItem] %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
