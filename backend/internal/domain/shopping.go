package domain

type ShoppingItem struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	Unit     string  `json:"unit"`
	Checked  bool    `json:"checked"`
	IsCustom bool    `json:"isCustom"`
}

type ShoppingCategory struct {
	Name  string         `json:"name"`
	Items []ShoppingItem `json:"items"`
}

type ShoppingList struct {
	MenuID     string             `json:"menuId"`
	Categories []ShoppingCategory `json:"categories"`
}

type UpdateShoppingItemRequest struct {
	Checked bool `json:"checked"`
}

type CustomShoppingItem struct {
	ID          string  `json:"id"`
	MenuID      string  `json:"menuId"`
	HouseholdID string  `json:"householdId"`
	Name        string  `json:"name"`
	Unit        string  `json:"unit"`
	Amount      float64 `json:"amount"`
	Checked     bool    `json:"checked"`
}

type CreateCustomItemRequest struct {
	Name   string  `json:"name"`
	Unit   string  `json:"unit"`
	Amount float64 `json:"amount"`
}
