package tls

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

// Config holds TLS configuration options for secure communication
// with registry services, including the ability to skip certificate
// verification for self-signed certificates in development environments.
type Config struct {
	// InsecureSkipVerify controls whether the client verifies the
	// server's certificate chain and host name. If InsecureSkipVerify
	// is true, TLS accepts any certificate presented by the server
	// and any host name in that certificate.
	// This should only be used for testing or development environments
	// with self-signed certificates.
	InsecureSkipVerify bool
}

// NewConfig creates a new TLS configuration with the specified options.
// By default, certificate verification is enabled for security.
//
// Parameters:
//   - insecure: when true, disables certificate verification (use only for development)
//
// Returns:
//   - *Config: configured TLS configuration object
func NewConfig(insecure bool) *Config {
	return &Config{
		InsecureSkipVerify: insecure,
	}
}

// ToTLSConfig converts the Config to a standard crypto/tls.Config
// that can be used with Go's standard library TLS implementations.
//
// Returns:
//   - *tls.Config: standard TLS configuration ready for use
func (c *Config) ToTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: c.InsecureSkipVerify,
	}
}

// ApplyToHTTPClient applies the TLS configuration to an HTTP client.
// If the client doesn't have a transport configured, a new one is created.
// This method ensures the client uses the specified TLS settings for HTTPS requests.
//
// Parameters:
//   - client: HTTP client to configure
func (c *Config) ApplyToHTTPClient(client *http.Client) {
	// Ensure the client has a transport
	if client.Transport == nil {
		client.Transport = &http.Transport{}
	}

	// Apply TLS config if the transport is the expected type
	if transport, ok := client.Transport.(*http.Transport); ok {
		transport.TLSClientConfig = c.ToTLSConfig()

		// Log warning for insecure configurations
		if c.InsecureSkipVerify {
			fmt.Printf("⚠️  TLS certificate verification disabled for HTTP client\n")
		}
	}
}

// ApplyToTransport applies the TLS configuration to an HTTP transport directly.
// This is useful when you have direct access to the transport object.
//
// Parameters:
//   - transport: HTTP transport to configure
func (c *Config) ApplyToTransport(transport *http.Transport) {
	transport.TLSClientConfig = c.ToTLSConfig()

	// Log warning for insecure configurations
	if c.InsecureSkipVerify {
		fmt.Printf("⚠️  TLS certificate verification disabled for HTTP transport\n")
	}
}

// IsSecure returns true if the configuration uses secure TLS settings
// (certificate verification enabled).
//
// Returns:
//   - bool: true if secure, false if insecure mode is enabled
func (c *Config) IsSecure() bool {
	return !c.InsecureSkipVerify
}

// String returns a string representation of the TLS configuration
// for logging and debugging purposes.
//
// Returns:
//   - string: human-readable description of the configuration
func (c *Config) String() string {
	if c.InsecureSkipVerify {
		return "TLS Config: INSECURE (certificate verification disabled)"
	}
	return "TLS Config: SECURE (certificate verification enabled)"
}