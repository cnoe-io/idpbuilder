# Registry Authentication Implementation Plan

## 📌 Effort Overview

**Effort ID**: E1.1.2B
**Effort Name**: registry-auth
**Phase**: 1, Wave: 1
**Branch**: `idpbuilder-oci-build-push/phase1/wave1/registry-auth`
**Base Branch**: `idpbuilder-oci-build-push/phase1/wave1/registry-types`
**Estimated Lines**: 350
**Can Parallelize**: No (depends on registry-types)
**Parallel With**: None
**Dependencies**: [registry-types (E1.1.2A)]
**Created**: 2025-09-18

## 🎯 Mission

Implement authentication handlers and middleware for OCI registry operations. This effort provides the authentication layer that sits between the registry types and actual registry operations, handling credential management, token negotiation, and authentication flow.

## 📦 Scope and Boundaries

### What This Effort Includes
- Authentication handler implementations
- Token management and refresh logic
- Credential validation
- Authentication middleware for registry operations
- Basic auth and token auth implementations

### What This Effort Excludes
- Core registry types (in registry-types)
- Helper utilities and convenience functions (in registry-helpers)
- Test implementations (in registry-tests)
- Actual registry client operations

## 📁 File Structure

```
pkg/
└── registry/
    └── auth/
        ├── authenticator.go      (~80 lines) - Core authenticator interface and factory
        ├── basic.go              (~60 lines) - Basic auth implementation
        ├── token.go              (~90 lines) - Token auth implementation
        ├── middleware.go         (~70 lines) - Auth middleware for HTTP clients
        └── manager.go            (~50 lines) - Auth manager for credential lifecycle
```

## 🔧 Implementation Details

### 1. Core Authenticator Interface (`pkg/registry/auth/authenticator.go`)

```go
package auth

import (
    "context"
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
    switch config.AuthType {
    case types.AuthTypeBasic:
        return NewBasicAuthenticator(config)
    case types.AuthTypeToken:
        return NewTokenAuthenticator(config)
    case types.AuthTypeNone:
        return NewNoOpAuthenticator(), nil
    default:
        return nil, fmt.Errorf("unsupported auth type: %s", config.AuthType)
    }
}

// NoOpAuthenticator for registries without auth
type NoOpAuthenticator struct{}

func NewNoOpAuthenticator() *NoOpAuthenticator {
    return &NoOpAuthenticator{}
}

func (n *NoOpAuthenticator) Authenticate(ctx context.Context, req *http.Request) error {
    return nil
}

func (n *NoOpAuthenticator) Refresh(ctx context.Context) error {
    return nil
}

func (n *NoOpAuthenticator) IsValid() bool {
    return true
}
```

### 2. Basic Authentication (`pkg/registry/auth/basic.go`)

```go
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
    if config.Username == "" || config.Password == "" {
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

func (b *BasicAuthenticator) Authenticate(ctx context.Context, req *http.Request) error {
    req.Header.Set("Authorization", "Basic "+b.encoded)
    return nil
}

func (b *BasicAuthenticator) Refresh(ctx context.Context) error {
    // Basic auth doesn't need refresh
    return nil
}

func (b *BasicAuthenticator) IsValid() bool {
    return b.encoded != ""
}
```

### 3. Token Authentication (`pkg/registry/auth/token.go`)

```go
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

func (t *TokenAuthenticator) Refresh(ctx context.Context) error {
    if t.tokenClient == nil {
        return fmt.Errorf("no token client configured for refresh")
    }

    resp, err := t.tokenClient.RequestToken(ctx, t.authConfig)
    if err != nil {
        return fmt.Errorf("token request failed: %w", err)
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

func (t *TokenAuthenticator) IsValid() bool {
    t.mu.RLock()
    defer t.mu.RUnlock()

    if t.token == "" {
        return false
    }

    // Check if token is expired with 30-second buffer
    return time.Now().Add(30 * time.Second).Before(t.expiresAt)
}
```

### 4. Authentication Middleware (`pkg/registry/auth/middleware.go`)

```go
package auth

import (
    "context"
    "net/http"
    "github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

// Transport wraps an http.RoundTripper with authentication
type Transport struct {
    Base          http.RoundTripper
    Authenticator Authenticator
}

// NewTransport creates a new authenticated transport
func NewTransport(base http.RoundTripper, auth Authenticator) *Transport {
    if base == nil {
        base = http.DefaultTransport
    }

    return &Transport{
        Base:          base,
        Authenticator: auth,
    }
}

// RoundTrip implements http.RoundTripper
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
    // Clone the request to avoid modifying the original
    clonedReq := req.Clone(req.Context())

    // Apply authentication
    if t.Authenticator != nil {
        // Check if auth needs refresh
        if !t.Authenticator.IsValid() {
            if err := t.Authenticator.Refresh(req.Context()); err != nil {
                return nil, fmt.Errorf("auth refresh failed: %w", err)
            }
        }

        if err := t.Authenticator.Authenticate(req.Context(), clonedReq); err != nil {
            return nil, fmt.Errorf("authentication failed: %w", err)
        }
    }

    // Execute the request
    resp, err := t.Base.RoundTrip(clonedReq)
    if err != nil {
        return nil, err
    }

    // Handle 401 Unauthorized by refreshing auth and retrying once
    if resp.StatusCode == http.StatusUnauthorized && t.Authenticator != nil {
        resp.Body.Close()

        if err := t.Authenticator.Refresh(req.Context()); err != nil {
            return nil, fmt.Errorf("auth refresh after 401 failed: %w", err)
        }

        // Retry with refreshed auth
        retryReq := req.Clone(req.Context())
        if err := t.Authenticator.Authenticate(req.Context(), retryReq); err != nil {
            return nil, fmt.Errorf("re-authentication failed: %w", err)
        }

        return t.Base.RoundTrip(retryReq)
    }

    return resp, nil
}
```

### 5. Auth Manager (`pkg/registry/auth/manager.go`)

```go
package auth

import (
    "context"
    "fmt"
    "sync"
    "github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

// Manager manages authentication for multiple registries
type Manager struct {
    mu      sync.RWMutex
    auths   map[string]Authenticator
    store   types.CredentialStore
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
    m.mu.RLock()
    auth, exists := m.auths[registry]
    m.mu.RUnlock()

    if exists && auth.IsValid() {
        return auth, nil
    }

    // Load credentials from store
    creds, err := m.store.GetCredentials(registry)
    if err != nil {
        return nil, fmt.Errorf("failed to get credentials: %w", err)
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
```

## 📊 Size Estimation Breakdown

| File | Estimated Lines | Purpose |
|------|-----------------|---------|
| authenticator.go | 80 | Core interface and factory |
| basic.go | 60 | Basic auth implementation |
| token.go | 90 | Token auth with refresh |
| middleware.go | 70 | HTTP transport wrapper |
| manager.go | 50 | Multi-registry auth management |
| **TOTAL** | **350** | Within limit |

## 🔗 Dependencies and Integration

### Internal Dependencies
- `github.com/cnoe-io/idpbuilder/pkg/registry/types` - From registry-types effort (E1.1.2A)
  - Uses: `AuthConfig`, `AuthType`, `TokenResponse`, `CredentialStore`

### External Dependencies
- Standard library only (`net/http`, `context`, `sync`, `time`, `encoding/base64`)

### Integration Points
1. **Registry Types**: Import and use types from registry-types package
2. **Registry Client**: Will be used by registry client (future effort) via middleware
3. **Registry Helpers**: Helpers will use these authenticators for operations

## ⚡ Implementation Strategy

### Phase 1: Core Structure (50 lines)
1. Create package structure
2. Define Authenticator interface
3. Implement NoOpAuthenticator

### Phase 2: Basic Auth (60 lines)
1. Implement BasicAuthenticator
2. Add header generation
3. Test with mock requests

### Phase 3: Token Auth (90 lines)
1. Define TokenClient interface
2. Implement TokenAuthenticator
3. Add refresh logic with expiry

### Phase 4: Middleware (70 lines)
1. Create Transport wrapper
2. Add authentication injection
3. Implement 401 retry logic

### Phase 5: Manager (50 lines)
1. Implement auth caching
2. Add credential store integration
3. Handle multi-registry scenarios

### Phase 6: Integration (30 lines)
1. Wire components together
2. Add error handling
3. Final testing and validation

## 🧪 Testing Strategy

Testing will be handled in the separate registry-tests effort (E1.1.2D), but key test scenarios include:

1. **Unit Tests**:
   - Each authenticator type
   - Token refresh logic
   - Middleware behavior
   - Manager caching

2. **Integration Tests**:
   - Full auth flow with mock registry
   - Credential store integration
   - Multi-registry scenarios

## ✅ Success Criteria

1. **Functional Requirements**:
   - ✅ Basic authentication works
   - ✅ Token authentication with refresh
   - ✅ Middleware properly injects auth
   - ✅ Manager handles multiple registries

2. **Non-Functional Requirements**:
   - ✅ Thread-safe operations
   - ✅ Proper error handling
   - ✅ Clean separation from other efforts
   - ✅ Under 350 lines limit

## 🚀 Next Steps

After this effort is complete:
1. **registry-helpers (E1.1.2C)** will build convenience functions using these authenticators
2. **registry-tests (E1.1.2D)** will provide comprehensive test coverage
3. Future efforts will use this auth layer for actual registry operations

## 📝 Notes

- This effort focuses ONLY on authentication logic
- No actual registry operations are implemented here
- Clean interfaces allow for easy extension (OAuth2, etc.)
- Thread safety is critical for concurrent operations
- Error messages should be clear and actionable