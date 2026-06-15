package domain

import "time"

type MenuDay struct {
	Date     string `json:"date"`
	RecipeID string `json:"recipeId,omitempty"`
	Servings int    `json:"servings"`
	Skip     bool   `json:"skip,omitempty"`

	// Leftover marks a meal-prep "eat the leftovers" day: it reuses the recipe
	// cooked on CookDate (the batch cook day, which carries 2× servings), so this
	// day requires no new groceries and is excluded from the shopping list.
	Leftover bool `json:"leftover,omitempty"`
	// CookDate is the date of the cook day whose batch this leftover day eats
	// from (only set when Leftover is true).
	CookDate string `json:"cookDate,omitempty"`
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

	// LockedDays pins specific dates to a chosen recipe (date → recipeId). The
	// generator keeps these recipes in place and never replaces them, but still
	// counts them when scoring the rest of the week (overlap, protein variety,
	// recency, vegetarian quota). An unknown recipeId is silently ignored.
	LockedDays map[string]string `json:"lockedDays,omitempty"`

	// PrepMode opts the week into meal-prep (batch-cooking) placement: the
	// selector cooks a batchable recipe once at 2× servings and reuses it as
	// leftovers the next day. A nil pointer falls back to the household's saved
	// PrepModeDefault preference; an explicit false forces classic Phase-1
	// generation regardless of the default.
	PrepMode *bool `json:"prepMode,omitempty"`
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

	// Leftover / CookDate mirror MenuDay (see there): a meal-prep leftovers day
	// reuses the recipe cooked on CookDate and buys no new groceries.
	Leftover bool   `json:"leftover,omitempty"`
	CookDate string `json:"cookDate,omitempty"`
}

type MenuResponse struct {
	ID      string            `json:"id"`
	Days    []MenuResponseDay `json:"days"`
	Economy *MenuEconomy      `json:"economy,omitempty"`
}

// SharedIngredient is a non-staple ingredient that appears across two or more
// recipes in the generated week. Name is the first raw (display) name seen for
// the canonical, so the UI can show "Lök" rather than the canonical key.
type SharedIngredient struct {
	CanonicalName string `json:"canonicalName"`
	Name          string `json:"name"`
	RecipeCount   int    `json:"recipeCount"`
}

// MenuEconomy summarizes how much the generated week reuses ingredients across
// its recipes. Pantry staples (salt, oil, flour, …) are excluded from every
// figure so the numbers reflect what actually has to be bought.
type MenuEconomy struct {
	SharedIngredients   []SharedIngredient `json:"sharedIngredients"`
	DistinctItemsToBuy  int                `json:"distinctItemsToBuy"`
	TotalIngredientRefs int                `json:"totalIngredientRefs"`
}

// Scoring constants for the smart menu generator. The selector maximizes a
// single signed score: λ·overlap − μ·proteinRepeats − duplicates − recency.
const (
	// OverlapReward (λ) is awarded per extra recipe sharing a non-staple
	// ingredient — it pulls the week toward dishes that reuse groceries.
	OverlapReward = 1.0
	// ProteinVarietyPenalty (μ) discourages repeating the same main protein.
	ProteinVarietyPenalty = 2.0
	// DuplicateRecipePenalty makes the selector treat the same recipe (or a
	// near-duplicate) appearing twice as a near-hard error.
	DuplicateRecipePenalty = 100.0
	// RecencyPenalty discourages picking recipes used in the recent weeks.
	RecencyPenalty = 1.5
	// RecencyWindowMenus is how many of the household's most recent menus feed
	// the recency penalty.
	RecencyWindowMenus = 3
	// SelectorRestarts is how many randomized greedy constructions are tried;
	// the best-scoring week wins (ties broken by lowest restart index).
	SelectorRestarts = 24
	// BatchMultiplier is how many servings of base portions a meal-prep cook day
	// produces (one day to eat + one day of leftovers).
	BatchMultiplier = 2
	// PrepPairsPerWeek is the maximum number of cook→leftovers batch pairs the
	// prep-aware selector places in a single generated week.
	PrepPairsPerWeek = 1
)

// DietCompatible reports whether a recipe's diet class is acceptable for a
// household diet profile. Both an empty profile and an empty recipe class are
// treated as "unknown" and accepted gracefully, so missing metadata never
// silently drops recipes.
func DietCompatible(profile, recipeClass string) bool {
	if profile == "" || recipeClass == "" {
		return true
	}
	switch profile {
	case DietClassOmnivore:
		return true
	case DietClassVegetarian:
		return recipeClass == DietClassVegetarian || recipeClass == DietClassVegan
	case DietClassVegan:
		return recipeClass == DietClassVegan
	case DietClassPescetarian:
		return recipeClass == DietClassPescetarian ||
			recipeClass == DietClassVegetarian ||
			recipeClass == DietClassVegan
	default:
		// Unknown profile value — accept everything rather than filter blindly.
		return true
	}
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

	// PrepModeDefault is the household's default for meal-prep (batch cooking).
	// When true, a generate request that omits prepMode opts into prep placement.
	PrepModeDefault bool `json:"prepModeDefault"`
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

	PrepModeDefault bool `json:"prepModeDefault"`
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
		PrepModeDefault:     false,
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
