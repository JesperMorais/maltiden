package handlers

import (
	"bytes"
	"encoding/json"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/pkg/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockMenuStorage implements domain.MenuRepository for handler-level tests.
type mockMenuStorage struct {
	currentMenu    *domain.Menu
	getErr         error
	updateErr      error
	updatedMenu    *domain.Menu
}

func (m *mockMenuStorage) Create(menu *domain.Menu) error { return nil }

func (m *mockMenuStorage) Update(menu *domain.Menu) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedMenu = menu
	return nil
}

func (m *mockMenuStorage) GetCurrentByHousehold(householdID string) (*domain.Menu, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.currentMenu, nil
}

func (m *mockMenuStorage) GetByID(id string) (*domain.Menu, error)                      { return nil, nil }
func (m *mockMenuStorage) GetHouseholdIDByMenuID(menuID string) (string, error)          { return "", nil }

// mockRecipeStorage implements domain.RecipeRepository (needed by MenuService).
type mockRecipeStorage struct{}

func (m *mockRecipeStorage) GetAll(_ *domain.RecipeFilter, _ string) ([]domain.RecipeSummary, error) {
	return nil, nil
}
func (m *mockRecipeStorage) GetAllPaginated(_ *domain.RecipeFilter, _ string, _, _ int) ([]domain.RecipeSummary, int, error) {
	return nil, 0, nil
}
func (m *mockRecipeStorage) GetByID(_ string) (*domain.Recipe, error)                        { return nil, nil }
func (m *mockRecipeStorage) GetByIDs(ids []string) (map[string]*domain.Recipe, error)         { return map[string]*domain.Recipe{}, nil }
func (m *mockRecipeStorage) Create(_ *domain.Recipe) error                                    { return nil }
func (m *mockRecipeStorage) Update(_ *domain.Recipe) error                                    { return nil }
func (m *mockRecipeStorage) Delete(_ string) error                                            { return nil }

func newTestMenuHandler(menuStore domain.MenuRepository) *MenuHandler {
	svc := services.NewMenuService(menuStore, &mockRecipeStorage{})
	return NewMenuHandler(svc)
}

// TestMenusCurrentUpdate_200_HappyPath: authed member with valid body, existing menu → 200.
func TestMenusCurrentUpdate_200_HappyPath(t *testing.T) {
	existing := &domain.Menu{
		ID:          "menu_test-1111-2222-3333-444444444444",
		HouseholdID: "hh_test-aaaa-bbbb-cccc-dddddddddddd",
		Days: []domain.MenuDay{
			{Date: "2026-01-01", Servings: 4},
		},
	}
	store := &mockMenuStorage{currentMenu: existing}
	h := newTestMenuHandler(store)

	body := `{"days":[{"date":"2026-01-02","servings":2}]}`
	req := httptest.NewRequest(http.MethodPut, "/menus/current", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithAuthContext(req.Context(),
		"usr_test-aaaa-bbbb-cccc-dddddddddddd",
		"hh_test-aaaa-bbbb-cccc-dddddddddddd"))

	rr := httptest.NewRecorder()
	h.UpdateCurrent(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp domain.MenuResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != existing.ID {
		t.Errorf("expected menu ID %q, got %q", existing.ID, resp.ID)
	}
	if len(resp.Days) != 1 {
		t.Errorf("expected 1 day in response, got %d", len(resp.Days))
	}
}

// TestMenusCurrentUpdate_400_InvalidPayload: authed member, empty days array → 400 invalid_days.
func TestMenusCurrentUpdate_400_InvalidPayload(t *testing.T) {
	existing := &domain.Menu{
		ID:          "menu_test-1111-2222-3333-444444444444",
		HouseholdID: "hh_test-aaaa-bbbb-cccc-dddddddddddd",
		Days:        []domain.MenuDay{{Date: "2026-01-01", Servings: 4}},
	}
	store := &mockMenuStorage{currentMenu: existing}
	h := newTestMenuHandler(store)

	// Empty days array → service returns ErrInvalidDays
	body := `{"days":[]}`
	req := httptest.NewRequest(http.MethodPut, "/menus/current", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithAuthContext(req.Context(),
		"usr_test-aaaa-bbbb-cccc-dddddddddddd",
		"hh_test-aaaa-bbbb-cccc-dddddddddddd"))

	rr := httptest.NewRecorder()
	h.UpdateCurrent(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp) //nolint:errcheck
	if errResp["error"] != "invalid_days" {
		t.Errorf("expected error 'invalid_days', got %q", errResp["error"])
	}
}

// TestMenusCurrentUpdate_401_NoJWT: no Authorization header → handler sees empty householdID → 401.
func TestMenusCurrentUpdate_401_NoJWT(t *testing.T) {
	store := &mockMenuStorage{}
	h := newTestMenuHandler(store)

	body := `{"days":[{"date":"2026-01-02","servings":2}]}`
	req := httptest.NewRequest(http.MethodPut, "/menus/current", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	// No auth context set — simulates missing JWT (middleware would have rejected, or context is empty)

	rr := httptest.NewRecorder()
	h.UpdateCurrent(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp) //nolint:errcheck
	if errResp["error"] != "unauthorized" {
		t.Errorf("expected error 'unauthorized', got %q", errResp["error"])
	}
}

// TestMenusCurrentUpdate_NonMember_NoMenu: user from a different household → no menu found → 404.
// The handler does not implement a 403; cross-household access results in a "no_active_menu" 404
// because the storage lookup uses the JWT's householdID, which has no menu.
func TestMenusCurrentUpdate_NonMember_NoMenu(t *testing.T) {
	// menu exists for hh_owner, but the requester belongs to hh_other
	store := &mockMenuStorage{currentMenu: nil} // GetCurrentByHousehold returns nil for hh_other
	h := newTestMenuHandler(store)

	body := `{"days":[{"date":"2026-01-02","servings":2}]}`
	req := httptest.NewRequest(http.MethodPut, "/menus/current", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithAuthContext(req.Context(),
		"usr_test-xxxx-yyyy-zzzz-000000000000",
		"hh_test-other-household-000000000000"))

	rr := httptest.NewRecorder()
	h.UpdateCurrent(rr, req)

	// Handler has no 403 path; cross-household access returns 404 (no_active_menu).
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-member (no 403 implemented), got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp) //nolint:errcheck
	if errResp["error"] != "no_active_menu" {
		t.Errorf("expected error 'no_active_menu', got %q", errResp["error"])
	}
}
