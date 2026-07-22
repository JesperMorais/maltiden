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

// recordingPrefsStorage is a MenuPreferencesRepository that records the
// householdID passed to each call, so handler tests can assert that scoping
// derives from the auth context and never from the request body/params.
type recordingPrefsStorage struct {
	stored map[string]*domain.MenuPreferences

	getErr    error
	upsertErr error

	gotGetHousehold    string
	gotUpsertHousehold string
	upsertCalls        int
}

func newRecordingPrefsStorage() *recordingPrefsStorage {
	return &recordingPrefsStorage{stored: map[string]*domain.MenuPreferences{}}
}

func (m *recordingPrefsStorage) Get(householdID string) (*domain.MenuPreferences, error) {
	m.gotGetHousehold = householdID
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.stored[householdID], nil
}

func (m *recordingPrefsStorage) Upsert(prefs *domain.MenuPreferences) error {
	m.upsertCalls++
	m.gotUpsertHousehold = prefs.HouseholdID
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.stored[prefs.HouseholdID] = prefs
	return nil
}

// newPrefsMenuHandler wires a MenuHandler backed by the recording prefs store
// (menu + recipe stores are the standard handler-test fakes).
func newPrefsMenuHandler(prefs domain.MenuPreferencesRepository) *MenuHandler {
	svc := services.NewMenuService(defaultMenuStorageForHandler(), recipeStorageWithRecipes(), prefs)
	return NewMenuHandler(svc)
}

// validPrefsRequest is a request body that passes all Validate() bounds.
func validPrefsRequest() domain.UpdateMenuPreferencesRequest {
	return domain.UpdateMenuPreferencesRequest{
		ExcludedTags:    []string{"fish", "pork"},
		DefaultDays:     5,
		DefaultServings: 4,
		VegetarianDays:  2,
	}
}

// --- GetPreferences tests ---

func TestMenuHandler_GetPreferences_NoneSetReturnsDefaults(t *testing.T) {
	store := newRecordingPrefsStorage() // empty: Get returns nil -> defaults
	h := newPrefsMenuHandler(store)

	req := httptest.NewRequest("GET", "/menus/preferences", nil)
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.GetPreferences(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var prefs domain.MenuPreferences
	if err := json.NewDecoder(rr.Body).Decode(&prefs); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	def := domain.DefaultMenuPreferences(testHouseholdID)
	if prefs.DefaultDays != def.DefaultDays {
		t.Errorf("expected default days %d, got %d", def.DefaultDays, prefs.DefaultDays)
	}
	if prefs.DefaultServings != def.DefaultServings {
		t.Errorf("expected default servings %d, got %d", def.DefaultServings, prefs.DefaultServings)
	}
	if prefs.VegetarianDays != def.VegetarianDays {
		t.Errorf("expected default vegetarian days %d, got %d", def.VegetarianDays, prefs.VegetarianDays)
	}
	if len(prefs.ExcludedTags) != 0 {
		t.Errorf("expected no excluded tags, got %v", prefs.ExcludedTags)
	}
	if prefs.HouseholdID != testHouseholdID {
		t.Errorf("expected household %q, got %q", testHouseholdID, prefs.HouseholdID)
	}
}

func TestMenuHandler_GetPreferences_ReturnsStored(t *testing.T) {
	store := newRecordingPrefsStorage()
	store.stored[testHouseholdID] = &domain.MenuPreferences{
		HouseholdID:     testHouseholdID,
		ExcludedTags:    []string{"shellfish"},
		DefaultDays:     3,
		DefaultServings: 6,
		VegetarianDays:  1,
	}
	h := newPrefsMenuHandler(store)

	req := httptest.NewRequest("GET", "/menus/preferences", nil)
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.GetPreferences(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var prefs domain.MenuPreferences
	json.NewDecoder(rr.Body).Decode(&prefs)
	if prefs.DefaultDays != 3 || prefs.DefaultServings != 6 || prefs.VegetarianDays != 1 {
		t.Errorf("stored prefs not returned: %+v", prefs)
	}
}

func TestMenuHandler_GetPreferences_NoAuth(t *testing.T) {
	h := newPrefsMenuHandler(newRecordingPrefsStorage())

	req := httptest.NewRequest("GET", "/menus/preferences", nil)
	// no setAuthContext
	rr := httptest.NewRecorder()
	h.GetPreferences(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestMenuHandler_GetPreferences_StorageError(t *testing.T) {
	store := newRecordingPrefsStorage()
	store.getErr = errors.New("db exploded")
	h := newPrefsMenuHandler(store)

	req := httptest.NewRequest("GET", "/menus/preferences", nil)
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.GetPreferences(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rr.Code, rr.Body.String())
	}
}

// GetPreferences must scope by the auth-context household, never by anything
// a caller could supply, so it cannot read another household's prefs.
func TestMenuHandler_GetPreferences_ScopesToAuthHousehold(t *testing.T) {
	store := newRecordingPrefsStorage()
	h := newPrefsMenuHandler(store)

	req := httptest.NewRequest("GET", "/menus/preferences?householdId=hh_victim", nil)
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.GetPreferences(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if store.gotGetHousehold != testHouseholdID {
		t.Errorf("IDOR: Get scoped to %q, expected auth household %q", store.gotGetHousehold, testHouseholdID)
	}
}

// --- UpdatePreferences tests ---

func TestMenuHandler_UpdatePreferences_Happy(t *testing.T) {
	store := newRecordingPrefsStorage()
	h := newPrefsMenuHandler(store)

	body, _ := json.Marshal(validPrefsRequest())
	req := httptest.NewRequest("PUT", "/menus/preferences", bytes.NewReader(body))
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.UpdatePreferences(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var prefs domain.MenuPreferences
	if err := json.NewDecoder(rr.Body).Decode(&prefs); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if prefs.DefaultDays != 5 || prefs.DefaultServings != 4 || prefs.VegetarianDays != 2 {
		t.Errorf("unexpected returned prefs: %+v", prefs)
	}
	if store.upsertCalls != 1 {
		t.Errorf("expected 1 upsert, got %d", store.upsertCalls)
	}
	saved := store.stored[testHouseholdID]
	if saved == nil {
		t.Fatalf("nothing persisted for household %q", testHouseholdID)
	}
	if len(saved.ExcludedTags) != 2 {
		t.Errorf("expected 2 excluded tags persisted, got %v", saved.ExcludedTags)
	}
}

func TestMenuHandler_UpdatePreferences_NoAuth(t *testing.T) {
	store := newRecordingPrefsStorage()
	h := newPrefsMenuHandler(store)

	body, _ := json.Marshal(validPrefsRequest())
	req := httptest.NewRequest("PUT", "/menus/preferences", bytes.NewReader(body))
	// no setAuthContext
	rr := httptest.NewRecorder()
	h.UpdatePreferences(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if store.upsertCalls != 0 {
		t.Errorf("unauthenticated request must not persist; got %d upserts", store.upsertCalls)
	}
}

func TestMenuHandler_UpdatePreferences_InvalidDays(t *testing.T) {
	tests := []struct {
		name string
		days int
	}{
		{"zero", 0},
		{"too_high", 32},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := newRecordingPrefsStorage()
			h := newPrefsMenuHandler(store)

			r := validPrefsRequest()
			r.DefaultDays = tc.days
			r.VegetarianDays = 0
			body, _ := json.Marshal(r)
			req := httptest.NewRequest("PUT", "/menus/preferences", bytes.NewReader(body))
			req = setAuthContext(req, testUserID, testHouseholdID)
			rr := httptest.NewRecorder()
			h.UpdatePreferences(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
			}
			assertErrorCode(t, rr, "invalid_days")
			if store.upsertCalls != 0 {
				t.Errorf("invalid request must not persist; got %d upserts", store.upsertCalls)
			}
		})
	}
}

func TestMenuHandler_UpdatePreferences_InvalidServings(t *testing.T) {
	tests := []struct {
		name     string
		servings int
	}{
		{"zero", 0},
		{"too_high", 101},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := newRecordingPrefsStorage()
			h := newPrefsMenuHandler(store)

			r := validPrefsRequest()
			r.DefaultServings = tc.servings
			body, _ := json.Marshal(r)
			req := httptest.NewRequest("PUT", "/menus/preferences", bytes.NewReader(body))
			req = setAuthContext(req, testUserID, testHouseholdID)
			rr := httptest.NewRecorder()
			h.UpdatePreferences(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
			}
			assertErrorCode(t, rr, "invalid_servings")
		})
	}
}

func TestMenuHandler_UpdatePreferences_InvalidVegetarianDays(t *testing.T) {
	store := newRecordingPrefsStorage()
	h := newPrefsMenuHandler(store)

	// VegetarianDays must not exceed DefaultDays.
	r := validPrefsRequest()
	r.DefaultDays = 3
	r.VegetarianDays = 5
	body, _ := json.Marshal(r)
	req := httptest.NewRequest("PUT", "/menus/preferences", bytes.NewReader(body))
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.UpdatePreferences(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	assertErrorCode(t, rr, "invalid_vegetarian_days")
}

func TestMenuHandler_UpdatePreferences_TooManyExcludedTags(t *testing.T) {
	store := newRecordingPrefsStorage()
	h := newPrefsMenuHandler(store)

	tags := make([]string, 51)
	for i := range tags {
		tags[i] = "tag"
	}
	r := validPrefsRequest()
	r.ExcludedTags = tags
	body, _ := json.Marshal(r)
	req := httptest.NewRequest("PUT", "/menus/preferences", bytes.NewReader(body))
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.UpdatePreferences(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	assertErrorCode(t, rr, "too_many_excluded_tags")
}

func TestMenuHandler_UpdatePreferences_InvalidExcludedTag(t *testing.T) {
	tests := []struct {
		name string
		tag  string
	}{
		{"empty", ""},
		{"too_long", string(make([]byte, 51))},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := newRecordingPrefsStorage()
			h := newPrefsMenuHandler(store)

			r := validPrefsRequest()
			r.ExcludedTags = []string{tc.tag}
			body, _ := json.Marshal(r)
			req := httptest.NewRequest("PUT", "/menus/preferences", bytes.NewReader(body))
			req = setAuthContext(req, testUserID, testHouseholdID)
			rr := httptest.NewRecorder()
			h.UpdatePreferences(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
			}
			assertErrorCode(t, rr, "invalid_excluded_tag")
		})
	}
}

func TestMenuHandler_UpdatePreferences_StorageError(t *testing.T) {
	store := newRecordingPrefsStorage()
	store.upsertErr = errors.New("db exploded")
	h := newPrefsMenuHandler(store)

	body, _ := json.Marshal(validPrefsRequest())
	req := httptest.NewRequest("PUT", "/menus/preferences", bytes.NewReader(body))
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.UpdatePreferences(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rr.Code, rr.Body.String())
	}
}

// UpdatePreferences must persist under the auth-context household, never one a
// caller could inject, so it cannot overwrite another household's prefs (IDOR).
func TestMenuHandler_UpdatePreferences_ScopesToAuthHousehold(t *testing.T) {
	store := newRecordingPrefsStorage()
	h := newPrefsMenuHandler(store)

	// Craft a body that smuggles a foreign householdId; the request DTO has no
	// such field, but a malicious client could still send it. It must be ignored.
	rawBody := `{"householdId":"hh_victim","excludedTags":["fish"],"defaultDays":5,"defaultServings":4,"vegetarianDays":2}`
	req := httptest.NewRequest("PUT", "/menus/preferences?householdId=hh_victim", bytes.NewBufferString(rawBody))
	req = setAuthContext(req, testUserID, testHouseholdID)
	rr := httptest.NewRecorder()
	h.UpdatePreferences(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if store.gotUpsertHousehold != testHouseholdID {
		t.Errorf("IDOR: persisted under %q, expected auth household %q", store.gotUpsertHousehold, testHouseholdID)
	}
	if _, leaked := store.stored["hh_victim"]; leaked {
		t.Errorf("IDOR: prefs leaked into foreign household hh_victim")
	}
	saved := store.stored[testHouseholdID]
	if saved == nil || saved.HouseholdID != testHouseholdID {
		t.Errorf("expected prefs persisted under auth household, got %+v", saved)
	}
}

// assertErrorCode decodes the standard {"error": "..."} envelope and checks it.
func assertErrorCode(t *testing.T, rr *httptest.ResponseRecorder, want string) {
	t.Helper()
	var errResp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if errResp["error"] != want {
		t.Errorf("expected error %q, got %q", want, errResp["error"])
	}
}
