// Package push implements OCI registry push functionality for idpbuilder.
package push

import (
	"fmt"
	"os"
)

// Credentials holds resolved authentication credentials for registry operations.
// Either Username/Password pair OR Token is used, never both.
// The struct intentionally has no String() method to prevent accidental
// credential logging (security requirement P1.3).
type Credentials struct {
	// Username for basic authentication
	Username string
	// Password for basic authentication
	Password string
	// Token for bearer token authentication (takes precedence over basic auth)
	Token string
	// IsAnonymous indicates no credentials were provided
	IsAnonymous bool
}

// CredentialFlags contains CLI flag values for credential resolution.
// These values take precedence over environment variables per REQ-014.
type CredentialFlags struct {
	Username string
	Password string
	Token    string
}

// EnvironmentLookup abstracts environment variable access for testing.
// This allows tests to inject mock environment values without modifying os.Environ.
type EnvironmentLookup interface {
	// Get retrieves the value of an environment variable.
	// Returns empty string if not set.
	Get(key string) string
}

// CredentialResolver resolves authentication credentials from multiple sources.
// Resolution priority: CLI flags > environment variables > anonymous access.
type CredentialResolver interface {
	// Resolve determines credentials based on flags and environment.
	// Returns Credentials with IsAnonymous=true if no credentials found.
	// Returns error if both token and username/password are provided.
	Resolve(flags CredentialFlags, env EnvironmentLookup) (*Credentials, error)
}

// Environment variable names for credential resolution
const (
	EnvRegistryUsername = "IDPBUILDER_REGISTRY_USERNAME"
	EnvRegistryPassword = "IDPBUILDER_REGISTRY_PASSWORD"
	EnvRegistryToken    = "IDPBUILDER_REGISTRY_TOKEN"
)

// DefaultEnvironment implements EnvironmentLookup using os.Getenv.
type DefaultEnvironment struct{}

// Get implements EnvironmentLookup.Get using os.Getenv.
func (e *DefaultEnvironment) Get(key string) string {
	return os.Getenv(key)
}

// DefaultCredentialResolver implements CredentialResolver.
// Priority order: flags > environment > anonymous
type DefaultCredentialResolver struct{}

// Resolve implements CredentialResolver.Resolve.
// Validates that either basic auth (username+password) or token is provided, not both.
// Resolution follows REQ-014 precedence: CLI flags override environment variables.
func (r *DefaultCredentialResolver) Resolve(flags CredentialFlags, env EnvironmentLookup) (*Credentials, error) {
	creds := &Credentials{}

	// Token resolution: flag takes precedence over environment (REQ-014)
	creds.Token = flags.Token
	if creds.Token == "" {
		creds.Token = env.Get(EnvRegistryToken)
	}

	// Username resolution: flag takes precedence over environment (REQ-014)
	creds.Username = flags.Username
	if creds.Username == "" {
		creds.Username = env.Get(EnvRegistryUsername)
	}

	// Password resolution: flag takes precedence over environment (REQ-014)
	creds.Password = flags.Password
	if creds.Password == "" {
		creds.Password = env.Get(EnvRegistryPassword)
	}

	// Determine auth mode
	hasToken := creds.Token != ""
	hasBasic := creds.Username != "" || creds.Password != ""

	// Validate: cannot have both token and basic auth
	if hasToken && hasBasic {
		return nil, fmt.Errorf("cannot specify both token and username/password credentials")
	}

	// If token is provided, clear basic auth fields for consistency
	if hasToken {
		creds.Username = ""
		creds.Password = ""
	}

	// If no credentials at all, mark as anonymous
	creds.IsAnonymous = !hasToken && !hasBasic

	return creds, nil
}
