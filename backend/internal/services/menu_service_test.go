package services

import (
	"maltiden/internal/domain"
	"maltiden/internal/storage/sqlite"
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
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	// Create a user (which creates a household) to satisfy FK constraint
	user := createTestUser(t, authService, "menu-test@test.com", "MenuTester")

	return &menuTestEnv{
		menuService:   NewMenuService(menuStorage, recipeStorage),
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
		})
		if err != nil {
			t.Fatalf("failed to seed recipe %d: %v", i, err)
		}
		ids[i] = resp.ID
	}
	return ids
}

func TestMenuGenerate_DefaultDays(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipes(t, 5)

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     0, // should default to 5
		Servings: 4,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Days) != 5 {
		t.Errorf("expected 5 days (default), got %d", len(resp.Days))
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
