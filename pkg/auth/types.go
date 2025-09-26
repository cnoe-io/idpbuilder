package auth

import "errors"

// Flag name constants for consistency across the application
const (
	UsernameFlagName = "username"
	PasswordFlagName = "password"
)

// Credentials holds registry authentication credentials
type Credentials struct {
	Username string
	Password string
}

// AuthConfig holds authentication configuration and metadata
type AuthConfig struct {
	Credentials Credentials
	Required    bool
}

// AuthValidator defines the interface for credential validation
type AuthValidator interface {
	ValidateCredentials(creds Credentials) error
	IsAuthRequired(creds Credentials) bool
}

// Common authentication errors
var (
	ErrEmptyUsername = errors.New("username cannot be empty when password is provided")
	ErrEmptyPassword = errors.New("password cannot be empty when username is provided")
	ErrInvalidUsernameFormat = errors.New("username contains invalid characters")
	ErrUsernameTooLong = errors.New("username exceeds maximum length")
	ErrPasswordTooLong = errors.New("password exceeds maximum length")
)

// Validation constants
const (
	MaxUsernameLength = 256
	MaxPasswordLength = 1024
)

// NewCredentials creates a new Credentials instance with the provided username and password
func NewCredentials(username, password string) Credentials {
	return Credentials{
		Username: username,
		Password: password,
	}
}

// NewAuthConfig creates a new AuthConfig instance with the provided credentials
func NewAuthConfig(creds Credentials) AuthConfig {
	return AuthConfig{
		Credentials: creds,
		Required:    creds.Username != "" || creds.Password != "",
	}
}

// IsEmpty returns true if both username and password are empty
func (c Credentials) IsEmpty() bool {
	return c.Username == "" && c.Password == ""
}

// IsComplete returns true if both username and password are provided
func (c Credentials) IsComplete() bool {
	return c.Username != "" && c.Password != ""
}