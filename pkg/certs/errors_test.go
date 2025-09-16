package certs

import (
	"errors"
	"testing"
)

func TestCertError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *CertError
		expected string
	}{
		{
			name: "error without context",
			err: &CertError{
				Op:   "extraction",
				Kind: "file_copy",
				Err:  errors.New("file not found"),
			},
			expected: "extraction: file_copy: file not found",
		},
		{
			name: "error with context",
			err: &CertError{
				Op:   "storage",
				Kind: "permission",
				Err:  errors.New("access denied"),
				Context: map[string]string{
					"path": "/tmp/certs",
					"user": "testuser",
				},
			},
			expected: "storage: permission: access denied (context: map[path:/tmp/certs user:testuser])",
		},
		{
			name: "error with empty context",
			err: &CertError{
				Op:      "validation",
				Kind:    "expiry",
				Err:     errors.New("expired"),
				Context: map[string]string{},
			},
			expected: "validation: expiry: expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("CertError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCertError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	certErr := &CertError{
		Op:   "test",
		Kind: "test",
		Err:  originalErr,
	}

	unwrapped := certErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("CertError.Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

func TestNewCertError(t *testing.T) {
	originalErr := errors.New("test error")
	certErr := NewCertError("operation", "kind", originalErr)

	if certErr.Op != "operation" {
		t.Errorf("NewCertError().Op = %v, want %v", certErr.Op, "operation")
	}
	if certErr.Kind != "kind" {
		t.Errorf("NewCertError().Kind = %v, want %v", certErr.Kind, "kind")
	}
	if certErr.Err != originalErr {
		t.Errorf("NewCertError().Err = %v, want %v", certErr.Err, originalErr)
	}
	if certErr.Context != nil {
		t.Errorf("NewCertError().Context = %v, want nil", certErr.Context)
	}
}

func TestNewCertErrorWithContext(t *testing.T) {
	originalErr := errors.New("test error")
	context := map[string]string{"key": "value"}
	certErr := NewCertErrorWithContext("operation", "kind", originalErr, context)

	if certErr.Op != "operation" {
		t.Errorf("NewCertErrorWithContext().Op = %v, want %v", certErr.Op, "operation")
	}
	if certErr.Kind != "kind" {
		t.Errorf("NewCertErrorWithContext().Kind = %v, want %v", certErr.Kind, "kind")
	}
	if certErr.Err != originalErr {
		t.Errorf("NewCertErrorWithContext().Err = %v, want %v", certErr.Err, originalErr)
	}
	if len(certErr.Context) != 1 || certErr.Context["key"] != "value" {
		t.Errorf("NewCertErrorWithContext().Context = %v, want %v", certErr.Context, context)
	}
}

func TestPredefinedErrors(t *testing.T) {
	errors := []error{
		ErrNoKindCluster,
		ErrGiteaPodNotFound,
		ErrCertNotInPod,
		ErrInvalidCertData,
		ErrCertNotFound,
		ErrStoragePermission,
		ErrStorageFull,
		ErrCertExpired,
		ErrCertNotYetValid,
		ErrCertInvalidKeyUsage,
		ErrCertSelfSigned,
	}

	for i, err := range errors {
		if err == nil {
			t.Errorf("Predefined error at index %d is nil", i)
		}
		if err.Error() == "" {
			t.Errorf("Predefined error at index %d has empty message", i)
		}
	}
}
