package services

import (
	"maltiden/internal/domain"
	"unicode"
)

// ValidatePasswordStrength enforces the project-wide password policy:
// minimum 8 characters and at least 3 of 4 character categories
// (uppercase letter, lowercase letter, digit, special character).
//
// Returns domain.ErrWeakPassword if the password fails either check.
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return domain.ErrWeakPassword
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}
	types := 0
	for _, has := range []bool{hasUpper, hasLower, hasDigit, hasSpecial} {
		if has {
			types++
		}
	}
	if types < 3 {
		return domain.ErrWeakPassword
	}
	return nil
}
