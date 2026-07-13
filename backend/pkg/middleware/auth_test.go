package middleware

import (
	"context"
	"encoding/json"
	"maltiden/pkg/utils"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeValidator struct {
	claims *utils.Claims
	err    error
}

func (f *fakeValidator) ValidateToken(string) (*utils.Claims, error) {
	return f.claims, f.err
}

type fakeVersionChecker struct {
	version int
	err     error
}

func (f *fakeVersionChecker) GetTokenVersion(string) (int, error) {
	return f.version, f.err
}

func newNext() (http.Handler, *string, *string) {
	var gotUser, gotHousehold string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser = GetUserID(r)
		gotHousehold = GetHouseholdID(r)
		w.WriteHeader(http.StatusOK)
	})
	return h, &gotUser, &gotHousehold
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return body["error"]
}

func TestRequireAuth_FailurePaths(t *testing.T) {
	claims := &utils.Claims{UserID: "usr_1", HouseholdID: "hh_1", TokenVersion: 2}

	tests := []struct {
		name       string
		authHeader string
		validator  *fakeValidator
		checker    *fakeVersionChecker
		wantCode   string
	}{
		{"no header", "", &fakeValidator{}, &fakeVersionChecker{}, "unauthorized"},
		{"bad scheme", "Basic abc", &fakeValidator{}, &fakeVersionChecker{}, "invalid_token_format"},
		{"validate err", "Bearer tok", &fakeValidator{err: errFake}, &fakeVersionChecker{}, "invalid_token"},
		{"version check err", "Bearer tok", &fakeValidator{claims: claims}, &fakeVersionChecker{err: errFake}, "invalid_token"},
		{"version mismatch", "Bearer tok", &fakeValidator{claims: claims}, &fakeVersionChecker{version: 1}, "token_revoked"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			next, _, _ := newNext()
			handler := RequireAuth(tc.validator, tc.checker)(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", rec.Code)
			}
			if got := decodeError(t, rec); got != tc.wantCode {
				t.Fatalf("error = %q, want %q", got, tc.wantCode)
			}
		})
	}
}

func TestRequireAuth_Success(t *testing.T) {
	claims := &utils.Claims{UserID: "usr_1", HouseholdID: "hh_1", TokenVersion: 2}
	next, gotUser, gotHousehold := newNext()
	handler := RequireAuth(&fakeValidator{claims: claims}, &fakeVersionChecker{version: 2})(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer tok")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if *gotUser != "usr_1" || *gotHousehold != "hh_1" {
		t.Fatalf("context ids = %q/%q, want usr_1/hh_1", *gotUser, *gotHousehold)
	}
}

func TestOptionalAuth(t *testing.T) {
	claims := &utils.Claims{UserID: "usr_1", HouseholdID: "hh_1", TokenVersion: 2}

	tests := []struct {
		name          string
		authHeader    string
		validator     *fakeValidator
		checker       *fakeVersionChecker
		wantUser      string
		wantHousehold string
	}{
		{"no header", "", &fakeValidator{}, &fakeVersionChecker{}, "", ""},
		{"invalid token", "Bearer bad", &fakeValidator{err: errFake}, &fakeVersionChecker{}, "", ""},
		{"valid token", "Bearer tok", &fakeValidator{claims: claims}, &fakeVersionChecker{version: 2}, "usr_1", "hh_1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			next, gotUser, gotHousehold := newNext()
			handler := OptionalAuth(tc.validator, tc.checker)(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if *gotUser != tc.wantUser || *gotHousehold != tc.wantHousehold {
				t.Fatalf("context ids = %q/%q, want %q/%q", *gotUser, *gotHousehold, tc.wantUser, tc.wantHousehold)
			}
		})
	}
}

func TestGetters_EmptyContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if GetUserID(req) != "" || GetHouseholdID(req) != "" {
		t.Fatal("expected empty ids on bare context")
	}
}

func TestWithAuthContext_RoundTrip(t *testing.T) {
	ctx := WithAuthContext(context.Background(), "usr_2", "hh_2")
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	if GetUserID(req) != "usr_2" || GetHouseholdID(req) != "hh_2" {
		t.Fatal("WithAuthContext did not round-trip ids")
	}
}

var errFake = &fakeError{"fake error"}

type fakeError struct{ msg string }

func (e *fakeError) Error() string { return e.msg }
