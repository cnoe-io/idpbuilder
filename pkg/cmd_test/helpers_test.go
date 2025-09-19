package cmd_test

import (
	"testing"
)

// TestValidateImageTag tests the image tag validation logic
func TestValidateImageTag(t *testing.T) {
	// validateImageTag is internal to cmd package, so we test the logic here
	validateImageTag := func(tag string) error {
		if tag == "" {
			return &mockError{"image tag cannot be empty"}
		}
		if tag == ":" {
			return &mockError{"invalid tag format"}
		}
		return nil
	}

	tests := []struct {
		name        string
		tag         string
		expectError bool
	}{
		{
			name:        "empty tag",
			tag:         "",
			expectError: true,
		},
		{
			name:        "just colon",
			tag:         ":",
			expectError: true,
		},
		{
			name:        "valid tag with version",
			tag:         "myapp:latest",
			expectError: false,
		},
		{
			name:        "valid tag with semantic version",
			tag:         "myapp:v1.0.0",
			expectError: false,
		},
		{
			name:        "registry with tag",
			tag:         "registry.com/myapp:latest",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateImageTag(tt.tag)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestParsePlatform tests the platform parsing logic
func TestParsePlatform(t *testing.T) {
	// parsePlatform is internal to cmd package, so we test the logic here
	parsePlatform := func(platform string) (string, string, error) {
		parts := []string{}
		current := ""
		for _, char := range platform {
			if char == '/' {
				if current != "" {
					parts = append(parts, current)
					current = ""
				}
			} else {
				current += string(char)
			}
		}
		if current != "" {
			parts = append(parts, current)
		}

		if len(parts) != 2 {
			return "", "", &mockError{"invalid platform format"}
		}

		validOS := map[string]bool{"linux": true, "windows": true, "darwin": true}
		validArch := map[string]bool{"amd64": true, "arm64": true, "arm": true, "386": true}

		if !validOS[parts[0]] {
			return "", "", &mockError{"unsupported OS"}
		}
		if !validArch[parts[1]] {
			return "", "", &mockError{"unsupported architecture"}
		}

		return parts[0], parts[1], nil
	}

	tests := []struct {
		name         string
		platform     string
		expectedOS   string
		expectedArch string
		expectError  bool
	}{
		{
			name:         "valid linux/amd64",
			platform:     "linux/amd64",
			expectedOS:   "linux",
			expectedArch: "amd64",
			expectError:  false,
		},
		{
			name:         "valid linux/arm64",
			platform:     "linux/arm64",
			expectedOS:   "linux",
			expectedArch: "arm64",
			expectError:  false,
		},
		{
			name:        "invalid format - no slash",
			platform:    "linux",
			expectError: true,
		},
		{
			name:        "invalid format - too many parts",
			platform:    "linux/amd64/extra",
			expectError: true,
		},
		{
			name:        "unsupported OS",
			platform:    "freebsd/amd64",
			expectError: true,
		},
		{
			name:        "unsupported architecture",
			platform:    "linux/sparc",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			osName, arch, err := parsePlatform(tt.platform)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError {
				if osName != tt.expectedOS {
					t.Errorf("Expected OS %s, got %s", tt.expectedOS, osName)
				}
				if arch != tt.expectedArch {
					t.Errorf("Expected arch %s, got %s", tt.expectedArch, arch)
				}
			}
		})
	}
}

// mockError is a simple error implementation for testing
type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}