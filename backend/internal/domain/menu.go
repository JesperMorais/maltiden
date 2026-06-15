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

type MenuResponseDay struct {
	Date       string `json:"date"`
	RecipeID   string `json:"recipeId,omitempty"`
	RecipeName string `json:"recipeName,omitempty"`
	Emoji      string `json:"emoji,omitempty"`
	Servings   int    `json:"servings"`
	Skip       bool   `json:"skip,omitempty"`
}

// SharedIngredient is one ingredient reused across two or more recipes in a
// generated week. It surfaces the shopping-economy benefit of the greedy
// overlap selection (#258) to the user: fewer distinct items to buy. Pantry
// staples (spices, oils, sauces) are excluded, mirroring the overlap scorer.
type SharedIngredient struct {
	Name        string `json:"name"`        // display name (first-seen casing)
	RecipeCount int    `json:"recipeCount"` // distinct recipes this week using it (>= 2)
}

type MenuResponse struct {
	ID   string            `json:"id"`
	Days []MenuResponseDay `json:"days"`
	// SharedIngredients lists ingredients reused across the week's recipes,
	// most-shared first. Empty when nothing is shared (or no ingredient data).
	SharedIngredients []SharedIngredient `json:"sharedIngredients,omitempty"`
}

// MenuPreferences holds a household's persisted menu-generation preferences.
// These are applied by the selector core as hard filters (ExcludedTags,
// DislikedIngredients) and as defaults (DefaultDays, DefaultServings) when a
// generate request omits a value. VegetarianDays is the minimum number of
// vegetarian days to prefer per week. DislikedIngredients is a list of
// ingredient names; any recipe containing one (name-normalized match) is
// hard-filtered out of generation.
type MenuPreferences struct {
	HouseholdID         string    `json:"householdId"`
	ExcludedTags        []string  `json:"excludedTags"`
	DislikedIngredients []string  `json:"dislikedIngredients"`
	DefaultDays         int       `json:"defaultDays"`
	DefaultServings     int       `json:"defaultServings"`
	VegetarianDays      int       `json:"vegetarianDays"`
	UpdatedAt           time.Time `json:"updatedAt,omitempty"`
}

// UpdateMenuPreferencesRequest is the payload for upserting a household's
// menu preferences.
type UpdateMenuPreferencesRequest struct {
	ExcludedTags        []string `json:"excludedTags"`
	DislikedIngredients []string `json:"dislikedIngredients"`
	DefaultDays         int      `json:"defaultDays"`
	DefaultServings     int      `json:"defaultServings"`
	VegetarianDays      int      `json:"vegetarianDays"`
}

// DefaultMenuPreferences returns the preferences applied to a household that
// has never saved any (no excluded tags, no disliked ingredients, full week,
// 4 servings).
func DefaultMenuPreferences(householdID string) MenuPreferences {
	return MenuPreferences{
		HouseholdID:         householdID,
		ExcludedTags:        []string{},
		DislikedIngredients: []string{},
		DefaultDays:         7,
		DefaultServings:     4,
		VegetarianDays:      0,
	}
}

// Validate checks an update request against the same bounds the generator
// enforces, so preferences can never persist values the generator would reject.
func (r UpdateMenuPreferencesRequest) Validate() error {
	if r.DefaultDays < 1 || r.DefaultDays > 31 {
		return ErrInvalidDays
	}
	if r.DefaultServings < 1 || r.DefaultServings > 100 {
		return ErrInvalidServings
	}
	if r.VegetarianDays < 0 || r.VegetarianDays > r.DefaultDays {
		return ErrInvalidVegetarianDays
	}
	if len(r.ExcludedTags) > 50 {
		return ErrTooManyExcludedTags
	}
	for _, tag := range r.ExcludedTags {
		if len(tag) == 0 || len(tag) > 50 {
			return ErrInvalidExcludedTag
		}
	}
	if len(r.DislikedIngredients) > 50 {
		return ErrTooManyDislikedIngredients
	}
	for _, ing := range r.DislikedIngredients {
		if len(ing) == 0 || len(ing) > 80 {
			return ErrInvalidDislikedIngredient
		}
	}
	return nil
}
