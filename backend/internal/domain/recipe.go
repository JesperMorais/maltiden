package domain

import "time"

type Ingredient struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Unit   string  `json:"unit"`
}

type Recipe struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Servings     int          `json:"servings"`
	Emoji        string       `json:"emoji,omitempty"`
	Tags         []string     `json:"tags"`
	Ingredients  []Ingredient `json:"ingredients"`
	Instructions []string     `json:"instructions"`
	CreatedAt    time.Time    `json:"createdAt"`
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
