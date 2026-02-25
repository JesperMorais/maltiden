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
	"time"
)

// mockFeedbackStorage implements domain.FeedbackRepository for handler tests.
type mockFeedbackStorage struct {
	created     []*domain.Feedback
	recentCount int
	createErr   error
}

func (m *mockFeedbackStorage) Create(f *domain.Feedback) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.created = append(m.created, f)
	return nil
}

func (m *mockFeedbackStorage) GetAll() ([]domain.Feedback, error) {
	return nil, nil
}

func (m *mockFeedbackStorage) CountRecentByUser(_ string, _ time.Time) (int, error) {
	return m.recentCount, nil
}

// newTestFeedbackHandler creates a FeedbackHandler wired to a mock storage.
func newTestFeedbackHandler(store *mockFeedbackStorage) *FeedbackHandler {
	svc := services.NewFeedbackService(store)
	return NewFeedbackHandler(svc)
}

// setAuthContext adds userID and householdID to request context like middleware.RequireAuth does.
func setAuthContext(r *http.Request, userID, householdID string) *http.Request {
	ctx := middleware.WithAuthContext(r.Context(), userID, householdID)
	return r.WithContext(ctx)
}

func TestFeedbackHandler_Create_ValidBody(t *testing.T) {
	store := &mockFeedbackStorage{}
	h := newTestFeedbackHandler(store)

	body := `{"mood":"good","categories":["recipes","menu"],"comment":"Bra app!"}`
	req := httptest.NewRequest("POST", "/feedback", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1234-5678-9abc-def012345678", "hh_test-1234-5678-9abc-def012345678")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp domain.CreateFeedbackResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID == "" {
		t.Fatal("expected non-empty feedback ID")
	}
	if len(store.created) != 1 {
		t.Fatalf("expected 1 feedback stored, got %d", len(store.created))
	}
}

func TestFeedbackHandler_Create_InvalidMood(t *testing.T) {
	store := &mockFeedbackStorage{}
	h := newTestFeedbackHandler(store)

	body := `{"mood":"terrible"}`
	req := httptest.NewRequest("POST", "/feedback", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1234-5678-9abc-def012345678", "")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "invalid_mood" {
		t.Errorf("expected error 'invalid_mood', got %q", errResp["error"])
	}
}

func TestFeedbackHandler_Create_MissingMood(t *testing.T) {
	store := &mockFeedbackStorage{}
	h := newTestFeedbackHandler(store)

	// Empty mood → service validates as invalid
	body := `{"comment":"No mood provided"}`
	req := httptest.NewRequest("POST", "/feedback", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1234-5678-9abc-def012345678", "")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing mood, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestFeedbackHandler_Create_InvalidJSON(t *testing.T) {
	store := &mockFeedbackStorage{}
	h := newTestFeedbackHandler(store)

	body := `not json at all`
	req := httptest.NewRequest("POST", "/feedback", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1234-5678-9abc-def012345678", "")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", rr.Code)
	}
}

func TestFeedbackHandler_Create_RateLimited(t *testing.T) {
	store := &mockFeedbackStorage{recentCount: 5}
	h := newTestFeedbackHandler(store)

	body := `{"mood":"bad"}`
	req := httptest.NewRequest("POST", "/feedback", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1234-5678-9abc-def012345678", "")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "feedback_rate_limited" {
		t.Errorf("expected error 'feedback_rate_limited', got %q", errResp["error"])
	}
}

func TestFeedbackHandler_Create_CommentTooLong(t *testing.T) {
	store := &mockFeedbackStorage{}
	h := newTestFeedbackHandler(store)

	longComment := make([]byte, 501)
	for i := range longComment {
		longComment[i] = 'a'
	}

	payload := map[string]interface{}{
		"mood":    "good",
		"comment": string(longComment),
	}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/feedback", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1234-5678-9abc-def012345678", "")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "comment_too_long" {
		t.Errorf("expected error 'comment_too_long', got %q", errResp["error"])
	}
}

func TestFeedbackHandler_Create_InvalidCategory(t *testing.T) {
	store := &mockFeedbackStorage{}
	h := newTestFeedbackHandler(store)

	body := `{"mood":"okay","categories":["recipes","nonexistent"]}`
	req := httptest.NewRequest("POST", "/feedback", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1234-5678-9abc-def012345678", "")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "invalid_category" {
		t.Errorf("expected error 'invalid_category', got %q", errResp["error"])
	}
}

func TestFeedbackHandler_Create_MoodOnly(t *testing.T) {
	store := &mockFeedbackStorage{}
	h := newTestFeedbackHandler(store)

	body := `{"mood":"bad"}`
	req := httptest.NewRequest("POST", "/feedback", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1234-5678-9abc-def012345678", "")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestFeedbackHandler_Create_StoresCorrectUserAndHousehold(t *testing.T) {
	store := &mockFeedbackStorage{}
	h := newTestFeedbackHandler(store)

	body := `{"mood":"good","comment":"test"}`
	req := httptest.NewRequest("POST", "/feedback", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-aaaa-bbbb-cccc-dddddddddddd", "hh_test-eeee-ffff-0000-111111111111")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	if len(store.created) != 1 {
		t.Fatalf("expected 1 feedback, got %d", len(store.created))
	}
	fb := store.created[0]
	if fb.UserID != "usr_test-aaaa-bbbb-cccc-dddddddddddd" {
		t.Errorf("expected userID 'usr_test-aaaa-bbbb-cccc-dddddddddddd', got %q", fb.UserID)
	}
	if fb.HouseholdID != "hh_test-eeee-ffff-0000-111111111111" {
		t.Errorf("expected householdID 'hh_test-eeee-ffff-0000-111111111111', got %q", fb.HouseholdID)
	}
}
