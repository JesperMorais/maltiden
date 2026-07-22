package sqlite

import (
	"testing"
	"time"

	"maltiden/internal/domain"
)

func setupFeedbackTestDB(t *testing.T) (*FeedbackStorage, string) {
	t.Helper()
	tmpFile := t.TempDir() + "/test.db"
	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	userID := "usr_fb_test"
	if _, err := db.Exec(`INSERT INTO households (id, name) VALUES (?, ?)`, "hh_fb_test", "Feedback Test Household"); err != nil {
		t.Fatalf("seed household: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO users (id, email, password_hash, name, household_id, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		userID, "fb@example.com", "$2a$12$testhash", "FB User", "hh_fb_test", time.Now().UTC(),
	); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	return NewFeedbackStorage(db), userID
}

func newTestFeedback(id, userID string, createdAt time.Time) *domain.Feedback {
	return &domain.Feedback{
		ID:        id,
		UserID:    userID,
		Mood:      "good",
		CreatedAt: createdAt,
	}
}

func TestFeedbackStorage_Create(t *testing.T) {
	s, userID := setupFeedbackTestDB(t)
	fb := newTestFeedback("fb_create_1", userID, time.Now().UTC())

	if err := s.Create(fb); err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestFeedbackStorage_CountRecentByUser(t *testing.T) {
	s, userID := setupFeedbackTestDB(t)

	now := time.Now().UTC()
	window := 1 * time.Hour
	since := now.Add(-window)

	// Insert 3 rows inside the window.
	for i := range 3 {
		ts := now.Add(-time.Duration(i+1) * 10 * time.Minute) // 10m, 20m, 30m ago
		fb := newTestFeedback("fb_recent_"+string(rune('a'+i)), userID, ts)
		if err := s.Create(fb); err != nil {
			t.Fatalf("Create recent[%d]: %v", i, err)
		}
	}

	// Insert 1 row outside the window.
	old := newTestFeedback("fb_old_1", userID, now.Add(-2*time.Hour))
	if err := s.Create(old); err != nil {
		t.Fatalf("Create old: %v", err)
	}

	count, err := s.CountRecentByUser(userID, since)
	if err != nil {
		t.Fatalf("CountRecentByUser: %v", err)
	}
	if count != 3 {
		t.Errorf("CountRecentByUser: got %d, want 3", count)
	}
}

func TestFeedbackStorage_CountRecentByUser_Empty(t *testing.T) {
	s, userID := setupFeedbackTestDB(t)

	count, err := s.CountRecentByUser(userID, time.Now().UTC())
	if err != nil {
		t.Fatalf("CountRecentByUser empty: %v", err)
	}
	if count != 0 {
		t.Errorf("CountRecentByUser empty: got %d, want 0", count)
	}
}

func TestFeedbackStorage_CountRecentByUser_OtherUser(t *testing.T) {
	s, userID := setupFeedbackTestDB(t)

	fb := newTestFeedback("fb_other_1", userID, time.Now().UTC())
	if err := s.Create(fb); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Query for a different user — should return 0.
	count, err := s.CountRecentByUser("usr_somebody_else", time.Now().UTC().Add(-time.Hour))
	if err != nil {
		t.Fatalf("CountRecentByUser other user: %v", err)
	}
	if count != 0 {
		t.Errorf("CountRecentByUser other user: got %d, want 0", count)
	}
}
