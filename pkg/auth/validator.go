package auth

import (
	"regexp"
	"strings"
)

// DefaultValidator implements the AuthValidator interface with standard validation rules
type DefaultValidator struct{}

// ValidateCredentials validates the provided credentials according to authentication rules
func (v *DefaultValidator) ValidateCredentials(creds Credentials) error {
	// If both are empty, authentication is not required - this is valid
	if creds.IsEmpty() {
		return nil
	}

	// If only one is provided, this is invalid
	if creds.Username == "" && creds.Password != "" {
		return ErrEmptyUsername
	}

	if creds.Username != "" && creds.Password == "" {
		return ErrEmptyPassword
	}

	// Validate username format - should not contain invalid characters that could break URLs
	if err := v.validateUsernameFormat(creds.Username); err != nil {
		return err
	}

	// Validate length constraints
	if len(creds.Username) > MaxUsernameLength {
		return ErrUsernameTooLong
	}

	if len(creds.Password) > MaxPasswordLength {
		return ErrPasswordTooLong
	}

	return nil
}

// IsAuthRequired returns true if authentication is required based on the provided credentials
func (v *DefaultValidator) IsAuthRequired(creds Credentials) bool {
	return !creds.IsEmpty()
}

// validateUsernameFormat checks if the username contains only valid characters
func (v *DefaultValidator) validateUsernameFormat(username string) error {
	// Username should not contain characters that could break URL parsing or HTTP headers
	// Allow alphanumeric, hyphens, underscores, dots, and @ symbol for email-style usernames
	validUsernameRegex := regexp.MustCompile(`^[a-zA-Z0-9._@-]+$`)

	if !validUsernameRegex.MatchString(username) {
		return ErrInvalidUsernameFormat
	}

	// Additional checks for problematic characters
	if strings.Contains(username, " ") ||
	   strings.Contains(username, "\t") ||
	   strings.Contains(username, "\n") ||
	   strings.Contains(username, ":") ||
	   strings.Contains(username, "/") {
		return ErrInvalidUsernameFormat
	}

	return nil
}