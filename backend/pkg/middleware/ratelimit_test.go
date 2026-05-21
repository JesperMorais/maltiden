package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimitBurstBoundary(t *testing.T) {
	const burst = 5
	// rate=0 means no token refill during test — purely burst-limited
	rl := NewRateLimiter(0, burst)
	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("allows burst then blocks", func(t *testing.T) {
		for i := 0; i < burst; i++ {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "1.2.3.4:1234"
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("request %d: want 200, got %d", i+1, rec.Code)
			}
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusTooManyRequests {
			t.Fatalf("request %d: want 429, got %d", burst+1, rec.Code)
		}
	})

	t.Run("different IP gets fresh bucket", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "5.6.7.8:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("new IP first request: want 200, got %d", rec.Code)
		}
	})
}
