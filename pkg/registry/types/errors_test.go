package types

import (
	"errors"
	"testing"
)

func TestRegistryError(t *testing.T) {
	tests := []struct {
		name           string
		err            *RegistryError
		expectedString string
		expectedCode   string
	}{
		{
			name: "basic error without detail",
			err: &RegistryError{
				Code:    "AUTH_FAILED",
				Message: "authentication failed",
			},
			expectedString: "AUTH_FAILED: authentication failed",
			expectedCode:   "AUTH_FAILED",
		},
		{
			name: "error with detail",
			err: &RegistryError{
				Code:    ErrCodeConnectionFailed,
				Message: "connection timeout",
				Detail:  "network unreachable",
			},
			expectedString: "CONNECTION_FAILED: connection timeout (detail: network unreachable)",
			expectedCode:   ErrCodeConnectionFailed,
		},
		{
			name: "error with complex detail",
			err: &RegistryError{
				Code:    ErrCodeTLSVerification,
				Message: "certificate verification failed",
				Detail:  map[string]string{"cert": "expired", "host": "registry.example.com"},
			},
			expectedString: "TLS_VERIFICATION_FAILED: certificate verification failed (detail: map[cert:expired host:registry.example.com])",
			expectedCode:   ErrCodeTLSVerification,
		},
		{
			name: "empty message - edge case",
			err: &RegistryError{
				Code:    ErrCodeInvalidConfig,
				Message: "",
			},
			expectedString: "INVALID_CONFIG: ",
			expectedCode:   ErrCodeInvalidConfig,
		},
		{
			name: "empty code - edge case",
			err: &RegistryError{
				Code:    "",
				Message: "unknown error",
			},
			expectedString: ": unknown error",
			expectedCode:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Error() method
			if tt.err.Error() != tt.expectedString {
				t.Errorf("Error() = %q, want %q", tt.err.Error(), tt.expectedString)
			}

			// Test Code field
			if tt.err.Code != tt.expectedCode {
				t.Errorf("Code = %q, want %q", tt.err.Code, tt.expectedCode)
			}

			// Test that it implements error interface
			var _ error = tt.err
		})
	}
}

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name         string
		constructor  func() error
		expectedCode string
		message      string
		hasDetail    bool
	}{
		{
			name: "unauthorized error",
			constructor: func() error {
				return NewUnauthorizedError("invalid credentials")
			},
			expectedCode: ErrCodeUnauthorized,
			message:      "invalid credentials",
			hasDetail:    false,
		},
		{
			name: "not found error",
			constructor: func() error {
				return NewNotFoundError("repository not found")
			},
			expectedCode: ErrCodeNotFound,
			message:      "repository not found",
			hasDetail:    false,
		},
		{
			name: "connection error with detail",
			constructor: func() error {
				return NewConnectionError("network timeout", "connection refused")
			},
			expectedCode: ErrCodeConnectionFailed,
			message:      "network timeout",
			hasDetail:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor()

			// Test that error is not nil
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			// Test error type
			regErr, ok := err.(*RegistryError)
			if !ok {
				t.Fatalf("expected *RegistryError, got %T", err)
			}

			// Test error code
			if regErr.Code != tt.expectedCode {
				t.Errorf("Code = %q, want %q", regErr.Code, tt.expectedCode)
			}

			// Test error message
			if regErr.Message != tt.message {
				t.Errorf("Message = %q, want %q", regErr.Message, tt.message)
			}

			// Test detail presence
			hasDetail := regErr.Detail != nil
			if hasDetail != tt.hasDetail {
				t.Errorf("has detail = %v, want %v", hasDetail, tt.hasDetail)
			}

			// Test that Error() method includes the code
			errorString := err.Error()
			if errorString == "" {
				t.Error("Error() returned empty string")
			}
		})
	}
}

func TestErrorConstants(t *testing.T) {
	expectedCodes := map[string]string{
		"UNAUTHORIZED":           ErrCodeUnauthorized,
		"FORBIDDEN":              ErrCodeForbidden,
		"NOT_FOUND":              ErrCodeNotFound,
		"TIMEOUT":                ErrCodeTimeout,
		"CONNECTION_FAILED":      ErrCodeConnectionFailed,
		"INVALID_CONFIG":         ErrCodeInvalidConfig,
		"TLS_VERIFICATION_FAILED": ErrCodeTLSVerification,
	}

	for expected, actual := range expectedCodes {
		if actual != expected {
			t.Errorf("error code constant = %q, want %q", actual, expected)
		}
		if actual == "" {
			t.Errorf("error code constant should not be empty")
		}
	}
}

func TestErrorWrapping(t *testing.T) {
	// Test that RegistryError can be used with errors.Is()
	baseErr := &RegistryError{Code: ErrCodeNotFound, Message: "test"}
	wrappedErr := &RegistryError{
		Code:    ErrCodeConnectionFailed,
		Message: "connection failed",
		Detail:  baseErr,
	}

	if !errors.Is(wrappedErr, wrappedErr) {
		t.Error("error should be equal to itself")
	}

	// Test nil error edge case
	var nilErr *RegistryError
	if nilErr != nil {
		t.Error("nil RegistryError should be nil")
	}

	// Test that error can be compared
	err1 := &RegistryError{Code: "TEST", Message: "test"}
	err2 := &RegistryError{Code: "TEST", Message: "test"}

	if err1.Error() != err2.Error() {
		t.Error("identical errors should produce same string")
	}
}

func TestErrorMethods(t *testing.T) {
	err := &RegistryError{
		Code:    ErrCodeTimeout,
		Message: "operation timed out",
		Detail:  30.5, // numeric detail
	}

	// Test that it implements the error interface properly
	var iface error = err
	if iface.Error() != err.Error() {
		t.Error("error interface implementation mismatch")
	}

	// Test string representation
	expected := "TIMEOUT: operation timed out (detail: 30.5)"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}
