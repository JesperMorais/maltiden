package sqlite

import (
	"testing"
	"time"

	"maltiden/internal/domain"
)

func newTestUser() *domain.User {
	return &domain.User{
		ID:           "usr_email_test_001",
		Email:        "email@example.com",
		PasswordHash: "$2a$12$originalhashedpassword",
		Name:         "Email Test User",
		HouseholdID:  "hh_email_test",
		CreatedAt:    time.Now().UTC().Truncate(time.Second),
	}
}

func TestUserStorage_GetByEmail_Hit(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	storage := NewUserStorage(db)
	user := newTestUser()

	if err := storage.Create(user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := storage.GetByEmail(user.Email)
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if got == nil {
		t.Fatal("GetByEmail returned nil, expected user")
	}
	if got.ID != user.ID {
		t.Errorf("ID: got %q, want %q", got.ID, user.ID)
	}
	if got.Email != user.Email {
		t.Errorf("Email: got %q, want %q", got.Email, user.Email)
	}
	if got.PasswordHash != user.PasswordHash {
		t.Errorf("PasswordHash: got %q, want %q", got.PasswordHash, user.PasswordHash)
	}
}

func TestUserStorage_GetByEmail_Miss(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	storage := NewUserStorage(db)

	got, err := storage.GetByEmail("nobody@example.com")
	if err != nil {
		t.Fatalf("GetByEmail missing: expected nil error, got %v", err)
	}
	if got != nil {
		t.Fatalf("GetByEmail missing: expected nil, got %+v", got)
	}
}

func TestUserStorage_UpdatePassword_Persists(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	storage := NewUserStorage(db)
	user := newTestUser()

	if err := storage.Create(user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	newHash := "$2a$12$newhashedpassword"
	if err := storage.UpdatePassword(user.ID, newHash); err != nil {
		t.Fatalf("UpdatePassword: %v", err)
	}

	got, err := storage.GetByID(user.ID)
	if err != nil {
		t.Fatalf("GetByID after UpdatePassword: %v", err)
	}
	if got == nil {
		t.Fatal("GetByID returned nil after UpdatePassword")
	}
	if got.PasswordHash != newHash {
		t.Errorf("PasswordHash after update: got %q, want %q", got.PasswordHash, newHash)
	}
}
