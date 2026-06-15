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
	GetTokenVersion(userID string) (int, error)
	IncrementTokenVersion(userID string) error
	IncrementTokenVersionTx(tx *sql.Tx, userID string) error
	UpdatePassword(userID, newHash string) error
	UpdatePasswordTx(tx *sql.Tx, userID, newHash string) error
	CreatePasswordResetToken(token *PasswordResetToken) error
	GetPasswordResetToken(token string) (*PasswordResetToken, error)
	MarkPasswordResetTokenUsed(token string) error
	MarkPasswordResetTokenUsedTx(tx *sql.Tx, token string) error
}

// HouseholdRepository defines the interface for household storage operations.
type HouseholdRepository interface {
	Create(household *Household) error
	CreateTx(tx *sql.Tx, household *Household) error
	UpdateName(householdID, name string) error
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
	GetAll(filter *RecipeFilter, householdID string) ([]RecipeSummary, error)
	GetAllPaginated(filter *RecipeFilter, householdID string, limit, offset int) ([]RecipeSummary, int, error)
	GetByID(id string) (*Recipe, error)
	GetByIDs(ids []string) (map[string]*Recipe, error)
	Create(recipe *Recipe) error
	Update(recipe *Recipe) error
	Delete(id string) error
}

// MenuRepository defines the interface for menu storage operations.
type MenuRepository interface {
	Create(menu *Menu) error
	Update(menu *Menu) error
	GetCurrentByHousehold(householdID string) (*Menu, error)
	GetByID(id string) (*Menu, error)
	GetHouseholdIDByMenuID(menuID string) (string, error)
	// GetRecentRecipeIDs returns the set of recipe ids used across the
	// household's windowMenus most recent menus, for the recency penalty.
	GetRecentRecipeIDs(householdID string, windowMenus int) (map[string]bool, error)
}

// MenuPreferencesRepository defines storage operations for per-household
// menu-generation preferences.
type MenuPreferencesRepository interface {
	// Get returns the preferences for a household, or nil if none have been
	// saved (callers should fall back to DefaultMenuPreferences).
	Get(householdID string) (*MenuPreferences, error)
	// Upsert creates or replaces the preferences for a household.
	Upsert(prefs *MenuPreferences) error
}

// ShoppingRepository defines the interface for shopping storage operations.
type ShoppingRepository interface {
	GetCheckedItems(menuID string) (map[string]bool, error)
	SetChecked(menuID, itemID string, checked bool) error
	CreateCustomItem(item *CustomShoppingItem) error
	CreateCustomItemWithCap(item *CustomShoppingItem, maxItems int) (bool, error)
	DeleteCustomItem(id, householdID string) error
	GetCustomItems(menuID, householdID string) ([]CustomShoppingItem, error)
	SetCustomItemChecked(id, householdID string, checked bool) error
}
