package sqlite

import (
	"maltiden/internal/domain"
	"testing"
	"time"
)

func setupFeedbackTestDB(t *testing.T) (*FeedbackStorage, string) {
	t.Helper()

	tmpFile := t.TempDir() + "/test.db"
	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	userID := "usr_feedback_test"
	if _, err := db.Exec(
		`INSERT INTO users (id, email, password_hash, name, household_id) VALUES (?, ?, ?, ?, ?)`,
		userID, "test@example.com", "hash", "Test User", "hh_test",
	); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	return NewFeedbackStorage(db), userID
}

func TestFeedbackStorage_Create(t *testing.T) {
	storage, userID := setupFeedbackTestDB(t)

	fb := &domain.Feedback{
		ID:        "fb_001",
		UserID:    userID,
		Mood:      "good",
		CreatedAt: time.Now().UTC(),
	}

	if err := storage.Create(fb); err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestFeedbackStorage_Create_WithAllFields(t *testing.T) {
	storage, userID := setupFeedbackTestDB(t)

	fb := &domain.Feedback{
		ID:            "fb_002",
		UserID:        userID,
		HouseholdID:   "hh_someplace",
		Mood:          "okay",
		Categories:    []string{"navigation", "performance"},
		Comment:       "Feels a bit slow",
		Page:          "/menu",
		ViewportWidth: 1440,
		UserAgent:     "Mozilla/5.0",
		CreatedAt:     time.Now().UTC(),
	}

	if err := storage.Create(fb); err != nil {
		t.Fatalf("Create with all fields: %v", err)
	}
}

func TestFeedbackStorage_CountRecentByUser_Empty(t *testing.T) {
	storage, userID := setupFeedbackTestDB(t)

	count, err := storage.CountRecentByUser(userID, time.Now().Add(-1*time.Hour).UTC())
	if err != nil {
		t.Fatalf("CountRecentByUser: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}
}

func TestFeedbackStorage_CountRecentByUser_AfterCreate(t *testing.T) {
	storage, userID := setupFeedbackTestDB(t)

	for i, id := range []string{"fb_a", "fb_b"} {
		fb := &domain.Feedback{
			ID:        id,
			UserID:    userID,
			Mood:      []string{"good", "bad"}[i],
			CreatedAt: time.Now().UTC(),
		}
		if err := storage.Create(fb); err != nil {
			t.Fatalf("Create %s: %v", id, err)
		}
	}

	count, err := storage.CountRecentByUser(userID, time.Now().Add(-1*time.Hour).UTC())
	if err != nil {
		t.Fatalf("CountRecentByUser: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}

func TestFeedbackStorage_CountRecentByUser_ExcludesOldRows(t *testing.T) {
	storage, userID := setupFeedbackTestDB(t)

	fb := &domain.Feedback{
		ID:        "fb_old",
		UserID:    userID,
		Mood:      "good",
		CreatedAt: time.Now().Add(-48 * time.Hour).UTC(),
	}
	if err := storage.Create(fb); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Query for the last hour — the row above is 48h old and must not appear.
	count, err := storage.CountRecentByUser(userID, time.Now().Add(-1*time.Hour).UTC())
	if err != nil {
		t.Fatalf("CountRecentByUser: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 (old row excluded), got %d", count)
	}
}
