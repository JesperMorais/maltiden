package services

import (
	"maltiden/internal/domain"
	"maltiden/internal/storage/sqlite"
	"math/rand/v2"
	"strings"
	"testing"
	"time"
)

type menuTestEnv struct {
	menuService   *MenuService
	recipeService *RecipeService
	householdID   string
}

func newMenuTestEnv(t *testing.T) *menuTestEnv {
	t.Helper()
	db := setupTestDB(t)
	recipeStorage := sqlite.NewRecipeStorage(db)
	menuStorage := sqlite.NewMenuStorage(db)
	menuPrefsStorage := sqlite.NewMenuPreferencesStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	// Create a user (which creates a household) to satisfy FK constraint
	user := createTestUser(t, authService, "menu-test@test.com", "MenuTester")

	return &menuTestEnv{
		menuService:   NewMenuService(menuStorage, recipeStorage, menuPrefsStorage),
		recipeService: NewRecipeService(recipeStorage),
		householdID:   user.User.HouseholdID,
	}
}

func (env *menuTestEnv) seedRecipes(t *testing.T, count int) []string {
	t.Helper()
	ids := make([]string, count)
	for i := 0; i < count; i++ {
		resp, err := env.recipeService.Create(domain.CreateRecipeRequest{
			Name:         "Recipe " + strings.Repeat("X", i),
			Servings:     4,
			Ingredients:  []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "st"}},
			Instructions: []string{"Do something"},
		}, env.householdID)
		if err != nil {
			t.Fatalf("failed to seed recipe %d: %v", i, err)
		}
		ids[i] = resp.ID
	}
	return ids
}

// seedTaggedRecipe creates one recipe with the given name and tags, returning
// its ID. Used by the batch-cooking tests, which depend on the `batchcook` tag.
func (env *menuTestEnv) seedTaggedRecipe(t *testing.T, name string, tags ...string) string {
	t.Helper()
	resp, err := env.recipeService.Create(domain.CreateRecipeRequest{
		Name:         name,
		Servings:     4,
		Tags:         tags,
		Ingredients:  []domain.Ingredient{{Name: "Test", Amount: 1, Unit: "st"}},
		Instructions: []string{"Do something"},
	}, env.householdID)
	if err != nil {
		t.Fatalf("failed to seed tagged recipe %q: %v", name, err)
	}
	return resp.ID
}

func TestMenuGenerate_PrepModeOffNeverBatches(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedTaggedRecipe(t, "Batchgryta", "batchcook")
	env.seedTaggedRecipe(t, "Vanlig rätt")

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     5,
		Servings: 4,
		// PrepMode omitted (off): even with a batchable recipe present, no day
		// may be marked as a batch/leftovers day.
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, day := range resp.Days {
		if day.PrepMode != "" || day.LeftoverOf != "" {
			t.Errorf("day %d: prep mode off but got prepMode=%q leftoverOf=%q", i, day.PrepMode, day.LeftoverOf)
		}
		if day.Servings != 4 {
			t.Errorf("day %d: expected unchanged servings=4, got %d", i, day.Servings)
		}
	}
}

// batchStubRecipeRepo is a minimal RecipeRepository returning a fixed, fully
// controlled catalog, so the batch-cooking activation test is deterministic and
// isolated from the migration-seeded recipes the real DB carries.
type batchStubRecipeRepo struct {
	summaries []domain.RecipeSummary
}

func (r *batchStubRecipeRepo) GetAll(_ *domain.RecipeFilter, _ string) ([]domain.RecipeSummary, error) {
	return r.summaries, nil
}
func (r *batchStubRecipeRepo) GetAllPaginated(_ *domain.RecipeFilter, _ string, _, _ int) ([]domain.RecipeSummary, int, error) {
	return r.summaries, len(r.summaries), nil
}
func (r *batchStubRecipeRepo) GetByID(_ string) (*domain.Recipe, error) { return nil, nil }
func (r *batchStubRecipeRepo) GetByIDs(ids []string) (map[string]*domain.Recipe, error) {
	out := make(map[string]*domain.Recipe)
	for _, s := range r.summaries {
		out[s.ID] = &domain.Recipe{ID: s.ID, Name: s.Name, Servings: s.Servings, Tags: s.Tags}
	}
	return out, nil
}
func (r *batchStubRecipeRepo) Create(_ *domain.Recipe) error { return nil }
func (r *batchStubRecipeRepo) Update(_ *domain.Recipe) error { return nil }
func (r *batchStubRecipeRepo) Delete(_ string) error         { return nil }

// captureMenuRepo records the menu passed to Create so the test can assert what
// the service persisted, and serves it back from GetCurrentByHousehold.
type captureMenuRepo struct{ menu *domain.Menu }

func (m *captureMenuRepo) Create(menu *domain.Menu) error { m.menu = menu; return nil }
func (m *captureMenuRepo) Update(menu *domain.Menu) error { m.menu = menu; return nil }
func (m *captureMenuRepo) GetCurrentByHousehold(_ string) (*domain.Menu, error) {
	return m.menu, nil
}
func (m *captureMenuRepo) GetByID(_ string) (*domain.Menu, error)          { return m.menu, nil }
func (m *captureMenuRepo) GetHouseholdIDByMenuID(_ string) (string, error) { return "", nil }

func TestMenuGenerate_PrepModeBatchesCookOnceEatTwice(t *testing.T) {
	// Two batchable recipes only: cycling over a 4-day week guarantees a
	// batchable recipe lands on an early slot, and the post-pass must pair it
	// with a later day. Fully deterministic — no migration-seeded noise.
	recipes := []domain.RecipeSummary{
		{ID: "rec_batch_a", Name: "Batchgryta", Servings: 4, Tags: []string{"batchcook"}},
		{ID: "rec_batch_b", Name: "Köttfärssås", Servings: 4, Tags: []string{"batchcook"}},
	}
	menuRepo := &captureMenuRepo{}
	svc := NewMenuService(menuRepo, &batchStubRecipeRepo{summaries: recipes}, nil)

	resp, err := svc.Generate("hh_1", domain.GenerateMenuRequest{Days: 4, Servings: 4, PrepMode: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cookDays, leftoverDays := 0, 0
	byDate := make(map[string]domain.MenuResponseDay, len(resp.Days))
	for _, d := range resp.Days {
		byDate[d.Date] = d
		switch {
		case d.PrepMode == domain.PrepModeBatch:
			cookDays++
			if d.Servings != 8 {
				t.Errorf("batch cook-day %s: expected doubled servings=8, got %d", d.Date, d.Servings)
			}
		case d.LeftoverOf != "":
			leftoverDays++
			if d.RecipeID == "" {
				t.Errorf("leftovers day %s has no recipe", d.Date)
			}
		}
	}
	if cookDays == 0 {
		t.Fatal("expected at least one batch cook-day with prep mode on and a batchable recipe present")
	}
	if cookDays != leftoverDays {
		t.Errorf("cook-days (%d) and leftovers-days (%d) must pair up", cookDays, leftoverDays)
	}
	for _, d := range resp.Days {
		if d.LeftoverOf == "" {
			continue
		}
		cook, ok := byDate[d.LeftoverOf]
		if !ok {
			t.Errorf("leftovers day %s references missing cook-day %s", d.Date, d.LeftoverOf)
			continue
		}
		if cook.PrepMode != domain.PrepModeBatch {
			t.Errorf("cook-day %s referenced by leftovers %s is not a batch day", cook.Date, d.Date)
		}
		if cook.RecipeID != d.RecipeID {
			t.Errorf("leftovers %s recipe %q != cook-day %s recipe %q", d.Date, d.RecipeID, cook.Date, cook.RecipeID)
		}
	}
}

func TestMenuGenerate_BatchFieldsPersistAndRoundTrip(t *testing.T) {
	// Drive the real storage round-trip: save a menu carrying batch markers via
	// UpdateCurrent, then reload it. Proves migration 018's columns persist and
	// scan back. Deterministic (no selector involved).
	env := newMenuTestEnv(t)
	rid := env.seedTaggedRecipe(t, "Batchgryta", "batchcook")

	// Seed an initial menu so UpdateCurrent has something to update.
	if _, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{Days: 2, Servings: 4}); err != nil {
		t.Fatalf("seed menu: %v", err)
	}

	days := []domain.MenuDay{
		{Date: "2026-06-15", RecipeID: rid, Servings: 8, PrepMode: domain.PrepModeBatch},
		{Date: "2026-06-16", RecipeID: rid, Servings: 4, LeftoverOf: "2026-06-15"},
	}
	if _, err := env.menuService.UpdateCurrent(env.householdID, domain.UpdateMenuRequest{Days: days}); err != nil {
		t.Fatalf("update current: %v", err)
	}

	cur, err := env.menuService.GetCurrent(env.householdID)
	if err != nil {
		t.Fatalf("get current: %v", err)
	}
	if cur == nil || len(cur.Days) != 2 {
		t.Fatalf("reloaded menu mismatch: %+v", cur)
	}
	if cur.Days[0].PrepMode != domain.PrepModeBatch || cur.Days[0].Servings != 8 {
		t.Errorf("cook-day not persisted: %+v", cur.Days[0])
	}
	if cur.Days[1].LeftoverOf != "2026-06-15" {
		t.Errorf("leftoverOf not persisted: got %q", cur.Days[1].LeftoverOf)
	}
}

func TestMenuGenerate_DefaultDays(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipes(t, 7)

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     0, // should default to 7 (full Mon–Sun week)
		Servings: 4,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Days) != 7 {
		t.Errorf("expected 7 days (default), got %d", len(resp.Days))
	}
}

func TestMenuGenerate_DefaultServings(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipes(t, 3)

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     3,
		Servings: 0, // should default to 4
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, day := range resp.Days {
		if day.Servings != 4 {
			t.Errorf("day %d: expected servings=4 (default), got %d", i, day.Servings)
		}
	}
}

func TestMenuGenerate_InvalidDaysTooLow(t *testing.T) {
	env := newMenuTestEnv(t)

	_, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     -1,
		Servings: 4,
	})
	if err != domain.ErrInvalidDays {
		t.Errorf("expected ErrInvalidDays, got %v", err)
	}
}

func TestMenuGenerate_InvalidDaysTooHigh(t *testing.T) {
	env := newMenuTestEnv(t)

	_, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     32,
		Servings: 4,
	})
	if err != domain.ErrInvalidDays {
		t.Errorf("expected ErrInvalidDays, got %v", err)
	}
}

func TestMenuGenerate_InvalidServingsTooLow(t *testing.T) {
	env := newMenuTestEnv(t)

	_, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     5,
		Servings: -1,
	})
	if err != domain.ErrInvalidServings {
		t.Errorf("expected ErrInvalidServings, got %v", err)
	}
}

func TestMenuGenerate_InvalidServingsTooHigh(t *testing.T) {
	env := newMenuTestEnv(t)

	_, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     5,
		Servings: 101,
	})
	if err != domain.ErrInvalidServings {
		t.Errorf("expected ErrInvalidServings, got %v", err)
	}
}

func TestMenuGenerate_WithSeededRecipes(t *testing.T) {
	// Migrations seed 5 recipes, so even without adding more, Generate should succeed
	env := newMenuTestEnv(t)

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     3,
		Servings: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response with seeded recipes")
	}
	if len(resp.Days) != 3 {
		t.Errorf("expected 3 days, got %d", len(resp.Days))
	}
}

func TestMenuGenerate_Success(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipes(t, 5)

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     5,
		Servings: 4,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(resp.ID, "menu_") {
		t.Errorf("expected menu_ prefix, got %q", resp.ID)
	}
	if len(resp.Days) != 5 {
		t.Fatalf("expected 5 days, got %d", len(resp.Days))
	}

	for i, day := range resp.Days {
		if day.RecipeID == "" {
			t.Errorf("day %d: expected non-empty recipe ID", i)
		}
		if !strings.HasPrefix(day.RecipeID, "rec_") {
			t.Errorf("day %d: expected rec_ prefix, got %q", i, day.RecipeID)
		}
		if day.Servings != 4 {
			t.Errorf("day %d: expected servings=4, got %d", i, day.Servings)
		}
	}
}

func TestMenuGenerate_MoreDaysThanRecipes(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipes(t, 2)

	// 7 days with only 2 recipes should cycle without panicking
	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     7,
		Servings: 4,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Days) != 7 {
		t.Fatalf("expected 7 days, got %d", len(resp.Days))
	}
	for i, day := range resp.Days {
		if day.RecipeID == "" {
			t.Errorf("day %d: expected recipe ID even when cycling", i)
		}
	}
}

func TestMenuGenerate_SkipDays(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipes(t, 5)

	today := time.Now()
	skipDate := today.AddDate(0, 0, 1).Format("2006-01-02") // skip tomorrow

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     3,
		Servings: 4,
		SkipDays: []string{skipDate},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Days) != 3 {
		t.Fatalf("expected 3 days, got %d", len(resp.Days))
	}

	// Day 1 (tomorrow) should be skipped
	if !resp.Days[1].Skip {
		t.Error("expected day 1 to be skipped")
	}
	if resp.Days[1].RecipeID != "" {
		t.Errorf("skipped day should have no recipe, got %q", resp.Days[1].RecipeID)
	}

	// Other days should not be skipped
	if resp.Days[0].Skip {
		t.Error("day 0 should not be skipped")
	}
	if resp.Days[2].Skip {
		t.Error("day 2 should not be skipped")
	}
}

func TestMenuGenerate_ExtraPortions(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipes(t, 5)

	today := time.Now()
	extraDate := today.AddDate(0, 0, 2).Format("2006-01-02")

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:          3,
		Servings:      4,
		ExtraPortions: map[string]int{extraDate: 3},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Day 2 should have 4+3=7 servings
	if resp.Days[2].Servings != 7 {
		t.Errorf("expected 7 servings on extra-portions day, got %d", resp.Days[2].Servings)
	}
	// Other days should have base servings
	if resp.Days[0].Servings != 4 {
		t.Errorf("expected 4 servings on normal day, got %d", resp.Days[0].Servings)
	}
}

func TestMenuPreferences_DefaultsWhenUnsaved(t *testing.T) {
	env := newMenuTestEnv(t)

	prefs, err := env.menuService.GetPreferences(env.householdID)
	if err != nil {
		t.Fatalf("GetPreferences: %v", err)
	}
	if prefs.DefaultDays != 7 || prefs.DefaultServings != 4 {
		t.Errorf("expected default 7 days / 4 servings, got %+v", prefs)
	}
	if len(prefs.ExcludedTags) != 0 {
		t.Errorf("expected no excluded tags by default, got %v", prefs.ExcludedTags)
	}
}

func TestMenuPreferences_UpdateValidation(t *testing.T) {
	env := newMenuTestEnv(t)

	tests := []struct {
		name    string
		req     domain.UpdateMenuPreferencesRequest
		wantErr error
	}{
		{"days too low", domain.UpdateMenuPreferencesRequest{DefaultDays: 0, DefaultServings: 4}, domain.ErrInvalidDays},
		{"days too high", domain.UpdateMenuPreferencesRequest{DefaultDays: 32, DefaultServings: 4}, domain.ErrInvalidDays},
		{"servings too low", domain.UpdateMenuPreferencesRequest{DefaultDays: 7, DefaultServings: 0}, domain.ErrInvalidServings},
		{"veg days exceed total", domain.UpdateMenuPreferencesRequest{DefaultDays: 3, DefaultServings: 4, VegetarianDays: 4}, domain.ErrInvalidVegetarianDays},
		{"empty tag", domain.UpdateMenuPreferencesRequest{DefaultDays: 7, DefaultServings: 4, ExcludedTags: []string{""}}, domain.ErrInvalidExcludedTag},
		{"valid", domain.UpdateMenuPreferencesRequest{DefaultDays: 5, DefaultServings: 3, VegetarianDays: 1, ExcludedTags: []string{"fisk"}}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := env.menuService.UpdatePreferences(env.householdID, tt.req)
			if err != tt.wantErr {
				t.Errorf("UpdatePreferences() err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestMenuPreferences_UpsertRoundTrip(t *testing.T) {
	env := newMenuTestEnv(t)

	_, err := env.menuService.UpdatePreferences(env.householdID, domain.UpdateMenuPreferencesRequest{
		DefaultDays:     4,
		DefaultServings: 6,
		VegetarianDays:  2,
		ExcludedTags:    []string{"fisk"},
	})
	if err != nil {
		t.Fatalf("UpdatePreferences: %v", err)
	}

	got, err := env.menuService.GetPreferences(env.householdID)
	if err != nil {
		t.Fatalf("GetPreferences: %v", err)
	}
	if got.DefaultDays != 4 || got.DefaultServings != 6 || got.VegetarianDays != 2 {
		t.Errorf("round-trip scalar mismatch: %+v", got)
	}
}

func TestMenuGenerate_UsesPreferenceDefaults(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipes(t, 5)

	if _, err := env.menuService.UpdatePreferences(env.householdID, domain.UpdateMenuPreferencesRequest{
		DefaultDays:     4,
		DefaultServings: 6,
	}); err != nil {
		t.Fatalf("UpdatePreferences: %v", err)
	}

	// Request omits days+servings, so saved preferences should drive both.
	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(resp.Days) != 4 {
		t.Errorf("expected 4 days from preference, got %d", len(resp.Days))
	}
	for i, d := range resp.Days {
		if d.Servings != 6 {
			t.Errorf("day %d: expected 6 servings from preference, got %d", i, d.Servings)
		}
	}
}

func TestMenuGenerate_ExplicitRequestOverridesPreferences(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipes(t, 5)

	if _, err := env.menuService.UpdatePreferences(env.householdID, domain.UpdateMenuPreferencesRequest{
		DefaultDays:     4,
		DefaultServings: 6,
	}); err != nil {
		t.Fatalf("UpdatePreferences: %v", err)
	}

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     2,
		Servings: 8,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(resp.Days) != 2 {
		t.Errorf("explicit days should win: expected 2, got %d", len(resp.Days))
	}
	if resp.Days[0].Servings != 8 {
		t.Errorf("explicit servings should win: expected 8, got %d", resp.Days[0].Servings)
	}
}

func TestMenuGenerate_ExcludedTagsFilterRecipes(t *testing.T) {
	env := newMenuTestEnv(t)

	// Tag a recipe with a unique excluded tag so we can assert it is never
	// selected. (Seed recipes with NULL household_id are also visible, so we
	// check the excluded one is absent rather than asserting a single allowed ID.)
	excluded, err := env.recipeService.Create(domain.CreateRecipeRequest{
		Name:         "Fish Stew",
		Servings:     4,
		Tags:         []string{"mt-exclude-marker"},
		Ingredients:  []domain.Ingredient{{Name: "Torsk", Amount: 1, Unit: "st"}},
		Instructions: []string{"Cook"},
	}, env.householdID)
	if err != nil {
		t.Fatalf("create excluded recipe: %v", err)
	}

	if _, err := env.menuService.UpdatePreferences(env.householdID, domain.UpdateMenuPreferencesRequest{
		DefaultDays:     7,
		DefaultServings: 4,
		ExcludedTags:    []string{"mt-exclude-marker"},
	}); err != nil {
		t.Fatalf("UpdatePreferences: %v", err)
	}

	// Generate a long menu so cycling would surface the excluded recipe if the
	// filter were not applied.
	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{Days: 20, Servings: 4})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	for i, d := range resp.Days {
		if d.RecipeID == excluded.ID {
			t.Errorf("day %d selected excluded recipe %q", i, excluded.ID)
		}
	}
}

func TestMenuGenerate_PrefersIngredientOverlap(t *testing.T) {
	// End-to-end proof that the service can load real recipe ingredients from
	// storage (loadIngredients) and that feeding them to the selector makes the
	// greedy overlap step prefer the sharing recipe. We isolate the assertion to
	// our own three recipes by driving the selector directly — the same code
	// path Generate uses (loadIngredients + newMenuSelector) — rather than
	// fighting the 100+ seed recipes that also live in the household's pool.
	env := newMenuTestEnv(t)

	mk := func(name string, ings []domain.Ingredient) (string, domain.RecipeSummary) {
		r, err := env.recipeService.Create(domain.CreateRecipeRequest{
			Name:         name,
			Servings:     4,
			Ingredients:  ings,
			Instructions: []string{"Cook"},
		}, env.householdID)
		if err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		return r.ID, domain.RecipeSummary{ID: r.ID, Name: name, Servings: 4}
	}

	aID, aSum := mk("Kycklinggryta", ing("kyckling", "ris"))
	bID, bSum := mk("Kycklingsallad", ing("kyckling", "sallad"))
	cID, cSum := mk("Torskrätt", ing("torsk", "potatis"))

	// Build the ingredient map the way the service does: load full recipes by
	// ID via the real storage. This exercises loadIngredients' data source.
	ingMap := env.menuService.loadIngredients([]domain.RecipeSummary{aSum, bSum, cSum})
	if len(ingMap[aID]) == 0 || len(ingMap[bID]) == 0 || len(ingMap[cID]) == 0 {
		t.Fatalf("loadIngredients did not return ingredients for all recipes: %v", ingMap)
	}

	// Drive the selector over exactly our three recipes across several seeds.
	// The first pick is a 0-overlap tie (shuffle order), but once one kyckling
	// recipe is chosen the other jumps to overlap=1 and must be picked before
	// the unrelated torsk recipe — so the two kyckling recipes always end up
	// adjacent in the 3-pick sequence.
	for _, seed := range []uint64{1, 7, 42, 100} {
		sel, ok := newMenuSelector([]domain.RecipeSummary{aSum, bSum, cSum},
			domain.DefaultMenuPreferences(env.householdID),
			rand.New(rand.NewPCG(seed, seed+1)), ingMap)
		if !ok {
			t.Fatalf("seed %d: selector failed to build", seed)
		}
		order := []string{sel.Next(), sel.Next(), sel.Next()}
		posA, posB := indexOf(order, aID), indexOf(order, bID)
		if abs(posA-posB) != 1 {
			t.Errorf("seed %d: kyckling recipes should be adjacent via overlap, got order %v (c=%s)", seed, order, cID)
		}
	}
}

func indexOf(s []string, v string) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return -1
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func TestAverageMealNutrition(t *testing.T) {
	f := func(v float64) *float64 { return &v }
	macros := func(cal, prot, carb, fat float64) *domain.Nutrition {
		return &domain.Nutrition{Calories: f(cal), ProteinG: f(prot), CarbsG: f(carb), FatG: f(fat)}
	}

	t.Run("averages per-serving macros across meals", func(t *testing.T) {
		days := []domain.MenuResponseDay{
			{Date: "d1", RecipeID: "r1", Servings: 4, Nutrition: macros(500, 30, 50, 20)},
			{Date: "d2", RecipeID: "r2", Servings: 2, Nutrition: macros(600, 40, 60, 25)},
		}
		got := averageMealNutrition(days)
		if got == nil {
			t.Fatal("expected non-nil average")
		}
		// Per serving, not scaled by servings: (500+600)/2 = 550 kcal; (30+40)/2 = 35 g protein.
		if got.Calories != 550 || got.ProteinG != 35 {
			t.Errorf("unexpected averages: %+v", got)
		}
		if got.Partial {
			t.Error("Partial should be false when every meal has data")
		}
	})

	t.Run("counts leftover days as meals, ignores skipped", func(t *testing.T) {
		days := []domain.MenuResponseDay{
			{Date: "d1", RecipeID: "r1", Servings: 8, Nutrition: macros(400, 0, 0, 0), PrepMode: domain.PrepModeBatch},
			{Date: "d2", RecipeID: "r1", Servings: 4, Nutrition: macros(400, 0, 0, 0), LeftoverOf: "d1"},
			{Date: "d3", Skip: true, Servings: 0, Nutrition: macros(999, 0, 0, 0)},
		}
		got := averageMealNutrition(days)
		if got == nil {
			t.Fatal("expected non-nil average")
		}
		// Two meals (cook + leftover) at 400 per serving each; skipped day excluded.
		if got.Calories != 400 {
			t.Errorf("expected 400 kcal average per meal, got %v", got.Calories)
		}
		if got.Partial {
			t.Error("Partial should be false when both meals have data")
		}
	})

	t.Run("flags partial when a meal lacks data and averages the rest", func(t *testing.T) {
		days := []domain.MenuResponseDay{
			{Date: "d1", RecipeID: "r1", Servings: 4, Nutrition: macros(500, 30, 50, 20)},
			{Date: "d2", RecipeID: "r2", Servings: 4}, // no nutrition
		}
		got := averageMealNutrition(days)
		if got == nil || !got.Partial {
			t.Fatalf("expected non-nil partial average, got %+v", got)
		}
		// Averaged over the one meal with data only.
		if got.Calories != 500 {
			t.Errorf("expected 500 kcal (over the meal with data), got %v", got.Calories)
		}
	})

	t.Run("nil when no meal carries data", func(t *testing.T) {
		days := []domain.MenuResponseDay{
			{Date: "d1", RecipeID: "r1", Servings: 4},
			{Date: "d2", Skip: true},
		}
		if got := averageMealNutrition(days); got != nil {
			t.Errorf("expected nil when no nutrition data, got %+v", got)
		}
	})
}

func TestMenuGenerate_ExcludeRecipeIDs(t *testing.T) {
	// Two recipes; excluding one must keep it out of the whole generated week.
	recipes := []domain.RecipeSummary{
		{ID: "rec_keep", Name: "Behåll", Servings: 4},
		{ID: "rec_drop", Name: "Uteslut", Servings: 4},
	}
	svc := NewMenuService(&captureMenuRepo{}, &batchStubRecipeRepo{summaries: recipes}, nil)

	resp, err := svc.Generate("hh_1", domain.GenerateMenuRequest{
		Days: 6, Servings: 4, ExcludeRecipeIDs: []string{"rec_drop"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, d := range resp.Days {
		if d.RecipeID == "rec_drop" {
			t.Errorf("day %d used excluded recipe rec_drop", i)
		}
	}

	// Excluding every recipe is ignored (better a repeat than an empty week).
	resp, err = svc.Generate("hh_1", domain.GenerateMenuRequest{
		Days: 3, Servings: 4, ExcludeRecipeIDs: []string{"rec_keep", "rec_drop"},
	})
	if err != nil {
		t.Fatalf("unexpected error (exclude-all): %v", err)
	}
	for i, d := range resp.Days {
		if d.RecipeID == "" {
			t.Errorf("day %d empty when all recipes excluded; exclusion should be ignored", i)
		}
	}
}

func TestMenuGetCurrent_NoMenu(t *testing.T) {
	env := newMenuTestEnv(t)

	resp, err := env.menuService.GetCurrent("hh_nonexistent")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp != nil {
		t.Errorf("expected nil response for no menu, got %+v", resp)
	}
}

func TestMenuGetCurrent_Success(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipes(t, 3)

	generated, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     3,
		Servings: 4,
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	current, err := env.menuService.GetCurrent(env.householdID)
	if err != nil {
		t.Fatalf("GetCurrent failed: %v", err)
	}
	if current == nil {
		t.Fatal("expected non-nil menu")
	}
	if current.ID != generated.ID {
		t.Errorf("expected menu ID %q, got %q", generated.ID, current.ID)
	}
	if len(current.Days) != 3 {
		t.Errorf("expected 3 days, got %d", len(current.Days))
	}
}

func TestMenuGenerate_BoundaryDays(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipes(t, 3)

	// days=1 should work
	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     1,
		Servings: 4,
	})
	if err != nil {
		t.Fatalf("days=1 should succeed, got %v", err)
	}
	if len(resp.Days) != 1 {
		t.Errorf("expected 1 day, got %d", len(resp.Days))
	}

	// days=31 should work
	resp, err = env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     31,
		Servings: 4,
	})
	if err != nil {
		t.Fatalf("days=31 should succeed, got %v", err)
	}
	if len(resp.Days) != 31 {
		t.Errorf("expected 31 days, got %d", len(resp.Days))
	}
}
