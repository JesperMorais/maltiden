package services

import (
	"maltiden/internal/domain"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
)

type MenuService struct {
	menuStorage   domain.MenuRepository
	recipeStorage domain.RecipeRepository
	prefsStorage  domain.MenuPreferencesRepository
}

func NewMenuService(menuStorage domain.MenuRepository, recipeStorage domain.RecipeRepository, prefsStorage domain.MenuPreferencesRepository) *MenuService {
	return &MenuService{
		menuStorage:   menuStorage,
		recipeStorage: recipeStorage,
		prefsStorage:  prefsStorage,
	}
}

// effectivePreferences loads a household's saved menu preferences, falling back
// to the defaults when none exist. A nil prefsStorage (or a load error) also
// yields defaults so generation never hard-fails on the preferences layer.
func (s *MenuService) effectivePreferences(householdID string) domain.MenuPreferences {
	if s.prefsStorage == nil {
		return domain.DefaultMenuPreferences(householdID)
	}
	prefs, err := s.prefsStorage.Get(householdID)
	if err != nil || prefs == nil {
		return domain.DefaultMenuPreferences(householdID)
	}
	return *prefs
}

// GetPreferences returns the household's saved menu preferences, or the
// defaults if none have been saved yet.
func (s *MenuService) GetPreferences(householdID string) (*domain.MenuPreferences, error) {
	if s.prefsStorage == nil {
		def := domain.DefaultMenuPreferences(householdID)
		return &def, nil
	}
	prefs, err := s.prefsStorage.Get(householdID)
	if err != nil {
		return nil, err
	}
	if prefs == nil {
		def := domain.DefaultMenuPreferences(householdID)
		return &def, nil
	}
	return prefs, nil
}

// UpdatePreferences validates and upserts a household's menu preferences,
// returning the stored result.
func (s *MenuService) UpdatePreferences(householdID string, req domain.UpdateMenuPreferencesRequest) (*domain.MenuPreferences, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	tags := req.ExcludedTags
	if tags == nil {
		tags = []string{}
	}
	prefs := &domain.MenuPreferences{
		HouseholdID:     householdID,
		ExcludedTags:    tags,
		DefaultDays:     req.DefaultDays,
		DefaultServings: req.DefaultServings,
		VegetarianDays:  req.VegetarianDays,
	}
	if err := s.prefsStorage.Upsert(prefs); err != nil {
		sentry.CaptureException(err)
		return nil, err
	}
	return prefs, nil
}

// loadIngredients batch-fetches the full recipes for the given summaries and
// returns a map of recipe ID -> ingredients, used by the selector for overlap
// scoring. A load error is non-fatal: it returns an empty map so generation
// proceeds with overlap scoring disabled rather than hard-failing.
func (s *MenuService) loadIngredients(recipes []domain.RecipeSummary) map[string][]domain.Ingredient {
	ids := make([]string, 0, len(recipes))
	for _, r := range recipes {
		if r.ID != "" {
			ids = append(ids, r.ID)
		}
	}
	out := make(map[string][]domain.Ingredient, len(ids))
	if len(ids) == 0 {
		return out
	}
	full, err := s.recipeStorage.GetByIDs(ids)
	if err != nil {
		sentry.CaptureException(err)
		return out
	}
	for id, r := range full {
		if r != nil {
			out[id] = r.Ingredients
		}
	}
	return out
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
	// Load the household's saved preferences (or defaults). These supply the
	// fallback day/serving counts and the excluded-tag / vegetarian-day rules
	// the selector core enforces.
	prefs := s.effectivePreferences(householdID)

	// Validate and default days (VALID-13). An omitted value falls back to the
	// household preference rather than a hardcoded week.
	days := req.Days
	if days == 0 {
		days = prefs.DefaultDays
	}
	if days < 1 || days > 31 {
		return nil, domain.ErrInvalidDays
	}

	// Validate and default servings (VALID-14).
	servings := req.Servings
	if servings == 0 {
		servings = prefs.DefaultServings
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

	// Fetch full recipes so the selector can score ingredient overlap. The
	// summaries from GetAll carry no ingredients, so we batch-load them by ID
	// (same access pattern as enrichMenuDays / the shopping list). On error we
	// degrade to overlap-less selection rather than failing generation.
	ingredientsByID := s.loadIngredients(recipes)

	// Build the selector core: hard-filters excluded tags, applies the
	// vegetarian-day preference, and greedily prefers high ingredient-overlap
	// dishes. If every recipe is filtered out, there is nothing to generate
	// from (treated the same as an empty catalog).
	selector, ok := newMenuSelector(recipes, prefs, nil, ingredientsByID)
	if !ok {
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
			// Pick the next recipe from the selector (preference-aware, cycling).
			day.RecipeID = selector.Next()

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
		sentry.CaptureException(err)
		return nil, err
	}

	enrichedDays, err := s.enrichMenuDays(menu.Days)
	if err != nil {
		return nil, err
	}

	// Surface the overlap economy to the user: which ingredients are reused
	// across the week's recipes. Reuses the ingredient map already loaded for
	// overlap scoring, so no extra storage round-trip.
	chosenIDs := make([]string, 0, len(menu.Days))
	for _, d := range menu.Days {
		if d.RecipeID != "" {
			chosenIDs = append(chosenIDs, d.RecipeID)
		}
	}

	return &domain.MenuResponse{
		ID:                menu.ID,
		Days:              enrichedDays,
		SharedIngredients: computeSharedIngredients(chosenIDs, ingredientsByID),
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
		sentry.CaptureException(err)
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
