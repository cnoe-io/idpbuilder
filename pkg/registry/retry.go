// Package registry provides interfaces and types for pushing OCI images to registries.
package registry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"time"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (default: 10)
	MaxRetries int
	// InitialDelay is the initial delay before first retry (default: 1s)
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries (default: 30s)
	MaxDelay time.Duration
	// BackoffMultiplier is the exponential multiplier (default: 2.0)
	BackoffMultiplier float64
	// NotifyFunc is called before each retry to notify the user
	NotifyFunc func(attempt int, delay time.Duration, err error)
}

// DefaultRetryConfig returns production defaults per PRD
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        10,
		InitialDelay:      1 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		NotifyFunc:        nil,
	}
}

// RetryableClient wraps RegistryClient with retry logic for transient errors
type RetryableClient struct {
	client RegistryClient
	config RetryConfig
}

// NewRetryableClient creates a new RetryableClient that wraps the given RegistryClient
// with retry logic for transient errors. It applies defaults for zero-valued config fields.
func NewRetryableClient(client RegistryClient, config RetryConfig) *RetryableClient {
	// Apply defaults for zero values
	if config.MaxRetries == 0 {
		config.MaxRetries = DefaultRetryConfig().MaxRetries
	}
	if config.InitialDelay == 0 {
		config.InitialDelay = DefaultRetryConfig().InitialDelay
	}
	if config.MaxDelay == 0 {
		config.MaxDelay = DefaultRetryConfig().MaxDelay
	}
	if config.BackoffMultiplier == 0 {
		config.BackoffMultiplier = DefaultRetryConfig().BackoffMultiplier
	}

	return &RetryableClient{
		client: client,
		config: config,
	}
}

// Push implements retry loop with context cancellation support (REQ-013)
// It retries transient errors up to MaxRetries times with exponential backoff.
func (r *RetryableClient) Push(ctx context.Context, imageRef, destRef string, progress ProgressReporter) (*PushResult, error) {
	var lastErr error

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		// Check context before attempting (REQ-013: graceful Ctrl+C handling)
		select {
		case <-ctx.Done():
			return nil, &RegistryError{
				Message:     "push cancelled",
				IsTransient: false,
				Cause:       ctx.Err(),
			}
		default:
		}

		// Attempt push
		result, err := r.client.Push(ctx, imageRef, destRef, progress)
		if err == nil {
			return result, nil
		}

		// Classify error - permanent errors return immediately
		if !r.isTransient(err) {
			return nil, err
		}

		// Check if we have retries left
		if attempt >= r.config.MaxRetries {
			lastErr = err
			break
		}

		// Calculate delay with exponential backoff (REQ-008: 1s, 2s, 4s, ...)
		delay := r.calculateDelay(attempt)

		// Notify before retry (REQ-010: user notification before each retry)
		if r.config.NotifyFunc != nil {
			r.config.NotifyFunc(attempt+1, delay, err)
		}

		// Wait with context cancellation support (REQ-013: graceful Ctrl+C during retry wait)
		select {
		case <-ctx.Done():
			return nil, &RegistryError{
				Message:     "push cancelled during retry wait",
				IsTransient: false,
				Cause:       ctx.Err(),
			}
		case <-time.After(delay):
			// Continue to next retry
		}

		lastErr = err
	}

	// All retries exhausted (REQ-009: maximum 10 retry attempts)
	return nil, &RegistryError{
		Message:     fmt.Sprintf("push failed after %d attempts", r.config.MaxRetries+1),
		IsTransient: false,
		Cause:       lastErr,
	}
}

// isTransient classifies errors for retry decisions
// Returns false for permanent errors (AuthError) and true for transient errors.
func (r *RetryableClient) isTransient(err error) bool {
	if err == nil {
		return false
	}

	// AuthError - never retry authentication failures
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return false
	}

	// RegistryError with explicit flag
	var regErr *RegistryError
	if errors.As(err, &regErr) {
		return regErr.IsTransient
	}

	// String pattern matching for transient errors
	transientPatterns := []string{
		"timeout", "connection refused", "connection reset",
		"temporary failure", "network unreachable", "no such host",
		"i/o timeout", "EOF", "broken pipe",
	}

	errStr := err.Error()
	for _, pattern := range transientPatterns {
		if containsIgnoreCase(errStr, pattern) {
			return true
		}
	}

	return false
}

// calculateDelay computes exponential backoff delay (REQ-008)
// Formula: delay = InitialDelay * (BackoffMultiplier ^ attempt)
// Returns capped at MaxDelay (30s default)
func (r *RetryableClient) calculateDelay(attempt int) time.Duration {
	delay := float64(r.config.InitialDelay) * math.Pow(r.config.BackoffMultiplier, float64(attempt))
	if delay > float64(r.config.MaxDelay) {
		return r.config.MaxDelay
	}
	return time.Duration(delay)
}

// StderrRetryNotifier returns a NotifyFunc that writes retry notifications to the given io.Writer
// This is the default notifier for user-facing operations
func StderrRetryNotifier(out io.Writer) func(attempt int, delay time.Duration, err error) {
	return func(attempt int, delay time.Duration, err error) {
		fmt.Fprintf(out, "Push attempt %d failed: %v\n", attempt, err)
		fmt.Fprintf(out, "Retrying in %v...\n", delay.Round(time.Millisecond))
	}
}

// containsIgnoreCase checks if substring is in str (case-insensitive)
func containsIgnoreCase(str, substring string) bool {
	return contains(toLower(str), toLower(substring))
}

// contains checks if substring is in str (case-sensitive)
func contains(str, substring string) bool {
	return len(str) >= len(substring) && findSubstring(str, substring)
}

// findSubstring is a helper that finds substring in str
func findSubstring(str, substring string) bool {
	for i := 0; i <= len(str)-len(substring); i++ {
		match := true
		for j := 0; j < len(substring); j++ {
			if str[i+j] != substring[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// toLower converts ASCII uppercase letters to lowercase
// This is a simple ASCII-only lowercase conversion (not Unicode-aware)
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b >= 'A' && b <= 'Z' {
			result[i] = b + ('a' - 'A')
		} else {
			result[i] = b
		}
	}
	return string(result)
}
