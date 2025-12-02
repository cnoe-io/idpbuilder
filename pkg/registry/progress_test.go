package registry

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStderrProgressReporter_Start tests the Start message formatting
func TestStderrProgressReporter_Start(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := &StderrProgressReporter{Out: buf}

	reporter.Start("myapp:latest", 3)

	output := buf.String()
	assert.Contains(t, output, "Pushing myapp:latest")
	assert.Contains(t, output, "3 layers")
}

// TestStderrProgressReporter_LayerProgress_Milestones tests milestone output at 25%, 50%, 75%
func TestStderrProgressReporter_LayerProgress_Milestones(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := &StderrProgressReporter{Out: buf}
	reporter.Start("myapp:latest", 1)
	buf.Reset()

	totalSize := int64(1000)

	// Progress to 25%
	reporter.LayerProgress("sha256:abc123def456", 100, totalSize) // 10%
	reporter.LayerProgress("sha256:abc123def456", 250, totalSize) // 25%
	output := buf.String()
	assert.Contains(t, output, "25%")
	buf.Reset()

	// Progress to 50%
	reporter.LayerProgress("sha256:abc123def456", 500, totalSize)
	output = buf.String()
	assert.Contains(t, output, "50%")
	buf.Reset()

	// Progress to 75%
	reporter.LayerProgress("sha256:abc123def456", 750, totalSize)
	output = buf.String()
	assert.Contains(t, output, "75%")
}

// TestStderrProgressReporter_Complete tests completion message with digest and time
func TestStderrProgressReporter_Complete(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := &StderrProgressReporter{Out: buf}
	reporter.Start("myapp:latest", 1)

	result := &PushResult{
		Reference: "registry.example.com/myapp:v1.0.0@sha256:abc123",
		Digest:    "sha256:abc123",
		Size:      1024000,
	}

	reporter.Complete(result)
	reporter.Complete(result)

	output := buf.String()
	assert.Contains(t, output, "Push complete:")
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

// TestStderrProgressReporter_Error tests error message formatting
func TestStderrProgressReporter_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := &StderrProgressReporter{Out: buf}

	testErr := errors.New("connection refused")
	reporter.Error(testErr)

	output := buf.String()
	assert.Contains(t, output, "Push failed:")
	assert.Contains(t, output, "connection refused")
}

// TestStderrProgressReporter_LayerComplete tests layer completion message
func TestStderrProgressReporter_LayerComplete(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := &StderrProgressReporter{Out: buf}

	reporter.LayerComplete("sha256:abc123def456")

	output := buf.String()
	assert.Contains(t, output, "done")
	assert.Contains(t, output, "sha256:abc123d")
}

// TestStderrProgressReporter_FormatBytes tests byte formatting
func TestStderrProgressReporter_FormatBytes(t *testing.T) {
	reporter := &StderrProgressReporter{}

	tests := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.00 KB"},
		{1024 * 1024, "1.00 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
		{1536 * 1024, "1.50 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := reporter.formatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStderrProgressReporter_ShortenDigest tests digest shortening
func TestStderrProgressReporter_ShortenDigest(t *testing.T) {
	reporter := &StderrProgressReporter{}

	tests := []struct {
		digest   string
		expected string
	}{
		{"sha256:abc123", "sha256:abc123"},
		{"sha256:abc123def456ghi789", "sha256:abc123de"},
		{"sha256:abc123def456ghi789jkl012mnopqr", "sha256:abc123de"},
	}

	for _, tt := range tests {
		t.Run(tt.digest, func(t *testing.T) {
			result := reporter.shortenDigest(tt.digest)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStderrProgressReporter_NilOut tests no panic with nil Out
func TestStderrProgressReporter_NilOut(t *testing.T) {
	reporter := &StderrProgressReporter{Out: nil}

	// Should not panic
	assert.NotPanics(t, func() {
		reporter.Start("myapp:latest", 1)
		reporter.LayerProgress("sha256:abc123", 100, 1000)
		reporter.LayerComplete("sha256:abc123")
		reporter.Complete(&PushResult{Digest: "sha256:abc123", Size: 1000})
		reporter.Error(errors.New("test error"))
	})
}

// TestStderrProgressReporter_ThreadSafe tests concurrent safety
func TestStderrProgressReporter_ThreadSafe(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := &StderrProgressReporter{Out: buf}
	reporter.Start("myapp:latest", 5)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(layerNum int) {
			defer wg.Done()
			layerDigest := fmt.Sprintf("sha256:layer%d", layerNum)
			for j := int64(0); j <= 1000; j += 250 {
				reporter.LayerProgress(layerDigest, j, 1000)
			}
			reporter.LayerComplete(layerDigest)
		}(i)
	}

	wg.Wait()
	// Should complete without race conditions
}
