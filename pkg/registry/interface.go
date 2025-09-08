package registry

import (
	"context"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// Registry defines operations for OCI registry interaction.
// This interface provides the core functionality needed for Gitea registry operations
// including image push, authentication, and repository listing.
type Registry interface {
	// Push uploads an image to the registry using the provided reference.
	// The reference should be in the format: registry.example.com/namespace/image:tag
	// Returns error if push fails due to authentication, network, or registry issues.
	Push(ctx context.Context, image v1.Image, reference string) error
	
	// Authenticate performs registry authentication using configured credentials.
	// Must be called before performing registry operations that require authentication.
	// Returns error if authentication fails or credentials are invalid.
	Authenticate(ctx context.Context) error
	
	// ListRepositories returns a list of available repository names in the registry.
	// Requires authentication to succeed. Returns empty slice if no repositories found.
	// Returns error if listing fails due to permissions or network issues.
	ListRepositories(ctx context.Context) ([]string, error)
	
	// GetRemoteOptions returns configured remote options for registry operations.
	// Includes TLS configuration, authentication, and other transport settings.
	GetRemoteOptions() []remote.Option
}

// RegistryConfig holds configuration for registry connection and authentication.
// All fields are required except Insecure which defaults to false.
type RegistryConfig struct {
	// URL is the base URL of the registry (e.g. https://gitea.example.com)
	URL string `json:"url" yaml:"url"`
	
	// Username for basic authentication
	Username string `json:"username" yaml:"username"`
	
	// Password for basic authentication
	Password string `json:"password" yaml:"password"`
	
	// Insecure allows TLS certificate verification to be skipped.
	// Should only be used for development or when explicitly required.
	Insecure bool `json:"insecure" yaml:"insecure"`
	
	// Timeout for registry operations in seconds (default: 30)
	TimeoutSeconds int `json:"timeout_seconds" yaml:"timeout_seconds"`
}

// AuthenticatedTransport wraps transport configuration for authenticated requests
type AuthenticatedTransport interface {
	// ConfigureTransport sets up the HTTP transport with proper authentication
	ConfigureTransport() error
	
	// IsAuthenticated returns true if authentication has been completed
	IsAuthenticated() bool
}