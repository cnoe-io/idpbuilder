// pkg/daemon/client.go
package daemon

import (
	"context"
	"io"
)

// ImageInfo contains metadata about a local Docker image.
type ImageInfo struct {
	// ID is the image's unique identifier (digest)
	ID string
	// RepoTags are the image's repository tags (e.g., ["myapp:latest", "myapp:v1.0"])
	RepoTags []string
	// Size is the total image size in bytes
	Size int64
	// LayerCount is the number of layers in the image
	LayerCount int
}

// DaemonClient defines operations for interacting with the local Docker daemon.
// All operations work with the daemon's image store.
type DaemonClient interface {
	// GetImage retrieves an image from the local Docker daemon.
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - reference: Image reference (e.g., "myapp:latest", "sha256:abc...")
	// Returns:
	//   - ImageInfo with metadata about the image
	//   - ImageReader for accessing image content (caller must close)
	//   - Error if image not found or daemon unavailable
	GetImage(ctx context.Context, reference string) (*ImageInfo, ImageReader, error)

	// ImageExists checks if an image exists in the local Docker daemon.
	// This is a lighter-weight check than GetImage when you only need presence.
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - reference: Image reference to check
	// Returns:
	//   - true if image exists, false otherwise
	//   - Error only if daemon communication fails (not for missing images)
	ImageExists(ctx context.Context, reference string) (bool, error)

	// Ping checks connectivity to the Docker daemon.
	// Returns error if daemon is not available.
	Ping(ctx context.Context) error
}

// ImageReader provides access to image content for push operations.
// Callers must call Close() when done to release resources.
type ImageReader interface {
	io.ReadCloser
}

// DaemonError represents an error from the Docker daemon.
type DaemonError struct {
	// Message is a human-readable error description
	Message string
	// IsNotRunning indicates the daemon is not available
	IsNotRunning bool
	// Cause is the underlying error
	Cause error
}

// Error implements the error interface.
func (e *DaemonError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Unwrap implements errors.Unwrap for error chaining.
func (e *DaemonError) Unwrap() error {
	return e.Cause
}

// ImageNotFoundError is returned when a requested image doesn't exist locally.
type ImageNotFoundError struct {
	// Reference is the image reference that was not found
	Reference string
}

// Error implements the error interface.
func (e *ImageNotFoundError) Error() string {
	return "image not found: " + e.Reference
}
