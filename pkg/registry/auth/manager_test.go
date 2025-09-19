package auth

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

// mockCredentialStore implements types.CredentialStore for testing
type mockCredentialStore struct {
	credentials map[string]*types.AuthConfig
	getError    error
	storeError  error
	removeError error
	callCounts  map[string]int
	mu          sync.RWMutex
}

func newMockCredentialStore() *mockCredentialStore {
	return &mockCredentialStore{
		credentials: make(map[string]*types.AuthConfig),
		callCounts:  make(map[string]int),
	}
}

func (m *mockCredentialStore) GetCredentials(registry string) (*types.AuthConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCounts["get"]++

	if m.getError != nil {
		return nil, m.getError
	}

	creds, exists := m.credentials[registry]
	if !exists {
		return &types.AuthConfig{AuthType: types.AuthTypeNone}, nil
	}
	return creds, nil
}

func (m *mockCredentialStore) StoreCredentials(registry string, creds *types.AuthConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCounts["store"]++

	if m.storeError != nil {
		return m.storeError
	}

	m.credentials[registry] = creds
	return nil
}

func (m *mockCredentialStore) RemoveCredentials(registry string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCounts["remove"]++

	if m.removeError != nil {
		return m.removeError
	}

	delete(m.credentials, registry)
	return nil
}

func (m *mockCredentialStore) setCredentials(registry string, creds *types.AuthConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.credentials[registry] = creds
}

func (m *mockCredentialStore) getCallCount(operation string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCounts[operation]
}

func (m *mockCredentialStore) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCounts = make(map[string]int)
}

func TestNewManager(t *testing.T) {
	tests := []struct {
		name  string
		store types.CredentialStore
	}{
		{
			name:  "with credential store",
			store: newMockCredentialStore(),
		},
		{
			name:  "without credential store",
			store: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager(tt.store)

			if manager == nil {
				t.Fatal("NewManager() returned nil")
			}

			if manager.store != tt.store {
				t.Error("NewManager() did not set store correctly")
			}

			if manager.auths == nil {
				t.Error("NewManager() did not initialize auths map")
			}
		})
	}
}

func TestManager_GetAuthenticator(t *testing.T) {
	tests := []struct {
		name        string
		registry    string
		setupStore  func(*mockCredentialStore)
		expectError bool
		expectType  string
	}{
		{
			name:     "empty registry returns error",
			registry: "",
			setupStore: func(store *mockCredentialStore) {
				// No setup needed
			},
			expectError: true,
		},
		{
			name:     "basic auth from store",
			registry: "registry.example.com",
			setupStore: func(store *mockCredentialStore) {
				store.setCredentials("registry.example.com", &types.AuthConfig{
					AuthType: types.AuthTypeBasic,
					Username: "user",
					Password: "pass",
				})
			},
			expectType: "*auth.BasicAuthenticator",
		},
		{
			name:     "token auth from store",
			registry: "private-registry.com",
			setupStore: func(store *mockCredentialStore) {
				store.setCredentials("private-registry.com", &types.AuthConfig{
					AuthType: types.AuthTypeToken,
					Token:    "token123",
				})
			},
			expectType: "*auth.TokenAuthenticator",
		},
		{
			name:     "no auth from store",
			registry: "public-registry.com",
			setupStore: func(store *mockCredentialStore) {
				store.setCredentials("public-registry.com", &types.AuthConfig{
					AuthType: types.AuthTypeNone,
				})
			},
			expectType: "*auth.NoOpAuthenticator",
		},
		{
			name:     "store returns error",
			registry: "error-registry.com",
			setupStore: func(store *mockCredentialStore) {
				store.getError = fmt.Errorf("credential store error")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMockCredentialStore()
			if tt.setupStore != nil {
				tt.setupStore(store)
			}

			manager := NewManager(store)

			auth, err := manager.GetAuthenticator(context.Background(), tt.registry)

			if tt.expectError {
				if err == nil {
					t.Errorf("GetAuthenticator() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetAuthenticator() unexpected error: %v", err)
				return
			}

			if auth == nil {
				t.Fatal("GetAuthenticator() returned nil authenticator")
			}

			// Check authenticator type
			if tt.expectType != "" {
				authType := getTypeName(auth)
				if authType != tt.expectType {
					t.Errorf("GetAuthenticator() created type %s, want %s", authType, tt.expectType)
				}
			}

			// Verify authenticator was cached
			auth2, err := manager.GetAuthenticator(context.Background(), tt.registry)
			if err != nil {
				t.Errorf("GetAuthenticator() second call error: %v", err)
			}

			if auth != auth2 {
				t.Error("GetAuthenticator() should return cached authenticator on second call")
			}
		})
	}
}

func TestManager_GetAuthenticator_NoStore(t *testing.T) {
	manager := NewManager(nil)

	auth, err := manager.GetAuthenticator(context.Background(), "registry.example.com")
	if err != nil {
		t.Errorf("GetAuthenticator() with no store error: %v", err)
	}

	if auth == nil {
		t.Fatal("GetAuthenticator() returned nil authenticator")
	}

	// Should default to no-op auth when no store is configured
	authType := getTypeName(auth)
	if authType != "*auth.NoOpAuthenticator" {
		t.Errorf("GetAuthenticator() with no store created type %s, want *auth.NoOpAuthenticator", authType)
	}
}

func TestManager_GetAuthenticator_RefreshInvalidAuth(t *testing.T) {
	store := newMockCredentialStore()
	store.setCredentials("registry.example.com", &types.AuthConfig{
		AuthType: types.AuthTypeBasic,
		Username: "user",
		Password: "pass",
	})

	manager := NewManager(store)

	// Get authenticator first time to populate cache
	_, err := manager.GetAuthenticator(context.Background(), "registry.example.com")
	if err != nil {
		t.Fatalf("GetAuthenticator() first call error: %v", err)
	}

	// Manually add an invalid authenticator to cache to test refresh logic
	invalidAuth := &mockAuthenticator{isValid: false}
	manager.mu.Lock()
	manager.auths["registry.example.com"] = invalidAuth
	manager.mu.Unlock()

	store.reset()

	// Get authenticator again - should refresh due to invalid auth
	auth2, err := manager.GetAuthenticator(context.Background(), "registry.example.com")
	if err != nil {
		t.Errorf("GetAuthenticator() refresh call error: %v", err)
	}

	// Should create new authenticator, not return the invalid cached one
	if auth2 == invalidAuth {
		t.Error("GetAuthenticator() returned invalid cached authenticator instead of creating new one")
	}

	// Should have called store to get credentials again
	if store.getCallCount("get") != 1 {
		t.Errorf("Expected 1 call to credential store during refresh, got %d", store.getCallCount("get"))
	}
}

func TestManager_Clear(t *testing.T) {
	store := newMockCredentialStore()
	store.setCredentials("registry1.com", &types.AuthConfig{
		AuthType: types.AuthTypeBasic,
		Username: "user",
		Password: "pass",
	})
	store.setCredentials("registry2.com", &types.AuthConfig{
		AuthType: types.AuthTypeToken,
		Token:    "token123",
	})

	manager := NewManager(store)

	// Get authenticators to populate cache
	_, err := manager.GetAuthenticator(context.Background(), "registry1.com")
	if err != nil {
		t.Fatalf("GetAuthenticator() error: %v", err)
	}
	_, err = manager.GetAuthenticator(context.Background(), "registry2.com")
	if err != nil {
		t.Fatalf("GetAuthenticator() error: %v", err)
	}

	// Verify both are cached
	manager.mu.RLock()
	if len(manager.auths) != 2 {
		t.Fatalf("Expected 2 cached authenticators, got %d", len(manager.auths))
	}
	manager.mu.RUnlock()

	// Clear one registry
	manager.Clear("registry1.com")

	// Verify only one remains
	manager.mu.RLock()
	if len(manager.auths) != 1 {
		t.Errorf("Expected 1 cached authenticator after clear, got %d", len(manager.auths))
	}
	if _, exists := manager.auths["registry1.com"]; exists {
		t.Error("registry1.com should be cleared from cache")
	}
	if _, exists := manager.auths["registry2.com"]; !exists {
		t.Error("registry2.com should still be in cache")
	}
	manager.mu.RUnlock()
}

func TestManager_ClearAll(t *testing.T) {
	store := newMockCredentialStore()
	store.setCredentials("registry1.com", &types.AuthConfig{
		AuthType: types.AuthTypeBasic,
		Username: "user",
		Password: "pass",
	})
	store.setCredentials("registry2.com", &types.AuthConfig{
		AuthType: types.AuthTypeToken,
		Token:    "token123",
	})

	manager := NewManager(store)

	// Get authenticators to populate cache
	_, err := manager.GetAuthenticator(context.Background(), "registry1.com")
	if err != nil {
		t.Fatalf("GetAuthenticator() error: %v", err)
	}
	_, err = manager.GetAuthenticator(context.Background(), "registry2.com")
	if err != nil {
		t.Fatalf("GetAuthenticator() error: %v", err)
	}

	// Verify both are cached
	manager.mu.RLock()
	if len(manager.auths) != 2 {
		t.Fatalf("Expected 2 cached authenticators, got %d", len(manager.auths))
	}
	manager.mu.RUnlock()

	// Clear all
	manager.ClearAll()

	// Verify cache is empty
	manager.mu.RLock()
	if len(manager.auths) != 0 {
		t.Errorf("Expected empty cache after ClearAll, got %d entries", len(manager.auths))
	}
	manager.mu.RUnlock()
}

func TestManager_ConcurrentAccess(t *testing.T) {
	store := newMockCredentialStore()

	// Set up multiple registries
	for i := 0; i < 5; i++ {
		registry := fmt.Sprintf("registry%d.com", i)
		store.setCredentials(registry, &types.AuthConfig{
			AuthType: types.AuthTypeBasic,
			Username: "user" + fmt.Sprint(i),
			Password: "pass" + fmt.Sprint(i),
		})
	}

	manager := NewManager(store)

	var wg sync.WaitGroup
	errors := make(chan error, 50)

	// Test concurrent GetAuthenticator calls
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < 5; j++ {
				registry := fmt.Sprintf("registry%d.com", j)

				auth, err := manager.GetAuthenticator(context.Background(), registry)
				if err != nil {
					errors <- fmt.Errorf("worker %d: GetAuthenticator error: %w", workerID, err)
					return
				}

				if auth == nil {
					errors <- fmt.Errorf("worker %d: GetAuthenticator returned nil", workerID)
					return
				}

				if !auth.IsValid() {
					errors <- fmt.Errorf("worker %d: GetAuthenticator returned invalid auth", workerID)
					return
				}
			}
		}(i)
	}

	// Test concurrent Clear calls
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < 5; j++ {
				registry := fmt.Sprintf("registry%d.com", j)
				manager.Clear(registry)
			}
		}(i)
	}

	// Test concurrent ClearAll calls
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			manager.ClearAll()
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}

func TestManager_GetAuthenticator_CreateError(t *testing.T) {
	store := newMockCredentialStore()
	store.setCredentials("invalid-registry.com", &types.AuthConfig{
		AuthType: types.AuthTypeBasic,
		// Missing username/password to cause creation error
	})

	manager := NewManager(store)

	auth, err := manager.GetAuthenticator(context.Background(), "invalid-registry.com")

	if err == nil {
		t.Error("GetAuthenticator() expected error for invalid config, got nil")
	}

	if auth != nil {
		t.Error("GetAuthenticator() expected nil auth on error, got authenticator")
	}

	// Verify error message contains expected text
	expectedErrorText := "failed to create authenticator"
	if err != nil && !containsString(err.Error(), expectedErrorText) {
		t.Errorf("GetAuthenticator() error = %q, expected to contain %q", err.Error(), expectedErrorText)
	}
}

func TestManager_GetAuthenticator_CacheAfterCreation(t *testing.T) {
	store := newMockCredentialStore()
	store.setCredentials("cache-test.com", &types.AuthConfig{
		AuthType: types.AuthTypeBasic,
		Username: "user",
		Password: "pass",
	})

	manager := NewManager(store)

	// First call should create and cache authenticator
	store.reset()
	auth1, err := manager.GetAuthenticator(context.Background(), "cache-test.com")
	if err != nil {
		t.Fatalf("GetAuthenticator() first call error: %v", err)
	}

	if store.getCallCount("get") != 1 {
		t.Errorf("Expected 1 call to store on first access, got %d", store.getCallCount("get"))
	}

	// Second call should return cached authenticator without calling store
	store.reset()
	auth2, err := manager.GetAuthenticator(context.Background(), "cache-test.com")
	if err != nil {
		t.Errorf("GetAuthenticator() second call error: %v", err)
	}

	if store.getCallCount("get") != 0 {
		t.Errorf("Expected 0 calls to store on cached access, got %d", store.getCallCount("get"))
	}

	// Should be same authenticator instance
	if auth1 != auth2 {
		t.Error("GetAuthenticator() should return same cached instance")
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsStringHelper(s, substr)))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}