package sqlite

import (
	"context"
	"database/sql"
	"maltiden/internal/domain"
	"time"
)

type UserStorage struct {
	db *sql.DB
}

func NewUserStorage(db *sql.DB) *UserStorage {
	return &UserStorage{db: db}
}

func (s *UserStorage) Create(user *domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO users (id, email, password_hash, name, household_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash,
		user.Name, user.HouseholdID, user.CreatedAt,
	)
	return err
}

// CreateTx inserts a user within a transaction.
func (s *UserStorage) CreateTx(tx *sql.Tx, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, name, household_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := tx.Exec(query,
		user.ID, user.Email, user.PasswordHash,
		user.Name, user.HouseholdID, user.CreatedAt,
	)
	return err
}

func (s *UserStorage) GetByEmail(email string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, email, password_hash, name, household_id, token_version, created_at
			  FROM users WHERE email = ?`

	var user domain.User
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash,
		&user.Name, &user.HouseholdID, &user.TokenVersion, &user.CreatedAt,
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, email, password_hash, name, household_id, token_version, created_at
			  FROM users WHERE id = ?`

	var user domain.User
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash,
		&user.Name, &user.HouseholdID, &user.TokenVersion, &user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // User does not exist
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserStorage) GetTokenVersion(userID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var version int
	err := s.db.QueryRowContext(ctx,
		`SELECT token_version FROM users WHERE id = ?`, userID,
	).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// GetAuthInfo returns the user's current token version together with the
// household they currently belong to. The household is resolved live from
// household_members (the authoritative source), NOT from the JWT claim or the
// users.household_id column — both of which go stale when a user joins a
// different household after their token was issued. Resolving it here keeps
// every household-scoped endpoint (menus, member statuses, invites) consistent
// with GET /households/me, which also reads household_members.
//
// householdID is empty if the user belongs to no household.
func (s *UserStorage) GetAuthInfo(userID string) (tokenVersion int, householdID string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var hh sql.NullString
	err = s.db.QueryRowContext(ctx,
		`SELECT u.token_version, hm.household_id
		 FROM users u
		 LEFT JOIN household_members hm ON hm.user_id = u.id
		 WHERE u.id = ?`,
		userID,
	).Scan(&tokenVersion, &hh)
	if err != nil {
		return 0, "", err
	}
	return tokenVersion, hh.String, nil
}

func (s *UserStorage) IncrementTokenVersion(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET token_version = token_version + 1 WHERE id = ?`, userID,
	)
	return err
}

// IncrementTokenVersionTx increments token_version within a transaction.
func (s *UserStorage) IncrementTokenVersionTx(tx *sql.Tx, userID string) error {
	_, err := tx.Exec(
		`UPDATE users SET token_version = token_version + 1 WHERE id = ?`, userID,
	)
	return err
}

// UpdatePassword updates the password hash for a given user.
func (s *UserStorage) UpdatePassword(userID, newHash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ? WHERE id = ?`, newHash, userID,
	)
	return err
}

// UpdatePasswordTx updates the password hash within a transaction.
func (s *UserStorage) UpdatePasswordTx(tx *sql.Tx, userID, newHash string) error {
	_, err := tx.Exec(
		`UPDATE users SET password_hash = ? WHERE id = ?`, newHash, userID,
	)
	return err
}

// CreatePasswordResetToken inserts a new password reset token.
func (s *UserStorage) CreatePasswordResetToken(token *domain.PasswordResetToken) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO password_reset_tokens (token, user_id, expires_at, created_at)
		 VALUES (?, ?, ?, ?)`,
		token.Token, token.UserID, token.ExpiresAt, token.CreatedAt,
	)
	return err
}

// GetPasswordResetToken retrieves a password reset token by its value.
// Returns (nil, nil) if the token does not exist.
func (s *UserStorage) GetPasswordResetToken(token string) (*domain.PasswordResetToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var t domain.PasswordResetToken
	var usedAt sql.NullTime
	err := s.db.QueryRowContext(ctx,
		`SELECT token, user_id, expires_at, used_at, created_at
		 FROM password_reset_tokens WHERE token = ?`, token,
	).Scan(&t.Token, &t.UserID, &t.ExpiresAt, &usedAt, &t.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if usedAt.Valid {
		t.UsedAt = &usedAt.Time
	}
	return &t, nil
}

// MarkPasswordResetTokenUsed marks a token as used at the current time.
func (s *UserStorage) MarkPasswordResetTokenUsed(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = ? WHERE token = ?`,
		time.Now(), token,
	)
	return err
}

// MarkPasswordResetTokenUsedTx marks a token as used within a transaction.
func (s *UserStorage) MarkPasswordResetTokenUsedTx(tx *sql.Tx, token string) error {
	_, err := tx.Exec(
		`UPDATE password_reset_tokens SET used_at = ? WHERE token = ?`,
		time.Now(), token,
	)
	return err
}
