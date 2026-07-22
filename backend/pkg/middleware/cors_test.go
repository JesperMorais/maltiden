package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_AllowedOriginReceivesHeader(t *testing.T) {
	allowed := "https://example.com"
	handler := CORS([]string{allowed})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/recipes", nil)
	req.Header.Set("Origin", allowed)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != allowed {
		t.Fatalf("expected Access-Control-Allow-Origin %q, got %q", allowed, got)
	}
}

func TestCORS_DisallowedOriginOmitsHeader(t *testing.T) {
	handler := CORS([]string{"https://example.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/recipes", nil)
	req.Header.Set("Origin", "https://evil.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no Access-Control-Allow-Origin for unlisted origin, got %q", got)
	}
}

func TestCORS_PreflightReturns200(t *testing.T) {
	allowed := "https://example.com"
	var nextCalled bool
	handler := CORS([]string{allowed})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/recipes", nil)
	req.Header.Set("Origin", allowed)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for OPTIONS preflight, got %d", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("expected Access-Control-Allow-Methods to be set on preflight")
	}
	if got := rr.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Fatal("expected Access-Control-Allow-Headers to be set on preflight")
	}
	if nextCalled {
		t.Fatal("expected next handler NOT to be called on OPTIONS preflight")
	}
}
