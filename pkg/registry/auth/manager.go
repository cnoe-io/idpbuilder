// Package auth provides authentication implementations for OCI registry operations.
// It supports multiple authentication methods including basic auth, bearer tokens,
// and anonymous access.
package auth

import (
	"context"
	"fmt"
	"sync"

	"github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

// Manager manages authentication for multiple registries
type Manager struct {
	mu    sync.RWMutex
	auths map[string]Authenticator
	store types.CredentialStore
}

// NewManager creates a new auth manager
func NewManager(store types.CredentialStore) *Manager {
	return &Manager{
		auths: make(map[string]Authenticator),
		store: store,
	}
}

// GetAuthenticator gets or creates an authenticator for a registry
func (m *Manager) GetAuthenticator(ctx context.Context, registry string) (Authenticator, error) {
	if registry == "" {
		return nil, fmt.Errorf("registry cannot be empty")
	}

	m.mu.RLock()
	auth, exists := m.auths[registry]
	m.mu.RUnlock()

	if exists && auth.IsValid() {
		return auth, nil
	}

	// Load credentials from store if available
	var creds *types.AuthConfig
	var err error
	if m.store != nil {
		creds, err = m.store.GetCredentials(registry)
		if err != nil {
			return nil, fmt.Errorf("failed to get credentials: %w", err)
		}
	} else {
		// Default to no auth if no store configured
		creds = &types.AuthConfig{AuthType: types.AuthTypeNone}
	}

	// Create new authenticator
	auth, err = NewAuthenticator(creds)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator: %w", err)
	}

	// Cache the authenticator
	m.mu.Lock()
	m.auths[registry] = auth
	m.mu.Unlock()

	return auth, nil
}

// Clear removes cached authenticator for a registry
func (m *Manager) Clear(registry string) {
	m.mu.Lock()
	delete(m.auths, registry)
	m.mu.Unlock()
}

// ClearAll removes all cached authenticators
func (m *Manager) ClearAll() {
	m.mu.Lock()
	m.auths = make(map[string]Authenticator)
	m.mu.Unlock()
}