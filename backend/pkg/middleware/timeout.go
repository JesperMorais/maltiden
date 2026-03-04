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
// directly to the response body when the deadline is exceeded.
func Timeout(next http.Handler) http.Handler {
	return http.TimeoutHandler(next, defaultTimeout, `{"error":"request_timeout"}`)
}
