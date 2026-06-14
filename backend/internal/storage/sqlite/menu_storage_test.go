package sqlite

import (
	"testing"
	"time"

	"maltiden/internal/domain"
)

func setupMenuTestDB(t *testing.T) (*MenuStorage, string, string) {
	t.Helper()
	db, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	const hhID = "hh_menu_test"
	if _, err := db.Exec(`INSERT INTO households (id, name) VALUES (?, ?)`, hhID, "Menu Test HH"); err != nil {
		t.Fatalf("seed household: %v", err)
	}

	// Seed a recipe for use in menu_days
	recID := "rec_menu_seed"
	rs := NewRecipeStorage(db)
	seedRec := &domain.Recipe{
		ID:           recID,
		Name:         "Seed Recipe",
		Servings:     2,
		Tags:         []string{},
		Ingredients:  []domain.Ingredient{},
		Instructions: []string{"step"},
		HouseholdID:  hhID,
		CreatedAt:    time.Now().UTC().Truncate(time.Second),
	}
	if err := rs.Create(seedRec); err != nil {
		t.Fatalf("seed recipe: %v", err)
	}

	return NewMenuStorage(db), hhID, recID
}

func newTestMenu(id, hhID, recID string) *domain.Menu {
	return &domain.Menu{
		ID:          id,
		HouseholdID: hhID,
		CreatedAt:   time.Now().UTC().Truncate(time.Second),
		Days: []domain.MenuDay{
			{Date: "2026-06-16", RecipeID: recID, Servings: 2},
			{Date: "2026-06-17", Skip: true, Servings: 0},
		},
	}
}

func TestMenuStorage_CreateGetByID(t *testing.T) {
	s, hhID, recID := setupMenuTestDB(t)
	m := newTestMenu("menu_1", hhID, recID)

	if err := s.Create(m); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.GetByID(m.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got == nil {
		t.Fatal("GetByID: expected menu, got nil")
	}
	if got.ID != m.ID || got.HouseholdID != m.HouseholdID {
		t.Errorf("field mismatch: got %+v", got)
	}
	if len(got.Days) != 2 {
		t.Errorf("expected 2 days, got %d", len(got.Days))
	}
	if got.Days[0].RecipeID != recID {
		t.Errorf("recipe_id not preserved: %q", got.Days[0].RecipeID)
	}
	if !got.Days[1].Skip {
		t.Error("skip flag not preserved")
	}
}

func TestMenuStorage_GetByID_Missing(t *testing.T) {
	s, _, _ := setupMenuTestDB(t)
	got, err := s.GetByID("menu_does_not_exist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestMenuStorage_Update(t *testing.T) {
	s, hhID, recID := setupMenuTestDB(t)
	m := newTestMenu("menu_upd", hhID, recID)
	if err := s.Create(m); err != nil {
		t.Fatalf("Create: %v", err)
	}

	m.Days = []domain.MenuDay{
		{Date: "2026-06-20", RecipeID: recID, Servings: 4},
	}
	if err := s.Update(m); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := s.GetByID(m.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if len(got.Days) != 1 {
		t.Errorf("expected 1 day after update, got %d", len(got.Days))
	}
	if got.Days[0].Servings != 4 {
		t.Errorf("servings not updated: %d", got.Days[0].Servings)
	}
}

func TestMenuStorage_GetCurrentByHousehold_Hit(t *testing.T) {
	s, hhID, recID := setupMenuTestDB(t)
	m := newTestMenu("menu_curr", hhID, recID)
	if err := s.Create(m); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.GetCurrentByHousehold(hhID)
	if err != nil {
		t.Fatalf("GetCurrentByHousehold: %v", err)
	}
	if got == nil {
		t.Fatal("expected menu, got nil")
	}
	if got.ID != m.ID {
		t.Errorf("wrong menu returned: %q", got.ID)
	}
}

func TestMenuStorage_GetCurrentByHousehold_Miss(t *testing.T) {
	s, _, _ := setupMenuTestDB(t)
	got, err := s.GetCurrentByHousehold("hh_no_menu")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestMenuStorage_GetHouseholdIDByMenuID_Hit(t *testing.T) {
	s, hhID, recID := setupMenuTestDB(t)
	m := newTestMenu("menu_hhid", hhID, recID)
	if err := s.Create(m); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.GetHouseholdIDByMenuID(m.ID)
	if err != nil {
		t.Fatalf("GetHouseholdIDByMenuID: %v", err)
	}
	if got != hhID {
		t.Errorf("wrong householdID: got %q, want %q", got, hhID)
	}
}

func TestMenuStorage_GetHouseholdIDByMenuID_Miss(t *testing.T) {
	s, _, _ := setupMenuTestDB(t)
	got, err := s.GetHouseholdIDByMenuID("menu_gone")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestMenuStorage_Create_FK_Violation(t *testing.T) {
	s, _, recID := setupMenuTestDB(t)
	m := newTestMenu("menu_fk", "hh_nonexistent", recID)
	err := s.Create(m)
	if err == nil {
		t.Fatal("expected FK error, got nil")
	}
}
