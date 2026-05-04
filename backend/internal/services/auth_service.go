package services

import (
	"database/sql"
	"maltiden/internal/domain"
	"maltiden/pkg/utils"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AuthService struct {
	db               *sql.DB
	userStorage      domain.UserRepository
	householdStorage domain.HouseholdRepository
	jwtService       *utils.JWTService
}

func NewAuthService(db *sql.DB, userStorage domain.UserRepository, householdStorage domain.HouseholdRepository, jwtService *utils.JWTService) *AuthService {
	return &AuthService{
		db:               db,
		userStorage:      userStorage,
		householdStorage: householdStorage,
		jwtService:       jwtService,
	}
}

func (s *AuthService) Register(req domain.RegisterRequest) (*domain.AuthResponse, error) {
	// Trim whitespace from inputs (VALID-15)
	req.Email = strings.TrimSpace(req.Email)
	req.Name = strings.TrimSpace(req.Name)
	req.LastName = strings.TrimSpace(req.LastName)

	// Email format validation using net/mail
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, domain.ErrInvalidEmail
	}

	// Validate password strength: min 8 chars, at least 3 of 4 character types
	if err := ValidatePasswordStrength(req.Password); err != nil {
		return nil, err
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

	// Create household with provided name or default.
	// Prefer last name for the default ("Anderssons hushåll" reads more naturally
	// than "Annas hushåll" for a shared space). Fall back to first name if no
	// last name was given. Swedish genitive: names ending in s/x/z don't take an extra -s.
	householdName := req.HouseholdName
	if householdName == "" {
		base := req.LastName
		if base == "" {
			base = req.Name
		}
		householdName = base + genitiveSuffix(base) + " hushåll"
	}
	household := &domain.Household{
		ID:        householdID,
		Name:      householdName,
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

	// Generate JWT (new user starts at token_version 1)
	token, err := s.jwtService.GenerateToken(user.ID, user.HouseholdID, 1)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// genitiveSuffix returns the Swedish genitive "s" suffix for a name, or an empty
// string for names ending in s/x/z (which don't take an extra -s).
func genitiveSuffix(name string) string {
	lower := strings.ToLower(name)
	if strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") || strings.HasSuffix(lower, "z") {
		return ""
	}
	return "s"
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

	// Generate JWT with current token_version
	token, err := s.jwtService.GenerateToken(user.ID, user.HouseholdID, user.TokenVersion)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}
