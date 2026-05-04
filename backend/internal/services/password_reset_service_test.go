package services

import (
	"database/sql"
	"maltiden/internal/domain"
	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/email"
	"maltiden/pkg/utils"
	"sync"
	"testing"
	"time"
)

// captureSender records email Send calls for assertions in tests.
type captureSender struct {
	mu    sync.Mutex
	calls []captureCall
}

type captureCall struct {
	To      string
	Subject string
	Body    string
}

func (c *captureSender) Send(to, subject, body string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, captureCall{To: to, Subject: subject, Body: body})
	return nil
}

func (c *captureSender) Calls() []captureCall {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]captureCall, len(c.calls))
	copy(out, c.calls)
	return out
}

// setupResetTest builds the dependencies for password-reset tests and registers
// one user.
func setupResetTest(t *testing.T) (*PasswordResetService, *sqlite.UserStorage, *captureSender, *domain.AuthResponse) {
	t.Helper()
	db := setupTestDB(t)
	userStorage := sqlite.NewUserStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	user := createTestUser(t, authService, "reset@test.com", "Reset")

	sender := &captureSender{}
	prs := NewPasswordResetService(db, userStorage, sender)
	return prs, userStorage, sender, user
}

func TestRequestReset_CreatesToken(t *testing.T) {
	prs, _, sender, _ := setupResetTest(t)

	if err := prs.RequestReset("reset@test.com"); err != nil {
		t.Fatalf("RequestReset: %v", err)
	}

	calls := sender.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(calls))
	}
	if calls[0].To != "reset@test.com" {
		t.Errorf("wrong recipient: %s", calls[0].To)
	}
	if len(calls[0].Body) == 0 {
		t.Error("empty email body")
	}
}

func TestRequestReset_NoLeakOnUnknownEmail(t *testing.T) {
	prs, _, sender, _ := setupResetTest(t)

	if err := prs.RequestReset("unknown@nowhere.com"); err != nil {
		t.Fatalf("RequestReset: %v", err)
	}

	if got := len(sender.Calls()); got != 0 {
		t.Errorf("expected zero emails for unknown account, got %d", got)
	}
}

// extractTokenFromEmail pulls the token query value from a reset URL embedded
// in the email body.
func extractTokenFromEmail(body string) string {
	const marker = "token="
	idx := -1
	for i := 0; i+len(marker) <= len(body); i++ {
		if body[i:i+len(marker)] == marker {
			idx = i + len(marker)
			break
		}
	}
	if idx < 0 {
		return ""
	}
	end := idx
	for end < len(body) {
		c := body[end]
		// Tokens are hex; stop at any non-hex char.
		isHex := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
		if !isHex {
			break
		}
		end++
	}
	return body[idx:end]
}

func TestResetPassword_Success(t *testing.T) {
	prs, userStorage, sender, user := setupResetTest(t)

	if err := prs.RequestReset(user.User.Email); err != nil {
		t.Fatalf("RequestReset: %v", err)
	}
	calls := sender.Calls()
	token := extractTokenFromEmail(calls[0].Body)
	if token == "" {
		t.Fatalf("could not extract token from email body: %s", calls[0].Body)
	}

	if err := prs.ResetPassword(token, "Newpassword123"); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}

	// Verify the password actually changed in DB.
	updated, err := userStorage.GetByID(user.User.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if !utils.CheckPassword("Newpassword123", updated.PasswordHash) {
		t.Error("new password does not verify")
	}
}

func TestResetPassword_RejectsUsedToken(t *testing.T) {
	prs, _, sender, user := setupResetTest(t)

	if err := prs.RequestReset(user.User.Email); err != nil {
		t.Fatalf("RequestReset: %v", err)
	}
	token := extractTokenFromEmail(sender.Calls()[0].Body)

	if err := prs.ResetPassword(token, "Newpassword123"); err != nil {
		t.Fatalf("first ResetPassword: %v", err)
	}
	err := prs.ResetPassword(token, "Anotherpw123A")
	if err != domain.ErrUsedResetToken {
		t.Errorf("expected ErrUsedResetToken, got %v", err)
	}
}

func TestResetPassword_RejectsExpiredToken(t *testing.T) {
	prs, userStorage, _, user := setupResetTest(t)

	// Insert a manually-expired token directly via the repo.
	expired := &domain.PasswordResetToken{
		Token:     "deadbeef" + "00000000000000000000000000000000000000000000000000000000",
		UserID:    user.User.ID,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}
	if err := userStorage.CreatePasswordResetToken(expired); err != nil {
		t.Fatalf("CreatePasswordResetToken: %v", err)
	}

	err := prs.ResetPassword(expired.Token, "Newpassword123")
	if err != domain.ErrExpiredResetToken {
		t.Errorf("expected ErrExpiredResetToken, got %v", err)
	}
}

func TestResetPassword_RejectsInvalidToken(t *testing.T) {
	prs, _, _, _ := setupResetTest(t)

	err := prs.ResetPassword("nonexistent-token", "Newpassword123")
	if err != domain.ErrInvalidResetToken {
		t.Errorf("expected ErrInvalidResetToken, got %v", err)
	}
}

func TestResetPassword_BumpsTokenVersion(t *testing.T) {
	prs, userStorage, sender, user := setupResetTest(t)

	versionBefore, err := userStorage.GetTokenVersion(user.User.ID)
	if err != nil {
		t.Fatalf("GetTokenVersion: %v", err)
	}

	if err := prs.RequestReset(user.User.Email); err != nil {
		t.Fatalf("RequestReset: %v", err)
	}
	token := extractTokenFromEmail(sender.Calls()[0].Body)
	if err := prs.ResetPassword(token, "Newpassword123"); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}

	versionAfter, err := userStorage.GetTokenVersion(user.User.ID)
	if err != nil {
		t.Fatalf("GetTokenVersion: %v", err)
	}
	if versionAfter != versionBefore+1 {
		t.Errorf("expected token_version to increment by 1 (was %d, now %d)", versionBefore, versionAfter)
	}
}

func TestResetPassword_RejectsWeakPassword(t *testing.T) {
	prs, _, sender, user := setupResetTest(t)

	if err := prs.RequestReset(user.User.Email); err != nil {
		t.Fatalf("RequestReset: %v", err)
	}
	token := extractTokenFromEmail(sender.Calls()[0].Body)

	err := prs.ResetPassword(token, "short")
	if err != domain.ErrWeakPassword {
		t.Errorf("expected ErrWeakPassword, got %v", err)
	}
}

// failingTokenVersionStorage wraps UserStorage and forces IncrementTokenVersionTx
// to fail. All other methods delegate to the embedded storage.
type failingTokenVersionStorage struct {
	*sqlite.UserStorage
}

func (f *failingTokenVersionStorage) IncrementTokenVersionTx(_ *sql.Tx, _ string) error {
	return errInjectedFailure
}

var errInjectedFailure = errInjected("injected failure")

type errInjected string

func (e errInjected) Error() string { return string(e) }

// TestResetPassword_Atomic_RollbackOnFailure verifies that if the final write
// (IncrementTokenVersion) fails, the entire reset rolls back — the password is
// NOT changed and the token is NOT marked used.
func TestResetPassword_Atomic_RollbackOnFailure(t *testing.T) {
	db := setupTestDB(t)
	userStorage := sqlite.NewUserStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)
	user := createTestUser(t, authService, "atomic@test.com", "Atomic")

	// Capture original password hash so we can verify it's unchanged.
	originalUser, err := userStorage.GetByID(user.User.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	originalHash := originalUser.PasswordHash

	// Build the service with a storage that fails on IncrementTokenVersionTx.
	failing := &failingTokenVersionStorage{UserStorage: userStorage}
	sender := &captureSender{}
	prs := NewPasswordResetService(db, failing, sender)

	if err := prs.RequestReset(user.User.Email); err != nil {
		t.Fatalf("RequestReset: %v", err)
	}
	token := extractTokenFromEmail(sender.Calls()[0].Body)

	if err := prs.ResetPassword(token, "Newpassword123"); err == nil {
		t.Fatal("expected ResetPassword to fail due to injected error, got nil")
	}

	// Password must NOT have changed (transaction rolled back).
	after, err := userStorage.GetByID(user.User.ID)
	if err != nil {
		t.Fatalf("GetByID after: %v", err)
	}
	if after.PasswordHash != originalHash {
		t.Errorf("password hash changed despite rollback (was %q, now %q)", originalHash, after.PasswordHash)
	}
	if utils.CheckPassword("Newpassword123", after.PasswordHash) {
		t.Error("new password verifies — transaction did not roll back")
	}

	// Token must NOT be marked used (transaction rolled back), so a retry is
	// still possible once the underlying issue is resolved.
	prt, err := userStorage.GetPasswordResetToken(token)
	if err != nil {
		t.Fatalf("GetPasswordResetToken: %v", err)
	}
	if prt == nil {
		t.Fatal("token disappeared")
	}
	if prt.UsedAt != nil {
		t.Error("token marked used despite rollback")
	}
}

// Sanity check that the LogSender doesn't panic.
func TestLogSender_DoesNotError(t *testing.T) {
	s := email.NewLogSender("from@test")
	if err := s.Send("to@test", "subject", "body"); err != nil {
		t.Errorf("LogSender.Send returned error: %v", err)
	}
}
