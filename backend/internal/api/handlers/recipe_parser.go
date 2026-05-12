package handlers

import (
	"errors"
	"log/slog"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/pkg/middleware"
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
		if isContentSentinel(err) {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
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
		if isContentSentinel(err) {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		slog.Error("ParseAndSave parse failed", "error", err)
		WriteError(w, http.StatusInternalServerError, "parse_failed")
		return
	}

	householdID := middleware.GetHouseholdID(r)
	created, err := h.recipeService.Create(parsed.Recipe, householdID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "save_failed")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":         created.ID,
		"recipe":     parsed.Recipe,
		"confidence": parsed.Confidence,
		"warnings":   parsed.Warnings,
	})
}

func isContentSentinel(err error) bool {
	return errors.Is(err, domain.ErrContainsHTML) ||
		errors.Is(err, domain.ErrContainsURL) ||
		errors.Is(err, domain.ErrContainsMarkdown) ||
		errors.Is(err, domain.ErrContainsScraping) ||
		errors.Is(err, domain.ErrContainsProfanity) ||
		errors.Is(err, domain.ErrContainsInjection)
}
