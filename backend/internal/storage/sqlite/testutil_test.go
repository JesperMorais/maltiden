package sqlite

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"maltiden/migrations"
)

// setupTestDB opens an in-memory SQLite database and applies all migrations.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := runMigrations(db, migrations.FS); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return db
}

// seedTestUser inserts a user row and returns its ID.
func seedTestUser(t *testing.T, db *sql.DB, id, email, name, householdID string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO users (id, email, password_hash, name, household_id) VALUES (?, ?, ?, ?, ?)`,
		id, email, "$2a$12$placeholder", name, householdID,
	)
	if err != nil {
		t.Fatalf("seedTestUser: %v", err)
	}
}

// seedTestHousehold inserts a household and an owner membership row.
func seedTestHousehold(t *testing.T, db *sql.DB, householdID, name, ownerUserID string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO households (id, name) VALUES (?, ?)`, householdID, name)
	if err != nil {
		t.Fatalf("seedTestHousehold (household): %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO household_members (id, household_id, user_id, role) VALUES (?, ?, ?, 'owner')`,
		"hm_"+householdID, householdID, ownerUserID,
	)
	if err != nil {
		t.Fatalf("seedTestHousehold (member): %v", err)
	}
}

// seedTestRecipe inserts a minimal recipe row and returns its ID.
func seedTestRecipe(t *testing.T, db *sql.DB, id, name string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO recipes (id, name, servings, tags, ingredients, instructions) VALUES (?, ?, 4, '[]', '[]', '[]')`,
		id, name,
	)
	if err != nil {
		t.Fatalf("seedTestRecipe: %v", err)
	}
}

// seedTestMenu inserts a menu row and returns its ID.
func seedTestMenu(t *testing.T, db *sql.DB, id, householdID string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO menus (id, household_id) VALUES (?, ?)`, id, householdID)
	if err != nil {
		t.Fatalf("seedTestMenu: %v", err)
	}
}

func TestSetupTestDB(t *testing.T) {
	db := setupTestDB(t)

	var fkEnabled int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}
	if fkEnabled != 1 {
		t.Errorf("expected foreign_keys=1, got %d", fkEnabled)
	}
}
