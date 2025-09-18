package helpers

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cnoe-io/idpbuilder/pkg/registry/auth"
	"github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

// NewAuthenticatedClient creates an HTTP client with authentication configured
func NewAuthenticatedClient(authConfig *types.AuthConfig, options *types.ConnectionOptions) (*http.Client, error) {
	// Create authenticator
	authenticator, err := auth.NewAuthenticator(authConfig)
	if err != nil {
		return nil, err
	}

	// Start with default transport or provided one
	transport := http.DefaultTransport
	if options != nil && options.HTTPClient != nil && options.HTTPClient.Transport != nil {
		transport = options.HTTPClient.Transport
	}

	// Wrap with authentication transport
	authTransport := auth.NewTransport(transport, authenticator)

	// Create client
	client := &http.Client{
		Transport: authTransport,
		Timeout:   30 * time.Second, // Default timeout
	}

	// Apply options if provided
	if options != nil {
		if options.HTTPClient != nil {
			// Preserve timeout and other settings
			if options.HTTPClient.Timeout > 0 {
				client.Timeout = options.HTTPClient.Timeout
			}
			client.CheckRedirect = options.HTTPClient.CheckRedirect
			client.Jar = options.HTTPClient.Jar
		}

		// Apply TLS config if provided
		if options.TLSConfig != nil {
			if transport, ok := client.Transport.(*auth.Transport); ok {
				if httpTransport, ok := transport.Base.(*http.Transport); ok {
					httpTransport.TLSClientConfig = options.TLSConfig
				}
			}
		}
	}

	return client, nil
}

// NewRegistryClient creates an HTTP client configured for registry operations
func NewRegistryClient(config *types.RegistryConfig, authConfig *types.AuthConfig) (*http.Client, error) {
	// Validate config
	if err := ValidateRegistryConfig(config); err != nil {
		return nil, err
	}

	// Build connection options from registry config
	options := &types.ConnectionOptions{
		Headers: make(map[string]string),
	}

	// Set User-Agent
	options.UserAgent = "idpbuilder-registry-client/1.0"
	options.Headers["User-Agent"] = options.UserAgent

	// Set timeout from config
	if config.Timeout > 0 {
		options.HTTPClient = &http.Client{
			Timeout: config.Timeout,
		}
	}

	// Handle insecure and TLS skip verify options
	if config.Insecure || config.SkipTLSVerify {
		transport := &http.Transport{}

		if config.SkipTLSVerify {
			transport.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		if options.HTTPClient == nil {
			options.HTTPClient = &http.Client{}
		}
		options.HTTPClient.Transport = transport
	}

	return NewAuthenticatedClient(authConfig, options)
}

// CreateRequestWithAuth creates an HTTP request with authentication applied
func CreateRequestWithAuth(ctx context.Context, method, url string, body interface{}, authConfig *types.AuthConfig) (*http.Request, error) {
	// Create base request
	var req *http.Request
	var err error

	switch v := body.(type) {
	case nil:
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	case []byte:
		req, err = http.NewRequestWithContext(ctx, method, url, strings.NewReader(string(v)))
	case string:
		req, err = http.NewRequestWithContext(ctx, method, url, strings.NewReader(v))
	case io.Reader:
		req, err = http.NewRequestWithContext(ctx, method, url, v)
	default:
		return nil, fmt.Errorf("unsupported body type: %T", body)
	}

	if err != nil {
		return nil, err
	}

	// Apply authentication if provided
	if authConfig != nil {
		authenticator, err := auth.NewAuthenticator(authConfig)
		if err != nil {
			return nil, err
		}

		if err := authenticator.Authenticate(ctx, req); err != nil {
			return nil, err
		}
	}

	// Set common headers
	req.Header.Set("User-Agent", "idpbuilder-registry-client/1.0")
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json")

	return req, nil
}