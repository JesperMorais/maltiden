package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockMenuStorageForHandler wraps mockMenuStorage with controllable errors.
type mockMenuStorageForHandler struct {
	mockMenuStorage
	createErr  error
	updateErr  error
	currentMenu *domain.Menu
	getCurrentErr error
}

func (m *mockMenuStorageForHandler) Create(menu *domain.Menu) error {
	return m.createErr
}

func (m *mockMenuStorageForHandler) Update(menu *domain.Menu) error {
	return m.updateErr
}

func (m *mockMenuStorageForHandler) GetCurrentByHousehold(householdID string) (*domain.Menu, error) {
	return m.currentMenu, m.getCurrentErr
}

func (m *mockMenuStorageForHandler) GetByID(id string) (*domain.Menu, error) {
	return m.mockMenuStorage.GetByID(id)
}

func (m *mockMenuStorageForHandler) GetHouseholdIDByMenuID(menuID string) (string, error) {
	return m.mockMenuStorage.GetHouseholdIDByMenuID(menuID)
}

type mockRecipeStorageForMenus struct {
	recipes []domain.RecipeSummary
}

func (m *mockRecipeStorageForMenus) GetAll(filter *domain.RecipeFilter, householdID string) ([]domain.RecipeSummary, error) {
	return m.recipes, nil
}

func (m *mockRecipeStorageForMenus) GetAllPaginated(filter *domain.RecipeFilter, householdID string, limit, offset int) ([]domain.RecipeSummary, int, error) {
	return m.recipes, len(m.recipes), nil
}

func (m *mockRecipeStorageForMenus) GetByID(id string) (*domain.Recipe, error) {
	return nil, nil
}

func (m *mockRecipeStorageForMenus) GetByIDs(ids []string) (map[string]*domain.Recipe, error) {
	return map[string]*domain.Recipe{}, nil
}

func (m *mockRecipeStorageForMenus) Create(recipe *domain.Recipe) error { return nil }
func (m *mockRecipeStorageForMenus) Update(recipe *domain.Recipe) error { return nil }
func (m *mockRecipeStorageForMenus) Delete(id string) error             { return nil }

// mockMenuPrefsStorageForMenus is a no-op MenuPreferencesRepository: Get
// returns nil so the service falls back to DefaultMenuPreferences.
type mockMenuPrefsStorageForMenus struct{}

func (m *mockMenuPrefsStorageForMenus) Get(householdID string) (*domain.MenuPreferences, error) {
	return nil, nil
}

func (m *mockMenuPrefsStorageForMenus) Upsert(prefs *domain.MenuPreferences) error { return nil }

func newTestMenuHandler(menuStore domain.MenuRepository, recipeStore domain.RecipeRepository) *MenuHandler {
	svc := services.NewMenuService(menuStore, recipeStore, &mockMenuPrefsStorageForMenus{})
	return NewMenuHandler(svc)
}

func defaultMenuStorageForHandler() *mockMenuStorageForHandler {
	return &mockMenuStorageForHandler{
		mockMenuStorage: mockMenuStorage{
			householdIDByMenuID: map[string]string{},
			menuByID:            map[string]*domain.Menu{},
		},
		currentMenu: &domain.Menu{
			ID:          testMenuID,
			HouseholdID: testHouseholdID,
			Days:        []domain.MenuDay{},
		},
	}
}

func recipeStorageWithRecipes() *mockRecipeStorageForMenus {
	return &mockRecipeStorageForMenus{
		recipes: []domain.RecipeSummary{
			{ID: "rec_1", Name: "Pasta", Servings: 4},
			{ID: "rec_2", Name: "Pizza", Servings: 4},
		},
	}
}

// --- Generate tests ---

func TestMenuHandler_Generate_Happy(t *testing.T) {
	h := newTestMenuHandler(defaultMenuStorageForHandler(), recipeStorageWithRecipes())

	body, _ := json.Marshal(domain.GenerateMenuRequest{Days: 3, Servings: 2})
	req := httptest.NewRequest("POST", "/menus/generate", bytes.NewReader(body))
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.Generate(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp domain.MenuResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Days) != 3 {
		t.Errorf("expected 3 days, got %d", len(resp.Days))
	}
}

func TestMenuHandler_Generate_NoAuth(t *testing.T) {
	h := newTestMenuHandler(defaultMenuStorageForHandler(), recipeStorageWithRecipes())

	body, _ := json.Marshal(domain.GenerateMenuRequest{Days: 3})
	req := httptest.NewRequest("POST", "/menus/generate", bytes.NewReader(body))
	// no setAuthContext
	rr := httptest.NewRecorder()
	h.Generate(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestMenuHandler_Generate_InvalidDays(t *testing.T) {
	h := newTestMenuHandler(defaultMenuStorageForHandler(), recipeStorageWithRecipes())

	body, _ := json.Marshal(domain.GenerateMenuRequest{Days: 99})
	req := httptest.NewRequest("POST", "/menus/generate", bytes.NewReader(body))
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.Generate(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "invalid_days" {
		t.Errorf("expected invalid_days, got %q", errResp["error"])
	}
}

func TestMenuHandler_Generate_InvalidServings(t *testing.T) {
	h := newTestMenuHandler(defaultMenuStorageForHandler(), recipeStorageWithRecipes())

	body, _ := json.Marshal(domain.GenerateMenuRequest{Days: 3, Servings: 999})
	req := httptest.NewRequest("POST", "/menus/generate", bytes.NewReader(body))
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.Generate(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "invalid_servings" {
		t.Errorf("expected invalid_servings, got %q", errResp["error"])
	}
}

func TestMenuHandler_Generate_NoRecipesAvailable(t *testing.T) {
	emptyRecipes := &mockRecipeStorageForMenus{recipes: []domain.RecipeSummary{}}
	h := newTestMenuHandler(defaultMenuStorageForHandler(), emptyRecipes)

	body, _ := json.Marshal(domain.GenerateMenuRequest{Days: 3})
	req := httptest.NewRequest("POST", "/menus/generate", bytes.NewReader(body))
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.Generate(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "no_recipes_available" {
		t.Errorf("expected no_recipes_available, got %q", errResp["error"])
	}
}

// --- UpdateCurrent tests ---

func TestMenuHandler_UpdateCurrent_Happy(t *testing.T) {
	h := newTestMenuHandler(defaultMenuStorageForHandler(), recipeStorageWithRecipes())

	days := []domain.MenuDay{{Date: "2026-06-14", RecipeID: "rec_1", Servings: 4}}
	body, _ := json.Marshal(domain.UpdateMenuRequest{Days: days})
	req := httptest.NewRequest("PUT", "/menus/current", bytes.NewReader(body))
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.UpdateCurrent(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestMenuHandler_UpdateCurrent_NoAuth(t *testing.T) {
	h := newTestMenuHandler(defaultMenuStorageForHandler(), recipeStorageWithRecipes())

	days := []domain.MenuDay{{Date: "2026-06-14", RecipeID: "rec_1", Servings: 4}}
	body, _ := json.Marshal(domain.UpdateMenuRequest{Days: days})
	req := httptest.NewRequest("PUT", "/menus/current", bytes.NewReader(body))
	// no setAuthContext
	rr := httptest.NewRecorder()
	h.UpdateCurrent(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestMenuHandler_UpdateCurrent_NoActiveMenu(t *testing.T) {
	menuStore := defaultMenuStorageForHandler()
	menuStore.currentMenu = nil
	h := newTestMenuHandler(menuStore, recipeStorageWithRecipes())

	days := []domain.MenuDay{{Date: "2026-06-14", RecipeID: "rec_1", Servings: 4}}
	body, _ := json.Marshal(domain.UpdateMenuRequest{Days: days})
	req := httptest.NewRequest("PUT", "/menus/current", bytes.NewReader(body))
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.UpdateCurrent(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "no_active_menu" {
		t.Errorf("expected no_active_menu, got %q", errResp["error"])
	}
}

func TestMenuHandler_UpdateCurrent_InvalidDays(t *testing.T) {
	h := newTestMenuHandler(defaultMenuStorageForHandler(), recipeStorageWithRecipes())

	// empty Days slice triggers ErrInvalidDays
	body, _ := json.Marshal(domain.UpdateMenuRequest{Days: []domain.MenuDay{}})
	req := httptest.NewRequest("PUT", "/menus/current", bytes.NewReader(body))
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.UpdateCurrent(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "invalid_days" {
		t.Errorf("expected invalid_days, got %q", errResp["error"])
	}
}

// --- GetCurrent tests ---

func TestMenuHandler_GetCurrent_Happy(t *testing.T) {
	h := newTestMenuHandler(defaultMenuStorageForHandler(), recipeStorageWithRecipes())

	req := httptest.NewRequest("GET", "/menus/current", nil)
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.GetCurrent(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestMenuHandler_GetCurrent_NoAuth(t *testing.T) {
	h := newTestMenuHandler(defaultMenuStorageForHandler(), recipeStorageWithRecipes())

	req := httptest.NewRequest("GET", "/menus/current", nil)
	// no setAuthContext
	rr := httptest.NewRecorder()
	h.GetCurrent(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestMenuHandler_GetCurrent_NoActiveMenu(t *testing.T) {
	menuStore := defaultMenuStorageForHandler()
	menuStore.currentMenu = nil
	h := newTestMenuHandler(menuStore, recipeStorageWithRecipes())

	req := httptest.NewRequest("GET", "/menus/current", nil)
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.GetCurrent(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "no_active_menu" {
		t.Errorf("expected no_active_menu, got %q", errResp["error"])
	}
}

func TestMenuHandler_GetCurrent_StorageError(t *testing.T) {
	menuStore := defaultMenuStorageForHandler()
	menuStore.getCurrentErr = errors.New("db exploded")
	h := newTestMenuHandler(menuStore, recipeStorageWithRecipes())

	req := httptest.NewRequest("GET", "/menus/current", nil)
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.GetCurrent(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rr.Code, rr.Body.String())
	}
}
