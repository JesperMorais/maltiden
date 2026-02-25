package domain

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	HouseholdID  string    `json:"householdId"`
	TokenVersion int       `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}
