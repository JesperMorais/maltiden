package sqlite

import (
	"testing"
	"time"

	"maltiden/internal/domain"
)

func TestUserStorage_Create_GetByID_RoundTrip(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	storage := NewUserStorage(db)

	now := time.Now().UTC().Truncate(time.Second)
	user := &domain.User{
		ID:           "usr_test_create_001",
		Email:        "test@example.com",
		PasswordHash: "$2a$12$hashedpassword",
		Name:         "Test User",
		HouseholdID:  "hh_test_household",
		CreatedAt:    now,
	}

	if err := storage.Create(user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := storage.GetByID(user.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got == nil {
		t.Fatal("GetByID returned nil, expected user")
	}

	if got.ID != user.ID {
		t.Errorf("ID: got %q, want %q", got.ID, user.ID)
	}
	if got.Email != user.Email {
		t.Errorf("Email: got %q, want %q", got.Email, user.Email)
	}
	if got.Name != user.Name {
		t.Errorf("Name: got %q, want %q", got.Name, user.Name)
	}
	if got.HouseholdID != user.HouseholdID {
		t.Errorf("HouseholdID: got %q, want %q", got.HouseholdID, user.HouseholdID)
	}
	if got.PasswordHash != user.PasswordHash {
		t.Errorf("PasswordHash: got %q, want %q", got.PasswordHash, user.PasswordHash)
	}
	if !got.CreatedAt.Equal(user.CreatedAt) {
		t.Errorf("CreatedAt: got %v, want %v", got.CreatedAt, user.CreatedAt)
	}
}

func TestUserStorage_GetByID_NotFound(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	storage := NewUserStorage(db)

	got, err := storage.GetByID("usr_does_not_exist")
	if err != nil {
		t.Fatalf("GetByID missing id: expected nil error, got %v", err)
	}
	if got != nil {
		t.Fatalf("GetByID missing id: expected nil user, got %+v", got)
	}
}
