package middleware

import "net/http"

// DailyBudget is satisfied by *aibudget.Budget.
type DailyBudget interface {
	Allow(key string) bool
}

// AIBudget rejects a request with 429 ai_quota_exceeded once the caller's
// household (or the whole app) has used up today's paid-LLM quota. It must run
// inside RequireAuth so the household is known; a user without a household is
// keyed by user id instead.
func AIBudget(b DailyBudget) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := GetHouseholdID(r)
			if key == "" {
				key = "user:" + GetUserID(r)
			}
			if !b.Allow(key) {
				writeError(w, http.StatusTooManyRequests, "ai_quota_exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
