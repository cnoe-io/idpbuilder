package status_test

import (
	"testing"
	"time"

	"github.com/cnoe-io/idpbuilder/pkg/status"
)

// TestReporter_ArgoCDAndGiteaParallel verifies that both argocd and gitea substeps
// can be tracked in parallel and displayed correctly
func TestReporter_ArgoCDAndGiteaParallel(t *testing.T) {
	reporter := status.NewReporter(false)

	// Add workflow steps
	reporter.AddStep("cluster", "Creating Kubernetes cluster")
	reporter.AddStep("crds", "Installing Custom Resource Definitions")
	reporter.AddStep("networking", "Configuring networking and certificates")
	reporter.AddStep("resources", "Creating platform resources")
	reporter.AddStep("packages", "Installing and syncing packages")

	// Execute steps
	reporter.StartStep("cluster")
	time.Sleep(5 * time.Millisecond)
	reporter.CompleteStep("cluster")

	reporter.StartStep("crds")
	time.Sleep(5 * time.Millisecond)
	reporter.CompleteStep("crds")

	reporter.StartStep("networking")
	time.Sleep(5 * time.Millisecond)
	reporter.CompleteStep("networking")

	reporter.StartStep("resources")
	time.Sleep(5 * time.Millisecond)
	reporter.CompleteStep("resources")

	// Start packages step with sub-steps
	reporter.StartStep("packages")

	// Add both argocd and gitea sub-steps (as they are now tracked in parallel)
	reporter.AddSubStep("packages", "argocd", "argocd")
	reporter.AddSubStep("packages", "gitea", "gitea")

	// Verify sub-steps were added
	steps := reporter.GetSteps()
	if len(steps[4].SubSteps) != 2 {
		t.Errorf("Expected 2 sub-steps (argocd and gitea), got %d", len(steps[4].SubSteps))
	}

	// Simulate parallel installation - argocd starts
	reporter.UpdateSubStep("packages", "argocd", 1) // Running
	time.Sleep(5 * time.Millisecond)

	// Gitea starts (in parallel)
	reporter.UpdateSubStep("packages", "gitea", 1) // Running
	time.Sleep(5 * time.Millisecond)

	// Both are running - verify states
	steps = reporter.GetSteps()
	if steps[4].SubSteps[0].State != status.StateRunning {
		t.Errorf("ArgoCD should be running, got state %v", steps[4].SubSteps[0].State)
	}
	if steps[4].SubSteps[1].State != status.StateRunning {
		t.Errorf("Gitea should be running, got state %v", steps[4].SubSteps[1].State)
	}

	// ArgoCD completes first
	reporter.UpdateSubStep("packages", "argocd", 2) // Complete
	time.Sleep(5 * time.Millisecond)

	// Verify argocd is complete but gitea is still running
	steps = reporter.GetSteps()
	if steps[4].SubSteps[0].State != status.StateComplete {
		t.Errorf("ArgoCD should be complete, got state %v", steps[4].SubSteps[0].State)
	}
	if steps[4].SubSteps[1].State != status.StateRunning {
		t.Errorf("Gitea should still be running, got state %v", steps[4].SubSteps[1].State)
	}

	// Gitea completes
	reporter.UpdateSubStep("packages", "gitea", 2) // Complete
	time.Sleep(5 * time.Millisecond)

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

// TestReporter_GiteaFailureScenario verifies that gitea failure is properly tracked
func TestReporter_GiteaFailureScenario(t *testing.T) {
	reporter := status.NewReporter(false)

	reporter.AddStep("packages", "Installing and syncing packages")
	reporter.StartStep("packages")

	// Add both argocd and gitea sub-steps
	reporter.AddSubStep("packages", "argocd", "argocd")
	reporter.AddSubStep("packages", "gitea", "gitea")

	// Both start
	reporter.UpdateSubStep("packages", "argocd", 1) // Running
	reporter.UpdateSubStep("packages", "gitea", 1)  // Running

	// ArgoCD completes successfully
	reporter.UpdateSubStep("packages", "argocd", 2) // Complete

	// Gitea fails
	reporter.UpdateSubStep("packages", "gitea", 3) // Failed

	// Verify states
	steps := reporter.GetSteps()
	if steps[0].SubSteps[0].State != status.StateComplete {
		t.Errorf("ArgoCD should be complete, got state %v", steps[0].SubSteps[0].State)
	}
	if steps[0].SubSteps[1].State != status.StateFailed {
		t.Errorf("Gitea should be failed, got state %v", steps[0].SubSteps[1].State)
	}
}
