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