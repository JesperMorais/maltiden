package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// prMockUserStorage is a configurable domain.UserRepository stub for the
// password-reset handler tests. It is intentionally separate from
// mockUserStorage (auth_test.go) so each suite can drive its own behaviour.
type prMockUserStorage struct {
	user            *domain.User
	getByEmailErr   error
	resetToken      *domain.PasswordResetToken
	getTokenErr     error
	createTokenErr  error
	createTokenSeen bool
}

func (m *prMockUserStorage) GetByEmail(_ string) (*domain.User, error) {
	return m.user, m.getByEmailErr
}
func (m *prMockUserStorage) GetByID(_ string) (*domain.User, error)   { return nil, nil }
func (m *prMockUserStorage) Create(_ *domain.User) error              { return nil }
func (m *prMockUserStorage) CreateTx(_ *sql.Tx, _ *domain.User) error { return nil }
func (m *prMockUserStorage) GetTokenVersion(_ string) (int, error)    { return 0, nil }
func (m *prMockUserStorage) IncrementTokenVersion(_ string) error     { return nil }
func (m *prMockUserStorage) IncrementTokenVersionTx(_ *sql.Tx, _ string) error {
	return nil
}
func (m *prMockUserStorage) UpdatePassword(_, _ string) error              { return nil }
func (m *prMockUserStorage) UpdatePasswordTx(_ *sql.Tx, _, _ string) error { return nil }
func (m *prMockUserStorage) CreatePasswordResetToken(_ *domain.PasswordResetToken) error {
	m.createTokenSeen = true
	return m.createTokenErr
}
func (m *prMockUserStorage) GetPasswordResetToken(_ string) (*domain.PasswordResetToken, error) {
	return m.resetToken, m.getTokenErr
}
func (m *prMockUserStorage) MarkPasswordResetTokenUsed(_ string) error { return nil }
func (m *prMockUserStorage) MarkPasswordResetTokenUsedTx(_ *sql.Tx, _ string) error {
	return nil
}

// prMockEmailer records the last sent email and can be made to fail. A failing
// emailer must NOT change the handler's response (no account-existence leak).
type prMockEmailer struct {
	sent    bool
	to      string
	sendErr error
}

func (m *prMockEmailer) Send(to, _, _ string) error {
	m.sent = true
	m.to = to
	return m.sendErr
}

func newTestPasswordResetHandler(store *prMockUserStorage, emailer *prMockEmailer) *PasswordResetHandler {
	// db is nil: every path exercised here returns before db.Begin(), so the
	// transaction (success) path is left to the service-level test.
	svc := services.NewPasswordResetService(nil, store, emailer)
	return NewPasswordResetHandler(svc)
}

func doRequest(t *testing.T, h func(http.ResponseWriter, *http.Request), path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h(rr, req)
	return rr
}

func decodeErr(t *testing.T, rr *httptest.ResponseRecorder) string {
	t.Helper()
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode error body %q: %v", rr.Body.String(), err)
	}
	return resp["error"]
}

func knownUser() *domain.User {
	return &domain.User{
		ID:        "usr_known",
		Email:     "user@example.com",
		Name:      "Test",
		CreatedAt: time.Now(),
	}
}

// --- ForgotPassword ---------------------------------------------------------

// Happy path: a known email triggers a token + email and still returns 200/ok.
func TestPasswordResetHandler_ForgotPassword_KnownEmail(t *testing.T) {
	store := &prMockUserStorage{user: knownUser()}
	emailer := &prMockEmailer{}
	h := newTestPasswordResetHandler(store, emailer)

	rr := doRequest(t, h.ForgotPassword, "/auth/forgot-password", `{"email":"user@example.com"}`)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]bool
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !resp["ok"] {
		t.Errorf("expected {\"ok\":true}, got %s", rr.Body.String())
	}
	if !store.createTokenSeen {
		t.Error("expected a reset token to be created for a known email")
	}
	if !emailer.sent {
		t.Error("expected a reset email to be sent for a known email")
	}
}

// Unknown email: must NOT leak account existence — same 200/ok as the known
// case, and no token created / no email sent.
func TestPasswordResetHandler_ForgotPassword_UnknownEmail_NoLeak(t *testing.T) {
	store := &prMockUserStorage{user: nil} // GetByEmail returns no user
	emailer := &prMockEmailer{}
	h := newTestPasswordResetHandler(store, emailer)

	rr := doRequest(t, h.ForgotPassword, "/auth/forgot-password", `{"email":"nobody@example.com"}`)

	if rr.Code != http.StatusOK {
		t.Fatalf("unknown email must return 200 (no leak), got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]bool
	json.NewDecoder(rr.Body).Decode(&resp)
	if !resp["ok"] {
		t.Errorf("unknown email must return {\"ok\":true} (no leak), got %s", rr.Body.String())
	}
	if store.createTokenSeen {
		t.Error("no reset token should be created for an unknown email")
	}
	if emailer.sent {
		t.Error("no email should be sent for an unknown email")
	}
}

// The known and unknown responses must be byte-identical so a caller cannot
// distinguish registered from unregistered accounts.
func TestPasswordResetHandler_ForgotPassword_KnownAndUnknown_Indistinguishable(t *testing.T) {
	hKnown := newTestPasswordResetHandler(&prMockUserStorage{user: knownUser()}, &prMockEmailer{})
	hUnknown := newTestPasswordResetHandler(&prMockUserStorage{user: nil}, &prMockEmailer{})

	rrKnown := doRequest(t, hKnown.ForgotPassword, "/auth/forgot-password", `{"email":"user@example.com"}`)
	rrUnknown := doRequest(t, hUnknown.ForgotPassword, "/auth/forgot-password", `{"email":"nobody@example.com"}`)

	if rrKnown.Code != rrUnknown.Code {
		t.Errorf("status differs (known=%d unknown=%d) — leaks account existence", rrKnown.Code, rrUnknown.Code)
	}
	if rrKnown.Body.String() != rrUnknown.Body.String() {
		t.Errorf("body differs (known=%q unknown=%q) — leaks account existence", rrKnown.Body.String(), rrUnknown.Body.String())
	}
}

// Invalid email format is treated as a silent no-op and still returns 200/ok.
func TestPasswordResetHandler_ForgotPassword_InvalidEmail_NoLeak(t *testing.T) {
	store := &prMockUserStorage{user: nil}
	emailer := &prMockEmailer{}
	h := newTestPasswordResetHandler(store, emailer)

	rr := doRequest(t, h.ForgotPassword, "/auth/forgot-password", `{"email":"not-an-email"}`)

	if rr.Code != http.StatusOK {
		t.Fatalf("invalid email must return 200 (no leak), got %d: %s", rr.Code, rr.Body.String())
	}
	if store.createTokenSeen || emailer.sent {
		t.Error("invalid email must not create a token or send mail")
	}
}

// Missing/empty email behaves like an invalid email: 200/ok, no work done.
func TestPasswordResetHandler_ForgotPassword_MissingEmail_NoLeak(t *testing.T) {
	store := &prMockUserStorage{}
	emailer := &prMockEmailer{}
	h := newTestPasswordResetHandler(store, emailer)

	rr := doRequest(t, h.ForgotPassword, "/auth/forgot-password", `{}`)

	if rr.Code != http.StatusOK {
		t.Fatalf("missing email must return 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if store.createTokenSeen || emailer.sent {
		t.Error("missing email must not create a token or send mail")
	}
}

// A real DB lookup error must still surface 200/ok to the client (logged
// server-side only) — again, no account-existence signal.
func TestPasswordResetHandler_ForgotPassword_LookupError_StillOK(t *testing.T) {
	store := &prMockUserStorage{getByEmailErr: errors.New("db down")}
	emailer := &prMockEmailer{}
	h := newTestPasswordResetHandler(store, emailer)

	rr := doRequest(t, h.ForgotPassword, "/auth/forgot-password", `{"email":"user@example.com"}`)

	if rr.Code != http.StatusOK {
		t.Fatalf("internal lookup error must still return 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]bool
	json.NewDecoder(rr.Body).Decode(&resp)
	if !resp["ok"] {
		t.Errorf("expected {\"ok\":true} even on internal error, got %s", rr.Body.String())
	}
}

// A failing emailer must not change the response (transient outage must not
// become an account-existence oracle).
func TestPasswordResetHandler_ForgotPassword_EmailSendFails_StillOK(t *testing.T) {
	store := &prMockUserStorage{user: knownUser()}
	emailer := &prMockEmailer{sendErr: errors.New("smtp boom")}
	h := newTestPasswordResetHandler(store, emailer)

	rr := doRequest(t, h.ForgotPassword, "/auth/forgot-password", `{"email":"user@example.com"}`)

	if rr.Code != http.StatusOK {
		t.Fatalf("email send failure must still return 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestPasswordResetHandler_ForgotPassword_InvalidJSON(t *testing.T) {
	h := newTestPasswordResetHandler(&prMockUserStorage{}, &prMockEmailer{})

	rr := doRequest(t, h.ForgotPassword, "/auth/forgot-password", "not json")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed JSON, got %d: %s", rr.Code, rr.Body.String())
	}
}

// --- ResetPassword ----------------------------------------------------------

func TestPasswordResetHandler_ResetPassword_InvalidJSON(t *testing.T) {
	h := newTestPasswordResetHandler(&prMockUserStorage{}, &prMockEmailer{})

	rr := doRequest(t, h.ResetPassword, "/auth/reset-password", "not json")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed JSON, got %d: %s", rr.Code, rr.Body.String())
	}
}

// Empty/missing token short-circuits to invalid_reset_token before any password
// strength check.
func TestPasswordResetHandler_ResetPassword_MissingToken(t *testing.T) {
	h := newTestPasswordResetHandler(&prMockUserStorage{}, &prMockEmailer{})

	rr := doRequest(t, h.ResetPassword, "/auth/reset-password", `{"token":"","newPassword":"ValidPass1!"}`)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	if code := decodeErr(t, rr); code != "invalid_reset_token" {
		t.Errorf("expected 'invalid_reset_token', got %q", code)
	}
}

func TestPasswordResetHandler_ResetPassword_WeakPassword(t *testing.T) {
	h := newTestPasswordResetHandler(&prMockUserStorage{}, &prMockEmailer{})

	rr := doRequest(t, h.ResetPassword, "/auth/reset-password", `{"token":"sometoken","newPassword":"weak"}`)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	if code := decodeErr(t, rr); code != "weak_password" {
		t.Errorf("expected 'weak_password', got %q", code)
	}
}

// Token not found in storage → invalid_reset_token.
func TestPasswordResetHandler_ResetPassword_UnknownToken(t *testing.T) {
	store := &prMockUserStorage{resetToken: nil}
	h := newTestPasswordResetHandler(store, &prMockEmailer{})

	rr := doRequest(t, h.ResetPassword, "/auth/reset-password", `{"token":"deadbeef","newPassword":"ValidPass1!"}`)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	if code := decodeErr(t, rr); code != "invalid_reset_token" {
		t.Errorf("expected 'invalid_reset_token', got %q", code)
	}
}

// A token already consumed (UsedAt set) → used_reset_token.
func TestPasswordResetHandler_ResetPassword_UsedToken(t *testing.T) {
	used := time.Now().Add(-time.Minute)
	store := &prMockUserStorage{resetToken: &domain.PasswordResetToken{
		Token:     "deadbeef",
		UserID:    "usr_known",
		ExpiresAt: time.Now().Add(time.Hour),
		UsedAt:    &used,
	}}
	h := newTestPasswordResetHandler(store, &prMockEmailer{})

	rr := doRequest(t, h.ResetPassword, "/auth/reset-password", `{"token":"deadbeef","newPassword":"ValidPass1!"}`)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	if code := decodeErr(t, rr); code != "used_reset_token" {
		t.Errorf("expected 'used_reset_token', got %q", code)
	}
}

// A token past its expiry → expired_reset_token.
func TestPasswordResetHandler_ResetPassword_ExpiredToken(t *testing.T) {
	store := &prMockUserStorage{resetToken: &domain.PasswordResetToken{
		Token:     "deadbeef",
		UserID:    "usr_known",
		ExpiresAt: time.Now().Add(-time.Hour),
	}}
	h := newTestPasswordResetHandler(store, &prMockEmailer{})

	rr := doRequest(t, h.ResetPassword, "/auth/reset-password", `{"token":"deadbeef","newPassword":"ValidPass1!"}`)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	if code := decodeErr(t, rr); code != "expired_reset_token" {
		t.Errorf("expected 'expired_reset_token', got %q", code)
	}
}

// A non-domain storage error from token lookup maps to 500 internal_error.
func TestPasswordResetHandler_ResetPassword_StorageError(t *testing.T) {
	store := &prMockUserStorage{getTokenErr: errors.New("db down")}
	h := newTestPasswordResetHandler(store, &prMockEmailer{})

	rr := doRequest(t, h.ResetPassword, "/auth/reset-password", `{"token":"deadbeef","newPassword":"ValidPass1!"}`)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rr.Code, rr.Body.String())
	}
	if code := decodeErr(t, rr); code != "internal_error" {
		t.Errorf("expected 'internal_error', got %q", code)
	}
}
