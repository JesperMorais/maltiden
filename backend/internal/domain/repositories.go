package domain

import (
	"database/sql"
)

// UserRepository defines the interface for user storage operations.
type UserRepository interface {
	GetByEmail(email string) (*User, error)
	GetByID(id string) (*User, error)
	Create(user *User) error
	CreateTx(tx *sql.Tx, user *User) error
}

// HouseholdRepository defines the interface for household storage operations.
type HouseholdRepository interface {
	Create(household *Household) error
	CreateTx(tx *sql.Tx, household *Household) error
	GetByUserID(userID string) (*HouseholdResponse, error)
	CreateInviteCode(invite *InviteCode) error
	GetInviteByCode(code string) (*InviteCode, error)
	MarkInviteUsedTx(tx *sql.Tx, codeID, userID string) error
	GetMemberRole(householdID, userID string) (string, error)
	GetMemberStatuses(householdID string) ([]MemberStatus, error)
	UpdateMemberStatus(householdID, userID string, isEatingToday *bool, wantsLunchBox *bool) error
	RemoveMember(householdID, userID string) error
	IsMember(householdID, userID string) (bool, error)
	IsMemberTx(tx *sql.Tx, householdID, userID string) (bool, error)
	GetUserHouseholdID(userID string) (string, error)
	RemoveMemberTx(tx *sql.Tx, householdID, userID string) error
	AddMemberTx(tx *sql.Tx, member *HouseholdMember) error
	UpdateUserHouseholdTx(tx *sql.Tx, userID, householdID string) error
	DB() *sql.DB
}

// RecipeRepository defines the interface for recipe storage operations.
type RecipeRepository interface {
	GetAll(filter *RecipeFilter) ([]RecipeSummary, error)
	GetAllPaginated(filter *RecipeFilter, limit, offset int) ([]RecipeSummary, int, error)
	GetByID(id string) (*Recipe, error)
	GetByIDs(ids []string) (map[string]*Recipe, error)
	Create(recipe *Recipe) error
}

// MenuRepository defines the interface for menu storage operations.
type MenuRepository interface {
	Create(menu *Menu) error
	GetCurrentByHousehold(householdID string) (*Menu, error)
	GetByID(id string) (*Menu, error)
	GetHouseholdIDByMenuID(menuID string) (string, error)
}

// ShoppingRepository defines the interface for shopping storage operations.
type ShoppingRepository interface {
	GetCheckedItems(menuID string) (map[string]bool, error)
	SetChecked(menuID, itemID string, checked bool) error
}
