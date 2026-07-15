package services

import (
	"maltiden/internal/domain"
	"testing"
)

type fakeMenuRepo struct {
	menu *domain.Menu
}

func (f *fakeMenuRepo) Create(menu *domain.Menu) error { return nil }
func (f *fakeMenuRepo) Update(menu *domain.Menu) error { return nil }
func (f *fakeMenuRepo) GetCurrentByHousehold(householdID string) (*domain.Menu, error) {
	return f.menu, nil
}
func (f *fakeMenuRepo) GetByID(id string) (*domain.Menu, error) { return f.menu, nil }
func (f *fakeMenuRepo) GetHouseholdIDByMenuID(menuID string) (string, error) {
	return "", nil
}

type fakeRecipeRepo struct {
	recipes map[string]*domain.Recipe
}

func (f *fakeRecipeRepo) GetAll(filter *domain.RecipeFilter, householdID string) ([]domain.RecipeSummary, error) {
	return nil, nil
}
func (f *fakeRecipeRepo) GetAllPaginated(filter *domain.RecipeFilter, householdID string, limit, offset int) ([]domain.RecipeSummary, int, error) {
	return nil, 0, nil
}
func (f *fakeRecipeRepo) GetByID(id string) (*domain.Recipe, error) { return f.recipes[id], nil }
func (f *fakeRecipeRepo) GetByIDs(ids []string) (map[string]*domain.Recipe, error) {
	out := make(map[string]*domain.Recipe)
	for _, id := range ids {
		if r, ok := f.recipes[id]; ok {
			out[id] = r
		}
	}
	return out, nil
}
func (f *fakeRecipeRepo) Create(recipe *domain.Recipe) error { return nil }
func (f *fakeRecipeRepo) Update(recipe *domain.Recipe) error { return nil }
func (f *fakeRecipeRepo) Delete(id string) error             { return nil }

type fakeShoppingRepo struct {
	checked map[string]bool
}

func (f *fakeShoppingRepo) GetCheckedItems(menuID string) (map[string]bool, error) {
	if f.checked == nil {
		return map[string]bool{}, nil
	}
	return f.checked, nil
}
func (f *fakeShoppingRepo) SetChecked(menuID, itemID string, checked bool) error { return nil }

func findItem(t *testing.T, list *domain.ShoppingList, name string) *domain.ShoppingItem {
	t.Helper()
	for _, cat := range list.Categories {
		for _, it := range cat.Items {
			if it.Name == name {
				return &it
			}
		}
	}
	return nil
}

func TestGetShoppingList_AggregatesAcrossDays(t *testing.T) {
	recipe := &domain.Recipe{
		ID:       "rec_1",
		Name:     "Test",
		Servings: 2,
		Ingredients: []domain.Ingredient{
			{Name: "Lök", Amount: 1, Unit: "st"},
		},
	}
	menu := &domain.Menu{
		ID: "menu_1",
		Days: []domain.MenuDay{
			{Date: "2026-01-01", RecipeID: "rec_1", Servings: 2},
			{Date: "2026-01-02", RecipeID: "rec_1", Servings: 2},
		},
	}
	svc := NewShoppingService(
		&fakeMenuRepo{menu: menu},
		&fakeRecipeRepo{recipes: map[string]*domain.Recipe{"rec_1": recipe}},
		&fakeShoppingRepo{},
	)

	list, err := svc.GetShoppingList("menu_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := findItem(t, list, "Lök")
	if item == nil {
		t.Fatalf("expected Lök item")
	}
	if item.Amount != 2 {
		t.Errorf("expected amount 2 (1+1 across two days), got %v", item.Amount)
	}
}

func TestGetShoppingList_DedupsSameNameUnitAcrossRecipes(t *testing.T) {
	recipeA := &domain.Recipe{ID: "rec_a", Servings: 4, Ingredients: []domain.Ingredient{{Name: "Lök", Amount: 1, Unit: "st"}}}
	recipeB := &domain.Recipe{ID: "rec_b", Servings: 4, Ingredients: []domain.Ingredient{{Name: "lök", Amount: 1, Unit: "st"}}}
	menu := &domain.Menu{
		ID: "menu_1",
		Days: []domain.MenuDay{
			{Date: "2026-01-01", RecipeID: "rec_a", Servings: 4},
			{Date: "2026-01-02", RecipeID: "rec_b", Servings: 4},
		},
	}
	svc := NewShoppingService(
		&fakeMenuRepo{menu: menu},
		&fakeRecipeRepo{recipes: map[string]*domain.Recipe{"rec_a": recipeA, "rec_b": recipeB}},
		&fakeShoppingRepo{},
	)

	list, err := svc.GetShoppingList("menu_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count := 0
	for _, cat := range list.Categories {
		count += len(cat.Items)
	}
	if count != 1 {
		t.Errorf("expected 1 deduped item, got %d", count)
	}
}

func TestGetShoppingList_MergesSpicesAcrossUnits(t *testing.T) {
	recipe := &domain.Recipe{
		ID:       "rec_1",
		Servings: 4,
		Ingredients: []domain.Ingredient{
			{Name: "salt", Amount: 1, Unit: "tsk"},
			{Name: "Salt", Amount: 2, Unit: "krm"},
		},
	}
	menu := &domain.Menu{
		ID:   "menu_1",
		Days: []domain.MenuDay{{Date: "2026-01-01", RecipeID: "rec_1", Servings: 4}},
	}
	svc := NewShoppingService(
		&fakeMenuRepo{menu: menu},
		&fakeRecipeRepo{recipes: map[string]*domain.Recipe{"rec_1": recipe}},
		&fakeShoppingRepo{},
	)

	list, err := svc.GetShoppingList("menu_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := findItem(t, list, "salt")
	if item == nil {
		t.Fatalf("expected merged salt item")
	}
	if item.Amount != 0 || item.Unit != "" {
		t.Errorf("expected spice amount/unit dropped, got amount=%v unit=%q", item.Amount, item.Unit)
	}
}

func TestGetShoppingList_NilMenuReturnsNil(t *testing.T) {
	svc := NewShoppingService(
		&fakeMenuRepo{menu: nil},
		&fakeRecipeRepo{recipes: map[string]*domain.Recipe{}},
		&fakeShoppingRepo{},
	)
	list, err := svc.GetShoppingList("missing")
	if err != nil || list != nil {
		t.Fatalf("expected nil,nil got list=%v err=%v", list, err)
	}
}

func TestGetShoppingList_SkipsSkippedAndEmptyDays(t *testing.T) {
	recipe := &domain.Recipe{ID: "rec_1", Servings: 4, Ingredients: []domain.Ingredient{{Name: "Lök", Amount: 1, Unit: "st"}}}
	menu := &domain.Menu{
		ID: "menu_1",
		Days: []domain.MenuDay{
			{Date: "2026-01-01", RecipeID: "rec_1", Servings: 4, Skip: true},
			{Date: "2026-01-02", RecipeID: "", Servings: 4},
		},
	}
	svc := NewShoppingService(
		&fakeMenuRepo{menu: menu},
		&fakeRecipeRepo{recipes: map[string]*domain.Recipe{"rec_1": recipe}},
		&fakeShoppingRepo{},
	)

	list, err := svc.GetShoppingList("menu_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Categories) != 0 {
		t.Errorf("expected no categories, got %d", len(list.Categories))
	}
}
