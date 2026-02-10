package services

import (
	"maltiden/internal/domain"
	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/utils"
	"testing"
)

func TestRegister_WeakPassword_TooShort(t *testing.T) {
	db := setupTestDB(t)
	userStorage := sqlite.NewUserStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	_, err := authService.Register(domain.RegisterRequest{
		Email:    "test@example.com",
		Password: "Short1!",
		Name:     "Test",
	})
	if err != domain.ErrWeakPassword {
		t.Errorf("expected ErrWeakPassword, got: %v", err)
	}
}

func TestRegister_WeakPassword_LowDiversity(t *testing.T) {
	db := setupTestDB(t)
	userStorage := sqlite.NewUserStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	// Only lowercase + digits = 2 types (needs 3)
	_, err := authService.Register(domain.RegisterRequest{
		Email:    "test@example.com",
		Password: "testpassword123",
		Name:     "Test",
	})
	if err != domain.ErrWeakPassword {
		t.Errorf("expected ErrWeakPassword for low diversity, got: %v", err)
	}
}

func TestRegister_StrongPassword(t *testing.T) {
	db := setupTestDB(t)
	userStorage := sqlite.NewUserStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	// 3 types: uppercase + lowercase + digit
	resp, err := authService.Register(domain.RegisterRequest{
		Email:    "test@example.com",
		Password: "Testpassword123",
		Name:     "Test",
	})
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	db := setupTestDB(t)
	userStorage := sqlite.NewUserStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	_, err := authService.Register(domain.RegisterRequest{
		Email:    "not-an-email",
		Password: "Testpassword123",
		Name:     "Test",
	})
	if err != domain.ErrInvalidEmail {
		t.Errorf("expected ErrInvalidEmail, got: %v", err)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	userStorage := sqlite.NewUserStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	createTestUser(t, authService, "dup@example.com", "First")

	_, err := authService.Register(domain.RegisterRequest{
		Email:    "dup@example.com",
		Password: "Testpassword123",
		Name:     "Second",
	})
	if err != domain.ErrDuplicateEmail {
		t.Errorf("expected ErrDuplicateEmail, got: %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	db := setupTestDB(t)
	userStorage := sqlite.NewUserStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	createTestUser(t, authService, "login@example.com", "Login")

	resp, err := authService.Login(domain.LoginRequest{
		Email:    "login@example.com",
		Password: "Testpassword123",
	})
	if err != nil {
		t.Fatalf("expected login success, got: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	db := setupTestDB(t)
	userStorage := sqlite.NewUserStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	createTestUser(t, authService, "login@example.com", "Login")

	_, err := authService.Login(domain.LoginRequest{
		Email:    "login@example.com",
		Password: "WrongPassword1!",
	})
	if err != domain.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	db := setupTestDB(t)
	userStorage := sqlite.NewUserStorage(db)
	householdStorage := sqlite.NewHouseholdStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)

	_, err := authService.Login(domain.LoginRequest{
		Email:    "noone@example.com",
		Password: "Testpassword123",
	})
	if err != domain.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestJWT_ValidToken(t *testing.T) {
	jwtService := setupTestJWTService(t)

	token, err := jwtService.GenerateToken("usr_abc", "hh_def")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != "usr_abc" {
		t.Errorf("expected UserID usr_abc, got %s", claims.UserID)
	}
	if claims.HouseholdID != "hh_def" {
		t.Errorf("expected HouseholdID hh_def, got %s", claims.HouseholdID)
	}
}

func TestJWT_WrongSecret(t *testing.T) {
	jwtService1 := setupTestJWTService(t)
	jwtService2, _ := utils.NewJWTService("different-secret-key-1234567890abcdef")

	token, _ := jwtService1.GenerateToken("usr_abc", "hh_def")

	_, err := jwtService2.ValidateToken(token)
	if err == nil {
		t.Error("expected error validating token with wrong secret")
	}
}

func TestJWT_MalformedToken(t *testing.T) {
	jwtService := setupTestJWTService(t)

	_, err := jwtService.ValidateToken("not.a.valid.jwt")
	if err == nil {
		t.Error("expected error for malformed token")
	}
}

// IDOR: User from household A cannot update member status in household B
func TestIDOR_CrossHouseholdStatusUpdate(t *testing.T) {
	db := setupTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)
	householdService := NewHouseholdService(householdStorage, userStorage)

	// Create two separate households
	anna := createTestUser(t, authService, "anna@test.com", "Anna")
	erik := createTestUser(t, authService, "erik@test.com", "Erik")

	// Erik tries to update status in Anna's household
	trueVal := true
	err := householdService.UpdateMemberStatus(anna.User.HouseholdID, erik.User.ID, domain.UpdateMemberStatusRequest{
		IsEatingToday: &trueVal,
	})
	if err == nil {
		t.Error("expected error for cross-household status update")
	}
}

// IDOR: User from household A cannot remove members from household B
func TestIDOR_CrossHouseholdRemoveMember(t *testing.T) {
	db := setupTestDB(t)
	householdStorage := sqlite.NewHouseholdStorage(db)
	userStorage := sqlite.NewUserStorage(db)
	jwtService := setupTestJWTService(t)
	authService := NewAuthService(db, userStorage, householdStorage, jwtService)
	householdService := NewHouseholdService(householdStorage, userStorage)

	// Create two separate households
	anna := createTestUser(t, authService, "anna@test.com", "Anna")
	erik := createTestUser(t, authService, "erik@test.com", "Erik")

	// Erik tries to remove Anna from her own household
	err := householdService.RemoveMember(anna.User.HouseholdID, erik.User.ID, anna.User.ID)
	if err == nil {
		t.Error("expected error for cross-household member removal")
	}
}
