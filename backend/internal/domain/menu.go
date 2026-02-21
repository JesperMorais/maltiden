package domain

import "time"

type MenuDay struct {
	Date     string `json:"date"`
	RecipeID string `json:"recipeId,omitempty"`
	Servings int    `json:"servings"`
	Skip     bool   `json:"skip,omitempty"`
}

type Menu struct {
	ID          string    `json:"id"`
	HouseholdID string    `json:"householdId"`
	Days        []MenuDay `json:"days"`
	CreatedAt   time.Time `json:"createdAt"`
}

type GenerateMenuRequest struct {
	Days          int            `json:"days"`
	Servings      int            `json:"servings"`
	SkipDays      []string       `json:"skipDays"`
	ExtraPortions map[string]int `json:"extraPortions"`
}

type UpdateMenuRequest struct {
	Days []MenuDay `json:"days"`
}

type MenuResponse struct {
	ID   string    `json:"id"`
	Days []MenuDay `json:"days"`
}
