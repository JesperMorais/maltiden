package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/internal/storage/sqlite"
)

const accountTestPassword = "Testpassword123"

type accountTestEnv struct {
	db        *sql.DB
	handler   *AccountHandler
	auth      *services.AuthService
	household *services.HouseholdService
}

func newAccountTestEnv(t *testing.T) *accountTestEnv {
	t.Helper()
	db := setupHouseholdTestDB(t)
	users := sqlite.NewUserStorage(db)
	households := sqlite.NewHouseholdStorage(db)
	return &accountTestEnv{
		db:        db,
		handler:   NewAccountHandler(services.NewAccountService(sqlite.NewAccountStorage(db), users)),
		auth:      services.NewAuthService(db, users, households, setupHouseholdTestJWTService(t)),
		household: services.NewHouseholdService(households, users),
	}
}

func (e *accountTestEnv) exec(t *testing.T, query string, args ...any) {
	t.Helper()
	if _, err := e.db.Exec(query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

func (e *accountTestEnv) count(t *testing.T, query string, args ...any) int {
	t.Helper()
	var n int
	if err := e.db.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("count %q: %v", query, err)
	}
	return n
}

// seedHouseholdData gives a household one of everything the app stores.
func (e *accountTestEnv) seedHouseholdData(t *testing.T, userID, householdID, suffix string) {
	t.Helper()
	e.exec(t, `INSERT INTO recipes (id, name, servings, ingredients, instructions, household_id)
		VALUES (?, 'Egen gryta', 4, '[{"name":"Lök","amount":1,"unit":"st"}]', '["Koka"]', ?)`,
		"rec_own_"+suffix, householdID)
	e.exec(t, `INSERT INTO menus (id, household_id) VALUES (?, ?)`, "menu_"+suffix, householdID)
	e.exec(t, `INSERT INTO menu_days (id, menu_id, date, recipe_id) VALUES (?, ?, '2026-09-24', ?)`,
		"md_"+suffix, "menu_"+suffix, "rec_own_"+suffix)
	e.exec(t, `INSERT INTO custom_shopping_items (id, menu_id, household_id, name) VALUES (?, ?, ?, 'Kaffe')`,
		"csi_"+suffix, "menu_"+suffix, householdID)
	e.exec(t, `INSERT INTO menu_preferences (household_id, excluded_tags) VALUES (?, '["fisk"]')`, householdID)
	e.exec(t, `INSERT INTO feedback (id, user_id, household_id, mood, comment) VALUES (?, ?, ?, 'good', 'Bra app')`,
		"fb_"+suffix, userID, householdID)
}

func (e *accountTestEnv) deleteAccount(t *testing.T, userID, householdID, password string) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"password": password})
	req := httptest.NewRequest(http.MethodDelete, "/account", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, userID, householdID)
	rr := httptest.NewRecorder()
	e.handler.Delete(rr, req)
	return rr
}

func (e *accountTestEnv) join(t *testing.T, owner, joiner *domain.AuthResponse) {
	t.Helper()
	invite, err := e.household.CreateInvite(owner.User.HouseholdID)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if _, err := e.household.JoinHousehold(joiner.User.ID, domain.JoinHouseholdRequest{Code: invite.Code}); err != nil {
		t.Fatalf("JoinHousehold: %v", err)
	}
}

func TestAccountDelete_SoleMemberErasesHousehold(t *testing.T) {
	env := newAccountTestEnv(t)
	anna := createHouseholdTestUser(t, env.auth, "anna@test.com", "Anna")
	bo := createHouseholdTestUser(t, env.auth, "bo@test.com", "Bo")
	hh := anna.User.HouseholdID
	env.seedHouseholdData(t, anna.User.ID, hh, "a")
	env.exec(t, `INSERT INTO password_reset_tokens (token, user_id, expires_at) VALUES ('tok', ?, '2099-01-01')`, anna.User.ID)
	// Bo's menu points at Anna's household recipe: deletion must detach it,
	// not fail on the foreign key.
	env.exec(t, `INSERT INTO menus (id, household_id) VALUES ('menu_bo', ?)`, bo.User.HouseholdID)
	env.exec(t, `INSERT INTO menu_days (id, menu_id, date, recipe_id) VALUES ('md_bo', 'menu_bo', '2026-09-24', 'rec_own_a')`)
	seedBefore := env.count(t, `SELECT COUNT(*) FROM recipes WHERE household_id IS NULL`)

	rr := env.deleteAccount(t, anna.User.ID, hh, accountTestPassword)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204: %s", rr.Code, rr.Body.String())
	}

	checks := map[string]int{
		`SELECT COUNT(*) FROM users WHERE id = '` + anna.User.ID + `'`:                      0,
		`SELECT COUNT(*) FROM households WHERE id = '` + hh + `'`:                           0,
		`SELECT COUNT(*) FROM household_members WHERE household_id = '` + hh + `'`:          0,
		`SELECT COUNT(*) FROM recipes WHERE household_id = '` + hh + `'`:                    0,
		`SELECT COUNT(*) FROM menus WHERE household_id = '` + hh + `'`:                      0,
		`SELECT COUNT(*) FROM menu_days WHERE menu_id = 'menu_a'`:                           0,
		`SELECT COUNT(*) FROM custom_shopping_items WHERE household_id = '` + hh + `'`:      0,
		`SELECT COUNT(*) FROM menu_preferences WHERE household_id = '` + hh + `'`:           0,
		`SELECT COUNT(*) FROM feedback WHERE user_id = '` + anna.User.ID + `'`:              0,
		`SELECT COUNT(*) FROM password_reset_tokens WHERE user_id = '` + anna.User.ID + `'`: 0,
		`SELECT COUNT(*) FROM menu_days WHERE id = 'md_bo' AND recipe_id IS NULL`:           1,
		`SELECT COUNT(*) FROM users WHERE id = '` + bo.User.ID + `'`:                        1,
		`SELECT COUNT(*) FROM recipes WHERE household_id IS NULL`:                           seedBefore,
	}
	for q, want := range checks {
		if got := env.count(t, q); got != want {
			t.Errorf("%s = %d, want %d", q, got, want)
		}
	}
}

func TestAccountDelete_SharedHouseholdKeepsDataAndTransfersOwnership(t *testing.T) {
	env := newAccountTestEnv(t)
	anna := createHouseholdTestUser(t, env.auth, "anna@test.com", "Anna")
	bo := createHouseholdTestUser(t, env.auth, "bo@test.com", "Bo")
	env.join(t, anna, bo)
	hh := anna.User.HouseholdID
	env.seedHouseholdData(t, anna.User.ID, hh, "a")

	rr := env.deleteAccount(t, anna.User.ID, hh, accountTestPassword)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204: %s", rr.Code, rr.Body.String())
	}

	if n := env.count(t, `SELECT COUNT(*) FROM users WHERE id = ?`, anna.User.ID); n != 0 {
		t.Error("user row still present")
	}
	if n := env.count(t, `SELECT COUNT(*) FROM feedback WHERE user_id = ?`, anna.User.ID); n != 0 {
		t.Error("user's feedback still present")
	}
	if n := env.count(t, `SELECT COUNT(*) FROM menus WHERE household_id = ?`, hh); n != 1 {
		t.Errorf("shared household menus = %d, want kept (1)", n)
	}
	var role string
	if err := env.db.QueryRow(`SELECT role FROM household_members WHERE household_id = ? AND user_id = ?`, hh, bo.User.ID).Scan(&role); err != nil {
		t.Fatalf("load Bo's role: %v", err)
	}
	if role != "owner" {
		t.Errorf("remaining member role = %q, want owner", role)
	}
}

func TestAccountDelete_WrongPasswordKeepsAccount(t *testing.T) {
	env := newAccountTestEnv(t)
	anna := createHouseholdTestUser(t, env.auth, "anna@test.com", "Anna")

	rr := env.deleteAccount(t, anna.User.ID, anna.User.HouseholdID, "wrong-password")
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403: %s", rr.Code, rr.Body.String())
	}
	if n := env.count(t, `SELECT COUNT(*) FROM users WHERE id = ?`, anna.User.ID); n != 1 {
		t.Error("account deleted despite wrong password")
	}
}

func TestAccountDelete_RequiresAuthAndPassword(t *testing.T) {
	env := newAccountTestEnv(t)
	if rr := env.deleteAccount(t, "", "", accountTestPassword); rr.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated status = %d, want 401", rr.Code)
	}
	anna := createHouseholdTestUser(t, env.auth, "anna@test.com", "Anna")
	if rr := env.deleteAccount(t, anna.User.ID, anna.User.HouseholdID, ""); rr.Code != http.StatusBadRequest {
		t.Errorf("empty password status = %d, want 400", rr.Code)
	}
}

func TestAccountExport_ContainsUserAndHouseholdDataButNoCredentials(t *testing.T) {
	env := newAccountTestEnv(t)
	anna := createHouseholdTestUser(t, env.auth, "anna@test.com", "Anna")
	env.seedHouseholdData(t, anna.User.ID, anna.User.HouseholdID, "a")

	req := setAuthContext(httptest.NewRequest(http.MethodGet, "/account/export", nil), anna.User.ID, anna.User.HouseholdID)
	rr := httptest.NewRecorder()
	env.handler.Export(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rr.Code, rr.Body.String())
	}
	if cd := rr.Header().Get("Content-Disposition"); !strings.HasPrefix(cd, "attachment;") {
		t.Errorf("Content-Disposition = %q, want attachment", cd)
	}
	body := rr.Body.String()
	for _, secret := range []string{"password_hash", "token_version", "$2a$"} {
		if strings.Contains(body, secret) {
			t.Errorf("export leaks %q", secret)
		}
	}

	var doc struct {
		User struct {
			Email string `json:"email"`
		} `json:"user"`
		Feedback  []map[string]any `json:"feedback"`
		Household struct {
			Name    string `json:"name"`
			Recipes []struct {
				Name        string            `json:"name"`
				Ingredients []json.RawMessage `json:"ingredients"`
			} `json:"recipes"`
			Menus               []map[string]any `json:"menus"`
			MenuDays            []map[string]any `json:"menuDays"`
			CustomShoppingItems []map[string]any `json:"customShoppingItems"`
			MenuPreferences     []map[string]any `json:"menuPreferences"`
			Members             []map[string]any `json:"members"`
		} `json:"household"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &doc); err != nil {
		t.Fatalf("decode export: %v", err)
	}
	if doc.User.Email != "anna@test.com" {
		t.Errorf("email = %q", doc.User.Email)
	}
	h := doc.Household
	if len(h.Recipes) != 1 || h.Recipes[0].Name != "Egen gryta" || len(h.Recipes[0].Ingredients) != 1 {
		t.Errorf("recipes = %+v, want the household recipe with structured ingredients", h.Recipes)
	}
	if len(h.Menus) != 1 || len(h.MenuDays) != 1 || len(h.CustomShoppingItems) != 1 ||
		len(h.MenuPreferences) != 1 || len(h.Members) != 1 || len(doc.Feedback) != 1 {
		t.Errorf("section sizes menus=%d days=%d items=%d prefs=%d members=%d feedback=%d, want 1 each",
			len(h.Menus), len(h.MenuDays), len(h.CustomShoppingItems), len(h.MenuPreferences), len(h.Members), len(doc.Feedback))
	}
}
