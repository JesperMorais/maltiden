package sqlite

import (
	"testing"
	"time"

	"maltiden/internal/domain"
)

func setupUserTestDB(t *testing.T) (*UserStorage, string) {
	t.Helper()

	tmpFile := t.TempDir() + "/test.db"
	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	householdID := "hh_user_email_test"
	if _, err := db.Exec(
		`INSERT INTO households (id, name) VALUES (?, ?)`,
		householdID, "User Email Test Household",
	); err != nil {
		t.Fatalf("seed household: %v", err)
	}

	return NewUserStorage(db), householdID
}

func TestUserStorage_GetByEmail_HappyPath(t *testing.T) {
	t.Helper()
	storage, householdID := setupUserTestDB(t)

	user := &domain.User{
		ID:           "usr_email_test_1",
		Email:        "test@example.com",
		PasswordHash: "hash_original",
		Name:         "Test User",
		HouseholdID:  householdID,
		CreatedAt:    time.Now().UTC().Truncate(time.Second),
	}
	if err := storage.Create(user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := storage.GetByEmail(user.Email)
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if got == nil {
		t.Fatalf("GetByEmail returned nil, want user")
	}
	if got.ID != user.ID {
		t.Fatalf("ID: got %q, want %q", got.ID, user.ID)
	}
	if got.Email != user.Email {
		t.Fatalf("Email: got %q, want %q", got.Email, user.Email)
	}
	if got.PasswordHash != user.PasswordHash {
		t.Fatalf("PasswordHash: got %q, want %q", got.PasswordHash, user.PasswordHash)
	}
}

func TestUserStorage_GetByEmail_Missing(t *testing.T) {
	t.Helper()
	storage, _ := setupUserTestDB(t)

	got, err := storage.GetByEmail("nobody@example.com")
	if err != nil {
		t.Fatalf("GetByEmail missing: unexpected error %v", err)
	}
	if got != nil {
		t.Fatalf("GetByEmail missing: expected nil, got %+v", got)
	}
}

func TestUserStorage_UpdatePassword(t *testing.T) {
	t.Helper()
	storage, householdID := setupUserTestDB(t)

	user := &domain.User{
		ID:           "usr_pw_test_1",
		Email:        "pwtest@example.com",
		PasswordHash: "hash_old",
		Name:         "PW Test User",
		HouseholdID:  householdID,
		CreatedAt:    time.Now().UTC().Truncate(time.Second),
	}
	if err := storage.Create(user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	newHash := "hash_new"
	if err := storage.UpdatePassword(user.ID, newHash); err != nil {
		t.Fatalf("UpdatePassword: %v", err)
	}

	got, err := storage.GetByEmail(user.Email)
	if err != nil {
		t.Fatalf("GetByEmail after UpdatePassword: %v", err)
	}
	if got == nil {
		t.Fatalf("GetByEmail after UpdatePassword returned nil")
	}
	if got.PasswordHash != newHash {
		t.Fatalf("PasswordHash: got %q, want %q", got.PasswordHash, newHash)
	}
}
