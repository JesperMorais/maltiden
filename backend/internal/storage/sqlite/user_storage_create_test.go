package sqlite

import (
	"maltiden/internal/domain"
	"testing"
	"time"
)

func TestUserStorage_Create_RoundTrip(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	storage := NewUserStorage(db)

	want := &domain.User{
		ID:           "usr_test001",
		Email:        "test@example.com",
		PasswordHash: "$2a$12$hashedpassword",
		Name:         "Test User",
		HouseholdID:  "hh_test001",
		CreatedAt:    time.Now().UTC().Truncate(time.Second),
	}

	if err := storage.Create(want); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := storage.GetByID(want.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got == nil {
		t.Fatal("GetByID returned nil for existing user")
	}

	if got.ID != want.ID {
		t.Errorf("ID: got %q, want %q", got.ID, want.ID)
	}
	if got.Email != want.Email {
		t.Errorf("Email: got %q, want %q", got.Email, want.Email)
	}
	if got.PasswordHash != want.PasswordHash {
		t.Errorf("PasswordHash: got %q, want %q", got.PasswordHash, want.PasswordHash)
	}
	if got.Name != want.Name {
		t.Errorf("Name: got %q, want %q", got.Name, want.Name)
	}
	if got.HouseholdID != want.HouseholdID {
		t.Errorf("HouseholdID: got %q, want %q", got.HouseholdID, want.HouseholdID)
	}
}

func TestUserStorage_Create_GetByID_Missing(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	storage := NewUserStorage(db)

	got, err := storage.GetByID("usr_does_not_exist")
	if err != nil {
		t.Fatalf("GetByID for missing id returned error: %v", err)
	}
	if got != nil {
		t.Errorf("GetByID for missing id returned non-nil user: %+v", got)
	}
}
