package services

import (
	"database/sql"
	"errors"
	"maltiden/internal/domain"
	"maltiden/internal/storage/sqlite"
	"math"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// shoppingTestEnv wires a real SQLite test DB to a ShoppingService so we can
// catch IDOR scoping at the SQL level (not just through a mock).
type shoppingTestEnv struct {
	shoppingService *ShoppingService
	menuService     *MenuService
	recipeService   *RecipeService
	householdID     string
	menuID          string
}

func newShoppingTestEnv(t *testing.T) *shoppingTestEnv {
	t.Helper()
	db := setupTestDB(t)
	recipeStorage := sqlite.NewRecipeStorage(db)
	menuStorage := sqlite.NewMenuStorage(db)
	shoppingStorage := sqlite.NewShoppingStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	user := createTestUser(t, authService, "shop-test@test.com", "Shopper")

	recipeService := NewRecipeService(recipeStorage)
	menuService := NewMenuService(menuStorage, recipeStorage)
	shoppingService := NewShoppingService(menuStorage, recipeStorage, shoppingStorage)

	// Seed at least one recipe so menu generation succeeds.
	if _, err := recipeService.Create(domain.CreateRecipeRequest{
		Name:         "Pasta",
		Servings:     4,
		Ingredients:  []domain.Ingredient{{Name: "Tomat", Amount: 2, Unit: "st"}},
		Instructions: []string{"Cook"},
	}, user.User.HouseholdID); err != nil {
		t.Fatalf("seed recipe: %v", err)
	}

	menu, err := menuService.Generate(user.User.HouseholdID, domain.GenerateMenuRequest{
		Days:     3,
		Servings: 4,
	})
	if err != nil {
		t.Fatalf("generate menu: %v", err)
	}

	return &shoppingTestEnv{
		shoppingService: shoppingService,
		menuService:     menuService,
		recipeService:   recipeService,
		householdID:     user.User.HouseholdID,
		menuID:          menu.ID,
	}
}

// newSecondHousehold spins up a second household (for IDOR tests) sharing the
// same DB by reusing the env's services. Returns the second household ID and
// its own menu ID.
func newSecondHousehold(t *testing.T, env *shoppingTestEnv, db *sql.DB) (string, string) {
	t.Helper()
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)
	user := createTestUser(t, authService, "shop-test2@test.com", "Shopper2")

	if _, err := env.recipeService.Create(domain.CreateRecipeRequest{
		Name:         "Soup",
		Servings:     4,
		Ingredients:  []domain.Ingredient{{Name: "Lök", Amount: 1, Unit: "st"}},
		Instructions: []string{"Boil"},
	}, user.User.HouseholdID); err != nil {
		t.Fatalf("seed recipe2: %v", err)
	}
	menu, err := env.menuService.Generate(user.User.HouseholdID, domain.GenerateMenuRequest{
		Days: 2, Servings: 2,
	})
	if err != nil {
		t.Fatalf("generate menu2: %v", err)
	}
	return user.User.HouseholdID, menu.ID
}

// shoppingTestEnvWithDB exposes the underlying *sql.DB so cross-household
// tests can spin up a second user against the same database.
type shoppingTestEnvWithDB struct {
	*shoppingTestEnv
	db *sql.DB
}

func newShoppingTestEnvWithDB(t *testing.T) *shoppingTestEnvWithDB {
	t.Helper()
	db := setupTestDB(t)
	recipeStorage := sqlite.NewRecipeStorage(db)
	menuStorage := sqlite.NewMenuStorage(db)
	shoppingStorage := sqlite.NewShoppingStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	user := createTestUser(t, authService, "shop-test@test.com", "Shopper")

	recipeService := NewRecipeService(recipeStorage)
	menuService := NewMenuService(menuStorage, recipeStorage)
	shoppingService := NewShoppingService(menuStorage, recipeStorage, shoppingStorage)

	if _, err := recipeService.Create(domain.CreateRecipeRequest{
		Name:         "Pasta",
		Servings:     4,
		Ingredients:  []domain.Ingredient{{Name: "Tomat", Amount: 2, Unit: "st"}},
		Instructions: []string{"Cook"},
	}, user.User.HouseholdID); err != nil {
		t.Fatalf("seed recipe: %v", err)
	}
	menu, err := menuService.Generate(user.User.HouseholdID, domain.GenerateMenuRequest{
		Days: 3, Servings: 4,
	})
	if err != nil {
		t.Fatalf("generate menu: %v", err)
	}

	return &shoppingTestEnvWithDB{
		shoppingTestEnv: &shoppingTestEnv{
			shoppingService: shoppingService,
			menuService:     menuService,
			recipeService:   recipeService,
			householdID:     user.User.HouseholdID,
			menuID:          menu.ID,
		},
		db: db,
	}
}

// ---------- CreateCustomItem ----------

func TestCreateCustomItem_Success(t *testing.T) {
	env := newShoppingTestEnv(t)
	item, err := env.shoppingService.CreateCustomItem(env.menuID, env.householdID, domain.CreateCustomItemRequest{
		Name:   "Kaffe",
		Unit:   "g",
		Amount: 250,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(item.ID, "citem_") {
		t.Errorf("expected citem_ prefix, got %q", item.ID)
	}
	if item.Name != "Kaffe" || item.Unit != "g" || item.Amount != 250 {
		t.Errorf("unexpected fields: %+v", item)
	}
	if item.MenuID != env.menuID || item.HouseholdID != env.householdID {
		t.Errorf("scoping wrong: %+v", item)
	}
}

func TestCreateCustomItem_TrimsWhitespaceFromName(t *testing.T) {
	env := newShoppingTestEnv(t)
	item, err := env.shoppingService.CreateCustomItem(env.menuID, env.householdID, domain.CreateCustomItemRequest{
		Name: "  Mjölk  ", Unit: "l", Amount: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "Mjölk" {
		t.Errorf("expected trimmed name, got %q", item.Name)
	}
}

func TestCreateCustomItem_RejectsEmptyName(t *testing.T) {
	env := newShoppingTestEnv(t)
	for _, name := range []string{"", "   ", "\t \n"} {
		_, err := env.shoppingService.CreateCustomItem(env.menuID, env.householdID, domain.CreateCustomItemRequest{
			Name: name, Unit: "st", Amount: 1,
		})
		if !errors.Is(err, domain.ErrNameRequired) {
			t.Errorf("name=%q: expected ErrNameRequired, got %v", name, err)
		}
	}
}

func TestCreateCustomItem_RejectsLongName(t *testing.T) {
	env := newShoppingTestEnv(t)
	longName := strings.Repeat("a", 201)
	_, err := env.shoppingService.CreateCustomItem(env.menuID, env.householdID, domain.CreateCustomItemRequest{
		Name: longName, Unit: "st", Amount: 1,
	})
	if !errors.Is(err, domain.ErrNameTooLong) {
		t.Errorf("expected ErrNameTooLong, got %v", err)
	}
}

func TestCreateCustomItem_RejectsLongUnit(t *testing.T) {
	env := newShoppingTestEnv(t)
	_, err := env.shoppingService.CreateCustomItem(env.menuID, env.householdID, domain.CreateCustomItemRequest{
		Name: "Kaffe", Unit: strings.Repeat("u", 21), Amount: 1,
	})
	if !errors.Is(err, domain.ErrUnitTooLong) {
		t.Errorf("expected ErrUnitTooLong, got %v", err)
	}
}

func TestCreateCustomItem_RejectsInvalidAmount(t *testing.T) {
	env := newShoppingTestEnv(t)
	cases := []float64{math.NaN(), math.Inf(1), math.Inf(-1), -1, -0.0001}
	for _, amt := range cases {
		_, err := env.shoppingService.CreateCustomItem(env.menuID, env.householdID, domain.CreateCustomItemRequest{
			Name: "Kaffe", Unit: "g", Amount: amt,
		})
		if !errors.Is(err, domain.ErrInvalidAmount) {
			t.Errorf("amount=%v: expected ErrInvalidAmount, got %v", amt, err)
		}
	}
}

func TestCreateCustomItem_RejectsAmountTooLarge(t *testing.T) {
	env := newShoppingTestEnv(t)
	_, err := env.shoppingService.CreateCustomItem(env.menuID, env.householdID, domain.CreateCustomItemRequest{
		Name: "Kaffe", Unit: "g", Amount: 100001,
	})
	if !errors.Is(err, domain.ErrAmountTooLarge) {
		t.Errorf("expected ErrAmountTooLarge, got %v", err)
	}
}

func TestCreateCustomItem_DefaultsAmountAndUnit(t *testing.T) {
	env := newShoppingTestEnv(t)
	item, err := env.shoppingService.CreateCustomItem(env.menuID, env.householdID, domain.CreateCustomItemRequest{
		Name: "Bröd", Unit: "", Amount: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Unit != "st" {
		t.Errorf("expected default unit 'st', got %q", item.Unit)
	}
	if item.Amount != 1 {
		t.Errorf("expected default amount 1, got %v", item.Amount)
	}
}

func TestCreateCustomItem_RespectsRowCap(t *testing.T) {
	env := newShoppingTestEnv(t)
	for i := 0; i < 500; i++ {
		if _, err := env.shoppingService.CreateCustomItem(env.menuID, env.householdID, domain.CreateCustomItemRequest{
			Name: "Item", Unit: "st", Amount: 1,
		}); err != nil {
			t.Fatalf("create #%d failed: %v", i+1, err)
		}
	}
	_, err := env.shoppingService.CreateCustomItem(env.menuID, env.householdID, domain.CreateCustomItemRequest{
		Name: "Item", Unit: "st", Amount: 1,
	})
	if !errors.Is(err, domain.ErrTooManyItems) {
		t.Errorf("expected ErrTooManyItems on 501st create, got %v", err)
	}
}

func TestCreateCustomItem_VerifiesMenuOwnership(t *testing.T) {
	env := newShoppingTestEnvWithDB(t)
	otherHouseholdID, _ := newSecondHousehold(t, env.shoppingTestEnv, env.db)

	// Wrong household trying to add to first menu -> Forbidden.
	_, err := env.shoppingService.CreateCustomItem(env.menuID, otherHouseholdID, domain.CreateCustomItemRequest{
		Name: "X", Unit: "st", Amount: 1,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden for cross-household, got %v", err)
	}

	// Non-existent menu -> ErrMenuNotFound.
	_, err = env.shoppingService.CreateCustomItem("menu_does_not_exist", env.householdID, domain.CreateCustomItemRequest{
		Name: "X", Unit: "st", Amount: 1,
	})
	if !errors.Is(err, domain.ErrMenuNotFound) {
		t.Errorf("expected ErrMenuNotFound for missing menu, got %v", err)
	}
}

// ---------- UpdateItemChecked: the IDOR regression test ----------

func TestUpdateItemChecked_RejectsCrossHouseholdCustomItem(t *testing.T) {
	env := newShoppingTestEnvWithDB(t)
	otherHouseholdID, otherMenuID := newSecondHousehold(t, env.shoppingTestEnv, env.db)

	// Household A creates a custom item on its own menu.
	itemA, err := env.shoppingService.CreateCustomItem(env.menuID, env.householdID, domain.CreateCustomItemRequest{
		Name: "Hemlig", Unit: "st", Amount: 1,
	})
	if err != nil {
		t.Fatalf("create A: %v", err)
	}

	// Household B tries to toggle citem belonging to A, but routes through B's
	// own menu (passing the IDOR menu check). Must not affect A's row.
	err = env.shoppingService.UpdateItemChecked(otherMenuID, itemA.ID, otherHouseholdID, true)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows for cross-household toggle, got %v", err)
	}
}

func TestUpdateItemChecked_AcceptsSameHouseholdCustomItem(t *testing.T) {
	env := newShoppingTestEnv(t)
	item, err := env.shoppingService.CreateCustomItem(env.menuID, env.householdID, domain.CreateCustomItemRequest{
		Name: "Kex", Unit: "st", Amount: 2,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := env.shoppingService.UpdateItemChecked(env.menuID, item.ID, env.householdID, true); err != nil {
		t.Errorf("expected success toggling own item, got %v", err)
	}
}

// ---------- GetShoppingList: category ordering ----------

// TestGetShoppingList_OrdersCategoriesCorrectly seeds a recipe whose ingredients
// span several of the new categories (#168) and confirms the resulting shopping
// list returns them in the supermarket-walk order, not insertion order.
func TestGetShoppingList_OrdersCategoriesCorrectly(t *testing.T) {
	db := setupTestDB(t)
	recipeStorage := sqlite.NewRecipeStorage(db)
	menuStorage := sqlite.NewMenuStorage(db)
	shoppingStorage := sqlite.NewShoppingStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	user := createTestUser(t, authService, "order-test@test.com", "Orderer")
	recipeService := NewRecipeService(recipeStorage)
	shoppingService := NewShoppingService(menuStorage, recipeStorage, shoppingStorage)

	// One recipe touching seven of the new categories at once.
	recipe, err := recipeService.Create(domain.CreateRecipeRequest{
		Name:     "Allt i ett",
		Servings: 2,
		Ingredients: []domain.Ingredient{
			{Name: "Olivolja", Amount: 2, Unit: "msk"},          // Såser & olja
			{Name: "Krossade tomater", Amount: 1, Unit: "burk"}, // Konserver
			{Name: "Spaghetti", Amount: 200, Unit: "g"},         // Pasta, ris & spannmål
			{Name: "Persilja", Amount: 1, Unit: "kruka"},        // Färska örter
			{Name: "Vitlök", Amount: 2, Unit: "klyftor"},        // Grönsaker
			{Name: "Parmesan", Amount: 50, Unit: "g"},           // Mejeri & Ägg
			{Name: "Köttfärs", Amount: 400, Unit: "g"},          // Kött & Fisk
			{Name: "Salt", Amount: 1, Unit: "tsk"},              // Kryddor
		},
		Instructions: []string{"Koka"},
	}, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("seed recipe: %v", err)
	}

	// Construct the menu directly (rather than going through MenuService.Generate,
	// which would shuffle in seed recipes from migrations) so the shopping list
	// is built from this single hand-picked recipe.
	menu := &domain.Menu{
		ID:          "menu_" + uuid.New().String(),
		HouseholdID: user.User.HouseholdID,
		Days: []domain.MenuDay{
			{Date: "2026-01-01", RecipeID: recipe.ID, Servings: 2},
		},
	}
	if err := menuStorage.Create(menu); err != nil {
		t.Fatalf("create menu: %v", err)
	}

	list, err := shoppingService.GetShoppingList(menu.ID, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("get shopping list: %v", err)
	}
	if list == nil {
		t.Fatal("expected list, got nil")
	}

	got := make([]string, 0, len(list.Categories))
	for _, c := range list.Categories {
		got = append(got, c.Name)
	}

	// Categories with no items are omitted; the remaining ones must appear in
	// the canonical supermarket-walk order from #168.
	want := []string{
		"Kött & Fisk",
		"Mejeri & Ägg",
		"Grönsaker",
		"Färska örter",
		"Pasta, ris & spannmål",
		"Konserver",
		"Såser & olja",
		"Kryddor",
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d categories %v, got %d %v", len(want), want, len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("category[%d] = %q, want %q (full order: %v)", i, got[i], want[i], got)
		}
	}
}

// ---------- DeleteCustomItem ----------

func TestDeleteCustomItem_RejectsCrossHouseholdCustomItem(t *testing.T) {
	env := newShoppingTestEnvWithDB(t)
	otherHouseholdID, _ := newSecondHousehold(t, env.shoppingTestEnv, env.db)

	itemA, err := env.shoppingService.CreateCustomItem(env.menuID, env.householdID, domain.CreateCustomItemRequest{
		Name: "Hemlig", Unit: "st", Amount: 1,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	err = env.shoppingService.DeleteCustomItem(itemA.ID, otherHouseholdID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows for cross-household delete, got %v", err)
	}
}

func TestDeleteCustomItem_Success(t *testing.T) {
	env := newShoppingTestEnv(t)
	item, err := env.shoppingService.CreateCustomItem(env.menuID, env.householdID, domain.CreateCustomItemRequest{
		Name: "Kex", Unit: "st", Amount: 2,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := env.shoppingService.DeleteCustomItem(item.ID, env.householdID); err != nil {
		t.Errorf("expected delete success, got %v", err)
	}
	// Idempotency: second delete should fail with ErrNoRows.
	if err := env.shoppingService.DeleteCustomItem(item.ID, env.householdID); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows on second delete, got %v", err)
	}
}

// ---------- GetShoppingList: cross-day aggregation ----------

// findItem searches all categories in the shopping list for an item whose name
// matches (case-insensitive) and unit matches exactly. Returns nil if not found.
func findItem(list *domain.ShoppingList, name, unit string) *domain.ShoppingItem {
	nameLower := strings.ToLower(name)
	for i := range list.Categories {
		for j := range list.Categories[i].Items {
			it := &list.Categories[i].Items[j]
			if strings.ToLower(it.Name) == nameLower && it.Unit == unit {
				return it
			}
		}
	}
	return nil
}

// TestGetShoppingList_AggregatesAcrossMenuDays_ServingsScale verifies that
// per-day servings scaling is applied before summing: recipe (4 servings,
// 200 g Tomat) used on two days at 2 servings each → 100+100 = 200 g total.
func TestGetShoppingList_AggregatesAcrossMenuDays_ServingsScale(t *testing.T) {
	db := setupTestDB(t)
	recipeStorage := sqlite.NewRecipeStorage(db)
	menuStorage := sqlite.NewMenuStorage(db)
	shoppingStorage := sqlite.NewShoppingStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	user := createTestUser(t, authService, "agg-scale@test.com", "Scaler")
	recipeService := NewRecipeService(recipeStorage)
	shoppingService := NewShoppingService(menuStorage, recipeStorage, shoppingStorage)

	recipe, err := recipeService.Create(domain.CreateRecipeRequest{
		Name:         "Tomatgröt",
		Servings:     4,
		Ingredients:  []domain.Ingredient{{Name: "Tomat", Amount: 200, Unit: "g"}},
		Instructions: []string{"Koka"},
	}, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("seed recipe: %v", err)
	}

	menu := &domain.Menu{
		ID:          "menu_" + uuid.New().String(),
		HouseholdID: user.User.HouseholdID,
		Days: []domain.MenuDay{
			{Date: "2026-01-01", RecipeID: recipe.ID, Servings: 2},
			{Date: "2026-01-02", RecipeID: recipe.ID, Servings: 2},
		},
	}
	if err := menuStorage.Create(menu); err != nil {
		t.Fatalf("create menu: %v", err)
	}

	list, err := shoppingService.GetShoppingList(menu.ID, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("get shopping list: %v", err)
	}

	it := findItem(list, "Tomat", "g")
	if it == nil {
		t.Fatal("expected Tomat/g in shopping list, not found")
	}
	if it.Amount != 200 {
		t.Errorf("Tomat amount = %v, want 200 (100+100 after 0.5× scale each day)", it.Amount)
	}
}

// TestGetShoppingList_AggregatesAcrossMenuDays_DifferentUnitsStaySeparate
// verifies that same ingredient name with different units produces two distinct
// items (the service does NOT convert g↔dl etc.).
func TestGetShoppingList_AggregatesAcrossMenuDays_DifferentUnitsStaySeparate(t *testing.T) {
	db := setupTestDB(t)
	recipeStorage := sqlite.NewRecipeStorage(db)
	menuStorage := sqlite.NewMenuStorage(db)
	shoppingStorage := sqlite.NewShoppingStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	user := createTestUser(t, authService, "agg-units@test.com", "Units")
	recipeService := NewRecipeService(recipeStorage)
	shoppingService := NewShoppingService(menuStorage, recipeStorage, shoppingStorage)

	recipeA, err := recipeService.Create(domain.CreateRecipeRequest{
		Name:         "Mjölksås",
		Servings:     4,
		Ingredients:  []domain.Ingredient{{Name: "Mjölk", Amount: 2, Unit: "dl"}},
		Instructions: []string{"Värm"},
	}, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("seed recipeA: %v", err)
	}
	recipeB, err := recipeService.Create(domain.CreateRecipeRequest{
		Name:         "Mjölkgröt",
		Servings:     4,
		Ingredients:  []domain.Ingredient{{Name: "Mjölk", Amount: 1, Unit: "l"}},
		Instructions: []string{"Koka"},
	}, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("seed recipeB: %v", err)
	}

	menu := &domain.Menu{
		ID:          "menu_" + uuid.New().String(),
		HouseholdID: user.User.HouseholdID,
		Days: []domain.MenuDay{
			{Date: "2026-01-01", RecipeID: recipeA.ID, Servings: 4},
			{Date: "2026-01-02", RecipeID: recipeB.ID, Servings: 4},
		},
	}
	if err := menuStorage.Create(menu); err != nil {
		t.Fatalf("create menu: %v", err)
	}

	list, err := shoppingService.GetShoppingList(menu.ID, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("get shopping list: %v", err)
	}

	itDl := findItem(list, "Mjölk", "dl")
	itL := findItem(list, "Mjölk", "l")
	if itDl == nil {
		t.Error("expected Mjölk/dl item, not found")
	}
	if itL == nil {
		t.Error("expected Mjölk/l item, not found")
	}
	if itDl != nil && itDl.Amount != 2 {
		t.Errorf("Mjölk/dl amount = %v, want 2", itDl.Amount)
	}
	if itL != nil && itL.Amount != 1 {
		t.Errorf("Mjölk/l amount = %v, want 1", itL.Amount)
	}
}

// TestGetShoppingList_AggregatesAcrossMenuDays_SkipExcludesDay verifies that a
// day with Skip==true contributes nothing to the shopping list.
func TestGetShoppingList_AggregatesAcrossMenuDays_SkipExcludesDay(t *testing.T) {
	db := setupTestDB(t)
	recipeStorage := sqlite.NewRecipeStorage(db)
	menuStorage := sqlite.NewMenuStorage(db)
	shoppingStorage := sqlite.NewShoppingStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	user := createTestUser(t, authService, "agg-skip@test.com", "Skipper")
	recipeService := NewRecipeService(recipeStorage)
	shoppingService := NewShoppingService(menuStorage, recipeStorage, shoppingStorage)

	recipe, err := recipeService.Create(domain.CreateRecipeRequest{
		Name:         "Morotssoppa",
		Servings:     4,
		Ingredients:  []domain.Ingredient{{Name: "Morot", Amount: 100, Unit: "g"}},
		Instructions: []string{"Koka"},
	}, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("seed recipe: %v", err)
	}

	menu := &domain.Menu{
		ID:          "menu_" + uuid.New().String(),
		HouseholdID: user.User.HouseholdID,
		Days: []domain.MenuDay{
			{Date: "2026-01-01", RecipeID: recipe.ID, Servings: 4, Skip: false},
			{Date: "2026-01-02", RecipeID: recipe.ID, Servings: 4, Skip: true},
		},
	}
	if err := menuStorage.Create(menu); err != nil {
		t.Fatalf("create menu: %v", err)
	}

	list, err := shoppingService.GetShoppingList(menu.ID, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("get shopping list: %v", err)
	}

	it := findItem(list, "Morot", "g")
	if it == nil {
		t.Fatal("expected Morot/g in shopping list, not found")
	}
	// Only one day contributes (the other is skipped), so amount should be 100.
	if it.Amount != 100 {
		t.Errorf("Morot amount = %v, want 100 (skipped day must be excluded)", it.Amount)
	}
}

// TestGetShoppingList_AggregatesAcrossMenuDays_CaseInsensitiveNameMatch
// verifies that "tomat" and "Tomat" (same unit) collapse into a single item.
func TestGetShoppingList_AggregatesAcrossMenuDays_CaseInsensitiveNameMatch(t *testing.T) {
	db := setupTestDB(t)
	recipeStorage := sqlite.NewRecipeStorage(db)
	menuStorage := sqlite.NewMenuStorage(db)
	shoppingStorage := sqlite.NewShoppingStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	user := createTestUser(t, authService, "agg-case@test.com", "Caser")
	recipeService := NewRecipeService(recipeStorage)
	shoppingService := NewShoppingService(menuStorage, recipeStorage, shoppingStorage)

	recipeA, err := recipeService.Create(domain.CreateRecipeRequest{
		Name:         "Recept A",
		Servings:     4,
		Ingredients:  []domain.Ingredient{{Name: "tomat", Amount: 100, Unit: "g"}},
		Instructions: []string{"Koka"},
	}, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("seed recipeA: %v", err)
	}
	recipeB, err := recipeService.Create(domain.CreateRecipeRequest{
		Name:         "Recept B",
		Servings:     4,
		Ingredients:  []domain.Ingredient{{Name: "Tomat", Amount: 100, Unit: "g"}},
		Instructions: []string{"Koka"},
	}, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("seed recipeB: %v", err)
	}

	menu := &domain.Menu{
		ID:          "menu_" + uuid.New().String(),
		HouseholdID: user.User.HouseholdID,
		Days: []domain.MenuDay{
			{Date: "2026-01-01", RecipeID: recipeA.ID, Servings: 4},
			{Date: "2026-01-02", RecipeID: recipeB.ID, Servings: 4},
		},
	}
	if err := menuStorage.Create(menu); err != nil {
		t.Fatalf("create menu: %v", err)
	}

	list, err := shoppingService.GetShoppingList(menu.ID, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("get shopping list: %v", err)
	}

	// Count how many items with name "tomat" (case-insensitive) and unit "g" appear.
	count := 0
	totalAmount := 0.0
	for _, cat := range list.Categories {
		for _, it := range cat.Items {
			if strings.ToLower(it.Name) == "tomat" && it.Unit == "g" {
				count++
				totalAmount += it.Amount
			}
		}
	}
	if count != 1 {
		t.Errorf("expected 1 tomat/g item (case-insensitive merge), got %d", count)
	}
	if totalAmount != 200 {
		t.Errorf("tomat/g total amount = %v, want 200 (100+100 merged)", totalAmount)
	}
}

// TestGetShoppingList_AggregatesAcrossMenuDays_RoundAmountSt verifies that
// unit "st" rounds the aggregated total to the nearest integer.
func TestGetShoppingList_AggregatesAcrossMenuDays_RoundAmountSt(t *testing.T) {
	db := setupTestDB(t)
	recipeStorage := sqlite.NewRecipeStorage(db)
	menuStorage := sqlite.NewMenuStorage(db)
	shoppingStorage := sqlite.NewShoppingStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	user := createTestUser(t, authService, "agg-round-st@test.com", "RoundSt")
	recipeService := NewRecipeService(recipeStorage)
	shoppingService := NewShoppingService(menuStorage, recipeStorage, shoppingStorage)

	// 3-serving recipe with 2 st; used twice at 2 servings each.
	// scale = 2/3 ≈ 0.6667 per day → 2×(2×0.6667) ≈ 2.6667 st total.
	// roundAmount("st") must round to nearest integer → 3.
	recipe, err := recipeService.Create(domain.CreateRecipeRequest{
		Name:         "Äggröra",
		Servings:     3,
		Ingredients:  []domain.Ingredient{{Name: "Ägg", Amount: 2, Unit: "st"}},
		Instructions: []string{"Vispa"},
	}, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("seed recipe: %v", err)
	}

	menu := &domain.Menu{
		ID:          "menu_" + uuid.New().String(),
		HouseholdID: user.User.HouseholdID,
		Days: []domain.MenuDay{
			{Date: "2026-01-01", RecipeID: recipe.ID, Servings: 2},
			{Date: "2026-01-02", RecipeID: recipe.ID, Servings: 2},
		},
	}
	if err := menuStorage.Create(menu); err != nil {
		t.Fatalf("create menu: %v", err)
	}

	list, err := shoppingService.GetShoppingList(menu.ID, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("get shopping list: %v", err)
	}

	it := findItem(list, "Ägg", "st")
	if it == nil {
		t.Fatal("expected Ägg/st in shopping list, not found")
	}
	// Must be an integer (no fractional part).
	if it.Amount != math.Round(it.Amount) {
		t.Errorf("Ägg/st amount = %v, expected integer rounding for unit 'st'", it.Amount)
	}
	if it.Amount != 3 {
		t.Errorf("Ägg/st amount = %v, want 3 (math.Round(2.6667))", it.Amount)
	}
}

// TestGetShoppingList_AggregatesAcrossMenuDays_RoundAmountTwoDecimals verifies
// that non-st units round the aggregated total to 2 decimal places.
func TestGetShoppingList_AggregatesAcrossMenuDays_RoundAmountTwoDecimals(t *testing.T) {
	db := setupTestDB(t)
	recipeStorage := sqlite.NewRecipeStorage(db)
	menuStorage := sqlite.NewMenuStorage(db)
	shoppingStorage := sqlite.NewShoppingStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	user := createTestUser(t, authService, "agg-round-g@test.com", "RoundG")
	recipeService := NewRecipeService(recipeStorage)
	shoppingService := NewShoppingService(menuStorage, recipeStorage, shoppingStorage)

	// Same 3/2-servings setup as the st test but unit "g":
	// 2×(2×2/3) ≈ 2.6667 g → roundAmount("g") = math.Round(2.6667*100)/100 = 2.67.
	recipe, err := recipeService.Create(domain.CreateRecipeRequest{
		Name:         "Kryddsås",
		Servings:     3,
		Ingredients:  []domain.Ingredient{{Name: "Ingefära", Amount: 2, Unit: "g"}},
		Instructions: []string{"Blanda"},
	}, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("seed recipe: %v", err)
	}

	menu := &domain.Menu{
		ID:          "menu_" + uuid.New().String(),
		HouseholdID: user.User.HouseholdID,
		Days: []domain.MenuDay{
			{Date: "2026-01-01", RecipeID: recipe.ID, Servings: 2},
			{Date: "2026-01-02", RecipeID: recipe.ID, Servings: 2},
		},
	}
	if err := menuStorage.Create(menu); err != nil {
		t.Fatalf("create menu: %v", err)
	}

	list, err := shoppingService.GetShoppingList(menu.ID, user.User.HouseholdID)
	if err != nil {
		t.Fatalf("get shopping list: %v", err)
	}

	it := findItem(list, "Ingefära", "g")
	if it == nil {
		t.Fatal("expected Ingefära/g in shopping list, not found")
	}
	// Must be rounded to at most 2 decimal places.
	rounded := math.Round(it.Amount*100) / 100
	if it.Amount != rounded {
		t.Errorf("Ingefära/g amount = %v, expected at most 2 decimal places (rounded = %v)", it.Amount, rounded)
	}
	if it.Amount != 2.67 {
		t.Errorf("Ingefära/g amount = %v, want 2.67 (math.Round(2.6667*100)/100)", it.Amount)
	}
}
