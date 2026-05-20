package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBodyLimit_AllowsUnderLimit(t *testing.T) {
	body := []byte(`{"name":"test"}`)
	var downstreamCalled bool
	var receivedBody string

	handler := BodyLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		downstreamCalled = true
		b, _ := io.ReadAll(r.Body)
		receivedBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !downstreamCalled {
		t.Fatal("expected downstream handler to be called, but it was not")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if receivedBody != string(body) {
		t.Fatalf("expected body %q, got %q", string(body), receivedBody)
	}
}

func TestBodyLimit_Rejects413(t *testing.T) {
	oversized := strings.Repeat("x", int(MaxBodyBytes)+1)
	var downstreamCalled bool

	handler := BodyLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		downstreamCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(oversized))
	req.ContentLength = int64(len(oversized))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if downstreamCalled {
		t.Fatal("expected downstream handler NOT to be called for oversized body")
	}
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status 413, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "request_body_too_large") {
		t.Fatalf("expected response body to contain 'request_body_too_large', got %q", rr.Body.String())
	}
}
