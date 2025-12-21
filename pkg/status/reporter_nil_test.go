package status

import (
	"errors"
	"testing"
)

// TestNilReporter_NoPanic verifies that calling methods on a nil Reporter doesn't panic
// This is important when --status-output=none is used and the reporter is nil
func TestNilReporter_NoPanic(t *testing.T) {
	var r *Reporter // nil reporter

	// Test all public methods to ensure they handle nil receiver gracefully
	t.Run("SetSimpleMode", func(t *testing.T) {
		defer func() {
			if rec := recover(); rec != nil {
				t.Errorf("SetSimpleMode panicked on nil receiver: %v", rec)
			}
		}()
		r.SetSimpleMode(true)
	})

	t.Run("GetSteps", func(t *testing.T) {
		defer func() {
			if rec := recover(); rec != nil {
				t.Errorf("GetSteps panicked on nil receiver: %v", rec)
			}
		}()
		steps := r.GetSteps()
		if steps != nil {
			t.Errorf("GetSteps on nil receiver should return nil, got %v", steps)
		}
	})

	t.Run("AddStep", func(t *testing.T) {
		defer func() {
			if rec := recover(); rec != nil {
				t.Errorf("AddStep panicked on nil receiver: %v", rec)
			}
		}()
		r.AddStep("test", "test description")
	})

	t.Run("StartStep", func(t *testing.T) {
		defer func() {
			if rec := recover(); rec != nil {
				t.Errorf("StartStep panicked on nil receiver: %v", rec)
			}
		}()
		r.StartStep("test")
	})

	t.Run("CompleteStep", func(t *testing.T) {
		defer func() {
			if rec := recover(); rec != nil {
				t.Errorf("CompleteStep panicked on nil receiver: %v", rec)
			}
		}()
		r.CompleteStep("test")
	})

	t.Run("FailStep", func(t *testing.T) {
		defer func() {
			if rec := recover(); rec != nil {
				t.Errorf("FailStep panicked on nil receiver: %v", rec)
			}
		}()
		r.FailStep("test", errors.New("test error"))
	})

	t.Run("AddSubStep", func(t *testing.T) {
		defer func() {
			if rec := recover(); rec != nil {
				t.Errorf("AddSubStep panicked on nil receiver: %v", rec)
			}
		}()
		r.AddSubStep("parent", "child", "test description")
	})

	t.Run("UpdateSubStep", func(t *testing.T) {
		defer func() {
			if rec := recover(); rec != nil {
				t.Errorf("UpdateSubStep panicked on nil receiver: %v", rec)
			}
		}()
		r.UpdateSubStep("parent", "child", 1)
	})

	t.Run("Summary", func(t *testing.T) {
		defer func() {
			if rec := recover(); rec != nil {
				t.Errorf("Summary panicked on nil receiver: %v", rec)
			}
		}()
		r.Summary()
	})
}
