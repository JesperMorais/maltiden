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
