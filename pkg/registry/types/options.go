package types

import (
	"crypto/tls"
	"io"
	"net/http"
)

// ConnectionOptions defines options for registry connections
type ConnectionOptions struct {
	// TLSConfig for secure connections
	TLSConfig *tls.Config

	// HTTPClient to use for requests
	HTTPClient *http.Client

	// Headers to add to all requests
	Headers map[string]string

	// UserAgent string
	UserAgent string

	// Debug enables debug logging
	Debug bool
}

// PushOptions defines options for pushing images
type PushOptions struct {
	// ProgressWriter for progress updates
	ProgressWriter io.Writer

	// Layers to push in parallel
	ParallelLayers int

	// Force push even if image exists
	Force bool

	// Platform specific push
	Platform string
}

// PullOptions defines options for pulling images
type PullOptions struct {
	// ProgressWriter for progress updates
	ProgressWriter io.Writer

	// VerifySignature checks image signatures
	VerifySignature bool

	// Platform specific pull
	Platform string
}

// ListOptions defines options for listing repositories/tags
type ListOptions struct {
	// Limit number of results
	Limit int

	// Offset for pagination
	Offset int

	// Filter by pattern
	Filter string
}