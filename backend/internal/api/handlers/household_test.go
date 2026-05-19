package handlers

import (
	"database/sql"
	"encoding/json"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockHouseholdStorage struct {
	roleByUser      map[string]string
	createdInvites  []*domain.InviteCode
	createInviteErr error
}

func (m *mockHouseholdStorage) GetMemberRole(_ string, userID string) (string, error) {
	return m.roleByUser[userID], nil
}

func (m *mockHouseholdStorage) CreateInviteCode(invite *domain.InviteCode) error {
	if m.createInviteErr != nil {
		return m.createInviteErr
	}
	m.createdInvites = append(m.createdInvites, invite)
	return nil
}

func (m *mockHouseholdStorage) DB() *sql.DB { return nil }

func (m *mockHouseholdStorage) Create(_ *domain.Household) error                        { return nil }
func (m *mockHouseholdStorage) CreateTx(_ *sql.Tx, _ *domain.Household) error           { return nil }
func (m *mockHouseholdStorage) UpdateName(_, _ string) error                            { return nil }
func (m *mockHouseholdStorage) GetByUserID(_ string) (*domain.HouseholdResponse, error) { return nil, nil }
func (m *mockHouseholdStorage) GetInviteByCode(_ string) (*domain.InviteCode, error)    { return nil, nil }
func (m *mockHouseholdStorage) MarkInviteUsedTx(_ *sql.Tx, _, _ string) error           { return nil }
func (m *mockHouseholdStorage) GetMemberStatuses(_ string) ([]domain.MemberStatus, error) {
	return nil, nil
}
func (m *mockHouseholdStorage) UpdateMemberStatus(_, _ string, _ *bool, _ *bool) error { return nil }
func (m *mockHouseholdStorage) RemoveMember(_, _ string) error                         { return nil }
func (m *mockHouseholdStorage) IsMember(_, _ string) (bool, error)                     { return false, nil }
func (m *mockHouseholdStorage) IsMemberTx(_ *sql.Tx, _, _ string) (bool, error)        { return false, nil }
func (m *mockHouseholdStorage) GetUserHouseholdID(_ string) (string, error)             { return "", nil }
func (m *mockHouseholdStorage) RemoveMemberTx(_ *sql.Tx, _, _ string) error             { return nil }
func (m *mockHouseholdStorage) AddMemberTx(_ *sql.Tx, _ *domain.HouseholdMember) error { return nil }
func (m *mockHouseholdStorage) UpdateUserHouseholdTx(_ *sql.Tx, _, _ string) error      { return nil }

func newTestHouseholdHandler(store *mockHouseholdStorage) *HouseholdHandler {
	svc := services.NewHouseholdService(store, nil)
	return NewHouseholdHandler(svc)
}

func TestHouseholdHandler_CreateInvite_Owner_Returns201(t *testing.T) {
	userID := "usr_owner-1234-5678-9abc-def012345678"
	householdID := "hh_test-1234-5678-9abc-def012345678"

	store := &mockHouseholdStorage{
		roleByUser: map[string]string{userID: "owner"},
	}
	h := newTestHouseholdHandler(store)

	req := httptest.NewRequest("POST", "/households/invite", nil)
	req = setAuthContext(req, userID, householdID)

	rr := httptest.NewRecorder()
	h.CreateInvite(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp domain.CreateInviteResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Code) != 8 {
		t.Errorf("expected 8-char code, got %q (len %d)", resp.Code, len(resp.Code))
	}
	in8Days := time.Now().Add(8 * 24 * time.Hour)
	if resp.ExpiresAt.IsZero() || resp.ExpiresAt.After(in8Days) {
		t.Errorf("ExpiresAt %v is not within 8 days of now", resp.ExpiresAt)
	}
}

func TestHouseholdHandler_CreateInvite_Member_Returns201(t *testing.T) {
	userID := "usr_member-1234-5678-9abc-def012345678"
	householdID := "hh_test-1234-5678-9abc-def012345678"

	store := &mockHouseholdStorage{
		roleByUser: map[string]string{userID: "member"},
	}
	h := newTestHouseholdHandler(store)

	req := httptest.NewRequest("POST", "/households/invite", nil)
	req = setAuthContext(req, userID, householdID)

	rr := httptest.NewRecorder()
	h.CreateInvite(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	if len(store.createdInvites) != 1 {
		t.Fatalf("expected 1 invite stored, got %d", len(store.createdInvites))
	}
}

func TestHouseholdHandler_CreateInvite_Guest_Returns403(t *testing.T) {
	userID := "usr_guest-1234-5678-9abc-def012345678"
	householdID := "hh_test-1234-5678-9abc-def012345678"

	store := &mockHouseholdStorage{
		roleByUser: map[string]string{userID: "guest"},
	}
	h := newTestHouseholdHandler(store)

	req := httptest.NewRequest("POST", "/households/invite", nil)
	req = setAuthContext(req, userID, householdID)

	rr := httptest.NewRecorder()
	h.CreateInvite(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rr.Code, rr.Body.String())
	}
	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "forbidden" {
		t.Errorf("expected error 'forbidden', got %q", errResp["error"])
	}
	if len(store.createdInvites) != 0 {
		t.Errorf("expected 0 invites stored, got %d", len(store.createdInvites))
	}
}

func TestHouseholdHandler_CreateInvite_NoRole_Returns403(t *testing.T) {
	userID := "usr_norole-1234-5678-9abc-def012345678"
	householdID := "hh_test-1234-5678-9abc-def012345678"

	store := &mockHouseholdStorage{
		roleByUser: map[string]string{},
	}
	h := newTestHouseholdHandler(store)

	req := httptest.NewRequest("POST", "/households/invite", nil)
	req = setAuthContext(req, userID, householdID)

	rr := httptest.NewRecorder()
	h.CreateInvite(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rr.Code, rr.Body.String())
	}
	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "forbidden" {
		t.Errorf("expected error 'forbidden', got %q", errResp["error"])
	}
	if len(store.createdInvites) != 0 {
		t.Errorf("expected 0 invites stored, got %d", len(store.createdInvites))
	}
}
