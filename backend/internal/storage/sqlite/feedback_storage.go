package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"maltiden/internal/domain"
	"time"
)

type FeedbackStorage struct {
	db *sql.DB
}

func NewFeedbackStorage(db *sql.DB) *FeedbackStorage {
	return &FeedbackStorage{db: db}
}

func (s *FeedbackStorage) Create(feedback *domain.Feedback) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var categoriesJSON *string
	if len(feedback.Categories) > 0 {
		b, err := json.Marshal(feedback.Categories)
		if err != nil {
			return err
		}
		str := string(b)
		categoriesJSON = &str
	}

	query := `
		INSERT INTO feedback (id, user_id, household_id, mood, categories, comment, page, viewport_width, user_agent, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var householdID *string
	if feedback.HouseholdID != "" {
		householdID = &feedback.HouseholdID
	}

	var comment *string
	if feedback.Comment != "" {
		comment = &feedback.Comment
	}

	var page *string
	if feedback.Page != "" {
		page = &feedback.Page
	}

	var viewportWidth *int
	if feedback.ViewportWidth > 0 {
		viewportWidth = &feedback.ViewportWidth
	}

	var userAgent *string
	if feedback.UserAgent != "" {
		userAgent = &feedback.UserAgent
	}

	_, err := s.db.ExecContext(ctx, query,
		feedback.ID,
		feedback.UserID,
		householdID,
		feedback.Mood,
		categoriesJSON,
		comment,
		page,
		viewportWidth,
		userAgent,
		feedback.CreatedAt,
	)
	return err
}

func (s *FeedbackStorage) CountRecentByUser(userID string, since time.Time) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int
	err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM feedback WHERE user_id = ? AND created_at >= ?",
		userID, since,
	).Scan(&count)
	return count, err
}
