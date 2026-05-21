package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORS(t *testing.T) {
	t.Run("preflight OPTIONS short-circuits and does not call next", func(t *testing.T) {
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		handler := CORS([]string{"http://localhost:5173"})(next)

		req := httptest.NewRequest(http.MethodOptions, "/api/anything", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		methods := rec.Header().Get("Access-Control-Allow-Methods")
		for _, want := range []string{"OPTIONS", "GET", "POST"} {
			if !strings.Contains(methods, want) {
				t.Errorf("Access-Control-Allow-Methods %q missing %q", methods, want)
			}
		}

		if nextCalled {
			t.Error("next handler must not be called for OPTIONS preflight")
		}
	})

	t.Run("allowed origin echoes in Access-Control-Allow-Origin", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		handler := CORS([]string{"http://localhost:5173"})(next)

		req := httptest.NewRequest(http.MethodOptions, "/api/anything", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
			t.Errorf("expected Access-Control-Allow-Origin %q, got %q", "http://localhost:5173", got)
		}
	})
}
