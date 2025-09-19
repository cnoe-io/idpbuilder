package types

import (
	"testing"
	"time"
)

func TestRegistryConfig(t *testing.T) {
	tests := []struct {
		name           string
		cfg            *RegistryConfig
		expectedURL    string
		expectedSecure bool
		hasTimeout     bool
		wantValid      bool
	}{
		{
			name: "valid secure config with namespace",
			cfg: &RegistryConfig{
				URL:       "registry.example.com",
				Namespace: "myorg",
				Timeout:   30 * time.Second,
			},
			expectedURL:    "registry.example.com",
			expectedSecure: true,
			hasTimeout:     true,
			wantValid:      true,
		},
		{
			name: "valid insecure config",
			cfg: &RegistryConfig{
				URL:      "localhost:5000",
				Insecure: true,
			},
			expectedURL:    "localhost:5000",
			expectedSecure: false,
			hasTimeout:     false,
			wantValid:      true,
		},
		{
			name: "config with skip TLS verification",
			cfg: &RegistryConfig{
				URL:           "self-signed.registry.com",
				SkipTLSVerify: true,
				Timeout:       5 * time.Second,
			},
			expectedURL:    "self-signed.registry.com",
			expectedSecure: true,
			hasTimeout:     true,
			wantValid:      true,
		},
		{
			name: "config with retry policy",
			cfg: &RegistryConfig{
				URL: "reliable.registry.com",
				RetryPolicy: &RetryPolicy{
					MaxAttempts:       3,
					InitialDelay:      100 * time.Millisecond,
					MaxDelay:          5 * time.Second,
					BackoffMultiplier: 2.0,
				},
			},
			expectedURL:    "reliable.registry.com",
			expectedSecure: true,
			hasTimeout:     false,
			wantValid:      true,
		},
		{
			name: "empty URL - edge case",
			cfg: &RegistryConfig{
				URL:       "",
				Namespace: "myorg",
			},
			expectedURL:    "",
			expectedSecure: true,
			hasTimeout:     false,
			wantValid:      false,
		},
		{
			name:      "nil config - edge case",
			cfg:       nil,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cfg == nil {
				if tt.wantValid {
					t.Error("nil config should not be valid")
				}
				return
			}

			// Validate URL field
			if tt.cfg.URL != tt.expectedURL {
				t.Errorf("URL = %q, want %q", tt.cfg.URL, tt.expectedURL)
			}

			// Validate security settings
			isSecure := !tt.cfg.Insecure
			if isSecure != tt.expectedSecure {
				t.Errorf("secure mode = %v, want %v", isSecure, tt.expectedSecure)
			}

			// Validate timeout configuration
			hasTimeout := tt.cfg.Timeout > 0
			if hasTimeout != tt.hasTimeout {
				t.Errorf("has timeout = %v, want %v (timeout: %v)", hasTimeout, tt.hasTimeout, tt.cfg.Timeout)
			}

			// Validate overall configuration validity
			isValid := tt.cfg.URL != ""
			if isValid != tt.wantValid {
				t.Errorf("config validity = %v, want %v", isValid, tt.wantValid)
			}

			// Additional validations for non-nil configs
			if tt.cfg.RetryPolicy != nil {
				if tt.cfg.RetryPolicy.MaxAttempts <= 0 {
					t.Error("retry policy should have positive max attempts")
				}
				if tt.cfg.RetryPolicy.BackoffMultiplier <= 0 {
					t.Error("retry policy should have positive backoff multiplier")
				}
			}
		})
	}
}

func TestRetryPolicy(t *testing.T) {
	tests := []struct {
		name   string
		policy *RetryPolicy
		valid  bool
	}{
		{
			name: "valid retry policy",
			policy: &RetryPolicy{
				MaxAttempts:       3,
				InitialDelay:      100 * time.Millisecond,
				MaxDelay:          5 * time.Second,
				BackoffMultiplier: 2.0,
			},
			valid: true,
		},
		{
			name: "invalid retry policy - zero attempts",
			policy: &RetryPolicy{
				MaxAttempts:       0,
				InitialDelay:      100 * time.Millisecond,
				BackoffMultiplier: 2.0,
			},
			valid: false,
		},
		{
			name: "invalid retry policy - negative multiplier",
			policy: &RetryPolicy{
				MaxAttempts:       2,
				InitialDelay:      100 * time.Millisecond,
				BackoffMultiplier: -1.0,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.policy.MaxAttempts > 0 && tt.policy.BackoffMultiplier > 0
			if valid != tt.valid {
				t.Errorf("retry policy validity = %v, want %v", valid, tt.valid)
			}
		})
	}
}

func TestRegistryInfo(t *testing.T) {
	info := &RegistryInfo{
		Scheme:       "https",
		Host:         "registry.example.com",
		Port:         443,
		APIVersion:   "v2",
		Capabilities: []string{CapabilityPush, CapabilityPull},
	}

	if info.Scheme != "https" {
		t.Errorf("Scheme = %q, want https", info.Scheme)
	}
	if info.Host != "registry.example.com" {
		t.Errorf("Host = %q, want registry.example.com", info.Host)
	}
	if info.Port != 443 {
		t.Errorf("Port = %d, want 443", info.Port)
	}
	if len(info.Capabilities) != 2 {
		t.Errorf("Capabilities length = %d, want 2", len(info.Capabilities))
	}

	// Test capability checking
	hasPush := false
	hasPull := false
	for _, cap := range info.Capabilities {
		if cap == CapabilityPush {
			hasPush = true
		}
		if cap == CapabilityPull {
			hasPull = true
		}
	}
	if !hasPush {
		t.Error("expected push capability")
	}
	if !hasPull {
		t.Error("expected pull capability")
	}
}

func TestImageReference(t *testing.T) {
	tests := []struct {
		name      string
		ref       *ImageReference
		wantValid bool
	}{
		{
			name: "complete reference with tag",
			ref: &ImageReference{
				Registry:   "registry.example.com",
				Namespace:  "myorg",
				Repository: "myapp",
				Tag:        "v1.0.0",
			},
			wantValid: true,
		},
		{
			name: "reference with digest",
			ref: &ImageReference{
				Registry:   "registry.example.com",
				Repository: "myapp",
				Digest:     "sha256:abcd1234",
			},
			wantValid: true,
		},
		{
			name: "incomplete reference - missing registry",
			ref: &ImageReference{
				Repository: "myapp",
				Tag:        "latest",
			},
			wantValid: false,
		},
		{
			name: "incomplete reference - missing repository",
			ref: &ImageReference{
				Registry:  "registry.example.com",
				Namespace: "myorg",
				Tag:       "latest",
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation: must have registry and repository
			isValid := tt.ref.Registry != "" && tt.ref.Repository != ""
			if isValid != tt.wantValid {
				t.Errorf("reference validity = %v, want %v", isValid, tt.wantValid)
			}

			// Additional field validations for valid references
			if tt.wantValid {
				if tt.ref.Registry == "" {
					t.Error("valid reference should have registry")
				}
				if tt.ref.Repository == "" {
					t.Error("valid reference should have repository")
				}
				// Should have either tag or digest
				hasIdentifier := tt.ref.Tag != "" || tt.ref.Digest != ""
				if !hasIdentifier {
					t.Error("valid reference should have tag or digest")
				}
			}
		})
	}
}
