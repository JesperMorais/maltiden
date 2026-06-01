package sqlite

import (
	"testing"
	"time"

	"maltiden/internal/domain"
)

func setupUserTestDB(t *testing.T) *UserStorage {
	t.Helper()
	tmpFile := t.TempDir() + "/test.db"
	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Exec(`INSERT INTO households (id, name) VALUES (?, ?)`, "hh_usr_test", "User Test Household"); err != nil {
		t.Fatalf("seed household: %v", err)
	}
	return NewUserStorage(db)
}

func newTestUser(id, email string) *domain.User {
	return &domain.User{
		ID:           id,
		Email:        email,
		PasswordHash: "$2a$12$testhash",
		Name:         "Test User",
		HouseholdID:  "hh_usr_test",
		CreatedAt:    time.Now().UTC().Truncate(time.Second),
	}
}

func TestUserStorage_Create(t *testing.T) {
	s := setupUserTestDB(t)
	u := newTestUser("usr_create_1", "create@example.com")

	if err := s.Create(u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.GetByID(u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got == nil {
		t.Fatal("GetByID: expected user, got nil")
	}
	if got.ID != u.ID || got.Email != u.Email || got.Name != u.Name {
		t.Errorf("field mismatch: got %+v, want %+v", got, u)
	}
}

func TestUserStorage_GetByID_Missing(t *testing.T) {
	s := setupUserTestDB(t)

	got, err := s.GetByID("usr_does_not_exist")
	if err != nil {
		t.Fatalf("GetByID missing: unexpected error %v", err)
	}
	if got != nil {
		t.Fatalf("GetByID missing: expected nil, got %+v", got)
	}
}

func TestUserStorage_GetByEmail(t *testing.T) {
	s := setupUserTestDB(t)
	u := newTestUser("usr_email_1", "email@example.com")

	if err := s.Create(u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.GetByEmail(u.Email)
	if err != nil {
		t.Fatalf("GetByEmail hit: %v", err)
	}
	if got == nil {
		t.Fatal("GetByEmail hit: expected user, got nil")
	}
	if got.Email != u.Email {
		t.Errorf("email mismatch: got %q, want %q", got.Email, u.Email)
	}
}

func TestUserStorage_GetByEmail_Missing(t *testing.T) {
	s := setupUserTestDB(t)

	got, err := s.GetByEmail("nobody@example.com")
	if err != nil {
		t.Fatalf("GetByEmail missing: unexpected error %v", err)
	}
	if got != nil {
		t.Fatalf("GetByEmail missing: expected nil, got %+v", got)
	}
}

func TestUserStorage_UpdatePassword(t *testing.T) {
	s := setupUserTestDB(t)
	u := newTestUser("usr_pw_1", "pw@example.com")

	if err := s.Create(u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	newHash := "$2a$12$newHash"
	if err := s.UpdatePassword(u.ID, newHash); err != nil {
		t.Fatalf("UpdatePassword: %v", err)
	}

	got, err := s.GetByID(u.ID)
	if err != nil {
		t.Fatalf("GetByID after UpdatePassword: %v", err)
	}
	if got == nil {
		t.Fatal("GetByID after UpdatePassword: expected user, got nil")
	}
	if got.PasswordHash != newHash {
		t.Errorf("password hash not updated: got %q, want %q", got.PasswordHash, newHash)
	}
}
