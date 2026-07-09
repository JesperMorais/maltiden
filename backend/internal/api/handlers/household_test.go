package handlers

import (
	"database/sql"
	"encoding/json"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/middleware"
	"maltiden/pkg/utils"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	tmpFile := t.TempDir() + "/test.db"

	origDir, _ := os.Getwd()
	os.Chdir(getBackendRoot(t))
	defer os.Chdir(origDir)

	db, err := sqlite.Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func getBackendRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// This file is at backend/internal/api/handlers/household_test.go
	return dir + "/../../.."
}

func newTestHouseholdHandler(t *testing.T) (*HouseholdHandler, *domain.AuthResponse) {
	t.Helper()

	db := setupTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	jwtService, err := utils.NewJWTService("test-secret-key-for-jwt-testing-1234567890")
	if err != nil {
		t.Fatalf("failed to create jwt service: %v", err)
	}
	authService := services.NewAuthService(db, userStorage, householdStorage, jwtService)
	householdService := services.NewHouseholdService(householdStorage, userStorage)

	user, err := authService.Register(domain.RegisterRequest{
		Email:    "anna@test.com",
		Password: "Testpassword123",
		Name:     "Anna",
	})
	if err != nil {
		t.Fatalf("failed to register test user: %v", err)
	}

	return NewHouseholdHandler(householdService), user
}

func TestHouseholdHandler_CreateInvite(t *testing.T) {
	h, user := newTestHouseholdHandler(t)

	req := httptest.NewRequest("POST", "/households/invite", nil)
	ctx := middleware.WithAuthContext(req.Context(), user.User.ID, user.User.HouseholdID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.CreateInvite(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp domain.CreateInviteResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Code) != 8 {
		t.Errorf("expected 8-char invite code, got %d chars: %q", len(resp.Code), resp.Code)
	}
}

func TestHouseholdHandler_GetMyHousehold_NotFound(t *testing.T) {
	h, _ := newTestHouseholdHandler(t)

	req := httptest.NewRequest("GET", "/households/me", nil)
	ctx := middleware.WithAuthContext(req.Context(), "usr_does-not-exist", "")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.GetMyHousehold(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "household_not_found" {
		t.Errorf("expected error 'household_not_found', got %q", errResp["error"])
	}
}
