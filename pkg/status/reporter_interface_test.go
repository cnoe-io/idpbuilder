package status

import (
	"testing"
)

// TestNilReporterAsInterface verifies that a nil *Reporter passed as an interface
// doesn't panic when methods are called. This simulates the real-world scenario
// where --status-output=none is used and the reporter is nil.
func TestNilReporterAsInterface(t *testing.T) {
	// This simulates how the reporter is used in the actual code
	var reporter *Reporter // nil

	// Create an interface that accepts a nil pointer (the root cause of the bug)
	var iface interface {
		AddSubStep(parentName, subStepName, description string)
		UpdateSubStep(parentName, subStepName string, state int)
	} = reporter

	// Verify that the interface is not nil (even though the underlying pointer is)
	// This is the Go gotcha that caused the original panic
	if iface == nil {
		t.Fatal("Interface should not be nil even when underlying reporter is nil")
	}

	// Now verify that calling methods doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Methods on nil Reporter passed as interface should not panic: %v", r)
		}
	}()

	// These calls would panic before the fix
	iface.AddSubStep("parent", "child", "description")
	iface.UpdateSubStep("parent", "child", 1)
}
