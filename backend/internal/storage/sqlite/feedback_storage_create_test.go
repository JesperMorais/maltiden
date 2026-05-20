package sqlite

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"maltiden/internal/domain"
)

func setupFeedbackTestDB(t *testing.T) (*sql.DB, string, string) {
	t.Helper()

	tmpFile := t.TempDir() + "/test.db"
	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	householdID := "hh_feedback_test"
	if _, err := db.Exec(
		`INSERT INTO households (id, name) VALUES (?, ?)`,
		householdID, "Feedback Test Household",
	); err != nil {
		t.Fatalf("seed household: %v", err)
	}

	userID := "usr_feedback_test"
	if _, err := db.Exec(
		`INSERT INTO users (id, email, password_hash, name, household_id) VALUES (?, ?, ?, ?, ?)`,
		userID, "feedback@test.com", "hash", "Test User", householdID,
	); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	return db, householdID, userID
}

func TestFeedbackStorage_Create(t *testing.T) {
	db, householdID, userID := setupFeedbackTestDB(t)
	storage := NewFeedbackStorage(db)

	createdAt := time.Now().UTC().Truncate(time.Second)
	fb := domain.Feedback{
		ID:            "fb_test_001",
		UserID:        userID,
		HouseholdID:   householdID,
		Mood:          "good",
		Categories:    []string{"bug", "ui"},
		Comment:       "Great app!",
		Page:          "/dashboard",
		ViewportWidth: 1440,
		UserAgent:     "Mozilla/5.0 (Test)",
		CreatedAt:     createdAt,
	}

	if err := storage.Create(&fb); err != nil {
		t.Fatalf("Create: %v", err)
	}

	var (
		gotID            string
		gotUserID        string
		gotHouseholdID   sql.NullString
		gotMood          string
		gotCategories    sql.NullString
		gotComment       sql.NullString
		gotPage          sql.NullString
		gotViewportWidth sql.NullInt64
		gotUserAgent     sql.NullString
		gotCreatedAt     time.Time
	)

	err := db.QueryRow(
		`SELECT id, user_id, household_id, mood, categories, comment, page, viewport_width, user_agent, created_at
		 FROM feedback WHERE id = ?`,
		fb.ID,
	).Scan(
		&gotID,
		&gotUserID,
		&gotHouseholdID,
		&gotMood,
		&gotCategories,
		&gotComment,
		&gotPage,
		&gotViewportWidth,
		&gotUserAgent,
		&gotCreatedAt,
	)
	if err != nil {
		t.Fatalf("SELECT: %v", err)
	}

	if gotID != fb.ID {
		t.Errorf("id: got %q, want %q", gotID, fb.ID)
	}
	if gotUserID != fb.UserID {
		t.Errorf("user_id: got %q, want %q", gotUserID, fb.UserID)
	}
	if !gotHouseholdID.Valid || gotHouseholdID.String != fb.HouseholdID {
		t.Errorf("household_id: got %v, want %q", gotHouseholdID, fb.HouseholdID)
	}
	if gotMood != fb.Mood {
		t.Errorf("mood: got %q, want %q", gotMood, fb.Mood)
	}

	if !gotCategories.Valid {
		t.Fatal("categories: expected non-NULL")
	}
	var gotCats []string
	if err := json.Unmarshal([]byte(gotCategories.String), &gotCats); err != nil {
		t.Fatalf("categories unmarshal: %v", err)
	}
	if len(gotCats) != len(fb.Categories) {
		t.Errorf("categories length: got %d, want %d", len(gotCats), len(fb.Categories))
	} else {
		for i := range fb.Categories {
			if gotCats[i] != fb.Categories[i] {
				t.Errorf("categories[%d]: got %q, want %q", i, gotCats[i], fb.Categories[i])
			}
		}
	}

	if !gotComment.Valid || gotComment.String != fb.Comment {
		t.Errorf("comment: got %v, want %q", gotComment, fb.Comment)
	}
	if !gotPage.Valid || gotPage.String != fb.Page {
		t.Errorf("page: got %v, want %q", gotPage, fb.Page)
	}
	if !gotViewportWidth.Valid || int(gotViewportWidth.Int64) != fb.ViewportWidth {
		t.Errorf("viewport_width: got %v, want %d", gotViewportWidth, fb.ViewportWidth)
	}
	if !gotUserAgent.Valid || gotUserAgent.String != fb.UserAgent {
		t.Errorf("user_agent: got %v, want %q", gotUserAgent, fb.UserAgent)
	}
	if !gotCreatedAt.Equal(createdAt) {
		t.Errorf("created_at: got %v, want %v", gotCreatedAt, createdAt)
	}
}
