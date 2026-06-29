package middleware

import (
	"context"
	"encoding/json"
	"log"
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

// AuthInfoProvider returns the user's current token_version and the household
// they currently belong to. The household is resolved live (from the DB) rather
// than read from the JWT claim, because a user's household can change (e.g. when
// they join another household via an invite code) after their token was issued.
type AuthInfoProvider interface {
	GetAuthInfo(userID string) (tokenVersion int, householdID string, err error)
}

func RequireAuth(validator TokenValidator, authInfo AuthInfoProvider) func(http.Handler) http.Handler {
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
			// household removal or password change — and resolve the user's
			// CURRENT household (which may differ from the JWT claim if they
			// joined another household after the token was issued).
			currentVersion, householdID, err := authInfo.GetAuthInfo(claims.UserID)
			if err != nil {
				log.Printf("ERROR [RequireAuth] auth info lookup for user %s: %v", claims.UserID, err)
				writeError(w, http.StatusUnauthorized, "invalid_token")
				return
			}
			if claims.TokenVersion != currentVersion {
				writeError(w, http.StatusUnauthorized, "token_revoked")
				return
			}

			// Add userID and householdID to context. householdID comes from the
			// DB, not the token, so household-scoped resources stay consistent.
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, HouseholdIDKey, householdID)

			// Call next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth extracts JWT claims into context if a valid token is present,
// but does NOT reject requests without a token. Used for public routes that
// behave differently for authenticated users (e.g., scoped recipe listing).
// If the token's version doesn't match the DB, the request proceeds as unauthenticated.
func OptionalAuth(validator TokenValidator, authInfo AuthInfoProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					if claims, err := validator.ValidateToken(parts[1]); err == nil {
						// Check token version — treat as unauthenticated if stale.
						// Resolve the current household live (see RequireAuth).
						if currentVersion, householdID, err := authInfo.GetAuthInfo(claims.UserID); err == nil && claims.TokenVersion == currentVersion {
							ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
							ctx = context.WithValue(ctx, HouseholdIDKey, householdID)
							r = r.WithContext(ctx)
						}
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
