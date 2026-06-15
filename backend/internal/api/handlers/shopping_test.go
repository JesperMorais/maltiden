package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- Mock implementations ---

type mockMenuStorage struct {
	householdIDByMenuID map[string]string
	menuByID            map[string]*domain.Menu
}

func (m *mockMenuStorage) Create(menu *domain.Menu) error                        { return nil }
func (m *mockMenuStorage) Update(menu *domain.Menu) error                        { return nil }
func (m *mockMenuStorage) GetCurrentByHousehold(householdID string) (*domain.Menu, error) {
	return nil, nil
}
func (m *mockMenuStorage) GetByID(id string) (*domain.Menu, error) {
	if menu, ok := m.menuByID[id]; ok {
		return menu, nil
	}
	return &domain.Menu{ID: id, Days: []domain.MenuDay{}}, nil
}
func (m *mockMenuStorage) GetHouseholdIDByMenuID(menuID string) (string, error) {
	if hid, ok := m.householdIDByMenuID[menuID]; ok {
		return hid, nil
	}
	return "", nil
}
func (m *mockMenuStorage) GetRecentRecipeIDs(householdID string, windowMenus int) (map[string]bool, error) {
	return map[string]bool{}, nil
}

type mockShoppingRecipeStorage struct{}

func (m *mockShoppingRecipeStorage) GetAll(filter *domain.RecipeFilter, householdID string) ([]domain.RecipeSummary, error) {
	return nil, nil
}
func (m *mockShoppingRecipeStorage) GetAllPaginated(filter *domain.RecipeFilter, householdID string, limit, offset int) ([]domain.RecipeSummary, int, error) {
	return nil, 0, nil
}
func (m *mockShoppingRecipeStorage) GetByID(id string) (*domain.Recipe, error) { return nil, nil }
func (m *mockShoppingRecipeStorage) GetByIDs(ids []string) (map[string]*domain.Recipe, error) {
	return map[string]*domain.Recipe{}, nil
}
func (m *mockShoppingRecipeStorage) Create(recipe *domain.Recipe) error { return nil }
func (m *mockShoppingRecipeStorage) Update(recipe *domain.Recipe) error { return nil }
func (m *mockShoppingRecipeStorage) Delete(id string) error             { return nil }

type mockShoppingStorage struct {
	checkedItems         map[string]bool
	customItems          []domain.CustomShoppingItem
	capInserted          bool
	setCheckedErr        error
	setCustomCheckedErr  error
	deleteErr            error
}

func (m *mockShoppingStorage) GetCheckedItems(menuID string) (map[string]bool, error) {
	if m.checkedItems != nil {
		return m.checkedItems, nil
	}
	return map[string]bool{}, nil
}
func (m *mockShoppingStorage) SetChecked(menuID, itemID string, checked bool) error {
	return m.setCheckedErr
}
func (m *mockShoppingStorage) CreateCustomItem(item *domain.CustomShoppingItem) error { return nil }
func (m *mockShoppingStorage) CreateCustomItemWithCap(item *domain.CustomShoppingItem, maxItems int) (bool, error) {
	return m.capInserted, nil
}
func (m *mockShoppingStorage) DeleteCustomItem(id, householdID string) error {
	return m.deleteErr
}
func (m *mockShoppingStorage) GetCustomItems(menuID, householdID string) ([]domain.CustomShoppingItem, error) {
	return m.customItems, nil
}
func (m *mockShoppingStorage) SetCustomItemChecked(id, householdID string, checked bool) error {
	return m.setCustomCheckedErr
}

// --- Test helpers ---

const (
	testUserID      = "usr_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	testHouseholdID = "hh_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	testMenuID      = "menu_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	otherHousehold  = "hh_bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
)

func newTestShoppingHandler(menuStore *mockMenuStorage, shoppingStore *mockShoppingStorage) *ShoppingHandler {
	svc := services.NewShoppingService(menuStore, &mockShoppingRecipeStorage{}, shoppingStore)
	return NewShoppingHandler(svc, menuStore)
}

func defaultMenuStore() *mockMenuStorage {
	return &mockMenuStorage{
		householdIDByMenuID: map[string]string{
			testMenuID: testHouseholdID,
		},
		menuByID: map[string]*domain.Menu{
			testMenuID: {ID: testMenuID, HouseholdID: testHouseholdID, Days: []domain.MenuDay{}},
		},
	}
}

func idorMenuStore() *mockMenuStorage {
	return &mockMenuStorage{
		householdIDByMenuID: map[string]string{
			testMenuID: otherHousehold,
		},
	}
}

// --- GetShoppingList tests ---

func TestShoppingHandler_GetShoppingList_Happy(t *testing.T) {
	h := newTestShoppingHandler(defaultMenuStore(), &mockShoppingStorage{capInserted: true})

	req := httptest.NewRequest("GET", "/shopping-list?menuId="+testMenuID, nil)
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.GetShoppingList(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestShoppingHandler_GetShoppingList_NoAuth(t *testing.T) {
	h := newTestShoppingHandler(defaultMenuStore(), &mockShoppingStorage{})

	req := httptest.NewRequest("GET", "/shopping-list?menuId="+testMenuID, nil)
	// no setAuthContext → householdID empty
	rr := httptest.NewRecorder()
	h.GetShoppingList(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestShoppingHandler_GetShoppingList_MissingMenuID(t *testing.T) {
	h := newTestShoppingHandler(defaultMenuStore(), &mockShoppingStorage{})

	req := httptest.NewRequest("GET", "/shopping-list", nil)
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.GetShoppingList(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestShoppingHandler_GetShoppingList_IDOR(t *testing.T) {
	h := newTestShoppingHandler(idorMenuStore(), &mockShoppingStorage{})

	req := httptest.NewRequest("GET", "/shopping-list?menuId="+testMenuID, nil)
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.GetShoppingList(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["error"] != "forbidden" {
		t.Errorf("expected error 'forbidden', got %q", resp["error"])
	}
}

// --- UpdateItem tests ---

func TestShoppingHandler_UpdateItem_Happy(t *testing.T) {
	h := newTestShoppingHandler(defaultMenuStore(), &mockShoppingStorage{})

	body := `{"checked":true}`
	req := httptest.NewRequest("PATCH", "/shopping-list/items/item_abc123?menuId="+testMenuID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, testUserID, testHouseholdID)
	req.SetPathValue("id", "item_abc123")
	rr := httptest.NewRecorder()
	h.UpdateItem(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestShoppingHandler_UpdateItem_NoAuth(t *testing.T) {
	h := newTestShoppingHandler(defaultMenuStore(), &mockShoppingStorage{})

	body := `{"checked":true}`
	req := httptest.NewRequest("PATCH", "/shopping-list/items/item_abc123?menuId="+testMenuID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "item_abc123")
	rr := httptest.NewRecorder()
	h.UpdateItem(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestShoppingHandler_UpdateItem_MissingMenuID(t *testing.T) {
	h := newTestShoppingHandler(defaultMenuStore(), &mockShoppingStorage{})

	body := `{"checked":true}`
	req := httptest.NewRequest("PATCH", "/shopping-list/items/item_abc123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, testUserID, testHouseholdID)
	req.SetPathValue("id", "item_abc123")
	rr := httptest.NewRecorder()
	h.UpdateItem(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestShoppingHandler_UpdateItem_IDOR(t *testing.T) {
	h := newTestShoppingHandler(idorMenuStore(), &mockShoppingStorage{})

	body := `{"checked":true}`
	req := httptest.NewRequest("PATCH", "/shopping-list/items/item_abc123?menuId="+testMenuID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, testUserID, testHouseholdID)
	req.SetPathValue("id", "item_abc123")
	rr := httptest.NewRecorder()
	h.UpdateItem(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["error"] != "forbidden" {
		t.Errorf("expected error 'forbidden', got %q", resp["error"])
	}
}

func TestShoppingHandler_UpdateItem_NotFound(t *testing.T) {
	store := &mockShoppingStorage{setCheckedErr: sql.ErrNoRows}
	h := newTestShoppingHandler(defaultMenuStore(), store)

	body := `{"checked":true}`
	req := httptest.NewRequest("PATCH", "/shopping-list/items/item_abc123?menuId="+testMenuID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, testUserID, testHouseholdID)
	req.SetPathValue("id", "item_abc123")
	rr := httptest.NewRecorder()
	h.UpdateItem(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

// --- AddCustomItem tests ---

func TestShoppingHandler_AddCustomItem_Happy(t *testing.T) {
	store := &mockShoppingStorage{capInserted: true}
	h := newTestShoppingHandler(defaultMenuStore(), store)

	body := `{"name":"Mjölk","unit":"liter","amount":2}`
	req := httptest.NewRequest("POST", "/shopping-list/custom?menuId="+testMenuID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.AddCustomItem(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var item domain.CustomShoppingItem
	if err := json.NewDecoder(rr.Body).Decode(&item); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if item.ID == "" {
		t.Fatal("expected non-empty item ID")
	}
}

func TestShoppingHandler_AddCustomItem_NoAuth(t *testing.T) {
	h := newTestShoppingHandler(defaultMenuStore(), &mockShoppingStorage{})

	body := `{"name":"Mjölk"}`
	req := httptest.NewRequest("POST", "/shopping-list/custom?menuId="+testMenuID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.AddCustomItem(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestShoppingHandler_AddCustomItem_MissingMenuID(t *testing.T) {
	h := newTestShoppingHandler(defaultMenuStore(), &mockShoppingStorage{})

	body := `{"name":"Mjölk"}`
	req := httptest.NewRequest("POST", "/shopping-list/custom", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.AddCustomItem(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestShoppingHandler_AddCustomItem_IDOR(t *testing.T) {
	h := newTestShoppingHandler(idorMenuStore(), &mockShoppingStorage{})

	body := `{"name":"Mjölk"}`
	req := httptest.NewRequest("POST", "/shopping-list/custom?menuId="+testMenuID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.AddCustomItem(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["error"] != "forbidden" {
		t.Errorf("expected error 'forbidden', got %q", resp["error"])
	}
}

func TestShoppingHandler_AddCustomItem_CapExceeded(t *testing.T) {
	store := &mockShoppingStorage{capInserted: false}
	h := newTestShoppingHandler(defaultMenuStore(), store)

	body := `{"name":"Mjölk","unit":"liter","amount":1}`
	req := httptest.NewRequest("POST", "/shopping-list/custom?menuId="+testMenuID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.AddCustomItem(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["error"] != "too_many_items" {
		t.Errorf("expected error 'too_many_items', got %q", resp["error"])
	}
}

func TestShoppingHandler_AddCustomItem_NameRequired(t *testing.T) {
	store := &mockShoppingStorage{capInserted: true}
	h := newTestShoppingHandler(defaultMenuStore(), store)

	body := `{"name":"","unit":"liter","amount":1}`
	req := httptest.NewRequest("POST", "/shopping-list/custom?menuId="+testMenuID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.AddCustomItem(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["error"] != "name_required" {
		t.Errorf("expected error 'name_required', got %q", resp["error"])
	}
}

// --- DeleteCustomItem tests ---

func TestShoppingHandler_DeleteCustomItem_Happy(t *testing.T) {
	h := newTestShoppingHandler(defaultMenuStore(), &mockShoppingStorage{})

	itemID := "citem_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	req := httptest.NewRequest("DELETE", "/shopping-list/custom/"+itemID, nil)
	req = setAuthContext(req, testUserID, testHouseholdID)
	req.SetPathValue("id", itemID)
	rr := httptest.NewRecorder()
	h.DeleteCustomItem(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestShoppingHandler_DeleteCustomItem_NoAuth(t *testing.T) {
	h := newTestShoppingHandler(defaultMenuStore(), &mockShoppingStorage{})

	itemID := "citem_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	req := httptest.NewRequest("DELETE", "/shopping-list/custom/"+itemID, nil)
	req.SetPathValue("id", itemID)
	rr := httptest.NewRecorder()
	h.DeleteCustomItem(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestShoppingHandler_DeleteCustomItem_NotFound(t *testing.T) {
	store := &mockShoppingStorage{deleteErr: sql.ErrNoRows}
	h := newTestShoppingHandler(defaultMenuStore(), store)

	itemID := "citem_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	req := httptest.NewRequest("DELETE", "/shopping-list/custom/"+itemID, nil)
	req = setAuthContext(req, testUserID, testHouseholdID)
	req.SetPathValue("id", itemID)
	rr := httptest.NewRecorder()
	h.DeleteCustomItem(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["error"] != "not_found" {
		t.Errorf("expected error 'not_found', got %q", resp["error"])
	}
}

func TestShoppingHandler_DeleteCustomItem_InvalidID(t *testing.T) {
	h := newTestShoppingHandler(defaultMenuStore(), &mockShoppingStorage{})

	req := httptest.NewRequest("DELETE", "/shopping-list/custom/not-valid", nil)
	req = setAuthContext(req, testUserID, testHouseholdID)
	req.SetPathValue("id", "not-valid")
	rr := httptest.NewRecorder()
	h.DeleteCustomItem(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
