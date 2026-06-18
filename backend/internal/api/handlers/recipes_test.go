package handlers

import (
	"bytes"
	"encoding/json"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockRecipeStorage implements domain.RecipeRepository for handler tests.
type mockRecipeStorage struct {
	recipes   map[string]*domain.Recipe
	createErr error
	updateErr error
	deleteErr error
	getErr    error
}

func newMockRecipeStorage() *mockRecipeStorage {
	return &mockRecipeStorage{recipes: make(map[string]*domain.Recipe)}
}

func (m *mockRecipeStorage) GetAll(filter *domain.RecipeFilter, householdID string) ([]domain.RecipeSummary, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	var out []domain.RecipeSummary
	for _, r := range m.recipes {
		if householdID != "" && r.HouseholdID != "" && r.HouseholdID != householdID {
			continue
		}
		out = append(out, domain.RecipeSummary{ID: r.ID, Name: r.Name, Servings: r.Servings, Emoji: r.Emoji, Tags: r.Tags})
	}
	return out, nil
}

func (m *mockRecipeStorage) GetAllPaginated(filter *domain.RecipeFilter, householdID string, limit, offset int) ([]domain.RecipeSummary, int, error) {
	summaries, err := m.GetAll(filter, householdID)
	return summaries, len(summaries), err
}

func (m *mockRecipeStorage) GetByID(id string) (*domain.Recipe, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	r, ok := m.recipes[id]
	if !ok {
		return nil, nil
	}
	return r, nil
}

func (m *mockRecipeStorage) GetByIDs(ids []string) (map[string]*domain.Recipe, error) {
	out := make(map[string]*domain.Recipe)
	for _, id := range ids {
		if r, ok := m.recipes[id]; ok {
			out[id] = r
		}
	}
	return out, nil
}

func (m *mockRecipeStorage) Create(recipe *domain.Recipe) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.recipes[recipe.ID] = recipe
	return nil
}

func (m *mockRecipeStorage) Update(recipe *domain.Recipe) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.recipes[recipe.ID] = recipe
	return nil
}

func (m *mockRecipeStorage) Delete(id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.recipes, id)
	return nil
}

func newTestRecipeHandler(store *mockRecipeStorage) *RecipeHandler {
	svc := services.NewRecipeService(store, nil)
	return NewRecipeHandler(svc)
}

func validCreateBody() []byte {
	body, _ := json.Marshal(map[string]interface{}{
		"name":         "Köttbullar",
		"servings":     4,
		"ingredients":  []map[string]interface{}{{"name": "nötkött", "amount": 500, "unit": "g"}},
		"instructions": []string{"Blanda ingredienserna", "Rulla köttbullar", "Stek tills genomstekta"},
	})
	return body
}

func TestRecipeHandler_Create_Valid(t *testing.T) {
	store := newMockRecipeStorage()
	h := newTestRecipeHandler(store)

	req := httptest.NewRequest("POST", "/recipes", bytes.NewBuffer(validCreateBody()))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1111-2222-3333-444444444444", "hh_test-1111-2222-3333-444444444444")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp domain.CreateRecipeResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID == "" {
		t.Fatal("expected non-empty recipe ID")
	}
}

func TestRecipeHandler_Create_MissingName(t *testing.T) {
	store := newMockRecipeStorage()
	h := newTestRecipeHandler(store)

	body, _ := json.Marshal(map[string]interface{}{
		"servings":     2,
		"ingredients":  []map[string]interface{}{{"name": "mjöl", "amount": 200, "unit": "g"}},
		"instructions": []string{"Blanda"},
	})
	req := httptest.NewRequest("POST", "/recipes", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1111-2222-3333-444444444444", "hh_test-1111-2222-3333-444444444444")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "name_required" {
		t.Errorf("expected 'name_required', got %q", errResp["error"])
	}
}

func TestRecipeHandler_Create_InvalidServings(t *testing.T) {
	store := newMockRecipeStorage()
	h := newTestRecipeHandler(store)

	body, _ := json.Marshal(map[string]interface{}{
		"name":         "Soppa",
		"servings":     0,
		"ingredients":  []map[string]interface{}{{"name": "vatten", "amount": 1, "unit": "l"}},
		"instructions": []string{"Koka"},
	})
	req := httptest.NewRequest("POST", "/recipes", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1111-2222-3333-444444444444", "hh_test-1111-2222-3333-444444444444")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "invalid_servings" {
		t.Errorf("expected 'invalid_servings', got %q", errResp["error"])
	}
}

func TestRecipeHandler_Create_InvalidJSON(t *testing.T) {
	store := newMockRecipeStorage()
	h := newTestRecipeHandler(store)

	req := httptest.NewRequest("POST", "/recipes", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1111-2222-3333-444444444444", "hh_test-1111-2222-3333-444444444444")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRecipeHandler_GetAll_Empty(t *testing.T) {
	store := newMockRecipeStorage()
	h := newTestRecipeHandler(store)

	req := httptest.NewRequest("GET", "/recipes", nil)

	rr := httptest.NewRecorder()
	h.GetAll(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp domain.RecipesResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	// GetAll returns nil slice when empty — just check no error
}

func TestRecipeHandler_GetAll_WithRecipes(t *testing.T) {
	store := newMockRecipeStorage()
	store.recipes["rec_00000000-0000-0000-0000-000000000001"] = &domain.Recipe{
		ID: "rec_00000000-0000-0000-0000-000000000001", Name: "Pannkaka", Servings: 4,
		HouseholdID: "hh_test-1111-2222-3333-444444444444",
		Tags:        []string{},
		CreatedAt:   time.Now(),
	}
	h := newTestRecipeHandler(store)

	req := httptest.NewRequest("GET", "/recipes", nil)

	rr := httptest.NewRecorder()
	h.GetAll(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp domain.RecipesResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Recipes) != 1 {
		t.Errorf("expected 1 recipe, got %d", len(resp.Recipes))
	}
}

func TestRecipeHandler_GetByID_NotFound(t *testing.T) {
	store := newMockRecipeStorage()
	h := newTestRecipeHandler(store)

	req := httptest.NewRequest("GET", "/recipes/rec_00000000-0000-0000-0000-000000000099", nil)
	req.SetPathValue("id", "rec_00000000-0000-0000-0000-000000000099")

	rr := httptest.NewRecorder()
	h.GetByID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestRecipeHandler_GetByID_Found(t *testing.T) {
	const recID = "rec_00000000-0000-0000-0000-000000000002"
	store := newMockRecipeStorage()
	store.recipes[recID] = &domain.Recipe{
		ID: recID, Name: "Lasagne", Servings: 6,
		Tags: []string{"pasta"}, Ingredients: []domain.Ingredient{{Name: "köttfärs", Amount: 400, Unit: "g"}},
		Instructions: []string{"Laga köttfärssåsen"}, HouseholdID: "hh_test-1111-2222-3333-444444444444",
		CreatedAt: time.Now(),
	}
	h := newTestRecipeHandler(store)

	req := httptest.NewRequest("GET", "/recipes/"+recID, nil)
	req.SetPathValue("id", recID)

	rr := httptest.NewRecorder()
	h.GetByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var recipe domain.Recipe
	if err := json.NewDecoder(rr.Body).Decode(&recipe); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if recipe.Name != "Lasagne" {
		t.Errorf("expected 'Lasagne', got %q", recipe.Name)
	}
}

func TestRecipeHandler_Delete_NotFound(t *testing.T) {
	store := newMockRecipeStorage()
	h := newTestRecipeHandler(store)

	req := httptest.NewRequest("DELETE", "/recipes/rec_00000000-0000-0000-0000-000000000099", nil)
	req.SetPathValue("id", "rec_00000000-0000-0000-0000-000000000099")
	req = setAuthContext(req, "usr_test-1111-2222-3333-444444444444", "hh_test-1111-2222-3333-444444444444")

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestRecipeHandler_Delete_Forbidden(t *testing.T) {
	const recID = "rec_00000000-0000-0000-0000-000000000003"
	store := newMockRecipeStorage()
	store.recipes[recID] = &domain.Recipe{
		ID: recID, Name: "Gryta", Servings: 4,
		HouseholdID:  "hh_other-household",
		Tags:         []string{},
		Ingredients:  []domain.Ingredient{{Name: "kött", Amount: 500, Unit: "g"}},
		Instructions: []string{"Koka"},
		CreatedAt:    time.Now(),
	}
	h := newTestRecipeHandler(store)

	req := httptest.NewRequest("DELETE", "/recipes/"+recID, nil)
	req.SetPathValue("id", recID)
	req = setAuthContext(req, "usr_test-1111-2222-3333-444444444444", "hh_test-1111-2222-3333-444444444444")

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestRecipeHandler_Delete_Success(t *testing.T) {
	const recID = "rec_00000000-0000-0000-0000-000000000004"
	store := newMockRecipeStorage()
	store.recipes[recID] = &domain.Recipe{
		ID: recID, Name: "Smörgås", Servings: 1,
		HouseholdID:  "hh_test-1111-2222-3333-444444444444",
		Tags:         []string{},
		Ingredients:  []domain.Ingredient{{Name: "bröd", Amount: 2, Unit: "skivor"}},
		Instructions: []string{"Lägg på pålägg"},
		CreatedAt:    time.Now(),
	}
	h := newTestRecipeHandler(store)

	req := httptest.NewRequest("DELETE", "/recipes/"+recID, nil)
	req.SetPathValue("id", recID)
	req = setAuthContext(req, "usr_test-1111-2222-3333-444444444444", "hh_test-1111-2222-3333-444444444444")

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if _, ok := store.recipes[recID]; ok {
		t.Error("expected recipe to be removed from store")
	}
}

func TestRecipeHandler_Update_NotFound(t *testing.T) {
	store := newMockRecipeStorage()
	h := newTestRecipeHandler(store)

	body, _ := json.Marshal(map[string]interface{}{
		"name":         "Uppdaterad Rätt",
		"servings":     3,
		"ingredients":  []map[string]interface{}{{"name": "ingrediens", "amount": 100, "unit": "g"}},
		"instructions": []string{"Laga"},
	})
	const missingID = "rec_00000000-0000-0000-0000-000000000099"
	req := httptest.NewRequest("PUT", "/recipes/"+missingID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", missingID)
	req = setAuthContext(req, "usr_test-1111-2222-3333-444444444444", "hh_test-1111-2222-3333-444444444444")

	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestRecipeHandler_Update_Success(t *testing.T) {
	const recID = "rec_00000000-0000-0000-0000-000000000005"
	store := newMockRecipeStorage()
	store.recipes[recID] = &domain.Recipe{
		ID: recID, Name: "Original", Servings: 2,
		HouseholdID:  "hh_test-1111-2222-3333-444444444444",
		Tags:         []string{},
		Ingredients:  []domain.Ingredient{{Name: "vatten", Amount: 1, Unit: "l"}},
		Instructions: []string{"Koka"},
		CreatedAt:    time.Now(),
	}
	h := newTestRecipeHandler(store)

	body, _ := json.Marshal(map[string]interface{}{
		"name":         "Uppdaterad",
		"servings":     4,
		"ingredients":  []map[string]interface{}{{"name": "mjölk", "amount": 500, "unit": "ml"}},
		"instructions": []string{"Värm upp"},
	})
	req := httptest.NewRequest("PUT", "/recipes/"+recID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", recID)
	req = setAuthContext(req, "usr_test-1111-2222-3333-444444444444", "hh_test-1111-2222-3333-444444444444")

	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var recipe domain.Recipe
	json.NewDecoder(rr.Body).Decode(&recipe)
	if recipe.Name != "Uppdaterad" {
		t.Errorf("expected 'Uppdaterad', got %q", recipe.Name)
	}
	if recipe.Servings != 4 {
		t.Errorf("expected servings 4, got %d", recipe.Servings)
	}
}
