package services

import (
	"fmt"
	"hash/fnv"
	"maltiden/internal/domain"
	"math"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
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

func (s *ShoppingService) GetShoppingList(menuID, householdID string) (*domain.ShoppingList, error) {
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

	// Append custom items as "Egna varor" category
	customItems, err := s.shoppingStorage.GetCustomItems(menuID, householdID)
	if err != nil {
		return nil, err
	}
	if len(customItems) > 0 {
		var items []domain.ShoppingItem
		for _, ci := range customItems {
			items = append(items, domain.ShoppingItem{
				ID:       ci.ID,
				Name:     ci.Name,
				Amount:   ci.Amount,
				Unit:     ci.Unit,
				Checked:  ci.Checked,
				IsCustom: true,
			})
		}
		categories = append(categories, domain.ShoppingCategory{
			Name:  "Egna varor",
			Items: items,
		})
	}

	return &domain.ShoppingList{
		MenuID:     menuID,
		Categories: categories,
	}, nil
}

func (s *ShoppingService) CreateCustomItem(menuID, householdID string, req domain.CreateCustomItemRequest) (*domain.CustomShoppingItem, error) {
	// IDOR protection: verify menu belongs to caller's household.
	menuHouseholdID, err := s.menuStorage.GetHouseholdIDByMenuID(menuID)
	if err != nil {
		return nil, err
	}
	if menuHouseholdID == "" {
		return nil, domain.ErrMenuNotFound
	}
	if menuHouseholdID != householdID {
		return nil, domain.ErrForbidden
	}

	// Trim leading/trailing whitespace before validating presence.
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, domain.ErrNameRequired
	}
	// Use rune count so Swedish characters (å, ä, ö) count as single chars.
	if utf8.RuneCountInString(name) > 200 {
		return nil, domain.ErrNameTooLong
	}
	if utf8.RuneCountInString(req.Unit) > 20 {
		return nil, domain.ErrUnitTooLong
	}
	if math.IsNaN(req.Amount) || math.IsInf(req.Amount, 0) {
		return nil, domain.ErrInvalidAmount
	}
	// Negative amounts are rejected; zero is treated as "unspecified" and
	// defaults to 1 below for ergonomic input.
	if req.Amount < 0 {
		return nil, domain.ErrInvalidAmount
	}
	if req.Amount > 100000 {
		return nil, domain.ErrAmountTooLarge
	}

	unit := req.Unit
	if unit == "" {
		unit = "st"
	}
	amount := req.Amount
	if amount == 0 {
		amount = 1
	}

	item := &domain.CustomShoppingItem{
		ID:          "citem_" + uuid.New().String(),
		MenuID:      menuID,
		HouseholdID: householdID,
		Name:        name,
		Unit:        unit,
		Amount:      amount,
	}

	// Atomic count-and-insert: enforces the per-menu cap in a single SQL
	// statement so concurrent requests cannot both pass the count check and
	// exceed the limit.
	inserted, err := s.shoppingStorage.CreateCustomItemWithCap(item, 500)
	if err != nil {
		return nil, err
	}
	if !inserted {
		return nil, domain.ErrTooManyItems
	}

	return item, nil
}

func (s *ShoppingService) DeleteCustomItem(itemID, householdID string) error {
	return s.shoppingStorage.DeleteCustomItem(itemID, householdID)
}

func (s *ShoppingService) UpdateItemChecked(menuID, itemID, householdID string, checked bool) error {
	// IDOR protection: verify menu belongs to caller's household.
	menuHouseholdID, err := s.menuStorage.GetHouseholdIDByMenuID(menuID)
	if err != nil {
		return err
	}
	if menuHouseholdID == "" {
		return domain.ErrMenuNotFound
	}
	if menuHouseholdID != householdID {
		return domain.ErrForbidden
	}

	// Custom items are stored in a separate table and scoped by household for IDOR protection
	if strings.HasPrefix(itemID, "citem_") {
		return s.shoppingStorage.SetCustomItemChecked(itemID, householdID, checked)
	}
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
