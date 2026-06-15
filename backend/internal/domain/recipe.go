package domain

import (
	"encoding/json"
	"fmt"
	"time"
	"unicode"
	"unicode/utf8"
)

type Ingredient struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Unit   string  `json:"unit"`

	// Metadata enriched by the parser/backfill. All fields are additive and
	// stored inside the existing ingredients JSON column. The zero value of
	// each field means "unknown" and must never be treated as a real signal.
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

	// Recipe-level metadata (stored in dedicated scalar columns). Empty/0
	// means "unknown".
	MainProtein string `json:"mainProtein,omitempty"`
	DietClass   string `json:"dietClass,omitempty"`
	Batchable   bool   `json:"batchable,omitempty"`
	CookMinutes int    `json:"cookMinutes,omitempty"`
}

type RecipeSummary struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Servings int      `json:"servings"`
	Emoji    string   `json:"emoji,omitempty"`
	Tags     []string `json:"tags"`

	MainProtein string `json:"mainProtein,omitempty"`
	DietClass   string `json:"dietClass,omitempty"`
	Batchable   bool   `json:"batchable,omitempty"`
	CookMinutes int    `json:"cookMinutes,omitempty"`
}

type CreateRecipeRequest struct {
	Name         string       `json:"name"`
	Servings     int          `json:"servings"`
	Emoji        string       `json:"emoji,omitempty"`
	Tags         []string     `json:"tags"`
	Ingredients  []Ingredient `json:"ingredients"`
	Instructions []string     `json:"instructions"`

	MainProtein string `json:"mainProtein,omitempty"`
	DietClass   string `json:"dietClass,omitempty"`
	Batchable   bool   `json:"batchable,omitempty"`
	CookMinutes int    `json:"cookMinutes,omitempty"`
}

type UpdateRecipeRequest struct {
	Name         string       `json:"name"`
	Servings     int          `json:"servings"`
	Emoji        string       `json:"emoji,omitempty"`
	Tags         []string     `json:"tags"`
	Ingredients  []Ingredient `json:"ingredients"`
	Instructions []string     `json:"instructions"`

	MainProtein string `json:"mainProtein,omitempty"`
	DietClass   string `json:"dietClass,omitempty"`
	Batchable   bool   `json:"batchable,omitempty"`
	CookMinutes int    `json:"cookMinutes,omitempty"`
}

// DietClass enum values. Empty string means "unknown" and is always accepted.
const (
	DietClassOmnivore    = "omnivore"
	DietClassVegetarian  = "vegetarian"
	DietClassVegan       = "vegan"
	DietClassPescetarian = "pescetarian"
)

// IsValidDietClass reports whether s is a recognized diet class. The empty
// string ("unknown") is accepted; any other unrecognized value is rejected.
func IsValidDietClass(s string) bool {
	switch s {
	case "", DietClassOmnivore, DietClassVegetarian, DietClassVegan, DietClassPescetarian:
		return true
	default:
		return false
	}
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

func hasControlChar(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func (r CreateRecipeRequest) Validate() error {
	// Name
	if utf8.RuneCountInString(r.Name) == 0 {
		return ErrNameRequired
	}
	if utf8.RuneCountInString(r.Name) > 200 {
		return fmt.Errorf("name: %w", ErrNameTooLong)
	}
	if hasControlChar(r.Name) {
		return fmt.Errorf("name: %w", ErrContainsControlChar)
	}

	// Servings
	if r.Servings < 1 || r.Servings > 50 {
		return ErrInvalidServings
	}

	// Diet class (empty = unknown, always accepted)
	if !IsValidDietClass(r.DietClass) {
		return ErrInvalidDietClass
	}

	// Emoji
	if utf8.RuneCountInString(r.Emoji) > 8 {
		return ErrEmojiTooLong
	}

	// Tags
	if len(r.Tags) > 12 {
		return ErrTooManyTags
	}
	for _, tag := range r.Tags {
		if utf8.RuneCountInString(tag) > 50 {
			return fmt.Errorf("tag: %w", ErrTagTooLong)
		}
		if hasControlChar(tag) {
			return fmt.Errorf("tag: %w", ErrContainsControlChar)
		}
	}

	// Ingredients
	if len(r.Ingredients) > 40 {
		return ErrTooManyIngredients
	}
	for _, ing := range r.Ingredients {
		if utf8.RuneCountInString(ing.Name) > 80 {
			return fmt.Errorf("ingredient name: %w", ErrIngredientNameTooLong)
		}
		if hasControlChar(ing.Name) {
			return fmt.Errorf("ingredient name: %w", ErrContainsControlChar)
		}
		if utf8.RuneCountInString(ing.CanonicalName) > 80 {
			return fmt.Errorf("ingredient canonical name: %w", ErrIngredientNameTooLong)
		}
		if hasControlChar(ing.CanonicalName) {
			return fmt.Errorf("ingredient canonical name: %w", ErrContainsControlChar)
		}
		if ing.Amount < 0 {
			return ErrInvalidAmount
		}
		if ing.Amount > 10000 {
			return ErrAmountTooLarge
		}
		if utf8.RuneCountInString(ing.Unit) > 20 {
			return fmt.Errorf("unit: %w", ErrUnitTooLong)
		}
		if hasControlChar(ing.Unit) {
			return fmt.Errorf("unit: %w", ErrContainsControlChar)
		}
	}

	// Instructions
	if len(r.Instructions) > 30 {
		return ErrTooManyInstructions
	}
	for _, step := range r.Instructions {
		if utf8.RuneCountInString(step) > 500 {
			return fmt.Errorf("instruction: %w", ErrInstructionTooLong)
		}
		if hasControlChar(step) {
			return fmt.Errorf("instruction: %w", ErrContainsControlChar)
		}
	}

	// Total JSON size backstop
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	if len(data) > 32*1024 {
		return ErrRecipeTooLarge
	}

	return nil
}
