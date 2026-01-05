package sqlite

import (
	"database/sql"
	"maltiden/internal/domain"
)

type HouseholdStorage struct {
	db *sql.DB
}

func NewHouseholdStorage(db *sql.DB) *HouseholdStorage {
	return &HouseholdStorage{db: db}
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
