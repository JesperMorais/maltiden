package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORS_AllowedOriginEchoed(t *testing.T) {
	const allowed = "https://app.example.com"

	invoked := false
	handler := CORS([]string{allowed})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		invoked = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", allowed)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != allowed {
		t.Fatalf("expected Access-Control-Allow-Origin %q, got %q", allowed, got)
	}
	if !invoked {
		t.Fatal("expected downstream handler to be invoked, but it was not")
	}
}

func TestCORS_DisallowedOriginNotEchoed(t *testing.T) {
	handler := CORS([]string{"https://app.example.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected empty Access-Control-Allow-Origin for disallowed origin, got %q", got)
	}
}

func TestCORS_OPTIONSPreflightShortCircuits(t *testing.T) {
	const allowed = "https://app.example.com"

	invoked := false
	handler := CORS([]string{allowed})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		invoked = true
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", allowed)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if invoked {
		t.Fatal("expected downstream handler NOT to be invoked on OPTIONS preflight, but it was")
	}

	if status := rr.Code; status < 200 || status >= 300 {
		t.Fatalf("expected 2xx status on preflight, got %d", status)
	}

	methods := rr.Header().Get("Access-Control-Allow-Methods")
	for _, m := range []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"} {
		if !strings.Contains(methods, m) {
			t.Fatalf("expected Access-Control-Allow-Methods to contain %q, got %q", m, methods)
		}
	}

	headers := rr.Header().Get("Access-Control-Allow-Headers")
	for _, h := range []string{"Content-Type", "Authorization"} {
		if !strings.Contains(headers, h) {
			t.Fatalf("expected Access-Control-Allow-Headers to contain %q, got %q", h, headers)
		}
	}
}
