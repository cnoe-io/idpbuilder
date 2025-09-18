// Package auth provides authentication implementations for OCI registry operations.
// It supports multiple authentication methods including basic auth, bearer tokens,
// and anonymous access.
package auth

import (
	"context"
	"net/http"
	"testing"

	"github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

func TestNewAuthenticator(t *testing.T) {
	tests := []struct {
		name        string
		config      *types.AuthConfig
		wantType    string
		expectError bool
	}{
		{
			name:     "nil config creates NoOp authenticator",
			config:   nil,
			wantType: "*auth.NoOpAuthenticator",
		},
		{
			name: "basic auth type creates BasicAuthenticator",
			config: &types.AuthConfig{
				AuthType: types.AuthTypeBasic,
				Username: "user",
				Password: "pass",
			},
			wantType: "*auth.BasicAuthenticator",
		},
		{
			name: "token auth type creates TokenAuthenticator",
			config: &types.AuthConfig{
				AuthType: types.AuthTypeToken,
				Token:    "test-token",
			},
			wantType: "*auth.TokenAuthenticator",
		},
		{
			name: "none auth type creates NoOp authenticator",
			config: &types.AuthConfig{
				AuthType: types.AuthTypeNone,
			},
			wantType: "*auth.NoOpAuthenticator",
		},
		{
			name: "unsupported auth type returns error",
			config: &types.AuthConfig{
				AuthType: types.AuthTypeOAuth2,
			},
			expectError: true,
		},
		{
			name: "basic auth without credentials returns error",
			config: &types.AuthConfig{
				AuthType: types.AuthTypeBasic,
				Username: "user",
				// missing password
			},
			expectError: true,
		},
		{
			name: "token auth without token returns error",
			config: &types.AuthConfig{
				AuthType: types.AuthTypeToken,
				// missing token and no client provided
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewAuthenticator(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("NewAuthenticator() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewAuthenticator() unexpected error: %v", err)
				return
			}

			if auth == nil {
				t.Errorf("NewAuthenticator() returned nil authenticator")
				return
			}

			// Check if we got the expected type
			authType := getTypeName(auth)
			if authType != tt.wantType {
				t.Errorf("NewAuthenticator() created type %s, want %s", authType, tt.wantType)
			}

			// Verify the authenticator implements the interface correctly
			if !auth.IsValid() && tt.wantType != "*auth.TokenAuthenticator" {
				t.Errorf("NewAuthenticator() created invalid authenticator")
			}
		})
	}
}

func TestNoOpAuthenticator(t *testing.T) {
	auth := NewNoOpAuthenticator()

	if auth == nil {
		t.Fatal("NewNoOpAuthenticator() returned nil")
	}

	t.Run("Authenticate does nothing", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		originalHeaders := len(req.Header)

		err := auth.Authenticate(context.Background(), req)
		if err != nil {
			t.Errorf("NoOpAuthenticator.Authenticate() error = %v, want nil", err)
		}

		// Should not add any headers
		if len(req.Header) != originalHeaders {
			t.Errorf("NoOpAuthenticator.Authenticate() modified request headers")
		}
	})

	t.Run("Refresh does nothing", func(t *testing.T) {
		err := auth.Refresh(context.Background())
		if err != nil {
			t.Errorf("NoOpAuthenticator.Refresh() error = %v, want nil", err)
		}
	})

	t.Run("IsValid always returns true", func(t *testing.T) {
		if !auth.IsValid() {
			t.Errorf("NoOpAuthenticator.IsValid() = false, want true")
		}
	})
}

func TestNoOpAuthenticator_Interface(t *testing.T) {
	var _ Authenticator = (*NoOpAuthenticator)(nil)
}

// Helper function to get the type name for testing
func getTypeName(auth Authenticator) string {
	switch auth.(type) {
	case *NoOpAuthenticator:
		return "*auth.NoOpAuthenticator"
	case *BasicAuthenticator:
		return "*auth.BasicAuthenticator"
	case *TokenAuthenticator:
		return "*auth.TokenAuthenticator"
	default:
		return "unknown"
	}
}