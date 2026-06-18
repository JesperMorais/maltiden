package domain

import "math"

// Macros are per-serving in API responses; computed per-recipe from matched livsmedel.
type Macros struct {
	Kcal    float64 `json:"kcal"`
	Protein float64 `json:"protein"` // grams
	Carbs   float64 `json:"carbs"`   // grams
	Fat     float64 `json:"fat"`     // grams
}

// Livsmedel is one row of Livsmedelsverket's food-composition table (per 100 g).
type Livsmedel struct {
	Livsmedelsnummer int     `json:"livsmedelsnummer"`
	Namn             string  `json:"namn"`
	KcalPer100g      float64 `json:"kcalPer100g"`
	ProteinPer100g   float64 `json:"proteinPer100g"`
	CarbsPer100g     float64 `json:"carbsPer100g"`
	FatPer100g       float64 `json:"fatPer100g"`
}

// MaxProteinTargetPerDay bounds the soft per-day protein target a household may
// set. It is generous enough to cover any realistic goal while rejecting clearly
// nonsensical input.
const MaxProteinTargetPerDay = 500

// IsValidProteinTarget reports whether a per-day protein target is in the
// accepted range [0, MaxProteinTargetPerDay]. Zero means "no target".
func IsValidProteinTarget(t float64) bool {
	return t >= 0 && t <= MaxProteinTargetPerDay
}

// round1 rounds to one decimal so persisted/returned macros stay tidy and
// deterministic across the compute → store → read round-trip.
func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

// ComputeRecipeMacros sums per-serving macros for a recipe from its matched
// livsmedel rows. For each ingredient with a non-zero Livsmedelsnummer, a
// positive GramsEquiv, and a row present in table, it adds
// per100g * GramsEquiv / 100 to the recipe totals, then divides by Servings
// (treating Servings < 1 as 1). It returns nil when no ingredient contributed,
// so un-enriched recipes carry nil macros (graceful). Pure, no I/O.
func ComputeRecipeMacros(r Recipe, table map[int]Livsmedel) *Macros {
	var total Macros
	matched := false
	for _, ing := range r.Ingredients {
		if ing.Livsmedelsnummer == 0 || ing.GramsEquiv <= 0 {
			continue
		}
		lv, ok := table[ing.Livsmedelsnummer]
		if !ok {
			continue
		}
		factor := ing.GramsEquiv / 100
		total.Kcal += lv.KcalPer100g * factor
		total.Protein += lv.ProteinPer100g * factor
		total.Carbs += lv.CarbsPer100g * factor
		total.Fat += lv.FatPer100g * factor
		matched = true
	}
	if !matched {
		return nil
	}

	servings := r.Servings
	if servings < 1 {
		servings = 1
	}
	div := float64(servings)
	return &Macros{
		Kcal:    round1(total.Kcal / div),
		Protein: round1(total.Protein / div),
		Carbs:   round1(total.Carbs / div),
		Fat:     round1(total.Fat / div),
	}
}
