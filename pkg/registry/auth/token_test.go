package auth

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

// mockTokenClient implements TokenClient for testing
type mockTokenClient struct {
	token     string
	expiresIn int
	err       error
	callCount int
	mu        sync.Mutex
}

func (m *mockTokenClient) RequestToken(ctx context.Context, config *types.AuthConfig) (*types.TokenResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if m.err != nil {
		return nil, m.err
	}

	return &types.TokenResponse{
		Token:     m.token,
		ExpiresIn: m.expiresIn,
		IssuedAt:  time.Now(),
	}, nil
}

func (m *mockTokenClient) getCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

func (m *mockTokenClient) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount = 0
}

// mockNilResponseTokenClient for testing nil response handling
type mockNilResponseTokenClient struct {
	mu        sync.Mutex
	callCount int
}

func (m *mockNilResponseTokenClient) RequestToken(ctx context.Context, config *types.AuthConfig) (*types.TokenResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	return nil, nil // Return nil response
}

func (m *mockNilResponseTokenClient) getCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

func TestNewTokenAuthenticator(t *testing.T) {
	tests := []struct {
		name        string
		config      *types.AuthConfig
		client      TokenClient
		expectError bool
		wantToken   string
	}{
		{
			name:        "nil config returns error",
			config:      nil,
			expectError: true,
		},
		{
			name: "direct token creates authenticator",
			config: &types.AuthConfig{
				Token: "test-token",
			},
			wantToken: "test-token",
		},
		{
			name: "token client without direct token creates authenticator",
			config: &types.AuthConfig{
				AuthType: types.AuthTypeToken,
			},
			client: &mockTokenClient{
				token:     "client-token",
				expiresIn: 3600,
			},
		},
		{
			name: "no token and no client returns error",
			config: &types.AuthConfig{
				AuthType: types.AuthTypeToken,
			},
			expectError: true,
		},
		{
			name: "both token and client works",
			config: &types.AuthConfig{
				Token: "direct-token",
			},
			client: &mockTokenClient{
				token:     "client-token",
				expiresIn: 3600,
			},
			wantToken: "direct-token", // Direct token takes precedence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewTokenAuthenticator(tt.config, tt.client)

			if tt.expectError {
				if err == nil {
					t.Errorf("NewTokenAuthenticator() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewTokenAuthenticator() unexpected error: %v", err)
				return
			}

			if auth == nil {
				t.Fatal("NewTokenAuthenticator() returned nil")
			}

			// If we expected a specific token, check it
			if tt.wantToken != "" {
				if auth.token != tt.wantToken {
					t.Errorf("NewTokenAuthenticator() token = %q, want %q", auth.token, tt.wantToken)
				}
			}

			// Verify tokenClient is set correctly
			if tt.client != nil {
				if auth.tokenClient != tt.client {
					t.Errorf("NewTokenAuthenticator() client not set correctly")
				}
			}
		})
	}
}

func TestTokenAuthenticator_Authenticate(t *testing.T) {
	tests := []struct {
		name        string
		config      *types.AuthConfig
		client      TokenClient
		expectError bool
		wantHeader  string
	}{
		{
			name: "authenticator with direct token",
			config: &types.AuthConfig{
				Token: "test-token",
			},
			wantHeader: "Bearer test-token",
		},
		{
			name: "authenticator with token client",
			config: &types.AuthConfig{
				AuthType: types.AuthTypeToken,
			},
			client: &mockTokenClient{
				token:     "client-token",
				expiresIn: 3600,
			},
			wantHeader: "Bearer client-token",
		},
		{
			name: "token client returns error",
			config: &types.AuthConfig{
				AuthType: types.AuthTypeToken,
			},
			client: &mockTokenClient{
				err: fmt.Errorf("token request failed"),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewTokenAuthenticator(tt.config, tt.client)
			if err != nil {
				t.Fatalf("NewTokenAuthenticator() error = %v", err)
			}

			req, err := http.NewRequest("GET", "http://example.com", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			err = auth.Authenticate(context.Background(), req)

			if tt.expectError {
				if err == nil {
					t.Errorf("TokenAuthenticator.Authenticate() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("TokenAuthenticator.Authenticate() error = %v", err)
				return
			}

			authHeader := req.Header.Get("Authorization")
			if authHeader != tt.wantHeader {
				t.Errorf("Authorization header = %q, want %q", authHeader, tt.wantHeader)
			}
		})
	}
}

func TestTokenAuthenticator_Refresh(t *testing.T) {
	tests := []struct {
		name        string
		client      TokenClient
		expectError bool
		wantToken   string
	}{
		{
			name: "successful token refresh",
			client: &mockTokenClient{
				token:     "refreshed-token",
				expiresIn: 7200,
			},
			wantToken: "refreshed-token",
		},
		{
			name: "token client returns error",
			client: &mockTokenClient{
				err: fmt.Errorf("refresh failed"),
			},
			expectError: true,
		},
		{
			name:        "no token client configured",
			client:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &types.AuthConfig{
				AuthType: types.AuthTypeToken,
			}

			auth, err := NewTokenAuthenticator(config, tt.client)
			if err != nil && tt.client != nil {
				t.Fatalf("NewTokenAuthenticator() error = %v", err)
			}
			if tt.client == nil {
				// Create authenticator without client for the "no client" test
				auth = &TokenAuthenticator{
					authConfig:  config,
					tokenClient: nil,
				}
			}

			err = auth.Refresh(context.Background())

			if tt.expectError {
				if err == nil {
					t.Errorf("TokenAuthenticator.Refresh() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("TokenAuthenticator.Refresh() error = %v", err)
				return
			}

			if tt.wantToken != "" {
				auth.mu.RLock()
				token := auth.token
				auth.mu.RUnlock()

				if token != tt.wantToken {
					t.Errorf("TokenAuthenticator.Refresh() token = %q, want %q", token, tt.wantToken)
				}
			}
		})
	}
}

func TestTokenAuthenticator_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		setupAuth func() *TokenAuthenticator
		expected  bool
	}{
		{
			name: "valid token not expired",
			setupAuth: func() *TokenAuthenticator {
				return &TokenAuthenticator{
					token:     "valid-token",
					expiresAt: time.Now().Add(10 * time.Minute),
				}
			},
			expected: true,
		},
		{
			name: "empty token is invalid",
			setupAuth: func() *TokenAuthenticator {
				return &TokenAuthenticator{
					token:     "",
					expiresAt: time.Now().Add(10 * time.Minute),
				}
			},
			expected: false,
		},
		{
			name: "expired token is invalid",
			setupAuth: func() *TokenAuthenticator {
				return &TokenAuthenticator{
					token:     "expired-token",
					expiresAt: time.Now().Add(-10 * time.Minute),
				}
			},
			expected: false,
		},
		{
			name: "token expiring within 30 seconds is invalid",
			setupAuth: func() *TokenAuthenticator {
				return &TokenAuthenticator{
					token:     "expiring-token",
					expiresAt: time.Now().Add(15 * time.Second),
				}
			},
			expected: false,
		},
		{
			name: "token expiring in more than 30 seconds is valid",
			setupAuth: func() *TokenAuthenticator {
				return &TokenAuthenticator{
					token:     "valid-token",
					expiresAt: time.Now().Add(60 * time.Second),
				}
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := tt.setupAuth()
			isValid := auth.IsValid()

			if isValid != tt.expected {
				t.Errorf("TokenAuthenticator.IsValid() = %v, want %v", isValid, tt.expected)
			}
		})
	}
}

func TestTokenAuthenticator_RefreshOnNilResponse(t *testing.T) {
	client := &mockNilResponseTokenClient{}

	config := &types.AuthConfig{
		AuthType: types.AuthTypeToken,
	}

	auth, err := NewTokenAuthenticator(config, client)
	if err != nil {
		t.Fatalf("NewTokenAuthenticator() error = %v", err)
	}

	err = auth.Refresh(context.Background())
	if err == nil {
		t.Errorf("TokenAuthenticator.Refresh() expected error for nil response, got nil")
	}

	expectedError := "token response is nil"
	if err.Error() != expectedError {
		t.Errorf("TokenAuthenticator.Refresh() error = %q, want %q", err.Error(), expectedError)
	}
}

func TestTokenAuthenticator_ConcurrentAccess(t *testing.T) {
	client := &mockTokenClient{
		token:     "concurrent-token",
		expiresIn: 3600,
	}

	config := &types.AuthConfig{
		AuthType: types.AuthTypeToken,
	}

	auth, err := NewTokenAuthenticator(config, client)
	if err != nil {
		t.Fatalf("NewTokenAuthenticator() error = %v", err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Test concurrent calls to various methods
	for i := 0; i < 10; i++ {
		wg.Add(3) // Three goroutines per iteration

		// Concurrent Authenticate calls
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			if err := auth.Authenticate(context.Background(), req); err != nil {
				errors <- fmt.Errorf("Authenticate error: %w", err)
			}
		}()

		// Concurrent Refresh calls
		go func() {
			defer wg.Done()
			if err := auth.Refresh(context.Background()); err != nil {
				errors <- fmt.Errorf("Refresh error: %w", err)
			}
		}()

		// Concurrent IsValid calls
		go func() {
			defer wg.Done()
			_ = auth.IsValid() // IsValid doesn't return error
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}

	// Verify that token client was called (at least some calls should succeed)
	callCount := client.getCallCount()
	if callCount == 0 {
		t.Error("Expected at least one call to token client during concurrent access")
	}
}

func TestTokenAuthenticator_AuthenticateTriggerRefresh(t *testing.T) {
	client := &mockTokenClient{
		token:     "fresh-token",
		expiresIn: 3600,
	}

	config := &types.AuthConfig{
		AuthType: types.AuthTypeToken,
	}

	// Create authenticator without initial token
	auth := &TokenAuthenticator{
		authConfig:  config,
		tokenClient: client,
		token:       "", // No initial token
	}

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client.reset()

	err = auth.Authenticate(context.Background(), req)
	if err != nil {
		t.Errorf("TokenAuthenticator.Authenticate() error = %v", err)
	}

	// Verify that Refresh was called to get token
	if client.getCallCount() != 1 {
		t.Errorf("Expected 1 call to token client, got %d", client.getCallCount())
	}

	// Verify the Authorization header was set
	authHeader := req.Header.Get("Authorization")
	if authHeader != "Bearer fresh-token" {
		t.Errorf("Authorization header = %q, want %q", authHeader, "Bearer fresh-token")
	}
}

func TestTokenAuthenticator_Interface(t *testing.T) {
	var _ Authenticator = (*TokenAuthenticator)(nil)
}

func TestTokenAuthenticator_ExpiryCalculation(t *testing.T) {
	client := &mockTokenClient{
		token:     "expiry-test-token",
		expiresIn: 7200, // 2 hours
	}

	config := &types.AuthConfig{
		AuthType: types.AuthTypeToken,
	}

	auth, err := NewTokenAuthenticator(config, client)
	if err != nil {
		t.Fatalf("NewTokenAuthenticator() error = %v", err)
	}

	beforeRefresh := time.Now()
	err = auth.Refresh(context.Background())
	if err != nil {
		t.Fatalf("TokenAuthenticator.Refresh() error = %v", err)
	}
	afterRefresh := time.Now()

	auth.mu.RLock()
	expiresAt := auth.expiresAt
	auth.mu.RUnlock()

	// Verify that expiry time is approximately 2 hours from now
	expectedMinExpiry := beforeRefresh.Add(2 * time.Hour)
	expectedMaxExpiry := afterRefresh.Add(2 * time.Hour)

	if expiresAt.Before(expectedMinExpiry) || expiresAt.After(expectedMaxExpiry) {
		t.Errorf("Token expiry time %v is not within expected range %v-%v",
			expiresAt, expectedMinExpiry, expectedMaxExpiry)
	}
}

func TestTokenAuthenticator_DefaultExpiry(t *testing.T) {
	client := &mockTokenClient{
		token:     "default-expiry-token",
		expiresIn: 0, // No expiry specified
	}

	config := &types.AuthConfig{
		AuthType: types.AuthTypeToken,
	}

	auth, err := NewTokenAuthenticator(config, client)
	if err != nil {
		t.Fatalf("NewTokenAuthenticator() error = %v", err)
	}

	beforeRefresh := time.Now()
	err = auth.Refresh(context.Background())
	if err != nil {
		t.Fatalf("TokenAuthenticator.Refresh() error = %v", err)
	}
	afterRefresh := time.Now()

	auth.mu.RLock()
	expiresAt := auth.expiresAt
	auth.mu.RUnlock()

	// Verify that default expiry of 1 hour was used
	expectedMinExpiry := beforeRefresh.Add(1 * time.Hour)
	expectedMaxExpiry := afterRefresh.Add(1 * time.Hour)

	if expiresAt.Before(expectedMinExpiry) || expiresAt.After(expectedMaxExpiry) {
		t.Errorf("Token expiry time %v is not within expected default range %v-%v",
			expiresAt, expectedMinExpiry, expectedMaxExpiry)
	}
}