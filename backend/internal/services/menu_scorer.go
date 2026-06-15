package services

import (
	"maltiden/internal/domain"
	"sort"
	"strings"
)

// economyMaxShared caps how many shared ingredients the economy summary lists,
// keeping the UI focused on the most-reused groceries.
const economyMaxShared = 8

// computeMenuEconomy summarizes ingredient reuse across the chosen week.
//
//   - DistinctItemsToBuy = number of unique non-staple canonical ingredients.
//   - TotalIngredientRefs = sum of per-recipe non-staple ingredient counts.
//   - SharedIngredients = non-staple canonicals appearing in ≥2 distinct
//     recipes, sorted by recipe count (desc, then name), capped to the top few.
//
// Pantry staples are excluded everywhere. recipesByID supplies full recipes
// (with ingredients) for each chosen id; ids missing from the map are skipped.
func computeMenuEconomy(chosenIDs []string, recipesByID map[string]domain.Recipe) *domain.MenuEconomy {
	type agg struct {
		recipes  map[string]struct{} // distinct recipe ids carrying this canonical
		firstRaw string               // first raw display name seen
	}

	byCanon := make(map[string]*agg)
	totalRefs := 0

	for _, id := range chosenIDs {
		r, ok := recipesByID[id]
		if !ok {
			continue
		}
		hasMeta := recipeHasMetadata(r)
		// Per-recipe set so the same canonical inside one recipe counts once
		// toward both refs and the shared-recipe tally.
		seen := make(map[string]struct{})
		for _, ing := range r.Ingredients {
			if effectiveStaple(ing, hasMeta) {
				continue
			}
			c := effectiveCanonical(ing)
			if c == "" {
				continue
			}
			if _, dup := seen[c]; dup {
				continue
			}
			seen[c] = struct{}{}
			totalRefs++

			a := byCanon[c]
			if a == nil {
				a = &agg{recipes: make(map[string]struct{})}
				if name := firstRawName(ing); name != "" {
					a.firstRaw = name
				}
				byCanon[c] = a
			}
			if a.firstRaw == "" {
				a.firstRaw = firstRawName(ing)
			}
			a.recipes[id] = struct{}{}
		}
	}

	shared := make([]domain.SharedIngredient, 0)
	for canon, a := range byCanon {
		if len(a.recipes) >= 2 {
			name := a.firstRaw
			if name == "" {
				name = canon
			}
			shared = append(shared, domain.SharedIngredient{
				CanonicalName: canon,
				Name:          name,
				RecipeCount:   len(a.recipes),
			})
		}
	}

	sort.Slice(shared, func(i, j int) bool {
		if shared[i].RecipeCount != shared[j].RecipeCount {
			return shared[i].RecipeCount > shared[j].RecipeCount
		}
		return shared[i].CanonicalName < shared[j].CanonicalName
	})
	if len(shared) > economyMaxShared {
		shared = shared[:economyMaxShared]
	}

	return &domain.MenuEconomy{
		SharedIngredients:   shared,
		DistinctItemsToBuy:  len(byCanon),
		TotalIngredientRefs: totalRefs,
	}
}

// firstRawName returns the ingredient's raw display name (its Name field
// trimmed), falling back to the canonical when the raw name is blank.
func firstRawName(ing domain.Ingredient) string {
	if n := strings.TrimSpace(ing.Name); n != "" {
		return n
	}
	return effectiveCanonical(ing)
}
