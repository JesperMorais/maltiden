package sqlite

import (
	"testing"
	"time"

	"maltiden/internal/domain"
)

func setupHouseholdTestDB(t *testing.T) (*HouseholdStorage, *UserStorage) {
	t.Helper()
	db, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewHouseholdStorage(db), NewUserStorage(db)
}

func newTestHousehold(id, name string) *domain.Household {
	return &domain.Household{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
}

func newTestMember(id, hhID, userID, role string) *domain.HouseholdMember {
	return &domain.HouseholdMember{
		ID:          id,
		HouseholdID: hhID,
		UserID:      userID,
		Role:        role,
		JoinedAt:    time.Now().UTC().Truncate(time.Second),
	}
}

func seedUserForHH(t *testing.T, us *UserStorage, id, email, hhID string) {
	t.Helper()
	u := &domain.User{
		ID:           id,
		Email:        email,
		PasswordHash: "$2a$12$testhash",
		Name:         "Test User",
		HouseholdID:  hhID,
		CreatedAt:    time.Now().UTC().Truncate(time.Second),
	}
	if err := us.Create(u); err != nil {
		t.Fatalf("seed user %s: %v", id, err)
	}
}

func TestHouseholdStorage_Create(t *testing.T) {
	hs, _ := setupHouseholdTestDB(t)
	hh := newTestHousehold("hh_create_1", "Test Household")

	if err := hs.Create(hh); err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestHouseholdStorage_UpdateName(t *testing.T) {
	hs, _ := setupHouseholdTestDB(t)
	hh := newTestHousehold("hh_upd_1", "Original Name")
	if err := hs.Create(hh); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := hs.UpdateName(hh.ID, "New Name"); err != nil {
		t.Fatalf("UpdateName: %v", err)
	}
}

func TestHouseholdStorage_AddMember_GetByUserID(t *testing.T) {
	hs, us := setupHouseholdTestDB(t)
	hh := newTestHousehold("hh_mem_1", "Member HH")
	if err := hs.Create(hh); err != nil {
		t.Fatalf("Create household: %v", err)
	}

	seedUserForHH(t, us, "usr_mem_1", "member1@example.com", hh.ID)
	if err := hs.AddMember(newTestMember("hm_1", hh.ID, "usr_mem_1", "owner")); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	resp, err := hs.GetByUserID("usr_mem_1")
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if resp == nil {
		t.Fatal("expected household response, got nil")
	}
	if resp.ID != hh.ID {
		t.Errorf("household id mismatch: got %q, want %q", resp.ID, hh.ID)
	}
	if len(resp.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(resp.Members))
	}
	if resp.Members[0].Role != "owner" {
		t.Errorf("role mismatch: got %q, want %q", resp.Members[0].Role, "owner")
	}
}

func TestHouseholdStorage_GetByUserID_Missing(t *testing.T) {
	hs, _ := setupHouseholdTestDB(t)

	resp, err := hs.GetByUserID("usr_does_not_exist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil, got %+v", resp)
	}
}

func TestHouseholdStorage_IsMember(t *testing.T) {
	hs, us := setupHouseholdTestDB(t)
	hh := newTestHousehold("hh_ismem_1", "IsMember HH")
	if err := hs.Create(hh); err != nil {
		t.Fatalf("Create household: %v", err)
	}

	seedUserForHH(t, us, "usr_ismem_1", "ismem@example.com", hh.ID)
	if err := hs.AddMember(newTestMember("hm_ismem_1", hh.ID, "usr_ismem_1", "member")); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	ok, err := hs.IsMember(hh.ID, "usr_ismem_1")
	if err != nil {
		t.Fatalf("IsMember: %v", err)
	}
	if !ok {
		t.Error("expected true, got false")
	}

	ok, err = hs.IsMember(hh.ID, "usr_nobody")
	if err != nil {
		t.Fatalf("IsMember not-member: %v", err)
	}
	if ok {
		t.Error("expected false for non-member, got true")
	}
}

func TestHouseholdStorage_GetMemberRole(t *testing.T) {
	hs, us := setupHouseholdTestDB(t)
	hh := newTestHousehold("hh_role_1", "Role HH")
	if err := hs.Create(hh); err != nil {
		t.Fatalf("Create household: %v", err)
	}

	seedUserForHH(t, us, "usr_role_1", "role@example.com", hh.ID)
	if err := hs.AddMember(newTestMember("hm_role_1", hh.ID, "usr_role_1", "owner")); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	role, err := hs.GetMemberRole(hh.ID, "usr_role_1")
	if err != nil {
		t.Fatalf("GetMemberRole: %v", err)
	}
	if role != "owner" {
		t.Errorf("role mismatch: got %q, want %q", role, "owner")
	}

	role, err = hs.GetMemberRole(hh.ID, "usr_nobody")
	if err != nil {
		t.Fatalf("GetMemberRole missing: %v", err)
	}
	if role != "" {
		t.Errorf("expected empty role for non-member, got %q", role)
	}
}

func TestHouseholdStorage_RemoveMember(t *testing.T) {
	hs, us := setupHouseholdTestDB(t)
	hh := newTestHousehold("hh_rm_1", "Remove Member HH")
	if err := hs.Create(hh); err != nil {
		t.Fatalf("Create household: %v", err)
	}

	seedUserForHH(t, us, "usr_rm_1", "rm@example.com", hh.ID)
	if err := hs.AddMember(newTestMember("hm_rm_1", hh.ID, "usr_rm_1", "member")); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	if err := hs.RemoveMember(hh.ID, "usr_rm_1"); err != nil {
		t.Fatalf("RemoveMember: %v", err)
	}

	ok, err := hs.IsMember(hh.ID, "usr_rm_1")
	if err != nil {
		t.Fatalf("IsMember after remove: %v", err)
	}
	if ok {
		t.Error("expected false after removal, got true")
	}
}

func TestHouseholdStorage_InviteCodes(t *testing.T) {
	hs, _ := setupHouseholdTestDB(t)
	hh := newTestHousehold("hh_inv_1", "Invite HH")
	if err := hs.Create(hh); err != nil {
		t.Fatalf("Create household: %v", err)
	}

	invite := &domain.InviteCode{
		ID:          "inv_1",
		HouseholdID: hh.ID,
		Code:        "TESTCODE123",
		ExpiresAt:   time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second),
		CreatedAt:   time.Now().UTC().Truncate(time.Second),
	}
	if err := hs.CreateInviteCode(invite); err != nil {
		t.Fatalf("CreateInviteCode: %v", err)
	}

	got, err := hs.GetInviteByCode("TESTCODE123")
	if err != nil {
		t.Fatalf("GetInviteByCode: %v", err)
	}
	if got == nil {
		t.Fatal("expected invite, got nil")
	}
	if got.Code != invite.Code || got.HouseholdID != hh.ID {
		t.Errorf("invite mismatch: got %+v", got)
	}
	if got.UsedBy != nil {
		t.Error("expected UsedBy nil on fresh invite")
	}
}

func TestHouseholdStorage_GetInviteByCode_Missing(t *testing.T) {
	hs, _ := setupHouseholdTestDB(t)

	got, err := hs.GetInviteByCode("DOESNOTEXIST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestHouseholdStorage_MemberStatus(t *testing.T) {
	hs, us := setupHouseholdTestDB(t)
	hh := newTestHousehold("hh_status_1", "Status HH")
	if err := hs.Create(hh); err != nil {
		t.Fatalf("Create household: %v", err)
	}

	seedUserForHH(t, us, "usr_stat_1", "stat@example.com", hh.ID)
	if err := hs.AddMember(newTestMember("hm_stat_1", hh.ID, "usr_stat_1", "member")); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	statuses, err := hs.GetMemberStatuses(hh.ID)
	if err != nil {
		t.Fatalf("GetMemberStatuses: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}

	eating := true
	lunch := false
	if err := hs.UpdateMemberStatus(hh.ID, "usr_stat_1", &eating, &lunch); err != nil {
		t.Fatalf("UpdateMemberStatus: %v", err)
	}

	statuses, err = hs.GetMemberStatuses(hh.ID)
	if err != nil {
		t.Fatalf("GetMemberStatuses after update: %v", err)
	}
	if !statuses[0].IsEatingToday {
		t.Error("expected IsEatingToday=true")
	}
	if statuses[0].WantsLunchBox {
		t.Error("expected WantsLunchBox=false")
	}
}

func TestHouseholdStorage_UpdateMemberStatus_Partial(t *testing.T) {
	hs, us := setupHouseholdTestDB(t)
	hh := newTestHousehold("hh_partial_1", "Partial Status HH")
	if err := hs.Create(hh); err != nil {
		t.Fatalf("Create household: %v", err)
	}

	seedUserForHH(t, us, "usr_partial_1", "partial@example.com", hh.ID)
	if err := hs.AddMember(newTestMember("hm_partial_1", hh.ID, "usr_partial_1", "member")); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	lunch := true
	if err := hs.UpdateMemberStatus(hh.ID, "usr_partial_1", nil, &lunch); err != nil {
		t.Fatalf("UpdateMemberStatus lunch only: %v", err)
	}

	statuses, err := hs.GetMemberStatuses(hh.ID)
	if err != nil {
		t.Fatalf("GetMemberStatuses: %v", err)
	}
	if !statuses[0].WantsLunchBox {
		t.Error("expected WantsLunchBox=true")
	}
}

func TestHouseholdStorage_GetUserHouseholdID(t *testing.T) {
	hs, us := setupHouseholdTestDB(t)
	hh := newTestHousehold("hh_uid_1", "UserHH")
	if err := hs.Create(hh); err != nil {
		t.Fatalf("Create household: %v", err)
	}

	seedUserForHH(t, us, "usr_uid_1", "uid@example.com", hh.ID)

	got, err := hs.GetUserHouseholdID("usr_uid_1")
	if err != nil {
		t.Fatalf("GetUserHouseholdID: %v", err)
	}
	if got != hh.ID {
		t.Errorf("household id mismatch: got %q, want %q", got, hh.ID)
	}
}

func TestHouseholdStorage_UpdateUserHousehold(t *testing.T) {
	hs, us := setupHouseholdTestDB(t)
	hh1 := newTestHousehold("hh_uuh_1", "HH One")
	hh2 := newTestHousehold("hh_uuh_2", "HH Two")
	if err := hs.Create(hh1); err != nil {
		t.Fatalf("Create hh1: %v", err)
	}
	if err := hs.Create(hh2); err != nil {
		t.Fatalf("Create hh2: %v", err)
	}

	seedUserForHH(t, us, "usr_uuh_1", "uuh@example.com", hh1.ID)
	if err := hs.UpdateUserHousehold("usr_uuh_1", hh2.ID); err != nil {
		t.Fatalf("UpdateUserHousehold: %v", err)
	}

	got, err := hs.GetUserHouseholdID("usr_uuh_1")
	if err != nil {
		t.Fatalf("GetUserHouseholdID after update: %v", err)
	}
	if got != hh2.ID {
		t.Errorf("expected %q, got %q", hh2.ID, got)
	}
}

func TestHouseholdStorage_MultipleMembers(t *testing.T) {
	hs, us := setupHouseholdTestDB(t)
	hh := newTestHousehold("hh_multi_1", "Multi Member HH")
	if err := hs.Create(hh); err != nil {
		t.Fatalf("Create household: %v", err)
	}

	for i, spec := range []struct{ id, email, role string }{
		{"usr_multi_1", "m1@example.com", "owner"},
		{"usr_multi_2", "m2@example.com", "member"},
	} {
		seedUserForHH(t, us, spec.id, spec.email, hh.ID)
		mem := newTestMember("hm_multi_"+string(rune('1'+i)), hh.ID, spec.id, spec.role)
		if err := hs.AddMember(mem); err != nil {
			t.Fatalf("AddMember %s: %v", spec.id, err)
		}
	}

	resp, err := hs.GetByUserID("usr_multi_1")
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if len(resp.Members) != 2 {
		t.Errorf("expected 2 members, got %d", len(resp.Members))
	}
}
