package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimit_AllowsRequestsWithinBurst(t *testing.T) {
	rl := NewRateLimiter(1, 5)
	var callCount int
	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rr.Code)
		}
	}
	if callCount != 5 {
		t.Fatalf("expected 5 calls to next handler, got %d", callCount)
	}
}

func TestRateLimit_Returns429AfterBurstExhausted(t *testing.T) {
	rl := NewRateLimiter(1, 3)
	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust the burst
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:9999"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// Next request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:9999"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}
}

func TestRateLimit_IsolatesPerIP(t *testing.T) {
	rl := NewRateLimiter(1, 2)
	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust burst for IP A
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "1.2.3.4:80"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// IP A is rate limited
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:80"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("IP A: expected 429, got %d", rr.Code)
	}

	// IP B is not affected
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "5.6.7.8:80"
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("IP B: expected 200, got %d", rr2.Code)
	}
}

func TestRateLimit_MissingPortFallsBack(t *testing.T) {
	rl := NewRateLimiter(10, 10)
	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1" // no port
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for addr without port, got %d", rr.Code)
	}
}

func TestRateLimit_BehindFlyProxyKeysOnFlyClientIP(t *testing.T) {
	prev := behindFlyProxy
	behindFlyProxy = true
	t.Cleanup(func() { behindFlyProxy = prev })

	rl := NewRateLimiter(0.001, 1)
	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Same proxy address, different visitors: each gets its own bucket.
	for _, visitor := range []string{"203.0.113.1", "203.0.113.2"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "172.16.0.1:443"
		req.Header.Set("Fly-Client-IP", visitor)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("visitor %s: expected 200, got %d", visitor, rr.Code)
		}
	}

	// The first visitor's bucket is now empty.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "172.16.0.1:443"
	req.Header.Set("Fly-Client-IP", "203.0.113.1")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for repeat visitor, got %d", rr.Code)
	}
}

func TestRateLimit_IgnoresFlyClientIPOffFly(t *testing.T) {
	prev := behindFlyProxy
	behindFlyProxy = false
	t.Cleanup(func() { behindFlyProxy = prev })

	rl := NewRateLimiter(0.001, 1)
	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// A client rotating a spoofed header must not get fresh buckets.
	codes := []int{}
	for _, spoofed := range []string{"198.51.100.1", "198.51.100.2"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.5:1234"
		req.Header.Set("Fly-Client-IP", spoofed)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		codes = append(codes, rr.Code)
	}
	if codes[1] != http.StatusTooManyRequests {
		t.Fatalf("expected spoofed header to be ignored (429 on 2nd), got %v", codes)
	}
}
