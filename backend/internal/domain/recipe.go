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

type UpdateRecipeRequest struct {
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
