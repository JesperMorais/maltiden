package services

import (
	"maltiden/internal/domain"
	"reflect"
	"testing"
)

// recBatch builds a batchable recipe with one distinguishing ingredient so it is
// not a near-duplicate of the others.
func recBatch(id, protein string, batchable bool, canonical string) domain.Recipe {
	r := recWith(id, protein, "", nil, ing(canonical, canonical, false))
	r.Batchable = batchable
	return r
}

// allWeekdays returns a weekday slice where every slot ranks as "any" (Tuesday),
// so window ranking falls back purely to lowest cook-slot index.
func allWeekdays(n, day int) []int {
	w := make([]int, n)
	for i := range w {
		w[i] = day
	}
	return w
}

func TestSelectWeekPrep_PlacesBatchPairOnConsecutiveSlots(t *testing.T) {
	recipes := []domain.Recipe{
		recBatch("batch_1", "nöt", true, "köttfärs"),
		recBatch("free_1", "fisk", false, "lax"),
		recBatch("free_2", "fågel", false, "kyckling"),
		recBatch("free_3", "fläsk", false, "korv"),
	}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	week, pairs := sel.SelectWeekPrep(4, make([]bool, 4), make([]string, 4), allWeekdays(4, 2), true)
	if len(pairs) != 1 {
		t.Fatalf("expected exactly 1 batch pair, got %d (%v)", len(pairs), pairs)
	}
	p := pairs[0]
	if p.LeftoverSlot != p.CookSlot+1 {
		t.Errorf("pair must span consecutive slots, got cook=%d leftover=%d", p.CookSlot, p.LeftoverSlot)
	}
	if week[p.CookSlot] != p.RecipeID || week[p.LeftoverSlot] != p.RecipeID {
		t.Errorf("both pair slots must hold the batch recipe %q, got %v", p.RecipeID, week)
	}
	if r := sel.byID[p.RecipeID]; !r.Batchable {
		t.Errorf("pair recipe %q must be batchable", p.RecipeID)
	}
	// Every slot filled.
	for i, id := range week {
		if id == "" {
			t.Errorf("slot %d empty in %v", i, week)
		}
	}
}

func TestSelectWeekPrep_PrefersSundayThenMondayCookDay(t *testing.T) {
	recipes := []domain.Recipe{
		recBatch("batch_1", "nöt", true, "köttfärs"),
		recBatch("free_1", "fisk", false, "lax"),
		recBatch("free_2", "fågel", false, "kyckling"),
		recBatch("free_3", "fläsk", false, "korv"),
		recBatch("free_4", "lamm", false, "lammfärs"),
	}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}

	// weekday: slot0=Fri(5) slot1=Sat(6) slot2=Sun(0) slot3=Mon(1) slot4=Tue(2).
	// Sunday cook day is slot 2.
	weekday := []int{5, 6, 0, 1, 2}
	_, pairs := sel.SelectWeekPrep(5, make([]bool, 5), make([]string, 5), weekday, true)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	if pairs[0].CookSlot != 2 {
		t.Errorf("expected Sunday (slot 2) cook day, got cook=%d", pairs[0].CookSlot)
	}

	// No Sunday window: slot weekdays Sat,Mon,Tue,Wed → Monday(slot1) preferred.
	weekday = []int{6, 1, 2, 3}
	_, pairs = sel.SelectWeekPrep(4, make([]bool, 4), make([]string, 4), weekday, true)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	if pairs[0].CookSlot != 1 {
		t.Errorf("expected Monday (slot 1) cook day, got cook=%d", pairs[0].CookSlot)
	}
}

func TestSelectWeekPrep_NoBatchableIsNoOp(t *testing.T) {
	recipes := []domain.Recipe{
		recBatch("free_1", "fisk", false, "lax"),
		recBatch("free_2", "fågel", false, "kyckling"),
		recBatch("free_3", "fläsk", false, "korv"),
	}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	week, pairs := sel.SelectWeekPrep(4, make([]bool, 4), make([]string, 4), allWeekdays(4, 2), true)
	if len(pairs) != 0 {
		t.Errorf("no batchable recipe → no pairs, got %v", pairs)
	}
	for i, id := range week {
		if id == "" {
			t.Errorf("slot %d empty in %v", i, week)
		}
	}
}

func TestSelectWeekPrep_NoConsecutiveWindowIsNoOp(t *testing.T) {
	recipes := []domain.Recipe{
		recBatch("batch_1", "nöt", true, "köttfärs"),
		recBatch("free_1", "fisk", false, "lax"),
		recBatch("free_2", "fågel", false, "kyckling"),
	}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	// Skip the middle slot so no two open slots are consecutive (open: 0 and 2).
	week, pairs := sel.SelectWeekPrep(3, []bool{false, true, false}, make([]string, 3), allWeekdays(3, 2), true)
	if len(pairs) != 0 {
		t.Errorf("no consecutive open window → no pairs, got %v", pairs)
	}
	if week[1] != "" {
		t.Errorf("skip slot must stay empty, got %q", week[1])
	}
}

func TestSelectWeekPrep_OffEqualsPhase1(t *testing.T) {
	recipes := []domain.Recipe{
		recBatch("batch_1", "nöt", true, "köttfärs"),
		recBatch("batch_2", "fågel", true, "kyckling"),
		recBatch("free_1", "fisk", false, "lax"),
		recBatch("free_2", "fläsk", false, "korv"),
		recBatch("free_3", "lamm", false, "lammfärs"),
	}
	// Same seed both ways.
	sel1, _ := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	sel2, _ := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())

	phase1 := sel1.SelectWeek(5, make([]bool, 5), make([]string, 5))
	prepOff, pairs := sel2.SelectWeekPrep(5, make([]bool, 5), make([]string, 5), allWeekdays(5, 0), false)
	if len(pairs) != 0 {
		t.Errorf("prep=false must place no pairs, got %v", pairs)
	}
	if !reflect.DeepEqual(phase1, prepOff) {
		t.Errorf("prep off must be byte-identical to Phase 1:\n  Phase1 = %v\n  prepOff = %v", phase1, prepOff)
	}
}

func TestSelectWeekPrep_Deterministic(t *testing.T) {
	recipes := []domain.Recipe{
		recBatch("batch_1", "nöt", true, "köttfärs"),
		recBatch("batch_2", "fågel", true, "kyckling"),
		recBatch("free_1", "fisk", false, "lax"),
		recBatch("free_2", "fläsk", false, "korv"),
	}
	run := func() ([]string, []PrepPair) {
		sel, _ := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
		week, pairs := sel.SelectWeekPrep(5, make([]bool, 5), make([]string, 5), allWeekdays(5, 0), true)
		return week, pairs
	}
	w1, p1 := run()
	w2, p2 := run()
	if !reflect.DeepEqual(w1, w2) || !reflect.DeepEqual(p1, p2) {
		t.Errorf("same seed must yield identical week+pairs:\n  %v / %v\n  %v / %v", w1, p1, w2, p2)
	}
}

func TestSelectWeekPrep_LocksNeverSpanned(t *testing.T) {
	recipes := []domain.Recipe{
		recBatch("batch_1", "nöt", true, "köttfärs"),
		recBatch("locked", "fisk", false, "lax"),
		recBatch("free_1", "fågel", false, "kyckling"),
		recBatch("free_2", "fläsk", false, "korv"),
	}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	// Lock slot 1; the only consecutive all-open window is slots 2,3.
	locked := []string{"", "locked", "", ""}
	week, pairs := sel.SelectWeekPrep(4, make([]bool, 4), locked, allWeekdays(4, 2), true)
	if week[1] != "locked" {
		t.Errorf("lock must be honored, got %q", week[1])
	}
	for _, p := range pairs {
		if p.CookSlot == 1 || p.LeftoverSlot == 1 {
			t.Errorf("pair must not span the locked slot, got %+v", p)
		}
	}
}

func TestSelectWeekPrep_LockedBatchableStaysSingleDay(t *testing.T) {
	recipes := []domain.Recipe{
		recBatch("batch_locked", "nöt", true, "köttfärs"),
		recBatch("batch_free", "fågel", true, "kyckling"),
		recBatch("free_1", "fisk", false, "lax"),
		recBatch("free_2", "fläsk", false, "korv"),
	}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	// Lock a batchable recipe on slot 0; it must NOT be auto-expanded into a pair.
	locked := []string{"batch_locked", "", "", ""}
	week, pairs := sel.SelectWeekPrep(4, make([]bool, 4), locked, allWeekdays(4, 2), true)
	if week[0] != "batch_locked" {
		t.Errorf("locked batchable must stay on its slot, got %q", week[0])
	}
	// The locked batchable must appear exactly once across the week.
	count := 0
	for _, id := range week {
		if id == "batch_locked" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("locked batchable must stay single-day, appeared %d times in %v", count, week)
	}
	// Any pair placed must be the other batchable, not the locked one.
	for _, p := range pairs {
		if p.RecipeID == "batch_locked" {
			t.Errorf("locked batchable must not be auto-expanded into a pair: %+v", p)
		}
	}
}

func TestSelectWeekPrep_OddDaysGraceful(t *testing.T) {
	recipes := []domain.Recipe{
		recBatch("batch_1", "nöt", true, "köttfärs"),
		recBatch("free_1", "fisk", false, "lax"),
		recBatch("free_2", "fågel", false, "kyckling"),
	}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	// 1 day: no consecutive window possible → no pair, single filled slot.
	week, pairs := sel.SelectWeekPrep(1, make([]bool, 1), make([]string, 1), allWeekdays(1, 2), true)
	if len(pairs) != 0 {
		t.Errorf("1 day cannot host a pair, got %v", pairs)
	}
	if week[0] == "" {
		t.Errorf("single slot should be filled, got empty")
	}
}

func TestSelectWeekPrep_BatchPairNotDuplicatePenalized(t *testing.T) {
	// A batch pair places the SAME recipe on two slots but must be scored as one
	// dish (no duplicate penalty), so the resulting week's incremental score must
	// match the oracle scored over DISTINCT recipes (counting the batch once).
	recipes := []domain.Recipe{
		recBatch("batch_1", "nöt", true, "köttfärs"),
		recBatch("free_1", "fisk", false, "lax"),
		recBatch("free_2", "fågel", false, "kyckling"),
		recBatch("free_3", "fläsk", false, "korv"),
	}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	week, pairs := sel.SelectWeekPrep(4, make([]bool, 4), make([]string, 4), allWeekdays(4, 2), true)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	// Build the deduped week (collapse the batch pair to one occurrence) and score
	// it with the oracle; the batch must not incur the -100 duplicate penalty.
	deduped := make([]string, 0, len(week))
	seenBatch := false
	for _, id := range week {
		if id == pairs[0].RecipeID {
			if seenBatch {
				continue
			}
			seenBatch = true
		}
		deduped = append(deduped, id)
	}
	score := oracleScoreWeek(sel, deduped)
	if score <= -domain.DuplicateRecipePenalty {
		t.Errorf("batch pair appears duplicate-penalized; deduped oracle score = %v", score)
	}
}

// vegBatch builds a batchable VEGETARIAN recipe (carries the veg tag so the
// quota counts it) with a distinguishing ingredient.
func vegBatch(id, canonical string) domain.Recipe {
	r := recWith(id, "", "", []string{vegetarianTag}, ing(canonical, canonical, false))
	r.Batchable = true
	return r
}

func TestSelectWeekPrep_VegBatchPairSatisfiesTwoVegDays(t *testing.T) {
	// A single VEGETARIAN batch pair fills two vegetarian days (cook + leftover,
	// same veg recipe). With VegetarianDays=2 it must fully satisfy the quota so
	// the selector reserves ZERO extra veg slots. To make the reservation visible
	// independent of scoring, there is exactly ONE batchable veg recipe (chosen for
	// the pair) and one MORE veg recipe; the open slots are filled from a pool of
	// distinct-protein meats. With the correct credit-of-2 the extra veg recipe is
	// not forced, so the only veg days come from the pair (== 2). With the buggy
	// credit-of-1, vegNeeded would stay 1 and force the extra veg recipe → 3.
	// veg_extra is a near-duplicate of veg_batch (identical non-staple ingredient
	// set), so the greedy never picks it on score — pairing the two veg recipes
	// would cost the -100 duplicate penalty. It can therefore only land in the week
	// if a veg slot is wrongly RESERVED for it, which is exactly the bug under test.
	recipes := []domain.Recipe{
		vegBatch("veg_batch", "linser"),
		recWith("veg_extra", "", domain.DietClassVegetarian, []string{vegetarianTag}, ing("Linser", "linser", false)),
		recBatch("meat_1", "nöt", false, "köttfärs"),
		recBatch("meat_2", "fisk", false, "lax"),
		recBatch("meat_3", "fågel", false, "kyckling"),
		recBatch("meat_4", "fläsk", false, "fläskkarré"),
		recBatch("meat_5", "lamm", false, "lammfärs"),
	}
	prefs := defaultPrefs()
	prefs.VegetarianDays = 2

	sel, ok := newMenuSelector(recipes, prefs, nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}

	// 5 days: pair lands at the lowest open window (slots 0,1); 3 open slots remain
	// and are filled from the 5 distinct-protein meats, so the second veg recipe is
	// only placed if a veg slot was wrongly reserved.
	week, pairs := sel.SelectWeekPrep(5, make([]bool, 5), make([]string, 5), allWeekdays(5, 2), true)
	if len(pairs) != 1 {
		t.Fatalf("expected exactly 1 batch pair, got %d (%v)", len(pairs), pairs)
	}
	if pairs[0].RecipeID != "veg_batch" {
		t.Fatalf("expected the vegetarian batchable to be chosen, got %q", pairs[0].RecipeID)
	}

	// Count vegetarian DAYS (slots holding a veg recipe). The batch pair already
	// contributes 2; the quota of 2 is met, so no extra veg recipe is reserved →
	// exactly 2 veg days total (the pair's cook + leftover).
	vegDays := 0
	for _, id := range week {
		if r, ok := sel.byID[id]; ok && isVegetarian(r) {
			vegDays++
		}
	}
	if vegDays != 2 {
		t.Errorf("veg batch pair should satisfy exactly 2 veg days (no reserved extra), got %d veg days in %v", vegDays, week)
	}
}
