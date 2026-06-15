package services

import (
	"log"
	"maltiden/internal/domain"
	"math/rand/v2"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
)

type MenuService struct {
	menuStorage   domain.MenuRepository
	recipeStorage domain.RecipeRepository
	prefsStorage  domain.MenuPreferencesRepository

	// newRNG produces the RNG seeded for one Generate run. The default seeds
	// from the wall clock so each generate (and regenerate) yields a different
	// valid week; tests override it to pin determinism.
	newRNG func() *rand.Rand
}

func NewMenuService(menuStorage domain.MenuRepository, recipeStorage domain.RecipeRepository, prefsStorage domain.MenuPreferencesRepository) *MenuService {
	return &MenuService{
		menuStorage:   menuStorage,
		recipeStorage: recipeStorage,
		prefsStorage:  prefsStorage,
		newRNG: func() *rand.Rand {
			seed := uint64(time.Now().UnixNano())
			return rand.New(rand.NewPCG(seed, seed>>1|1))
		},
	}
}

// WithRNG overrides the RNG factory so tests can pin the seed for deterministic
// menu generation. It returns the receiver for fluent setup.
func (s *MenuService) WithRNG(factory func() *rand.Rand) *MenuService {
	s.newRNG = factory
	return s
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
	disliked := req.DislikedIngredients
	if disliked == nil {
		disliked = []string{}
	}
	prefs := &domain.MenuPreferences{
		HouseholdID:         householdID,
		ExcludedTags:        tags,
		DefaultDays:         req.DefaultDays,
		DefaultServings:     req.DefaultServings,
		VegetarianDays:      req.VegetarianDays,
		DietProfile:         req.DietProfile,
		DislikedIngredients: disliked,
	}
	if err := s.prefsStorage.Upsert(prefs); err != nil {
		sentry.CaptureException(err)
		return nil, err
	}
	return prefs, nil
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

	// Get recipe summaries visible to this household (own + seed). Summaries
	// carry tags/metadata but NOT ingredients, so we collect the candidate ids
	// and re-fetch full recipes (with ingredients) in a single batch below —
	// two queries, no N+1, mirroring shopping_service.
	summaries, err := s.recipeStorage.GetAll(nil, householdID)
	if err != nil {
		return nil, err
	}

	if len(summaries) == 0 {
		return nil, nil
	}

	// Build the precomputed date list so locks/skips can be mapped to slots.
	today := time.Now()
	dates := make([]string, days)
	for i := 0; i < days; i++ {
		dates[i] = today.AddDate(0, 0, i).Format("2006-01-02")
	}

	// SECURITY: only recipe ids in THIS household's catalog (its scoped summary
	// set) may be fetched or pinned. We never fetch a locked id that is not in
	// the summary set — otherwise a client could pin (and have returned) another
	// household's private recipe via GetByIDs, which has no household filter.
	summarySet := make(map[string]bool, len(summaries))
	allIDs := make([]string, 0, len(summaries))
	for _, sum := range summaries {
		if !summarySet[sum.ID] {
			summarySet[sum.ID] = true
			allIDs = append(allIDs, sum.ID)
		}
	}

	// The scorer needs ingredients, so re-fetch full recipes for the household's
	// candidate set (one batch, no N+1, mirroring shopping_service).
	recipeMap, err := s.recipeStorage.GetByIDs(allIDs)
	if err != nil {
		return nil, err
	}

	// Candidate pool = full recipes for the summary set.
	candidates := make([]domain.Recipe, 0, len(summaries))
	for _, sum := range summaries {
		if r, ok := recipeMap[sum.ID]; ok {
			candidates = append(candidates, *r)
		}
	}

	// Recency: recipe ids used in the household's most recent menus. Recency is
	// an optimization, not a correctness requirement, so a storage error must
	// not abort generation — log it and proceed with an empty recency set.
	recency, err := s.menuStorage.GetRecentRecipeIDs(householdID, domain.RecencyWindowMenus)
	if err != nil {
		sentry.CaptureException(err)
		log.Printf("WARN [GenerateMenu] recency lookup failed, proceeding without: %v", err)
		recency = nil
	}

	// Build the selector core: hard-filters (excluded tags, diet profile,
	// disliked ingredients), then greedily builds the best-scoring week. If
	// every recipe is filtered out, there is nothing to generate from (treated
	// the same as an empty catalog).
	selector, ok := newMenuSelector(candidates, prefs, recency, s.newRNG())
	if !ok {
		return nil, nil
	}

	// Map skip days and locked days to slot indexes.
	skipDays := make(map[string]bool)
	for _, d := range req.SkipDays {
		skipDays[d] = true
	}
	skip := make([]bool, days)
	locked := make([]string, days)
	for i, date := range dates {
		if skipDays[date] {
			skip[i] = true
			continue
		}
		// Only honor a lock whose recipe is in this household's catalog. Unknown
		// or foreign ids are dropped silently (implements the documented
		// "unknown locked recipeId → treat as unlocked" and prevents leaking
		// another household's recipe).
		if recipeID, ok := req.LockedDays[date]; ok && summarySet[recipeID] {
			locked[i] = recipeID
		}
	}

	chosen := selector.SelectWeek(days, skip, locked)

	// Build menu days from the chosen week.
	menuDays := make([]domain.MenuDay, 0, days)
	for i, date := range dates {
		day := domain.MenuDay{
			Date:     date,
			Servings: servings,
		}
		if skip[i] {
			day.Skip = true
		} else {
			day.RecipeID = chosen[i]
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

	// Economy summary over the chosen (non-skip) recipes. Build a value map of
	// full recipes for the scorer.
	chosenIDs := make([]string, 0, days)
	for i := range menuDays {
		if !menuDays[i].Skip && menuDays[i].RecipeID != "" {
			chosenIDs = append(chosenIDs, menuDays[i].RecipeID)
		}
	}
	recipesByID := make(map[string]domain.Recipe, len(recipeMap))
	for id, r := range recipeMap {
		recipesByID[id] = *r
	}
	economy := computeMenuEconomy(chosenIDs, recipesByID)

	return &domain.MenuResponse{
		ID:      menu.ID,
		Days:    enrichedDays,
		Economy: economy,
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
