package helpers

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

func TestNewAuthenticatedClient(t *testing.T) {
	tests := []struct {
		name       string
		authConfig *types.AuthConfig
		options    *types.ConnectionOptions
		wantErr    bool
	}{
		{
			name: "Basic auth configuration",
			authConfig: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				AuthType: types.AuthTypeBasic,
			},
			options: nil,
			wantErr: false,
		},
		{
			name: "Token auth configuration",
			authConfig: &types.AuthConfig{
				Token:    "test-token",
				AuthType: types.AuthTypeToken,
			},
			options: nil,
			wantErr: false,
		},
		{
			name: "With connection options",
			authConfig: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				AuthType: types.AuthTypeBasic,
			},
			options: &types.ConnectionOptions{
				HTTPClient: &http.Client{
					Timeout: 60 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "With TLS config",
			authConfig: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				AuthType: types.AuthTypeBasic,
			},
			options: &types.ConnectionOptions{
				TLSConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			wantErr: false,
		},
		{
			name: "With custom transport",
			authConfig: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				AuthType: types.AuthTypeBasic,
			},
			options: &types.ConnectionOptions{
				HTTPClient: &http.Client{
					Transport: &http.Transport{
						MaxIdleConns: 10,
					},
					Timeout: 45 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "With redirect policy",
			authConfig: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				AuthType: types.AuthTypeBasic,
			},
			options: &types.ConnectionOptions{
				HTTPClient: &http.Client{
					CheckRedirect: func(req *http.Request, via []*http.Request) error {
						return http.ErrUseLastResponse
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "Nil auth config",
			authConfig: nil,
			options:    nil,
			wantErr:    false, // Actually works with nil auth config - creates client without auth
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewAuthenticatedClient(tt.authConfig, tt.options)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewAuthenticatedClient() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewAuthenticatedClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if client == nil {
				t.Errorf("NewAuthenticatedClient() returned nil client")
				return
			}

			// Check client has proper timeout
			if client.Timeout <= 0 {
				t.Errorf("NewAuthenticatedClient() client has no timeout set")
			}

			// Check transport is not nil
			if client.Transport == nil {
				t.Errorf("NewAuthenticatedClient() client transport is nil")
			}

			// If options provided with timeout, verify it's applied
			if tt.options != nil && tt.options.HTTPClient != nil && tt.options.HTTPClient.Timeout > 0 {
				if client.Timeout != tt.options.HTTPClient.Timeout {
					t.Errorf("NewAuthenticatedClient() timeout = %v, want %v",
						client.Timeout, tt.options.HTTPClient.Timeout)
				}
			}
		})
	}
}

func TestNewRegistryClient(t *testing.T) {
	tests := []struct {
		name       string
		config     *types.RegistryConfig
		authConfig *types.AuthConfig
		wantErr    bool
	}{
		{
			name: "Basic registry config",
			config: &types.RegistryConfig{
				URL: "registry.example.com:5000",
			},
			authConfig: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				AuthType: types.AuthTypeBasic,
			},
			wantErr: false,
		},
		{
			name: "Registry with timeout",
			config: &types.RegistryConfig{
				URL:     "registry.example.com:5000",
				Timeout: 30 * time.Second,
			},
			authConfig: &types.AuthConfig{
				Token:    "test-token",
				AuthType: types.AuthTypeToken,
			},
			wantErr: false,
		},
		{
			name: "Insecure registry",
			config: &types.RegistryConfig{
				URL:      "registry.example.com:5000",
				Insecure: true,
			},
			authConfig: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				AuthType: types.AuthTypeBasic,
			},
			wantErr: false,
		},
		{
			name: "Skip TLS verify",
			config: &types.RegistryConfig{
				URL:           "registry.example.com:5000",
				SkipTLSVerify: true,
			},
			authConfig: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				AuthType: types.AuthTypeBasic,
			},
			wantErr: false,
		},
		{
			name: "Both insecure and skip TLS verify",
			config: &types.RegistryConfig{
				URL:           "registry.example.com:5000",
				Insecure:      true,
				SkipTLSVerify: true,
			},
			authConfig: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				AuthType: types.AuthTypeBasic,
			},
			wantErr: false,
		},
		{
			name:       "Nil registry config",
			config:     nil,
			authConfig: &types.AuthConfig{Username: "test", Password: "test", AuthType: types.AuthTypeBasic},
			wantErr:    true,
		},
		{
			name: "Invalid registry config",
			config: &types.RegistryConfig{
				URL: "", // Empty URL should cause validation error
			},
			authConfig: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				AuthType: types.AuthTypeBasic,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewRegistryClient(tt.config, tt.authConfig)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewRegistryClient() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewRegistryClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if client == nil {
				t.Errorf("NewRegistryClient() returned nil client")
				return
			}

			// Check client has proper timeout
			if client.Timeout <= 0 {
				t.Errorf("NewRegistryClient() client has no timeout set")
			}

			// If config has timeout, verify it's applied
			if tt.config.Timeout > 0 {
				if client.Timeout != tt.config.Timeout {
					t.Errorf("NewRegistryClient() timeout = %v, want %v",
						client.Timeout, tt.config.Timeout)
				}
			}
		})
	}
}

func TestCreateRequestWithAuth(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		method     string
		url        string
		body       interface{}
		authConfig *types.AuthConfig
		wantErr    bool
	}{
		{
			name:   "GET request with no body",
			method: "GET",
			url:    "https://registry.example.com/v2/",
			body:   nil,
			authConfig: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				AuthType: types.AuthTypeBasic,
			},
			wantErr: false,
		},
		{
			name:   "POST request with string body",
			method: "POST",
			url:    "https://registry.example.com/v2/manifest",
			body:   `{"test": "data"}`,
			authConfig: &types.AuthConfig{
				Token:    "test-token",
				AuthType: types.AuthTypeToken,
			},
			wantErr: false,
		},
		{
			name:   "PUT request with byte slice body",
			method: "PUT",
			url:    "https://registry.example.com/v2/manifest",
			body:   []byte(`{"test": "data"}`),
			authConfig: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				AuthType: types.AuthTypeBasic,
			},
			wantErr: false,
		},
		{
			name:   "POST request with io.Reader body",
			method: "POST",
			url:    "https://registry.example.com/v2/blob",
			body:   strings.NewReader("test content"),
			authConfig: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				AuthType: types.AuthTypeBasic,
			},
			wantErr: false,
		},
		{
			name:       "Request without auth config",
			method:     "GET",
			url:        "https://registry.example.com/v2/",
			body:       nil,
			authConfig: nil,
			wantErr:    false, // Should work without auth
		},
		{
			name:   "Request with unsupported body type",
			method: "POST",
			url:    "https://registry.example.com/v2/",
			body:   map[string]string{"test": "data"}, // Unsupported type
			authConfig: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				AuthType: types.AuthTypeBasic,
			},
			wantErr: true,
		},
		{
			name:   "Request with invalid URL",
			method: "GET",
			url:    "://invalid-url",
			body:   nil,
			authConfig: &types.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				AuthType: types.AuthTypeBasic,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := CreateRequestWithAuth(ctx, tt.method, tt.url, tt.body, tt.authConfig)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateRequestWithAuth() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("CreateRequestWithAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if req == nil {
				t.Errorf("CreateRequestWithAuth() returned nil request")
				return
			}

			// Check method and URL
			if req.Method != tt.method {
				t.Errorf("CreateRequestWithAuth() method = %v, want %v", req.Method, tt.method)
			}

			if req.URL.String() != tt.url {
				t.Errorf("CreateRequestWithAuth() url = %v, want %v", req.URL.String(), tt.url)
			}

			// Check required headers are set
			userAgent := req.Header.Get("User-Agent")
			if userAgent == "" {
				t.Errorf("CreateRequestWithAuth() User-Agent header not set")
			}

			accept := req.Header.Get("Accept")
			if accept == "" {
				t.Errorf("CreateRequestWithAuth() Accept header not set")
			}

			// Check context
			if req.Context() != ctx {
				t.Errorf("CreateRequestWithAuth() context not preserved")
			}

			// Check body handling based on type
			if tt.body != nil {
				if req.Body == nil {
					t.Errorf("CreateRequestWithAuth() body is nil when body was provided")
				}

				// Read body to verify it was set correctly
				if req.Body != nil {
					bodyBytes, err := io.ReadAll(req.Body)
					if err != nil {
						t.Errorf("CreateRequestWithAuth() failed to read body: %v", err)
					} else {
						switch v := tt.body.(type) {
						case string:
							if string(bodyBytes) != v {
								t.Errorf("CreateRequestWithAuth() body = %v, want %v", string(bodyBytes), v)
							}
						case []byte:
							if string(bodyBytes) != string(v) {
								t.Errorf("CreateRequestWithAuth() body = %v, want %v", string(bodyBytes), string(v))
							}
						}
					}
				}
			}
		})
	}
}

func TestCreateRequestWithAuth_ContextHandling(t *testing.T) {
	// Test that context cancellation is properly handled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req, err := CreateRequestWithAuth(ctx, "GET", "https://registry.example.com/v2/", nil, nil)
	if err != nil {
		t.Errorf("CreateRequestWithAuth() with cancelled context error = %v", err)
		return
	}

	if req == nil {
		t.Errorf("CreateRequestWithAuth() returned nil request")
		return
	}

	// Verify the context is cancelled
	select {
	case <-req.Context().Done():
		// Expected - context should be done
	default:
		t.Errorf("CreateRequestWithAuth() context should be cancelled")
	}
}

func TestCreateRequestWithAuth_HeaderValidation(t *testing.T) {
	ctx := context.Background()

	req, err := CreateRequestWithAuth(ctx, "GET", "https://registry.example.com/v2/", nil, nil)
	if err != nil {
		t.Errorf("CreateRequestWithAuth() error = %v", err)
		return
	}

	// Check specific header values
	userAgent := req.Header.Get("User-Agent")
	expectedUserAgent := "idpbuilder-registry-client/1.0"
	if userAgent != expectedUserAgent {
		t.Errorf("CreateRequestWithAuth() User-Agent = %v, want %v", userAgent, expectedUserAgent)
	}

	accept := req.Header.Get("Accept")
	expectedAccept := "application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json"
	if accept != expectedAccept {
		t.Errorf("CreateRequestWithAuth() Accept = %v, want %v", accept, expectedAccept)
	}
}