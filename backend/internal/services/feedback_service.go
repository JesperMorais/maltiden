package services

import (
	"maltiden/internal/domain"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type FeedbackService struct {
	feedbackStorage domain.FeedbackRepository
}

func NewFeedbackService(feedbackStorage domain.FeedbackRepository) *FeedbackService {
	return &FeedbackService{feedbackStorage: feedbackStorage}
}

func (s *FeedbackService) Create(req domain.CreateFeedbackRequest, userID, householdID string) (*domain.CreateFeedbackResponse, error) {
	// Validate mood
	if req.Mood != "good" && req.Mood != "okay" && req.Mood != "bad" {
		return nil, domain.ErrInvalidMood
	}

	// Validate comment length
	if utf8.RuneCountInString(req.Comment) > 500 {
		return nil, domain.ErrCommentTooLong
	}

	// Validate categories
	validCategories := map[string]bool{
		"recipes": true, "menu": true, "shopping": true, "design": true, "other": true,
	}
	for _, cat := range req.Categories {
		if !validCategories[cat] {
			return nil, domain.ErrInvalidCategory
		}
	}

	// Rate limit: max 5 per user per hour
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	count, err := s.feedbackStorage.CountRecentByUser(userID, oneHourAgo)
	if err != nil {
		return nil, err
	}
	if count >= 5 {
		return nil, domain.ErrFeedbackRateLimited
	}

	feedback := &domain.Feedback{
		ID:            "fb_" + uuid.New().String(),
		UserID:        userID,
		HouseholdID:   householdID,
		Mood:          req.Mood,
		Categories:    req.Categories,
		Comment:       req.Comment,
		Page:          req.Page,
		ViewportWidth: req.ViewportWidth,
		UserAgent:     req.UserAgent,
		CreatedAt:     time.Now(),
	}

	if err := s.feedbackStorage.Create(feedback); err != nil {
		return nil, err
	}

	return &domain.CreateFeedbackResponse{ID: feedback.ID}, nil
}
