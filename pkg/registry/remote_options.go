package registry

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// RemoteOptions configures the behavior of remote registry operations
type RemoteOptions struct {
	// Connection settings
	Timeout      time.Duration
	MaxRetries   int
	RetryDelay   time.Duration
	UserAgent    string
	
	// TLS/SSL settings
	Insecure           bool
	TLSConfig          *tls.Config
	CertFile           string
	KeyFile            string
	CAFile             string
	SkipTLSVerify      bool
	
	// Proxy settings
	ProxyURL           string
	ProxyUsername      string
	ProxyPassword      string
	NoProxy            []string
	
	// Authentication settings
	AuthScope          string
	AuthService        string
	TokenRefreshMargin time.Duration
	
	// Performance settings
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int
	IdleConnTimeout     time.Duration
	
	// Registry-specific settings
	RegistryVersion     string
	AcceptedMediaTypes  []string
	CustomHeaders       map[string]string
}

// DefaultRemoteOptions returns a RemoteOptions struct with sensible defaults
func DefaultRemoteOptions() *RemoteOptions {
	return &RemoteOptions{
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		RetryDelay:   1 * time.Second,
		UserAgent:    "idpbuilder-gitea-client/1.0",
		Insecure:      false,
		SkipTLSVerify: false,
		TokenRefreshMargin: 5 * time.Minute,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     0,
		IdleConnTimeout:     90 * time.Second,
		RegistryVersion: "v2",
		AcceptedMediaTypes: []string{
			"application/vnd.docker.distribution.manifest.v2+json",
			"application/vnd.docker.distribution.manifest.list.v2+json",
			"application/vnd.oci.image.manifest.v1+json",
			"application/vnd.oci.image.index.v1+json",
		},
		CustomHeaders: make(map[string]string),
	}
}

// Validate checks if the RemoteOptions configuration is valid
func (r *RemoteOptions) Validate() error {
	if r.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	if r.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	if r.RetryDelay < 0 {
		return fmt.Errorf("retry delay cannot be negative")
	}
	if r.UserAgent == "" {
		return fmt.Errorf("user agent cannot be empty")
	}
	if r.ProxyURL != "" {
		if _, err := url.Parse(r.ProxyURL); err != nil {
			return fmt.Errorf("invalid proxy URL: %w", err)
		}
	}
	if r.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections cannot be negative")
	}
	if r.MaxIdleConnsPerHost < 0 {
		return fmt.Errorf("max idle connections per host cannot be negative")
	}
	if r.IdleConnTimeout < 0 {
		return fmt.Errorf("idle connection timeout cannot be negative")
	}
	if r.TokenRefreshMargin < 0 {
		return fmt.Errorf("token refresh margin cannot be negative")
	}
	return nil
}

// ApplyToTransport configures an HTTP transport with these options
func (r *RemoteOptions) ApplyToTransport(transport *http.Transport) error {
	if transport == nil {
		return fmt.Errorf("transport cannot be nil")
	}
	
	// Apply TLS configuration
	if r.TLSConfig != nil {
		transport.TLSClientConfig = r.TLSConfig.Clone()
	} else {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: r.Insecure || r.SkipTLSVerify,
		}
	}
	
	// Load client certificates if specified
	if r.CertFile != "" && r.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(r.CertFile, r.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to load client certificate: %w", err)
		}
		
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.Certificates = []tls.Certificate{cert}
	}
	
	// Configure proxy
	if r.ProxyURL != "" {
		proxyURL, err := url.Parse(r.ProxyURL)
		if err != nil {
			return fmt.Errorf("invalid proxy URL: %w", err)
		}
		
		if r.ProxyUsername != "" {
			proxyURL.User = url.UserPassword(r.ProxyUsername, r.ProxyPassword)
		}
		
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	
	// Configure connection pooling
	transport.MaxIdleConns = r.MaxIdleConns
	transport.MaxIdleConnsPerHost = r.MaxIdleConnsPerHost
	transport.MaxConnsPerHost = r.MaxConnsPerHost
	transport.IdleConnTimeout = r.IdleConnTimeout
	
	return nil
}

// CreateHTTPClient creates a new HTTP client with these options
func (r *RemoteOptions) CreateHTTPClient() (*http.Client, error) {
	if err := r.Validate(); err != nil {
		return nil, fmt.Errorf("invalid remote options: %w", err)
	}
	
	transport := &http.Transport{}
	if err := r.ApplyToTransport(transport); err != nil {
		return nil, err
	}
	
	return &http.Client{
		Transport: transport,
		Timeout:   r.Timeout,
	}, nil
}

// Clone creates a deep copy of the RemoteOptions
func (r *RemoteOptions) Clone() *RemoteOptions {
	clone := &RemoteOptions{
		Timeout:             r.Timeout,
		MaxRetries:          r.MaxRetries,
		RetryDelay:          r.RetryDelay,
		UserAgent:           r.UserAgent,
		Insecure:            r.Insecure,
		CertFile:            r.CertFile,
		KeyFile:             r.KeyFile,
		CAFile:              r.CAFile,
		SkipTLSVerify:       r.SkipTLSVerify,
		ProxyURL:            r.ProxyURL,
		ProxyUsername:       r.ProxyUsername,
		ProxyPassword:       r.ProxyPassword,
		AuthScope:           r.AuthScope,
		AuthService:         r.AuthService,
		TokenRefreshMargin:  r.TokenRefreshMargin,
		MaxIdleConns:        r.MaxIdleConns,
		MaxIdleConnsPerHost: r.MaxIdleConnsPerHost,
		MaxConnsPerHost:     r.MaxConnsPerHost,
		IdleConnTimeout:     r.IdleConnTimeout,
		RegistryVersion:     r.RegistryVersion,
	}
	
	// Clone TLS config
	if r.TLSConfig != nil {
		clone.TLSConfig = r.TLSConfig.Clone()
	}
	
	// Clone slices
	if r.NoProxy != nil {
		clone.NoProxy = make([]string, len(r.NoProxy))
		copy(clone.NoProxy, r.NoProxy)
	}
	
	if r.AcceptedMediaTypes != nil {
		clone.AcceptedMediaTypes = make([]string, len(r.AcceptedMediaTypes))
		copy(clone.AcceptedMediaTypes, r.AcceptedMediaTypes)
	}
	
	// Clone map
	if r.CustomHeaders != nil {
		clone.CustomHeaders = make(map[string]string, len(r.CustomHeaders))
		for k, v := range r.CustomHeaders {
			clone.CustomHeaders[k] = v
		}
	}
	
	return clone
}

// WithTimeout returns a copy with the specified timeout
func (r *RemoteOptions) WithTimeout(timeout time.Duration) *RemoteOptions {
	clone := r.Clone()
	clone.Timeout = timeout
	return clone
}

// WithInsecure returns a copy with insecure mode enabled/disabled
func (r *RemoteOptions) WithInsecure(insecure bool) *RemoteOptions {
	clone := r.Clone()
	clone.Insecure = insecure
	clone.SkipTLSVerify = insecure
	return clone
}