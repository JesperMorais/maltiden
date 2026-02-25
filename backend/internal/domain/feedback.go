package domain

import "time"

type Feedback struct {
	ID            string    `json:"id"`
	UserID        string    `json:"userId"`
	HouseholdID   string    `json:"householdId,omitempty"`
	Mood          string    `json:"mood"`
	Categories    []string  `json:"categories,omitempty"`
	Comment       string    `json:"comment,omitempty"`
	Page          string    `json:"page,omitempty"`
	ViewportWidth int       `json:"viewportWidth,omitempty"`
	UserAgent     string    `json:"-"`
	CreatedAt     time.Time `json:"createdAt"`
}

type CreateFeedbackRequest struct {
	Mood          string   `json:"mood"`
	Categories    []string `json:"categories,omitempty"`
	Comment       string   `json:"comment,omitempty"`
	Page          string   `json:"page,omitempty"`
	ViewportWidth int      `json:"viewportWidth,omitempty"`
	UserAgent     string   `json:"-"`
}

type CreateFeedbackResponse struct {
	ID string `json:"id"`
}

type FeedbackRepository interface {
	Create(feedback *Feedback) error
	GetAll() ([]Feedback, error)
	CountRecentByUser(userID string, since time.Time) (int, error)
}
