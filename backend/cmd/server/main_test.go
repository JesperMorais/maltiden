package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	browserAccept = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
	axiosAccept   = "application/json, text/plain, */*"
)

func newSPATestHandler(t *testing.T) http.Handler {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<!doctype html>spa"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("js"), 0o644); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /recipes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"recipes":[]}`))
	})
	mux.HandleFunc("GET /menus/current", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"no_current_menu"}`))
	})
	return withSPA(dir, mux)
}

func TestWithSPA(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		accept   string
		wantCode int
		wantBody string
	}{
		{"page load of a client route that is also an API route", "/recipes", browserAccept, 200, "spa"},
		{"page load of an unknown path", "/dashboard", browserAccept, 200, "spa"},
		{"API call reaches the API", "/recipes", axiosAccept, 200, `{"recipes":[]}`},
		{"API 404 stays 404", "/menus/current", axiosAccept, 404, "no_current_menu"},
		{"unknown API path is a 404, not the SPA", "/nope", axiosAccept, 404, "404 page not found"},
		{"static file", "/app.js", axiosAccept, 200, "js"},
	}
	h := newSPATestHandler(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("Accept", tt.accept)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != tt.wantCode {
				t.Fatalf("code = %d, want %d", rr.Code, tt.wantCode)
			}
			if !strings.Contains(rr.Body.String(), tt.wantBody) {
				t.Fatalf("body = %q, want it to contain %q", rr.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestWithSPA_NoFrontendBuiltFallsThroughToAPI(t *testing.T) {
	mux := http.NewServeMux()
	h := withSPA(t.TempDir(), mux)
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req.Header.Set("Accept", browserAccept)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", rr.Code)
	}
}
