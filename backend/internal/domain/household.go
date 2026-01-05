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
