package services

import (
	"maltiden/internal/domain"
	"math/rand/v2"
	"strings"
)

// vegetarianTag is the recipe tag the selector treats as marking a meatless
// dish. Matches the Swedish tag used across the seed recipes.
const vegetarianTag = "vegetariskt"

// menuSelector holds the prepared candidate pool for one generate run. It is
// the deterministic "selector core": given a recipe catalog and a household's
// preferences, it (1) hard-filters out recipes carrying any excluded tag, then
// (2) hands back recipes one slot at a time, preferring vegetarian dishes until
// the household's vegetarian-days quota is met, and cycling through the pool
// when there are fewer recipes than days.
//
// It carries no I/O and no clock, so it is fully unit-testable with a seeded RNG.
type menuSelector struct {
	veg    []domain.RecipeSummary // vegetarian candidates (shuffled)
	nonVeg []domain.RecipeSummary // everything else (shuffled)
	vegIdx int
	genIdx int
	pool   []domain.RecipeSummary // veg + nonVeg, the cycling fallback order

	vegRemaining int // vegetarian slots still owed by the preference
}

// filterByExcludedTags returns the subset of recipes that carry none of the
// excluded tags. Tag matching is case-insensitive and trims surrounding space
// so user-entered preferences ("Fisk ") match seed tags ("fisk").
func filterByExcludedTags(recipes []domain.RecipeSummary, excludedTags []string) []domain.RecipeSummary {
	if len(excludedTags) == 0 {
		out := make([]domain.RecipeSummary, len(recipes))
		copy(out, recipes)
		return out
	}

	excluded := make(map[string]bool, len(excludedTags))
	for _, t := range excludedTags {
		excluded[normalizeTag(t)] = true
	}

	out := make([]domain.RecipeSummary, 0, len(recipes))
	for _, r := range recipes {
		if recipeHasExcludedTag(r, excluded) {
			continue
		}
		out = append(out, r)
	}
	return out
}

func recipeHasExcludedTag(r domain.RecipeSummary, excluded map[string]bool) bool {
	for _, t := range r.Tags {
		if excluded[normalizeTag(t)] {
			return true
		}
	}
	return false
}

func isVegetarian(r domain.RecipeSummary) bool {
	for _, t := range r.Tags {
		if normalizeTag(t) == vegetarianTag {
			return true
		}
	}
	return false
}

func normalizeTag(t string) string {
	return strings.ToLower(strings.TrimSpace(t))
}

// newMenuSelector builds a selector from the catalog and preferences. rng is
// injected so tests can be deterministic. Returns (nil, false) when no recipe
// survives the excluded-tag filter — the caller should treat that as "no
// recipes available", mirroring the empty-catalog case.
func newMenuSelector(recipes []domain.RecipeSummary, prefs domain.MenuPreferences, rng *rand.Rand) (*menuSelector, bool) {
	candidates := filterByExcludedTags(recipes, prefs.ExcludedTags)
	if len(candidates) == 0 {
		return nil, false
	}

	sel := &menuSelector{vegRemaining: prefs.VegetarianDays}
	for _, r := range candidates {
		if isVegetarian(r) {
			sel.veg = append(sel.veg, r)
		} else {
			sel.nonVeg = append(sel.nonVeg, r)
		}
	}

	shuffle(sel.veg, rng)
	shuffle(sel.nonVeg, rng)

	// Cycling pool prefers vegetarian first so the quota can be honored even
	// when veg recipes are scarce, then falls back to the full set.
	sel.pool = append(sel.pool, sel.veg...)
	sel.pool = append(sel.pool, sel.nonVeg...)

	return sel, true
}

func shuffle(s []domain.RecipeSummary, rng *rand.Rand) {
	if rng == nil {
		rand.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
		return
	}
	rng.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
}

// Next returns the recipe ID to place in the next non-skipped slot. It prefers
// a fresh vegetarian dish while the vegetarian quota is unmet, otherwise draws
// from the combined pool, cycling once exhausted.
func (s *menuSelector) Next() string {
	if s.vegRemaining > 0 && s.vegIdx < len(s.veg) {
		id := s.veg[s.vegIdx].ID
		s.vegIdx++
		s.vegRemaining--
		return id
	}

	id := s.pool[s.genIdx%len(s.pool)].ID
	s.genIdx++
	return id
}
