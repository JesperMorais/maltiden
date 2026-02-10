package services

import (
	"maltiden/internal/domain"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type MenuService struct {
	menuStorage   domain.MenuRepository
	recipeStorage domain.RecipeRepository
}

func NewMenuService(menuStorage domain.MenuRepository, recipeStorage domain.RecipeRepository) *MenuService {
	return &MenuService{
		menuStorage:   menuStorage,
		recipeStorage: recipeStorage,
	}
}

func (s *MenuService) Generate(householdID string, req domain.GenerateMenuRequest) (*domain.MenuResponse, error) {
	// Validate and default days (VALID-13)
	days := req.Days
	if days == 0 {
		days = 5 // backwards-compatible default
	}
	if days < 1 || days > 31 {
		return nil, domain.ErrInvalidDays
	}

	// Validate and default servings (VALID-14)
	servings := req.Servings
	if servings == 0 {
		servings = 4 // backwards-compatible default
	}
	if servings < 1 || servings > 100 {
		return nil, domain.ErrInvalidServings
	}

	// Get all recipes
	recipes, err := s.recipeStorage.GetAll(nil)
	if err != nil {
		return nil, err
	}

	if len(recipes) == 0 {
		return nil, nil
	}

	// Build skip days map
	skipDays := make(map[string]bool)
	for _, d := range req.SkipDays {
		skipDays[d] = true
	}

	// Generate menu days
	menuDays := make([]domain.MenuDay, 0, days)
	today := time.Now()

	for i := 0; i < days; i++ {
		date := today.AddDate(0, 0, i).Format("2006-01-02")

		day := domain.MenuDay{
			Date:     date,
			Servings: servings,
		}

		// Check if this day should be skipped
		if skipDays[date] {
			day.Skip = true
		} else {
			// Pick a random recipe
			recipe := recipes[rand.Intn(len(recipes))]
			day.RecipeID = recipe.ID

			// Check for extra portions
			if extra, ok := req.ExtraPortions[date]; ok {
				day.Servings = servings + extra
			}
		}

		menuDays = append(menuDays, day)
	}

	// Create menu
	menu := &domain.Menu{
		ID:          "menu_" + uuid.New().String(),
		HouseholdID: householdID,
		Days:        menuDays,
		CreatedAt:   time.Now(),
	}

	if err := s.menuStorage.Create(menu); err != nil {
		return nil, err
	}

	return &domain.MenuResponse{
		ID:   menu.ID,
		Days: menu.Days,
	}, nil
}

func (s *MenuService) GetCurrent(householdID string) (*domain.MenuResponse, error) {
	menu, err := s.menuStorage.GetCurrentByHousehold(householdID)
	if err != nil {
		return nil, err
	}

	if menu == nil {
		return nil, nil
	}

	return &domain.MenuResponse{
		ID:   menu.ID,
		Days: menu.Days,
	}, nil
}
