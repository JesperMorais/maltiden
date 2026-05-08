package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_GeneratesWhenAbsent(t *testing.T) {
	var captured string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = GetRequestID(r)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("X-Request-ID")
	if got == "" {
		t.Fatal("expected response X-Request-ID to be set, got empty string")
	}
	if captured == "" {
		t.Fatal("expected GetRequestID to return non-empty value from context")
	}
	if got != captured {
		t.Fatalf("response header (%q) and context value (%q) differ", got, captured)
	}
}

func TestRequestID_PreservesIncomingHeader(t *testing.T) {
	const incoming = "test-id-abc123"

	var captured string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = GetRequestID(r)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", incoming)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("X-Request-ID"); got != incoming {
		t.Fatalf("expected response X-Request-ID %q, got %q", incoming, got)
	}
	if captured != incoming {
		t.Fatalf("expected context value %q, got %q", incoming, captured)
	}
}

func TestGetRequestID_NoMiddleware(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := GetRequestID(req); got != "" {
		t.Fatalf("expected empty string when no middleware wrapped request, got %q", got)
	}
}
