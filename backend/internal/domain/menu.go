package domain

import "time"

// PrepModeBatch marks a menu day as a batch-cook day: the recipe is cooked
// once at double servings and reused as leftovers on a later day (#248 Phase 2,
// "laga en gång, ät två gånger"). The empty string means an ordinary day.
const PrepModeBatch = "batch"

type MenuDay struct {
	Date     string `json:"date"`
	RecipeID string `json:"recipeId,omitempty"`
	Servings int    `json:"servings"`
	Skip     bool   `json:"skip,omitempty"`
	// PrepMode marks how the day is prepared. "" is an ordinary day;
	// PrepModeBatch marks a cook-once-eat-twice day (batch cooking, #248).
	PrepMode string `json:"prepMode,omitempty"`
	// LeftoverOf, when set, is the date of the batch cook-day this day reuses.
	// A leftovers day carries the same RecipeID as its cook-day but is not
	// cooked again. Empty for cook-days and ordinary days.
	LeftoverOf string `json:"leftoverOf,omitempty"`
}

type Menu struct {
	ID          string    `json:"id"`
	HouseholdID string    `json:"householdId"`
	Days        []MenuDay `json:"days"`
	CreatedAt   time.Time `json:"createdAt"`
	// Rationale is a short Swedish explanation of the week, written by the
	// optional Phase-4 AI experience layer. Empty when the layer is disabled.
	Rationale string `json:"rationale,omitempty"`
}

type GenerateMenuRequest struct {
	Days          int            `json:"days"`
	Servings      int            `json:"servings"`
	SkipDays      []string       `json:"skipDays"`
	ExtraPortions map[string]int `json:"extraPortions"`
	// PrepMode, when true, enables batch cooking for the week (#248 Phase 2):
	// a batchable recipe is placed on a cook-day at double servings and reused
	// as leftovers on the next eligible day. Defaults to off.
	PrepMode bool `json:"prepMode,omitempty"`
	// Wishes is optional free-text Swedish input ("två vegetariska dagar, ingen
	// fisk"). When the Phase-4 experience layer is enabled it is parsed into
	// per-run constraints; otherwise it is ignored (WishesIgnored in response).
	Wishes string `json:"wishes,omitempty"`
	// Arrange, when true, asks the Phase-4 experience layer to place the chosen
	// recipes across weekdays and write a Swedish rationale. No-op without the
	// layer; the deterministic order is kept.
	Arrange bool `json:"arrange,omitempty"`
	// ExcludeRecipeIDs are recipe IDs to keep OUT of this generation — e.g.
	// recipes on locked days during a regenerate, so the same dish isn't reused
	// across the week (#248). Ignored if it would exclude every recipe.
	ExcludeRecipeIDs []string `json:"excludeRecipeIds,omitempty"`
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
	// PrepMode / LeftoverOf surface batch cooking to the UX (#248 Phase 2).
	// PrepModeBatch marks a cook-once-eat-twice cook-day; LeftoverOf is the
	// cook-day date a leftovers day reuses. Both omitempty for ordinary days.
	PrepMode   string `json:"prepMode,omitempty"`
	LeftoverOf string `json:"leftoverOf,omitempty"`
	// Nutrition is the recipe's per-serving nutrition for this day (#248 Phase
	// 3), when known. Nil when the recipe carries no nutrition data, so the UI
	// shows macros only where they exist rather than misleading zeros.
	Nutrition *Nutrition `json:"nutrition,omitempty"`
}

// MenuNutrition is the average per-meal nutrition for a generated or current
// menu (#248 Phase 3): the mean of each meal's per-serving macros — what a
// typical plate looks like, which is the figure users actually reason about
// (not a whole-week sum). Each non-skipped day with a recipe is one meal
// (leftover days included). Partial is true when at least one meal's recipe had
// no nutrition data, so the average is over the subset that did.
type MenuNutrition struct {
	Calories float64 `json:"calories"`
	ProteinG float64 `json:"proteinG"`
	CarbsG   float64 `json:"carbsG"`
	FatG     float64 `json:"fatG"`
	Partial  bool    `json:"partial"`
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
	// Rationale is the optional Phase-4 Swedish explanation of the week. Empty
	// when the experience layer is disabled or arrangement was not requested.
	Rationale string `json:"rationale,omitempty"`
	// WishesIgnored is true when the request carried free-text wishes but the
	// experience layer was unavailable (or failed), so they had no effect.
	WishesIgnored bool `json:"wishesIgnored,omitempty"`
	// Nutrition is the week's aggregated macro total (#248 Phase 3). Nil — and
	// omitted — when no day in the menu carries nutrition data.
	Nutrition *MenuNutrition `json:"nutrition,omitempty"`
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

// MaxRationaleLen is the maximum number of runes kept from an AI arranger
// rationale; longer text is truncated before persisting / returning.
const MaxRationaleLen = 600

// ParsedWishConstraints is the structured result of parsing a household's
// free-text weekly wishes (Phase 4). Pointer fields distinguish "the model said
// nothing about this" (nil) from an explicit zero value. Non-nil fields are
// merged into the effective MenuPreferences (or the request) for one generation
// only; explicit typed request fields still win over wishes.
type ParsedWishConstraints struct {
	VegetarianDays           *int     `json:"vegetarianDays"`
	PrepMode                 *bool    `json:"prepMode"`
	Days                     *int     `json:"days"`
	Servings                 *int     `json:"servings"`
	ExtraExcludedTags        []string `json:"extraExcludedTags"`
	ExtraDislikedIngredients []string `json:"extraDislikedIngredients"`
}

// Clamp bounds every parsed wish field to the same ranges the generator and
// preferences validation enforce, so an out-of-range model output can never
// produce an invalid generation. A nil receiver is a no-op.
func (c *ParsedWishConstraints) Clamp() {
	if c == nil {
		return
	}
	if c.Days != nil {
		*c.Days = clampInt(*c.Days, 1, 31)
	}
	if c.Servings != nil {
		*c.Servings = clampInt(*c.Servings, 1, 100)
	}
	if c.VegetarianDays != nil {
		*c.VegetarianDays = clampInt(*c.VegetarianDays, 0, 31)
	}
	c.ExtraExcludedTags = clampStringList(c.ExtraExcludedTags)
	c.ExtraDislikedIngredients = clampStringList(c.ExtraDislikedIngredients)
}

// clampStringList drops empty and oversized (>50-char) entries and caps the
// list at 50 items, matching the preference-validation bounds.
func clampStringList(items []string) []string {
	if items == nil {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if len(item) == 0 || len(item) > 50 {
			continue
		}
		out = append(out, item)
		if len(out) >= 50 {
			break
		}
	}
	return out
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
