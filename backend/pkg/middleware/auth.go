package middleware

import (
	"context"
	"encoding/json"
	"maltiden/pkg/utils"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const HouseholdIDKey contextKey = "household_id"

// writeError writes a JSON error response from middleware.
// Duplicated here to avoid import cycle with handlers package.
func writeError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": code})
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get auth header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Extract token (format: Bearer <token>)
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeError(w, http.StatusUnauthorized, "invalid_token_format")
			return
		}
		token := parts[1]

		// Validate JWT
		claims, err := utils.ValidateToken(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid_token")
			return
		}

		// Add userID and householdID to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, HouseholdIDKey, claims.HouseholdID)

		// Call next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper to get userID from context
func GetUserID(r *http.Request) string {
	userID, _ := r.Context().Value(UserIDKey).(string)
	return userID
}

// Helper to get householdID from context
func GetHouseholdID(r *http.Request) string {
	householdID, _ := r.Context().Value(HouseholdIDKey).(string)
	return householdID
}
