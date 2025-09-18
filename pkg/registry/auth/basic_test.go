package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"

	"github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

func TestNewBasicAuthenticator(t *testing.T) {
	tests := []struct {
		name        string
		config      *types.AuthConfig
		expectError bool
		wantUser    string
		wantPass    string
	}{
		{
			name: "valid credentials creates authenticator",
			config: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
			},
			wantUser: "testuser",
			wantPass: "testpass",
		},
		{
			name:        "nil config returns error",
			config:      nil,
			expectError: true,
		},
		{
			name: "missing username returns error",
			config: &types.AuthConfig{
				Username: "",
				Password: "testpass",
			},
			expectError: true,
		},
		{
			name: "missing password returns error",
			config: &types.AuthConfig{
				Username: "testuser",
				Password: "",
			},
			expectError: true,
		},
		{
			name: "empty username and password returns error",
			config: &types.AuthConfig{
				Username: "",
				Password: "",
			},
			expectError: true,
		},
		{
			name: "special characters in credentials",
			config: &types.AuthConfig{
				Username: "user@domain.com",
				Password: "P@ssw0rd!#$",
			},
			wantUser: "user@domain.com",
			wantPass: "P@ssw0rd!#$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewBasicAuthenticator(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("NewBasicAuthenticator() expected error, got nil")
				}
				if auth != nil {
					t.Errorf("NewBasicAuthenticator() expected nil auth on error, got %v", auth)
				}
				return
			}

			if err != nil {
				t.Errorf("NewBasicAuthenticator() unexpected error: %v", err)
				return
			}

			if auth == nil {
				t.Fatal("NewBasicAuthenticator() returned nil")
			}

			// Verify internal fields are set correctly
			if auth.username != tt.wantUser {
				t.Errorf("NewBasicAuthenticator() username = %q, want %q", auth.username, tt.wantUser)
			}

			if auth.password != tt.wantPass {
				t.Errorf("NewBasicAuthenticator() password = %q, want %q", auth.password, tt.wantPass)
			}

			// Verify encoded credential is correct
			expectedEncoded := base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf("%s:%s", tt.wantUser, tt.wantPass)),
			)
			if auth.encoded != expectedEncoded {
				t.Errorf("NewBasicAuthenticator() encoded = %q, want %q", auth.encoded, expectedEncoded)
			}
		})
	}
}

func TestBasicAuthenticator_Authenticate(t *testing.T) {
	auth, err := NewBasicAuthenticator(&types.AuthConfig{
		Username: "testuser",
		Password: "testpass",
	})
	if err != nil {
		t.Fatalf("NewBasicAuthenticator() error = %v", err)
	}

	tests := []struct {
		name   string
		method string
		url    string
	}{
		{
			name:   "GET request",
			method: "GET",
			url:    "http://example.com",
		},
		{
			name:   "POST request",
			method: "POST",
			url:    "https://registry.example.com/v2/",
		},
		{
			name:   "PUT request",
			method: "PUT",
			url:    "https://registry.example.com/v2/repo/manifests/latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Ensure no auth header exists initially
			if req.Header.Get("Authorization") != "" {
				t.Fatal("Request already has Authorization header")
			}

			err = auth.Authenticate(context.Background(), req)
			if err != nil {
				t.Errorf("BasicAuthenticator.Authenticate() error = %v", err)
			}

			// Check that Authorization header was added
			authHeader := req.Header.Get("Authorization")
			if authHeader == "" {
				t.Error("BasicAuthenticator.Authenticate() did not set Authorization header")
			}

			// Verify the header format
			expectedHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte("testuser:testpass"))
			if authHeader != expectedHeader {
				t.Errorf("Authorization header = %q, want %q", authHeader, expectedHeader)
			}
		})
	}
}

func TestBasicAuthenticator_Refresh(t *testing.T) {
	auth, err := NewBasicAuthenticator(&types.AuthConfig{
		Username: "testuser",
		Password: "testpass",
	})
	if err != nil {
		t.Fatalf("NewBasicAuthenticator() error = %v", err)
	}

	// Basic auth doesn't need refresh, should always return nil
	err = auth.Refresh(context.Background())
	if err != nil {
		t.Errorf("BasicAuthenticator.Refresh() error = %v, want nil", err)
	}
}

func TestBasicAuthenticator_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		config   *types.AuthConfig
		expected bool
	}{
		{
			name: "valid authenticator is valid",
			config: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
			},
			expected: true,
		},
		{
			name: "authenticator with empty credentials after creation",
			config: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
			},
			expected: true, // Even after creation, it should be valid since we have encoded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewBasicAuthenticator(tt.config)
			if err != nil {
				t.Fatalf("NewBasicAuthenticator() error = %v", err)
			}

			isValid := auth.IsValid()
			if isValid != tt.expected {
				t.Errorf("BasicAuthenticator.IsValid() = %v, want %v", isValid, tt.expected)
			}
		})
	}
}

func TestBasicAuthenticator_EncodedCredentialFormat(t *testing.T) {
	testCases := []struct {
		username string
		password string
	}{
		{"user", "pass"},
		{"admin", "secret123"},
		{"test@example.com", "complex!P@ssw0rd"},
		{"user with spaces", "pass with spaces"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s:%s", tc.username, tc.password), func(t *testing.T) {
			auth, err := NewBasicAuthenticator(&types.AuthConfig{
				Username: tc.username,
				Password: tc.password,
			})
			if err != nil {
				t.Fatalf("NewBasicAuthenticator() error = %v", err)
			}

			// Test that we can decode the encoded credential
			expectedCreds := tc.username + ":" + tc.password
			decoded, err := base64.StdEncoding.DecodeString(auth.encoded)
			if err != nil {
				t.Errorf("Failed to decode encoded credential: %v", err)
			}

			if string(decoded) != expectedCreds {
				t.Errorf("Decoded credential = %q, want %q", string(decoded), expectedCreds)
			}
		})
	}
}

func TestBasicAuthenticator_Interface(t *testing.T) {
	var _ Authenticator = (*BasicAuthenticator)(nil)
}

func TestBasicAuthenticator_ConcurrentAccess(t *testing.T) {
	auth, err := NewBasicAuthenticator(&types.AuthConfig{
		Username: "testuser",
		Password: "testpass",
	})
	if err != nil {
		t.Fatalf("NewBasicAuthenticator() error = %v", err)
	}

	// Test concurrent access to methods
	done := make(chan struct{})
	errors := make(chan error, 100)

	// Start multiple goroutines calling methods concurrently
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- struct{}{} }()

			req, _ := http.NewRequest("GET", "http://example.com", nil)

			for j := 0; j < 10; j++ {
				if err := auth.Authenticate(context.Background(), req); err != nil {
					errors <- err
					return
				}

				if err := auth.Refresh(context.Background()); err != nil {
					errors <- err
					return
				}

				if !auth.IsValid() {
					errors <- fmt.Errorf("auth became invalid during concurrent access")
					return
				}
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}