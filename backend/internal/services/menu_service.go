package services

import (
	"context"
	"log"
	"maltiden/internal/domain"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
)

type MenuService struct {
	menuStorage   domain.MenuRepository
	recipeStorage domain.RecipeRepository
	prefsStorage  domain.MenuPreferencesRepository

	// experience is the optional Phase-4 AI layer (Gemini). It is nil when no
	// API key is configured; every use is guarded so generation degrades to the
	// deterministic Phase 1–3 result and makes zero network calls by default.
	experience *MenuExperienceService
}

// experienceCallTimeout bounds each opt-in AI experience call. It is kept well
// under the generate route's timeout so a slow or hung Gemini call fails fast to
// the deterministic fallback rather than letting the HTTP route time out — which
// would otherwise return an error to the client while the menu was still being
// built and persisted (a "ghost menu"). The gemini client honors the context
// deadline, so this actually cancels the in-flight request.
const experienceCallTimeout = 20 * time.Second

func NewMenuService(menuStorage domain.MenuRepository, recipeStorage domain.RecipeRepository, prefsStorage domain.MenuPreferencesRepository) *MenuService {
	return &MenuService{
		menuStorage:   menuStorage,
		recipeStorage: recipeStorage,
		prefsStorage:  prefsStorage,
	}
}

// WithExperience injects the optional Phase-4 AI experience layer. A nil service
// (the default) keeps generation fully deterministic. Returns the receiver for
// fluent setup.
func (s *MenuService) WithExperience(e *MenuExperienceService) *MenuService {
	s.experience = e
	return s
}

// mergeWishConstraints folds non-nil parsed wish constraints into a household's
// effective preferences (and the prep-mode request flag) for one generation. It
// runs BEFORE the explicit request day/serving defaulting, by writing into the
// preference fallbacks, so a typed request value always wins over a wish. String
// lists are appended and case-insensitively deduped onto the existing lists. A
// prep-mode wish can only enable batch cooking, never disable an explicit one.
func mergeWishConstraints(prefs *domain.MenuPreferences, req *domain.GenerateMenuRequest, pc *domain.ParsedWishConstraints) {
	if pc == nil {
		return
	}
	if pc.VegetarianDays != nil {
		prefs.VegetarianDays = *pc.VegetarianDays
	}
	if pc.Days != nil {
		prefs.DefaultDays = *pc.Days
	}
	if pc.Servings != nil {
		prefs.DefaultServings = *pc.Servings
	}
	if pc.PrepMode != nil && *pc.PrepMode {
		req.PrepMode = true
	}
	prefs.ExcludedTags = appendDedupeFold(prefs.ExcludedTags, pc.ExtraExcludedTags)
	prefs.DislikedIngredients = appendDedupeFold(prefs.DislikedIngredients, pc.ExtraDislikedIngredients)
}

// appendDedupeFold appends extra items to base, skipping case-insensitive
// duplicates of items already present (in base or earlier in extra).
func appendDedupeFold(base, extra []string) []string {
	if len(extra) == 0 {
		return base
	}
	seen := make(map[string]bool, len(base)+len(extra))
	for _, b := range base {
		seen[strings.ToLower(strings.TrimSpace(b))] = true
	}
	out := base
	for _, e := range extra {
		key := strings.ToLower(strings.TrimSpace(e))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, e)
	}
	return out
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
		DislikedIngredients: disliked,
		DefaultDays:         req.DefaultDays,
		DefaultServings:     req.DefaultServings,
		VegetarianDays:      req.VegetarianDays,
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
			Date:       d.Date,
			RecipeID:   d.RecipeID,
			Servings:   d.Servings,
			Skip:       d.Skip,
			PrepMode:   d.PrepMode,
			LeftoverOf: d.LeftoverOf,
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

	// Opt-in wish parsing (Phase 4): free-text Swedish wishes become per-run
	// constraint overrides merged into the effective prefs BEFORE the day/serving
	// defaulting below, so a typed request field still wins. This only runs when
	// the request carries wishes; with no experience layer the wishes are
	// reported as ignored and generation proceeds deterministically.
	wishesIgnored := false
	if strings.TrimSpace(req.Wishes) != "" {
		if s.experience != nil {
			ctx, cancel := context.WithTimeout(context.Background(), experienceCallTimeout)
			pc, err := s.experience.ParseWishes(ctx, req.Wishes)
			cancel()
			if err != nil {
				sentry.CaptureException(err)
				log.Printf("WARN [Generate] wish parse failed, ignoring wishes: %v", err)
				wishesIgnored = true
			} else {
				pc.Clamp()
				mergeWishConstraints(&prefs, &req, pc)
			}
		} else {
			wishesIgnored = true
		}
	}

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

	// Weekday arrangement + rationale (Phase 4): when requested and the
	// experience layer is available, ask it to permute the already-chosen
	// recipes across the week's weekdays and write a short Swedish rationale.
	// Runs BEFORE batch cooking so leftover pairing respects the final
	// placement. The arranger is guard-railed to a strict permutation of the
	// offered slots/recipes; on any failure the deterministic order is kept and
	// the rationale stays empty (graceful degradation).
	rationale := ""
	if req.Arrange && s.experience != nil {
		nameByID := make(map[string]string, len(recipes))
		for _, r := range recipes {
			nameByID[r.ID] = r.Name
		}
		slots := make([]ArrangeSlot, 0, len(menuDays))
		for i, d := range menuDays {
			if d.Skip || d.RecipeID == "" {
				continue
			}
			weekday := time.Sunday
			if t, err := time.Parse("2006-01-02", d.Date); err == nil {
				weekday = t.Weekday()
			}
			slots = append(slots, ArrangeSlot{
				Slot:     i,
				Weekday:  int(weekday),
				RecipeID: d.RecipeID,
				Name:     nameByID[d.RecipeID],
			})
		}
		if len(slots) > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), experienceCallTimeout)
			assignment, r, err := s.experience.ArrangeWeek(ctx, slots)
			cancel()
			if err != nil {
				sentry.CaptureException(err)
				log.Printf("WARN [Generate] arrange failed, keeping deterministic order: %v", err)
			} else {
				for slot, recipeID := range assignment {
					menuDays[slot].RecipeID = recipeID
				}
				rationale = r
			}
		}
	}

	// Batch cooking (#248 Phase 2): when prep mode is on, pair each batchable
	// recipe's cook-day with a later leftovers day (cook once, eat twice). This
	// is a deterministic post-pass that never alters which recipes the selector
	// chose, so the prefs / overlap / variety / disliked / veg-quota chain is
	// untouched — it only relates two already-placed slots.
	if req.PrepMode {
		batchableIDs := make(map[string]bool)
		for _, r := range recipes {
			if isBatchable(r) {
				batchableIDs[r.ID] = true
			}
		}
		menuDays = applyBatchCooking(menuDays, batchableIDs)
	}

	// Create menu
	menu := &domain.Menu{
		ID:          "menu_" + uuid.New().String(),
		HouseholdID: householdID,
		Days:        menuDays,
		CreatedAt:   time.Now(),
		Rationale:   rationale,
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
		Rationale:         menu.Rationale,
		WishesIgnored:     wishesIgnored,
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
		ID:        menu.ID,
		Days:      enrichedDays,
		Rationale: menu.Rationale,
	}, nil
}
