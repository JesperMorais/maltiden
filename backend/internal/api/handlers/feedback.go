package handlers

import (
	"errors"
	"log/slog"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"maltiden/pkg/middleware"
	"net/http"
)

type FeedbackHandler struct {
	feedbackService *services.FeedbackService
}

func NewFeedbackHandler(feedbackService *services.FeedbackService) *FeedbackHandler {
	return &FeedbackHandler{feedbackService: feedbackService}
}

func (h *FeedbackHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	householdID := middleware.GetHouseholdID(r)

	var req domain.CreateFeedbackRequest
	if !DecodeJSON(w, r, maxBodySize, &req) {
		return
	}

	// Override client-supplied UserAgent with actual HTTP header
	req.UserAgent = r.Header.Get("User-Agent")

	resp, err := h.feedbackService.Create(req, userID, householdID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidMood):
			WriteError(w, http.StatusBadRequest, "invalid_mood")
		case errors.Is(err, domain.ErrCommentTooLong):
			WriteError(w, http.StatusBadRequest, "comment_too_long")
		case errors.Is(err, domain.ErrInvalidCategory):
			WriteError(w, http.StatusBadRequest, "invalid_category")
		case errors.Is(err, domain.ErrFeedbackRateLimited):
			WriteError(w, http.StatusTooManyRequests, "feedback_rate_limited")
		default:
			slog.Error("CreateFeedback failed", "error", err)
			WriteError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}

	WriteJSON(w, http.StatusCreated, resp)
}
