package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestRequestID(t *testing.T) {
	cases := []struct {
		name          string
		incoming      string
		wantPreserved bool
	}{
		{
			name:          "generates when absent",
			incoming:      "",
			wantPreserved: false,
		},
		{
			name:          "preserves valid incoming",
			incoming:      "test-id-abc123",
			wantPreserved: true,
		},
		{
			name:          "preserves boundary-length incoming",
			incoming:      strings.Repeat("B", maxRequestIDLen),
			wantPreserved: true,
		},
		{
			name:          "replaces oversized incoming",
			incoming:      strings.Repeat("A", maxRequestIDLen+1),
			wantPreserved: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var contextID string
			handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				contextID = GetRequestID(r)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.incoming != "" {
				req.Header.Set("X-Request-ID", tc.incoming)
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			got := rr.Header().Get("X-Request-ID")

			if got == "" {
				t.Fatal("response X-Request-ID must not be empty")
			}
			if contextID != got {
				t.Fatalf("context value %q differs from response header %q", contextID, got)
			}
			if tc.wantPreserved {
				if got != tc.incoming {
					t.Fatalf("expected preserved incoming %q, got %q", tc.incoming, got)
				}
			} else {
				if got == tc.incoming {
					t.Fatalf("expected incoming %q to be replaced, but it was preserved", tc.incoming)
				}
				if len(got) > maxRequestIDLen {
					t.Fatalf("generated ID length %d exceeds cap %d", len(got), maxRequestIDLen)
				}
			}
		})
	}
}

// TestRequestID_OversizedIncomingIsReplaced verifies the maxRequestIDLen
// cap: a client-supplied X-Request-ID longer than the cap is treated as
// absent, and the middleware generates a fresh ID instead of echoing the
// oversized value into logs/headers (claude review feedback on #223).
func TestRequestID_OversizedIncomingIsReplaced(t *testing.T) {
	oversized := strings.Repeat("A", maxRequestIDLen+1)

	var captured string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = GetRequestID(r)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", oversized)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if captured == oversized {
		t.Fatalf("expected oversized incoming ID to be replaced, got it preserved (len=%d)", len(captured))
	}
	if captured == "" {
		t.Fatal("expected a freshly-generated ID, got empty string")
	}
	if len(captured) > maxRequestIDLen {
		t.Fatalf("generated replacement exceeds cap: len=%d, cap=%d", len(captured), maxRequestIDLen)
	}
	// Also at the boundary — exactly maxRequestIDLen chars should be preserved.
	atBoundary := strings.Repeat("B", maxRequestIDLen)
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-Request-ID", atBoundary)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if got := rr2.Header().Get("X-Request-ID"); got != atBoundary {
		t.Fatalf("expected boundary-length ID %q preserved, got %q", atBoundary, got)
	}
}
