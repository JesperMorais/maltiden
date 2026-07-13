package services

import (
	"maltiden/internal/domain"
	"time"

	"github.com/google/uuid"
)

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
	if len(req.Name) > 200 {
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
	// Tags: max 20 tags, each tag max 50 chars
	if len(req.Tags) > 20 {
		return nil, domain.ErrTooManyTags
	}
	for _, tag := range req.Tags {
		if len(tag) > 50 {
			return nil, domain.ErrTagTooLong
		}
	}
	for _, ing := range req.Ingredients {
		if ing.Amount < 0 {
			return nil, domain.ErrInvalidAmount
		}
		if ing.Amount > 10000 {
			return nil, domain.ErrAmountTooLarge
		}
	}

	// Update fields on existing recipe
	existing.Name = req.Name
	existing.Servings = req.Servings
	existing.Emoji = req.Emoji
	existing.Tags = req.Tags
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
	// VALID-10: name length upper bound
	if len(req.Name) > 200 {
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
	// Tags: max 20 tags, each tag max 50 chars
	if len(req.Tags) > 20 {
		return nil, domain.ErrTooManyTags
	}
	for _, tag := range req.Tags {
		if len(tag) > 50 {
			return nil, domain.ErrTagTooLong
		}
	}
	for _, ing := range req.Ingredients {
		if ing.Amount < 0 {
			return nil, domain.ErrInvalidAmount
		}
		if ing.Amount > 10000 {
			return nil, domain.ErrAmountTooLarge
		}
	}

	recipe := &domain.Recipe{
		ID:           "rec_" + uuid.New().String(),
		Name:         req.Name,
		Servings:     req.Servings,
		Emoji:        req.Emoji,
		Tags:         req.Tags,
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
