package registry

import (
	"context"
	"io"
)

// Registry defines the interface for container registry operations
type Registry interface {
	// Push uploads a container image to the registry
	Push(ctx context.Context, image string, content io.Reader) error
	
	// List returns a list of repositories in the registry
	List(ctx context.Context) ([]string, error)
	
	// Exists checks if a repository exists in the registry
	Exists(ctx context.Context, repository string) (bool, error)
	
	// Delete removes a repository from the registry
	Delete(ctx context.Context, repository string) error
	
	// Close cleans up any resources used by the registry client
	Close() error
}

// RegistryConfig holds configuration for registry connections
type RegistryConfig struct {
	URL      string
	Username string
	Token    string
	Insecure bool
}