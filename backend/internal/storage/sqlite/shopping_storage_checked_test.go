package sqlite

import (
	"database/sql"
	"testing"
)

// setupCheckedTestDB opens a fresh SQLite DB (with all migrations applied)
// and seeds a household + menu so shopping_items rows can satisfy their
// FK constraint to menus.id.
func setupCheckedTestDB(t *testing.T) (*sql.DB, string) {
	t.Helper()

	tmpFile := t.TempDir() + "/test.db"
	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	householdID := "hh_checked_test"
	if _, err := db.Exec(
		`INSERT INTO households (id, name) VALUES (?, ?)`,
		householdID, "Checked Test Household",
	); err != nil {
		t.Fatalf("seed household: %v", err)
	}

	menuID := "menu_checked_test"
	if _, err := db.Exec(
		`INSERT INTO menus (id, household_id) VALUES (?, ?)`,
		menuID, householdID,
	); err != nil {
		t.Fatalf("seed menu: %v", err)
	}

	return db, menuID
}

func TestShoppingStorage_SetChecked_TrueThenFalse(t *testing.T) {
	db, menuID := setupCheckedTestDB(t)
	storage := NewShoppingStorage(db)

	itemID := "tomat"

	// Initially unknown items are not checked.
	checked, err := storage.IsChecked(menuID, itemID)
	if err != nil {
		t.Fatalf("IsChecked before set: %v", err)
	}
	if checked {
		t.Fatalf("expected IsChecked=false before any SetChecked, got true")
	}

	// Set true → IsChecked returns true.
	if err := storage.SetChecked(menuID, itemID, true); err != nil {
		t.Fatalf("SetChecked(true): %v", err)
	}
	checked, err = storage.IsChecked(menuID, itemID)
	if err != nil {
		t.Fatalf("IsChecked after SetChecked(true): %v", err)
	}
	if !checked {
		t.Fatalf("expected IsChecked=true after SetChecked(true), got false")
	}

	// Flip back to false → IsChecked returns false (upsert path).
	if err := storage.SetChecked(menuID, itemID, false); err != nil {
		t.Fatalf("SetChecked(false): %v", err)
	}
	checked, err = storage.IsChecked(menuID, itemID)
	if err != nil {
		t.Fatalf("IsChecked after SetChecked(false): %v", err)
	}
	if checked {
		t.Fatalf("expected IsChecked=false after SetChecked(false), got true")
	}
}

func TestShoppingStorage_IsChecked_UnknownItem(t *testing.T) {
	db, menuID := setupCheckedTestDB(t)
	storage := NewShoppingStorage(db)

	checked, err := storage.IsChecked(menuID, "never-inserted")
	if err != nil {
		t.Fatalf("IsChecked unknown: %v", err)
	}
	if checked {
		t.Fatalf("expected IsChecked=false for unknown item, got true")
	}
}
