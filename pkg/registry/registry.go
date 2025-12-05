package registry

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// DefaultClient implements RegistryClient using go-containerregistry
type DefaultClient struct {
	config     RegistryConfig
	auth       authn.Authenticator
	httpClient *http.Client
}

// NewDefaultClient creates a new registry client with the given configuration
func NewDefaultClient(config RegistryConfig) (*DefaultClient, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("registry URL is required")
	}

	// Configure authentication based on config
	var auth authn.Authenticator
	if config.Token != "" {
		auth = &authn.Bearer{Token: config.Token}
	} else if config.Username != "" && config.Password != "" {
		auth = &authn.Basic{Username: config.Username, Password: config.Password}
	} else {
		auth = authn.Anonymous
	}

	// Configure HTTP client with optional TLS insecure mode
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.Insecure,
			},
			DisableKeepAlives: false,
			MaxIdleConns:      10,
		},
	}

	return &DefaultClient{
		config:     config,
		auth:       auth,
		httpClient: httpClient,
	}, nil
}

// Push implements RegistryClient.Push using go-containerregistry
func (c *DefaultClient) Push(ctx context.Context, imageRef, destRef string, progress ProgressReporter) (*PushResult, error) {
	// Parse source reference for daemon access
	sourceRef, err := name.ParseReference(imageRef, name.WeakValidation)
	if err != nil {
		if progress != nil {
			progress.Error(err)
		}
		return nil, fmt.Errorf("invalid source reference: %w", err)
	}

	// Parse destination reference for registry
	destRefParsed, err := name.ParseReference(destRef, name.WeakValidation)
	if err != nil {
		if progress != nil {
			progress.Error(err)
		}
		return nil, fmt.Errorf("invalid destination reference: %w", err)
	}

	// Get image from local daemon
	img, err := daemon.Image(sourceRef)
	if err != nil {
		if progress != nil {
			progress.Error(err)
		}
		return nil, fmt.Errorf("failed to read image from daemon: %w", err)
	}

	// Get layer count for progress reporting
	layers, err := img.Layers()
	if err != nil {
		if progress != nil {
			progress.Error(err)
		}
		return nil, fmt.Errorf("failed to get image layers: %w", err)
	}

	// Call progress.Start() if progress is provided
	if progress != nil {
		progress.Start(imageRef, len(layers))
	}

	// Build remote options with auth, context, and optional TLS
	opts := []remote.Option{
		remote.WithAuth(c.auth),
		remote.WithContext(ctx),
	}

	// Add custom HTTP client for TLS configuration
	if c.config.Insecure {
		opts = append(opts, remote.WithTransport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DisableKeepAlives: false,
			MaxIdleConns:      10,
		}))
	}

	// Push image to registry
	err = remote.Write(destRefParsed, img, opts...)
	if err != nil {
		err = classifyRemoteError(err)
		if progress != nil {
			progress.Error(err)
		}
		return nil, err
	}

	// Get digest from the image configuration
	configDigest, err := img.ConfigName()
	if err != nil {
		if progress != nil {
			progress.Error(err)
		}
		return nil, fmt.Errorf("failed to get image digest: %w", err)
	}

	// Calculate total size from layers
	totalSize := int64(0)
	for _, layer := range layers {
		size, err := layer.Size()
		if err == nil {
			totalSize += size
		}
	}

	// Build and return PushResult
	result := &PushResult{
		Reference: destRef + "@" + configDigest.String(),
		Digest:    configDigest.String(),
		Size:      totalSize,
	}

	// Call progress.Complete() if progress is provided
	if progress != nil {
		progress.Complete(result)
	}

	return result, nil
}

// classifyRemoteError converts go-containerregistry errors to our error types
func classifyRemoteError(err error) error {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())

	// Check for auth errors (401, 403, unauthorized, forbidden)
	if strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "forbidden") ||
		strings.Contains(errStr, "authentication") {
		return &AuthError{
			Message: "authentication failed",
			Cause:   err,
		}
	}

	// Check for transient errors (5xx, timeout, connection refused)
	isTransient := false

	// Check for 5xx status codes
	if strings.Contains(errStr, "50") || strings.Contains(errStr, "51") ||
		strings.Contains(errStr, "52") || strings.Contains(errStr, "53") {
		isTransient = true
	}

	// Check for timeout-like errors
	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded") ||
		strings.Contains(errStr, "temporary failure") {
		isTransient = true
	}

	// Check for connection-related errors
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "network unreachable") ||
		strings.Contains(errStr, "host unreachable") {
		isTransient = true
	}

	// Extract HTTP status code if present
	statusCode := 0
	if sc, ok := extractStatusCode(err); ok {
		statusCode = sc
		if sc >= 500 {
			isTransient = true
		}
	}

	// Check for net.Error temporary flag
	if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
		isTransient = true
	}

	return &RegistryError{
		StatusCode:  statusCode,
		Message:     "registry operation failed",
		IsTransient: isTransient,
		Cause:       err,
	}
}

// extractStatusCode attempts to extract HTTP status code from error
func extractStatusCode(err error) (int, bool) {
	errStr := err.Error()

	// Look for status code patterns like "status 500" or "500 Internal Server Error"
	parts := strings.Fields(errStr)
	for _, part := range parts {
		// Remove non-numeric characters (like colons, commas)
		cleanPart := strings.TrimFunc(part, func(r rune) bool {
			return !('0' <= r && r <= '9')
		})
		if cleanPart != "" {
			if sc, err := strconv.Atoi(cleanPart); err == nil && sc >= 100 && sc < 600 {
				return sc, true
			}
		}
	}

	return 0, false
}
