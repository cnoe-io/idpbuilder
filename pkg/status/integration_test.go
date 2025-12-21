package status_test

import (
	"testing"
	"time"

	"github.com/cnoe-io/idpbuilder/pkg/status"
)

// TestReporter_SimulateRealWorkflow simulates a realistic workflow with sub-steps
func TestReporter_SimulateRealWorkflow(t *testing.T) {
	reporter := status.NewReporter(false)

	// Add workflow steps
	reporter.AddStep("cluster", "Creating Kubernetes cluster")
	reporter.AddStep("crds", "Installing Custom Resource Definitions")
	reporter.AddStep("networking", "Configuring networking and certificates")
	reporter.AddStep("resources", "Creating platform resources")
	reporter.AddStep("packages", "Installing and syncing packages")

	// Execute steps
	reporter.StartStep("cluster")
	time.Sleep(10 * time.Millisecond)
	reporter.CompleteStep("cluster")

	reporter.StartStep("crds")
	time.Sleep(10 * time.Millisecond)
	reporter.CompleteStep("crds")

	reporter.StartStep("networking")
	time.Sleep(10 * time.Millisecond)
	reporter.CompleteStep("networking")

	reporter.StartStep("resources")
	time.Sleep(10 * time.Millisecond)
	reporter.CompleteStep("resources")

	// Start packages step with sub-steps
	reporter.StartStep("packages")

	// Add package sub-steps
	reporter.AddSubStep("packages", "argocd", "argocd")
	reporter.AddSubStep("packages", "custom-dir-0", "/example/custom")
	reporter.AddSubStep("packages", "custom-url-0", "https://github.com/example/pkg")

	// Verify sub-steps were added
	steps := reporter.GetSteps()
	if len(steps[4].SubSteps) != 3 {
		t.Errorf("Expected 3 sub-steps, got %d", len(steps[4].SubSteps))
	}

	// Process argocd
	reporter.UpdateSubStep("packages", "argocd", 1) // Running
	time.Sleep(20 * time.Millisecond)
	reporter.UpdateSubStep("packages", "argocd", 2) // Complete

	// Process custom-dir-0
	reporter.UpdateSubStep("packages", "custom-dir-0", 1)
	time.Sleep(15 * time.Millisecond)
	reporter.UpdateSubStep("packages", "custom-dir-0", 2)

	// Process custom-url-0
	reporter.UpdateSubStep("packages", "custom-url-0", 1)
	time.Sleep(15 * time.Millisecond)
	reporter.UpdateSubStep("packages", "custom-url-0", 2)

	// Complete packages step
	reporter.CompleteStep("packages")

	// Verify all sub-steps are complete
	steps = reporter.GetSteps()
	for i, substep := range steps[4].SubSteps {
		if substep.State != status.StateComplete {
			t.Errorf("Sub-step %d (%s) should be complete, got state %v",
				i, substep.Name, substep.State)
		}
	}
}
