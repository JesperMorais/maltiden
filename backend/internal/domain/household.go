package domain

import "time"

type Household struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type HouseholdMember struct {
	ID          string    `json:"id"`
	HouseholdID string    `json:"householdId"`
	UserID      string    `json:"userId"`
	Role        string    `json:"role"` // owner, member, guest
	JoinedAt    time.Time `json:"joinedAt"`
}

type HouseholdResponse struct {
	ID      string                    `json:"id"`
	Name    string                    `json:"name"`
	Members []HouseholdMemberResponse `json:"members"`
}

type HouseholdMemberResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// Invite codes

type InviteCode struct {
	ID          string    `json:"id"`
	HouseholdID string    `json:"householdId"`
	Code        string    `json:"code"`
	ExpiresAt   time.Time `json:"expiresAt"`
	UsedBy      *string   `json:"usedBy,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

type CreateInviteResponse struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type JoinHouseholdRequest struct {
	Code string `json:"code"`
}

type JoinHouseholdResponse struct {
	HouseholdID string `json:"householdId"`
}

// Member status

type MemberStatus struct {
	ID             string `json:"id"`
	IsEatingToday  bool   `json:"isEatingToday"`
	WantsLunchBox  bool   `json:"wantsLunchBox"`
}

type MemberStatusListResponse struct {
	Members []MemberStatus `json:"members"`
}

type UpdateMemberStatusRequest struct {
	IsEatingToday *bool `json:"isEatingToday,omitempty"`
	WantsLunchBox *bool `json:"wantsLunchBox,omitempty"`
}

type UpdateHouseholdRequest struct {
	Name string `json:"name"`
}

// Planning preferences

type HouseholdPreferences struct {
	DietProfile           string   `json:"dietProfile"`
	VegetarianDaysPerWeek int      `json:"vegetarianDaysPerWeek"`
	DislikedIngredients   []string `json:"dislikedIngredients"`
}
