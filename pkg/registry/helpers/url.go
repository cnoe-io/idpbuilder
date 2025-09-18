// Package helpers provides utility functions for registry operations,
// building on top of the authentication and types packages.
package helpers

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

// ParseRegistryURL parses a registry URL and extracts components
func ParseRegistryURL(rawURL string) (*types.RegistryInfo, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("registry URL cannot be empty")
	}

	// Add scheme if missing
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid registry URL: %w", err)
	}

	info := &types.RegistryInfo{
		Scheme: parsedURL.Scheme,
		Host:   parsedURL.Hostname(),
	}

	// Parse port
	if parsedURL.Port() != "" {
		port, err := strconv.Atoi(parsedURL.Port())
		if err != nil {
			return nil, fmt.Errorf("invalid port in URL: %w", err)
		}
		info.Port = port
	} else {
		// Set default ports
		if info.Scheme == "https" {
			info.Port = 443
		} else {
			info.Port = 80
		}
	}

	return info, nil
}

// BuildRegistryURL constructs a registry URL from components
func BuildRegistryURL(info *types.RegistryInfo) string {
	if info == nil {
		return ""
	}

	scheme := info.Scheme
	if scheme == "" {
		scheme = "https"
	}

	host := info.Host
	if host == "" {
		return ""
	}

	// Add port if non-default
	if (scheme == "https" && info.Port != 443) || (scheme == "http" && info.Port != 80) {
		host = fmt.Sprintf("%s:%d", host, info.Port)
	}

	return fmt.Sprintf("%s://%s", scheme, host)
}

// ValidateRegistryConfig validates a registry configuration
func ValidateRegistryConfig(config *types.RegistryConfig) error {
	if config == nil {
		return fmt.Errorf("registry config cannot be nil")
	}

	if config.URL == "" {
		return fmt.Errorf("registry URL is required")
	}

	// Parse URL to validate format
	_, err := ParseRegistryURL(config.URL)
	if err != nil {
		return fmt.Errorf("invalid registry URL: %w", err)
	}

	// Validate retry policy if provided
	if config.RetryPolicy != nil {
		if config.RetryPolicy.MaxAttempts < 1 {
			return fmt.Errorf("retry policy MaxAttempts must be >= 1")
		}
		if config.RetryPolicy.BackoffMultiplier <= 0 {
			return fmt.Errorf("retry policy BackoffMultiplier must be > 0")
		}
	}

	return nil
}

// NormalizeRegistryURL normalizes a registry URL for consistent comparisons
func NormalizeRegistryURL(rawURL string) (string, error) {
	info, err := ParseRegistryURL(rawURL)
	if err != nil {
		return "", err
	}

	normalized := BuildRegistryURL(info)

	// Remove trailing slashes
	normalized = strings.TrimSuffix(normalized, "/")

	return normalized, nil
}