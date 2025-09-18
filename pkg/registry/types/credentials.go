package types

import "time"

// AuthConfig represents authentication configuration for a registry
type AuthConfig struct {
	Username string
	Password string
	Token    string
	AuthType AuthType
}

// AuthType represents the type of authentication
type AuthType string

const (
	AuthTypeBasic  AuthType = "basic"
	AuthTypeToken  AuthType = "token"
	AuthTypeOAuth2 AuthType = "oauth2"
	AuthTypeNone   AuthType = "none"
)

// TokenResponse represents a token response from a registry
type TokenResponse struct {
	Token     string
	ExpiresIn int
	IssuedAt  time.Time
}

// CredentialStore interface for credential storage
type CredentialStore interface {
	GetCredentials(registry string) (*AuthConfig, error)
	StoreCredentials(registry string, creds *AuthConfig) error
	RemoveCredentials(registry string) error
}