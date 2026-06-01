package sqlite

import (
	"testing"

	"maltiden/internal/domain"
)

func newTestCustomItem(menuID, householdID string) *domain.CustomShoppingItem {
	return &domain.CustomShoppingItem{
		ID:          "csi_test_001",
		MenuID:      menuID,
		HouseholdID: householdID,
		Name:        "Mjölk",
		Unit:        "liter",
		Amount:      2.5,
	}
}

func TestShoppingStorage_CreateCustomItem_PersistsFields(t *testing.T) {
	db, menuID := setupCheckedTestDB(t)
	storage := NewShoppingStorage(db)

	householdID := "hh_checked_test"
	item := newTestCustomItem(menuID, householdID)

	if err := storage.CreateCustomItem(item); err != nil {
		t.Fatalf("CreateCustomItem: %v", err)
	}

	items, err := storage.GetCustomItems(menuID, householdID)
	if err != nil {
		t.Fatalf("GetCustomItems: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("GetCustomItems: expected 1 item, got %d", len(items))
	}

	got := items[0]
	if got.ID != item.ID {
		t.Errorf("ID: got %q, want %q", got.ID, item.ID)
	}
	if got.Name != item.Name {
		t.Errorf("Name: got %q, want %q", got.Name, item.Name)
	}
	if got.Unit != item.Unit {
		t.Errorf("Unit: got %q, want %q", got.Unit, item.Unit)
	}
	if got.Amount != item.Amount {
		t.Errorf("Amount: got %v, want %v", got.Amount, item.Amount)
	}
	if got.Checked {
		t.Error("Checked: expected false on new item, got true")
	}
}

func TestShoppingStorage_CreateCustomItem_IsolatedByHousehold(t *testing.T) {
	db, menuID := setupCheckedTestDB(t)
	storage := NewShoppingStorage(db)

	householdID := "hh_checked_test"
	item := newTestCustomItem(menuID, householdID)

	if err := storage.CreateCustomItem(item); err != nil {
		t.Fatalf("CreateCustomItem: %v", err)
	}

	// A different household should see no items for the same menu.
	items, err := storage.GetCustomItems(menuID, "hh_other")
	if err != nil {
		t.Fatalf("GetCustomItems other household: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for other household, got %d", len(items))
	}
}

func TestShoppingStorage_CreateCustomItemWithCap_EnforcesLimit(t *testing.T) {
	db, menuID := setupCheckedTestDB(t)
	storage := NewShoppingStorage(db)

	householdID := "hh_checked_test"

	first := &domain.CustomShoppingItem{
		ID:          "csi_cap_001",
		MenuID:      menuID,
		HouseholdID: householdID,
		Name:        "Smör",
		Unit:        "st",
		Amount:      1,
	}
	inserted, err := storage.CreateCustomItemWithCap(first, 1)
	if err != nil {
		t.Fatalf("CreateCustomItemWithCap first: %v", err)
	}
	if !inserted {
		t.Fatal("expected first insert to succeed, got false")
	}

	second := &domain.CustomShoppingItem{
		ID:          "csi_cap_002",
		MenuID:      menuID,
		HouseholdID: householdID,
		Name:        "Ost",
		Unit:        "st",
		Amount:      1,
	}
	inserted, err = storage.CreateCustomItemWithCap(second, 1)
	if err != nil {
		t.Fatalf("CreateCustomItemWithCap second: %v", err)
	}
	if inserted {
		t.Fatal("expected second insert to be rejected by cap, got true")
	}
}
