package services

import (
	"errors"
	"maltiden/internal/domain"
	"maltiden/internal/storage/sqlite"
	"time"

	"github.com/google/uuid"
)

type RecipeService struct {
	recipeStorage *sqlite.RecipeStorage
}

func NewRecipeService(recipeStorage *sqlite.RecipeStorage) *RecipeService {
	return &RecipeService{recipeStorage: recipeStorage}
}

func (s *RecipeService) GetAll(filter *domain.RecipeFilter) ([]domain.RecipeSummary, error) {
	return s.recipeStorage.GetAll(filter)
}

func (s *RecipeService) GetByID(id string) (*domain.Recipe, error) {
	return s.recipeStorage.GetByID(id)
}

func (s *RecipeService) Create(req domain.CreateRecipeRequest) (*domain.CreateRecipeResponse, error) {
	if req.Name == "" {
		return nil, errors.New("name_required")
	}
	if req.Servings <= 0 {
		return nil, errors.New("invalid_servings")
	}
	if len(req.Ingredients) == 0 {
		return nil, errors.New("ingredients_required")
	}
	if len(req.Instructions) == 0 {
		return nil, errors.New("instructions_required")
	}

	recipe := &domain.Recipe{
		ID:           "rec_" + uuid.New().String(),
		Name:         req.Name,
		Servings:     req.Servings,
		Emoji:        req.Emoji,
		Tags:         req.Tags,
		Ingredients:  req.Ingredients,
		Instructions: req.Instructions,
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
