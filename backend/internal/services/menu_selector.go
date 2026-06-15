package services

import (
	"maltiden/internal/domain"
	"math/rand/v2"
	"sort"
	"strings"
)

// vegetarianTag is the recipe tag the selector treats as marking a meatless
// dish. Matches the Swedish tag used across the seed recipes.
const vegetarianTag = "vegetariskt"

// Scoring weights. The selector's score is
//
//	score = overlapReward·overlap − varietyWeight·varietyPenalty − recencyWeight·recencyPenalty
//
// overlap is weighted above the penalties so a genuine shopping-economy win
// (a shared perishable ingredient) still beats a small repeat-a-cuisine cost,
// while the penalties break what would otherwise be overlap ties toward more
// varied, better-spaced weeks. They are deliberately small relative to one
// overlap point so they never override the ingredient-overlap feature (#258)
// and never the vegetarian quota (which is enforced before scoring).
const (
	overlapReward  = 1.0
	varietyWeight  = 0.34 // per prior use of a shared tag this week
	recencyWeight  = 0.5  // scaled by how recently that tag was last used
	recencyDecaySl = 3    // recency penalty fades to ~0 after this many slots
)

// scheduleTagsExcluded lists tags that describe *when* a dish is eaten or who
// it is for (weekday/weekend/kid-friendly) rather than what it is. Penalizing
// repeats of these would wrongly push the week away from, e.g., weeknight-
// friendly food, so they are excluded from variety/recency scoring. The
// vegetarian tag is also excluded so variety never fights the veg-day quota.
var scheduleTagsExcluded = map[string]bool{
	vegetarianTag: true,
	"vardag":      true,
	"helg":        true,
	"barn":        true,
	"snabb":       true,
	"snabbt":      true,
}

// varietyTags returns the normalized tags of a recipe that count toward the
// variety/recency penalties: cuisine/protein/course markers, with schedule and
// vegetarian tags dropped. Blank tags are ignored.
func varietyTags(r domain.RecipeSummary) []string {
	out := make([]string, 0, len(r.Tags))
	for _, t := range r.Tags {
		n := normalizeTag(t)
		if n == "" || scheduleTagsExcluded[n] {
			continue
		}
		out = append(out, n)
	}
	return out
}

// menuSelector holds the prepared candidate pool for one generate run. It is
// the deterministic "selector core": given a recipe catalog and a household's
// preferences, it (1) hard-filters out recipes carrying any excluded tag, then
// (2) hands back recipes one slot at a time, preferring vegetarian dishes until
// the household's vegetarian-days quota is met, and within each phase greedily
// picking the highest-scoring candidate. The score rewards ingredient overlap
// with the recipes already chosen this week (so the shopping list shares
// ingredients) and penalizes repeating a cuisine/protein (variety) or repeating
// one too soon within the week (in-week recency), so a week is varied and
// similar dishes are spaced out. It cycles through the pool when there are
// fewer recipes than days.
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

	// tags maps recipe ID -> its normalized, overlap-relevant tags (cuisine /
	// protein / course markers like "fisk", "kyckling", "asiatiskt"). These are
	// the only categorical metadata the catalog carries today, so they stand in
	// for "same cuisine / same protein" until Phase 0 adds explicit mainProtein /
	// dietClass fields (issue #248). The vegetarian tag is dropped here so the
	// variety penalty never fights the vegetarian-day quota.
	tags map[string][]string

	// chosenTags counts how many already-picked recipes this week carry each tag,
	// driving the VARIETY penalty (repeating a cuisine/protein costs score).
	// lastUsedTag records the pick index at which each tag was most recently
	// used, driving the in-week RECENCY penalty (a cuisine used last night costs
	// more than one used at the start of the week, so similar dishes spread out).
	chosenTags  map[string]int
	lastUsedTag map[string]int
	pickIndex   int // monotonically increasing slot counter for recency spacing
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

// computeSharedIngredients reports which non-staple ingredients are used by two
// or more of the given recipes, so the UX can surface the shopping-economy
// benefit of overlap selection (#248/#258). recipeIDs is the week's chosen
// recipes (duplicates from cycling are de-duplicated, so a recipe appearing
// twice counts once); ingredientsByID is the same ingredient map the selector
// scored against. Names are matched case-insensitively (same keying as overlap
// scoring) but displayed in their first-seen casing. Pantry staples are
// excluded. The result is sorted most-shared first, then alphabetically for a
// stable order, and is empty when nothing is shared.
func computeSharedIngredients(recipeIDs []string, ingredientsByID map[string][]domain.Ingredient) []domain.SharedIngredient {
	counts := make(map[string]int)      // normalized key -> distinct recipe count
	display := make(map[string]string)  // normalized key -> first-seen display name
	seenRecipe := make(map[string]bool) // de-duplicate cycled recipes
	for _, id := range recipeIDs {
		if id == "" || seenRecipe[id] {
			continue
		}
		seenRecipe[id] = true
		// Count each ingredient once per recipe (a recipe listing an ingredient
		// twice must not look "shared" with itself).
		seenKey := make(map[string]bool)
		for _, ing := range ingredientsByID[id] {
			name := strings.TrimSpace(ing.Name)
			if name == "" || isPantryStaple(name) {
				continue
			}
			key := strings.ToLower(name)
			if seenKey[key] {
				continue
			}
			seenKey[key] = true
			counts[key]++
			if _, ok := display[key]; !ok {
				display[key] = name
			}
		}
	}

	shared := make([]domain.SharedIngredient, 0)
	for key, n := range counts {
		if n >= 2 {
			shared = append(shared, domain.SharedIngredient{Name: display[key], RecipeCount: n})
		}
	}
	sort.Slice(shared, func(i, j int) bool {
		if shared[i].RecipeCount != shared[j].RecipeCount {
			return shared[i].RecipeCount > shared[j].RecipeCount
		}
		return shared[i].Name < shared[j].Name
	})
	return shared
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
		tags:         make(map[string][]string, len(candidates)),
		chosenTags:   make(map[string]int),
		lastUsedTag:  make(map[string]int),
	}
	for _, r := range candidates {
		if isVegetarian(r) {
			sel.veg = append(sel.veg, r)
		} else {
			sel.nonVeg = append(sel.nonVeg, r)
		}
		if vt := varietyTags(r); len(vt) > 0 {
			sel.tags[r.ID] = vt
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

// varietyPenalty returns how many times the candidate's variety tags have
// already been used this week. A dish whose cuisine/protein has not appeared
// yet scores 0; each prior repeat of a shared tag adds 1, so a third fish dish
// is penalized more than the second. Recipes with no variety tags (or before
// any pick) are never penalized.
func (s *menuSelector) varietyPenalty(id string) float64 {
	penalty := 0.0
	for _, t := range s.tags[id] {
		penalty += float64(s.chosenTags[t])
	}
	return penalty
}

// recencyPenalty returns a spacing penalty for repeating a tag that was used
// recently *within this week*. A tag used in the immediately preceding slot
// costs the full weight; the cost decays linearly to 0 once recencyDecaySl
// slots have passed, so similar dishes get pushed apart rather than clustered.
// This is in-week recency only — cross-week recency (recipes used in prior
// weeks) is not scored because the catalog/menu layer exposes no per-recipe
// last-used data today (issue #248 Phase 0 groundwork).
func (s *menuSelector) recencyPenalty(id string) float64 {
	penalty := 0.0
	for _, t := range s.tags[id] {
		last, ok := s.lastUsedTag[t]
		if !ok {
			continue
		}
		gap := s.pickIndex - last
		if gap >= recencyDecaySl {
			continue
		}
		// gap==1 (used last slot) -> full weight; fades with distance.
		penalty += float64(recencyDecaySl-gap) / float64(recencyDecaySl)
	}
	return penalty
}

// score combines the ingredient-overlap reward with the variety and in-week
// recency penalties into the single value pickBest maximizes.
func (s *menuSelector) score(id string) float64 {
	return overlapReward*float64(s.overlapScore(id)) -
		varietyWeight*s.varietyPenalty(id) -
		recencyWeight*s.recencyPenalty(id)
}

// commit records a chosen recipe: it marks the ID as picked, folds its
// ingredient keys into the running week so subsequent picks can score against
// them, and updates the per-tag variety counts and recency markers.
func (s *menuSelector) commit(id string) {
	s.picked[id] = true
	for _, k := range s.ingredients[id] {
		s.chosenKeys[k]++
	}
	s.pickIndex++
	for _, t := range s.tags[id] {
		s.chosenTags[t]++
		s.lastUsedTag[t] = s.pickIndex
	}
}

// pickBest chooses the highest-scoring recipe from candidates, preferring those
// not yet used this week. The score rewards ingredient overlap and penalizes
// repeating a cuisine/protein (variety) or repeating one too soon (in-week
// recency); see score(). Ties (including the very first pick, when nothing has
// been chosen yet) resolve to the candidate earliest in the shuffled order, so
// selection stays deterministic under a seeded RNG. onlyFresh restricts the
// choice to recipes not already picked; when every candidate is already picked
// (the cycling case) it is called again with onlyFresh=false. Returns "" when
// candidates is empty.
func (s *menuSelector) pickBest(candidates []domain.RecipeSummary, onlyFresh bool) string {
	bestID := ""
	var bestScore float64
	for _, r := range candidates {
		if onlyFresh && s.picked[r.ID] {
			continue
		}
		score := s.score(r.ID)
		if bestID == "" || score > bestScore {
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
