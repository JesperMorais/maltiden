package domain

import "time"

type Ingredient struct {
	Name           string  `json:"name"`
	Amount         float64 `json:"amount"`
	Unit           string  `json:"unit"`
	CanonicalName  string  `json:"canonicalName,omitempty"`
	GramsEquiv     float64 `json:"gramsEquiv,omitempty"`
	IsPantryStaple bool    `json:"isPantryStaple,omitempty"`
	IsPerishable   bool    `json:"isPerishable,omitempty"`
}

type Recipe struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Servings     int          `json:"servings"`
	Emoji        string       `json:"emoji,omitempty"`
	Tags         []string     `json:"tags"`
	Ingredients  []Ingredient `json:"ingredients"`
	Instructions []string     `json:"instructions"`
	HouseholdID  string       `json:"householdId,omitempty"`
	CreatedAt    time.Time    `json:"createdAt"`
	MainProtein  string       `json:"mainProtein,omitempty"`
	DietClass    string       `json:"dietClass,omitempty"`
	Batchable    bool         `json:"batchable,omitempty"`
	CookMinutes  int          `json:"cookMinutes,omitempty"`
}

type RecipeSummary struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Servings int      `json:"servings"`
	Emoji    string   `json:"emoji,omitempty"`
	Tags     []string `json:"tags"`
}

type CreateRecipeRequest struct {
	Name         string       `json:"name"`
	Servings     int          `json:"servings"`
	Emoji        string       `json:"emoji,omitempty"`
	Tags         []string     `json:"tags"`
	Ingredients  []Ingredient `json:"ingredients"`
	Instructions []string     `json:"instructions"`
	MainProtein  string       `json:"mainProtein,omitempty"`
	DietClass    string       `json:"dietClass,omitempty"`
	Batchable    bool         `json:"batchable,omitempty"`
	CookMinutes  int          `json:"cookMinutes,omitempty"`
}

type UpdateRecipeRequest struct {
	Name         string       `json:"name"`
	Servings     int          `json:"servings"`
	Emoji        string       `json:"emoji,omitempty"`
	Tags         []string     `json:"tags"`
	Ingredients  []Ingredient `json:"ingredients"`
	Instructions []string     `json:"instructions"`
	MainProtein  string       `json:"mainProtein,omitempty"`
	DietClass    string       `json:"dietClass,omitempty"`
	Batchable    bool         `json:"batchable,omitempty"`
	CookMinutes  int          `json:"cookMinutes,omitempty"`
}

type RecipesResponse struct {
	Recipes []RecipeSummary `json:"recipes"`
}

type CreateRecipeResponse struct {
	ID string `json:"id"`
}

type RecipeFilter struct {
	Name string
	Tag  string
}
