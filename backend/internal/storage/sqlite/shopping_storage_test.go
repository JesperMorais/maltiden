package sqlite

import (
	"database/sql"
	"testing"

	"maltiden/internal/domain"
)

func setupShoppingTestDB(t *testing.T) (*sql.DB, string, string) {
	t.Helper()

	tmpFile := t.TempDir() + "/test.db"
	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	householdID := "hh_shopping_test"
	if _, err := db.Exec(
		`INSERT INTO households (id, name) VALUES (?, ?)`,
		householdID, "Shopping Test Household",
	); err != nil {
		t.Fatalf("seed household: %v", err)
	}

	menuID := "menu_shopping_test"
	if _, err := db.Exec(
		`INSERT INTO menus (id, household_id) VALUES (?, ?)`,
		menuID, householdID,
	); err != nil {
		t.Fatalf("seed menu: %v", err)
	}

	return db, menuID, householdID
}

func TestShoppingStorage_CreateCustomItem_CheckedFalse(t *testing.T) {
	db, menuID, householdID := setupShoppingTestDB(t)
	storage := NewShoppingStorage(db)

	item := &domain.CustomShoppingItem{
		ID:          "custom_1",
		MenuID:      menuID,
		HouseholdID: householdID,
		Name:        "Mjölk",
		Unit:        "liter",
		Amount:      1.0,
	}

	if err := storage.CreateCustomItem(item); err != nil {
		t.Fatalf("CreateCustomItem: %v", err)
	}

	items, err := storage.GetCustomItems(menuID, householdID)
	if err != nil {
		t.Fatalf("GetCustomItems: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Checked {
		t.Fatalf("expected Checked=false, got true")
	}
	if items[0].Name != "Mjölk" {
		t.Fatalf("expected Name=Mjölk, got %q", items[0].Name)
	}
}

func TestShoppingStorage_CreateCustomItemWithCap_RespectsCap(t *testing.T) {
	db, menuID, householdID := setupShoppingTestDB(t)
	storage := NewShoppingStorage(db)

	mkItem := func(id, name string) *domain.CustomShoppingItem {
		return &domain.CustomShoppingItem{
			ID:          id,
			MenuID:      menuID,
			HouseholdID: householdID,
			Name:        name,
			Unit:        "st",
			Amount:      1.0,
		}
	}

	ok1, err := storage.CreateCustomItemWithCap(mkItem("c1", "Ägg"), 2)
	if err != nil {
		t.Fatalf("insert 1: %v", err)
	}
	if !ok1 {
		t.Fatal("expected insert 1 to succeed")
	}

	ok2, err := storage.CreateCustomItemWithCap(mkItem("c2", "Smör"), 2)
	if err != nil {
		t.Fatalf("insert 2: %v", err)
	}
	if !ok2 {
		t.Fatal("expected insert 2 to succeed")
	}

	ok3, err := storage.CreateCustomItemWithCap(mkItem("c3", "Ost"), 2)
	if err != nil {
		t.Fatalf("insert 3 error: %v", err)
	}
	if ok3 {
		t.Fatal("expected insert 3 to be rejected by cap, got ok=true")
	}

	items, err := storage.GetCustomItems(menuID, householdID)
	if err != nil {
		t.Fatalf("GetCustomItems: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items after cap hit, got %d", len(items))
	}
}

func TestShoppingStorage_GetCustomItems_OrderedAndScoped(t *testing.T) {
	db, menuID, householdID := setupShoppingTestDB(t)
	storage := NewShoppingStorage(db)

	// Seed second household + menu to test scoping.
	if _, err := db.Exec(
		`INSERT INTO households (id, name) VALUES (?, ?)`,
		"hh_other", "Other Household",
	); err != nil {
		t.Fatalf("seed other household: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO menus (id, household_id) VALUES (?, ?)`,
		"menu_other", "hh_other",
	); err != nil {
		t.Fatalf("seed other menu: %v", err)
	}

	// Insert items for the target menu.
	for _, item := range []*domain.CustomShoppingItem{
		{ID: "i1", MenuID: menuID, HouseholdID: householdID, Name: "Alpha", Unit: "g", Amount: 100},
		{ID: "i2", MenuID: menuID, HouseholdID: householdID, Name: "Beta", Unit: "g", Amount: 200},
	} {
		if err := storage.CreateCustomItem(item); err != nil {
			t.Fatalf("insert %s: %v", item.ID, err)
		}
	}

	// Insert item for other menu — must not appear.
	other := &domain.CustomShoppingItem{
		ID: "i3", MenuID: "menu_other", HouseholdID: "hh_other",
		Name: "Gamma", Unit: "g", Amount: 300,
	}
	if err := storage.CreateCustomItem(other); err != nil {
		t.Fatalf("insert other: %v", err)
	}

	items, err := storage.GetCustomItems(menuID, householdID)
	if err != nil {
		t.Fatalf("GetCustomItems: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "i1" || items[1].ID != "i2" {
		t.Fatalf("expected order [i1, i2], got [%s, %s]", items[0].ID, items[1].ID)
	}
}
