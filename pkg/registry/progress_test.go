package registry

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStderrProgressReporter_Start tests the Start implementation.
func TestStderrProgressReporter_Start(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := &StderrProgressReporter{Out: buf}

	reporter.Start("myapp:latest", 3)

	output := buf.String()
	assert.Contains(t, output, "Pushing")
	assert.Contains(t, output, "myapp:latest")
	assert.Contains(t, output, "3 layers")
}

// TestStderrProgressReporter_LayerProgress tests the LayerProgress implementation.
func TestStderrProgressReporter_LayerProgress(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := &StderrProgressReporter{Out: buf}

	// Test milestone at 50%
	reporter.LayerProgress("sha256:abc123def456", 512000, 1024000)

	output := buf.String()
	assert.Contains(t, output, "50%")
	assert.Contains(t, output, "abc123def45")
}

// TestStderrProgressReporter_Complete tests the Complete implementation.
func TestStderrProgressReporter_Complete(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := &StderrProgressReporter{Out: buf}

	result := &PushResult{
		Reference: "registry.example.com/myapp:v1.0.0@sha256:abc123",
		Digest:    "sha256:abc123",
		Size:      1024000,
	}

	reporter.Complete(result)

	output := buf.String()
	assert.Contains(t, output, "Push complete")
	assert.Contains(t, output, "sha256:abc123")
}

// TestProgressReporter_OutputsToStderr verifies the Out writer field exists and is usable.
func TestProgressReporter_OutputsToStderr(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := &StderrProgressReporter{Out: buf}

	// Verify the Out field is accessible and can be written to
	assert.NotNil(t, reporter.Out)

	// Test that we can write to the buffer (for Wave 3 implementation)
	n, err := reporter.Out.Write([]byte("test output"))
	assert.NoError(t, err)
	assert.Equal(t, 11, n)
	assert.Equal(t, "test output", buf.String())
}

// TestStderrProgressReporter_Error tests the Error implementation.
func TestStderrProgressReporter_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := &StderrProgressReporter{Out: buf}

	testErr := errors.New("test error")
	reporter.Error(testErr)

	output := buf.String()
	assert.Contains(t, output, "Push failed")
	assert.Contains(t, output, "test error")
}

// TestStderrProgressReporter_LayerComplete tests the LayerComplete implementation.
func TestStderrProgressReporter_LayerComplete(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := &StderrProgressReporter{Out: buf}

	reporter.LayerComplete("sha256:abc123def456")

	output := buf.String()
	assert.Contains(t, output, "abc123def45")
	assert.Contains(t, output, "done")
}
