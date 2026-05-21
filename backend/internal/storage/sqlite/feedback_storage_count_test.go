package sqlite

import (
	"testing"
	"time"
)

func TestFeedbackStorage_CountRecentByUser(t *testing.T) {
	tmpFile := t.TempDir() + "/feedback_test.db"
	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	userID := "usr_feedback_count_test"
	if _, err := db.Exec(
		`INSERT INTO users (id, email, password_hash, name, household_id) VALUES (?, ?, ?, ?, ?)`,
		userID, "feedback_count@test.example", "hash", "Test User", "hh_feedback_count_test",
	); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	now := time.Now().Truncate(time.Second)
	rows := []struct {
		id  string
		at  time.Time
	}{
		{"fb_count_1", now.Add(-2 * time.Hour)},
		{"fb_count_2", now.Add(-30 * time.Minute)},
		{"fb_count_3", now.Add(-5 * time.Minute)},
	}
	for _, r := range rows {
		if _, err := db.Exec(
			`INSERT INTO feedback (id, user_id, mood, created_at) VALUES (?, ?, ?, ?)`,
			r.id, userID, "good", r.at,
		); err != nil {
			t.Fatalf("seed feedback %s: %v", r.id, err)
		}
	}

	storage := NewFeedbackStorage(db)

	tests := []struct {
		name  string
		since time.Time
		want  int
	}{
		{"last 1h returns 2", now.Add(-1 * time.Hour), 2},
		{"last 3h returns all 3", now.Add(-3 * time.Hour), 3},
		{"last 1m returns 0", now.Add(-1 * time.Minute), 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			got, err := storage.CountRecentByUser(userID, tc.since)
			if err != nil {
				t.Fatalf("CountRecentByUser: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}
