package sqlite

import (
	"testing"
	"time"

	"maltiden/internal/domain"
)

func setupRecipeTestDB(t *testing.T) (*RecipeStorage, string) {
	t.Helper()
	db, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	const hhID = "hh_rec_test"
	if _, err := db.Exec(`INSERT INTO households (id, name) VALUES (?, ?)`, hhID, "Recipe Test HH"); err != nil {
		t.Fatalf("seed household: %v", err)
	}
	return NewRecipeStorage(db), hhID
}

func newTestRecipe(id, hhID string) *domain.Recipe {
	return &domain.Recipe{
		ID:           id,
		Name:         "Pasta Carbonara",
		Servings:     4,
		Emoji:        "🍝",
		Tags:         []string{"italian", "quick"},
		Ingredients:  []domain.Ingredient{{Name: "pasta", Amount: 200, Unit: "g"}},
		Instructions: []string{"Boil water", "Cook pasta"},
		HouseholdID:  hhID,
		CreatedAt:    time.Now().UTC().Truncate(time.Second),
	}
}

func TestRecipeStorage_CreateGetByID(t *testing.T) {
	s, hhID := setupRecipeTestDB(t)
	r := newTestRecipe("rec_test_cg1", hhID)

	if err := s.Create(r); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.GetByID(r.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got == nil {
		t.Fatal("GetByID: expected recipe, got nil")
	}
	if got.ID != r.ID || got.Name != r.Name || got.Servings != r.Servings {
		t.Errorf("field mismatch: got %+v", got)
	}
	if len(got.Ingredients) != 1 || got.Ingredients[0].Name != "pasta" {
		t.Errorf("ingredients not preserved: %+v", got.Ingredients)
	}
	if len(got.Instructions) != 2 || got.Instructions[0] != "Boil water" {
		t.Errorf("instructions not preserved: %+v", got.Instructions)
	}
}

func TestRecipeStorage_GetByID_Missing(t *testing.T) {
	s, _ := setupRecipeTestDB(t)
	got, err := s.GetByID("rec_does_not_exist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestRecipeStorage_Update(t *testing.T) {
	s, hhID := setupRecipeTestDB(t)
	r := newTestRecipe("rec_upd", hhID)
	if err := s.Create(r); err != nil {
		t.Fatalf("Create: %v", err)
	}

	r.Name = "Updated Pasta"
	r.Servings = 2
	r.Instructions = []string{"Step 1"}
	if err := s.Update(r); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := s.GetByID(r.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "Updated Pasta" || got.Servings != 2 {
		t.Errorf("update not persisted: %+v", got)
	}
	if len(got.Instructions) != 1 {
		t.Errorf("instructions not updated: %+v", got.Instructions)
	}
}

func TestRecipeStorage_Delete(t *testing.T) {
	s, hhID := setupRecipeTestDB(t)
	r := newTestRecipe("rec_del", hhID)
	if err := s.Create(r); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := s.Delete(r.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	got, err := s.GetByID(r.ID)
	if err != nil {
		t.Fatalf("GetByID after delete: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestRecipeStorage_GetAll(t *testing.T) {
	s, hhID := setupRecipeTestDB(t)
	for i, id := range []string{"rec_all_1", "rec_all_2"} {
		r := newTestRecipe(id, hhID)
		r.Name = "Recipe " + string(rune('A'+i))
		if err := s.Create(r); err != nil {
			t.Fatalf("Create %s: %v", id, err)
		}
	}

	recipes, err := s.GetAll(nil, hhID)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(recipes) < 2 {
		t.Errorf("expected >=2 recipes, got %d", len(recipes))
	}
}

func TestRecipeStorage_GetAllPaginated(t *testing.T) {
	s, hhID := setupRecipeTestDB(t)
	for _, id := range []string{"rec_pag_1", "rec_pag_2", "rec_pag_3"} {
		r := newTestRecipe(id, hhID)
		if err := s.Create(r); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	recipes, total, err := s.GetAllPaginated(nil, hhID, 2, 0)
	if err != nil {
		t.Fatalf("GetAllPaginated: %v", err)
	}
	if total < 3 {
		t.Errorf("expected total>=3, got %d", total)
	}
	if len(recipes) != 2 {
		t.Errorf("expected page of 2, got %d", len(recipes))
	}
}

func TestRecipeStorage_GetByIDs(t *testing.T) {
	s, hhID := setupRecipeTestDB(t)
	ids := []string{"rec_map_1", "rec_map_2"}
	for _, id := range ids {
		r := newTestRecipe(id, hhID)
		if err := s.Create(r); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	result, err := s.GetByIDs(ids)
	if err != nil {
		t.Fatalf("GetByIDs: %v", err)
	}
	for _, id := range ids {
		if _, ok := result[id]; !ok {
			t.Errorf("missing %s in result map", id)
		}
	}
}

func fptr(v float64) *float64 { return &v }

func TestRecipeStorage_NutritionRoundTrip(t *testing.T) {
	s, hhID := setupRecipeTestDB(t)
	r := newTestRecipe("rec_nutri", hhID)
	r.Nutrition = &domain.Nutrition{
		Calories: fptr(520),
		ProteinG: fptr(31.5),
		CarbsG:   fptr(48),
		FatG:     fptr(18.2),
	}

	if err := s.Create(r); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.GetByID(r.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Nutrition == nil {
		t.Fatal("expected nutrition to round-trip, got nil")
	}
	if got.Nutrition.Calories == nil || *got.Nutrition.Calories != 520 {
		t.Errorf("calories not preserved: %+v", got.Nutrition.Calories)
	}
	if got.Nutrition.ProteinG == nil || *got.Nutrition.ProteinG != 31.5 {
		t.Errorf("protein not preserved: %+v", got.Nutrition.ProteinG)
	}
	if got.Nutrition.CarbsG == nil || *got.Nutrition.CarbsG != 48 {
		t.Errorf("carbs not preserved: %+v", got.Nutrition.CarbsG)
	}
	if got.Nutrition.FatG == nil || *got.Nutrition.FatG != 18.2 {
		t.Errorf("fat not preserved: %+v", got.Nutrition.FatG)
	}

	// GetByIDs should also surface nutrition.
	m, err := s.GetByIDs([]string{r.ID})
	if err != nil {
		t.Fatalf("GetByIDs: %v", err)
	}
	if m[r.ID] == nil || m[r.ID].Nutrition == nil || m[r.ID].Nutrition.Calories == nil {
		t.Errorf("GetByIDs did not surface nutrition: %+v", m[r.ID])
	}
}

func TestRecipeStorage_NutritionNullableAndPartial(t *testing.T) {
	s, hhID := setupRecipeTestDB(t)

	// A recipe with no nutrition data at all stays valid; Nutrition is nil.
	noData := newTestRecipe("rec_nutri_none", hhID)
	if err := s.Create(noData); err != nil {
		t.Fatalf("Create no-data: %v", err)
	}
	got, err := s.GetByID(noData.ID)
	if err != nil {
		t.Fatalf("GetByID no-data: %v", err)
	}
	if got.Nutrition != nil {
		t.Errorf("expected nil nutrition for recipe without data, got %+v", got.Nutrition)
	}

	// Partial nutrition (only protein) round-trips with the rest NULL.
	partial := newTestRecipe("rec_nutri_partial", hhID)
	partial.Nutrition = &domain.Nutrition{ProteinG: fptr(40)}
	if err := s.Create(partial); err != nil {
		t.Fatalf("Create partial: %v", err)
	}
	gotP, err := s.GetByID(partial.ID)
	if err != nil {
		t.Fatalf("GetByID partial: %v", err)
	}
	if gotP.Nutrition == nil || gotP.Nutrition.ProteinG == nil || *gotP.Nutrition.ProteinG != 40 {
		t.Fatalf("partial protein not preserved: %+v", gotP.Nutrition)
	}
	if gotP.Nutrition.Calories != nil || gotP.Nutrition.CarbsG != nil || gotP.Nutrition.FatG != nil {
		t.Errorf("expected unset fields to stay nil: %+v", gotP.Nutrition)
	}
}

func TestRecipeStorage_NutritionUpdate(t *testing.T) {
	s, hhID := setupRecipeTestDB(t)
	r := newTestRecipe("rec_nutri_upd", hhID)
	if err := s.Create(r); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Add nutrition via Update.
	r.Nutrition = &domain.Nutrition{Calories: fptr(600), ProteinG: fptr(25)}
	if err := s.Update(r); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err := s.GetByID(r.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Nutrition == nil || got.Nutrition.Calories == nil || *got.Nutrition.Calories != 600 {
		t.Fatalf("nutrition not persisted on update: %+v", got.Nutrition)
	}

	// Clearing nutrition via Update writes NULLs back.
	r.Nutrition = nil
	if err := s.Update(r); err != nil {
		t.Fatalf("Update clear: %v", err)
	}
	gotCleared, err := s.GetByID(r.ID)
	if err != nil {
		t.Fatalf("GetByID after clear: %v", err)
	}
	if gotCleared.Nutrition != nil {
		t.Errorf("expected nutrition cleared to nil, got %+v", gotCleared.Nutrition)
	}
}

func TestRecipeStorage_Create_DuplicateID(t *testing.T) {
	s, hhID := setupRecipeTestDB(t)
	r := newTestRecipe("rec_dup_test", hhID)
	if err := s.Create(r); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	err := s.Create(r)
	if err == nil {
		t.Fatal("expected UNIQUE error on duplicate id, got nil")
	}
}
