package services

import (
	"maltiden/internal/domain"
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

// enrichMenuDays converts MenuDay slice to MenuResponseDay slice,
// populating recipeName and emoji from recipe storage.
func (s *MenuService) enrichMenuDays(days []domain.MenuDay) ([]domain.MenuResponseDay, error) {
	// Collect unique recipe IDs
	ids := make([]string, 0, len(days))
	for _, d := range days {
		if d.RecipeID != "" {
			ids = append(ids, d.RecipeID)
		}
	}

	// Batch-fetch recipes
	recipeMap := make(map[string]*domain.Recipe)
	if len(ids) > 0 {
		var err error
		recipeMap, err = s.recipeStorage.GetByIDs(ids)
		if err != nil {
			return nil, err
		}
	}

	result := make([]domain.MenuResponseDay, len(days))
	for i, d := range days {
		rd := domain.MenuResponseDay{
			Date:     d.Date,
			RecipeID: d.RecipeID,
			Servings: d.Servings,
			Skip:     d.Skip,
		}
		if r, ok := recipeMap[d.RecipeID]; ok {
			rd.RecipeName = r.Name
			rd.Emoji = r.Emoji
		}
		result[i] = rd
	}
	return result, nil
}

func (s *MenuService) Generate(householdID string, req domain.GenerateMenuRequest) (*domain.MenuResponse, error) {
	// Validate and default days (VALID-13)
	days := req.Days
	if days == 0 {
		days = 7 // default to a full Mon–Sun week
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

	// Get recipes visible to this household (own + seed)
	recipes, err := s.recipeStorage.GetAll(nil, householdID)
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

	today := time.Now()

	// Determine which slots are skipped and fetch full recipe data
	// (with ingredients) for the greedy overlap scorer.
	skipSlots := make(map[int]bool)
	for i := 0; i < days; i++ {
		date := today.AddDate(0, 0, i).Format("2006-01-02")
		if skipDays[date] {
			skipSlots[i] = true
		}
	}

	ids := make([]string, len(recipes))
	for i, r := range recipes {
		ids[i] = r.ID
	}
	recipeData, err := s.recipeStorage.GetByIDs(ids)
	if err != nil {
		return nil, err
	}

	selected := selectRecipesForWeek(recipes, days, skipSlots, recipeData, uint64(today.UnixNano()))

	// Generate menu days
	menuDays := make([]domain.MenuDay, 0, days)

	for i := 0; i < days; i++ {
		date := today.AddDate(0, 0, i).Format("2006-01-02")

		day := domain.MenuDay{
			Date:     date,
			Servings: servings,
		}

		if skipSlots[i] {
			day.Skip = true
		} else {
			day.RecipeID = selected[i]

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

	enrichedDays, err := s.enrichMenuDays(menu.Days)
	if err != nil {
		return nil, err
	}

	return &domain.MenuResponse{
		ID:   menu.ID,
		Days: enrichedDays,
	}, nil
}

func (s *MenuService) UpdateCurrent(householdID string, req domain.UpdateMenuRequest) (*domain.MenuResponse, error) {
	// Get current menu for this household
	menu, err := s.menuStorage.GetCurrentByHousehold(householdID)
	if err != nil {
		return nil, err
	}
	if menu == nil {
		return nil, domain.ErrNotFound
	}

	// Validate days
	if len(req.Days) == 0 || len(req.Days) > 31 {
		return nil, domain.ErrInvalidDays
	}

	// Update the menu days
	menu.Days = req.Days

	if err := s.menuStorage.Update(menu); err != nil {
		return nil, err
	}

	enrichedDays, err := s.enrichMenuDays(menu.Days)
	if err != nil {
		return nil, err
	}

	return &domain.MenuResponse{
		ID:   menu.ID,
		Days: enrichedDays,
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

	enrichedDays, err := s.enrichMenuDays(menu.Days)
	if err != nil {
		return nil, err
	}

	return &domain.MenuResponse{
		ID:   menu.ID,
		Days: enrichedDays,
	}, nil
}
