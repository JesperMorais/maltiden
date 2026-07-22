package services

import (
	"maltiden/internal/domain"
	"math/rand/v2"
)

const menuSelectorRestarts = 8

// selectRecipesForWeek picks one recipe per non-skip slot, maximizing shared
// (non-pantry) ingredient overlap across the week while penalizing tags that
// would appear more than twice. It runs several randomized restarts and
// keeps the best-scoring week.
//
// recipes is the candidate pool (cycled if fewer than slotCount). skipSlots
// marks indices that don't need a recipe; those positions are left as "" in
// the returned slice. recipeData supplies ingredient/tag lookups keyed by
// recipe ID.
func selectRecipesForWeek(recipes []domain.RecipeSummary, slotCount int, skipSlots map[int]bool, recipeData map[string]*domain.Recipe, seed uint64) []string {
	if len(recipes) == 0 || slotCount == 0 {
		return make([]string, slotCount)
	}

	var best []string
	bestScore := -1.0

	for restart := 0; restart < menuSelectorRestarts; restart++ {
		rng := rand.New(rand.NewPCG(seed, uint64(restart)))

		order := make([]int, len(recipes))
		for i := range order {
			order[i] = i
		}
		rng.Shuffle(len(order), func(i, j int) {
			order[i], order[j] = order[j], order[i]
		})

		candidate, score := greedyFill(recipes, order, slotCount, skipSlots, recipeData, rng)
		if score > bestScore {
			bestScore = score
			best = candidate
		}
	}

	return best
}

// greedyFill fills each non-skip slot in order, picking the candidate (drawn
// from recipes, cycling through the shuffled `order`) that maximizes overlap
// with already-picked recipes minus a variety penalty for over-used tags.
func greedyFill(recipes []domain.RecipeSummary, order []int, slotCount int, skipSlots map[int]bool, recipeData map[string]*domain.Recipe, rng *rand.Rand) ([]string, float64) {
	result := make([]string, slotCount)
	pickedIngredients := make(map[string]int) // canonical ingredient -> count picked so far
	pickedTags := make(map[string]int)
	totalScore := 0.0
	cycle := 0

	for slot := 0; slot < slotCount; slot++ {
		if skipSlots[slot] {
			continue
		}

		bestIdx := -1
		bestScore := 0.0
		bestOffset := 0

		// Consider a full pass over the shuffled pool starting from the
		// current cycle position, so ties break according to rng order.
		for offset := 0; offset < len(order); offset++ {
			idx := order[(cycle+offset)%len(order)]
			recipe := recipes[idx]
			score := candidateScore(recipe, pickedIngredients, pickedTags, recipeData)
			if bestIdx == -1 || score > bestScore {
				bestIdx = idx
				bestScore = score
				bestOffset = offset
			}
		}

		chosen := recipes[bestIdx]
		result[slot] = chosen.ID
		totalScore += bestScore
		cycle += bestOffset + 1

		if data, ok := recipeData[chosen.ID]; ok {
			for _, ing := range data.Ingredients {
				canonical := domain.NormalizeIngredientName(ing.Name)
				if ing.IsPantryStaple || domain.IsPantryStapleName(canonical) {
					continue
				}
				pickedIngredients[canonical]++
			}
		}
		for _, tag := range chosen.Tags {
			pickedTags[tag]++
		}
	}

	return result, totalScore
}

// candidateScore rewards recipes that share non-pantry ingredients with
// what's already picked, and penalizes tags that would exceed 2 uses/week.
func candidateScore(recipe domain.RecipeSummary, pickedIngredients map[string]int, pickedTags map[string]int, recipeData map[string]*domain.Recipe) float64 {
	overlapReward := 0.0
	if data, ok := recipeData[recipe.ID]; ok {
		for _, ing := range data.Ingredients {
			canonical := domain.NormalizeIngredientName(ing.Name)
			if ing.IsPantryStaple || domain.IsPantryStapleName(canonical) {
				continue
			}
			if count := pickedIngredients[canonical]; count > 0 {
				overlapReward += float64(count)
			}
		}
	}

	varietyPenalty := 0.0
	for _, tag := range recipe.Tags {
		if pickedTags[tag] >= 2 {
			varietyPenalty += float64(pickedTags[tag]-1)
		}
	}

	return overlapReward - varietyPenalty
}
