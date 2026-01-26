package handlers

import (
	"encoding/json"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"net/http"
	"strings"
)

type RecipeHandler struct {
	recipeService *services.RecipeService
}

func NewRecipeHandler(recipeService *services.RecipeService) *RecipeHandler {
	return &RecipeHandler{recipeService: recipeService}
}

func (h *RecipeHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	recipes, err := h.recipeService.GetAll()
	if err != nil {
		http.Error(w, `{"error":"internal_error"}`, http.StatusInternalServerError)
		return
	}

	response := domain.RecipesResponse{Recipes: recipes}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *RecipeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path: /recipes/{id}
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
		return
	}
	id := parts[len(parts)-1]

	recipe, err := h.recipeService.GetByID(id)
	if err != nil {
		http.Error(w, `{"error":"internal_error"}`, http.StatusInternalServerError)
		return
	}

	if recipe == nil {
		http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recipe)
}

func (h *RecipeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid_request"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.recipeService.Create(req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
