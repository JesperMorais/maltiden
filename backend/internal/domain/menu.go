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

type MenuResponse struct {
	ID   string            `json:"id"`
	Days []MenuResponseDay `json:"days"`
}

// MenuPreferences holds a household's persisted menu-generation preferences.
// These are applied by the selector core as hard filters (ExcludedTags) and as
// defaults (DefaultDays, DefaultServings) when a generate request omits a value.
// VegetarianDays is the minimum number of vegetarian days to prefer per week.
type MenuPreferences struct {
	HouseholdID     string    `json:"householdId"`
	ExcludedTags    []string  `json:"excludedTags"`
	DefaultDays     int       `json:"defaultDays"`
	DefaultServings int       `json:"defaultServings"`
	VegetarianDays  int       `json:"vegetarianDays"`
	UpdatedAt       time.Time `json:"updatedAt,omitempty"`

	// DietProfile is an optional household-wide diet class (one of the
	// DietClass enum values, or empty for "unknown"). DislikedIngredients are
	// free-text strings the household wants to avoid; Phase 0 only STORES them.
	DietProfile         string   `json:"dietProfile,omitempty"`
	DislikedIngredients []string `json:"dislikedIngredients"`
}

// UpdateMenuPreferencesRequest is the payload for upserting a household's
// menu preferences.
type UpdateMenuPreferencesRequest struct {
	ExcludedTags    []string `json:"excludedTags"`
	DefaultDays     int      `json:"defaultDays"`
	DefaultServings int      `json:"defaultServings"`
	VegetarianDays  int      `json:"vegetarianDays"`

	DietProfile         string   `json:"dietProfile,omitempty"`
	DislikedIngredients []string `json:"dislikedIngredients"`
}

// DefaultMenuPreferences returns the preferences applied to a household that
// has never saved any (no excluded tags, full week, 4 servings).
func DefaultMenuPreferences(householdID string) MenuPreferences {
	return MenuPreferences{
		HouseholdID:         householdID,
		ExcludedTags:        []string{},
		DefaultDays:         7,
		DefaultServings:     4,
		VegetarianDays:      0,
		DietProfile:         "",
		DislikedIngredients: []string{},
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
	if err := validateStringList(r.ExcludedTags, ErrTooManyExcludedTags, ErrInvalidExcludedTag); err != nil {
		return err
	}
	// Diet profile (empty = unknown, always accepted)
	if !IsValidDietClass(r.DietProfile) {
		return ErrInvalidDietProfile
	}
	if err := validateStringList(r.DislikedIngredients, ErrTooManyDislikedIngredients, ErrInvalidDislikedIngredient); err != nil {
		return err
	}
	return nil
}

// validateStringList enforces the shared "≤50 entries, each 1–50 chars" bound
// used by both excluded tags and disliked ingredients.
func validateStringList(items []string, errTooMany, errInvalid error) error {
	if len(items) > 50 {
		return errTooMany
	}
	for _, item := range items {
		if len(item) == 0 || len(item) > 50 {
			return errInvalid
		}
	}
	return nil
}
