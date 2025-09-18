package helpers

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/cnoe-io/idpbuilder/pkg/registry/types"
)

func TestDefaultRetryPolicy(t *testing.T) {
	policy := DefaultRetryPolicy()

	if policy.MaxAttempts != 3 {
		t.Errorf("DefaultRetryPolicy() MaxAttempts = %v, want 3", policy.MaxAttempts)
	}

	if policy.InitialDelay != 1*time.Second {
		t.Errorf("DefaultRetryPolicy() InitialDelay = %v, want 1s", policy.InitialDelay)
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "Registry timeout error",
			err:      &types.RegistryError{Code: types.ErrCodeTimeout, Message: "timeout"},
			expected: true,
		},
		{
			name:     "Registry unauthorized error",
			err:      &types.RegistryError{Code: types.ErrCodeUnauthorized, Message: "unauthorized"},
			expected: false,
		},
		{
			name:     "Connection refused error",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "Generic error",
			err:      errors.New("some generic error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryableError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsRetryableHTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		expected bool
	}{
		{
			name:     "200 OK",
			status:   http.StatusOK,
			expected: false,
		},
		{
			name:     "429 Too Many Requests",
			status:   http.StatusTooManyRequests,
			expected: true,
		},
		{
			name:     "500 Internal Server Error",
			status:   http.StatusInternalServerError,
			expected: true,
		},
		{
			name:     "404 Not Found",
			status:   http.StatusNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableHTTPStatus(tt.status)
			if result != tt.expected {
				t.Errorf("IsRetryableHTTPStatus() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRetryWithBackoff(t *testing.T) {
	tests := []struct {
		name          string
		attempts      int
		failAttempts  int
		expectedCalls int
		expectError   bool
	}{
		{
			name:          "Success on first attempt",
			attempts:      3,
			failAttempts:  0,
			expectedCalls: 1,
			expectError:   false,
		},
		{
			name:          "Success on second attempt",
			attempts:      3,
			failAttempts:  1,
			expectedCalls: 2,
			expectError:   false,
		},
		{
			name:          "All attempts fail",
			attempts:      2,
			failAttempts:  3,
			expectedCalls: 2,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			fn := func(ctx context.Context) error {
				callCount++
				if callCount <= tt.failAttempts {
					return &types.RegistryError{Code: types.ErrCodeTimeout, Message: "timeout"}
				}
				return nil
			}

			policy := &types.RetryPolicy{
				MaxAttempts:       tt.attempts,
				InitialDelay:      1 * time.Millisecond,
				MaxDelay:          10 * time.Millisecond,
				BackoffMultiplier: 2.0,
			}

			ctx := context.Background()
			err := RetryWithBackoff(ctx, fn, policy)

			if (err != nil) != tt.expectError {
				t.Errorf("RetryWithBackoff() error = %v, expectError %v", err, tt.expectError)
			}

			if callCount != tt.expectedCalls {
				t.Errorf("RetryWithBackoff() calls = %d, want %d", callCount, tt.expectedCalls)
			}
		})
	}
}

func TestRetryWithBackoffContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	fn := func(ctx context.Context) error {
		cancel() // Cancel context during execution
		time.Sleep(10 * time.Millisecond) // Small delay to ensure cancellation is processed
		return &types.RegistryError{Code: types.ErrCodeTimeout, Message: "timeout"}
	}

	policy := &types.RetryPolicy{
		MaxAttempts:       3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 2.0,
	}

	err := RetryWithBackoff(ctx, fn, policy)
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}
}