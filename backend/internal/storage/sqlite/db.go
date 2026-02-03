package sqlite

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func Open(path string) (*sql.DB, error) {
	// Create dir if not exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Run Migrations
	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	// Create migrations tracking table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Run all migrations in order
	migrations := []struct {
		version int
		file    string
	}{
		{1, "migrations/001_create_users.sql"},
		{2, "migrations/002_create_households.sql"},
		{3, "migrations/003_create_recipes.sql"},
		{4, "migrations/004_seed_recipes.sql"},
		{5, "migrations/005_create_menus.sql"},
		{6, "migrations/006_household_invites_and_member_status.sql"},
	}

	for _, m := range migrations {
		// Check if migration already applied
		var exists int
		err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", m.version).Scan(&exists)
		if err != nil {
			return err
		}

		if exists > 0 {
			// Migration already applied, skip
			continue
		}

		// Read and execute migration file
		sqlBytes, err := os.ReadFile(m.file)
		if err != nil {
			return err
		}

		_, err = db.Exec(string(sqlBytes))
		if err != nil {
			return err
		}

		// Mark migration as applied
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", m.version)
		if err != nil {
			return err
		}
	}

	return nil
}
