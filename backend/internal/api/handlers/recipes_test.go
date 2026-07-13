package handlers

import (
	"bytes"
	"encoding/json"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/internal/storage/sqlite"
	"maltiden/pkg/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecipeHandler_Create_NegativeAmount(t *testing.T) {
	db := setupTestDB(t)
	recipeStorage := sqlite.NewRecipeStorage(db)
	recipeService := services.NewRecipeService(recipeStorage)
	h := NewRecipeHandler(recipeService)

	body, _ := json.Marshal(domain.CreateRecipeRequest{
		Name:         "Test",
		Servings:     2,
		Ingredients:  []domain.Ingredient{{Name: "Salt", Amount: -500, Unit: "g"}},
		Instructions: []string{"Mix"},
	})

	req := httptest.NewRequest("POST", "/recipes", bytes.NewReader(body))
	ctx := middleware.WithAuthContext(req.Context(), "usr_test", "hh_test")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["error"] != "invalid_amount" {
		t.Fatalf(`expected {"error":"invalid_amount"}, got %s`, rr.Body.String())
	}
}
