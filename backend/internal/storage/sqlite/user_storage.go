package sqlite

import (
	"database/sql"
	"maltiden/internal/domain"
)

type UserStorage struct {
	db *sql.DB
}

func NewUserStorage(db *sql.DB) *UserStorage {
	return &UserStorage{db: db}
}

func (s *UserStorage) Create(user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, name, household_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query,
		user.ID, user.Email, user.PasswordHash,
		user.Name, user.HouseholdID, user.CreatedAt,
	)
	return err
}

func (s *UserStorage) GetByEmail(email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, name, household_id, created_at
			  FROM users WHERE email = ?`

	var user domain.User
	err := s.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash,
		&user.Name, &user.HouseholdID, &user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // User does not exist
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserStorage) GetByID(id string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, name, household_id, created_at
			  FROM users WHERE id = ?`

	var user domain.User
	err := s.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash,
		&user.Name, &user.HouseholdID, &user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // User does not exist
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}
