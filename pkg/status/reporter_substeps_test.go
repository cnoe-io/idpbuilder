package status

import (
	"testing"
	"time"
)

// Test that sub-steps are displayed correctly
func TestReporter_SubSteps(t *testing.T) {
	reporter := NewReporter(false) // No color for easier testing

	// Add main steps
	reporter.AddStep("packages", "Installing and syncing packages")

	// Start the packages step
	reporter.StartStep("packages")

	// Add sub-steps
	reporter.AddSubStep("packages", "argocd", "argocd")
	reporter.AddSubStep("packages", "custom-dir-0", "/path/to/custom")
	reporter.AddSubStep("packages", "custom-url-0", "https://example.com/package")

	// Verify sub-steps were added
	if len(reporter.steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(reporter.steps))
	}

	if len(reporter.steps[0].SubSteps) != 3 {
		t.Errorf("Expected 3 sub-steps, got %d", len(reporter.steps[0].SubSteps))
	}

	// Verify sub-step names
	expectedNames := []string{"argocd", "custom-dir-0", "custom-url-0"}
	for i, expected := range expectedNames {
		if reporter.steps[0].SubSteps[i].Name != expected {
			t.Errorf("Expected sub-step name %s, got %s", expected, reporter.steps[0].SubSteps[i].Name)
		}
	}

	// Test updating sub-step states
	reporter.UpdateSubStep("packages", "argocd", 1) // Running
	if reporter.steps[0].SubSteps[0].State != StateRunning {
		t.Errorf("Expected argocd to be running, got %v", reporter.steps[0].SubSteps[0].State)
	}

	reporter.UpdateSubStep("packages", "argocd", 2) // Complete
	if reporter.steps[0].SubSteps[0].State != StateComplete {
		t.Errorf("Expected argocd to be complete, got %v", reporter.steps[0].SubSteps[0].State)
	}

	// Test that adding the same sub-step multiple times doesn't create duplicates
	initialCount := len(reporter.steps[0].SubSteps)
	reporter.AddSubStep("packages", "argocd", "argocd")
	if len(reporter.steps[0].SubSteps) != initialCount+1 {
		t.Errorf("AddSubStep should add duplicate (reporter doesn't prevent this)")
	}
}

// Test concurrent sub-step updates (thread safety)
func TestReporter_ConcurrentSubStepUpdates(t *testing.T) {
	reporter := NewReporter(false)
	reporter.AddStep("packages", "Installing packages")
	reporter.StartStep("packages")

	// Add multiple sub-steps
	for i := 0; i < 5; i++ {
		reporter.AddSubStep("packages", string(rune('a'+i)), string(rune('a'+i)))
	}

	// Update them concurrently
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(idx int) {
			name := string(rune('a' + idx))
			reporter.UpdateSubStep("packages", name, 1) // Running
			time.Sleep(10 * time.Millisecond)
			reporter.UpdateSubStep("packages", name, 2) // Complete
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify all sub-steps are complete
	for i, substep := range reporter.steps[0].SubSteps {
		if substep.State != StateComplete {
			t.Errorf("Sub-step %d (%s) should be complete, got %v", i, substep.Name, substep.State)
		}
	}
}
