package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestHealthHandler_Check_OK(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	defer db.Close()

	h := NewHealthHandler(db)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	h.Check(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", resp["status"])
	}
	if resp["db"] != "connected" {
		t.Errorf("expected db 'connected', got %q", resp["db"])
	}
	if resp["version"] == "" {
		t.Errorf("expected non-empty version, got %q", resp["version"])
	}
}

func TestHealthHandler_Check_ResponseShape(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	defer db.Close()

	h := NewHealthHandler(db)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	h.Check(rr, req)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := resp["status"]; !ok {
		t.Errorf("response body missing key 'status'")
	}
	if _, ok := resp["version"]; !ok {
		t.Errorf("response body missing key 'version'")
	}
}

func TestHealthHandler_Check_DBDown(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	// Close immediately so PingContext fails.
	db.Close()

	h := NewHealthHandler(db)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	h.Check(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["status"] != "error" {
		t.Errorf("expected status 'error', got %q", resp["status"])
	}
	if resp["db"] != "disconnected" {
		t.Errorf("expected db 'disconnected', got %q", resp["db"])
	}
	if resp["version"] == "" {
		t.Errorf("expected non-empty version, got %q", resp["version"])
	}
}
