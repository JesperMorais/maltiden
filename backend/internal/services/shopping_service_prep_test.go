package services

import (
	"maltiden/internal/domain"
	"maltiden/internal/storage/sqlite"
	"testing"

	"github.com/google/uuid"
)

// totalAmountFor sums the amount of a named ingredient across all categories.
func totalAmountFor(list *domain.ShoppingList, name string) float64 {
	total := 0.0
	for _, c := range list.Categories {
		for _, it := range c.Items {
			if it.Name == name {
				total += it.Amount
			}
		}
	}
	return total
}

// TestGetShoppingList_LeftoverDaysExcluded verifies a meal-prep leftover day adds
// nothing to the shopping list: the cook day's 2×-servings batch already covers
// it. The amounts must equal a lone cook day, NOT cook + leftover.
func TestGetShoppingList_LeftoverDaysExcluded(t *testing.T) {
	db := setupTestDB(t)
	recipeStorage := sqlite.NewRecipeStorage(db)
	menuStorage := sqlite.NewMenuStorage(db)
	shoppingStorage := sqlite.NewShoppingStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	user := createTestUser(t, authService, "shop-prep@test.com", "ShopPrep")
	recipeService := NewRecipeService(recipeStorage, nil)
	shoppingService := NewShoppingService(menuStorage, recipeStorage, shoppingStorage)

	// Recipe base = 4 servings, 500 g köttfärs.
	recipe, err := recipeService.Create(domain.CreateRecipeRequest{
		Name:         "Köttfärssås",
		Servings:     4,
		Batchable:    true,
		Ingredients:  []domain.Ingredient{{Name: "Köttfärs", CanonicalName: "köttfärs", Amount: 500, Unit: "g"}},
		Instructions: []string{"Koka"},
	}, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("seed recipe: %v", err)
	}

	// A prep week: cook day carries 2× servings (8), leftover day reuses it.
	prepMenu := &domain.Menu{
		ID:          "menu_" + uuid.New().String(),
		HouseholdID: user.User.HouseholdID,
		Days: []domain.MenuDay{
			{Date: "2026-01-01", RecipeID: recipe.ID, Servings: 8},                                          // cook (2×4)
			{Date: "2026-01-02", RecipeID: recipe.ID, Servings: 4, Leftover: true, CookDate: "2026-01-01"}, // leftovers
		},
	}
	if err := menuStorage.Create(prepMenu); err != nil {
		t.Fatalf("create prep menu: %v", err)
	}

	prepList, err := shoppingService.GetShoppingList(prepMenu.ID, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("get prep shopping list: %v", err)
	}

	// Reference: a single cook day at 2× servings, no leftover.
	refMenu := &domain.Menu{
		ID:          "menu_" + uuid.New().String(),
		HouseholdID: user.User.HouseholdID,
		Days: []domain.MenuDay{
			{Date: "2026-02-01", RecipeID: recipe.ID, Servings: 8},
		},
	}
	if err := menuStorage.Create(refMenu); err != nil {
		t.Fatalf("create ref menu: %v", err)
	}
	refList, err := shoppingService.GetShoppingList(refMenu.ID, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("get ref shopping list: %v", err)
	}

	gotPrep := totalAmountFor(prepList, "Köttfärs")
	gotRef := totalAmountFor(refList, "Köttfärs")

	// Cook day scales 2×: 500 g × (8/4) = 1000 g.
	if gotRef != 1000 {
		t.Fatalf("reference cook-day amount = %v, want 1000 (2× scale)", gotRef)
	}
	// The leftover day must NOT add another batch.
	if gotPrep != gotRef {
		t.Errorf("prep list (cook + leftover) amount = %v, want %v (leftover excluded)", gotPrep, gotRef)
	}
}
