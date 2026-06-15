package services

import (
	"fmt"
	"hash/fnv"
	"maltiden/internal/domain"
	"math"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/getsentry/sentry-go"
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

// Swedish ingredient categories.
//
// Categories are organized to mirror a typical Swedish supermarket walk so the
// shopping list matches the order shoppers naturally encounter aisles. When
// adding ingredients, prefer the most common Swedish form (lowercase, with å/ä/ö).
// "basilika"/"oregano"/"timjan"/"rosmarin" without a "färsk" prefix default to
// the dried form (Kryddor) since that is by far the more common pantry staple;
// fresh variants are listed explicitly under "Färska örter".
var ingredientCategories = map[string]string{
	// Kött & Fisk
	"köttfärs":          "Kött & Fisk",
	"nötfärs":           "Kött & Fisk",
	"blandfärs":         "Kött & Fisk",
	"fläskfärs":         "Kött & Fisk",
	"kycklingfärs":      "Kött & Fisk",
	"bacon":             "Kött & Fisk",
	"skinka":            "Kött & Fisk",
	"kassler":           "Kött & Fisk",
	"falukorv":          "Kött & Fisk",
	"prinskorv":         "Kött & Fisk",
	"chorizo":           "Kött & Fisk",
	"salami":            "Kött & Fisk",
	"kycklingfilé":      "Kött & Fisk",
	"kycklinglårfilé":   "Kött & Fisk",
	"kyckling":          "Kött & Fisk",
	"kalkon":            "Kött & Fisk",
	"kalkonfilé":        "Kött & Fisk",
	"oxfilé":            "Kött & Fisk",
	"entrecôte":         "Kött & Fisk",
	"ryggbiff":          "Kött & Fisk",
	"fläskkarré":        "Kött & Fisk",
	"fläskfilé":         "Kött & Fisk",
	"lammfärs":          "Kött & Fisk",
	"lammstek":          "Kött & Fisk",
	"laxfilé":           "Kött & Fisk",
	"lax":               "Kött & Fisk",
	"rökt lax":          "Kött & Fisk",
	"torsk":             "Kött & Fisk",
	"torskfilé":         "Kött & Fisk",
	"sej":               "Kött & Fisk",
	"sejfilé":           "Kött & Fisk",
	"fisk":              "Kött & Fisk",
	"räkor":             "Kött & Fisk",
	"handskalade räkor": "Kött & Fisk",
	"kräftor":           "Kött & Fisk",
	"musslor":           "Kött & Fisk",
	"korv":              "Kött & Fisk",

	// Mejeri & Ägg
	"ägg":            "Mejeri & Ägg",
	"mjölk":          "Mejeri & Ägg",
	"lättmjölk":      "Mejeri & Ägg",
	"mellanmjölk":    "Mejeri & Ägg",
	"standardmjölk":  "Mejeri & Ägg",
	"laktosfri mjölk": "Mejeri & Ägg",
	"havremjölk":     "Mejeri & Ägg",
	"sojamjölk":      "Mejeri & Ägg",
	"mandelmjölk":    "Mejeri & Ägg",
	"grädde":         "Mejeri & Ägg",
	"vispgrädde":     "Mejeri & Ägg",
	"matlagningsgrädde": "Mejeri & Ägg",
	"crème fraîche":  "Mejeri & Ägg",
	"gräddfil":       "Mejeri & Ägg",
	"smör":           "Mejeri & Ägg",
	"bregott":        "Mejeri & Ägg",
	"margarin":       "Mejeri & Ägg",
	"ost":            "Mejeri & Ägg",
	"riven ost":      "Mejeri & Ägg",
	"hushållsost":    "Mejeri & Ägg",
	"prästost":       "Mejeri & Ägg",
	"västerbottensost": "Mejeri & Ägg",
	"parmesan":       "Mejeri & Ägg",
	"mozzarella":     "Mejeri & Ägg",
	"fetaost":        "Mejeri & Ägg",
	"halloumi":       "Mejeri & Ägg",
	"chèvre":         "Mejeri & Ägg",
	"cottage cheese": "Mejeri & Ägg",
	"keso":           "Mejeri & Ägg",
	"kvarg":          "Mejeri & Ägg",
	"yoghurt":        "Mejeri & Ägg",
	"grekisk yoghurt": "Mejeri & Ägg",
	"turkisk yoghurt": "Mejeri & Ägg",
	"filmjölk":       "Mejeri & Ägg",
	"kefir":          "Mejeri & Ägg",

	// Frukt
	"äpple":       "Frukt",
	"päron":       "Frukt",
	"banan":       "Frukt",
	"apelsin":     "Frukt",
	"clementin":   "Frukt",
	"mandarin":    "Frukt",
	"citron":      "Frukt",
	"lime":        "Frukt",
	"grapefrukt":  "Frukt",
	"jordgubbar":  "Frukt",
	"blåbär":      "Frukt",
	"hallon":      "Frukt",
	"björnbär":    "Frukt",
	"vindruvor":   "Frukt",
	"druvor":      "Frukt",
	"ananas":      "Frukt",
	"mango":       "Frukt",
	"papaya":      "Frukt",
	"melon":       "Frukt",
	"vattenmelon": "Frukt",
	"kiwi":        "Frukt",
	"persika":     "Frukt",
	"nektarin":    "Frukt",
	"plommon":     "Frukt",
	"granatäpple": "Frukt",
	"avokado":     "Frukt",

	// Grönsaker
	"lök":              "Grönsaker",
	"gul lök":          "Grönsaker",
	"röd lök":          "Grönsaker",
	"vitlök":           "Grönsaker",
	"schalottenlök":    "Grönsaker",
	"purjolök":         "Grönsaker",
	"salladslök":       "Grönsaker",
	"morot":            "Grönsaker",
	"morötter":         "Grönsaker",
	"potatis":          "Grönsaker",
	"färskpotatis":     "Grönsaker",
	"sötpotatis":       "Grönsaker",
	"paprika":          "Grönsaker",
	"röd paprika":      "Grönsaker",
	"gul paprika":      "Grönsaker",
	"grön paprika":     "Grönsaker",
	"tomat":            "Grönsaker",
	"körsbärstomater":  "Grönsaker",
	"körsbärstomat":    "Grönsaker",
	"kvisttomater":     "Grönsaker",
	"gurka":            "Grönsaker",
	"sallad":           "Grönsaker",
	"isbergssallad":    "Grönsaker",
	"romansallad":      "Grönsaker",
	"ruccola":          "Grönsaker",
	"spenat":           "Grönsaker",
	"babyspenat":       "Grönsaker",
	"grönkål":          "Grönsaker",
	"broccoli":         "Grönsaker",
	"blomkål":          "Grönsaker",
	"zucchini":         "Grönsaker",
	"squash":           "Grönsaker",
	"aubergine":        "Grönsaker",
	"champinjoner":     "Grönsaker",
	"svamp":            "Grönsaker",
	"kantareller":      "Grönsaker",
	"karljohan":        "Grönsaker",
	"kålrot":           "Grönsaker",
	"palsternacka":     "Grönsaker",
	"rödbetor":         "Grönsaker",
	"rödbeta":          "Grönsaker",
	"rödkål":           "Grönsaker",
	"vitkål":           "Grönsaker",
	"spetskål":         "Grönsaker",
	"savojkål":         "Grönsaker",
	"brysselkål":       "Grönsaker",
	"sockerärtor":      "Grönsaker",
	"haricots verts":   "Grönsaker",
	"majskolv":         "Grönsaker",
	"selleri":          "Grönsaker",
	"rotselleri":       "Grönsaker",
	"fänkål":           "Grönsaker",
	"rädisor":          "Grönsaker",
	"sparris":          "Grönsaker",
	"chili":            "Grönsaker",
	"ingefära":         "Grönsaker",

	// Färska örter
	"persilja":        "Färska örter",
	"bladpersilja":    "Färska örter",
	"dill":            "Färska örter",
	"färsk basilika":  "Färska örter",
	"färsk koriander": "Färska örter",
	"koriander":       "Färska örter",
	"mynta":           "Färska örter",
	"gräslök":         "Färska örter",
	"färsk rosmarin":  "Färska örter",
	"färsk timjan":    "Färska örter",
	"färsk salvia":    "Färska örter",
	"färsk oregano":   "Färska örter",
	"körvel":          "Färska örter",
	"dragon":          "Färska örter",

	// Bröd
	"bröd":          "Bröd",
	"formbröd":      "Bröd",
	"limpa":         "Bröd",
	"baguette":      "Bröd",
	"frallor":       "Bröd",
	"frukostfralla": "Bröd",
	"knäckebröd":    "Bröd",
	"tunnbröd":      "Bröd",
	"tortilla":      "Bröd",
	"tortillabröd":  "Bröd",
	"pitabröd":      "Bröd",
	"hamburgerbröd": "Bröd",
	"korvbröd":      "Bröd",
	"naanbröd":      "Bröd",
	"surdegsbröd":   "Bröd",

	// Pasta, ris & spannmål
	"spaghetti":       "Pasta, ris & spannmål",
	"pasta":           "Pasta, ris & spannmål",
	"makaroner":       "Pasta, ris & spannmål",
	"penne":           "Pasta, ris & spannmål",
	"fusilli":         "Pasta, ris & spannmål",
	"tagliatelle":     "Pasta, ris & spannmål",
	"farfalle":        "Pasta, ris & spannmål",
	"rigatoni":        "Pasta, ris & spannmål",
	"lasagneplattor":  "Pasta, ris & spannmål",
	"cannelloni":      "Pasta, ris & spannmål",
	"gnocchi":         "Pasta, ris & spannmål",
	"nudlar":          "Pasta, ris & spannmål",
	"risnudlar":       "Pasta, ris & spannmål",
	"äggnudlar":       "Pasta, ris & spannmål",
	"ris":             "Pasta, ris & spannmål",
	"basmatiris":      "Pasta, ris & spannmål",
	"jasminris":       "Pasta, ris & spannmål",
	"råris":           "Pasta, ris & spannmål",
	"arborio":         "Pasta, ris & spannmål",
	"risottoris":      "Pasta, ris & spannmål",
	"couscous":        "Pasta, ris & spannmål",
	"bulgur":          "Pasta, ris & spannmål",
	"quinoa":          "Pasta, ris & spannmål",
	"havregryn":       "Pasta, ris & spannmål",
	"müsli":           "Pasta, ris & spannmål",
	"cornflakes":      "Pasta, ris & spannmål",
	"linser":          "Pasta, ris & spannmål",
	"röda linser":     "Pasta, ris & spannmål",
	"gröna linser":    "Pasta, ris & spannmål",
	"vetemjöl":        "Pasta, ris & spannmål",
	"rågmjöl":         "Pasta, ris & spannmål",
	"grahamsmjöl":     "Pasta, ris & spannmål",
	"mandelmjöl":      "Pasta, ris & spannmål",
	"strösocker":      "Pasta, ris & spannmål",
	"socker":          "Pasta, ris & spannmål",
	"florsocker":      "Pasta, ris & spannmål",
	"farinsocker":     "Pasta, ris & spannmål",
	"jäst":            "Pasta, ris & spannmål",
	"bakpulver":       "Pasta, ris & spannmål",
	"bikarbonat":      "Pasta, ris & spannmål",
	"ströbröd":        "Pasta, ris & spannmål",
	"panko":           "Pasta, ris & spannmål",

	// Konserver
	"krossade tomater":   "Konserver",
	"hela tomater":       "Konserver",
	"passerade tomater":  "Konserver",
	"tomatpuré":          "Konserver",
	"kokosmjölk":         "Konserver",
	"kokosgrädde":        "Konserver",
	"vita bönor":         "Konserver",
	"svarta bönor":       "Konserver",
	"röda bönor":         "Konserver",
	"kidneybönor":        "Konserver",
	"bönor":              "Konserver",
	"kikärtor":           "Konserver",
	"majs":               "Konserver",
	"tonfisk":            "Konserver",
	"sardiner":           "Konserver",
	"makrill":            "Konserver",
	"ansjovis":           "Konserver",
	"oliver":             "Konserver",
	"kapris":             "Konserver",
	"inlagd gurka":       "Konserver",
	"rödbetor inlagda":   "Konserver",
	"sylt":               "Konserver",
	"lingonsylt":         "Konserver",
	"jordgubbssylt":      "Konserver",
	"honung":             "Konserver",
	"jordnötssmör":       "Konserver",

	// Såser & olja
	"olivolja":         "Såser & olja",
	"rapsolja":         "Såser & olja",
	"solrosolja":       "Såser & olja",
	"sesamolja":        "Såser & olja",
	"kokosolja":        "Såser & olja",
	"smaksatt olja":    "Såser & olja",
	"soja":             "Såser & olja",
	"sojasås":          "Såser & olja",
	"japansk soja":     "Såser & olja",
	"balsamvinäger":    "Såser & olja",
	"vinäger":          "Såser & olja",
	"ättika":           "Såser & olja",
	"äppelcidervinäger": "Såser & olja",
	"ketchup":          "Såser & olja",
	"senap":            "Såser & olja",
	"dijonsenap":       "Såser & olja",
	"sötstark senap":   "Såser & olja",
	"majonnäs":         "Såser & olja",
	"aioli":            "Såser & olja",
	"sweet chili":      "Såser & olja",
	"sweet chilisås":   "Såser & olja",
	"hoisinsås":        "Såser & olja",
	"ostronsås":        "Såser & olja",
	"fisksås":          "Såser & olja",
	"sambal oelek":     "Såser & olja",
	"sriracha":         "Såser & olja",
	"worcestershiresås": "Såser & olja",
	"buljong":          "Såser & olja",
	"grönsaksbuljong":  "Såser & olja",
	"kycklingbuljong":  "Såser & olja",
	"köttbuljong":      "Såser & olja",
	"fond":             "Såser & olja",
	"pesto":            "Såser & olja",
	"röd pesto":        "Såser & olja",
	"tahini":           "Såser & olja",

	// Kryddor (dried herbs and seasonings; fresh herbs go to Färska örter)
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
	"salvia":        "Kryddor",
	"paprikapulver": "Kryddor",
	"rökt paprika":  "Kryddor",
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
	"garam masala":  "Kryddor",
	"tacokrydda":    "Kryddor",
	"köttbullskrydda": "Kryddor",
	"fänkålsfrön":   "Kryddor",
	"korianderfrön": "Kryddor",
	"senapsfrön":    "Kryddor",
	"vaniljsocker":  "Kryddor",
	"vaniljstång":   "Kryddor",
	"saffran":       "Kryddor",
	"ingefära mald": "Kryddor",
	"chiliflingor":  "Kryddor",

	// Frys
	"ärtor":                   "Frys",
	"frysta ärtor":            "Frys",
	"blandade grönsaker":      "Frys",
	"frysta grönsaker":        "Frys",
	"frysta wokgrönsaker":     "Frys",
	"wokgrönsaker":            "Frys",
	"frysta bär":              "Frys",
	"frysta hallon":           "Frys",
	"frysta blåbär":           "Frys",
	"frysta jordgubbar":       "Frys",
	"glass":                   "Frys",
	"vaniljglass":             "Frys",
	"fiskpinnar":              "Frys",
	"kycklingnuggets":         "Frys",
	"frysta köttbullar":       "Frys",
	"köttbullar":              "Frys",
	"piroger":                 "Frys",
	"pizzadeg":                "Frys",
	"smördeg":                 "Frys",
	"filodeg":                 "Frys",
	"tacoskal":                "Bröd",
}

func (s *ShoppingService) GetShoppingList(menuID, householdID string) (*domain.ShoppingList, error) {
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

	// Get menu
	menu, err := s.menuStorage.GetByID(menuID)
	if err != nil {
		sentry.CaptureException(err)
		return nil, err
	}
	if menu == nil {
		return nil, nil
	}

	// Get checked items
	checkedItems, err := s.shoppingStorage.GetCheckedItems(menuID)
	if err != nil {
		sentry.CaptureException(err)
		return nil, err
	}

	// Collect all unique recipe IDs from non-skip days
	recipeIDs := make([]string, 0, len(menu.Days))
	recipeIDSet := make(map[string]bool)
	for _, day := range menu.Days {
		// Leftovers days reuse the batch cooked on their cook day (whose 2×
		// servings already covers both), so they buy nothing and are skipped here.
		if !day.Skip && !day.Leftover && day.RecipeID != "" {
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
		// Leftovers days add no groceries: the cook day's 2× servings already
		// covers them (see the distinct-id collection above).
		if day.Leftover {
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

	// Build response with sorted categories.
	// Order mirrors a typical Swedish supermarket walk: meat/fish counter first,
	// then dairy, then produce (split into fruit, veg, herbs), bakery, dry goods,
	// canned/sauces, spices, frozen, and finally the catch-all "Övrigt".
	categoryOrder := []string{
		"Kött & Fisk",
		"Mejeri & Ägg",
		"Frukt",
		"Grönsaker",
		"Färska örter",
		"Bröd",
		"Pasta, ris & spannmål",
		"Konserver",
		"Såser & olja",
		"Kryddor",
		"Frys",
		"Övrigt",
	}
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
		sentry.CaptureException(err)
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
