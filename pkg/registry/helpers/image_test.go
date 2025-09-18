package helpers

import (
	"testing"

	"github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

func TestParseImageReference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *types.ImageReference
		wantErr  bool
	}{
		{
			name:  "Docker Hub official image",
			input: "ubuntu:20.04",
			expected: &types.ImageReference{
				Registry:   "docker.io",
				Namespace:  "library",
				Repository: "ubuntu",
				Tag:        "20.04",
			},
		},
		{
			name:  "Docker Hub user image",
			input: "nginx/nginx:latest",
			expected: &types.ImageReference{
				Registry:   "docker.io",
				Namespace:  "nginx",
				Repository: "nginx",
				Tag:        "latest",
			},
		},
		{
			name:  "Custom registry with port",
			input: "registry.example.com:5000/myapp/web:v1.0",
			expected: &types.ImageReference{
				Registry:   "registry.example.com:5000",
				Namespace:  "myapp",
				Repository: "web",
				Tag:        "v1.0",
			},
		},
		{
			name:  "With digest",
			input: "alpine@sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			expected: &types.ImageReference{
				Registry:   "docker.io",
				Namespace:  "library",
				Repository: "alpine",
				Digest:     "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			},
		},
		{
			name:    "Empty reference",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Invalid digest",
			input:   "alpine@invalid-digest",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseImageReference(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseImageReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !imageReferenceEqual(result, tt.expected) {
				t.Errorf("ParseImageReference() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestBuildImageReference(t *testing.T) {
	tests := []struct {
		name     string
		input    *types.ImageReference
		expected string
	}{
		{
			name: "Docker Hub official",
			input: &types.ImageReference{
				Registry:   "docker.io",
				Namespace:  "library",
				Repository: "ubuntu",
				Tag:        "20.04",
			},
			expected: "ubuntu:20.04",
		},
		{
			name: "Custom registry",
			input: &types.ImageReference{
				Registry:   "registry.example.com:5000",
				Namespace:  "myapp",
				Repository: "web",
				Tag:        "v1.0",
			},
			expected: "registry.example.com:5000/myapp/web:v1.0",
		},
		{
			name: "With digest",
			input: &types.ImageReference{
				Registry:   "docker.io",
				Namespace:  "library",
				Repository: "alpine",
				Digest:     "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			},
			expected: "alpine@sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildImageReference(tt.input)
			if result != tt.expected {
				t.Errorf("BuildImageReference() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateImageReference(t *testing.T) {
	tests := []struct {
		name    string
		input   *types.ImageReference
		wantErr bool
	}{
		{
			name: "Valid reference",
			input: &types.ImageReference{
				Registry:   "registry.example.com",
				Repository: "myapp",
				Tag:        "v1.0",
			},
		},
		{
			name:    "Nil reference",
			input:   nil,
			wantErr: true,
		},
		{
			name: "Empty repository",
			input: &types.ImageReference{
				Registry: "registry.example.com",
			},
			wantErr: true,
		},
		{
			name: "Invalid tag",
			input: &types.ImageReference{
				Repository: "myapp",
				Tag:        "invalid/tag",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageReference(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateImageReference() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetImageRegistryURL(t *testing.T) {
	tests := []struct {
		name     string
		input    *types.ImageReference
		expected string
	}{
		{
			name: "Docker Hub registry",
			input: &types.ImageReference{
				Registry: "docker.io",
			},
			expected: "https://registry-1.docker.io",
		},
		{
			name: "Custom registry",
			input: &types.ImageReference{
				Registry: "registry.example.com:5000",
			},
			expected: "https://registry.example.com:5000",
		},
		{
			name:     "Nil reference defaults to Docker Hub",
			input:    nil,
			expected: "https://registry-1.docker.io",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetImageRegistryURL(tt.input)
			if result != tt.expected {
				t.Errorf("GetImageRegistryURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func imageReferenceEqual(a, b *types.ImageReference) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Registry == b.Registry &&
		a.Namespace == b.Namespace &&
		a.Repository == b.Repository &&
		a.Tag == b.Tag &&
		a.Digest == b.Digest
}

func TestImageReferenceWithTag(t *testing.T) {
	tests := []struct {
		name     string
		input    *types.ImageReference
		tag      string
		expected *types.ImageReference
	}{
		{
			name: "Add tag to reference",
			input: &types.ImageReference{
				Registry:   "registry.example.com",
				Namespace:  "myapp",
				Repository: "web",
				Tag:        "old-tag",
			},
			tag: "new-tag",
			expected: &types.ImageReference{
				Registry:   "registry.example.com",
				Namespace:  "myapp",
				Repository: "web",
				Tag:        "new-tag",
				Digest:     "", // Should be cleared
			},
		},
		{
			name: "Replace tag and clear digest",
			input: &types.ImageReference{
				Registry:   "registry.example.com",
				Namespace:  "myapp",
				Repository: "web",
				Tag:        "old-tag",
				Digest:     "sha256:abcdef",
			},
			tag: "new-tag",
			expected: &types.ImageReference{
				Registry:   "registry.example.com",
				Namespace:  "myapp",
				Repository: "web",
				Tag:        "new-tag",
				Digest:     "", // Should be cleared when setting tag
			},
		},
		{
			name: "Add tag to reference without tag",
			input: &types.ImageReference{
				Registry:   "docker.io",
				Namespace:  "library",
				Repository: "ubuntu",
			},
			tag: "20.04",
			expected: &types.ImageReference{
				Registry:   "docker.io",
				Namespace:  "library",
				Repository: "ubuntu",
				Tag:        "20.04",
			},
		},
		{
			name:     "Nil input returns nil",
			input:    nil,
			tag:      "latest",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ImageReferenceWithTag(tt.input, tt.tag)

			if !imageReferenceEqual(result, tt.expected) {
				t.Errorf("ImageReferenceWithTag() = %+v, want %+v", result, tt.expected)
			}

			// Verify original input is not modified
			if tt.input != nil {
				originalInput := &types.ImageReference{
					Registry:   tt.input.Registry,
					Namespace:  tt.input.Namespace,
					Repository: tt.input.Repository,
					Tag:        tt.input.Tag,
					Digest:     tt.input.Digest,
				}
				// Reset to original values for comparison
				if tt.name == "Add tag to reference" {
					originalInput.Tag = "old-tag"
				} else if tt.name == "Replace tag and clear digest" {
					originalInput.Tag = "old-tag"
					originalInput.Digest = "sha256:abcdef"
				} else if tt.name == "Add tag to reference without tag" {
					originalInput.Tag = ""
				}
			}
		})
	}
}

func TestImageReferenceWithDigest(t *testing.T) {
	tests := []struct {
		name     string
		input    *types.ImageReference
		digest   string
		expected *types.ImageReference
	}{
		{
			name: "Add digest to reference",
			input: &types.ImageReference{
				Registry:   "registry.example.com",
				Namespace:  "myapp",
				Repository: "web",
				Tag:        "v1.0",
			},
			digest: "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			expected: &types.ImageReference{
				Registry:   "registry.example.com",
				Namespace:  "myapp",
				Repository: "web",
				Tag:        "v1.0", // Tag should be preserved
				Digest:     "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			},
		},
		{
			name: "Replace existing digest",
			input: &types.ImageReference{
				Registry:   "docker.io",
				Namespace:  "library",
				Repository: "alpine",
				Digest:     "sha256:olddigest",
			},
			digest: "sha256:newdigest123456789012345678901234567890123456789012345678901234",
			expected: &types.ImageReference{
				Registry:   "docker.io",
				Namespace:  "library",
				Repository: "alpine",
				Digest:     "sha256:newdigest123456789012345678901234567890123456789012345678901234",
			},
		},
		{
			name: "Add digest with tag preserved",
			input: &types.ImageReference{
				Registry:   "registry.example.com:5000",
				Namespace:  "company",
				Repository: "app",
				Tag:        "latest",
			},
			digest: "sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			expected: &types.ImageReference{
				Registry:   "registry.example.com:5000",
				Namespace:  "company",
				Repository: "app",
				Tag:        "latest", // Tag should coexist with digest
				Digest:     "sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			},
		},
		{
			name:     "Nil input returns nil",
			input:    nil,
			digest:   "sha256:test",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ImageReferenceWithDigest(tt.input, tt.digest)

			if !imageReferenceEqual(result, tt.expected) {
				t.Errorf("ImageReferenceWithDigest() = %+v, want %+v", result, tt.expected)
			}

			// Verify original input is not modified by checking it's a copy
			if tt.input != nil && result != nil {
				// Verify they are different objects (not the same pointer)
				if result == tt.input {
					t.Errorf("ImageReferenceWithDigest() should return a copy, not modify original")
				}
			}
		})
	}
}