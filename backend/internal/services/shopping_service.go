package services

import (
	"fmt"
	"hash/fnv"
	"maltiden/internal/domain"
	"math"
	"sort"
	"strings"
)

type ShoppingService struct {
	menuStorage     domain.MenuRepository
	recipeStorage   domain.RecipeRepository
	shoppingStorage domain.ShoppingRepository
}

func NewShoppingService(
	menuStorage domain.MenuRepository,
	recipeStorage domain.RecipeRepository,
	shoppingStorage domain.ShoppingRepository,
) *ShoppingService {
	return &ShoppingService{
		menuStorage:     menuStorage,
		recipeStorage:   recipeStorage,
		shoppingStorage: shoppingStorage,
	}
}

// Swedish ingredient categories
var ingredientCategories = map[string]string{
	// Kött & Fisk
	"köttfärs":     "Kött & Fisk",
	"bacon":        "Kött & Fisk",
	"kycklingfilé": "Kött & Fisk",
	"kyckling":     "Kött & Fisk",
	"laxfilé":      "Kött & Fisk",
	"lax":          "Kött & Fisk",
	"fisk":         "Kött & Fisk",
	"fläskfilé":    "Kött & Fisk",
	"korv":         "Kött & Fisk",

	// Mejeri
	"ägg":       "Mejeri",
	"parmesan":  "Mejeri",
	"ost":       "Mejeri",
	"riven ost": "Mejeri",
	"grädde":    "Mejeri",
	"mjölk":     "Mejeri",
	"smör":      "Mejeri",

	// Frukt & Grönt
	"tomat":          "Frukt & Grönt",
	"sallad":         "Frukt & Grönt",
	"lök":            "Frukt & Grönt",
	"vitlök":         "Frukt & Grönt",
	"citron":         "Frukt & Grönt",
	"wokgrönsaker":   "Frukt & Grönt",
	"potatis":        "Frukt & Grönt",
	"dill":           "Frukt & Grönt",

	// Skafferi
	"spaghetti":        "Skafferi",
	"pasta":            "Skafferi",
	"ris":              "Skafferi",
	"krossade tomater": "Skafferi",
	"sojasås":          "Skafferi",
	"tacokrydda":       "Skafferi",
	"tacoskal":         "Skafferi",

	// Kryddor (dried herbs and seasonings; fresh herbs like dill/persilja stay in Frukt & Grönt)
	"salt":          "Kryddor",
	"havssalt":      "Kryddor",
	"flingsalt":     "Kryddor",
	"koksalt":       "Kryddor",
	"peppar":        "Kryddor",
	"svartpeppar":   "Kryddor",
	"vitpeppar":     "Kryddor",
	"oregano":       "Kryddor",
	"basilika":      "Kryddor",
	"timjan":        "Kryddor",
	"rosmarin":      "Kryddor",
	"paprikapulver": "Kryddor",
	"chiliflakes":   "Kryddor",
	"chilipulver":   "Kryddor",
	"kanel":         "Kryddor",
	"muskot":        "Kryddor",
	"kardemumma":    "Kryddor",
	"lagerblad":     "Kryddor",
	"kryddpeppar":   "Kryddor",
	"spiskummin":    "Kryddor",
	"gurkmeja":      "Kryddor",
	"curry":         "Kryddor",
	"currypulver":   "Kryddor",
	"kajennpeppar":  "Kryddor",
}

func (s *ShoppingService) GetShoppingList(menuID string) (*domain.ShoppingList, error) {
	// Get menu
	menu, err := s.menuStorage.GetByID(menuID)
	if err != nil {
		return nil, err
	}
	if menu == nil {
		return nil, nil
	}

	// Get checked items
	checkedItems, err := s.shoppingStorage.GetCheckedItems(menuID)
	if err != nil {
		return nil, err
	}

	// Collect all unique recipe IDs from non-skip days
	recipeIDs := make([]string, 0, len(menu.Days))
	recipeIDSet := make(map[string]bool)
	for _, day := range menu.Days {
		if !day.Skip && day.RecipeID != "" {
			if !recipeIDSet[day.RecipeID] {
				recipeIDs = append(recipeIDs, day.RecipeID)
				recipeIDSet[day.RecipeID] = true
			}
		}
	}

	// Batch fetch all recipes (eliminates N+1 query problem)
	recipes, err := s.recipeStorage.GetByIDs(recipeIDs)
	if err != nil {
		return nil, err
	}

	// Aggregate ingredients from all recipes
	aggregated := make(map[string]*domain.ShoppingItem)

	for _, day := range menu.Days {
		if day.Skip || day.RecipeID == "" {
			continue
		}

		recipe, ok := recipes[day.RecipeID]
		if !ok {
			continue
		}

		// Scale ingredients based on servings
		scale := float64(day.Servings) / float64(recipe.Servings)

		for _, ing := range recipe.Ingredients {
			key := strings.ToLower(ing.Name) + "_" + ing.Unit
			scaledAmount := ing.Amount * scale

			if existing, ok := aggregated[key]; ok {
				existing.Amount += scaledAmount
			} else {
				itemID := generateItemID(menuID, ing.Name, ing.Unit)
				aggregated[key] = &domain.ShoppingItem{
					ID:      itemID,
					Name:    ing.Name,
					Amount:  scaledAmount,
					Unit:    ing.Unit,
					Checked: checkedItems[itemID],
				}
			}
		}
	}

	// Group by category; for spices we drop amounts (a shopping list just needs
	// "buy salt", not "3.66 st"), and dedupe by name since the same spice may
	// have been entered with different units (tsk + krm + st) across recipes.
	categoryMap := make(map[string][]domain.ShoppingItem)
	seenSpice := make(map[string]bool)
	for _, item := range aggregated {
		category := categorizeIngredient(item.Name)
		if category == "Kryddor" {
			nameKey := strings.ToLower(item.Name)
			if seenSpice[nameKey] {
				continue
			}
			seenSpice[nameKey] = true
			it := *item
			it.Amount = 0
			it.Unit = ""
			it.ID = generateItemID(menuID, it.Name, "")
			it.Checked = checkedItems[it.ID]
			categoryMap[category] = append(categoryMap[category], it)
			continue
		}
		it := *item
		it.Amount = roundAmount(it.Amount, it.Unit)
		categoryMap[category] = append(categoryMap[category], it)
	}

	// Build response with sorted categories
	categoryOrder := []string{"Kött & Fisk", "Mejeri", "Frukt & Grönt", "Skafferi", "Kryddor", "Övrigt"}
	var categories []domain.ShoppingCategory

	for _, catName := range categoryOrder {
		if items, ok := categoryMap[catName]; ok {
			// Sort items by name
			sort.Slice(items, func(i, j int) bool {
				return items[i].Name < items[j].Name
			})
			categories = append(categories, domain.ShoppingCategory{
				Name:  catName,
				Items: items,
			})
		}
	}

	return &domain.ShoppingList{
		MenuID:     menuID,
		Categories: categories,
	}, nil
}

func (s *ShoppingService) UpdateItemChecked(menuID, itemID string, checked bool) error {
	return s.shoppingStorage.SetChecked(menuID, itemID, checked)
}

func categorizeIngredient(name string) string {
	nameLower := strings.ToLower(name)
	if cat, ok := ingredientCategories[nameLower]; ok {
		return cat
	}
	return "Övrigt"
}

// roundAmount keeps shopping-list amounts human-readable: integers for
// countable units like "st" (you don't buy 3.6 eggs) and 2-decimal precision
// for weight/volume. Raw scaling from float math would otherwise print things
// like "3.6666666666666665".
func roundAmount(amount float64, unit string) float64 {
	u := strings.ToLower(strings.TrimSpace(unit))
	if u == "st" || u == "styck" || u == "stycken" {
		return math.Round(amount)
	}
	return math.Round(amount*100) / 100
}

func generateItemID(menuID, name, unit string) string {
	h := fnv.New64a()
	h.Write([]byte(menuID + "_" + strings.ToLower(name) + "_" + unit))
	return fmt.Sprintf("item_%x", h.Sum64())
}
