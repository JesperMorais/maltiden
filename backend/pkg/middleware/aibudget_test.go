package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"maltiden/pkg/aibudget"
)

func TestAIBudget_BlocksAfterHouseholdQuota(t *testing.T) {
	calls := 0
	h := AIBudget(aibudget.New(1, 10))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))

	do := func(household string) int {
		req := httptest.NewRequest(http.MethodPost, "/recipes/parse", nil)
		req = req.WithContext(WithAuthContext(req.Context(), "usr_1", household))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	if code := do("hh_1"); code != http.StatusOK {
		t.Fatalf("first request = %d, want 200", code)
	}
	if code := do("hh_1"); code != http.StatusTooManyRequests {
		t.Fatalf("second request = %d, want 429", code)
	}
	if code := do("hh_2"); code != http.StatusOK {
		t.Fatalf("other household = %d, want 200", code)
	}
	if calls != 2 {
		t.Fatalf("handler ran %d times, want 2", calls)
	}
}
