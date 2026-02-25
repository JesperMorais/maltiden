package services

import (
	"maltiden/internal/domain"
	"testing"
	"time"
)

// mockFeedbackStorage implements domain.FeedbackRepository for testing.
type mockFeedbackStorage struct {
	created      []*domain.Feedback
	recentCount  int
	recentErr    error
	createErr    error
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
	return m.recentCount, m.recentErr
}

func TestFeedbackService_Create_Valid(t *testing.T) {
	store := &mockFeedbackStorage{}
	svc := NewFeedbackService(store)

	resp, err := svc.Create(domain.CreateFeedbackRequest{
		Mood:       "good",
		Categories: []string{"recipes", "menu"},
		Comment:    "Bra app!",
		Page:       "/dashboard",
	}, "usr_123", "hh_456")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.ID == "" {
		t.Fatal("expected non-empty response ID")
	}
	if len(store.created) != 1 {
		t.Fatalf("expected 1 feedback created, got %d", len(store.created))
	}

	fb := store.created[0]
	if fb.Mood != "good" {
		t.Errorf("expected mood 'good', got %q", fb.Mood)
	}
	if fb.UserID != "usr_123" {
		t.Errorf("expected userID 'usr_123', got %q", fb.UserID)
	}
	if fb.HouseholdID != "hh_456" {
		t.Errorf("expected householdID 'hh_456', got %q", fb.HouseholdID)
	}
	if fb.Comment != "Bra app!" {
		t.Errorf("expected comment 'Bra app!', got %q", fb.Comment)
	}
}

func TestFeedbackService_Create_InvalidMood(t *testing.T) {
	store := &mockFeedbackStorage{}
	svc := NewFeedbackService(store)

	_, err := svc.Create(domain.CreateFeedbackRequest{
		Mood: "terrible",
	}, "usr_123", "")

	if err != domain.ErrInvalidMood {
		t.Fatalf("expected ErrInvalidMood, got %v", err)
	}
}

func TestFeedbackService_Create_CommentTooLong(t *testing.T) {
	store := &mockFeedbackStorage{}
	svc := NewFeedbackService(store)

	longComment := make([]byte, 501)
	for i := range longComment {
		longComment[i] = 'a'
	}

	_, err := svc.Create(domain.CreateFeedbackRequest{
		Mood:    "good",
		Comment: string(longComment),
	}, "usr_123", "")

	if err != domain.ErrCommentTooLong {
		t.Fatalf("expected ErrCommentTooLong, got %v", err)
	}
}

func TestFeedbackService_Create_InvalidCategory(t *testing.T) {
	store := &mockFeedbackStorage{}
	svc := NewFeedbackService(store)

	_, err := svc.Create(domain.CreateFeedbackRequest{
		Mood:       "okay",
		Categories: []string{"recipes", "nonexistent"},
	}, "usr_123", "")

	if err != domain.ErrInvalidCategory {
		t.Fatalf("expected ErrInvalidCategory, got %v", err)
	}
}

func TestFeedbackService_Create_RateLimited(t *testing.T) {
	store := &mockFeedbackStorage{recentCount: 5}
	svc := NewFeedbackService(store)

	_, err := svc.Create(domain.CreateFeedbackRequest{
		Mood: "bad",
	}, "usr_123", "")

	if err != domain.ErrFeedbackRateLimited {
		t.Fatalf("expected ErrFeedbackRateLimited, got %v", err)
	}
}

func TestFeedbackService_Create_MoodOnly(t *testing.T) {
	store := &mockFeedbackStorage{}
	svc := NewFeedbackService(store)

	resp, err := svc.Create(domain.CreateFeedbackRequest{
		Mood: "bad",
	}, "usr_123", "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.ID == "" {
		t.Fatal("expected non-empty response ID")
	}
}
