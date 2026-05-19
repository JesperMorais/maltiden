package sqlite

import (
	"maltiden/internal/domain"
	"testing"
	"time"
)

func setupUserTestDB(t *testing.T) *UserStorage {
	t.Helper()
	db, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewUserStorage(db)
}

func newTestUser(id, email string) *domain.User {
	return &domain.User{
		ID:           id,
		Email:        email,
		PasswordHash: "$2a$12$testhash",
		Name:         "Test User",
		HouseholdID:  "hh_test",
		CreatedAt:    time.Now().UTC().Truncate(time.Second),
	}
}

func TestUserStorage_Create_HappyPath(t *testing.T) {
	s := setupUserTestDB(t)
	u := newTestUser("usr_1", "alice@example.com")

	if err := s.Create(u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.GetByID(u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got == nil {
		t.Fatal("GetByID returned nil after Create")
	}
	if got.Email != u.Email {
		t.Errorf("email mismatch: got %q want %q", got.Email, u.Email)
	}
	if got.Name != u.Name {
		t.Errorf("name mismatch: got %q want %q", got.Name, u.Name)
	}
}

func TestUserStorage_Create_DuplicateEmail(t *testing.T) {
	s := setupUserTestDB(t)

	if err := s.Create(newTestUser("usr_a", "dup@example.com")); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	err := s.Create(newTestUser("usr_b", "dup@example.com"))
	if err == nil {
		t.Fatal("expected error on duplicate email, got nil")
	}
}

func TestUserStorage_GetByID_Missing(t *testing.T) {
	s := setupUserTestDB(t)

	got, err := s.GetByID("usr_nonexistent")
	if err != nil {
		t.Fatalf("GetByID missing: unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil for missing user, got %+v", got)
	}
}

func TestUserStorage_GetByEmail_HappyPath(t *testing.T) {
	s := setupUserTestDB(t)
	u := newTestUser("usr_2", "bob@example.com")

	if err := s.Create(u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.GetByEmail(u.Email)
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if got == nil {
		t.Fatal("GetByEmail returned nil after Create")
	}
	if got.ID != u.ID {
		t.Errorf("id mismatch: got %q want %q", got.ID, u.ID)
	}
}

func TestUserStorage_GetByEmail_Missing(t *testing.T) {
	s := setupUserTestDB(t)

	got, err := s.GetByEmail("nobody@example.com")
	if err != nil {
		t.Fatalf("GetByEmail missing: unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil for missing email, got %+v", got)
	}
}
