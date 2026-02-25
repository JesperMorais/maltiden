package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"maltiden/migrations"
)

func Open(path string) (*sql.DB, error) {
	// Create dir if not exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Enable foreign keys (verification step for pooled connections)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	// Enable WAL mode for concurrent reads during writes
	var walMode string
	if err := db.QueryRow("PRAGMA journal_mode=WAL").Scan(&walMode); err != nil {
		return nil, fmt.Errorf("enable WAL mode: %w", err)
	}
	if walMode != "wal" {
		return nil, fmt.Errorf("failed to enable WAL mode, got: %s", walMode)
	}

	// Safe with WAL, reduces fsync calls
	if _, err := db.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		return nil, fmt.Errorf("set synchronous mode: %w", err)
	}

	// Wait 5s instead of failing immediately on lock contention
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}

	// Configure connection pool (conservative for SQLite single-writer)
	db.SetMaxOpenConns(10)    // single writer, multiple readers with WAL
	db.SetMaxIdleConns(5)     // keep warm connections close to max to avoid churn
	db.SetConnMaxLifetime(5 * time.Minute) // recycle connections periodically

	// Run Migrations
	if err := runMigrations(db, migrations.FS); err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(db *sql.DB, fs embed.FS) error {
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

	// Fetch all applied versions in a single query
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return err
	}
	applied := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return err
		}
		applied[v] = true
	}
	rows.Close()

	// Run all migrations in order
	migrationFiles := []struct {
		version int
		file    string
	}{
		{1, "001_create_users.sql"},
		{2, "002_create_households.sql"},
		{3, "003_create_recipes.sql"},
		{4, "004_seed_recipes.sql"},
		{5, "005_create_menus.sql"},
		{6, "006_household_invites_and_member_status.sql"},
		{7, "007_add_menu_date_index.sql"},
		{8, "008_seed_more_recipes.sql"},
		{9, "009_add_token_version.sql"},
		{10, "010_add_recipe_household_id.sql"},
	}

	for _, m := range migrationFiles {
		// Check if migration already applied
		if applied[m.version] {
			// Migration already applied, skip
			continue
		}

		// Read and execute migration file from embedded FS
		sqlBytes, err := fs.ReadFile(m.file)
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
