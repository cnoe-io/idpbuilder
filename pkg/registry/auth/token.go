// Package auth provides authentication implementations for OCI registry operations.
// It supports multiple authentication methods including basic auth, bearer tokens,
// and anonymous access.
package auth

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

// TokenAuthenticator implements bearer token authentication
type TokenAuthenticator struct {
	mu          sync.RWMutex
	token       string
	expiresAt   time.Time
	authConfig  *types.AuthConfig
	tokenClient TokenClient
}

// TokenClient interface for token operations
type TokenClient interface {
	RequestToken(ctx context.Context, config *types.AuthConfig) (*types.TokenResponse, error)
}

// NewTokenAuthenticator creates a new token authenticator
func NewTokenAuthenticator(config *types.AuthConfig, client TokenClient) (*TokenAuthenticator, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if config.Token == "" && client == nil {
		return nil, fmt.Errorf("token or token client required")
	}

	auth := &TokenAuthenticator{
		authConfig:  config,
		tokenClient: client,
	}

	if config.Token != "" {
		auth.token = config.Token
		// Set a default expiry if token is provided directly
		auth.expiresAt = time.Now().Add(1 * time.Hour)
	}

	return auth, nil
}

// Authenticate implements Authenticator interface
func (t *TokenAuthenticator) Authenticate(ctx context.Context, req *http.Request) error {
	t.mu.RLock()
	token := t.token
	t.mu.RUnlock()

	if token == "" {
		if err := t.Refresh(ctx); err != nil {
			return fmt.Errorf("failed to obtain token: %w", err)
		}
		t.mu.RLock()
		token = t.token
		t.mu.RUnlock()
	}

	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

// Refresh implements Authenticator interface
func (t *TokenAuthenticator) Refresh(ctx context.Context) error {
	if t.tokenClient == nil {
		return fmt.Errorf("no token client configured for refresh")
	}

	resp, err := t.tokenClient.RequestToken(ctx, t.authConfig)
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}

	if resp == nil {
		return fmt.Errorf("token response is nil")
	}

	t.mu.Lock()
	t.token = resp.Token
	if resp.ExpiresIn > 0 {
		t.expiresAt = time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
	} else {
		t.expiresAt = time.Now().Add(1 * time.Hour)
	}
	t.mu.Unlock()

	return nil
}

// IsValid implements Authenticator interface
func (t *TokenAuthenticator) IsValid() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.token == "" {
		return false
	}

	// Check if token is expired with 30-second buffer
	return time.Now().Add(30 * time.Second).Before(t.expiresAt)
}