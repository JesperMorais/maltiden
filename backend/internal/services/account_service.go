package services

import (
	"database/sql"
	"errors"

	"maltiden/internal/domain"
	"maltiden/pkg/utils"
)

// accountStore is satisfied by *sqlite.AccountStorage.
type accountStore interface {
	DeleteUser(userID string) error
	ExportUser(userID string) (map[string]any, error)
}

// AccountService backs the self-service privacy endpoints: delete my account
// and export my data.
type AccountService struct {
	accounts accountStore
	users    domain.UserRepository
}

func NewAccountService(accounts accountStore, users domain.UserRepository) *AccountService {
	return &AccountService{accounts: accounts, users: users}
}

// DeleteAccount erases the user after re-checking their password, so a stolen
// or forgotten-open session alone cannot wipe an account.
func (s *AccountService) DeleteAccount(userID, password string) error {
	user, err := s.users.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrNotFound
	}
	if !utils.CheckPassword(password, user.PasswordHash) {
		return domain.ErrInvalidCredentials
	}
	if err := s.accounts.DeleteUser(userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	return nil
}

// Export returns the user's data as a JSON-ready document.
func (s *AccountService) Export(userID string) (map[string]any, error) {
	data, err := s.accounts.ExportUser(userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return data, err
}
