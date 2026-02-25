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

func (s *FeedbackStorage) GetAll() ([]domain.Feedback, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, household_id, mood, categories, comment, page, viewport_width, user_agent, created_at
		FROM feedback
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbacks []domain.Feedback
	for rows.Next() {
		var f domain.Feedback
		var householdID sql.NullString
		var categoriesJSON sql.NullString
		var comment sql.NullString
		var page sql.NullString
		var viewportWidth sql.NullInt64
		var userAgent sql.NullString

		err := rows.Scan(
			&f.ID,
			&f.UserID,
			&householdID,
			&f.Mood,
			&categoriesJSON,
			&comment,
			&page,
			&viewportWidth,
			&userAgent,
			&f.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if householdID.Valid {
			f.HouseholdID = householdID.String
		}
		if categoriesJSON.Valid {
			json.Unmarshal([]byte(categoriesJSON.String), &f.Categories)
		}
		if comment.Valid {
			f.Comment = comment.String
		}
		if page.Valid {
			f.Page = page.String
		}
		if viewportWidth.Valid {
			f.ViewportWidth = int(viewportWidth.Int64)
		}
		if userAgent.Valid {
			f.UserAgent = userAgent.String
		}

		feedbacks = append(feedbacks, f)
	}

	return feedbacks, rows.Err()
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
