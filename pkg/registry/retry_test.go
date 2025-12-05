package registry

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TC-RT-001: TestRetryableClient_Push_Success - Success on first attempt
func TestRetryableClient_Push_Success(t *testing.T) {
	mockClient := new(MockRegistryClient)
	ctx := context.Background()

	expectedResult := &PushResult{
		Reference: "registry.example.com/image:v1.0.0",
		Digest:    "sha256:abc123",
		Size:      1024,
	}

	mockClient.On("Push", ctx, "image:latest", "registry.example.com/image:v1.0.0", mock.Anything).
		Return(expectedResult, nil)

	retryClient := NewRetryableClient(mockClient, DefaultRetryConfig())
	result, err := retryClient.Push(ctx, "image:latest", "registry.example.com/image:v1.0.0", nil)

	require.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	mockClient.AssertNumberOfCalls(t, "Push", 1)
}

// TC-RT-002: TestRetryableClient_Push_TransientError_ThenSuccess - Retry and succeed
func TestRetryableClient_Push_TransientError_ThenSuccess(t *testing.T) {
	mockClient := new(MockRegistryClient)
	ctx := context.Background()

	transientErr := &RegistryError{
		Message:     "connection timeout",
		IsTransient: true,
	}

	expectedResult := &PushResult{
		Reference: "registry.example.com/image:v1.0.0",
		Digest:    "sha256:abc123",
		Size:      1024,
	}

	mockClient.On("Push", ctx, "image:latest", "registry.example.com/image:v1.0.0", mock.Anything).
		Return(nil, transientErr).
		Once().
		On("Push", ctx, "image:latest", "registry.example.com/image:v1.0.0", mock.Anything).
		Return(expectedResult, nil)

	config := RetryConfig{
		MaxRetries:        10,
		InitialDelay:      1 * time.Millisecond,
		MaxDelay:          10 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	retryClient := NewRetryableClient(mockClient, config)
	result, err := retryClient.Push(ctx, "image:latest", "registry.example.com/image:v1.0.0", nil)

	require.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	mockClient.AssertNumberOfCalls(t, "Push", 2)
}

// TC-RT-003: TestRetryableClient_Push_PermanentError_NoRetry - No retry on AuthError
func TestRetryableClient_Push_PermanentError_NoRetry(t *testing.T) {
	mockClient := new(MockRegistryClient)
	ctx := context.Background()

	authErr := &AuthError{
		Message: "invalid credentials",
	}

	mockClient.On("Push", ctx, "image:latest", "registry.example.com/image:v1.0.0", mock.Anything).
		Return(nil, authErr)

	retryClient := NewRetryableClient(mockClient, DefaultRetryConfig())
	_, err := retryClient.Push(ctx, "image:latest", "registry.example.com/image:v1.0.0", nil)

	require.Error(t, err)
	assert.Equal(t, authErr, err)
	mockClient.AssertNumberOfCalls(t, "Push", 1)
}

// TC-RT-004: TestRetryableClient_Push_ExhaustedRetries - Property P1.2: exactly 10 retries
func TestRetryableClient_Push_ExhaustedRetries(t *testing.T) {
	mockClient := new(MockRegistryClient)
	ctx := context.Background()

	transientErr := &RegistryError{
		Message:     "connection timeout",
		IsTransient: true,
	}

	mockClient.On("Push", ctx, "image:latest", "registry.example.com/image:v1.0.0", mock.Anything).
		Return(nil, transientErr)

	maxRetries := 10
	config := RetryConfig{
		MaxRetries:        maxRetries,
		InitialDelay:      1 * time.Millisecond,
		MaxDelay:          10 * time.Millisecond,
		BackoffMultiplier: 1.5,
	}

	retryClient := NewRetryableClient(mockClient, config)
	_, err := retryClient.Push(ctx, "image:latest", "registry.example.com/image:v1.0.0", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "after 11 attempts")
	mockClient.AssertNumberOfCalls(t, "Push", maxRetries+1)
}

// TC-RT-005: TestRetryableClient_Push_ContextCancellation - Ctrl+C handling
func TestRetryableClient_Push_ContextCancellation(t *testing.T) {
	mockClient := new(MockRegistryClient)

	transientErr := &RegistryError{
		Message:     "connection timeout",
		IsTransient: true,
	}

	mockClient.On("Push", mock.Anything, "image:latest", "registry.example.com/image:v1.0.0", mock.Anything).
		Return(nil, transientErr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := RetryConfig{
		MaxRetries:        10,
		InitialDelay:      1 * time.Millisecond,
		MaxDelay:          10 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	retryClient := NewRetryableClient(mockClient, config)

	// Start push in goroutine to cancel during retry wait
	errChan := make(chan error, 1)
	go func() {
		_, err := retryClient.Push(ctx, "image:latest", "registry.example.com/image:v1.0.0", nil)
		errChan <- err
	}()

	// Cancel context after a brief delay to ensure first attempt and retry wait happen
	time.Sleep(5 * time.Millisecond)
	cancel()

	err := <-errChan
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
}

// TC-RT-006: TestRetryableClient_Push_NotifyFunc - User notification
func TestRetryableClient_Push_NotifyFunc(t *testing.T) {
	mockClient := new(MockRegistryClient)
	ctx := context.Background()

	transientErr := &RegistryError{
		Message:     "connection timeout",
		IsTransient: true,
	}

	expectedResult := &PushResult{
		Reference: "registry.example.com/image:v1.0.0",
		Digest:    "sha256:abc123",
		Size:      1024,
	}

	mockClient.On("Push", ctx, "image:latest", "registry.example.com/image:v1.0.0", mock.Anything).
		Return(nil, transientErr).
		Once().
		On("Push", ctx, "image:latest", "registry.example.com/image:v1.0.0", mock.Anything).
		Return(expectedResult, nil)

	notifyCount := 0
	config := RetryConfig{
		MaxRetries:        10,
		InitialDelay:      1 * time.Millisecond,
		MaxDelay:          10 * time.Millisecond,
		BackoffMultiplier: 2.0,
		NotifyFunc: func(attempt int, delay time.Duration, err error) {
			notifyCount++
			assert.Equal(t, 1, attempt)
			assert.True(t, delay > 0)
			assert.NotNil(t, err)
		},
	}

	retryClient := NewRetryableClient(mockClient, config)
	result, err := retryClient.Push(ctx, "image:latest", "registry.example.com/image:v1.0.0", nil)

	require.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	assert.Equal(t, 1, notifyCount)
}

// TC-RT-007: TestCalculateDelay_ExponentialBackoff - 1s, 2s, 4s, 8s pattern
func TestCalculateDelay_ExponentialBackoff(t *testing.T) {
	config := RetryConfig{
		InitialDelay:      1 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
	}
	client := NewRetryableClient(&MockRegistryClient{}, config)

	tests := []struct {
		attempt      int
		expectedDelay time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 16 * time.Second},
		{5, 30 * time.Second}, // capped at MaxDelay
	}

	for _, tt := range tests {
		delay := client.calculateDelay(tt.attempt)
		assert.Equal(t, tt.expectedDelay, delay, "attempt %d", tt.attempt)
	}
}

// TC-RT-008: TestCalculateDelay_MaxDelayCap - Capped at 30s
func TestCalculateDelay_MaxDelayCap(t *testing.T) {
	config := RetryConfig{
		InitialDelay:      1 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
	}
	client := NewRetryableClient(&MockRegistryClient{}, config)

	// Attempt 10 would exceed 30s without cap
	delay := client.calculateDelay(10)
	assert.Equal(t, 30*time.Second, delay)
	assert.True(t, delay <= config.MaxDelay)
}

// TC-RT-009: TestIsTransient_ErrorClassification - Error classification
func TestIsTransient_ErrorClassification(t *testing.T) {
	config := DefaultRetryConfig()
	client := NewRetryableClient(&MockRegistryClient{}, config)

	tests := []struct {
		name       string
		err        error
		isTransient bool
	}{
		{
			name:        "nil error",
			err:         nil,
			isTransient: false,
		},
		{
			name: "explicit transient",
			err: &RegistryError{
				Message:     "timeout",
				IsTransient: true,
			},
			isTransient: true,
		},
		{
			name: "explicit non-transient",
			err: &RegistryError{
				Message:     "not found",
				IsTransient: false,
			},
			isTransient: false,
		},
	}

	for _, tt := range tests {
		result := client.isTransient(tt.err)
		assert.Equal(t, tt.isTransient, result, "test: %s", tt.name)
	}
}

// TC-RT-010: TestIsTransient_AuthError_NotTransient - AuthError never retried
func TestIsTransient_AuthError_NotTransient(t *testing.T) {
	config := DefaultRetryConfig()
	client := NewRetryableClient(&MockRegistryClient{}, config)

	authErr := &AuthError{
		Message: "invalid credentials",
	}

	result := client.isTransient(authErr)
	assert.False(t, result)
}

// TC-RT-011: TestIsTransient_NetworkTimeout_Transient - Timeout is transient
func TestIsTransient_NetworkTimeout_Transient(t *testing.T) {
	config := DefaultRetryConfig()
	client := NewRetryableClient(&MockRegistryClient{}, config)

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "i/o timeout",
			err:  errors.New("i/o timeout"),
		},
		{
			name: "timeout in message",
			err:  errors.New("network timeout occurred"),
		},
		{
			name: "Timeout uppercase",
			err:  errors.New("Timeout waiting for response"),
		},
	}

	for _, tt := range tests {
		result := client.isTransient(tt.err)
		assert.True(t, result, "test: %s", tt.name)
	}
}

// TC-RT-012: TestIsTransient_ConnectionRefused_Transient - Connection refused is transient
func TestIsTransient_ConnectionRefused_Transient(t *testing.T) {
	config := DefaultRetryConfig()
	client := NewRetryableClient(&MockRegistryClient{}, config)

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "connection refused",
			err:  errors.New("connection refused"),
		},
		{
			name: "connection reset",
			err:  errors.New("connection reset by peer"),
		},
		{
			name: "EOF",
			err:  errors.New("EOF"),
		},
		{
			name: "broken pipe",
			err:  errors.New("broken pipe"),
		},
	}

	for _, tt := range tests {
		result := client.isTransient(tt.err)
		assert.True(t, result, "test: %s", tt.name)
	}
}

// TestStderrRetryNotifier - Verify notification format
func TestStderrRetryNotifier(t *testing.T) {
	type WriteFunc struct {
		io.Writer
		data string
	}

	w := &WriteFunc{}
	w.Writer = &mockWriter{data: &w.data}

	notifier := StderrRetryNotifier(w)
	err := errors.New("test error")

	notifier(1, 5*time.Second, err)

	assert.Contains(t, w.data, "Push attempt")
	assert.Contains(t, w.data, "failed")
	assert.Contains(t, w.data, "Retrying in")
}

type mockWriter struct {
	data *string
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	*m.data += string(p)
	return len(p), nil
}

// TestNewRetryableClient_DefaultsApplied - Verify zero values get defaults
func TestNewRetryableClient_DefaultsApplied(t *testing.T) {
	mockClient := &MockRegistryClient{}
	config := RetryConfig{} // All zero values

	client := NewRetryableClient(mockClient, config)

	assert.Equal(t, 10, client.config.MaxRetries)
	assert.Equal(t, 1*time.Second, client.config.InitialDelay)
	assert.Equal(t, 30*time.Second, client.config.MaxDelay)
	assert.Equal(t, 2.0, client.config.BackoffMultiplier)
}

// TestContainsIgnoreCase - Case-insensitive string matching
func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		str       string
		substring string
		expected  bool
	}{
		{"timeout", "timeout", true},
		{"TIMEOUT", "timeout", true},
		{"Timeout", "timeout", true},
		{"timeout error", "timeout", true},
		{"CONNECTION REFUSED", "connection refused", true},
		{"no match", "timeout", false},
	}

	for _, tt := range tests {
		result := containsIgnoreCase(tt.str, tt.substring)
		assert.Equal(t, tt.expected, result, "str=%s, substring=%s", tt.str, tt.substring)
	}
}

// TestToLower - ASCII lowercase conversion
func TestToLower(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ABC", "abc"},
		{"abc", "abc"},
		{"AbC", "abc"},
		{"Connection Refused", "connection refused"},
		{"123", "123"},
		{"!@#", "!@#"},
	}

	for _, tt := range tests {
		result := toLower(tt.input)
		assert.Equal(t, tt.expected, result, "input=%s", tt.input)
	}
}

// TestContains - Simple substring check
func TestContains(t *testing.T) {
	tests := []struct {
		str       string
		substring string
		expected  bool
	}{
		{"hello", "hello", true},
		{"hello world", "world", true},
		{"hello", "bye", false},
		{"", "", true},
		{"a", "b", false},
	}

	for _, tt := range tests {
		result := contains(tt.str, tt.substring)
		assert.Equal(t, tt.expected, result, "str=%s, substring=%s", tt.str, tt.substring)
	}
}
