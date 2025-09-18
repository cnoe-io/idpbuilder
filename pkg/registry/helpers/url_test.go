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

func registryInfoEqual(a, b *types.RegistryInfo) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Scheme == b.Scheme && a.Host == b.Host && a.Port == b.Port
}