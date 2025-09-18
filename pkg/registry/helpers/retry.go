package helpers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

// RetryableFunc defines a function that can be retried
type RetryableFunc func(ctx context.Context) error

// HTTPRetryableFunc defines an HTTP operation that can be retried
type HTTPRetryableFunc func(ctx context.Context) (*http.Response, error)

// DefaultRetryPolicy provides sensible defaults for retry operations
func DefaultRetryPolicy() *types.RetryPolicy {
	return &types.RetryPolicy{
		MaxAttempts:       3,
		InitialDelay:      1 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// IsRetryableError determines if an error should trigger a retry
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for registry-specific errors
	if regErr, ok := err.(*types.RegistryError); ok {
		switch regErr.Code {
		case types.ErrCodeTimeout, types.ErrCodeConnectionFailed:
			return true
		case types.ErrCodeUnauthorized, types.ErrCodeForbidden, types.ErrCodeNotFound:
			return false // Don't retry auth/permission errors
		}
	}

	// Check error message for common retryable patterns
	errStr := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"timeout",
		"temporary failure",
		"network is unreachable",
		"no such host",
	}

	for _, pattern := range retryablePatterns {
		if contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// IsRetryableHTTPStatus determines if an HTTP status code should trigger a retry
func IsRetryableHTTPStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// RetryWithBackoff executes a function with exponential backoff retry
func RetryWithBackoff(ctx context.Context, fn RetryableFunc, policy *types.RetryPolicy) error {
	if policy == nil {
		policy = DefaultRetryPolicy()
	}

	var lastErr error
	delay := policy.InitialDelay

	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Don't retry on last attempt or if error is not retryable
		if attempt == policy.MaxAttempts || !IsRetryableError(err) {
			break
		}

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		default:
		}

		// Wait before retry
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("retry cancelled during backoff: %w", ctx.Err())
		case <-timer.C:
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * policy.BackoffMultiplier)
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
		}
	}

	return fmt.Errorf("all retry attempts failed: %w", lastErr)
}

// RetryHTTPRequest executes an HTTP request with retry logic
func RetryHTTPRequest(ctx context.Context, fn HTTPRetryableFunc, policy *types.RetryPolicy) (*http.Response, error) {
	if policy == nil {
		policy = DefaultRetryPolicy()
	}

	var lastResp *http.Response
	var lastErr error
	delay := policy.InitialDelay

	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		// Execute the HTTP request
		resp, err := fn(ctx)

		// Check for success (2xx status codes and no error)
		if err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// Store response and error for potential retry logic
		lastResp = resp
		lastErr = err

		// Determine if we should retry
		shouldRetry := false
		if err != nil && IsRetryableError(err) {
			shouldRetry = true
		} else if resp != nil && IsRetryableHTTPStatus(resp.StatusCode) {
			shouldRetry = true
			// Close response body to avoid resource leak
			if resp.Body != nil {
				resp.Body.Close()
			}
		}

		// Don't retry on last attempt or if not retryable
		if attempt == policy.MaxAttempts || !shouldRetry {
			break
		}

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
			return nil, fmt.Errorf("retry cancelled: %w", ctx.Err())
		default:
		}

		// Wait before retry
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
			return nil, fmt.Errorf("retry cancelled during backoff: %w", ctx.Err())
		case <-timer.C:
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * policy.BackoffMultiplier)
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
		}
	}

	// Return the last response and error
	if lastErr != nil {
		return lastResp, fmt.Errorf("all HTTP retry attempts failed: %w", lastErr)
	}

	// If we have a response but it's not successful, wrap it in an appropriate error
	if lastResp != nil {
		return lastResp, fmt.Errorf("HTTP request failed with status %d", lastResp.StatusCode)
	}

	return nil, fmt.Errorf("HTTP request failed after %d attempts", policy.MaxAttempts)
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		   (s == substr ||
			len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
			 s[len(s)-len(substr):] == substr ||
			 indexSubstring(s, substr) >= 0))
}

// indexSubstring finds the index of a substring in a string
func indexSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}