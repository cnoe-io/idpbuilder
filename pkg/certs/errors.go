package certs

import (
	"errors"
	"fmt"
)

// Certificate operation errors
var (
	// Extraction errors
	ErrNoKindCluster    = errors.New("no Kind cluster found")
	ErrGiteaPodNotFound = errors.New("Gitea pod not found in cluster")
	ErrCertNotInPod     = errors.New("certificate not found in pod")
	ErrInvalidCertData  = errors.New("invalid certificate data")

	// Storage errors
	ErrCertNotFound      = errors.New("certificate not found in storage")
	ErrStoragePermission = errors.New("insufficient permissions for certificate storage")
	ErrStorageFull       = errors.New("certificate storage is full")

	// Validation errors
	ErrCertExpired         = errors.New("certificate has expired")
	ErrCertNotYetValid     = errors.New("certificate is not yet valid")
	ErrCertInvalidKeyUsage = errors.New("certificate has invalid key usage")
	ErrCertSelfSigned      = errors.New("certificate is self-signed")

	// Feature flag errors
	ErrFeatureDisabled = errors.New("certificate extraction feature is disabled")
)

// CertError wraps certificate errors with context
type CertError struct {
	Op      string            // Operation that failed
	Kind    string            // Kind of error
	Err     error             // Underlying error
	Context map[string]string // Additional context
}

// Error implements the error interface
func (e *CertError) Error() string {
	if e.Context != nil && len(e.Context) > 0 {
		return fmt.Sprintf("%s: %s: %v (context: %v)", e.Op, e.Kind, e.Err, e.Context)
	}
	return fmt.Sprintf("%s: %s: %v", e.Op, e.Kind, e.Err)
}

// Unwrap returns the underlying error
func (e *CertError) Unwrap() error {
	return e.Err
}

// NewCertError creates a new certificate error
func NewCertError(op, kind string, err error) *CertError {
	return &CertError{
		Op:   op,
		Kind: kind,
		Err:  err,
	}
}

// NewCertErrorWithContext creates a new certificate error with context
func NewCertErrorWithContext(op, kind string, err error, context map[string]string) *CertError {
	return &CertError{
		Op:      op,
		Kind:    kind,
		Err:     err,
		Context: context,
	}
}
