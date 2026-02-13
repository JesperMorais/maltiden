package middleware

import "net/http"

// CORS middleware with configurable allowed origins
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	// Build map for O(1) lookup
	originMap := make(map[string]bool)
	for _, origin := range allowedOrigins {
		originMap[origin] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if origin is allowed
			origin := r.Header.Get("Origin")
			if originMap[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// Allow required methods
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

			// Allow required headers (Content-Type for JSON, Authorization for JWT)
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Allow credentials (cookies, auth headers)
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Cache preflight response for 24 hours
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Handle preflight OPTIONS request
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}
