package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const requestIDKey contextKey = "requestID"

// maxRequestIDLen caps the length of a client-supplied X-Request-ID before
// we accept it into our log/header pipeline. Without this, a malicious
// client could send a megabyte-long header value that ends up echoed into
// every structured log line for that request. Generated IDs are always
// 16 hex chars (8 bytes), so 128 is generous-but-bounded.
const maxRequestIDLen = 128

// RequestID generates a unique request ID and adds it to the context and response headers
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Reuse incoming X-Request-ID if present (and within size cap),
		// otherwise generate one. Oversized inbound IDs are silently
		// replaced with a fresh generated ID — no need to error the request.
		requestID := r.Header.Get("X-Request-ID")
		if len(requestID) > maxRequestIDLen {
			requestID = ""
		}
		if requestID == "" {
			// Generate a short request ID (8 bytes = 16 hex chars)
			b := make([]byte, 8)
			if _, err := rand.Read(b); err != nil {
				// Fallback to a simple string if random generation fails
				b = []byte{0, 0, 0, 0, 0, 0, 0, 1}
			}
			requestID = hex.EncodeToString(b)
		}

		// Add to response header
		w.Header().Set("X-Request-ID", requestID)

		// Add to request context
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)

		// Continue to next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(r *http.Request) string {
	if id, ok := r.Context().Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}
