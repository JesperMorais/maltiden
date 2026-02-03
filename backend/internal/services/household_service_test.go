package services

import (
	"database/sql"
	"maltiden/internal/domain"
	"maltiden/internal/storage/sqlite"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Use temp file for SQLite (in-memory doesn't work well with multiple connections)
	tmpFile := t.TempDir() + "/test.db"

	// Need to set working directory so migrations can be found
	// The migrations are read from "migrations/" relative path
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

func getBackendRoot(t *testing.T) string {
	t.Helper()
	// Walk up from internal/services to backend/
	// This file is at backend/internal/services/household_service_test.go
	// Backend root is 2 levels up
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// We're in internal/services, go up two levels
	return dir + "/../.."
}

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

func TestCreateInvite(t *testing.T) {
	db := setupTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	authService := NewAuthService(userStorage, householdStorage)
	householdService := NewHouseholdService(householdStorage, userStorage)

	// Create a user (which creates a household)
	user := createTestUser(t, authService, "anna@test.com", "Anna")

	// Create invite
	resp, err := householdService.CreateInvite(user.User.HouseholdID)
	if err != nil {
		t.Fatalf("CreateInvite failed: %v", err)
	}

	if resp.Code == "" {
		t.Error("expected non-empty invite code")
	}
	if len(resp.Code) != 8 {
		t.Errorf("expected 8-char code, got %d chars: %q", len(resp.Code), resp.Code)
	}
	if resp.ExpiresAt.Before(time.Now()) {
		t.Error("expected expiry to be in the future")
	}
	if resp.ExpiresAt.After(time.Now().Add(8 * 24 * time.Hour)) {
		t.Error("expected expiry to be within ~7 days")
	}
}

func TestJoinHousehold(t *testing.T) {
	db := setupTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	authService := NewAuthService(userStorage, householdStorage)
	householdService := NewHouseholdService(householdStorage, userStorage)

	// Create owner
	owner := createTestUser(t, authService, "anna@test.com", "Anna")

	// Create invite code
	invite, err := householdService.CreateInvite(owner.User.HouseholdID)
	if err != nil {
		t.Fatalf("CreateInvite failed: %v", err)
	}

	// Create second user
	joiner := createTestUser(t, authService, "erik@test.com", "Erik")

	// Join household
	resp, err := householdService.JoinHousehold(joiner.User.ID, domain.JoinHouseholdRequest{
		Code: invite.Code,
	})
	if err != nil {
		t.Fatalf("JoinHousehold failed: %v", err)
	}

	if resp.HouseholdID != owner.User.HouseholdID {
		t.Errorf("expected household %s, got %s", owner.User.HouseholdID, resp.HouseholdID)
	}

	// Verify the joiner is now a member
	household, err := householdService.GetMyHousehold(joiner.User.ID)
	if err != nil {
		t.Fatalf("GetMyHousehold failed: %v", err)
	}

	// After joining, the user should only be in the owner's household (single-household enforcement)
	if household.ID != owner.User.HouseholdID {
		t.Errorf("expected joiner's household to be %s, got %s", owner.User.HouseholdID, household.ID)
	}

	// Verify joiner was removed from their original household
	isMember, err := householdStorage.IsMember(joiner.User.HouseholdID, joiner.User.ID)
	if err != nil {
		t.Fatalf("IsMember check failed: %v", err)
	}
	if isMember {
		t.Error("expected joiner to be removed from their original household")
	}
}

func TestJoinHousehold_InvalidCode(t *testing.T) {
	db := setupTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	authService := NewAuthService(userStorage, householdStorage)
	householdService := NewHouseholdService(householdStorage, userStorage)

	user := createTestUser(t, authService, "anna@test.com", "Anna")

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

func TestJoinHousehold_ExpiredCode(t *testing.T) {
	db := setupTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	authService := NewAuthService(userStorage, householdStorage)
	householdService := NewHouseholdService(householdStorage, userStorage)

	owner := createTestUser(t, authService, "anna@test.com", "Anna")

	// Manually create an expired invite code
	expiredInvite := &domain.InviteCode{
		ID:          "inv_" + uuid.New().String(),
		HouseholdID: owner.User.HouseholdID,
		Code:        "EXPIRD",
		ExpiresAt:   time.Now().Add(-1 * time.Hour), // expired 1 hour ago
		CreatedAt:   time.Now().Add(-8 * 24 * time.Hour),
	}
	if err := householdStorage.CreateInviteCode(expiredInvite); err != nil {
		t.Fatalf("failed to create expired invite: %v", err)
	}

	joiner := createTestUser(t, authService, "erik@test.com", "Erik")

	_, err := householdService.JoinHousehold(joiner.User.ID, domain.JoinHouseholdRequest{
		Code: "EXPIRD",
	})
	if err == nil {
		t.Error("expected error for expired code")
	}
	if err.Error() != "invalid_code" {
		t.Errorf("expected invalid_code error, got: %v", err)
	}
}

func TestJoinHousehold_AlreadyMember(t *testing.T) {
	db := setupTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	authService := NewAuthService(userStorage, householdStorage)
	householdService := NewHouseholdService(householdStorage, userStorage)

	owner := createTestUser(t, authService, "anna@test.com", "Anna")

	// Owner tries to join their own household
	invite, _ := householdService.CreateInvite(owner.User.HouseholdID)
	_, err := householdService.JoinHousehold(owner.User.ID, domain.JoinHouseholdRequest{
		Code: invite.Code,
	})
	if err == nil {
		t.Error("expected error when already a member")
	}
	if err.Error() != "already_member" {
		t.Errorf("expected already_member error, got: %v", err)
	}
}

func TestGetMemberStatuses(t *testing.T) {
	db := setupTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	authService := NewAuthService(userStorage, householdStorage)
	householdService := NewHouseholdService(householdStorage, userStorage)

	owner := createTestUser(t, authService, "anna@test.com", "Anna")

	resp, err := householdService.GetMemberStatuses(owner.User.HouseholdID)
	if err != nil {
		t.Fatalf("GetMemberStatuses failed: %v", err)
	}

	if len(resp.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(resp.Members))
	}

	// Default: eating today = true, wants lunch box = false
	if !resp.Members[0].IsEatingToday {
		t.Error("expected isEatingToday to be true by default")
	}
	if resp.Members[0].WantsLunchBox {
		t.Error("expected wantsLunchBox to be false by default")
	}
}

func TestUpdateMemberStatus(t *testing.T) {
	db := setupTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	authService := NewAuthService(userStorage, householdStorage)
	householdService := NewHouseholdService(householdStorage, userStorage)

	owner := createTestUser(t, authService, "anna@test.com", "Anna")

	// Update eating status
	falseVal := false
	trueVal := true
	err := householdService.UpdateMemberStatus(owner.User.HouseholdID, owner.User.ID, domain.UpdateMemberStatusRequest{
		IsEatingToday: &falseVal,
		WantsLunchBox: &trueVal,
	})
	if err != nil {
		t.Fatalf("UpdateMemberStatus failed: %v", err)
	}

	// Verify
	resp, _ := householdService.GetMemberStatuses(owner.User.HouseholdID)
	if resp.Members[0].IsEatingToday {
		t.Error("expected isEatingToday to be false after update")
	}
	if !resp.Members[0].WantsLunchBox {
		t.Error("expected wantsLunchBox to be true after update")
	}
}

func TestUpdateMemberStatus_PartialUpdate(t *testing.T) {
	db := setupTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	authService := NewAuthService(userStorage, householdStorage)
	householdService := NewHouseholdService(householdStorage, userStorage)

	owner := createTestUser(t, authService, "anna@test.com", "Anna")

	// Only update lunch box
	trueVal := true
	err := householdService.UpdateMemberStatus(owner.User.HouseholdID, owner.User.ID, domain.UpdateMemberStatusRequest{
		WantsLunchBox: &trueVal,
	})
	if err != nil {
		t.Fatalf("UpdateMemberStatus failed: %v", err)
	}

	// Verify eating status unchanged (should still be true = default)
	resp, _ := householdService.GetMemberStatuses(owner.User.HouseholdID)
	if !resp.Members[0].IsEatingToday {
		t.Error("expected isEatingToday to remain true (unchanged)")
	}
	if !resp.Members[0].WantsLunchBox {
		t.Error("expected wantsLunchBox to be true after update")
	}
}

func TestUpdateMemberStatus_NotMember(t *testing.T) {
	db := setupTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	authService := NewAuthService(userStorage, householdStorage)
	householdService := NewHouseholdService(householdStorage, userStorage)

	owner := createTestUser(t, authService, "anna@test.com", "Anna")

	trueVal := true
	err := householdService.UpdateMemberStatus(owner.User.HouseholdID, "usr_nonexistent", domain.UpdateMemberStatusRequest{
		IsEatingToday: &trueVal,
	})
	if err == nil {
		t.Error("expected error for non-member")
	}
	if err.Error() != "not_found" {
		t.Errorf("expected not_found error, got: %v", err)
	}
}

func TestRemoveMember(t *testing.T) {
	db := setupTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	authService := NewAuthService(userStorage, householdStorage)
	householdService := NewHouseholdService(householdStorage, userStorage)

	// Create owner and a member
	owner := createTestUser(t, authService, "anna@test.com", "Anna")
	invite, _ := householdService.CreateInvite(owner.User.HouseholdID)
	member := createTestUser(t, authService, "erik@test.com", "Erik")
	householdService.JoinHousehold(member.User.ID, domain.JoinHouseholdRequest{Code: invite.Code})

	// Owner removes member
	err := householdService.RemoveMember(owner.User.HouseholdID, owner.User.ID, member.User.ID)
	if err != nil {
		t.Fatalf("RemoveMember failed: %v", err)
	}

	// Verify member is gone
	isMember, _ := householdStorage.IsMember(owner.User.HouseholdID, member.User.ID)
	if isMember {
		t.Error("expected member to be removed")
	}
}

func TestRemoveMember_CannotRemoveSelf(t *testing.T) {
	db := setupTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	authService := NewAuthService(userStorage, householdStorage)
	householdService := NewHouseholdService(householdStorage, userStorage)

	owner := createTestUser(t, authService, "anna@test.com", "Anna")

	err := householdService.RemoveMember(owner.User.HouseholdID, owner.User.ID, owner.User.ID)
	if err == nil {
		t.Error("expected error when removing self")
	}
	if err.Error() != "cannot_remove" {
		t.Errorf("expected cannot_remove error, got: %v", err)
	}
}

func TestRemoveMember_CannotRemoveOwner(t *testing.T) {
	db := setupTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	authService := NewAuthService(userStorage, householdStorage)
	householdService := NewHouseholdService(householdStorage, userStorage)

	owner := createTestUser(t, authService, "anna@test.com", "Anna")
	invite, _ := householdService.CreateInvite(owner.User.HouseholdID)
	member := createTestUser(t, authService, "erik@test.com", "Erik")
	householdService.JoinHousehold(member.User.ID, domain.JoinHouseholdRequest{Code: invite.Code})

	// Member tries to remove owner
	err := householdService.RemoveMember(owner.User.HouseholdID, member.User.ID, owner.User.ID)
	if err == nil {
		t.Error("expected error when member tries to remove owner")
	}
	if err.Error() != "cannot_remove" {
		t.Errorf("expected cannot_remove error, got: %v", err)
	}
}

func TestRemoveMember_GuestCannotRemove(t *testing.T) {
	db := setupTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	authService := NewAuthService(userStorage, householdStorage)
	householdService := NewHouseholdService(householdStorage, userStorage)

	owner := createTestUser(t, authService, "anna@test.com", "Anna")

	// Add a guest and a member manually
	invite1, _ := householdService.CreateInvite(owner.User.HouseholdID)
	guest := createTestUser(t, authService, "lisa@test.com", "Lisa")
	householdService.JoinHousehold(guest.User.ID, domain.JoinHouseholdRequest{Code: invite1.Code})

	// Change guest's role to "guest"
	db.Exec(`UPDATE household_members SET role = 'guest' WHERE user_id = ?`, guest.User.ID)

	invite2, _ := householdService.CreateInvite(owner.User.HouseholdID)
	member := createTestUser(t, authService, "erik@test.com", "Erik")
	householdService.JoinHousehold(member.User.ID, domain.JoinHouseholdRequest{Code: invite2.Code})

	// Guest tries to remove member
	err := householdService.RemoveMember(owner.User.HouseholdID, guest.User.ID, member.User.ID)
	if err == nil {
		t.Error("expected error when guest tries to remove")
	}
	if err.Error() != "forbidden" {
		t.Errorf("expected forbidden error, got: %v", err)
	}
}

func TestFullFlow_InviteJoinAndManage(t *testing.T) {
	db := setupTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	authService := NewAuthService(userStorage, householdStorage)
	householdService := NewHouseholdService(householdStorage, userStorage)

	// 1. Anna registers (creates household)
	anna := createTestUser(t, authService, "anna@test.com", "Anna")

	// 2. Anna creates invite
	invite, err := householdService.CreateInvite(anna.User.HouseholdID)
	if err != nil {
		t.Fatalf("CreateInvite failed: %v", err)
	}

	// 3. Erik joins
	erik := createTestUser(t, authService, "erik@test.com", "Erik")
	_, err = householdService.JoinHousehold(erik.User.ID, domain.JoinHouseholdRequest{Code: invite.Code})
	if err != nil {
		t.Fatalf("Erik JoinHousehold failed: %v", err)
	}

	// 4. Check member statuses (should be 2 members now, both eating by default)
	statuses, err := householdService.GetMemberStatuses(anna.User.HouseholdID)
	if err != nil {
		t.Fatalf("GetMemberStatuses failed: %v", err)
	}
	if len(statuses.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(statuses.Members))
	}

	// 5. Erik marks himself as not eating today
	falseVal := false
	err = householdService.UpdateMemberStatus(anna.User.HouseholdID, erik.User.ID, domain.UpdateMemberStatusRequest{
		IsEatingToday: &falseVal,
	})
	if err != nil {
		t.Fatalf("UpdateMemberStatus failed: %v", err)
	}

	// 6. Verify Erik's status changed
	statuses, _ = householdService.GetMemberStatuses(anna.User.HouseholdID)
	for _, m := range statuses.Members {
		if m.ID == erik.User.ID && m.IsEatingToday {
			t.Error("expected Erik's isEatingToday to be false")
		}
	}

	// 7. Anna removes Erik
	err = householdService.RemoveMember(anna.User.HouseholdID, anna.User.ID, erik.User.ID)
	if err != nil {
		t.Fatalf("RemoveMember failed: %v", err)
	}

	// 8. Verify only Anna remains
	statuses, _ = householdService.GetMemberStatuses(anna.User.HouseholdID)
	if len(statuses.Members) != 1 {
		t.Errorf("expected 1 member after removal, got %d", len(statuses.Members))
	}
}
