package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/utils"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func setupHouseholdTestDB(t *testing.T) *sql.DB {
	t.Helper()

	tmpFile := t.TempDir() + "/test.db"

	origDir, _ := os.Getwd()
	os.Chdir(getHouseholdBackendRoot(t))
	defer os.Chdir(origDir)

	db, err := sqlite.Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func getHouseholdBackendRoot(t *testing.T) string {
	t.Helper()
	// This file is at backend/internal/api/handlers/ — go up 3 levels to reach backend/
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return dir + "/../../.."
}

func setupHouseholdTestJWTService(t *testing.T) *utils.JWTService {
	t.Helper()
	jwtService, err := utils.NewJWTService("test-secret-key-for-jwt-testing-1234567890")
	if err != nil {
		t.Fatalf("failed to create test JWT service: %v", err)
	}
	return jwtService
}

func createHouseholdTestUser(t *testing.T, authService *services.AuthService, email, name string) *domain.AuthResponse {
	t.Helper()
	resp, err := authService.Register(domain.RegisterRequest{
		Email:    email,
		Password: "Testpassword123",
		Name:     name,
	})
	if err != nil {
		t.Fatalf("failed to create test user %s: %v", name, err)
	}
	return resp
}

func newTestHouseholdHandler(t *testing.T) (*HouseholdHandler, *services.AuthService) {
	t.Helper()
	db := setupHouseholdTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	jwtService := setupHouseholdTestJWTService(t)
	authService := services.NewAuthService(db, userStorage, householdStorage, jwtService)
	householdService := services.NewHouseholdService(householdStorage, userStorage)
	handler := NewHouseholdHandler(householdService)
	return handler, authService
}

// TestHouseholdHandler_GetMyHousehold_Unauthorized checks that missing userID returns 401.
func TestHouseholdHandler_GetMyHousehold_Unauthorized(t *testing.T) {
	h, _ := newTestHouseholdHandler(t)

	req := httptest.NewRequest("GET", "/households/me", nil)
	rr := httptest.NewRecorder()
	h.GetMyHousehold(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "unauthorized" {
		t.Errorf("expected error 'unauthorized', got %q", errResp["error"])
	}
}

// TestHouseholdHandler_GetMyHousehold_OK checks that a seeded user gets their household.
func TestHouseholdHandler_GetMyHousehold_OK(t *testing.T) {
	h, authService := newTestHouseholdHandler(t)
	user := createHouseholdTestUser(t, authService, "anna@test.com", "Anna")

	req := httptest.NewRequest("GET", "/households/me", nil)
	req = setAuthContext(req, user.User.ID, user.User.HouseholdID)
	rr := httptest.NewRecorder()
	h.GetMyHousehold(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp domain.Household
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != user.User.HouseholdID {
		t.Errorf("expected household %s, got %s", user.User.HouseholdID, resp.ID)
	}
}

// TestHouseholdHandler_UpdateMyHousehold_Unauthorized checks that missing context returns 401.
func TestHouseholdHandler_UpdateMyHousehold_Unauthorized(t *testing.T) {
	h, _ := newTestHouseholdHandler(t)

	body := `{"name":"New Name"}`
	req := httptest.NewRequest("PUT", "/households/me", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateMyHousehold(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "unauthorized" {
		t.Errorf("expected error 'unauthorized', got %q", errResp["error"])
	}
}

// TestHouseholdHandler_UpdateMyHousehold_NameRequired checks empty name returns 400.
func TestHouseholdHandler_UpdateMyHousehold_NameRequired(t *testing.T) {
	h, authService := newTestHouseholdHandler(t)
	user := createHouseholdTestUser(t, authService, "anna@test.com", "Anna")

	body := `{"name":"   "}`
	req := httptest.NewRequest("PUT", "/households/me", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, user.User.ID, user.User.HouseholdID)
	rr := httptest.NewRecorder()
	h.UpdateMyHousehold(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "household_name_required" {
		t.Errorf("expected error 'household_name_required', got %q", errResp["error"])
	}
}

// TestHouseholdHandler_UpdateMyHousehold_OK checks that a valid name update succeeds.
func TestHouseholdHandler_UpdateMyHousehold_OK(t *testing.T) {
	h, authService := newTestHouseholdHandler(t)
	user := createHouseholdTestUser(t, authService, "anna@test.com", "Anna")

	body := `{"name":"Villa Solsidan"}`
	req := httptest.NewRequest("PUT", "/households/me", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, user.User.ID, user.User.HouseholdID)
	rr := httptest.NewRecorder()
	h.UpdateMyHousehold(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

// TestHouseholdHandler_CreateInvite_Unauthorized checks that missing context returns 401.
func TestHouseholdHandler_CreateInvite_Unauthorized(t *testing.T) {
	h, _ := newTestHouseholdHandler(t)

	req := httptest.NewRequest("POST", "/households/invite", nil)
	rr := httptest.NewRecorder()
	h.CreateInvite(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "unauthorized" {
		t.Errorf("expected error 'unauthorized', got %q", errResp["error"])
	}
}

// TestHouseholdHandler_CreateInvite_OK checks that owner gets a non-empty invite code.
func TestHouseholdHandler_CreateInvite_OK(t *testing.T) {
	h, authService := newTestHouseholdHandler(t)
	user := createHouseholdTestUser(t, authService, "anna@test.com", "Anna")

	req := httptest.NewRequest("POST", "/households/invite", nil)
	req = setAuthContext(req, user.User.ID, user.User.HouseholdID)
	rr := httptest.NewRecorder()
	h.CreateInvite(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp domain.CreateInviteResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code == "" {
		t.Error("expected non-empty invite code")
	}
}

// TestHouseholdHandler_JoinHousehold_Unauthorized checks missing userID returns 401.
func TestHouseholdHandler_JoinHousehold_Unauthorized(t *testing.T) {
	h, _ := newTestHouseholdHandler(t)

	body := `{"code":"ABCD1234"}`
	req := httptest.NewRequest("POST", "/households/join", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.JoinHousehold(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "unauthorized" {
		t.Errorf("expected error 'unauthorized', got %q", errResp["error"])
	}
}

// TestHouseholdHandler_JoinHousehold_CodeRequired checks empty code returns 400.
func TestHouseholdHandler_JoinHousehold_CodeRequired(t *testing.T) {
	h, authService := newTestHouseholdHandler(t)
	user := createHouseholdTestUser(t, authService, "anna@test.com", "Anna")

	body := `{"code":""}`
	req := httptest.NewRequest("POST", "/households/join", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, user.User.ID, user.User.HouseholdID)
	rr := httptest.NewRecorder()
	h.JoinHousehold(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "code_required" {
		t.Errorf("expected error 'code_required', got %q", errResp["error"])
	}
}

// TestHouseholdHandler_JoinHousehold_InvalidCode checks bogus code returns 400 invalid_code.
func TestHouseholdHandler_JoinHousehold_InvalidCode(t *testing.T) {
	h, authService := newTestHouseholdHandler(t)
	user := createHouseholdTestUser(t, authService, "anna@test.com", "Anna")

	body := `{"code":"BADCODE"}`
	req := httptest.NewRequest("POST", "/households/join", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, user.User.ID, user.User.HouseholdID)
	rr := httptest.NewRecorder()
	h.JoinHousehold(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "invalid_code" {
		t.Errorf("expected error 'invalid_code', got %q", errResp["error"])
	}
}

// TestHouseholdHandler_JoinHousehold_OK uses a real invite code from CreateInvite.
func TestHouseholdHandler_JoinHousehold_OK(t *testing.T) {
	h, authService := newTestHouseholdHandler(t)

	owner := createHouseholdTestUser(t, authService, "anna@test.com", "Anna")
	joiner := createHouseholdTestUser(t, authService, "erik@test.com", "Erik")

	// Get invite code via handler
	invReq := httptest.NewRequest("POST", "/households/invite", nil)
	invReq = setAuthContext(invReq, owner.User.ID, owner.User.HouseholdID)
	invRR := httptest.NewRecorder()
	h.CreateInvite(invRR, invReq)
	if invRR.Code != http.StatusCreated {
		t.Fatalf("CreateInvite failed: %d %s", invRR.Code, invRR.Body.String())
	}

	var invResp domain.CreateInviteResponse
	if err := json.NewDecoder(invRR.Body).Decode(&invResp); err != nil {
		t.Fatalf("decode invite: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"code": invResp.Code})
	req := httptest.NewRequest("POST", "/households/join", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, joiner.User.ID, joiner.User.HouseholdID)
	rr := httptest.NewRecorder()
	h.JoinHousehold(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}
