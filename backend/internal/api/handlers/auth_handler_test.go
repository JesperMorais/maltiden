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
	"strings"
	"testing"
)

func getHandlerBackendRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// We're in internal/api/handlers, go up three levels to reach backend/
	return dir + "/../../.."
}

func setupAuthTestDB(t *testing.T) *sql.DB {
	t.Helper()
	tmpFile := t.TempDir() + "/test.db"

	origDir, _ := os.Getwd()
	os.Chdir(getHandlerBackendRoot(t))
	defer os.Chdir(origDir)

	db, err := sqlite.Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func newTestAuthHandler(t *testing.T) (*AuthHandler, *sql.DB) {
	t.Helper()
	db := setupAuthTestDB(t)
	userStorage := sqlite.NewUserStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	jwtSvc, err := utils.NewJWTService("test-secret-key-for-jwt-testing-1234567890")
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}
	h := NewAuthHandler(services.NewAuthService(db, userStorage, householdStorage, jwtSvc))
	return h, db
}

func registerRequest(t *testing.T, h *AuthHandler, payload string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	return rr
}

func TestAuthHandler_Register(t *testing.T) {
	t.Run("201 happy path", func(t *testing.T) {
		h, _ := newTestAuthHandler(t)
		payload := `{"email":"anna@example.com","password":"Testpassword123","name":"Anna"}`
		rr := registerRequest(t, h, payload)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp domain.AuthResponse
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if resp.Token == "" {
			t.Error("expected non-empty token")
		}
		if !strings.HasPrefix(resp.User.ID, "usr_") {
			t.Errorf("expected user ID to start with 'usr_', got %q", resp.User.ID)
		}
	})

	t.Run("409 duplicate email", func(t *testing.T) {
		h, _ := newTestAuthHandler(t)
		payload := `{"email":"anna@example.com","password":"Testpassword123","name":"Anna"}`

		// First registration succeeds
		rr := registerRequest(t, h, payload)
		if rr.Code != http.StatusCreated {
			t.Fatalf("first register: expected 201, got %d: %s", rr.Code, rr.Body.String())
		}

		// Second registration with same email
		rr = registerRequest(t, h, payload)
		if rr.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d: %s", rr.Code, rr.Body.String())
		}

		var body map[string]string
		if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
			t.Fatalf("decode error body: %v", err)
		}
		if body["error"] != "email_already_exists" {
			t.Errorf("expected error 'email_already_exists', got %q", body["error"])
		}
	})

	t.Run("400 weak password", func(t *testing.T) {
		h, _ := newTestAuthHandler(t)
		payload := `{"email":"anna@example.com","password":"short","name":"Anna"}`
		rr := registerRequest(t, h, payload)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
		}

		var body map[string]string
		if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
			t.Fatalf("decode error body: %v", err)
		}
		if body["error"] == "" {
			t.Error("expected non-empty error field")
		}
	})
}
