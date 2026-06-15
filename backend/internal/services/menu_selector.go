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
// the household's vegetarian-days quota is met, and within each phase greedily
// preferring the candidate whose ingredients overlap most with the recipes
// already chosen this week (so the shopping list shares ingredients). It cycles
// through the pool when there are fewer recipes than days.
//
// It carries no I/O and no clock, so it is fully unit-testable with a seeded RNG.
type menuSelector struct {
	veg    []domain.RecipeSummary // vegetarian candidates (shuffled, tie-break order)
	nonVeg []domain.RecipeSummary // everything else (shuffled, tie-break order)
	pool   []domain.RecipeSummary // veg + nonVeg, the cycling fallback order

	vegRemaining int // vegetarian slots still owed by the preference

	// ingredients maps recipe ID -> its overlap-relevant ingredient keys (lower-
	// cased canonical names, pantry staples already dropped). Empty/missing when
	// no ingredient data was supplied, in which case the selector degrades to the
	// previous shuffle-order behavior.
	ingredients map[string][]string

	// picked tracks IDs already placed this week (so cycling re-picks are scored
	// against the full week) and the multiset of ingredient keys chosen so far.
	picked        map[string]bool
	chosenKeys    map[string]int
	hasIngredient bool // true once any ingredient data is present
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

// overlapKeys returns the lower-cased ingredient names of a recipe that count
// toward overlap scoring. Pantry staples (spices, oils, condiments, etc.) are
// dropped because they appear in almost every dish and would otherwise dominate
// the score with no shopping-economy benefit — the same categories the shopping
// list treats as Kryddor / Såser & olja. Blank names are ignored.
func overlapKeys(ings []domain.Ingredient) []string {
	keys := make([]string, 0, len(ings))
	for _, ing := range ings {
		name := strings.TrimSpace(ing.Name)
		if name == "" {
			continue
		}
		if isPantryStaple(name) {
			continue
		}
		keys = append(keys, strings.ToLower(name))
	}
	return keys
}

// isPantryStaple reports whether an ingredient is a pantry staple that should
// not drive overlap scoring. It reuses the shopping list's category map: spices
// and the oil/sauce shelf are bought once and reused, so sharing them across
// dishes is not a meaningful economy signal.
func isPantryStaple(name string) bool {
	switch categorizeIngredient(name) {
	case "Kryddor", "Såser & olja":
		return true
	default:
		return false
	}
}

// newMenuSelector builds a selector from the catalog and preferences. rng is
// injected so tests can be deterministic. ingredientsByID supplies the overlap-
// relevant ingredient keys per recipe ID; pass nil to disable overlap scoring
// (the selector then reproduces the prior shuffle-order behavior). Returns
// (nil, false) when no recipe survives the excluded-tag filter — the caller
// should treat that as "no recipes available", mirroring the empty-catalog case.
func newMenuSelector(recipes []domain.RecipeSummary, prefs domain.MenuPreferences, rng *rand.Rand, ingredientsByID map[string][]domain.Ingredient) (*menuSelector, bool) {
	candidates := filterByExcludedTags(recipes, prefs.ExcludedTags)
	if len(candidates) == 0 {
		return nil, false
	}

	sel := &menuSelector{
		vegRemaining: prefs.VegetarianDays,
		ingredients:  make(map[string][]string, len(candidates)),
		picked:       make(map[string]bool, len(candidates)),
		chosenKeys:   make(map[string]int),
	}
	for _, r := range candidates {
		if isVegetarian(r) {
			sel.veg = append(sel.veg, r)
		} else {
			sel.nonVeg = append(sel.nonVeg, r)
		}
		if ings, ok := ingredientsByID[r.ID]; ok {
			keys := overlapKeys(ings)
			sel.ingredients[r.ID] = keys
			if len(keys) > 0 {
				sel.hasIngredient = true
			}
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

// overlapScore counts how many of a recipe's overlap-relevant ingredients are
// already on this week's shopping list (i.e. used by an already-picked recipe).
// Higher means more shared ingredients, so fewer distinct items to buy.
func (s *menuSelector) overlapScore(id string) int {
	score := 0
	for _, k := range s.ingredients[id] {
		if s.chosenKeys[k] > 0 {
			score++
		}
	}
	return score
}

// commit records a chosen recipe: it marks the ID as picked and folds its
// ingredient keys into the running week so subsequent picks can score against
// them.
func (s *menuSelector) commit(id string) {
	s.picked[id] = true
	for _, k := range s.ingredients[id] {
		s.chosenKeys[k]++
	}
}

// pickBest chooses the highest-overlap recipe from candidates, preferring those
// not yet used this week; ties (including the very first pick, when nothing has
// been chosen yet) resolve to the candidate earliest in the shuffled order, so
// selection stays deterministic under a seeded RNG. onlyFresh restricts the
// choice to recipes not already picked; when every candidate is already picked
// (the cycling case) it is called again with onlyFresh=false. Returns "" when
// candidates is empty.
func (s *menuSelector) pickBest(candidates []domain.RecipeSummary, onlyFresh bool) string {
	bestID := ""
	bestScore := -1
	for _, r := range candidates {
		if onlyFresh && s.picked[r.ID] {
			continue
		}
		score := s.overlapScore(r.ID)
		if score > bestScore {
			bestScore = score
			bestID = r.ID
		}
	}
	return bestID
}

// Next returns the recipe ID to place in the next non-skipped slot. It prefers
// a fresh vegetarian dish while the vegetarian quota is unmet, otherwise draws
// from the combined pool. Within each phase it greedily picks the candidate
// that shares the most ingredients with the recipes already chosen this week,
// cycling once every recipe has been used.
func (s *menuSelector) Next() string {
	// Vegetarian phase: honor the quota with fresh vegetarian dishes while any
	// remain unused.
	if s.vegRemaining > 0 {
		if id := s.pickBest(s.veg, true); id != "" {
			s.vegRemaining--
			s.commit(id)
			return id
		}
		// No fresh vegetarian dish left; fall through to the general pool but
		// stop owing vegetarian slots so we don't spin here.
		s.vegRemaining = 0
	}

	// General phase: prefer a fresh recipe. When every recipe has been used,
	// start a new cycle — clear the picked set so each recipe gets used again
	// before any repeats, while still overlap-scoring within the fresh round.
	if id := s.pickBest(s.pool, true); id != "" {
		s.commit(id)
		return id
	}
	s.startNewCycle()
	id := s.pickBest(s.pool, true)
	if id == "" {
		// Defensive: pool is non-empty by construction, but never return "".
		id = s.pool[0].ID
	}
	s.commit(id)
	return id
}

// startNewCycle clears the picked set so the cycling fallback hands out every
// recipe again before repeating. The week's accumulated ingredient keys are
// kept, so a re-picked recipe is still scored against the full week.
func (s *menuSelector) startNewCycle() {
	s.picked = make(map[string]bool, len(s.pool))
}
