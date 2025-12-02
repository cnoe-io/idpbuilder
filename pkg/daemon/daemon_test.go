// pkg/daemon/daemon_test.go
package daemon

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// W2-DC-001: TestDefaultDaemonClient_ImageExists_True
func TestDefaultDaemonClient_ImageExists_True(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewDefaultClient()
	if err != nil {
		t.Skip("Docker daemon not available")
	}

	// Uses alpine:latest which should be available
	exists, err := client.ImageExists(context.Background(), "alpine:latest")
	require.NoError(t, err)
	assert.True(t, exists)
}

// W2-DC-002: TestDefaultDaemonClient_ImageExists_False
func TestDefaultDaemonClient_ImageExists_False(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewDefaultClient()
	if err != nil {
		t.Skip("Docker daemon not available")
	}

	exists, err := client.ImageExists(context.Background(), "nonexistent-image-12345:notag")
	require.NoError(t, err)
	assert.False(t, exists)
}

// W2-DC-003: TestDefaultDaemonClient_ImageExists_DaemonDown
func TestDefaultDaemonClient_ImageExists_DaemonDown(t *testing.T) {
	// Set invalid DOCKER_HOST to simulate daemon down
	original := os.Getenv("DOCKER_HOST")
	defer os.Setenv("DOCKER_HOST", original)
	os.Setenv("DOCKER_HOST", "unix:///invalid/path/docker.sock")

	_, err := NewDefaultClient()
	require.Error(t, err)

	var de *DaemonError
	assert.True(t, errors.As(err, &de))
	assert.True(t, de.IsNotRunning)
}

// W2-DC-004: TestDefaultDaemonClient_GetImage_Success
func TestDefaultDaemonClient_GetImage_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewDefaultClient()
	if err != nil {
		t.Skip("Docker daemon not available")
	}

	info, reader, err := client.GetImage(context.Background(), "alpine:latest")
	require.NoError(t, err)
	require.NotNil(t, info)
	require.NotNil(t, reader)
	defer reader.Close()

	assert.NotEmpty(t, info.ID)
	assert.Greater(t, info.Size, int64(0))
	assert.Greater(t, info.LayerCount, 0)
}

// W2-DC-005: TestDefaultDaemonClient_GetImage_NotFound
func TestDefaultDaemonClient_GetImage_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewDefaultClient()
	if err != nil {
		t.Skip("Docker daemon not available")
	}

	info, reader, err := client.GetImage(context.Background(), "nonexistent-12345:notag")
	require.Error(t, err)
	assert.Nil(t, info)
	assert.Nil(t, reader)

	var notFoundErr *ImageNotFoundError
	assert.True(t, errors.As(err, &notFoundErr))
	assert.Equal(t, "nonexistent-12345:notag", notFoundErr.Reference)
}

// W2-DC-006: TestDefaultDaemonClient_Ping_Success
func TestDefaultDaemonClient_Ping_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewDefaultClient()
	if err != nil {
		t.Skip("Docker daemon not available")
	}

	err = client.Ping(context.Background())
	require.NoError(t, err)
}

// W2-DC-007: TestDefaultDaemonClient_Ping_Failure
func TestDefaultDaemonClient_Ping_Failure(t *testing.T) {
	original := os.Getenv("DOCKER_HOST")
	defer os.Setenv("DOCKER_HOST", original)
	os.Setenv("DOCKER_HOST", "unix:///invalid/path/docker.sock")

	client := &DefaultDaemonClient{dockerHost: "unix:///invalid/path/docker.sock"}
	err := client.Ping(context.Background())
	require.Error(t, err)

	var de *DaemonError
	assert.True(t, errors.As(err, &de))
	assert.True(t, de.IsNotRunning)
}

// W2-DC-008: TestDefaultDaemonClient_DOCKER_HOST
func TestDefaultDaemonClient_DOCKER_HOST(t *testing.T) {
	original := os.Getenv("DOCKER_HOST")
	defer os.Setenv("DOCKER_HOST", original)

	customHost := "unix:///var/run/custom-docker.sock"
	os.Setenv("DOCKER_HOST", customHost)

	_, err := NewDefaultClient()
	// Expected error - socket doesn't exist
	if err != nil {
		var de *DaemonError
		assert.True(t, errors.As(err, &de))
	}
}

// W2-DC-009: TestDefaultDaemonClient_ErrorClassification
func TestDefaultDaemonClient_ErrorClassification(t *testing.T) {
	tests := []struct {
		name           string
		errorMsg       string
		wantNotFound   bool
		wantDaemonDown bool
	}{
		{"not_found", "No such image: myapp:latest", true, false},
		{"manifest_unknown", "manifest unknown", true, false},
		{"connection_refused", "connection refused", false, true},
		{"daemon_not_running", "Cannot connect to the Docker daemon", false, true},
		{"dial_unix", "dial unix /var/run/docker.sock: connect: no such file", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errorMsg)
			notFound := isNotFoundError(err)
			daemonDown := isDaemonUnavailable(err)

			assert.Equal(t, tt.wantNotFound, notFound, "isNotFoundError mismatch")
			assert.Equal(t, tt.wantDaemonDown, daemonDown, "isDaemonUnavailable mismatch")
		})
	}
}
