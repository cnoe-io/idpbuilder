// pkg/daemon/daemon.go
package daemon

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

// DefaultDaemonClient implements DaemonClient using go-containerregistry
type DefaultDaemonClient struct {
	// dockerHost is the Docker daemon socket path (from DOCKER_HOST)
	dockerHost string
}

// NewDefaultClient creates a new daemon client.
// Respects DOCKER_HOST environment variable (REQ-024).
func NewDefaultClient() (*DefaultDaemonClient, error) {
	client := &DefaultDaemonClient{}

	// Check for custom Docker host (REQ-024)
	if host := os.Getenv("DOCKER_HOST"); host != "" {
		client.dockerHost = host
	}

	// Verify daemon connectivity
	if err := client.Ping(context.Background()); err != nil {
		return nil, err
	}

	return client, nil
}

// GetImage implements DaemonClient.GetImage
func (c *DefaultDaemonClient) GetImage(ctx context.Context, reference string) (*ImageInfo, ImageReader, error) {
	// Parse the reference
	ref, err := name.ParseReference(reference, name.WeakValidation)
	if err != nil {
		return nil, nil, &ImageNotFoundError{Reference: reference}
	}

	// Get image from daemon
	img, err := daemon.Image(ref)
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil, &ImageNotFoundError{Reference: reference}
		}
		if isDaemonUnavailable(err) {
			return nil, nil, &DaemonError{
				Message:      "Cannot connect to Docker daemon",
				IsNotRunning: true,
				Cause:        err,
			}
		}
		return nil, nil, &DaemonError{
			Message: "failed to get image: " + reference,
			Cause:   err,
		}
	}

	// Get image metadata
	digest, err := img.Digest()
	if err != nil {
		return nil, nil, &DaemonError{Message: "failed to get digest", Cause: err}
	}

	layers, err := img.Layers()
	if err != nil {
		return nil, nil, &DaemonError{Message: "failed to get layers", Cause: err}
	}

	// Calculate size
	var totalSize int64
	for _, layer := range layers {
		size, _ := layer.Size()
		totalSize += size
	}

	info := &ImageInfo{
		ID:         digest.String(),
		RepoTags:   []string{reference},
		Size:       totalSize,
		LayerCount: len(layers),
	}

	// Create pipe reader for image content
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		if err := tarball.Write(ref, img, pw); err != nil {
			pw.CloseWithError(err)
		}
	}()

	return info, &pipeReader{pr}, nil
}

// ImageExists implements DaemonClient.ImageExists
func (c *DefaultDaemonClient) ImageExists(ctx context.Context, reference string) (bool, error) {
	ref, err := name.ParseReference(reference, name.WeakValidation)
	if err != nil {
		return false, nil // Invalid reference means image doesn't exist
	}

	_, err = daemon.Image(ref)
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		if isDaemonUnavailable(err) {
			return false, &DaemonError{
				Message:      "Cannot connect to Docker daemon",
				IsNotRunning: true,
				Cause:        err,
			}
		}
		return false, &DaemonError{
			Message: "failed to check image existence",
			Cause:   err,
		}
	}

	return true, nil
}

// Ping implements DaemonClient.Ping
func (c *DefaultDaemonClient) Ping(ctx context.Context) error {
	// Try to access a non-existent image to verify daemon connectivity
	ref, _ := name.ParseReference("__ping_check__:__ping__", name.WeakValidation)
	_, err := daemon.Image(ref)

	if err == nil {
		// Unexpectedly found the image - daemon is running
		return nil
	}

	if isDaemonUnavailable(err) {
		return &DaemonError{
			Message:      "Cannot connect to Docker daemon",
			IsNotRunning: true,
			Cause:        err,
		}
	}

	// Any other error (like image not found) means daemon is running
	return nil
}

// pipeReader wraps io.PipeReader to implement ImageReader
type pipeReader struct {
	*io.PipeReader
}

func (r *pipeReader) Read(p []byte) (n int, err error) {
	return r.PipeReader.Read(p)
}

func (r *pipeReader) Close() error {
	return r.PipeReader.Close()
}

// isNotFoundError checks if error indicates image not found
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "no such image") ||
		strings.Contains(errStr, "manifest unknown")
}

// isDaemonUnavailable checks if error indicates daemon is not running
func isDaemonUnavailable(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "cannot connect to the docker daemon") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "is the docker daemon running") ||
		strings.Contains(errStr, "dial unix") ||
		strings.Contains(errStr, "no such host")
}
