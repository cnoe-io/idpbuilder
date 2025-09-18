package types

import "testing"

func TestRegistryError(t *testing.T) {
	err := &RegistryError{
		Code:    "AUTH_FAILED",
		Message: "authentication failed",
	}

	if err.Error() != "AUTH_FAILED: authentication failed" {
		t.Errorf("Error() = %v", err.Error())
	}

	// Test error type checking
	if err.Code != "AUTH_FAILED" {
		t.Errorf("expected AUTH_FAILED")
	}
}

func TestErrorConstructors(t *testing.T) {
	err1 := NewUnauthorizedError("test")
	err2 := NewNotFoundError("not found")
	err3 := NewConnectionError("conn", "detail")
	if err1 == nil || err2 == nil || err3 == nil {
		t.Error("expected errors")
	}
}