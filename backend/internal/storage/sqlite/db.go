package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
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

	// Configure connection pool (conservative for SQLite)
	db.SetMaxOpenConns(25)    // single writer, multiple readers with WAL
	db.SetMaxIdleConns(5)     // keep a few warm connections
	db.SetConnMaxLifetime(5 * time.Minute) // recycle connections periodically

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
		{7, "migrations/007_add_menu_date_index.sql"},
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
