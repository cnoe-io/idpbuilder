package types

import (
	"time"
)

// RegistryConfig represents the configuration for a container registry
type RegistryConfig struct {
	// URL is the registry URL (e.g., "registry.example.com")
	URL string

	// Namespace is the registry namespace/organization
	Namespace string

	// Insecure allows insecure HTTP connections
	Insecure bool

	// SkipTLSVerify bypasses TLS certificate verification
	SkipTLSVerify bool

	// Timeout for registry operations
	Timeout time.Duration

	// RetryPolicy defines retry behavior
	RetryPolicy *RetryPolicy
}

// RetryPolicy defines retry behavior for registry operations
type RetryPolicy struct {
	MaxAttempts       int
	InitialDelay      time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
}

// RegistryInfo contains runtime information about a registry
type RegistryInfo struct {
	// Scheme (http/https)
	Scheme string

	// Host is the registry hostname
	Host string

	// Port number
	Port int

	// APIVersion of the registry
	APIVersion string

	// Capabilities supported by the registry
	Capabilities []string
}

// ImageReference represents a container image reference
type ImageReference struct {
	Registry   string
	Namespace  string
	Repository string
	Tag        string
	Digest     string
}

// Constants for common capabilities
const (
	CapabilityPush   = "push"
	CapabilityPull   = "pull"
	CapabilityDelete = "delete"
	CapabilityList   = "list"
)