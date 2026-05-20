package middleware

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

// MaxBodyBytes is the maximum allowed request body size (1 MiB).
const MaxBodyBytes int64 = 1 << 20

// BodyLimit rejects incoming requests whose body exceeds MaxBodyBytes with a
// 413 response. Methods that carry no body (GET, HEAD, DELETE) are passed
// through untouched.
func BodyLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodDelete:
			next.ServeHTTP(w, r)
			return
		}
		if r.ContentLength == 0 {
			next.ServeHTTP(w, r)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			var mbErr *http.MaxBytesError
			if errors.As(err, &mbErr) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				_, _ = w.Write([]byte(`{"error":"request_body_too_large"}`))
				return
			}
			http.Error(w, `{"error":"read_error"}`, http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(buf))
		next.ServeHTTP(w, r)
	})
}
