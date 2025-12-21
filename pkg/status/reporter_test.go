package status

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestReporter_BasicFlow(t *testing.T) {
	var buf bytes.Buffer
	r := NewReporter(false)
	r.writer = &buf

	// Add steps
	r.AddStep("step1", "First step")
	r.AddStep("step2", "Second step")
	r.AddStep("step3", "Third step")

	// Start and complete steps
	r.StartStep("step1")
	time.Sleep(10 * time.Millisecond)
	r.CompleteStep("step1")

	r.StartStep("step2")
	time.Sleep(10 * time.Millisecond)
	r.CompleteStep("step2")

	r.StartStep("step3")
	time.Sleep(10 * time.Millisecond)
	r.CompleteStep("step3")

	r.Summary()

	output := buf.String()

	// Verify output contains step descriptions
	if !strings.Contains(output, "First step") {
		t.Errorf("Output should contain 'First step'")
	}
	if !strings.Contains(output, "Second step") {
		t.Errorf("Output should contain 'Second step'")
	}
	if !strings.Contains(output, "Third step") {
		t.Errorf("Output should contain 'Third step'")
	}
	if !strings.Contains(output, "Build completed successfully") {
		t.Errorf("Output should contain success message")
	}
}

func TestReporter_FailedStep(t *testing.T) {
	var buf bytes.Buffer
	r := NewReporter(false)
	r.writer = &buf

	r.AddStep("step1", "First step")
	r.AddStep("step2", "Second step")

	r.StartStep("step1")
	r.CompleteStep("step1")

	r.StartStep("step2")
	r.FailStep("step2", nil)

	r.Summary()

	output := buf.String()

	if !strings.Contains(output, "Build failed") {
		t.Errorf("Output should contain failure message, got: %s", output)
	}
}

func TestReporter_StateSymbols(t *testing.T) {
	r := NewReporter(false)

	tests := []struct {
		state    State
		expected string
	}{
		{StatePending, "○"},
		{StateRunning, "●"},
		{StateComplete, "✓"},
		{StateFailed, "✗"},
	}

	for _, tt := range tests {
		got := r.getSymbol(tt.state)
		if got != tt.expected {
			t.Errorf("getSymbol(%v) = %s, want %s", tt.state, got, tt.expected)
		}
	}
}

func TestReporter_ColoredOutput(t *testing.T) {
	r := NewReporter(true)

	// Test that color codes are returned when colored is true
	if r.color(Green) != Green {
		t.Errorf("color() should return Green when colored is true")
	}

	r2 := NewReporter(false)

	// Test that empty string is returned when colored is false
	if r2.color(Green) != "" {
		t.Errorf("color() should return empty string when colored is false")
	}
}
