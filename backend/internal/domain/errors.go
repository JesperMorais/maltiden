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

	// Recipe validation
	ErrNameRequired         = errors.New("name_required")
	ErrInvalidServings      = errors.New("invalid_servings")
	ErrIngredientsRequired  = errors.New("ingredients_required")
	ErrInstructionsRequired = errors.New("instructions_required")
)
