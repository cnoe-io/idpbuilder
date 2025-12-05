// pkg/daemon/client_test.go
package daemon

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDaemonClient implements DaemonClient for testing.
type MockDaemonClient struct {
	mock.Mock
}

// GetImage implements DaemonClient.GetImage for mocking.
func (m *MockDaemonClient) GetImage(ctx context.Context, reference string) (*ImageInfo, ImageReader, error) {
	args := m.Called(ctx, reference)
	var info *ImageInfo
	var reader ImageReader
	if args.Get(0) != nil {
		info = args.Get(0).(*ImageInfo)
	}
	if args.Get(1) != nil {
		reader = args.Get(1).(ImageReader)
	}
	return info, reader, args.Error(2)
}

// ImageExists implements DaemonClient.ImageExists for mocking.
func (m *MockDaemonClient) ImageExists(ctx context.Context, reference string) (bool, error) {
	args := m.Called(ctx, reference)
	return args.Bool(0), args.Error(1)
}

// Ping implements DaemonClient.Ping for mocking.
func (m *MockDaemonClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockImageReader is a mock implementation of ImageReader for testing.
type MockImageReader struct {
	*bytes.Reader
}

func (m *MockImageReader) Close() error {
	return nil
}

// NewMockImageReader creates a new MockImageReader with the given data.
func NewMockImageReader(data []byte) *MockImageReader {
	return &MockImageReader{Reader: bytes.NewReader(data)}
}

// TestDaemonClient_GetImage_Success tests successful image retrieval.
func TestDaemonClient_GetImage_Success(t *testing.T) {
	// GIVEN a mock daemon client with a configured successful response
	ctx := context.Background()
	mockClient := new(MockDaemonClient)
	expectedInfo := &ImageInfo{
		ID:         "sha256:abc123def456",
		RepoTags:   []string{"myapp:latest", "myapp:v1.0"},
		Size:       1024 * 1024 * 50,
		LayerCount: 5,
	}
	mockReader := NewMockImageReader([]byte("mock image data"))
	mockClient.On("GetImage", ctx, "myapp:latest").Return(expectedInfo, mockReader, nil)

	// WHEN GetImage is called
	info, reader, err := mockClient.GetImage(ctx, "myapp:latest")

	// THEN no error is returned
	require.NoError(t, err)
	// AND image info matches expected
	assert.Equal(t, "sha256:abc123def456", info.ID)
	assert.Equal(t, []string{"myapp:latest", "myapp:v1.0"}, info.RepoTags)
	assert.Equal(t, int64(1024*1024*50), info.Size)
	assert.Equal(t, 5, info.LayerCount)
	// AND reader contains expected data
	data, _ := io.ReadAll(reader)
	assert.Equal(t, []byte("mock image data"), data)
	reader.Close()
	mockClient.AssertExpectations(t)
}

// TestDaemonClient_GetImage_NotFound tests image not found error.
func TestDaemonClient_GetImage_NotFound(t *testing.T) {
	// GIVEN a mock client configured to return ImageNotFoundError
	ctx := context.Background()
	mockClient := new(MockDaemonClient)
	notFoundErr := &ImageNotFoundError{Reference: "nonexistent:latest"}
	mockClient.On("GetImage", ctx, "nonexistent:latest").Return(nil, nil, notFoundErr)

	// WHEN GetImage is called with non-existent image
	info, reader, err := mockClient.GetImage(ctx, "nonexistent:latest")

	// THEN error is returned
	require.Error(t, err)
	// AND error is ImageNotFoundError
	var imgNotFound *ImageNotFoundError
	assert.True(t, errors.As(err, &imgNotFound))
	assert.Equal(t, "nonexistent:latest", imgNotFound.Reference)
	assert.Contains(t, imgNotFound.Error(), "image not found")
	// AND info and reader are nil
	assert.Nil(t, info)
	assert.Nil(t, reader)
	mockClient.AssertExpectations(t)
}

// TestDaemonClient_GetImage_DaemonNotRunning tests daemon unavailable error.
func TestDaemonClient_GetImage_DaemonNotRunning(t *testing.T) {
	// GIVEN a mock client configured to return DaemonError
	ctx := context.Background()
	mockClient := new(MockDaemonClient)
	daemonErr := &DaemonError{
		Message:      "Cannot connect to Docker daemon",
		IsNotRunning: true,
		Cause:        errors.New("connection refused"),
	}
	mockClient.On("GetImage", ctx, "myapp:latest").Return(nil, nil, daemonErr)

	// WHEN GetImage is called
	info, reader, err := mockClient.GetImage(ctx, "myapp:latest")

	// THEN DaemonError is returned
	require.Error(t, err)
	var de *DaemonError
	assert.True(t, errors.As(err, &de))
	assert.True(t, de.IsNotRunning)
	assert.Equal(t, "Cannot connect to Docker daemon", de.Message)
	assert.NotNil(t, de.Cause)
	// AND info and reader are nil
	assert.Nil(t, info)
	assert.Nil(t, reader)
	mockClient.AssertExpectations(t)
}

// TestDaemonClient_ImageExists_True tests image existence check returning true.
func TestDaemonClient_ImageExists_True(t *testing.T) {
	// GIVEN a mock client configured to return true for existing image
	ctx := context.Background()
	mockClient := new(MockDaemonClient)
	mockClient.On("ImageExists", ctx, "existing:latest").Return(true, nil)

	// WHEN ImageExists is called
	exists, err := mockClient.ImageExists(ctx, "existing:latest")

	// THEN no error is returned
	require.NoError(t, err)
	// AND exists is true
	assert.True(t, exists)
	mockClient.AssertExpectations(t)
}

// TestDaemonClient_ImageExists_False tests image existence check returning false.
func TestDaemonClient_ImageExists_False(t *testing.T) {
	// GIVEN a mock client configured to return false for non-existent image
	ctx := context.Background()
	mockClient := new(MockDaemonClient)
	mockClient.On("ImageExists", ctx, "nonexistent:latest").Return(false, nil)

	// WHEN ImageExists is called
	exists, err := mockClient.ImageExists(ctx, "nonexistent:latest")

	// THEN no error is returned
	require.NoError(t, err)
	// AND exists is false
	assert.False(t, exists)
	mockClient.AssertExpectations(t)
}

// TestDaemonClient_Ping_Success tests successful daemon ping.
func TestDaemonClient_Ping_Success(t *testing.T) {
	// GIVEN a mock client configured for successful ping
	ctx := context.Background()
	mockClient := new(MockDaemonClient)
	mockClient.On("Ping", ctx).Return(nil)

	// WHEN Ping is called
	err := mockClient.Ping(ctx)

	// THEN no error is returned
	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// TestDaemonClient_Ping_Failure tests daemon ping failure.
func TestDaemonClient_Ping_Failure(t *testing.T) {
	// GIVEN a mock client configured for failed ping
	ctx := context.Background()
	mockClient := new(MockDaemonClient)
	daemonErr := &DaemonError{
		Message:      "Cannot connect to Docker daemon",
		IsNotRunning: true,
		Cause:        errors.New("connection refused"),
	}
	mockClient.On("Ping", ctx).Return(daemonErr)

	// WHEN Ping is called
	err := mockClient.Ping(ctx)

	// THEN error is returned
	require.Error(t, err)
	// AND error is DaemonError
	var de *DaemonError
	assert.True(t, errors.As(err, &de))
	assert.True(t, de.IsNotRunning)
	mockClient.AssertExpectations(t)
}

// TestDaemonError_ErrorChaining tests error wrapping with Unwrap().
func TestDaemonError_ErrorChaining(t *testing.T) {
	// GIVEN a DaemonError with a cause
	causeErr := errors.New("underlying error")
	daemonErr := &DaemonError{
		Message:      "daemon error",
		IsNotRunning: true,
		Cause:        causeErr,
	}

	// WHEN Error() is called
	errMsg := daemonErr.Error()

	// THEN error message includes both message and cause
	assert.Contains(t, errMsg, "daemon error")
	assert.Contains(t, errMsg, "underlying error")

	// WHEN Unwrap() is called
	unwrapped := daemonErr.Unwrap()

	// THEN underlying error is returned
	assert.Equal(t, causeErr, unwrapped)

	// AND errors.Is works with error chaining
	assert.True(t, errors.Is(daemonErr, causeErr))
}

// TestImageNotFoundError tests ImageNotFoundError message format.
func TestImageNotFoundError(t *testing.T) {
	// GIVEN an ImageNotFoundError
	err := &ImageNotFoundError{Reference: "myimage:v1.2.3"}

	// WHEN Error() is called
	errMsg := err.Error()

	// THEN error message contains expected format
	assert.Contains(t, errMsg, "image not found")
	assert.Contains(t, errMsg, "myimage:v1.2.3")
	assert.Equal(t, "image not found: myimage:v1.2.3", errMsg)
}
