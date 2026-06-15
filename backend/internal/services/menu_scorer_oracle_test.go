package services

import (
	"maltiden/internal/domain"
	"math/rand/v2"
	"testing"
)

// oracleScoreWeek is a TEST-ONLY reference scorer. It recomputes the week's
// score from scratch with the same formula the production incremental scorer
// (weekState.marginalDelta / addRecipe) implements, so the two can be
// cross-checked. It deliberately reuses jaccardInts (the production int path)
// for near-duplicate detection rather than a second Jaccard implementation, so
// there is exactly one similarity function to keep correct.
//
// Empty slots ("") are ignored. Higher is better.
func oracleScoreWeek(s *menuSelector, week []string) float64 {
	chosen := make([]string, 0, len(week))
	for _, id := range week {
		if id != "" {
			chosen = append(chosen, id)
		}
	}

	// Overlap: per non-staple canonical appearing in ≥2 recipes, reward (count-1).
	canonCount := make(map[int]int)
	for _, id := range chosen {
		for _, ci := range s.canonInts[id] {
			canonCount[ci]++
		}
	}
	overlap := 0.0
	for _, c := range canonCount {
		if c >= 2 {
			overlap += float64(c - 1)
		}
	}

	// Protein repeats: per non-empty main protein, max(0, count-1).
	protCount := make(map[int]int)
	for _, id := range chosen {
		if pi, ok := s.protInts[id]; ok && pi >= 0 {
			protCount[pi]++
		}
	}
	proteinRepeats := 0.0
	for _, c := range protCount {
		if c > 1 {
			proteinRepeats += float64(c - 1)
		}
	}

	// Duplicate pairs: exact-id repeats + near-duplicate (Jaccard≥0.8) pairs
	// among distinct ids.
	dupPairs := 0.0
	idCount := make(map[string]int)
	for _, id := range chosen {
		idCount[id]++
	}
	for _, c := range idCount {
		if c > 1 {
			dupPairs += float64(c - 1)
		}
	}
	distinct := make([]string, 0, len(idCount))
	for id := range idCount {
		distinct = append(distinct, id)
	}
	for i := 0; i < len(distinct); i++ {
		for j := i + 1; j < len(distinct); j++ {
			if jaccardInts(s.canonInts[distinct[i]], s.canonInts[distinct[j]]) >= nearDuplicateJaccard {
				dupPairs++
			}
		}
	}

	// Recency hits: chosen ids present in the recency set.
	recencyHits := 0.0
	if len(s.recency) > 0 {
		for _, id := range chosen {
			if s.recency[id] {
				recencyHits++
			}
		}
	}

	return domain.OverlapReward*overlap -
		domain.ProteinVarietyPenalty*proteinRepeats -
		domain.DuplicateRecipePenalty*dupPairs -
		domain.RecencyPenalty*recencyHits
}

// incrementalScore replays a week through weekState (the production scorer) and
// returns its accumulated score — the value constructGreedy uses to rank
// restarts.
func incrementalScore(s *menuSelector, week []string) float64 {
	state := s.newWeekState()
	for _, id := range week {
		if id != "" {
			state.addRecipe(id)
		}
	}
	return state.score
}

func TestScorer_IncrementalMatchesOracle(t *testing.T) {
	// nd1/nd2 are near-duplicates (5 shared, Jaccard 5/6 ≈ 0.83 ≥ 0.8) so the
	// exact-repeat + near-duplicate interaction is exercised.
	ndShared := []domain.Ingredient{
		ing("Lök", "lök", false), ing("Vitlök", "vitlök2", false),
		ing("Tomat", "tomat", false), ing("Pasta", "pasta", false),
		ing("Basilika", "basilika", false),
	}
	nd1 := append(append([]domain.Ingredient{}, ndShared...), ing("Oregano", "oregano", false))
	recipes := []domain.Recipe{
		recWith("a", "nöt", "", nil, ing("Lök", "lök", false), ing("Grädde", "grädde", false), ing("Köttfärs", "köttfärs", false)),
		recWith("b", "fläsk", "", nil, ing("Lök", "lök", false), ing("Grädde", "grädde", false), ing("Fläsk", "fläsk", false)),
		recWith("c", "fisk", "", nil, ing("Lax", "lax", false), ing("Citron", "citron", false)),
		recWith("d", "nöt", "", nil, ing("Lök", "lök", false), ing("Tomat", "tomat", false)),
		recWith("e", "", "", nil), // metadata-poor: no non-staple canonicals
		recWith("nd1", "nöt", "", nil, nd1...),
		recWith("nd2", "fågel", "", nil, ndShared...),
	}
	recency := map[string]bool{"c": true}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), recency, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}

	// A spread of weeks: distinct, with empties, with an exact repeat, the
	// metadata-poor recipe alongside others, and near-duplicates mixed with
	// exact repeats of one near-dup partner.
	weeks := [][]string{
		{"a", "b", "c"},
		{"a", "d"},     // share lök but distinct
		{"a", "", "b"}, // empty slot ignored
		{"c", "c"},     // exact repeat
		{"e", "e"},     // two metadata-poor → must NOT be near-duplicates
		{"a", "b", "e"},
		{"a", "b", "c", "d", "e"},
		{"nd1", "nd2"},               // a single near-duplicate pair
		{"nd1", "nd2", "nd1"},        // near-dup pair + exact repeat of nd1
		{"nd1", "nd1", "nd2", "nd2"}, // both partners repeated
	}
	for _, week := range weeks {
		got := incrementalScore(sel, week)
		want := oracleScoreWeek(sel, week)
		if got != want {
			t.Errorf("incremental %v = %v, oracle = %v (week %v)", week, got, want, week)
		}
	}
}

func TestScorer_IncrementalMatchesOracle_Random(t *testing.T) {
	// Build a varied catalog and fuzz random weeks, asserting the incremental
	// scorer and the oracle agree on every one.
	ings := []string{"lök", "grädde", "tomat", "pasta", "ris", "ost", "kyckling", "lax", "köttfärs", "fläsk"}
	proteins := []string{"nöt", "fläsk", "fågel", "fisk", ""}
	var recipes []domain.Recipe
	for i := 0; i < 20; i++ {
		r := recWith(string(rune('A'+i)), proteins[i%len(proteins)], "", nil)
		for j := 0; j <= i%4; j++ {
			c := ings[(i+j)%len(ings)]
			r.Ingredients = append(r.Ingredients, ing(c, c, false))
		}
		recipes = append(recipes, r)
	}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), map[string]bool{"C": true, "H": true}, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}

	ids := make([]string, 0, len(recipes))
	for _, r := range recipes {
		ids = append(ids, r.ID)
	}
	rng := rand.New(rand.NewPCG(7, 11))
	for iter := 0; iter < 200; iter++ {
		n := 1 + rng.IntN(7)
		week := make([]string, n)
		for k := range week {
			if rng.IntN(6) == 0 {
				week[k] = "" // occasional empty slot
			} else {
				week[k] = ids[rng.IntN(len(ids))]
			}
		}
		got := incrementalScore(sel, week)
		want := oracleScoreWeek(sel, week)
		if got != want {
			t.Fatalf("mismatch on week %v: incremental=%v oracle=%v", week, got, want)
		}
	}
}

func TestJaccardInts_EmptySetsAreNotSimilar(t *testing.T) {
	tests := []struct {
		name string
		a, b []int
		want float64
	}{
		{"both empty", nil, nil, 0},
		{"left empty", nil, []int{1, 2}, 0},
		{"right empty", []int{1, 2}, nil, 0},
		{"disjoint", []int{1, 2}, []int{3, 4}, 0},
		{"identical", []int{1, 2, 3}, []int{1, 2, 3}, 1},
		{"half overlap", []int{1, 2}, []int{2, 3}, 1.0 / 3.0},
		{"subset 2 of 3", []int{1, 2}, []int{1, 2, 3}, 2.0 / 3.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jaccardInts(tt.a, tt.b); got != tt.want {
				t.Errorf("jaccardInts(%v,%v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSelectWeek_MetadataPoorRecipesPairFreely(t *testing.T) {
	// Two recipes with NO non-staple canonicals (all-staple / empty) must not be
	// treated as near-duplicates — otherwise the 100-point penalty would forbid
	// putting both in the same week. With 2 slots and only these 2 recipes, both
	// must be placed.
	recipes := []domain.Recipe{
		recWith("poor_1", "nöt", "", nil, ing("Salt", "salt", true)),
		recWith("poor_2", "fisk", "", nil, ing("Olivolja", "olivolja", true)),
	}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	week := sel.SelectWeek(2, []bool{false, false}, []string{"", ""})
	got := map[string]bool{week[0]: true, week[1]: true}
	if !got["poor_1"] || !got["poor_2"] {
		t.Errorf("metadata-poor recipes should pair freely, got %v", week)
	}
	// And the near-dup adjacency must not link them.
	if sel.nearDup["poor_1"]["poor_2"] {
		t.Errorf("metadata-poor recipes wrongly flagged as near-duplicates")
	}
}
