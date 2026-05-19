package sqlite

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"maltiden/internal/domain"
)

func TestFeedbackStorage_Create(t *testing.T) {
	t.Run("all fields populated", func(t *testing.T) {
		db := openFeedbackTestDB(t)
		seedFeedbackUser(t, db, "usr_fb_test")

		createdAt := time.Now().UTC().Truncate(time.Second)
		f := domain.Feedback{
			ID:            "fbk_test",
			UserID:        "usr_fb_test",
			HouseholdID:   "hh_fb_test",
			Mood:          "good",
			Categories:    []string{"ui", "speed"},
			Comment:       "great app",
			Page:          "/recipes",
			ViewportWidth: 1440,
			UserAgent:     "Mozilla/5.0",
			CreatedAt:     createdAt,
		}

		if err := NewFeedbackStorage(db).Create(&f); err != nil {
			t.Fatalf("Create: %v", err)
		}

		var (
			id            string
			userID        string
			householdID   sql.NullString
			mood          string
			categoriesRaw sql.NullString
			comment       sql.NullString
			page          sql.NullString
			viewportWidth sql.NullInt64
			userAgent     sql.NullString
			storedAt      time.Time
		)

		row := db.QueryRow(
			`SELECT id, user_id, household_id, mood, categories, comment, page, viewport_width, user_agent, created_at
			 FROM feedback WHERE id = ?`, f.ID,
		)
		if err := row.Scan(&id, &userID, &householdID, &mood, &categoriesRaw, &comment, &page, &viewportWidth, &userAgent, &storedAt); err != nil {
			t.Fatalf("Scan: %v", err)
		}

		if id != f.ID {
			t.Errorf("id: got %q, want %q", id, f.ID)
		}
		if userID != f.UserID {
			t.Errorf("user_id: got %q, want %q", userID, f.UserID)
		}
		if !householdID.Valid || householdID.String != f.HouseholdID {
			t.Errorf("household_id: got %v, want %q", householdID, f.HouseholdID)
		}
		if mood != f.Mood {
			t.Errorf("mood: got %q, want %q", mood, f.Mood)
		}
		if !categoriesRaw.Valid {
			t.Fatal("categories: expected non-NULL")
		}
		var gotCategories []string
		if err := json.Unmarshal([]byte(categoriesRaw.String), &gotCategories); err != nil {
			t.Fatalf("unmarshal categories: %v", err)
		}
		if !reflect.DeepEqual(gotCategories, f.Categories) {
			t.Errorf("categories: got %v, want %v", gotCategories, f.Categories)
		}
		if !comment.Valid || comment.String != f.Comment {
			t.Errorf("comment: got %v, want %q", comment, f.Comment)
		}
		if !page.Valid || page.String != f.Page {
			t.Errorf("page: got %v, want %q", page, f.Page)
		}
		if !viewportWidth.Valid || int(viewportWidth.Int64) != f.ViewportWidth {
			t.Errorf("viewport_width: got %v, want %d", viewportWidth, f.ViewportWidth)
		}
		if !userAgent.Valid || userAgent.String != f.UserAgent {
			t.Errorf("user_agent: got %v, want %q", userAgent, f.UserAgent)
		}
		if !storedAt.UTC().Truncate(time.Second).Equal(createdAt) {
			t.Errorf("created_at: got %v, want %v", storedAt, createdAt)
		}
	})

	t.Run("optional fields empty become NULL", func(t *testing.T) {
		db := openFeedbackTestDB(t)
		seedFeedbackUser(t, db, "usr_fb_null")

		f := domain.Feedback{
			ID:        "fbk_null",
			UserID:    "usr_fb_null",
			Mood:      "bad",
			CreatedAt: time.Now().UTC().Truncate(time.Second),
		}

		if err := NewFeedbackStorage(db).Create(&f); err != nil {
			t.Fatalf("Create: %v", err)
		}

		var (
			householdID   sql.NullString
			categoriesRaw sql.NullString
			comment       sql.NullString
			page          sql.NullString
			viewportWidth sql.NullInt64
			userAgent     sql.NullString
		)

		row := db.QueryRow(
			`SELECT household_id, categories, comment, page, viewport_width, user_agent
			 FROM feedback WHERE id = ?`, f.ID,
		)
		if err := row.Scan(&householdID, &categoriesRaw, &comment, &page, &viewportWidth, &userAgent); err != nil {
			t.Fatalf("Scan: %v", err)
		}

		if householdID.Valid {
			t.Errorf("household_id: expected NULL, got %q", householdID.String)
		}
		if categoriesRaw.Valid {
			t.Errorf("categories: expected NULL, got %q", categoriesRaw.String)
		}
		if comment.Valid {
			t.Errorf("comment: expected NULL, got %q", comment.String)
		}
		if page.Valid {
			t.Errorf("page: expected NULL, got %q", page.String)
		}
		if viewportWidth.Valid {
			t.Errorf("viewport_width: expected NULL, got %d", viewportWidth.Int64)
		}
		if userAgent.Valid {
			t.Errorf("user_agent: expected NULL, got %q", userAgent.String)
		}
	})
}

func openFeedbackTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func seedFeedbackUser(t *testing.T, db *sql.DB, id string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO users (id, email, password_hash, name, household_id) VALUES (?, ?, ?, ?, ?)`,
		id, id+"@test.com", "hash", "Test User", "hh_none",
	)
	if err != nil {
		t.Fatalf("seed user %q: %v", id, err)
	}
}
