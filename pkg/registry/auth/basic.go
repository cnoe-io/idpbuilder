// Package auth provides authentication implementations for OCI registry operations.
// It supports multiple authentication methods including basic auth, bearer tokens,
// and anonymous access.
package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

// BasicAuthenticator implements basic authentication
type BasicAuthenticator struct {
	username string
	password string
	encoded  string
}

// NewBasicAuthenticator creates a new basic authenticator
func NewBasicAuthenticator(config *types.AuthConfig) (*BasicAuthenticator, error) {
	if config == nil || config.Username == "" || config.Password == "" {
		return nil, fmt.Errorf("username and password required for basic auth")
	}

	auth := &BasicAuthenticator{
		username: config.Username,
		password: config.Password,
	}
	auth.encoded = base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", config.Username, config.Password)),
	)

	return auth, nil
}

// Authenticate implements Authenticator interface
func (b *BasicAuthenticator) Authenticate(ctx context.Context, req *http.Request) error {
	req.Header.Set("Authorization", "Basic "+b.encoded)
	return nil
}

// Refresh implements Authenticator interface
func (b *BasicAuthenticator) Refresh(ctx context.Context) error {
	// Basic auth doesn't need refresh
	return nil
}

// IsValid implements Authenticator interface
func (b *BasicAuthenticator) IsValid() bool {
	return b.encoded != ""
}