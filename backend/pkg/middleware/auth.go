package middleware

import (
	"context"
	"maltiden/pkg/utils"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get auth header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// Extract token (format: Bearer <token>)
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"invalid_token_format"}`, http.StatusUnauthorized)
			return
		}
		token := parts[1]

		// Validate JWT
		claims, err := utils.ValidateToken(token)
		if err != nil {
			http.Error(w, `{"error":"invalid_token"}`, http.StatusUnauthorized)
			return
		}

		// Add userID to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)

		// Call next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper to get userID from context
func GetUserID(r *http.Request) string {
	userID, _ := r.Context().Value(UserIDKey).(string)
	return userID
}
