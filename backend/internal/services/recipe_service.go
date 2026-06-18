package services

import (
	"maltiden/internal/domain"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
)

func validateRecipeContent(name string, tags []string, ingredients []domain.Ingredient, instructions []string) error {
	if err := domain.ValidateContent(name); err != nil {
		return err
	}
	for _, tag := range tags {
		if err := domain.ValidateContent(tag); err != nil {
			return err
		}
	}
	for _, ing := range ingredients {
		if err := domain.ValidateContent(ing.Name); err != nil {
			return err
		}
		if err := domain.ValidateContent(ing.Unit); err != nil {
			return err
		}
	}
	for _, step := range instructions {
		if err := domain.ValidateContent(step); err != nil {
			return err
		}
	}
	return nil
}

// maxCookMinutes bounds the cook_minutes metadata to a sane upper limit (24h).
const maxCookMinutes = 1440

// validateRecipeMetadata validates the optional Phase-0 metadata fields shared
// by create and update. Empty/zero values mean "unknown" and are accepted.
func validateRecipeMetadata(dietClass string, cookMinutes int, ingredients []domain.Ingredient) error {
	if !domain.IsValidDietClass(dietClass) {
		return domain.ErrInvalidDietClass
	}
	if cookMinutes < 0 || cookMinutes > maxCookMinutes {
		return domain.ErrInvalidCookMinutes
	}
	for _, ing := range ingredients {
		if ing.GramsEquiv < 0 {
			return domain.ErrInvalidGramsEquiv
		}
		if utf8.RuneCountInString(ing.CanonicalName) > 80 {
			return domain.ErrIngredientNameTooLong
		}
		if err := domain.ValidateContent(ing.CanonicalName); err != nil {
			return err
		}
	}
	return nil
}

func stripRecipeEmoji(r *domain.Recipe) {
	r.Name = domain.StripEmoji(r.Name)
	for i, tag := range r.Tags {
		r.Tags[i] = domain.StripEmoji(tag)
	}
	for i := range r.Ingredients {
		r.Ingredients[i].Name = domain.StripEmoji(r.Ingredients[i].Name)
		r.Ingredients[i].Unit = domain.StripEmoji(r.Ingredients[i].Unit)
	}
	for i, step := range r.Instructions {
		r.Instructions[i] = domain.StripEmoji(step)
	}
}

// normalizeTags trims whitespace, drops empty strings, and dedupes
// case-insensitively while preserving the first-seen casing and order.
func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return tags
	}
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		trimmed := strings.TrimSpace(t)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

type RecipeService struct {
	recipeStorage    domain.RecipeRepository
	livsmedelStorage domain.LivsmedelRepository // nullable: nil disables macro recompute (graceful)
}

func NewRecipeService(recipeStorage domain.RecipeRepository, livsmedelStorage domain.LivsmedelRepository) *RecipeService {
	return &RecipeService{recipeStorage: recipeStorage, livsmedelStorage: livsmedelStorage}
}

// recomputeMacros recomputes and sets r.Macros from the recipe's matched
// livsmedel. It is a graceful no-op when the livsmedel repo is unavailable, no
// ingredient is matched, or the lookup fails — macros are an optimization, not a
// correctness requirement, so they never block a write.
func (s *RecipeService) recomputeMacros(r *domain.Recipe) {
	if s.livsmedelStorage == nil {
		return
	}
	seen := make(map[int]struct{})
	numbers := make([]int, 0, len(r.Ingredients))
	for _, ing := range r.Ingredients {
		if ing.Livsmedelsnummer == 0 {
			continue
		}
		if _, ok := seen[ing.Livsmedelsnummer]; ok {
			continue
		}
		seen[ing.Livsmedelsnummer] = struct{}{}
		numbers = append(numbers, ing.Livsmedelsnummer)
	}
	if len(numbers) == 0 {
		r.Macros = nil
		return
	}
	table, err := s.livsmedelStorage.GetByNumbers(numbers)
	if err != nil {
		sentry.CaptureException(err)
		return
	}
	r.Macros = domain.ComputeRecipeMacros(*r, table)
}

// recipeHasComputableMacros reports whether any ingredient can actually
// contribute to a macro computation — i.e. it carries both a livsmedelsnummer
// and a positive grams-equivalent, matching ComputeRecipeMacros's contribution
// rule. Gating the read-time recompute on this (rather than on the number
// alone) avoids re-running a lookup every read for a matched-but-zero-grams
// recipe whose macros are legitimately nil.
func recipeHasComputableMacros(r *domain.Recipe) bool {
	for _, ing := range r.Ingredients {
		if ing.Livsmedelsnummer != 0 && ing.GramsEquiv > 0 {
			return true
		}
	}
	return false
}

func (s *RecipeService) GetAll(filter *domain.RecipeFilter, householdID string) ([]domain.RecipeSummary, error) {
	return s.recipeStorage.GetAll(filter, householdID)
}

func (s *RecipeService) GetByID(id string) (*domain.Recipe, error) {
	r, err := s.recipeStorage.GetByID(id)
	if err != nil || r == nil {
		return r, err
	}
	// Read-time fallback: if macros are absent but the recipe carries matched
	// ingredients (e.g. hand-edited after a match), recompute them so the response
	// stays correct. Persisted macros are preferred on the hot path; this only
	// fills the gap.
	if r.Macros == nil && recipeHasComputableMacros(r) {
		s.recomputeMacros(r)
	}
	return r, nil
}

// checkOwnership verifies the requesting household can modify this recipe.
// Seed recipes (empty HouseholdID) cannot be modified by anyone.
func checkOwnership(recipe *domain.Recipe, householdID string) error {
	if recipe.HouseholdID == "" {
		return domain.ErrForbidden // seed recipe
	}
	if recipe.HouseholdID != householdID {
		return domain.ErrForbidden // belongs to different household
	}
	return nil
}

func (s *RecipeService) Update(id string, householdID string, req domain.UpdateRecipeRequest) (*domain.Recipe, error) {
	// Verify recipe exists
	existing, err := s.recipeStorage.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrNotFound
	}

	// Check ownership
	if err := checkOwnership(existing, householdID); err != nil {
		return nil, err
	}

	// Validate fields (same rules as Create)
	if req.Name == "" {
		return nil, domain.ErrNameRequired
	}
	if utf8.RuneCountInString(req.Name) > 200 {
		return nil, domain.ErrNameTooLong
	}
	if req.Servings <= 0 || req.Servings > 100 {
		return nil, domain.ErrInvalidServings
	}
	if len(req.Ingredients) == 0 {
		return nil, domain.ErrIngredientsRequired
	}
	if len(req.Ingredients) > 50 {
		return nil, domain.ErrTooManyIngredients
	}
	if len(req.Instructions) == 0 {
		return nil, domain.ErrInstructionsRequired
	}
	// Tags: normalize (trim, drop empties, dedup), then enforce limits
	// (max 20 tags, each tag max 50 runes — Swedish chars like å/ä/ö are
	// 2 UTF-8 bytes, so byte-counting truncates valid tags).
	normalizedTags := normalizeTags(req.Tags)
	if len(normalizedTags) > 20 {
		return nil, domain.ErrTooManyTags
	}
	for _, tag := range normalizedTags {
		if utf8.RuneCountInString(tag) > 50 {
			return nil, domain.ErrTagTooLong
		}
	}

	if err := validateRecipeContent(req.Name, normalizedTags, req.Ingredients, req.Instructions); err != nil {
		return nil, err
	}

	if err := validateRecipeMetadata(req.DietClass, req.CookMinutes, req.Ingredients); err != nil {
		return nil, err
	}

	// Update fields on existing recipe
	existing.Name = req.Name
	existing.Servings = req.Servings
	existing.Emoji = req.Emoji
	existing.Tags = normalizedTags
	existing.Ingredients = req.Ingredients
	existing.Instructions = req.Instructions
	existing.MainProtein = req.MainProtein
	existing.DietClass = req.DietClass
	existing.Batchable = req.Batchable
	existing.CookMinutes = req.CookMinutes

	if existing.Tags == nil {
		existing.Tags = []string{}
	}

	stripRecipeEmoji(existing)

	// Recompute persisted macros from the (possibly changed) matched ingredients.
	s.recomputeMacros(existing)

	if err := s.recipeStorage.Update(existing); err != nil {
		sentry.CaptureException(err)
		return nil, err
	}

	return existing, nil
}

func (s *RecipeService) Delete(id string, householdID string) error {
	// Verify recipe exists
	existing, err := s.recipeStorage.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrNotFound
	}

	// Check ownership
	if err := checkOwnership(existing, householdID); err != nil {
		return err
	}

	return s.recipeStorage.Delete(id)
}

func (s *RecipeService) Create(req domain.CreateRecipeRequest, householdID string) (*domain.CreateRecipeResponse, error) {
	if req.Name == "" {
		return nil, domain.ErrNameRequired
	}
	// VALID-10: name length upper bound (rune count, not bytes)
	if utf8.RuneCountInString(req.Name) > 200 {
		return nil, domain.ErrNameTooLong
	}
	if req.Servings <= 0 {
		return nil, domain.ErrInvalidServings
	}
	// VALID-11: servings upper bound
	if req.Servings > 100 {
		return nil, domain.ErrInvalidServings
	}
	if len(req.Ingredients) == 0 {
		return nil, domain.ErrIngredientsRequired
	}
	// VALID-12: ingredients array size upper bound
	if len(req.Ingredients) > 50 {
		return nil, domain.ErrTooManyIngredients
	}
	if len(req.Instructions) == 0 {
		return nil, domain.ErrInstructionsRequired
	}
	// Tags: normalize (trim, drop empties, dedup), then enforce limits
	// (max 20 tags, each tag max 50 runes — Swedish chars like å/ä/ö are
	// 2 UTF-8 bytes, so byte-counting truncates valid tags).
	normalizedTags := normalizeTags(req.Tags)
	if len(normalizedTags) > 20 {
		return nil, domain.ErrTooManyTags
	}
	for _, tag := range normalizedTags {
		if utf8.RuneCountInString(tag) > 50 {
			return nil, domain.ErrTagTooLong
		}
	}

	if err := validateRecipeContent(req.Name, normalizedTags, req.Ingredients, req.Instructions); err != nil {
		return nil, err
	}

	if err := validateRecipeMetadata(req.DietClass, req.CookMinutes, req.Ingredients); err != nil {
		return nil, err
	}

	recipe := &domain.Recipe{
		ID:           "rec_" + uuid.New().String(),
		Name:         req.Name,
		Servings:     req.Servings,
		Emoji:        req.Emoji,
		Tags:         normalizedTags,
		Ingredients:  req.Ingredients,
		Instructions: req.Instructions,
		HouseholdID:  householdID,
		CreatedAt:    time.Now(),
		MainProtein:  req.MainProtein,
		DietClass:    req.DietClass,
		Batchable:    req.Batchable,
		CookMinutes:  req.CookMinutes,
	}

	if recipe.Tags == nil {
		recipe.Tags = []string{}
	}

	stripRecipeEmoji(recipe)

	// Compute persisted macros from any matched ingredients (graceful no-op when
	// none are matched or the livsmedel repo is unavailable).
	s.recomputeMacros(recipe)

	if err := s.recipeStorage.Create(recipe); err != nil {
		sentry.CaptureException(err)
		return nil, err
	}

	return &domain.CreateRecipeResponse{ID: recipe.ID}, nil
}
