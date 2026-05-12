package services

import (
	"maltiden/internal/domain"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

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
	recipeStorage domain.RecipeRepository
}

func NewRecipeService(recipeStorage domain.RecipeRepository) *RecipeService {
	return &RecipeService{recipeStorage: recipeStorage}
}

func (s *RecipeService) GetAll(filter *domain.RecipeFilter, householdID string) ([]domain.RecipeSummary, error) {
	return s.recipeStorage.GetAll(filter, householdID)
}

func (s *RecipeService) GetByID(id string) (*domain.Recipe, error) {
	return s.recipeStorage.GetByID(id)
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
	stripUpdateEmoji(&req)
	normalizedTags = normalizeTags(req.Tags)

	// Update fields on existing recipe
	existing.Name = req.Name
	existing.Servings = req.Servings
	existing.Emoji = req.Emoji
	existing.Tags = normalizedTags
	existing.Ingredients = req.Ingredients
	existing.Instructions = req.Instructions

	if existing.Tags == nil {
		existing.Tags = []string{}
	}

	if err := s.recipeStorage.Update(existing); err != nil {
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
	stripCreateEmoji(&req)
	normalizedTags = normalizeTags(req.Tags)

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
	}

	if recipe.Tags == nil {
		recipe.Tags = []string{}
	}

	if err := s.recipeStorage.Create(recipe); err != nil {
		return nil, err
	}

	return &domain.CreateRecipeResponse{ID: recipe.ID}, nil
}

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

func stripCreateEmoji(req *domain.CreateRecipeRequest) {
	req.Name = domain.StripEmoji(req.Name)
	for i, tag := range req.Tags {
		req.Tags[i] = domain.StripEmoji(tag)
	}
	for i := range req.Ingredients {
		req.Ingredients[i].Name = domain.StripEmoji(req.Ingredients[i].Name)
		req.Ingredients[i].Unit = domain.StripEmoji(req.Ingredients[i].Unit)
	}
	for i, step := range req.Instructions {
		req.Instructions[i] = domain.StripEmoji(step)
	}
	// req.Emoji is intentionally untouched
}

func stripUpdateEmoji(req *domain.UpdateRecipeRequest) {
	req.Name = domain.StripEmoji(req.Name)
	for i, tag := range req.Tags {
		req.Tags[i] = domain.StripEmoji(tag)
	}
	for i := range req.Ingredients {
		req.Ingredients[i].Name = domain.StripEmoji(req.Ingredients[i].Name)
		req.Ingredients[i].Unit = domain.StripEmoji(req.Ingredients[i].Unit)
	}
	for i, step := range req.Instructions {
		req.Instructions[i] = domain.StripEmoji(step)
	}
	// req.Emoji is intentionally untouched
}
