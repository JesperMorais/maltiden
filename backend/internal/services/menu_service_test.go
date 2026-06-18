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
	authService   *AuthService
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
		recipeService: NewRecipeService(recipeStorage, nil),
		authService:   authService,
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

// seedRecipesWithIngredients creates recipes carrying real (non-staple) and
// staple ingredients so the economy summary has something to compute. It
// returns the created recipe ids.
func (env *menuTestEnv) seedRecipesWithIngredients(t *testing.T) []string {
	t.Helper()
	specs := []struct {
		name       string
		mainProt   string
		ingredients []domain.Ingredient
	}{
		{"Köttfärssås", "nöt", []domain.Ingredient{
			{Name: "Köttfärs", CanonicalName: "köttfärs", Amount: 500, Unit: "g"},
			{Name: "Gul lök", CanonicalName: "lök", Amount: 1, Unit: "st"},
			{Name: "Salt", CanonicalName: "salt", Amount: 1, Unit: "tsk", IsPantryStaple: true},
		}},
		{"Korv Stroganoff", "fläsk", []domain.Ingredient{
			{Name: "Falukorv", CanonicalName: "korv", Amount: 400, Unit: "g"},
			{Name: "Lök", CanonicalName: "lök", Amount: 1, Unit: "st"},
			{Name: "Salt", CanonicalName: "salt", Amount: 1, Unit: "tsk", IsPantryStaple: true},
		}},
		{"Laxpasta", "fisk", []domain.Ingredient{
			{Name: "Lax", CanonicalName: "lax", Amount: 300, Unit: "g"},
			{Name: "Pasta", CanonicalName: "pasta", Amount: 400, Unit: "g"},
		}},
	}
	ids := make([]string, len(specs))
	for i, s := range specs {
		resp, err := env.recipeService.Create(domain.CreateRecipeRequest{
			Name:         s.name,
			Servings:     4,
			MainProtein:  s.mainProt,
			Ingredients:  s.ingredients,
			Instructions: []string{"Do something"},
		}, env.householdID)
		if err != nil {
			t.Fatalf("seed recipe %q: %v", s.name, err)
		}
		ids[i] = resp.ID
	}
	return ids
}

func TestMenuGenerate_SeededDeterministic(t *testing.T) {
	// A generate persists a menu, which then feeds the recency penalty — so two
	// runs in the SAME household see different recency. To prove "same seed →
	// same week", we use two fresh households that share the seed recipes (NULL
	// household_id) and each start with empty menu history. With identical
	// candidates, identical (empty) recency, and the same pinned seed, the weeks
	// must match exactly.
	env := newMenuTestEnv(t)
	env.menuService.WithRNG(func() *rand.Rand { return rand.New(rand.NewPCG(99, 99)) })

	second := createTestUser(t, env.authService, "menu-det-2@test.com", "MenuDet2")

	collect := func(hh string) []string {
		resp, err := env.menuService.Generate(hh, domain.GenerateMenuRequest{Days: 5, Servings: 4})
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		out := make([]string, len(resp.Days))
		for i, d := range resp.Days {
			out[i] = d.RecipeID
		}
		return out
	}

	a := collect(env.householdID)
	b := collect(second.User.HouseholdID)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("same seed + empty recency should yield identical week: %v vs %v", a, b)
		}
	}
}

func TestMenuGenerate_LockedDaysRoundTrip(t *testing.T) {
	env := newMenuTestEnv(t)
	ids := env.seedRecipesWithIngredients(t)

	today := time.Now()
	lockDate := today.AddDate(0, 0, 1).Format("2006-01-02")
	lockedRecipe := ids[2] // Laxpasta

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:       4,
		Servings:   4,
		LockedDays: map[string]string{lockDate: lockedRecipe},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if resp.Days[1].Date != lockDate {
		t.Fatalf("expected slot 1 to be the lock date")
	}
	if resp.Days[1].RecipeID != lockedRecipe {
		t.Errorf("locked day not honored: got %q want %q", resp.Days[1].RecipeID, lockedRecipe)
	}
}

func TestMenuGenerate_UnknownLockIgnored(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipesWithIngredients(t)

	today := time.Now()
	lockDate := today.Format("2006-01-02")

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:       3,
		Servings:   4,
		LockedDays: map[string]string{lockDate: "rec_does_not_exist"},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Unknown lock → slot treated as open and filled with a real recipe.
	if resp.Days[0].RecipeID == "" || resp.Days[0].RecipeID == "rec_does_not_exist" {
		t.Errorf("unknown lock should be filled with a real recipe, got %q", resp.Days[0].RecipeID)
	}
}

func TestMenuGenerate_ForeignLockNotLeaked(t *testing.T) {
	// SECURITY: a client must not be able to pin (and thus surface) a recipe that
	// belongs to another household by passing its id in lockedDays. The foreign
	// id must be dropped silently — never fetched, never placed.
	env := newMenuTestEnv(t)
	env.seedRecipesWithIngredients(t)

	// A second household with a PRIVATE recipe not visible to the first.
	other := createTestUser(t, env.authService, "menu-foreign@test.com", "ForeignHH")
	foreign, err := env.recipeService.Create(domain.CreateRecipeRequest{
		Name:         "Secret Recipe",
		Servings:     4,
		Ingredients:  []domain.Ingredient{{Name: "Truffel", CanonicalName: "truffel", Amount: 1, Unit: "st"}},
		Instructions: []string{"Secret"},
	}, other.User.HouseholdID)
	if err != nil {
		t.Fatalf("create foreign recipe: %v", err)
	}

	today := time.Now()
	lockDate := today.Format("2006-01-02")

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:       3,
		Servings:   4,
		LockedDays: map[string]string{lockDate: foreign.ID},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// The foreign recipe must appear nowhere in the generated week.
	for i, d := range resp.Days {
		if d.RecipeID == foreign.ID {
			t.Errorf("day %d leaked foreign recipe %q", i, foreign.ID)
		}
	}
	// The locked day is treated as open and filled with one of the first
	// household's own recipes.
	if resp.Days[0].RecipeID == "" {
		t.Errorf("foreign-locked day should be filled with an own recipe, got empty")
	}
}

func TestMenuGenerate_PopulatesEconomy(t *testing.T) {
	env := newMenuTestEnv(t)
	ids := env.seedRecipesWithIngredients(t)

	// Lock the two lök-sharing recipes onto the week so the economy is
	// deterministic regardless of which restart wins.
	today := time.Now()
	d0 := today.Format("2006-01-02")
	d1 := today.AddDate(0, 0, 1).Format("2006-01-02")

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     2,
		Servings: 4,
		LockedDays: map[string]string{
			d0: ids[0], // Köttfärssås (lök + köttfärs + salt-staple)
			d1: ids[1], // Korv Stroganoff (lök + korv + salt-staple)
		},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if resp.Economy == nil {
		t.Fatal("expected economy to be populated")
	}
	// Staples (salt) must never appear among shared ingredients or counts.
	for _, s := range resp.Economy.SharedIngredients {
		if s.CanonicalName == "salt" {
			t.Errorf("staple 'salt' leaked into economy shared ingredients")
		}
	}
	// lök is shared across both locked recipes.
	foundLok := false
	for _, s := range resp.Economy.SharedIngredients {
		if s.CanonicalName == "lök" && s.RecipeCount == 2 {
			foundLok = true
		}
	}
	if !foundLok {
		t.Errorf("expected 'lök' shared across 2 recipes, got %+v", resp.Economy.SharedIngredients)
	}
	// Distinct non-staple items: köttfärs, lök, korv = 3 (salt excluded).
	if resp.Economy.DistinctItemsToBuy != 3 {
		t.Errorf("expected 3 distinct items to buy, got %d", resp.Economy.DistinctItemsToBuy)
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
