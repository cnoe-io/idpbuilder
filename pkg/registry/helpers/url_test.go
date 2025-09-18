package helpers

import (
	"testing"

	"github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

func TestParseRegistryURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *types.RegistryInfo
		wantErr  bool
	}{
		{
			name:  "HTTPS URL with port",
			input: "https://registry.example.com:8080",
			expected: &types.RegistryInfo{
				Scheme: "https",
				Host:   "registry.example.com",
				Port:   8080,
			},
		},
		{
			name:  "HTTP URL default port",
			input: "http://localhost",
			expected: &types.RegistryInfo{
				Scheme: "http",
				Host:   "localhost",
				Port:   80,
			},
		},
		{
			name:  "No scheme defaults to HTTPS",
			input: "registry.example.com",
			expected: &types.RegistryInfo{
				Scheme: "https",
				Host:   "registry.example.com",
				Port:   443,
			},
		},
		{
			name:    "Empty URL",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Invalid port",
			input:   "https://registry.example.com:invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseRegistryURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRegistryURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !registryInfoEqual(result, tt.expected) {
				t.Errorf("ParseRegistryURL() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestBuildRegistryURL(t *testing.T) {
	tests := []struct {
		name     string
		input    *types.RegistryInfo
		expected string
	}{
		{
			name: "HTTPS with custom port",
			input: &types.RegistryInfo{
				Scheme: "https",
				Host:   "registry.example.com",
				Port:   8080,
			},
			expected: "https://registry.example.com:8080",
		},
		{
			name: "HTTPS with default port",
			input: &types.RegistryInfo{
				Scheme: "https",
				Host:   "registry.example.com",
				Port:   443,
			},
			expected: "https://registry.example.com",
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildRegistryURL(tt.input)
			if result != tt.expected {
				t.Errorf("BuildRegistryURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateRegistryConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *types.RegistryConfig
		wantErr bool
	}{
		{
			name: "Valid config",
			config: &types.RegistryConfig{
				URL: "https://registry.example.com",
			},
		},
		{
			name:    "Nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "Empty URL",
			config: &types.RegistryConfig{
				URL: "",
			},
			wantErr: true,
		},
		{
			name: "Invalid retry policy",
			config: &types.RegistryConfig{
				URL: "https://registry.example.com",
				RetryPolicy: &types.RetryPolicy{
					MaxAttempts: 0, // Invalid
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegistryConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRegistryConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeRegistryURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "HTTPS URL with port",
			input:    "https://registry.example.com:8080/",
			expected: "https://registry.example.com:8080",
		},
		{
			name:     "HTTP URL with trailing slash",
			input:    "http://localhost:5000/",
			expected: "http://localhost:5000",
		},
		{
			name:     "URL without scheme defaults to HTTPS",
			input:    "registry.example.com:5000",
			expected: "https://registry.example.com:5000",
		},
		{
			name:     "Docker Hub registry",
			input:    "docker.io",
			expected: "https://docker.io",
		},
		{
			name:     "Registry with default HTTPS port",
			input:    "https://registry.example.com:443/",
			expected: "https://registry.example.com", // Default ports are normalized away
		},
		{
			name:     "Registry with default HTTP port",
			input:    "http://registry.example.com:80/",
			expected: "http://registry.example.com", // Default ports are normalized away
		},
		{
			name:     "Localhost without port",
			input:    "localhost",
			expected: "https://localhost",
		},
		{
			name:     "IP address with port",
			input:    "192.168.1.100:5000",
			expected: "https://192.168.1.100:5000",
		},
		{
			name:     "URL with multiple trailing slashes",
			input:    "https://registry.example.com:5000///",
			expected: "https://registry.example.com:5000",
		},
		{
			name:     "Already normalized URL",
			input:    "https://registry.example.com:5000",
			expected: "https://registry.example.com:5000",
		},
		{
			name:     "Invalid URL with missing scheme",
			input:    "://invalid-url",
			expected: "", // Returns empty string, no error
		},
		{
			name:    "Empty URL",
			input:   "",
			wantErr: true,
		},
		{
			name:     "URL with double dots (accepted)",
			input:    "https://registry..example.com",
			expected: "https://registry..example.com", // Actually accepted by the function
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeRegistryURL(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NormalizeRegistryURL() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NormalizeRegistryURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != tt.expected {
				t.Errorf("NormalizeRegistryURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNormalizeRegistryURL_Idempotency(t *testing.T) {
	// Test that normalizing an already normalized URL returns the same result
	testURLs := []string{
		"https://registry.example.com:5000",
		"http://localhost:8080",
		"https://docker.io",
	}

	for _, url := range testURLs {
		t.Run(url, func(t *testing.T) {
			// First normalization
			normalized1, err := NormalizeRegistryURL(url)
			if err != nil {
				t.Errorf("First normalization failed: %v", err)
				return
			}

			// Second normalization
			normalized2, err := NormalizeRegistryURL(normalized1)
			if err != nil {
				t.Errorf("Second normalization failed: %v", err)
				return
			}

			if normalized1 != normalized2 {
				t.Errorf("NormalizeRegistryURL() is not idempotent: first=%v, second=%v",
					normalized1, normalized2)
			}
		})
	}
}

func registryInfoEqual(a, b *types.RegistryInfo) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Scheme == b.Scheme && a.Host == b.Host && a.Port == b.Port
}