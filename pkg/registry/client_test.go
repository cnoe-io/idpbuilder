package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRegistryClient implements RegistryClient for testing.
type MockRegistryClient struct {
	mock.Mock
}

// Push implements RegistryClient.Push for mocking.
func (m *MockRegistryClient) Push(ctx context.Context, imageRef, destRef string, progress ProgressReporter) (*PushResult, error) {
	args := m.Called(ctx, imageRef, destRef, progress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PushResult), args.Error(1)
}

// MockProgressReporter implements ProgressReporter for testing.
type MockProgressReporter struct {
	mock.Mock
}

func (m *MockProgressReporter) Start(imageRef string, totalLayers int) {
	m.Called(imageRef, totalLayers)
}

func (m *MockProgressReporter) LayerProgress(layerDigest string, current, total int64) {
	m.Called(layerDigest, current, total)
}

func (m *MockProgressReporter) LayerComplete(layerDigest string) {
	m.Called(layerDigest)
}

func (m *MockProgressReporter) Complete(result *PushResult) {
	m.Called(result)
}

func (m *MockProgressReporter) Error(err error) {
	m.Called(err)
}

// TestRegistryClient_Push_Success tests successful push operation.
func TestRegistryClient_Push_Success(t *testing.T) {
	mockClient := new(MockRegistryClient)
	ctx := context.Background()

	expectedResult := &PushResult{
		Reference: "registry.example.com/myapp:v1.0.0@sha256:abc123",
		Digest:    "sha256:abc123",
		Size:      1024000,
	}

	mockClient.On("Push", ctx, "myapp:latest", "registry.example.com/myapp:v1.0.0", mock.Anything).
		Return(expectedResult, nil)

	result, err := mockClient.Push(ctx, "myapp:latest", "registry.example.com/myapp:v1.0.0", nil)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "registry.example.com/myapp:v1.0.0@sha256:abc123", result.Reference)
	assert.Equal(t, "sha256:abc123", result.Digest)
	assert.Equal(t, int64(1024000), result.Size)
	mockClient.AssertExpectations(t)
}

// TestRegistryClient_Push_AuthError tests authentication failure.
func TestRegistryClient_Push_AuthError(t *testing.T) {
	mockClient := new(MockRegistryClient)
	ctx := context.Background()

	authErr := &AuthError{
		Message: "authentication failed",
		Cause:   errors.New("invalid credentials"),
	}

	mockClient.On("Push", ctx, "myapp:latest", "registry.example.com/myapp:v1.0.0", mock.Anything).
		Return(nil, authErr)

	result, err := mockClient.Push(ctx, "myapp:latest", "registry.example.com/myapp:v1.0.0", nil)

	require.Error(t, err)
	require.Nil(t, result)
	var ae *AuthError
	assert.True(t, errors.As(err, &ae))
	assert.Equal(t, "authentication failed", ae.Message)
	mockClient.AssertExpectations(t)
}

// TestRegistryClient_Push_TransientError tests transient error handling.
func TestRegistryClient_Push_TransientError(t *testing.T) {
	mockClient := new(MockRegistryClient)
	ctx := context.Background()

	transientErr := &RegistryError{
		StatusCode:  503,
		Message:     "service unavailable",
		IsTransient: true,
		Cause:       errors.New("temporary network issue"),
	}

	mockClient.On("Push", ctx, "myapp:latest", "registry.example.com/myapp:v1.0.0", mock.Anything).
		Return(nil, transientErr)

	result, err := mockClient.Push(ctx, "myapp:latest", "registry.example.com/myapp:v1.0.0", nil)

	require.Error(t, err)
	require.Nil(t, result)
	var re *RegistryError
	assert.True(t, errors.As(err, &re))
	assert.Equal(t, 503, re.StatusCode)
	assert.True(t, re.IsTransient)
	mockClient.AssertExpectations(t)
}

// TestRegistryClient_Push_WithProgress tests progress callback invocation.
func TestRegistryClient_Push_WithProgress(t *testing.T) {
	mockClient := new(MockRegistryClient)
	mockProgress := new(MockProgressReporter)
	ctx := context.Background()

	expectedResult := &PushResult{
		Reference: "registry.example.com/myapp:v1.0.0@sha256:abc123",
		Digest:    "sha256:abc123",
		Size:      1024000,
	}

	// Set up progress expectations
	mockProgress.On("Start", "myapp:latest", 3).Return()
	mockProgress.On("LayerProgress", "sha256:layer1", int64(512000), int64(512000)).Return()
	mockProgress.On("LayerComplete", "sha256:layer1").Return()
	mockProgress.On("Complete", expectedResult).Return()

	// Set up client to simulate push with progress
	mockClient.On("Push", ctx, "myapp:latest", "registry.example.com/myapp:v1.0.0", mockProgress).
		Run(func(args mock.Arguments) {
			prog := args.Get(3).(ProgressReporter)
			prog.Start("myapp:latest", 3)
			prog.LayerProgress("sha256:layer1", 512000, 512000)
			prog.LayerComplete("sha256:layer1")
			prog.Complete(expectedResult)
		}).
		Return(expectedResult, nil)

	result, err := mockClient.Push(ctx, "myapp:latest", "registry.example.com/myapp:v1.0.0", mockProgress)

	require.NoError(t, err)
	require.NotNil(t, result)
	mockClient.AssertExpectations(t)
	mockProgress.AssertExpectations(t)
}

// TestRegistryError_ErrorChaining tests error wrapping with Unwrap().
func TestRegistryError_ErrorChaining(t *testing.T) {
	cause := errors.New("connection refused")
	regErr := &RegistryError{
		StatusCode:  500,
		Message:     "registry error",
		IsTransient: true,
		Cause:       cause,
	}

	// Test Error() method
	assert.Contains(t, regErr.Error(), "registry error")
	assert.Contains(t, regErr.Error(), "connection refused")

	// Test Unwrap() method
	unwrapped := errors.Unwrap(regErr)
	assert.Equal(t, cause, unwrapped)

	// Test errors.Is compatibility
	assert.True(t, errors.Is(regErr, cause))
}

// TestAuthError_ErrorChaining tests AuthError wrapping.
func TestAuthError_ErrorChaining(t *testing.T) {
	cause := errors.New("invalid token")
	authErr := &AuthError{
		Message: "authentication failed",
		Cause:   cause,
	}

	// Test Error() method
	assert.Contains(t, authErr.Error(), "authentication failed")
	assert.Contains(t, authErr.Error(), "invalid token")

	// Test Unwrap() method
	unwrapped := errors.Unwrap(authErr)
	assert.Equal(t, cause, unwrapped)

	// Test errors.Is compatibility
	assert.True(t, errors.Is(authErr, cause))
}

// TestNoOpProgressReporter_DoesNothing tests that NoOp is safe to call.
func TestNoOpProgressReporter_DoesNothing(t *testing.T) {
	reporter := &NoOpProgressReporter{}

	// These should not panic
	assert.NotPanics(t, func() {
		reporter.Start("myapp:latest", 3)
	})
	assert.NotPanics(t, func() {
		reporter.LayerProgress("sha256:layer1", 100, 1000)
	})
	assert.NotPanics(t, func() {
		reporter.LayerComplete("sha256:layer1")
	})
	assert.NotPanics(t, func() {
		reporter.Complete(&PushResult{})
	})
	assert.NotPanics(t, func() {
		reporter.Error(errors.New("test error"))
	})
}
