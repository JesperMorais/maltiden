package handlers

import (
	"encoding/json"
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
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	menuID := r.URL.Query().Get("menuId")
	if menuID == "" {
		http.Error(w, `{"error":"menu_id_required"}`, http.StatusBadRequest)
		return
	}

	// IDOR protection: verify menu belongs to user's household
	menuHouseholdID, err := h.menuStorage.GetHouseholdIDByMenuID(menuID)
	if err != nil {
		http.Error(w, `{"error":"internal_error"}`, http.StatusInternalServerError)
		return
	}
	if menuHouseholdID == "" {
		http.Error(w, `{"error":"menu_not_found"}`, http.StatusNotFound)
		return
	}
	if menuHouseholdID != householdID {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}

	list, err := h.shoppingService.GetShoppingList(menuID)
	if err != nil {
		http.Error(w, `{"error":"internal_error"}`, http.StatusInternalServerError)
		return
	}

	if list == nil {
		http.Error(w, `{"error":"menu_not_found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *ShoppingHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	// Verify authenticated user has householdID
	householdID := middleware.GetHouseholdID(r.Context())
	if householdID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Extract item ID from path using PathValue
	itemID := r.PathValue("id")
	if itemID == "" {
		http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
		return
	}

	menuID := r.URL.Query().Get("menuId")
	if menuID == "" {
		http.Error(w, `{"error":"menu_id_required"}`, http.StatusBadRequest)
		return
	}

	// IDOR protection: verify menu belongs to user's household
	menuHouseholdID, err := h.menuStorage.GetHouseholdIDByMenuID(menuID)
	if err != nil {
		http.Error(w, `{"error":"internal_error"}`, http.StatusInternalServerError)
		return
	}
	if menuHouseholdID == "" {
		http.Error(w, `{"error":"menu_not_found"}`, http.StatusNotFound)
		return
	}
	if menuHouseholdID != householdID {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}

	var req domain.UpdateShoppingItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
		return
	}

	if err := h.shoppingService.UpdateItemChecked(menuID, itemID, req.Checked); err != nil {
		http.Error(w, `{"error":"internal_error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
