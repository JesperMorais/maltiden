package middleware

import (
	"net/http"
	"time"
)

// DefaultTimeout is the request processing timeout applied to most API handlers.
const DefaultTimeout = 10 * time.Second

// ParserTimeout is the longer timeout used for recipe parser routes that call
// the Claude API. Long recipes can routinely take 20–40s to parse, so the
// default 10s deadline produces spurious 503 responses for legitimate work.
const ParserTimeout = 60 * time.Second

// Timeout wraps the given handler with the default request deadline.
//
// If the handler does not complete within the timeout, the request is cancelled
// and the client receives a 503 response with a JSON error body. The timeout
// message must be valid JSON because http.TimeoutHandler writes it directly to
// the response body when the deadline is exceeded; we wrap the handler with an
// outer adapter that sets Content-Type: application/json so the frontend's
// Axios client parses the timeout body as JSON rather than text.
func Timeout(next http.Handler) http.Handler {
	return TimeoutWith(DefaultTimeout)(next)
}

// TimeoutWith returns a middleware that wraps a handler with the given deadline.
// Use this for routes that need a non-default timeout (e.g. Claude API calls).
func TimeoutWith(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		inner := http.TimeoutHandler(next, d, `{"error":"request_timeout"}`)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			inner.ServeHTTP(w, r)
		})
	}
}
