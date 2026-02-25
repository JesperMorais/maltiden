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
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}

// TokenValidator defines the interface for validating JWT tokens.
type TokenValidator interface {
	ValidateToken(tokenString string) (*utils.Claims, error)
}

// TokenVersionChecker checks whether a JWT's token_version is still current.
type TokenVersionChecker interface {
	GetTokenVersion(userID string) (int, error)
}

func RequireAuth(validator TokenValidator, versionChecker TokenVersionChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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
			claims, err := validator.ValidateToken(token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid_token")
				return
			}

			// Check token version against DB — rejects stale tokens after
			// household removal or password change
			currentVersion, err := versionChecker.GetTokenVersion(claims.UserID)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid_token")
				return
			}
			if claims.TokenVersion != currentVersion {
				writeError(w, http.StatusUnauthorized, "token_revoked")
				return
			}

			// Add userID and householdID to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, HouseholdIDKey, claims.HouseholdID)

			// Call next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth extracts JWT claims into context if a valid token is present,
// but does NOT reject requests without a token. Used for public routes that
// behave differently for authenticated users (e.g., scoped recipe listing).
func OptionalAuth(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					if claims, err := validator.ValidateToken(parts[1]); err == nil {
						ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
						ctx = context.WithValue(ctx, HouseholdIDKey, claims.HouseholdID)
						r = r.WithContext(ctx)
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
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

// WithAuthContext returns a context with userID and householdID set,
// matching what RequireAuth injects. Useful for handler tests.
func WithAuthContext(ctx context.Context, userID, householdID string) context.Context {
	ctx = context.WithValue(ctx, UserIDKey, userID)
	ctx = context.WithValue(ctx, HouseholdIDKey, householdID)
	return ctx
}
