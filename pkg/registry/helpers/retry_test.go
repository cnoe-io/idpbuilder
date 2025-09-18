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

func TestRetryHTTPRequest(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func() HTTPRetryableFunc
		policy       *types.RetryPolicy
		expectError  bool
		expectStatus int
	}{
		{
			name: "Successful request on first try",
			setupFunc: func() HTTPRetryableFunc {
				return func(ctx context.Context) (*http.Response, error) {
					resp := &http.Response{
						StatusCode: 200,
						Status:     "200 OK",
					}
					return resp, nil
				}
			},
			expectError:  false,
			expectStatus: 200,
		},
		{
			name: "Successful after retry",
			setupFunc: func() HTTPRetryableFunc {
				attempt := 0
				return func(ctx context.Context) (*http.Response, error) {
					attempt++
					if attempt == 1 {
						return &http.Response{StatusCode: 500}, nil
					}
					return &http.Response{StatusCode: 200}, nil
				}
			},
			policy: &types.RetryPolicy{
				MaxAttempts:       2,
				InitialDelay:      1 * time.Millisecond,
				MaxDelay:          10 * time.Millisecond,
				BackoffMultiplier: 1.5,
			},
			expectError:  false,
			expectStatus: 200,
		},
		{
			name: "All attempts fail with retryable error",
			setupFunc: func() HTTPRetryableFunc {
				return func(ctx context.Context) (*http.Response, error) {
					return nil, &types.RegistryError{Code: types.ErrCodeTimeout, Message: "timeout"}
				}
			},
			policy: &types.RetryPolicy{
				MaxAttempts:       2,
				InitialDelay:      1 * time.Millisecond,
				MaxDelay:          10 * time.Millisecond,
				BackoffMultiplier: 1.5,
			},
			expectError: true,
		},
		{
			name: "All attempts fail with retryable HTTP status",
			setupFunc: func() HTTPRetryableFunc {
				return func(ctx context.Context) (*http.Response, error) {
					return &http.Response{StatusCode: 503, Status: "503 Service Unavailable"}, nil
				}
			},
			policy: &types.RetryPolicy{
				MaxAttempts:       2,
				InitialDelay:      1 * time.Millisecond,
				MaxDelay:          10 * time.Millisecond,
				BackoffMultiplier: 1.5,
			},
			expectError: true,
		},
		{
			name: "Non-retryable error fails immediately",
			setupFunc: func() HTTPRetryableFunc {
				return func(ctx context.Context) (*http.Response, error) {
					return &http.Response{StatusCode: 404, Status: "404 Not Found"}, nil
				}
			},
			policy: &types.RetryPolicy{
				MaxAttempts:       3,
				InitialDelay:      1 * time.Millisecond,
				MaxDelay:          10 * time.Millisecond,
				BackoffMultiplier: 1.5,
			},
			expectError: true, // 404 still results in error, but only tries once since not retryable
		},
		{
			name: "Context cancellation",
			setupFunc: func() HTTPRetryableFunc {
				return func(ctx context.Context) (*http.Response, error) {
					// Check if context is already cancelled
					select {
					case <-ctx.Done():
						return nil, ctx.Err()
					default:
						return &http.Response{StatusCode: 500}, nil
					}
				}
			},
			policy: &types.RetryPolicy{
				MaxAttempts:       3,
				InitialDelay:      100 * time.Millisecond,
				MaxDelay:          1 * time.Second,
				BackoffMultiplier: 2.0,
			},
			expectError: true,
		},
		{
			name: "Nil policy uses default",
			setupFunc: func() HTTPRetryableFunc {
				return func(ctx context.Context) (*http.Response, error) {
					return &http.Response{StatusCode: 200}, nil
				}
			},
			policy:       nil, // Should use default policy
			expectError:  false,
			expectStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Special case for context cancellation test
			if tt.name == "Context cancellation" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				// Cancel after a short delay to test retry cancellation
				go func() {
					time.Sleep(50 * time.Millisecond)
					cancel()
				}()
			}

			fn := tt.setupFunc()
			resp, err := RetryHTTPRequest(ctx, fn, tt.policy)

			if tt.expectError {
				if err == nil {
					t.Errorf("RetryHTTPRequest() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("RetryHTTPRequest() error = %v, wantErr %v", err, tt.expectError)
					return
				}

				if resp == nil {
					t.Errorf("RetryHTTPRequest() returned nil response")
					return
				}

				if resp.StatusCode != tt.expectStatus {
					t.Errorf("RetryHTTPRequest() status = %d, want %d", resp.StatusCode, tt.expectStatus)
				}
			}
		})
	}
}

func TestRetryHTTPRequest_BackoffProgression(t *testing.T) {
	var delays []time.Duration
	startTime := time.Now()

	attempt := 0
	fn := func(ctx context.Context) (*http.Response, error) {
		if attempt > 0 {
			delays = append(delays, time.Since(startTime))
		}
		attempt++
		startTime = time.Now()
		return &http.Response{StatusCode: 500}, nil
	}

	policy := &types.RetryPolicy{
		MaxAttempts:       3,
		InitialDelay:      10 * time.Millisecond,
		MaxDelay:          100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	ctx := context.Background()
	_, err := RetryHTTPRequest(ctx, fn, policy)

	if err == nil {
		t.Error("Expected error after all retries failed")
	}

	if len(delays) != 2 { // Should have 2 delays between 3 attempts
		t.Errorf("Expected 2 delays, got %d", len(delays))
	}

	// Check that delays are roughly correct (with some tolerance for timing)
	expectedDelays := []time.Duration{10 * time.Millisecond, 20 * time.Millisecond}
	for i, expectedDelay := range expectedDelays {
		if i < len(delays) {
			actualDelay := delays[i]
			tolerance := 50 * time.Millisecond
			if actualDelay < expectedDelay-tolerance || actualDelay > expectedDelay+tolerance {
				t.Errorf("Delay %d: expected ~%v, got %v", i, expectedDelay, actualDelay)
			}
		}
	}
}