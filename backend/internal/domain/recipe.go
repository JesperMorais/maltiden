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

func (r CreateRecipeRequest) Validate() error {
	if err := validateRecipeFields(r.Name, r.Servings, r.Emoji, r.Tags, r.Ingredients, r.Instructions); err != nil {
		return err
	}
	data, err := json.Marshal(r)
	if err != nil || len(data) > 32768 {
		return fmt.Errorf("%w: payload", ErrRecipePayloadTooLarge)
	}
	return nil
}

func (r UpdateRecipeRequest) Validate() error {
	if err := validateRecipeFields(r.Name, r.Servings, r.Emoji, r.Tags, r.Ingredients, r.Instructions); err != nil {
		return err
	}
	data, err := json.Marshal(r)
	if err != nil || len(data) > 32768 {
		return fmt.Errorf("%w: payload", ErrRecipePayloadTooLarge)
	}
	return nil
}

func validateRecipeFields(name string, servings int, emoji string, tags []string, ingredients []Ingredient, instructions []string) error {
	if name == "" {
		return fmt.Errorf("%w: name", ErrNameRequired)
	}
	if utf8.RuneCountInString(name) > 200 {
		return fmt.Errorf("%w: name", ErrNameTooLong)
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return fmt.Errorf("%w: name", ErrContainsControlChar)
		}
	}

	if servings < 1 || servings > 50 {
		return fmt.Errorf("%w: servings", ErrInvalidServings)
	}

	if utf8.RuneCountInString(emoji) > 8 {
		return fmt.Errorf("%w: emoji", ErrEmojiTooLong)
	}

	if len(tags) > 12 {
		return fmt.Errorf("%w: tags", ErrTooManyTags)
	}
	for _, tag := range tags {
		if utf8.RuneCountInString(tag) > 50 {
			return fmt.Errorf("%w: tag", ErrTagTooLong)
		}
		for _, r := range tag {
			if unicode.IsControl(r) {
				return fmt.Errorf("%w: tag", ErrContainsControlChar)
			}
		}
	}

	if len(ingredients) == 0 {
		return fmt.Errorf("%w: ingredients", ErrIngredientsRequired)
	}
	if len(ingredients) > 40 {
		return fmt.Errorf("%w: ingredients", ErrTooManyIngredients)
	}
	for _, ing := range ingredients {
		if ing.Name == "" {
			return fmt.Errorf("%w: ingredient name", ErrIngredientNameRequired)
		}
		if utf8.RuneCountInString(ing.Name) > 80 {
			return fmt.Errorf("%w: ingredient name", ErrIngredientNameTooLong)
		}
		for _, r := range ing.Name {
			if unicode.IsControl(r) {
				return fmt.Errorf("%w: ingredient name", ErrContainsControlChar)
			}
		}
		if ing.Amount < 0 {
			return fmt.Errorf("%w: amount", ErrInvalidAmount)
		}
		if ing.Amount > 10000 {
			return fmt.Errorf("%w: amount", ErrAmountTooLarge)
		}
		if utf8.RuneCountInString(ing.Unit) > 20 {
			return fmt.Errorf("%w: unit", ErrUnitTooLong)
		}
		for _, r := range ing.Unit {
			if unicode.IsControl(r) {
				return fmt.Errorf("%w: unit", ErrContainsControlChar)
			}
		}
	}

	if len(instructions) == 0 {
		return fmt.Errorf("%w: instructions", ErrInstructionsRequired)
	}
	if len(instructions) > 30 {
		return fmt.Errorf("%w: instructions", ErrTooManyInstructions)
	}
	for _, step := range instructions {
		if utf8.RuneCountInString(step) > 500 {
			return fmt.Errorf("%w: instruction", ErrInstructionTooLong)
		}
		for _, r := range step {
			if unicode.IsControl(r) {
				return fmt.Errorf("%w: instruction", ErrContainsControlChar)
			}
		}
	}

	return nil
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
