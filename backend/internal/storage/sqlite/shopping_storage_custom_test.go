package sqlite

import (
	"testing"

	"maltiden/internal/domain"
)

func setupCustomTestDB(t *testing.T) (*ShoppingStorage, string, string) {
	t.Helper()

	db, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Exec(`INSERT INTO households (id, name) VALUES (?, ?)`, "hh_custom_a", "Household A"); err != nil {
		t.Fatalf("seed household: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO menus (id, household_id) VALUES (?, ?)`, "menu_custom_a", "hh_custom_a"); err != nil {
		t.Fatalf("seed menu: %v", err)
	}

	return NewShoppingStorage(db), "menu_custom_a", "hh_custom_a"
}

func TestShoppingStorage_Custom_RoundTrip(t *testing.T) {
	storage, menuID, householdID := setupCustomTestDB(t)

	items := []*domain.CustomShoppingItem{
		{ID: "csi_1", MenuID: menuID, HouseholdID: householdID, Name: "Mjölk", Unit: "l", Amount: 1.5},
		{ID: "csi_2", MenuID: menuID, HouseholdID: householdID, Name: "Smör", Unit: "g", Amount: 200},
	}
	for _, item := range items {
		if err := storage.CreateCustomItem(item); err != nil {
			t.Fatalf("CreateCustomItem %s: %v", item.ID, err)
		}
	}

	got, err := storage.GetCustomItems(menuID, householdID)
	if err != nil {
		t.Fatalf("GetCustomItems: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}

	byID := make(map[string]domain.CustomShoppingItem, len(got))
	for _, g := range got {
		byID[g.ID] = g
	}

	for _, want := range items {
		g, ok := byID[want.ID]
		if !ok {
			t.Errorf("item %s not found in result", want.ID)
			continue
		}
		if g.MenuID != want.MenuID {
			t.Errorf("%s: MenuID got %q, want %q", want.ID, g.MenuID, want.MenuID)
		}
		if g.HouseholdID != want.HouseholdID {
			t.Errorf("%s: HouseholdID got %q, want %q", want.ID, g.HouseholdID, want.HouseholdID)
		}
		if g.Name != want.Name {
			t.Errorf("%s: Name got %q, want %q", want.ID, g.Name, want.Name)
		}
		if g.Unit != want.Unit {
			t.Errorf("%s: Unit got %q, want %q", want.ID, g.Unit, want.Unit)
		}
		if g.Amount != want.Amount {
			t.Errorf("%s: Amount got %v, want %v", want.ID, g.Amount, want.Amount)
		}
		if g.Checked {
			t.Errorf("%s: Checked should be false after insert", want.ID)
		}
	}
}

func TestShoppingStorage_Custom_HouseholdIsolation(t *testing.T) {
	db, err := Open(t.TempDir() + "/isolation.db")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	for _, hh := range []struct{ id, name string }{
		{"hh_custom_a", "Household A"},
		{"hh_custom_b", "Household B"},
	} {
		if _, err := db.Exec(`INSERT INTO households (id, name) VALUES (?, ?)`, hh.id, hh.name); err != nil {
			t.Fatalf("seed household %s: %v", hh.id, err)
		}
	}
	for _, m := range []struct{ id, hhID string }{
		{"menu_custom_a", "hh_custom_a"},
		{"menu_custom_b", "hh_custom_b"},
	} {
		if _, err := db.Exec(`INSERT INTO menus (id, household_id) VALUES (?, ?)`, m.id, m.hhID); err != nil {
			t.Fatalf("seed menu %s: %v", m.id, err)
		}
	}

	storage := NewShoppingStorage(db)

	itemA := &domain.CustomShoppingItem{ID: "csi_a1", MenuID: "menu_custom_a", HouseholdID: "hh_custom_a", Name: "Ägg", Unit: "st", Amount: 6}
	if err := storage.CreateCustomItem(itemA); err != nil {
		t.Fatalf("CreateCustomItem A: %v", err)
	}

	itemB := &domain.CustomShoppingItem{ID: "csi_b1", MenuID: "menu_custom_b", HouseholdID: "hh_custom_b", Name: "Ost", Unit: "g", Amount: 100}
	if err := storage.CreateCustomItem(itemB); err != nil {
		t.Fatalf("CreateCustomItem B: %v", err)
	}

	got, err := storage.GetCustomItems("menu_custom_a", "hh_custom_a")
	if err != nil {
		t.Fatalf("GetCustomItems: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 item for household A, got %d", len(got))
	}
	if got[0].ID != "csi_a1" {
		t.Errorf("expected item csi_a1, got %s", got[0].ID)
	}
}
