package handlers

import (
	"errors"
	"log"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/pkg/middleware"
	"net/http"
)

type RecipeHandler struct {
	recipeService *services.RecipeService
}

func NewRecipeHandler(recipeService *services.RecipeService) *RecipeHandler {
	return &RecipeHandler{recipeService: recipeService}
}

func (h *RecipeHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	// householdID is empty for unauthenticated requests (public route)
	householdID := middleware.GetHouseholdID(r)

	// Parse query parameters for filtering
	filter := &domain.RecipeFilter{
		Name: r.URL.Query().Get("name"),
		Tag:  r.URL.Query().Get("tag"),
	}

	recipes, err := h.recipeService.GetAll(filter, householdID)
	if err != nil {
		log.Printf("ERROR [GetAllRecipes] %v", err)
		WriteError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	response := domain.RecipesResponse{Recipes: recipes}

	WriteJSON(w, http.StatusOK, response)
}

func (h *RecipeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Use PathValue instead of manual path splitting (VALID-03)
	id := r.PathValue("id")
	if !ValidateID(w, id, "recipe_id") {
		return
	}

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

func (h *RecipeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !ValidateID(w, id, "recipe_id") {
		return
	}

	householdID := middleware.GetHouseholdID(r)

	var req domain.UpdateRecipeRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	recipe, err := h.recipeService.Update(id, householdID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			WriteError(w, http.StatusNotFound, "not_found")
		case errors.Is(err, domain.ErrForbidden):
			WriteError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, domain.ErrNameRequired):
			WriteError(w, http.StatusBadRequest, "name_required")
		case errors.Is(err, domain.ErrInvalidServings):
			WriteError(w, http.StatusBadRequest, "invalid_servings")
		case errors.Is(err, domain.ErrIngredientsRequired):
			WriteError(w, http.StatusBadRequest, "ingredients_required")
		case errors.Is(err, domain.ErrInstructionsRequired):
			WriteError(w, http.StatusBadRequest, "instructions_required")
		case errors.Is(err, domain.ErrNameTooLong):
			WriteError(w, http.StatusBadRequest, "name_too_long")
		case errors.Is(err, domain.ErrTooManyIngredients):
			WriteError(w, http.StatusBadRequest, "too_many_ingredients")
		default:
			log.Printf("ERROR [UpdateRecipe] %v", err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	WriteJSON(w, http.StatusOK, recipe)
}

func (h *RecipeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !ValidateID(w, id, "recipe_id") {
		return
	}

	householdID := middleware.GetHouseholdID(r)

	err := h.recipeService.Delete(id, householdID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			WriteError(w, http.StatusNotFound, "not_found")
		case errors.Is(err, domain.ErrForbidden):
			WriteError(w, http.StatusForbidden, "forbidden")
		default:
			log.Printf("ERROR [DeleteRecipe] %v", err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RecipeHandler) Create(w http.ResponseWriter, r *http.Request) {
	householdID := middleware.GetHouseholdID(r)

	var req domain.CreateRecipeRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	resp, err := h.recipeService.Create(req, householdID)
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
		case errors.Is(err, domain.ErrNameTooLong):
			WriteError(w, http.StatusBadRequest, "name_too_long")
		case errors.Is(err, domain.ErrTooManyIngredients):
			WriteError(w, http.StatusBadRequest, "too_many_ingredients")
		default:
			log.Printf("ERROR [CreateRecipe] %v", err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	WriteJSON(w, http.StatusCreated, resp)
}
