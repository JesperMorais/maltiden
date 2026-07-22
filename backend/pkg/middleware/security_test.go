package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newSecurityHandler() http.Handler {
	return Security(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

func TestSecurity_Headers(t *testing.T) {
	handler := newSecurityHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	checks := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}
	for header, want := range checks {
		if got := rr.Header().Get(header); got != want {
			t.Errorf("%s: got %q, want %q", header, got, want)
		}
	}
}

func TestSecurity_HSTSOnlyInProduction(t *testing.T) {
	os.Unsetenv("FLY_APP_NAME")
	handler := newSecurityHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if got := rr.Header().Get("Strict-Transport-Security"); got != "" {
		t.Errorf("HSTS should not be set without FLY_APP_NAME, got %q", got)
	}

	t.Setenv("FLY_APP_NAME", "maltiden")
	handler = newSecurityHandler()
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if got := rr.Header().Get("Strict-Transport-Security"); got == "" {
		t.Error("HSTS should be set when FLY_APP_NAME is set")
	}
}
