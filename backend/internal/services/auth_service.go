package services

import (
	"database/sql"
	"maltiden/internal/domain"
	"maltiden/pkg/utils"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AuthService struct {
	db               *sql.DB
	userStorage      domain.UserRepository
	householdStorage domain.HouseholdRepository
}

func NewAuthService(db *sql.DB, userStorage domain.UserRepository, householdStorage domain.HouseholdRepository) *AuthService {
	return &AuthService{
		db:               db,
		userStorage:      userStorage,
		householdStorage: householdStorage,
	}
}

func (s *AuthService) Register(req domain.RegisterRequest) (*domain.AuthResponse, error) {
	// Trim whitespace from inputs (VALID-15)
	req.Email = strings.TrimSpace(req.Email)
	req.Name = strings.TrimSpace(req.Name)

	// Basic email format validation
	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		return nil, domain.ErrInvalidEmail
	}

	// Validate input
	if len(req.Password) < 8 {
		return nil, domain.ErrWeakPassword
	}

	// Check if email already exists
	existing, err := s.userStorage.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrDuplicateEmail
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Generate IDs
	userID := "usr_" + uuid.New().String()
	householdID := "hh_" + uuid.New().String()
	now := time.Now()

	// Begin transaction for atomic registration
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create household
	household := &domain.Household{
		ID:        householdID,
		Name:      req.Name + "'s household",
		CreatedAt: now,
	}
	if err := s.householdStorage.CreateTx(tx, household); err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		ID:           userID,
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		HouseholdID:  householdID,
		CreatedAt:    now,
	}
	if err := s.userStorage.CreateTx(tx, user); err != nil {
		return nil, err
	}

	// Add user as household owner
	member := &domain.HouseholdMember{
		ID:          "hm_" + uuid.New().String(),
		HouseholdID: householdID,
		UserID:      userID,
		Role:        "owner",
		JoinedAt:    now,
	}
	if err := s.householdStorage.AddMemberTx(tx, member); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Generate JWT
	token, err := utils.GenerateToken(user.ID, user.HouseholdID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *AuthService) Login(req domain.LoginRequest) (*domain.AuthResponse, error) {
	// Trim whitespace from email (VALID-15)
	req.Email = strings.TrimSpace(req.Email)

	// Get user from DB
	user, err := s.userStorage.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Validate password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	// Generate JWT
	token, err := utils.GenerateToken(user.ID, user.HouseholdID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}
