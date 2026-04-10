package handlers

import (
	"log/slog"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"net/http"
)

type RecipeParserHandler struct {
	parserService *services.RecipeParserService
	recipeService *services.RecipeService
}

func NewRecipeParserHandler(
	parserService *services.RecipeParserService,
	recipeService *services.RecipeService,
) *RecipeParserHandler {
	return &RecipeParserHandler{
		parserService: parserService,
		recipeService: recipeService,
	}
}

func (h *RecipeParserHandler) ParseRecipe(w http.ResponseWriter, r *http.Request) {
	var req domain.ParseRecipeRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	if req.RawText == "" {
		WriteError(w, http.StatusBadRequest, "raw_text_required")
		return
	}

	if len(req.RawText) > 10000 {
		WriteError(w, http.StatusBadRequest, "input_too_long")
		return
	}

	result, err := h.parserService.ParseRecipe(req.RawText)
	if err != nil {
		slog.Error("ParseRecipe failed", "error", err)
		WriteError(w, http.StatusInternalServerError, "parse_failed")
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (h *RecipeParserHandler) ParseAndSave(w http.ResponseWriter, r *http.Request) {
	var req domain.ParseRecipeRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	if req.RawText == "" {
		WriteError(w, http.StatusBadRequest, "raw_text_required")
		return
	}

	if len(req.RawText) > 10000 {
		WriteError(w, http.StatusBadRequest, "input_too_long")
		return
	}

	parsed, err := h.parserService.ParseRecipe(req.RawText)
	if err != nil {
		slog.Error("ParseAndSave parse failed", "error", err)
		WriteError(w, http.StatusInternalServerError, "parse_failed")
		return
	}

	created, err := h.recipeService.Create(parsed.Recipe)
	if err != nil {
		slog.Error("ParseAndSave save failed", "error", err)
		WriteError(w, http.StatusInternalServerError, "save_failed")
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         created.ID,
		"recipe":     parsed.Recipe,
		"confidence": parsed.Confidence,
		"warnings":   parsed.Warnings,
	})
}
