package services

import (
	"context"
	"database/sql"
	"maltiden/internal/domain"
	"maltiden/internal/storage/sqlite"
	"testing"
)

func countSessionEvents(t *testing.T, db *sql.DB) int {
	t.Helper()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM session_events").Scan(&count); err != nil {
		t.Fatalf("failed to count session_events: %v", err)
	}
	return count
}

func TestEventService_FlagOff_NoOp(t *testing.T) {
	db := setupTestDB(t)
	storage := sqlite.NewEventStorage(db)
	svc := NewEventService(storage, "req_1")

	if err := svc.Track(context.Background(), domain.EventMenuGenerated, "hh_123", ""); err != nil {
		t.Fatalf("expected no error when flag is off, got: %v", err)
	}

	if count := countSessionEvents(t, db); count != 0 {
		t.Fatalf("expected no rows written when flag is off, got %d", count)
	}
}

func TestEventService_RejectsNonAllowlistedType(t *testing.T) {
	t.Setenv("EVENTS_ENABLED", "true")
	db := setupTestDB(t)
	storage := sqlite.NewEventStorage(db)
	svc := NewEventService(storage, "req_1")

	if err := svc.Track(context.Background(), "not_a_real_event", "hh_123", ""); err != ErrEventTypeNotAllowed {
		t.Fatalf("expected ErrEventTypeNotAllowed, got: %v", err)
	}

	if count := countSessionEvents(t, db); count != 0 {
		t.Fatalf("expected no rows written for rejected event type, got %d", count)
	}
}

func TestEventService_StoresHashOnly(t *testing.T) {
	t.Setenv("EVENTS_ENABLED", "true")
	db := setupTestDB(t)
	storage := sqlite.NewEventStorage(db)
	svc := NewEventService(storage, "req_1")

	householdID := "hh_super_secret_id"
	if err := svc.Track(context.Background(), domain.EventMenuGenerated, householdID, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var stored string
	if err := db.QueryRow("SELECT household_id_hash FROM session_events LIMIT 1").Scan(&stored); err != nil {
		t.Fatalf("failed to read stored event: %v", err)
	}

	if stored == householdID {
		t.Fatalf("raw household_id was stored, expected a hash")
	}
	if len(stored) != 64 {
		t.Fatalf("expected a 64-char hex sha256 hash, got %d chars", len(stored))
	}
}
