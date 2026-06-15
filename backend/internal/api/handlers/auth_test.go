package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/pkg/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockUserStorage stubs domain.UserRepository for auth handler tests.
type mockUserStorage struct {
	user     *domain.User
	getErr   error
}

func (m *mockUserStorage) GetByEmail(_ string) (*domain.User, error) {
	return m.user, m.getErr
}
func (m *mockUserStorage) GetByID(_ string) (*domain.User, error)                      { return nil, nil }
func (m *mockUserStorage) Create(_ *domain.User) error                                  { return nil }
func (m *mockUserStorage) CreateTx(_ *sql.Tx, _ *domain.User) error                    { return nil }
func (m *mockUserStorage) GetTokenVersion(_ string) (int, error)                        { return 0, nil }
func (m *mockUserStorage) IncrementTokenVersion(_ string) error                         { return nil }
func (m *mockUserStorage) IncrementTokenVersionTx(_ *sql.Tx, _ string) error            { return nil }
func (m *mockUserStorage) UpdatePassword(_, _ string) error                             { return nil }
func (m *mockUserStorage) UpdatePasswordTx(_ *sql.Tx, _, _ string) error               { return nil }
func (m *mockUserStorage) CreatePasswordResetToken(_ *domain.PasswordResetToken) error  { return nil }
func (m *mockUserStorage) GetPasswordResetToken(_ string) (*domain.PasswordResetToken, error) {
	return nil, nil
}
func (m *mockUserStorage) MarkPasswordResetTokenUsed(_ string) error                    { return nil }
func (m *mockUserStorage) MarkPasswordResetTokenUsedTx(_ *sql.Tx, _ string) error      { return nil }

// mockHouseholdStorage stubs domain.HouseholdRepository for auth handler tests.
type mockHouseholdStorage struct{}

func (m *mockHouseholdStorage) Create(_ *domain.Household) error                               { return nil }
func (m *mockHouseholdStorage) CreateTx(_ *sql.Tx, _ *domain.Household) error                 { return nil }
func (m *mockHouseholdStorage) UpdateName(_, _ string) error                                   { return nil }
func (m *mockHouseholdStorage) GetByUserID(_ string) (*domain.HouseholdResponse, error)        { return nil, nil }
func (m *mockHouseholdStorage) CreateInviteCode(_ *domain.InviteCode) error                    { return nil }
func (m *mockHouseholdStorage) GetInviteByCode(_ string) (*domain.InviteCode, error)           { return nil, nil }
func (m *mockHouseholdStorage) MarkInviteUsedTx(_ *sql.Tx, _, _ string) error                 { return nil }
func (m *mockHouseholdStorage) GetMemberRole(_, _ string) (string, error)                      { return "", nil }
func (m *mockHouseholdStorage) GetMemberStatuses(_ string) ([]domain.MemberStatus, error)      { return nil, nil }
func (m *mockHouseholdStorage) UpdateMemberStatus(_ string, _ string, _ *bool, _ *bool) error { return nil }
func (m *mockHouseholdStorage) RemoveMember(_, _ string) error                                 { return nil }
func (m *mockHouseholdStorage) IsMember(_, _ string) (bool, error)                             { return false, nil }
func (m *mockHouseholdStorage) IsMemberTx(_ *sql.Tx, _, _ string) (bool, error)               { return false, nil }
func (m *mockHouseholdStorage) GetUserHouseholdID(_ string) (string, error)                    { return "", nil }
func (m *mockHouseholdStorage) RemoveMemberTx(_ *sql.Tx, _, _ string) error                   { return nil }
func (m *mockHouseholdStorage) AddMemberTx(_ *sql.Tx, _ *domain.HouseholdMember) error        { return nil }
func (m *mockHouseholdStorage) UpdateUserHouseholdTx(_ *sql.Tx, _, _ string) error            { return nil }
func (m *mockHouseholdStorage) DB() *sql.DB                                                    { return nil }

func newTestAuthHandler(userStore *mockUserStorage) *AuthHandler {
	jwtSvc, _ := utils.NewJWTService("test-secret")
	svc := services.NewAuthService(nil, userStore, &mockHouseholdStorage{}, jwtSvc)
	return NewAuthHandler(svc)
}

// Register tests — pre-DB error paths only

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	h := newTestAuthHandler(&mockUserStorage{})

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAuthHandler_Register_InvalidEmail(t *testing.T) {
	h := newTestAuthHandler(&mockUserStorage{})

	body := `{"email":"not-an-email","password":"ValidPass1!","name":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "invalid_email" {
		t.Errorf("expected error 'invalid_email', got %q", errResp["error"])
	}
}

func TestAuthHandler_Register_WeakPassword(t *testing.T) {
	h := newTestAuthHandler(&mockUserStorage{})

	body := `{"email":"user@example.com","password":"weak","name":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "weak_password" {
		t.Errorf("expected error 'weak_password', got %q", errResp["error"])
	}
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	existingUser := &domain.User{
		ID:        "usr_existing",
		Email:     "user@example.com",
		CreatedAt: time.Now(),
	}
	h := newTestAuthHandler(&mockUserStorage{user: existingUser})

	body := `{"email":"user@example.com","password":"ValidPass1!","name":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rr.Code, rr.Body.String())
	}
	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "email_already_exists" {
		t.Errorf("expected error 'email_already_exists', got %q", errResp["error"])
	}
}

// Login tests

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	h := newTestAuthHandler(&mockUserStorage{})

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAuthHandler_Login_UserNotFound(t *testing.T) {
	h := newTestAuthHandler(&mockUserStorage{user: nil})

	body := `{"email":"nobody@example.com","password":"ValidPass1!"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}
	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "invalid_credentials" {
		t.Errorf("expected error 'invalid_credentials', got %q", errResp["error"])
	}
}

func TestAuthHandler_Login_WrongPassword(t *testing.T) {
	hash, _ := utils.HashPassword("CorrectPass1!")
	existingUser := &domain.User{
		ID:           "usr_existing",
		Email:        "user@example.com",
		PasswordHash: hash,
		HouseholdID:  "hh_existing",
		CreatedAt:    time.Now(),
	}
	h := newTestAuthHandler(&mockUserStorage{user: existingUser})

	body := `{"email":"user@example.com","password":"WrongPass1!"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}
	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "invalid_credentials" {
		t.Errorf("expected error 'invalid_credentials', got %q", errResp["error"])
	}
}
