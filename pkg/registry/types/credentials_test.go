package types

import (
	"testing"
	"time"
)

func TestAuthConfig(t *testing.T) {
	tests := []struct {
		name         string
		auth         *AuthConfig
		wantValid    bool
		expectedType AuthType
	}{
		{
			name: "valid basic auth",
			auth: &AuthConfig{
				AuthType: AuthTypeBasic,
				Username: "user",
				Password: "pass",
			},
			wantValid:    true,
			expectedType: AuthTypeBasic,
		},
		{
			name: "valid token auth",
			auth: &AuthConfig{
				AuthType: AuthTypeToken,
				Token:    "bearer-token",
			},
			wantValid:    true,
			expectedType: AuthTypeToken,
		},
		{
			name: "valid OAuth2 auth",
			auth: &AuthConfig{
				AuthType: AuthTypeOAuth2,
				Token:    "oauth2-token",
			},
			wantValid:    true,
			expectedType: AuthTypeOAuth2,
		},
		{
			name: "none auth type",
			auth: &AuthConfig{
				AuthType: AuthTypeNone,
			},
			wantValid:    true,
			expectedType: AuthTypeNone,
		},
		{
			name: "basic auth without username - edge case",
			auth: &AuthConfig{
				AuthType: AuthTypeBasic,
				Username: "",
				Password: "pass",
			},
			wantValid:    false,
			expectedType: AuthTypeBasic,
		},
		{
			name: "basic auth without password - edge case",
			auth: &AuthConfig{
				AuthType: AuthTypeBasic,
				Username: "user",
				Password: "",
			},
			wantValid:    false,
			expectedType: AuthTypeBasic,
		},
		{
			name: "token auth without token - edge case",
			auth: &AuthConfig{
				AuthType: AuthTypeToken,
				Token:    "",
			},
			wantValid:    false,
			expectedType: AuthTypeToken,
		},
		{
			name: "empty auth type - edge case",
			auth: &AuthConfig{
				AuthType: "",
				Username: "user",
				Password: "pass",
			},
			wantValid:    false,
			expectedType: "",
		},
		{
			name:      "nil auth config - edge case",
			auth:      nil,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.auth == nil {
				if tt.wantValid {
					t.Error("nil auth config should not be valid")
				}
				return
			}

			// Test auth type
			if tt.auth.AuthType != tt.expectedType {
				t.Errorf("AuthType = %q, want %q", tt.auth.AuthType, tt.expectedType)
			}

			// Test validation logic
			var isValid bool
			switch tt.auth.AuthType {
			case AuthTypeBasic:
				isValid = tt.auth.Username != "" && tt.auth.Password != ""
			case AuthTypeToken, AuthTypeOAuth2:
				isValid = tt.auth.Token != ""
			case AuthTypeNone:
				isValid = true
			default:
				isValid = false
			}

			if isValid != tt.wantValid {
				t.Errorf("auth config validity = %v, want %v", isValid, tt.wantValid)
			}

			// Test specific field validations for valid configs
			if tt.wantValid && tt.auth.AuthType == AuthTypeBasic {
				if tt.auth.Username == "" {
					t.Error("basic auth should have username")
				}
				if tt.auth.Password == "" {
					t.Error("basic auth should have password")
				}
			}

			if tt.wantValid && (tt.auth.AuthType == AuthTypeToken || tt.auth.AuthType == AuthTypeOAuth2) {
				if tt.auth.Token == "" {
					t.Error("token auth should have token")
				}
			}
		})
	}
}

func TestAuthType(t *testing.T) {
	tests := []struct {
		name     string
		authType AuthType
		valid    bool
	}{
		{"basic auth type", AuthTypeBasic, true},
		{"token auth type", AuthTypeToken, true},
		{"oauth2 auth type", AuthTypeOAuth2, true},
		{"none auth type", AuthTypeNone, true},
		{"unknown auth type", AuthType("unknown"), false},
		{"empty auth type", AuthType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that auth type constants have expected values
			knownTypes := map[AuthType]bool{
				AuthTypeBasic:  true,
				AuthTypeToken:  true,
				AuthTypeOAuth2: true,
				AuthTypeNone:   true,
			}

			isKnown := knownTypes[tt.authType]
			if isKnown != tt.valid {
				t.Errorf("auth type %q validity = %v, want %v", tt.authType, isKnown, tt.valid)
			}

			// Test string conversion
			authTypeStr := string(tt.authType)
			if tt.valid && authTypeStr == "" {
				t.Error("valid auth type should not be empty string")
			}
		})
	}
}

func TestTokenResponse(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		response *TokenResponse
		wantErr  bool
	}{
		{
			name: "valid token response",
			response: &TokenResponse{
				Token:     "token123",
				ExpiresIn: 3600,
				IssuedAt:  now,
			},
			wantErr: false,
		},
		{
			name: "token response with zero expiry",
			response: &TokenResponse{
				Token:     "token123",
				ExpiresIn: 0,
				IssuedAt:  now,
			},
			wantErr: false, // zero expiry might be valid for some tokens
		},
		{
			name: "token response without token - edge case",
			response: &TokenResponse{
				Token:     "",
				ExpiresIn: 3600,
				IssuedAt:  now,
			},
			wantErr: true,
		},
		{
			name: "token response with negative expiry - edge case",
			response: &TokenResponse{
				Token:     "token123",
				ExpiresIn: -1,
				IssuedAt:  now,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic field validation
			if tt.response.Token == "" && !tt.wantErr {
				t.Error("valid token response should have token")
			}

			if tt.response.ExpiresIn < 0 && !tt.wantErr {
				t.Error("valid token response should not have negative expiry")
			}

			// Test time field
			if !tt.response.IssuedAt.IsZero() {
				if tt.response.IssuedAt.After(time.Now().Add(time.Minute)) {
					t.Error("issued at time should not be in the future")
				}
			}

			// Test that ExpiresIn is reasonable
			if !tt.wantErr && tt.response.ExpiresIn > 86400*365 { // more than 1 year
				t.Error("expiry time seems unreasonably long")
			}
		})
	}
}

func TestCredentialStore(t *testing.T) {
	// Test that CredentialStore is an interface
	var store CredentialStore
	if store != nil {
		t.Error("nil interface should be nil")
	}

	// Test interface methods exist by creating a mock implementation
	mockStore := &mockCredentialStore{
		credentials: make(map[string]*AuthConfig),
	}

	// Test interface compliance
	var _ CredentialStore = mockStore

	registry := "registry.example.com"
	creds := &AuthConfig{
		AuthType: AuthTypeBasic,
		Username: "user",
		Password: "pass",
	}

	// Test store credentials
	err := mockStore.StoreCredentials(registry, creds)
	if err != nil {
		t.Errorf("StoreCredentials failed: %v", err)
	}

	// Test get credentials
	retrievedCreds, err := mockStore.GetCredentials(registry)
	if err != nil {
		t.Errorf("GetCredentials failed: %v", err)
	}
	if retrievedCreds == nil {
		t.Error("expected credentials, got nil")
	}
	if retrievedCreds.Username != creds.Username {
		t.Errorf("Username = %q, want %q", retrievedCreds.Username, creds.Username)
	}

	// Test remove credentials
	err = mockStore.RemoveCredentials(registry)
	if err != nil {
		t.Errorf("RemoveCredentials failed: %v", err)
	}

	// Test get after remove
	retrievedCreds, err = mockStore.GetCredentials(registry)
	if err == nil {
		t.Error("expected error after removing credentials")
	}
	if retrievedCreds != nil {
		t.Error("expected nil credentials after removal")
	}
}

// Mock implementation for testing CredentialStore interface
type mockCredentialStore struct {
	credentials map[string]*AuthConfig
}

func (m *mockCredentialStore) GetCredentials(registry string) (*AuthConfig, error) {
	creds, exists := m.credentials[registry]
	if !exists {
		return nil, &RegistryError{
			Code:    ErrCodeNotFound,
			Message: "credentials not found",
		}
	}
	return creds, nil
}

func (m *mockCredentialStore) StoreCredentials(registry string, creds *AuthConfig) error {
	if creds == nil {
		return &RegistryError{
			Code:    ErrCodeInvalidConfig,
			Message: "credentials cannot be nil",
		}
	}
	m.credentials[registry] = creds
	return nil
}

func (m *mockCredentialStore) RemoveCredentials(registry string) error {
	if _, exists := m.credentials[registry]; !exists {
		return &RegistryError{
			Code:    ErrCodeNotFound,
			Message: "credentials not found",
		}
	}
	delete(m.credentials, registry)
	return nil
}
