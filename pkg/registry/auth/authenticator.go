// Package auth provides authentication implementations for OCI registry operations.
// It supports multiple authentication methods including basic auth, bearer tokens,
// and anonymous access.
package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

// Authenticator defines the interface for registry authentication
type Authenticator interface {
	// Authenticate adds authentication to the request
	Authenticate(ctx context.Context, req *http.Request) error

	// Refresh refreshes authentication credentials if needed
	Refresh(ctx context.Context) error

	// IsValid checks if current auth is still valid
	IsValid() bool
}

// NewAuthenticator creates an authenticator based on config
func NewAuthenticator(config *types.AuthConfig) (Authenticator, error) {
	if config == nil {
		return NewNoOpAuthenticator(), nil
	}

	switch config.AuthType {
	case types.AuthTypeBasic:
		return NewBasicAuthenticator(config)
	case types.AuthTypeToken:
		return NewTokenAuthenticator(config, nil)
	case types.AuthTypeNone:
		return NewNoOpAuthenticator(), nil
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", config.AuthType)
	}
}

// NoOpAuthenticator for registries without auth
type NoOpAuthenticator struct{}

// NewNoOpAuthenticator creates a new no-op authenticator
func NewNoOpAuthenticator() *NoOpAuthenticator {
	return &NoOpAuthenticator{}
}

// Authenticate implements Authenticator interface
func (n *NoOpAuthenticator) Authenticate(ctx context.Context, req *http.Request) error {
	return nil
}

// Refresh implements Authenticator interface
func (n *NoOpAuthenticator) Refresh(ctx context.Context) error {
	return nil
}

// IsValid implements Authenticator interface
func (n *NoOpAuthenticator) IsValid() bool {
	return true
}