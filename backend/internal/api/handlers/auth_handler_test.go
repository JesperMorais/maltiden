package handlers

import (
	"bytes"
	"encoding/json"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/utils"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupAuthHandlerTest(t *testing.T) *AuthHandler {
	t.Helper()

	tmpFile := t.TempDir() + "/auth_handler_test.db"
	db, err := sqlite.Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	jwtService, err := utils.NewJWTService("test-secret-key-for-jwt-testing-1234567890")
	if err != nil {
		t.Fatalf("create jwt service: %v", err)
	}

	userStorage := sqlite.NewUserStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	authService := services.NewAuthService(db, userStorage, householdStorage, jwtService)

	return NewAuthHandler(authService)
}

func seedLoginUser(t *testing.T, h *AuthHandler) {
	t.Helper()
	body := `{"email":"login@example.com","password":"Str0ng-Pass!word","name":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("seed register failed: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAuthHandler_Login_ValidCredentials(t *testing.T) {
	h := setupAuthHandlerTest(t)
	seedLoginUser(t, h)

	body := `{"email":"login@example.com","password":"Str0ng-Pass!word"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp domain.AuthResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestAuthHandler_Login_WrongPassword(t *testing.T) {
	h := setupAuthHandlerTest(t)
	seedLoginUser(t, h)

	body := `{"email":"login@example.com","password":"Wr0ng-Pass!word"}`
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

func TestAuthHandler_Login_UnknownUser(t *testing.T) {
	h := setupAuthHandlerTest(t)

	body := `{"email":"nobody@example.com","password":"Str0ng-Pass!word"}`
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

func TestAuthHandler_PasswordResetRequest_Valid(t *testing.T) {
	h := setupAuthHandlerTest(t)

	body := `{"email":"someone@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/password-reset/request", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.PasswordResetRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["message"] == "" {
		t.Error("expected non-empty message field")
	}
}

func TestAuthHandler_PasswordResetRequest_InvalidJSON(t *testing.T) {
	h := setupAuthHandlerTest(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/password-reset/request", bytes.NewBufferString(`not-json`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.PasswordResetRequest(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "invalid_request" {
		t.Errorf("expected error 'invalid_request', got %q", errResp["error"])
	}
}

func TestAuthHandler_PasswordResetRequest_MissingEmail(t *testing.T) {
	h := setupAuthHandlerTest(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/password-reset/request", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.PasswordResetRequest(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "invalid_request" {
		t.Errorf("expected error 'invalid_request', got %q", errResp["error"])
	}
}
