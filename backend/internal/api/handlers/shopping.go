package handlers

import (
	"encoding/json"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"net/http"
	"strings"
)

type ShoppingHandler struct {
	shoppingService *services.ShoppingService
}

func NewShoppingHandler(shoppingService *services.ShoppingService) *ShoppingHandler {
	return &ShoppingHandler{shoppingService: shoppingService}
}

func (h *ShoppingHandler) GetShoppingList(w http.ResponseWriter, r *http.Request) {
	menuID := r.URL.Query().Get("menuId")
	if menuID == "" {
		http.Error(w, `{"error":"menu_id_required"}`, http.StatusBadRequest)
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
	// Extract item ID from path: /shopping-list/items/{id}
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
		return
	}
	itemID := parts[len(parts)-1]

	menuID := r.URL.Query().Get("menuId")
	if menuID == "" {
		http.Error(w, `{"error":"menu_id_required"}`, http.StatusBadRequest)
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
