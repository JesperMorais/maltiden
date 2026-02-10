package services

import (
	"crypto/rand"
	"fmt"
	"maltiden/internal/domain"
	"math/big"
	"time"

	"github.com/google/uuid"
)

type HouseholdService struct {
	householdStorage domain.HouseholdRepository
	userStorage      domain.UserRepository
}

func NewHouseholdService(householdStorage domain.HouseholdRepository, userStorage domain.UserRepository) *HouseholdService {
	return &HouseholdService{
		householdStorage: householdStorage,
		userStorage:      userStorage,
	}
}

func (s *HouseholdService) GetMyHousehold(userID string) (*domain.HouseholdResponse, error) {
	return s.householdStorage.GetByUserID(userID)
}

func (s *HouseholdService) GetMemberRole(householdID, userID string) (string, error) {
	return s.householdStorage.GetMemberRole(householdID, userID)
}

// CreateInvite generates an 8-character invite code valid for 7 days.
func (s *HouseholdService) CreateInvite(householdID string) (*domain.CreateInviteResponse, error) {
	code, err := generateInviteCode()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	invite := &domain.InviteCode{
		ID:          "inv_" + uuid.New().String(),
		HouseholdID: householdID,
		Code:        code,
		ExpiresAt:   now.Add(7 * 24 * time.Hour),
		CreatedAt:   now,
	}

	if err := s.householdStorage.CreateInviteCode(invite); err != nil {
		return nil, err
	}

	return &domain.CreateInviteResponse{
		Code:      invite.Code,
		ExpiresAt: invite.ExpiresAt,
	}, nil
}

// JoinHousehold validates an invite code and adds the user to the household.
// The entire operation runs inside a database transaction to prevent race conditions.
// Users can only belong to one household — joining a new one removes them from the old one.
func (s *HouseholdService) JoinHousehold(userID string, req domain.JoinHouseholdRequest) (*domain.JoinHouseholdResponse, error) {
	if req.Code == "" {
		return nil, domain.ErrCodeRequired
	}

	// Validate invite code before starting the transaction
	invite, err := s.householdStorage.GetInviteByCode(req.Code)
	if err != nil {
		return nil, err
	}
	if invite == nil {
		return nil, domain.ErrInvalidCode
	}
	if time.Now().After(invite.ExpiresAt) {
		return nil, domain.ErrInvalidCode
	}
	if invite.UsedBy != nil {
		return nil, domain.ErrInvalidCode
	}

	// Begin transaction for the mutating operations
	tx, err := s.householdStorage.DB().Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Re-check membership inside transaction to prevent race conditions
	isMember, err := s.householdStorage.IsMemberTx(tx, invite.HouseholdID, userID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, domain.ErrAlreadyMember
	}

	// Remove user from their current household (single-household enforcement)
	currentHouseholdID, err := s.householdStorage.GetUserHouseholdID(userID)
	if err != nil {
		return nil, err
	}
	if currentHouseholdID != "" {
		if err := s.householdStorage.RemoveMemberTx(tx, currentHouseholdID, userID); err != nil {
			return nil, err
		}
	}

	// Add user as member of new household
	member := &domain.HouseholdMember{
		ID:          "hm_" + uuid.New().String(),
		HouseholdID: invite.HouseholdID,
		UserID:      userID,
		Role:        "member",
		JoinedAt:    time.Now(),
	}
	if err := s.householdStorage.AddMemberTx(tx, member); err != nil {
		return nil, err
	}

	// Update user's household_id
	if err := s.householdStorage.UpdateUserHouseholdTx(tx, userID, invite.HouseholdID); err != nil {
		return nil, err
	}

	// Mark invite as used
	if err := s.householdStorage.MarkInviteUsedTx(tx, invite.ID, userID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &domain.JoinHouseholdResponse{
		HouseholdID: invite.HouseholdID,
	}, nil
}

// GetMemberStatuses returns the eating/lunch-box status of all household members.
func (s *HouseholdService) GetMemberStatuses(householdID string) (*domain.MemberStatusListResponse, error) {
	statuses, err := s.householdStorage.GetMemberStatuses(householdID)
	if err != nil {
		return nil, err
	}

	if statuses == nil {
		statuses = []domain.MemberStatus{}
	}

	return &domain.MemberStatusListResponse{
		Members: statuses,
	}, nil
}

// UpdateMemberStatus updates the eating/lunch-box status of a specific member.
func (s *HouseholdService) UpdateMemberStatus(householdID, memberID string, req domain.UpdateMemberStatusRequest) error {
	// Verify the target is a member of this household
	isMember, err := s.householdStorage.IsMember(householdID, memberID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrNotFound
	}

	return s.householdStorage.UpdateMemberStatus(householdID, memberID, req.IsEatingToday, req.WantsLunchBox)
}

// RemoveMember removes a member from the household. Owners cannot be removed, and
// only owners/members can remove others.
func (s *HouseholdService) RemoveMember(householdID, requestingUserID, targetUserID string) error {
	// Can't remove yourself
	if requestingUserID == targetUserID {
		return domain.ErrCannotRemove
	}

	// Check requesting user's role
	requestingRole, err := s.householdStorage.GetMemberRole(householdID, requestingUserID)
	if err != nil {
		return err
	}
	if requestingRole == "" || requestingRole == "guest" {
		return domain.ErrForbidden
	}

	// Check target's role — can't remove the owner
	targetRole, err := s.householdStorage.GetMemberRole(householdID, targetUserID)
	if err != nil {
		return err
	}
	if targetRole == "" {
		return domain.ErrNotFound
	}
	if targetRole == "owner" {
		return domain.ErrCannotRemove
	}

	return s.householdStorage.RemoveMember(householdID, targetUserID)
}

// generateInviteCode creates a random 8-character alphanumeric code (~40 bits of entropy).
func generateInviteCode() (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no 0/O/1/I to avoid confusion
	code := make([]byte, 8)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("generating invite code: %w", err)
		}
		code[i] = charset[n.Int64()]
	}
	return string(code), nil
}
