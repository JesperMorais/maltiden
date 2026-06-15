package sqlite

import (
	"database/sql"
	"sort"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"maltiden/migrations"
)

// newMigratedMemoryDB opens a fresh in-memory SQLite DB and runs the
// embedded migrations against it, returning the live handle.
func newMigratedMemoryDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := runMigrations(db, migrations.FS); err != nil {
		t.Fatalf("runMigrations: %v", err)
	}
	return db
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var got string
	err := db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, name,
	).Scan(&got)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		t.Fatalf("query table %q: %v", name, err)
	}
	return got == name
}

func indexExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var got string
	err := db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type = 'index' AND name = ?`, name,
	).Scan(&got)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		t.Fatalf("query index %q: %v", name, err)
	}
	return got == name
}

func columnNames(t *testing.T, db *sql.DB, table string) map[string]bool {
	t.Helper()
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		t.Fatalf("PRAGMA table_info(%s): %v", table, err)
	}
	defer rows.Close()

	cols := make(map[string]bool)
	for rows.Next() {
		var (
			cid       int
			name      string
			ctype     string
			notnull   int
			dfltValue sql.NullString
			pk        int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			t.Fatalf("scan table_info(%s): %v", table, err)
		}
		cols[name] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err table_info(%s): %v", table, err)
	}
	return cols
}

// TestRunMigrations_CreatesExpectedTables asserts a fresh in-memory DB ends up
// with every table the migration set is responsible for.
func TestRunMigrations_CreatesExpectedTables(t *testing.T) {
	db := newMigratedMemoryDB(t)

	expected := []string{
		"schema_migrations",
		"users",
		"households",
		"household_members",
		"recipes",
		"menus",
		"menu_days",
		"shopping_items",
		"feedback",
		"password_reset_tokens",
	}
	for _, tbl := range expected {
		if !tableExists(t, db, tbl) {
			t.Errorf("expected table %q to exist after migrations", tbl)
		}
	}
}

// TestRunMigrations_RecordsAllVersions asserts every migration version is
// tracked in schema_migrations and matches the count of migration files.
func TestRunMigrations_RecordsAllVersions(t *testing.T) {
	db := newMigratedMemoryDB(t)

	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			t.Fatalf("scan version: %v", err)
		}
		versions = append(versions, v)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}

	// Migrations 1..19 are defined in runMigrations.
	want := make([]int, 0, 19)
	for v := 1; v <= 19; v++ {
		want = append(want, v)
	}
	if !sort.IntsAreSorted(versions) {
		t.Fatalf("versions not returned sorted: %v", versions)
	}
	if len(versions) != len(want) {
		t.Fatalf("expected %d applied migrations, got %d (%v)", len(want), len(versions), versions)
	}
	for i, v := range want {
		if versions[i] != v {
			t.Errorf("version[%d] = %d, want %d", i, versions[i], v)
		}
	}
}

// TestRunMigrations_ExpectedColumns spot-checks columns added across multiple
// migrations to confirm schema-init applied the later ALTERs/columns, not just
// the initial CREATE TABLE statements.
func TestRunMigrations_ExpectedColumns(t *testing.T) {
	db := newMigratedMemoryDB(t)

	cases := []struct {
		table string
		cols  []string
	}{
		{"users", []string{"id", "email", "password_hash", "name", "household_id", "created_at", "token_version"}},
		{"households", []string{"id", "name", "created_at"}},
		{"menu_days", []string{"id", "menu_id", "date", "recipe_id", "servings", "skip"}},
		{"recipes", []string{"id", "household_id"}},
		{"schema_migrations", []string{"version", "applied_at"}},
	}
	for _, c := range cases {
		cols := columnNames(t, db, c.table)
		for _, col := range c.cols {
			if !cols[col] {
				t.Errorf("expected column %q on table %q, present columns: %v", col, c.table, cols)
			}
		}
	}
}

// TestRunMigrations_Phase0Columns asserts the Phase 0 (issue #248) migrations
// added the recipe metadata columns and the preferences diet columns, and that
// a pre-existing recipe row reads back with the column defaults.
func TestRunMigrations_Phase0Columns(t *testing.T) {
	db := newMigratedMemoryDB(t)

	recipeCols := columnNames(t, db, "recipes")
	for _, col := range []string{"main_protein", "diet_class", "batchable", "cook_minutes"} {
		if !recipeCols[col] {
			t.Errorf("expected recipes column %q after migration 017, present: %v", col, recipeCols)
		}
	}

	prefCols := columnNames(t, db, "menu_preferences")
	for _, col := range []string{"diet_profile", "disliked_ingredients"} {
		if !prefCols[col] {
			t.Errorf("expected menu_preferences column %q after migration 018, present: %v", col, prefCols)
		}
	}

	// A row inserted without the new columns must read back with defaults.
	_, err := db.Exec(`INSERT INTO recipes (id, name, servings, ingredients, instructions)
	                   VALUES ('rec_phase0', 'Test', 4, '[]', '[]')`)
	if err != nil {
		t.Fatalf("insert legacy-shaped recipe: %v", err)
	}
	var (
		mainProtein, dietClass string
		batchable, cookMinutes int
	)
	err = db.QueryRow(`SELECT main_protein, diet_class, batchable, cook_minutes
	                   FROM recipes WHERE id = 'rec_phase0'`).
		Scan(&mainProtein, &dietClass, &batchable, &cookMinutes)
	if err != nil {
		t.Fatalf("read back metadata defaults: %v", err)
	}
	if mainProtein != "" || dietClass != "" || batchable != 0 || cookMinutes != 0 {
		t.Errorf("expected metadata defaults, got mainProtein=%q dietClass=%q batchable=%d cookMinutes=%d",
			mainProtein, dietClass, batchable, cookMinutes)
	}

	if reached := schemaVersion(t, db); reached != 19 {
		t.Errorf("expected schema to reach version 19, got %d", reached)
	}
}

// TestRunMigrations_Phase2Columns asserts the Phase 2 (issue #248, meal prep)
// migration added the menu_days leftover/cook_date columns and the
// menu_preferences prep_mode_default column, with the documented defaults on a
// legacy-shaped row.
func TestRunMigrations_Phase2Columns(t *testing.T) {
	db := newMigratedMemoryDB(t)

	menuDayCols := columnNames(t, db, "menu_days")
	for _, col := range []string{"leftover", "cook_date"} {
		if !menuDayCols[col] {
			t.Errorf("expected menu_days column %q after migration 019, present: %v", col, menuDayCols)
		}
	}

	prefCols := columnNames(t, db, "menu_preferences")
	if !prefCols["prep_mode_default"] {
		t.Errorf("expected menu_preferences column prep_mode_default after migration 019, present: %v", prefCols)
	}

	// A menu_days row inserted without the new columns must read back with defaults.
	if _, err := db.Exec(`INSERT INTO households (id, name) VALUES ('hh_p2', 'P2')`); err != nil {
		t.Fatalf("insert household: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO menus (id, household_id) VALUES ('menu_p2', 'hh_p2')`); err != nil {
		t.Fatalf("insert menu: %v", err)
	}
	_, err := db.Exec(`INSERT INTO menu_days (id, menu_id, date, servings)
	                   VALUES ('md_p2', 'menu_p2', '2026-01-01', 4)`)
	if err != nil {
		t.Fatalf("insert legacy-shaped menu_day: %v", err)
	}
	var (
		leftover int
		cookDate sql.NullString
	)
	err = db.QueryRow(`SELECT leftover, cook_date FROM menu_days WHERE id = 'md_p2'`).
		Scan(&leftover, &cookDate)
	if err != nil {
		t.Fatalf("read back menu_day defaults: %v", err)
	}
	if leftover != 0 || cookDate.Valid {
		t.Errorf("expected leftover=0 cook_date=NULL, got leftover=%d cookDate=%v", leftover, cookDate)
	}
}

func schemaVersion(t *testing.T, db *sql.DB) int {
	t.Helper()
	var v int
	if err := db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&v); err != nil {
		t.Fatalf("max schema version: %v", err)
	}
	return v
}

// TestRunMigrations_ExpectedIndexes asserts named indexes created by the
// migrations exist (covers the dedicated index migration #007 plus inline ones).
func TestRunMigrations_ExpectedIndexes(t *testing.T) {
	db := newMigratedMemoryDB(t)

	expected := []string{
		"idx_users_email",
		"idx_users_household",
		"idx_household_members_household",
		"idx_menus_household",
		"idx_menu_days_menu",
		"idx_menu_days_menu_date", // migration 007
		"idx_shopping_items_menu",
		"idx_recipes_diet_class",   // migration 017
		"idx_recipes_main_protein", // migration 017
	}
	for _, idx := range expected {
		if !indexExists(t, db, idx) {
			t.Errorf("expected index %q to exist after migrations", idx)
		}
	}
}

// TestRunMigrations_Idempotent asserts running migrations twice against the same
// DB does not error and does not duplicate version rows.
func TestRunMigrations_Idempotent(t *testing.T) {
	db := newMigratedMemoryDB(t)

	countVersions := func() int {
		var n int
		if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&n); err != nil {
			t.Fatalf("count schema_migrations: %v", err)
		}
		return n
	}

	first := countVersions()

	// Second run must be a no-op (all versions already applied).
	if err := runMigrations(db, migrations.FS); err != nil {
		t.Fatalf("second runMigrations: %v", err)
	}
	second := countVersions()

	if first != second {
		t.Errorf("idempotency violated: version count changed from %d to %d", first, second)
	}

	// And a third run for good measure — still no error, still stable.
	if err := runMigrations(db, migrations.FS); err != nil {
		t.Fatalf("third runMigrations: %v", err)
	}
	if third := countVersions(); third != first {
		t.Errorf("idempotency violated on third run: %d != %d", third, first)
	}
}
