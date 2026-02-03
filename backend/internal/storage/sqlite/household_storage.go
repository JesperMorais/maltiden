package sqlite

import (
	"database/sql"
	"maltiden/internal/domain"
	"strings"
)

type HouseholdStorage struct {
	db *sql.DB
}

func NewHouseholdStorage(db *sql.DB) *HouseholdStorage {
	return &HouseholdStorage{db: db}
}

func (s *HouseholdStorage) DB() *sql.DB {
	return s.db
}

func (s *HouseholdStorage) Create(household *domain.Household) error {
	query := `
		INSERT INTO households (id, name, created_at)
		VALUES (?, ?, ?)
	`
	_, err := s.db.Exec(query, household.ID, household.Name, household.CreatedAt)
	return err
}

func (s *HouseholdStorage) AddMember(member *domain.HouseholdMember) error {
	query := `
		INSERT INTO household_members (id, household_id, user_id, role, joined_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query,
		member.ID, member.HouseholdID, member.UserID, member.Role, member.JoinedAt,
	)
	return err
}

// AddMemberTx inserts a household member within a transaction.
func (s *HouseholdStorage) AddMemberTx(tx *sql.Tx, member *domain.HouseholdMember) error {
	query := `
		INSERT INTO household_members (id, household_id, user_id, role, joined_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := tx.Exec(query,
		member.ID, member.HouseholdID, member.UserID, member.Role, member.JoinedAt,
	)
	return err
}

func (s *HouseholdStorage) GetByUserID(userID string) (*domain.HouseholdResponse, error) {
	// Get household
	householdQuery := `
		SELECT h.id, h.name
		FROM households h
		JOIN household_members hm ON h.id = hm.household_id
		WHERE hm.user_id = ?
		LIMIT 1
	`

	var household domain.HouseholdResponse
	err := s.db.QueryRow(householdQuery, userID).Scan(&household.ID, &household.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Get members
	membersQuery := `
		SELECT u.id, u.name, hm.role
		FROM users u
		JOIN household_members hm ON u.id = hm.user_id
		WHERE hm.household_id = ?
		ORDER BY hm.joined_at ASC
	`

	rows, err := s.db.Query(membersQuery, household.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	household.Members = []domain.HouseholdMemberResponse{}
	for rows.Next() {
		var member domain.HouseholdMemberResponse
		if err := rows.Scan(&member.ID, &member.Name, &member.Role); err != nil {
			return nil, err
		}
		household.Members = append(household.Members, member)
	}

	return &household, nil
}

// Invite code methods

func (s *HouseholdStorage) CreateInviteCode(invite *domain.InviteCode) error {
	query := `
		INSERT INTO invite_codes (id, household_id, code, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query,
		invite.ID, invite.HouseholdID, invite.Code, invite.ExpiresAt, invite.CreatedAt,
	)
	return err
}

func (s *HouseholdStorage) GetInviteByCode(code string) (*domain.InviteCode, error) {
	query := `
		SELECT id, household_id, code, expires_at, used_by, created_at
		FROM invite_codes
		WHERE code = ?
	`

	var invite domain.InviteCode
	var usedBy sql.NullString
	err := s.db.QueryRow(query, code).Scan(
		&invite.ID, &invite.HouseholdID, &invite.Code,
		&invite.ExpiresAt, &usedBy, &invite.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if usedBy.Valid {
		invite.UsedBy = &usedBy.String
	}

	return &invite, nil
}

func (s *HouseholdStorage) MarkInviteUsedTx(tx *sql.Tx, codeID, userID string) error {
	_, err := tx.Exec(
		`UPDATE invite_codes SET used_by = ? WHERE id = ?`,
		userID, codeID,
	)
	return err
}

// Member status methods

func (s *HouseholdStorage) GetMemberStatuses(householdID string) ([]domain.MemberStatus, error) {
	query := `
		SELECT hm.user_id, hm.is_eating_today, hm.wants_lunch_box
		FROM household_members hm
		WHERE hm.household_id = ?
		ORDER BY hm.joined_at ASC
	`

	rows, err := s.db.Query(query, householdID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []domain.MemberStatus
	for rows.Next() {
		var ms domain.MemberStatus
		var eating, lunch int
		if err := rows.Scan(&ms.ID, &eating, &lunch); err != nil {
			return nil, err
		}
		ms.IsEatingToday = eating == 1
		ms.WantsLunchBox = lunch == 1
		statuses = append(statuses, ms)
	}

	return statuses, rows.Err()
}

// UpdateMemberStatus performs a single atomic UPDATE for both fields.
func (s *HouseholdStorage) UpdateMemberStatus(householdID, userID string, isEatingToday *bool, wantsLunchBox *bool) error {
	updates := []string{}
	args := []interface{}{}

	if isEatingToday != nil {
		updates = append(updates, "is_eating_today = ?")
		if *isEatingToday {
			args = append(args, 1)
		} else {
			args = append(args, 0)
		}
	}
	if wantsLunchBox != nil {
		updates = append(updates, "wants_lunch_box = ?")
		if *wantsLunchBox {
			args = append(args, 1)
		} else {
			args = append(args, 0)
		}
	}

	if len(updates) == 0 {
		return nil
	}

	args = append(args, householdID, userID)
	query := "UPDATE household_members SET " + strings.Join(updates, ", ") + " WHERE household_id = ? AND user_id = ?"
	_, err := s.db.Exec(query, args...)
	return err
}

// Member management

func (s *HouseholdStorage) GetMemberRole(householdID, userID string) (string, error) {
	var role string
	err := s.db.QueryRow(
		`SELECT role FROM household_members WHERE household_id = ? AND user_id = ?`,
		householdID, userID,
	).Scan(&role)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return role, err
}

func (s *HouseholdStorage) RemoveMember(householdID, userID string) error {
	_, err := s.db.Exec(
		`DELETE FROM household_members WHERE household_id = ? AND user_id = ?`,
		householdID, userID,
	)
	return err
}

// RemoveMemberTx removes a household member within a transaction.
func (s *HouseholdStorage) RemoveMemberTx(tx *sql.Tx, householdID, userID string) error {
	_, err := tx.Exec(
		`DELETE FROM household_members WHERE household_id = ? AND user_id = ?`,
		householdID, userID,
	)
	return err
}

func (s *HouseholdStorage) IsMember(householdID, userID string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM household_members WHERE household_id = ? AND user_id = ?`,
		householdID, userID,
	).Scan(&count)
	return count > 0, err
}

// IsMemberTx checks membership within a transaction (row-level read).
func (s *HouseholdStorage) IsMemberTx(tx *sql.Tx, householdID, userID string) (bool, error) {
	var count int
	err := tx.QueryRow(
		`SELECT COUNT(*) FROM household_members WHERE household_id = ? AND user_id = ?`,
		householdID, userID,
	).Scan(&count)
	return count > 0, err
}

func (s *HouseholdStorage) UpdateUserHousehold(userID, householdID string) error {
	_, err := s.db.Exec(
		`UPDATE users SET household_id = ? WHERE id = ?`,
		householdID, userID,
	)
	return err
}

// UpdateUserHouseholdTx updates the user's household_id within a transaction.
func (s *HouseholdStorage) UpdateUserHouseholdTx(tx *sql.Tx, userID, householdID string) error {
	_, err := tx.Exec(
		`UPDATE users SET household_id = ? WHERE id = ?`,
		householdID, userID,
	)
	return err
}

// GetUserHouseholdID returns the user's current household_id.
func (s *HouseholdStorage) GetUserHouseholdID(userID string) (string, error) {
	var householdID sql.NullString
	err := s.db.QueryRow(
		`SELECT household_id FROM users WHERE id = ?`, userID,
	).Scan(&householdID)
	if err != nil {
		return "", err
	}
	if householdID.Valid {
		return householdID.String, nil
	}
	return "", nil
}
