package domain

import "errors"

// Sentinel errors for business logic. Services return these directly (not wrapped)
// so that handlers can use errors.Is() for type-safe error checking.

var (
	// Authentication & authorization
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidCredentials = errors.New("invalid_credentials")
	ErrWeakPassword       = errors.New("weak_password")

	// User
	ErrDuplicateEmail = errors.New("email_already_exists")

	// Household
	ErrInvalidCode  = errors.New("invalid_code")
	ErrAlreadyMember = errors.New("already_member")
	ErrCannotRemove = errors.New("cannot_remove")
	ErrCodeRequired = errors.New("code_required")

	// Resource
	ErrNotFound = errors.New("not_found")

	// Resource
	ErrInvalidDays = errors.New("invalid_days")

	// Recipe validation
	ErrNameRequired         = errors.New("name_required")
	ErrInvalidServings      = errors.New("invalid_servings")
	ErrIngredientsRequired  = errors.New("ingredients_required")
	ErrInstructionsRequired = errors.New("instructions_required")
	ErrNameTooLong          = errors.New("name_too_long")
	ErrTooManyIngredients   = errors.New("too_many_ingredients")
	ErrTooManyTags          = errors.New("too_many_tags")
	ErrTagTooLong           = errors.New("tag_too_long")

	// Auth validation
	ErrInvalidEmail = errors.New("invalid_email")

	// Feedback validation
	ErrInvalidMood        = errors.New("invalid_mood")
	ErrCommentTooLong     = errors.New("comment_too_long")
	ErrInvalidCategory    = errors.New("invalid_category")
	ErrFeedbackRateLimited = errors.New("feedback_rate_limited")
)
