package handlers

import (
	"encoding/json"
	"errors"
	"log"
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
	// Parse query parameters for filtering
	filter := &domain.RecipeFilter{
		Name: r.URL.Query().Get("name"),
		Tag:  r.URL.Query().Get("tag"),
	}

	recipes, err := h.recipeService.GetAll(filter)
	if err != nil {
		log.Printf("ERROR [GetAllRecipes] %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	response := domain.RecipesResponse{Recipes: recipes}

	WriteJSON(w, http.StatusOK, response)
}

func (h *RecipeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path: /recipes/{id}
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		WriteError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	id := parts[len(parts)-1]

	recipe, err := h.recipeService.GetByID(id)
	if err != nil {
		log.Printf("ERROR [GetRecipeByID] %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	if recipe == nil {
		WriteError(w, http.StatusNotFound, "not_found")
		return
	}

	WriteJSON(w, http.StatusOK, recipe)
}

func (h *RecipeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	resp, err := h.recipeService.Create(req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNameRequired):
			WriteError(w, http.StatusBadRequest, "name_required")
		case errors.Is(err, domain.ErrInvalidServings):
			WriteError(w, http.StatusBadRequest, "invalid_servings")
		case errors.Is(err, domain.ErrIngredientsRequired):
			WriteError(w, http.StatusBadRequest, "ingredients_required")
		case errors.Is(err, domain.ErrInstructionsRequired):
			WriteError(w, http.StatusBadRequest, "instructions_required")
		default:
			log.Printf("ERROR [CreateRecipe] %v", err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	WriteJSON(w, http.StatusCreated, resp)
}
