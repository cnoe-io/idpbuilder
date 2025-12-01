// Package registry provides interfaces and types for pushing OCI images to registries.
package registry

import (
	"context"
	"io"
)

// PushResult contains information about a successful push operation.
type PushResult struct {
	// Reference is the full registry reference of the pushed image
	// Example: "registry.example.com/myapp:v1.0.0@sha256:abc..."
	Reference string
	// Digest is the content-addressable digest of the pushed manifest
	Digest string
	// Size is the total size of all layers pushed (in bytes)
	Size int64
}

// RegistryConfig holds configuration for connecting to an OCI registry.
type RegistryConfig struct {
	// URL is the registry URL (e.g., "registry.example.com" or "localhost:5000")
	URL string
	// Insecure allows connecting to HTTP registries or registries with invalid TLS
	Insecure bool
	// Username for basic authentication (mutually exclusive with Token)
	Username string
	// Password for basic authentication (mutually exclusive with Token)
	Password string
	// Token for bearer token authentication (mutually exclusive with Username/Password)
	Token string
}

// RegistryClient defines operations for pushing images to an OCI-compliant registry.
// Implementations handle authentication, layer upload, and manifest push.
type RegistryClient interface {
	// Push pushes an image from the local Docker daemon to the registry.
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - imageRef: Local image reference (e.g., "myapp:latest")
	//   - destRef: Destination reference (e.g., "registry.example.com/myapp:latest")
	//   - progress: Progress reporter for push status updates (can be nil)
	// Returns:
	//   - PushResult with reference and digest on success
	//   - Error if push fails (may be RegistryError, AuthError, or NetworkError)
	Push(ctx context.Context, imageRef, destRef string, progress ProgressReporter) (*PushResult, error)
}

// RegistryClientFactory creates RegistryClient instances with the given configuration.
type RegistryClientFactory interface {
	// NewClient creates a new RegistryClient with the provided configuration.
	NewClient(config RegistryConfig) (RegistryClient, error)
}

// RegistryError represents an error from the registry with classification.
type RegistryError struct {
	// StatusCode is the HTTP status code from the registry (0 if not HTTP)
	StatusCode int
	// Message is a human-readable error description
	Message string
	// IsTransient indicates if the error may be resolved by retry
	IsTransient bool
	// Cause is the underlying error
	Cause error
}

// Error implements the error interface.
func (e *RegistryError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Unwrap implements errors.Unwrap for error chaining.
func (e *RegistryError) Unwrap() error {
	return e.Cause
}

// AuthError represents an authentication failure.
type AuthError struct {
	Message string
	Cause   error
}

// Error implements the error interface.
func (e *AuthError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Unwrap implements errors.Unwrap for error chaining.
func (e *AuthError) Unwrap() error {
	return e.Cause
}

// ProgressReporter receives progress updates during push operations.
type ProgressReporter interface {
	// Start is called when the push operation begins.
	Start(imageRef string, totalLayers int)
	// LayerProgress is called during layer upload.
	// current is bytes uploaded, total is layer size.
	LayerProgress(layerDigest string, current, total int64)
	// LayerComplete is called when a layer finishes uploading.
	LayerComplete(layerDigest string)
	// Complete is called when the entire push succeeds.
	Complete(result *PushResult)
	// Error is called when the push fails.
	Error(err error)
}

// NoOpProgressReporter is a ProgressReporter that does nothing.
// Used when progress reporting is disabled or for testing.
type NoOpProgressReporter struct{}

func (n *NoOpProgressReporter) Start(imageRef string, totalLayers int)              {}
func (n *NoOpProgressReporter) LayerProgress(layerDigest string, current, total int64) {}
func (n *NoOpProgressReporter) LayerComplete(layerDigest string)                     {}
func (n *NoOpProgressReporter) Complete(result *PushResult)                          {}
func (n *NoOpProgressReporter) Error(err error)                                      {}

// StderrProgressReporter writes progress to stderr.
// This is the default progress reporter for user-facing operations.
type StderrProgressReporter struct {
	Out io.Writer
}

func (s *StderrProgressReporter) Start(imageRef string, totalLayers int) {
	// Implementation in Wave 3 (E1.3.2)
}

func (s *StderrProgressReporter) LayerProgress(layerDigest string, current, total int64) {
	// Implementation in Wave 3 (E1.3.2)
}

func (s *StderrProgressReporter) LayerComplete(layerDigest string) {
	// Implementation in Wave 3 (E1.3.2)
}

func (s *StderrProgressReporter) Complete(result *PushResult) {
	// Implementation in Wave 3 (E1.3.2)
}

func (s *StderrProgressReporter) Error(err error) {
	// Implementation in Wave 3 (E1.3.2)
}
