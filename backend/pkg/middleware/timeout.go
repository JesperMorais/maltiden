package middleware

import (
	"net/http"
	"time"
)

// defaultTimeout is the request processing timeout applied to all API handlers.
const defaultTimeout = 10 * time.Second

// Timeout wraps the given handler with a context-based deadline. If the handler
// does not complete within the timeout, the request is cancelled and the client
// receives a 503 response with a JSON error body.
//
// The timeout message must be valid JSON because http.TimeoutHandler writes it
// directly to the response body when the deadline is exceeded. We wrap the
// handler with an outer adapter that sets Content-Type: application/json so the
// frontend's Axios client parses the timeout body as JSON rather than text.
func Timeout(next http.Handler) http.Handler {
	inner := http.TimeoutHandler(next, defaultTimeout, `{"error":"request_timeout"}`)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		inner.ServeHTTP(w, r)
	})
}
