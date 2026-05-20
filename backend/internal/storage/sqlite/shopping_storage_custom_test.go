package sqlite

import (
	"database/sql"
	"testing"

	"maltiden/internal/domain"
)

func setupCustomTestDB(t *testing.T) (*sql.DB, string, string) {
	t.Helper()

	tmpFile := t.TempDir() + "/test.db"
	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	hhA := "hh_custom_a"
	hhB := "hh_custom_b"
	for _, hh := range []struct{ id, name string }{
		{hhA, "Household A"},
		{hhB, "Household B"},
	} {
		if _, err := db.Exec(`INSERT INTO households (id, name) VALUES (?, ?)`, hh.id, hh.name); err != nil {
			t.Fatalf("seed household %s: %v", hh.id, err)
		}
	}

	menuID := "menu_custom_test"
	if _, err := db.Exec(`INSERT INTO menus (id, household_id) VALUES (?, ?)`, menuID, hhA); err != nil {
		t.Fatalf("seed menu: %v", err)
	}

	return db, hhA, hhB
}

func TestShoppingStorage_CustomItems(t *testing.T) {
	db, hhA, hhB := setupCustomTestDB(t)
	storage := NewShoppingStorage(db)

	menuID := "menu_custom_test"

	// Insert 2 custom items for hh_a with IDs that sort lexicographically to guarantee order.
	itemA1 := &domain.CustomShoppingItem{
		ID:          "sci_a_001",
		MenuID:      menuID,
		HouseholdID: hhA,
		Name:        "Mjölk",
		Unit:        "liter",
		Amount:      1.5,
	}
	itemA2 := &domain.CustomShoppingItem{
		ID:          "sci_a_002",
		MenuID:      menuID,
		HouseholdID: hhA,
		Name:        "Bröd",
		Unit:        "st",
		Amount:      2,
	}
	itemB1 := &domain.CustomShoppingItem{
		ID:          "sci_b_001",
		MenuID:      menuID,
		HouseholdID: hhB,
		Name:        "Smör",
		Unit:        "g",
		Amount:      250,
	}

	if err := storage.CreateCustomItem(itemA1); err != nil {
		t.Fatalf("CreateCustomItem a1: %v", err)
	}
	if err := storage.CreateCustomItem(itemA2); err != nil {
		t.Fatalf("CreateCustomItem a2: %v", err)
	}
	if err := storage.CreateCustomItem(itemB1); err != nil {
		t.Fatalf("CreateCustomItem b1: %v", err)
	}

	// hh_a should see exactly 2 items.
	items, err := storage.GetCustomItems(menuID, hhA)
	if err != nil {
		t.Fatalf("GetCustomItems hh_a: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items for hh_a, got %d", len(items))
	}

	// Verify insertion order (ORDER BY created_at, id → sci_a_001 first).
	if items[0].ID != itemA1.ID {
		t.Errorf("expected first item ID %q, got %q", itemA1.ID, items[0].ID)
	}
	if items[1].ID != itemA2.ID {
		t.Errorf("expected second item ID %q, got %q", itemA2.ID, items[1].ID)
	}

	// Verify fields round-trip correctly.
	got := items[0]
	if got.MenuID != itemA1.MenuID {
		t.Errorf("MenuID: want %q, got %q", itemA1.MenuID, got.MenuID)
	}
	if got.HouseholdID != itemA1.HouseholdID {
		t.Errorf("HouseholdID: want %q, got %q", itemA1.HouseholdID, got.HouseholdID)
	}
	if got.Name != itemA1.Name {
		t.Errorf("Name: want %q, got %q", itemA1.Name, got.Name)
	}
	if got.Unit != itemA1.Unit {
		t.Errorf("Unit: want %q, got %q", itemA1.Unit, got.Unit)
	}
	if got.Amount != itemA1.Amount {
		t.Errorf("Amount: want %v, got %v", itemA1.Amount, got.Amount)
	}
	if got.Checked {
		t.Errorf("Checked: want false, got true")
	}

	// hh_b's item must not appear in hh_a's results.
	for _, item := range items {
		if item.ID == itemB1.ID {
			t.Errorf("hh_b item %q leaked into hh_a results", itemB1.ID)
		}
	}

	// hh_b should see exactly 1 item.
	bItems, err := storage.GetCustomItems(menuID, hhB)
	if err != nil {
		t.Fatalf("GetCustomItems hh_b: %v", err)
	}
	if len(bItems) != 1 {
		t.Fatalf("expected 1 item for hh_b, got %d", len(bItems))
	}
	if bItems[0].ID != itemB1.ID {
		t.Errorf("expected hh_b item ID %q, got %q", itemB1.ID, bItems[0].ID)
	}
}
