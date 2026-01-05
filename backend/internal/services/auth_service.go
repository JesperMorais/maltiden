package services

import (
	"errors"
	"maltiden/internal/domain"
	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/utils"
	"time"

	"github.com/google/uuid"
)

type AuthService struct {
	userStorage      *sqlite.UserStorage
	householdStorage *sqlite.HouseholdStorage
}

func NewAuthService(userStorage *sqlite.UserStorage, householdStorage *sqlite.HouseholdStorage) *AuthService {
	return &AuthService{
		userStorage:      userStorage,
		householdStorage: householdStorage,
	}
}

func (s *AuthService) Register(req domain.RegisterRequest) (*domain.AuthResponse, error) {
	// Validate input
	if len(req.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	// Check if email already exists
	existing, err := s.userStorage.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Generate IDs
	userID := "usr_" + uuid.New().String()
	householdID := "hh_" + uuid.New().String()
	now := time.Now()

	// Create household
	household := &domain.Household{
		ID:        householdID,
		Name:      req.Name + "'s household",
		CreatedAt: now,
	}
	if err := s.householdStorage.Create(household); err != nil {
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
	if err := s.userStorage.Create(user); err != nil {
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
	if err := s.householdStorage.AddMember(member); err != nil {
		return nil, err
	}

	// Generate JWT
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *AuthService) Login(req domain.LoginRequest) (*domain.AuthResponse, error) {
	// Get user from DB
	user, err := s.userStorage.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Validate password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	// Generate JWT
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}
