package status

import (
	"strings"
	"testing"
)

// Test that sub-steps with phase information are displayed correctly
func TestReporter_SubStepsWithPhase(t *testing.T) {
	reporter := NewReporter(false) // No color for easier testing

	// Add main steps
	reporter.AddStep("packages", "Installing and syncing packages")

	// Start the packages step
	reporter.StartStep("packages")

	// Add sub-steps
	reporter.AddSubStep("packages", "gitea", "gitea")
	reporter.AddSubStep("packages", "argocd", "argocd")
	reporter.AddSubStep("packages", "nginx", "nginx")

	// Verify sub-steps were added
	if len(reporter.steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(reporter.steps))
	}

	if len(reporter.steps[0].SubSteps) != 3 {
		t.Errorf("Expected 3 sub-steps, got %d", len(reporter.steps[0].SubSteps))
	}

	// Update sub-steps with phase information
	reporter.UpdateSubStepWithPhase("packages", "gitea", 1, "Installing") // Running
	if reporter.steps[0].SubSteps[0].State != StateRunning {
		t.Errorf("Expected gitea to be running, got %v", reporter.steps[0].SubSteps[0].State)
	}
	if !strings.Contains(reporter.steps[0].SubSteps[0].Description, "Installing") {
		t.Errorf("Expected gitea description to contain 'Installing', got %s", reporter.steps[0].SubSteps[0].Description)
	}

	reporter.UpdateSubStepWithPhase("packages", "gitea", 2, "Ready") // Complete
	if reporter.steps[0].SubSteps[0].State != StateComplete {
		t.Errorf("Expected gitea to be complete, got %v", reporter.steps[0].SubSteps[0].State)
	}
	if !strings.Contains(reporter.steps[0].SubSteps[0].Description, "Ready") {
		t.Errorf("Expected gitea description to contain 'Ready', got %s", reporter.steps[0].SubSteps[0].Description)
	}

	// Update argocd with different phase
	reporter.UpdateSubStepWithPhase("packages", "argocd", 1, "Pending")
	if !strings.Contains(reporter.steps[0].SubSteps[1].Description, "Pending") {
		t.Errorf("Expected argocd description to contain 'Pending', got %s", reporter.steps[0].SubSteps[1].Description)
	}

	reporter.UpdateSubStepWithPhase("packages", "argocd", 1, "Installing")
	if !strings.Contains(reporter.steps[0].SubSteps[1].Description, "Installing") {
		t.Errorf("Expected argocd description to contain 'Installing', got %s", reporter.steps[0].SubSteps[1].Description)
	}

	// Test with empty phase - should not crash
	reporter.UpdateSubStepWithPhase("packages", "nginx", 1, "")
	if reporter.steps[0].SubSteps[2].State != StateRunning {
		t.Errorf("Expected nginx to be running, got %v", reporter.steps[0].SubSteps[2].State)
	}
}

// Test that regular UpdateSubStep still works
func TestReporter_UpdateSubStepBackwardCompatibility(t *testing.T) {
	reporter := NewReporter(false)
	reporter.AddStep("packages", "Installing packages")
	reporter.StartStep("packages")

	reporter.AddSubStep("packages", "test", "test-package")

	// Use old UpdateSubStep method
	reporter.UpdateSubStep("packages", "test", 1) // Running
	if reporter.steps[0].SubSteps[0].State != StateRunning {
		t.Errorf("Expected test to be running")
	}

	reporter.UpdateSubStep("packages", "test", 2) // Complete
	if reporter.steps[0].SubSteps[0].State != StateComplete {
		t.Errorf("Expected test to be complete")
	}
}

// Test that empty phase clears previous phase information
func TestReporter_EmptyPhaseClearsPreviousPhase(t *testing.T) {
	reporter := NewReporter(false)
	reporter.AddStep("packages", "Installing packages")
	reporter.StartStep("packages")

	reporter.AddSubStep("packages", "test", "test-package")

	// Set a phase
	reporter.UpdateSubStepWithPhase("packages", "test", 1, "Installing")
	if !strings.Contains(reporter.steps[0].SubSteps[0].Description, "Installing") {
		t.Errorf("Expected description to contain 'Installing', got %s", reporter.steps[0].SubSteps[0].Description)
	}

	// Clear the phase with empty string
	reporter.UpdateSubStepWithPhase("packages", "test", 1, "")
	if strings.Contains(reporter.steps[0].SubSteps[0].Description, "Installing") {
		t.Errorf("Expected description to NOT contain 'Installing' after clearing, got %s", reporter.steps[0].SubSteps[0].Description)
	}
	if reporter.steps[0].SubSteps[0].Description != "test" {
		t.Errorf("Expected description to be 'test', got %s", reporter.steps[0].SubSteps[0].Description)
	}
}
