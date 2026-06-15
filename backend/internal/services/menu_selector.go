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

// nearDuplicateJaccard is the ingredient-set Jaccard threshold at or above
// which two distinct recipes are treated as near-duplicates and penalized like
// an exact repeat (e.g. two slightly different "korv stroganoff" variants).
const nearDuplicateJaccard = 0.8

// fallbackStaples is the hardcoded pantry-staple list used for recipes whose
// ingredients carry no enriched metadata (IsPantryStaple all false). It is a
// best-effort approximation so overlap and items-to-buy stay meaningful even
// for un-enriched seed recipes. Entries are normalized canonical names.
var fallbackStaples = map[string]bool{
	"salt":      true,
	"peppar":    true,
	"vatten":    true,
	"olja":      true,
	"olivolja":  true,
	"rapsolja":  true,
	"smör":      true,
	"socker":    true,
	"mjöl":      true,
	"vetemjöl":  true,
	"buljong":   true,
	"vitlök":    true,
}

// menuSelector holds the prepared candidate pool for one generate run. It is
// the deterministic "selector core": given a recipe catalog (with ingredients)
// and a household's preferences, it hard-filters out disallowed recipes, then
// builds the best week it can via randomized-restart greedy construction,
// scoring for shared-ingredient overlap, protein variety, recipe duplication,
// and recency.
//
// It carries no I/O and no clock, so it is fully unit-testable with a seeded RNG.
type menuSelector struct {
	rng *rand.Rand

	veg  []string // vegetarian candidate ids
	pool []string // veg + nonVeg ids, the cycling fallback order

	byID      map[string]domain.Recipe   // candidate recipes by id
	canonInts map[string][]int           // recipe id → non-staple canonicals as sorted dense int ids
	canonID   map[string]int             // canonical name → dense int id
	protInts  map[string]int             // recipe id → dense protein int id (-1 = none)
	nProteins int                        // number of distinct proteins (for count array sizing)
	nearDup   map[string]map[string]bool // id → set of ids that are near-duplicates (Jaccard≥0.8)

	vegRemaining int // vegetarian slots still owed by the preference
	recency      map[string]bool
}

// filterByExcludedTags returns the subset of recipes that carry none of the
// excluded tags. Tag matching is case-insensitive and trims surrounding space
// so user-entered preferences ("Fisk ") match seed tags ("fisk").
func filterByExcludedTags(recipes []domain.Recipe, excludedTags []string) []domain.Recipe {
	if len(excludedTags) == 0 {
		out := make([]domain.Recipe, len(recipes))
		copy(out, recipes)
		return out
	}

	excluded := make(map[string]bool, len(excludedTags))
	for _, t := range excludedTags {
		excluded[normalizeTag(t)] = true
	}

	out := make([]domain.Recipe, 0, len(recipes))
	for _, r := range recipes {
		if recipeHasExcludedTag(r, excluded) {
			continue
		}
		out = append(out, r)
	}
	return out
}

func recipeHasExcludedTag(r domain.Recipe, excluded map[string]bool) bool {
	for _, t := range r.Tags {
		if excluded[normalizeTag(t)] {
			return true
		}
	}
	return false
}

// isVegetarian reports whether a recipe counts toward the vegetarian-day quota.
// A recipe qualifies if it carries the Swedish "vegetariskt" tag OR its
// DietClass is vegetarian/vegan. Relying on DietClass too means the quota still
// works for households whose recipes have structured metadata but lack the tag
// (otherwise the veg quota silently no-ops for them).
func isVegetarian(r domain.Recipe) bool {
	switch r.DietClass {
	case domain.DietClassVegetarian, domain.DietClassVegan:
		return true
	}
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

// effectiveCanonical returns the canonical name used for overlap matching:
// CanonicalName when the parser populated it, else the lowercased/trimmed raw
// name as a graceful fallback.
func effectiveCanonical(ing domain.Ingredient) string {
	if c := strings.TrimSpace(ing.CanonicalName); c != "" {
		return strings.ToLower(c)
	}
	return strings.ToLower(strings.TrimSpace(ing.Name))
}

// recipeHasMetadata reports whether a recipe's ingredient metadata has been
// enriched. We treat any populated CanonicalName or IsPantryStaple flag as the
// signal that the staple flags are meaningful; otherwise we fall back to the
// hardcoded staple list.
func recipeHasMetadata(r domain.Recipe) bool {
	for _, ing := range r.Ingredients {
		if ing.IsPantryStaple || strings.TrimSpace(ing.CanonicalName) != "" {
			return true
		}
	}
	return false
}

// effectiveStaple reports whether an ingredient should be treated as a pantry
// staple (and thus excluded from overlap and items-to-buy). When the recipe has
// enriched metadata it trusts IsPantryStaple; otherwise it falls back to
// membership in the hardcoded staple list.
func effectiveStaple(ing domain.Ingredient, hasMeta bool) bool {
	if hasMeta {
		return ing.IsPantryStaple
	}
	return fallbackStaples[effectiveCanonical(ing)]
}

// nonStapleCanonicals returns the set of non-staple canonical names for a
// recipe. The result is a set (duplicates within one recipe collapse).
func nonStapleCanonicals(r domain.Recipe) map[string]struct{} {
	hasMeta := recipeHasMetadata(r)
	set := make(map[string]struct{}, len(r.Ingredients))
	for _, ing := range r.Ingredients {
		if effectiveStaple(ing, hasMeta) {
			continue
		}
		c := effectiveCanonical(ing)
		if c == "" {
			continue
		}
		set[c] = struct{}{}
	}
	return set
}

// passesHardFilters applies the diet-profile and disliked-ingredient hard
// filters. (Excluded tags are applied separately, up front.)
func passesHardFilters(r domain.Recipe, prefs domain.MenuPreferences, disliked map[string]bool) bool {
	if !domain.DietCompatible(prefs.DietProfile, r.DietClass) {
		return false
	}
	if len(disliked) > 0 {
		for _, ing := range r.Ingredients {
			if disliked[effectiveCanonical(ing)] {
				return false
			}
		}
	}
	return true
}

// newMenuSelector builds a selector from the catalog and preferences. recency
// is the set of recently-used recipe ids (may be nil). rng is injected so tests
// can be deterministic. Locks are honored only when they reference a recipe in
// the candidate set (the household's catalog), so every locked recipe is also a
// registered candidate. Returns (nil, false) when no candidate survives the
// hard filters — the caller treats that as "no recipes available", mirroring
// the empty-catalog case.
func newMenuSelector(recipes []domain.Recipe, prefs domain.MenuPreferences, recency map[string]bool, rng *rand.Rand) (*menuSelector, bool) {
	candidates := filterByExcludedTags(recipes, prefs.ExcludedTags)

	disliked := make(map[string]bool, len(prefs.DislikedIngredients))
	for _, d := range prefs.DislikedIngredients {
		if n := normalizeTag(d); n != "" {
			disliked[n] = true
		}
	}

	filtered := make([]domain.Recipe, 0, len(candidates))
	for _, r := range candidates {
		if passesHardFilters(r, prefs, disliked) {
			filtered = append(filtered, r)
		}
	}
	if len(filtered) == 0 {
		return nil, false
	}

	sel := &menuSelector{
		rng:          rng,
		vegRemaining: prefs.VegetarianDays,
		recency:      recency,
		byID:         make(map[string]domain.Recipe, len(filtered)),
		canonInts:    make(map[string][]int, len(filtered)),
		canonID:      make(map[string]int),
		protInts:     make(map[string]int, len(filtered)),
	}

	protID := make(map[string]int)
	// register populates the per-recipe caches the selector reasons about: the
	// recipe by id, its non-staple canonicals as sorted dense int ids (hot-path
	// counting without string hashing; sorted so the near-duplicate precompute
	// can intersect by merge), and its protein int id.
	register := func(r domain.Recipe) {
		if _, done := sel.byID[r.ID]; done {
			return
		}
		sel.byID[r.ID] = r

		set := nonStapleCanonicals(r)
		ints := make([]int, 0, len(set))
		for c := range set {
			id, ok := sel.canonID[c]
			if !ok {
				id = len(sel.canonID)
				sel.canonID[c] = id
			}
			ints = append(ints, id)
		}
		sort.Ints(ints)
		sel.canonInts[r.ID] = ints

		p := normalizeTag(r.MainProtein)
		if p == "" {
			sel.protInts[r.ID] = -1
		} else {
			id, ok := protID[p]
			if !ok {
				id = len(protID)
				protID[p] = id
			}
			sel.protInts[r.ID] = id
		}
	}

	var nonVeg []string
	for _, r := range filtered {
		register(r)
		if isVegetarian(r) {
			sel.veg = append(sel.veg, r.ID)
		} else {
			nonVeg = append(nonVeg, r.ID)
		}
	}
	sel.nProteins = len(protID)

	// Cycling pool prefers vegetarian first so the quota can be honored even
	// when veg recipes are scarce, then falls back to the full set.
	sel.pool = append(sel.pool, sel.veg...)
	sel.pool = append(sel.pool, nonVeg...)

	// Precompute the near-duplicate adjacency once over every registered recipe
	// (O(n²) Jaccard, but only here — not per greedy step), so scoring during
	// construction is a cheap map lookup instead of recomputing Jaccard for
	// every candidate/slot.
	allIDs := make([]string, 0, len(sel.byID))
	for id := range sel.byID {
		allIDs = append(allIDs, id)
	}
	sel.nearDup = make(map[string]map[string]bool, len(allIDs))
	for i := 0; i < len(allIDs); i++ {
		for j := i + 1; j < len(allIDs); j++ {
			a, b := allIDs[i], allIDs[j]
			if jaccardInts(sel.canonInts[a], sel.canonInts[b]) >= nearDuplicateJaccard {
				if sel.nearDup[a] == nil {
					sel.nearDup[a] = make(map[string]bool)
				}
				if sel.nearDup[b] == nil {
					sel.nearDup[b] = make(map[string]bool)
				}
				sel.nearDup[a][b] = true
				sel.nearDup[b][a] = true
			}
		}
	}

	return sel, true
}

func shuffle(s []string, rng *rand.Rand) {
	if rng == nil {
		rand.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
		return
	}
	rng.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
}

// canonIntsFor returns the dense int canonical ids for a recipe. Registered
// recipes are precomputed; for anything else (defensive) it returns nil so the
// recipe contributes no overlap rather than panicking.
func (s *menuSelector) canonIntsFor(id string) []int {
	return s.canonInts[id]
}

// protIntFor returns the dense protein int id for a recipe (-1 if none/unknown).
func (s *menuSelector) protIntFor(id string) int {
	if pi, ok := s.protInts[id]; ok {
		return pi
	}
	return -1
}

// SelectWeek builds the best week for the given open/fixed-slot layout.
//
//   - slots is the number of menu days (including skips and locks).
//   - skip[i] true marks a skipped day (left empty, unscored).
//   - locked[i] non-empty pins a recipe id to that slot (kept, not replaced, but
//     counted in scoring); an unknown id is treated as unlocked.
//
// It returns the chosen recipe id per slot ("" for skipped slots). Locks and
// the cycling fallback for fewer-recipes-than-days are preserved.
func (s *menuSelector) SelectWeek(slots int, skip []bool, locked []string) []string {
	// Classify slots.
	openIdx := make([]int, 0, slots)
	lockedAtSlot := make(map[int]string)
	for i := 0; i < slots; i++ {
		if i < len(skip) && skip[i] {
			continue
		}
		if i < len(locked) && locked[i] != "" {
			if _, ok := s.byID[locked[i]]; ok {
				lockedAtSlot[i] = locked[i]
				continue
			}
			// Unknown locked id → treat as a normal open slot.
		}
		openIdx = append(openIdx, i)
	}

	// Vegetarian quota covered by locked slots reduces what's still owed.
	vegFromLocks := 0
	for _, id := range lockedAtSlot {
		if r, ok := s.byID[id]; ok && isVegetarian(r) {
			vegFromLocks++
		}
	}
	vegNeeded := s.vegRemaining - vegFromLocks
	if vegNeeded < 0 {
		vegNeeded = 0
	}
	if vegNeeded > len(openIdx) {
		vegNeeded = len(openIdx)
	}
	if vegNeeded > len(s.veg) {
		vegNeeded = len(s.veg)
	}

	best := []string(nil)
	bestScore := 0.0
	haveBest := false

	for restart := 0; restart < domain.SelectorRestarts; restart++ {
		week, score := s.constructGreedy(slots, openIdx, lockedAtSlot, vegNeeded, restart)
		if !haveBest || score > bestScore {
			best = week
			bestScore = score
			haveBest = true
		}
	}

	return best
}

// weekState carries the running aggregates needed to score a (partial) week
// incrementally as recipes are added, so the greedy marginal delta is O(small)
// per candidate instead of an O(n²) full rescore. The aggregates implement the
// same scoring formula as the test oracle (oracleScoreWeek in the _test.go),
// which cross-checks them; addRecipe/marginalDelta keep score in lockstep.
type weekState struct {
	sel        *menuSelector
	canonCount []int          // dense canonical int id → occurrences across chosen
	protCount  []int          // dense protein int id → occurrences
	usedCount  map[string]int // recipe id → times chosen (distinct priors + exact-repeat detection)
	score      float64
}

func (s *menuSelector) newWeekState() *weekState {
	return &weekState{
		sel:        s,
		canonCount: make([]int, len(s.canonID)),
		protCount:  make([]int, s.nProteins),
		usedCount:  make(map[string]int),
	}
}

// marginalDelta returns the change in score from adding recipe id to the
// current state, without mutating it.
func (w *weekState) marginalDelta(id string) float64 {
	sel := w.sel
	delta := 0.0

	// Overlap: a canonical already present (count≥1) gains +1 to its
	// max(0,count-1) contribution when its count rises.
	for _, ci := range sel.canonIntsFor(id) {
		if ci >= 0 && ci < len(w.canonCount) && w.canonCount[ci] >= 1 {
			delta += domain.OverlapReward
		}
	}

	// Protein repeat: adding a protein already present adds one repeat.
	if pi := sel.protIntFor(id); pi >= 0 && pi < len(w.protCount) && w.protCount[pi] >= 1 {
		delta -= domain.ProteinVarietyPenalty
	}

	// Duplicates: adding this recipe creates new duplicate "pairs" worth one
	// penalty each, matching the batch oracle (which counts, over the DISTINCT
	// chosen ids, exact repeats as count-1 plus one per near-duplicate pair):
	//   - exact repeat: if id is already present, this extra occurrence adds one
	//     (a recipe used n times contributes n-1 across its n adds).
	//   - near-duplicate: a distinct pair {id, prev} is "created" exactly once,
	//     when id is FIRST added; count one per distinct prior that is a
	//     near-duplicate of id. Re-adding an existing id creates no new pair.
	// The near-dup scan is skipped unless id has near-duplicates at all, and the
	// exact-repeat check is a single map lookup — keeping the hot path cheap.
	dup := 0.0
	firstAdd := w.usedCount[id] == 0
	if !firstAdd {
		dup++ // exact repeat: this is one extra occurrence
	}
	if firstAdd {
		if adj := sel.nearDup[id]; adj != nil { // nil when id has no near-duplicates
			for prev := range w.usedCount {
				if prev != id && adj[prev] {
					dup++
				}
			}
		}
	}
	delta -= domain.DuplicateRecipePenalty * dup

	// Recency.
	if len(sel.recency) > 0 && sel.recency[id] {
		delta -= domain.RecencyPenalty
	}

	return delta
}

// addRecipe commits a recipe into the running state, updating aggregates and
// the cached score.
func (w *weekState) addRecipe(id string) {
	w.score += w.marginalDelta(id)
	for _, ci := range w.sel.canonIntsFor(id) {
		if ci >= 0 && ci < len(w.canonCount) {
			w.canonCount[ci]++
		}
	}
	if pi := w.sel.protIntFor(id); pi >= 0 && pi < len(w.protCount) {
		w.protCount[pi]++
	}
	w.usedCount[id]++
}

// constructGreedy builds one candidate week and returns it with its score. It
// pre-places locked recipes, then fills open slots — vegetarian-reserved ones
// first — by argmax marginal score delta over a per-restart shuffled candidate
// order. Exact duplicates are only allowed once the unique pool is exhausted
// (preserving the cycling fallback for fewer-recipes-than-days).
func (s *menuSelector) constructGreedy(slots int, openIdx []int, lockedAtSlot map[int]string, vegNeeded, restart int) ([]string, float64) {
	week := make([]string, slots)
	state := s.newWeekState()
	for slot, id := range lockedAtSlot {
		week[slot] = id
		state.addRecipe(id)
	}

	// Per-restart deterministic sub-stream so each restart explores a different
	// shuffle while the overall run stays reproducible for a given seed.
	var rng *rand.Rand
	if s.rng != nil {
		rng = rand.New(rand.NewPCG(s.rng.Uint64(), uint64(restart)+1))
	}

	veg := make([]string, len(s.veg))
	copy(veg, s.veg)
	shuffle(veg, rng)
	pool := make([]string, len(s.pool))
	copy(pool, s.pool)
	shuffle(pool, rng)

	used := make(map[string]bool)
	for _, id := range lockedAtSlot {
		used[id] = true
	}

	// Split open slots: reserve the first vegNeeded for vegetarian candidates.
	vegSlots := append([]int(nil), openIdx[:vegNeeded]...)
	openSlots := append([]int(nil), openIdx[vegNeeded:]...)

	// Per-candidate-list cycle index used only by the fallback when every
	// candidate is already used (catalog smaller than the number of slots). It
	// advances each fallback pick so duplicates spread evenly across the
	// shuffled list — round-robin — instead of clustering on candidates[0].
	cycle := make(map[*[]string]int)
	fill := func(slot int, candidates *[]string) {
		list := *candidates
		// Prefer an unused candidate that maximizes the marginal score.
		bestID := ""
		bestDelta := 0.0
		found := false
		for _, id := range list {
			if used[id] {
				continue
			}
			delta := state.marginalDelta(id)
			if !found || delta > bestDelta {
				bestID = id
				bestDelta = delta
				found = true
			}
		}
		if !found && len(list) > 0 {
			// Pool exhausted (fewer recipes than slots): cycle round-robin over
			// the shuffled list so duplicates are distributed evenly.
			bestID = list[cycle[candidates]%len(list)]
			cycle[candidates]++
		}
		if bestID != "" {
			week[slot] = bestID
			used[bestID] = true
			state.addRecipe(bestID)
		}
	}

	for _, slot := range vegSlots {
		fill(slot, &veg)
	}
	for _, slot := range openSlots {
		fill(slot, &pool)
	}

	return week, state.score
}

// jaccardInts returns the Jaccard similarity of two SORTED int-id slices via a
// linear merge — no map allocation or hashing. Used in the O(n²) near-duplicate
// precompute where it is the hot path.
//
// If EITHER set is empty the similarity is 0, never 1: a recipe with no
// non-staple canonicals (metadata-poor, or all-staple) shares nothing concrete
// with anything, so it must not be flagged a near-duplicate (which would slap a
// 100-point penalty on pairing it with another such recipe and effectively
// forbid metadata-poor recipes from coexisting in a week).
func jaccardInts(a, b []int) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	inter := 0
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			inter++
			i++
			j++
		case a[i] < b[j]:
			i++
		default:
			j++
		}
	}
	union := len(a) + len(b) - inter
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}
