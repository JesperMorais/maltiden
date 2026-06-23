package api

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/utils"
)

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("sqlite.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	jwtSvc, err := utils.NewJWTService("test-secret-for-router-tests")
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}
	return NewRouter(db, jwtSvc)
}

func TestRouterKnownRoutes(t *testing.T) {
	h := newTestRouter(t)

	tests := []struct {
		method string
		path   string
	}{
		// Public routes
		{"GET", "/health"},
		{"POST", "/auth/register"},
		{"POST", "/auth/login"},
		{"GET", "/offers/search"},
		{"GET", "/offers/stores"},
		{"GET", "/recipes"},
		{"GET", "/recipes/some-id"},
		// Protected routes — expect 401, not 404
		{"GET", "/households/me"},
		{"POST", "/households/invite"},
		{"POST", "/menus/generate"},
		{"GET", "/menus/current"},
		{"GET", "/shopping-list"},
		{"POST", "/feedback"},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code == http.StatusNotFound {
				t.Errorf("expected route to be registered, got 404")
			}
		})
	}
}

func TestRouterUnknownRoutes(t *testing.T) {
	h := newTestRouter(t)

	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/nonexistent"},
		{"GET", "/recipes/extra/segments/too/many"},
		{"GET", "/admin"},
		{"POST", "/nonexistent/path"},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusNotFound {
				t.Errorf("expected 404 for unregistered route, got %d", rec.Code)
			}
		})
	}
}
