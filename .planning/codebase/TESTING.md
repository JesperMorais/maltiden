# Testing Patterns

**Analysis Date:** 2026-02-06

## Overview

Testing exists **only on the Go backend**. The frontend has **zero test files** -- no test runner, no test framework, no test configuration. Frontend quality is maintained via TypeScript strict mode, ESLint, and type-checking in CI.

---

## Test Framework (Backend - Go)

**Runner:**
- Go standard `testing` package
- No third-party test framework (no testify, no gomock)

**Assertion Library:**
- Go standard `testing.T` methods only: `t.Error()`, `t.Errorf()`, `t.Fatalf()`, `t.Fatal()`
- No assertion helpers or custom matchers

**Run Commands:**
```bash
cd backend && go test ./... -v -race    # Run all tests with race detection
cd backend && go test ./internal/services/ -v -run TestCreateInvite  # Run specific test
```

**CI Configuration:**
- Defined in `.github/workflows/ci.yml`
- Runs `go test ./... -v -race` with `JWT_SECRET=ci-test` env var
- Also runs `go build ./...`, `go vet ./...`, and `govulncheck`

---

## Test File Organization

**Location:**
- Co-located with source files in the same package directory
- Only one test file exists: `backend/internal/services/household_service_test.go`

**Naming:**
- Standard Go convention: `{source_file}_test.go`
- The test file for `household_service.go` is `household_service_test.go`

**Coverage gaps:**
- No tests for: auth service, recipe service, menu service, shopping service, tjek service
- No tests for: any handler, any storage layer, any middleware, any utility

---

## Test Structure

**Test Function Naming:**
- `Test{FunctionName}` for happy path: `TestCreateInvite`, `TestJoinHousehold`
- `Test{FunctionName}_{Scenario}` for edge cases: `TestJoinHousehold_InvalidCode`, `TestJoinHousehold_ExpiredCode`, `TestJoinHousehold_AlreadyMember`
- `TestFullFlow_{Description}` for integration-style tests: `TestFullFlow_InviteJoinAndManage`

**Suite Organization:**
```go
func TestCreateInvite(t *testing.T) {
    // 1. Setup: create DB and service dependencies
    db := setupTestDB(t)
    householdStorage := sqlite.NewHouseholdStorage(db)
    userStorage := sqlite.NewUserStorage(db)
    authService := NewAuthService(userStorage, householdStorage)
    householdService := NewHouseholdService(householdStorage, userStorage)

    // 2. Arrange: create test data
    user := createTestUser(t, authService, "anna@test.com", "Anna")

    // 3. Act: perform the operation under test
    resp, err := householdService.CreateInvite(user.User.HouseholdID)

    // 4. Assert: check results
    if err != nil {
        t.Fatalf("CreateInvite failed: %v", err)
    }
    if resp.Code == "" {
        t.Error("expected non-empty invite code")
    }
    if len(resp.Code) != 8 {
        t.Errorf("expected 8-char code, got %d chars: %q", len(resp.Code), resp.Code)
    }
}
```
See: `backend/internal/services/household_service_test.go` lines 61-89

**Patterns:**
- Every test creates its own isolated DB (no shared state between tests)
- Setup is repeated per test (no `TestMain` or test suites)
- `t.Helper()` used on helper functions
- `t.Cleanup()` used for DB cleanup (not `defer` in tests)
- `t.Fatalf()` for setup failures (stops test immediately)
- `t.Errorf()` for assertion failures (continues running)
- `t.Error()` for simple boolean assertions

---

## Test Helpers

**Database Setup:**
```go
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()

    // Use temp file for SQLite (in-memory doesn't work well with multiple connections)
    tmpFile := t.TempDir() + "/test.db"

    // Need to set working directory so migrations can be found
    origDir, _ := os.Getwd()
    os.Chdir(getBackendRoot(t))
    defer os.Chdir(origDir)

    db, err := sqlite.Open(tmpFile)
    if err != nil {
        t.Fatalf("failed to open test db: %v", err)
    }

    t.Cleanup(func() { db.Close() })
    return db
}
```
See: `backend/internal/services/household_service_test.go` lines 14-33

Important notes:
- Uses `t.TempDir()` for automatic cleanup
- Changes working directory to find migration files (fragile pattern)
- Runs real migrations against the test DB

**Test User Creation:**
```go
func createTestUser(t *testing.T, authService *AuthService, email, name string) *domain.AuthResponse {
    t.Helper()
    resp, err := authService.Register(domain.RegisterRequest{
        Email:    email,
        Password: "testpassword123",
        Name:     name,
    })
    if err != nil {
        t.Fatalf("failed to create test user %s: %v", name, err)
    }
    return resp
}
```
See: `backend/internal/services/household_service_test.go` lines 48-59

---

## Mocking

**Framework:** None used

**Approach:** Integration-style testing with real SQLite database
- Tests create a real SQLite database in a temp directory
- Tests run real migrations
- Tests exercise the full service → storage → DB path
- No mocking of storage layers or external dependencies

**What gets mocked:** Nothing. All tests are integration tests against real SQLite.

**Implications:**
- Tests require migration files to be accessible (working directory must be `backend/`)
- Tests are slower than unit tests but catch more bugs
- No way to test error paths from the database layer

---

## Fixtures and Factories

**Test Data:**
- Created inline in each test using `createTestUser` helper
- Common test credentials: `email: "anna@test.com"`, `password: "testpassword123"`
- Second user pattern: `email: "erik@test.com"`, `name: "Erik"`
- Expired invites created manually with custom `ExpiresAt` in the past

**Location:**
- All helpers and factories are in the same test file: `backend/internal/services/household_service_test.go`
- No shared test utilities or fixture files

---

## Coverage

**Requirements:** None enforced. No coverage thresholds configured.

**Current state:** Only household service has tests. All other services, handlers, storage, middleware, and utilities are untested.

**View coverage:**
```bash
cd backend && go test ./... -coverprofile=coverage.out
cd backend && go tool cover -html=coverage.out
```

---

## Test Types

### Unit Tests
- **Not present.** No isolated unit tests with mocked dependencies.

### Integration Tests (Backend Only)
- **Scope:** Service layer with real SQLite database
- **File:** `backend/internal/services/household_service_test.go`
- **Tests present (13 total):**
  - `TestCreateInvite` -- Generate invite code
  - `TestJoinHousehold` -- Happy path join
  - `TestJoinHousehold_InvalidCode` -- Invalid code rejection
  - `TestJoinHousehold_ExpiredCode` -- Expired code rejection
  - `TestJoinHousehold_AlreadyMember` -- Duplicate join rejection
  - `TestGetMemberStatuses` -- Default status values
  - `TestUpdateMemberStatus` -- Full update
  - `TestUpdateMemberStatus_PartialUpdate` -- Partial update
  - `TestUpdateMemberStatus_NotMember` -- Non-member rejection
  - `TestRemoveMember` -- Happy path removal
  - `TestRemoveMember_CannotRemoveSelf` -- Self-removal rejection
  - `TestRemoveMember_CannotRemoveOwner` -- Owner removal rejection
  - `TestRemoveMember_GuestCannotRemove` -- Guest permission rejection
  - `TestFullFlow_InviteJoinAndManage` -- Full invite/join/manage lifecycle

### E2E Tests
- **Not present.** No Cypress, Playwright, or similar framework.

### Frontend Tests
- **Not present.** No Vitest, Jest, or any test runner configured.
- No test-related dependencies in `frontend/package.json`
- `frontend/tsconfig.app.json` excludes `src/**/__tests__/*` (suggests intent for future tests but none exist)

---

## Common Patterns

### Error Case Testing
Test that specific error strings are returned:
```go
func TestJoinHousehold_InvalidCode(t *testing.T) {
    // ... setup ...

    _, err := householdService.JoinHousehold(user.User.ID, domain.JoinHouseholdRequest{
        Code: "BADCODE",
    })
    if err == nil {
        t.Error("expected error for invalid code")
    }
    if err.Error() != "invalid_code" {
        t.Errorf("expected invalid_code error, got: %v", err)
    }
}
```
See: `backend/internal/services/household_service_test.go` lines 143-161

### Multi-Step Flow Testing
Full workflow tests exercise multiple operations in sequence:
```go
func TestFullFlow_InviteJoinAndManage(t *testing.T) {
    // 1. Anna registers (creates household)
    anna := createTestUser(t, authService, "anna@test.com", "Anna")

    // 2. Anna creates invite
    invite, err := householdService.CreateInvite(anna.User.HouseholdID)

    // 3. Erik joins
    erik := createTestUser(t, authService, "erik@test.com", "Erik")
    _, err = householdService.JoinHousehold(erik.User.ID, domain.JoinHouseholdRequest{Code: invite.Code})

    // 4. Check member statuses
    // 5. Erik marks himself as not eating
    // 6. Verify Erik's status changed
    // 7. Anna removes Erik
    // 8. Verify only Anna remains
}
```
See: `backend/internal/services/household_service_test.go` lines 422-482

### Side Effect Verification
After a mutation, query the database to verify the state changed:
```go
// Verify member is gone
isMember, _ := householdStorage.IsMember(owner.User.HouseholdID, member.User.ID)
if isMember {
    t.Error("expected member to be removed")
}
```
See: `backend/internal/services/household_service_test.go` lines 345-348

### Direct DB Manipulation in Tests
Some tests bypass the service layer to set up specific conditions:
```go
// Change guest's role to "guest" directly in DB
db.Exec(`UPDATE household_members SET role = 'guest' WHERE user_id = ?`, guest.User.ID)
```
See: `backend/internal/services/household_service_test.go` line 406

---

## CI Test Pipeline

Defined in `.github/workflows/ci.yml`:

**Backend job:**
1. `go build ./...` -- compilation check
2. `go test ./... -v -race` -- tests with race detector (env: `JWT_SECRET=ci-test`)
3. `go vet ./...` -- static analysis
4. `govulncheck` -- vulnerability scanning

**Frontend job:**
1. `npm ci` -- install dependencies
2. `npm run type-check` -- `vue-tsc --build` (TypeScript compilation)
3. `npm run lint` -- ESLint
4. `npm run build` -- Vite production build

---

## When Adding New Tests

**For new Go service tests:**
1. Create `{service_name}_test.go` in `backend/internal/services/`
2. Use `setupTestDB(t)` helper (currently in `household_service_test.go` -- consider extracting to a shared test helper file)
3. Follow `Test{FunctionName}` / `Test{FunctionName}_{Scenario}` naming
4. Use `t.Fatalf()` for setup failures, `t.Errorf()` for assertions
5. Create test data with `createTestUser` helper or similar

**For new frontend tests (not yet established):**
- No patterns exist yet. Would need to add Vitest and configure it.
- The `tsconfig.app.json` already excludes `__tests__/` directories, suggesting co-located test directories were planned.

---

*Testing analysis: 2026-02-06*
*Update when test patterns change*
