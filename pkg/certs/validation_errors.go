package certs

import (
	"fmt"
	"strings"
)

// ValidationErrorType defines the category of validation error
type ValidationErrorType int

const (
	// InvalidCertificate indicates the certificate is malformed or nil
	InvalidCertificate ValidationErrorType = iota
	
	// Expired indicates the certificate has expired
	Expired
	
	// NotYetValid indicates the certificate is not yet valid
	NotYetValid
	
	// UntrustedRoot indicates the root certificate is not trusted
	UntrustedRoot
	
	// InvalidChain indicates issues with the certificate chain structure
	InvalidChain
	
	// InvalidSignature indicates signature verification failed
	InvalidSignature
	
	// HostnameMismatch indicates the certificate doesn't match the hostname
	HostnameMismatch
	
	// InvalidKeyUsage indicates invalid key usage for the certificate
	InvalidKeyUsage
	
	// InvalidExtKeyUsage indicates invalid extended key usage
	InvalidExtKeyUsage
	
	// InvalidInput indicates invalid input parameters
	InvalidInput
)

// String returns the string representation of ValidationErrorType
func (t ValidationErrorType) String() string {
	switch t {
	case InvalidCertificate:
		return "INVALID_CERTIFICATE"
	case Expired:
		return "EXPIRED"
	case NotYetValid:
		return "NOT_YET_VALID"
	case UntrustedRoot:
		return "UNTRUSTED_ROOT"
	case InvalidChain:
		return "INVALID_CHAIN"
	case InvalidSignature:
		return "INVALID_SIGNATURE"
	case HostnameMismatch:
		return "HOSTNAME_MISMATCH"
	case InvalidKeyUsage:
		return "INVALID_KEY_USAGE"
	case InvalidExtKeyUsage:
		return "INVALID_EXT_KEY_USAGE"
	case InvalidInput:
		return "INVALID_INPUT"
	default:
		return "UNKNOWN"
	}
}

// ValidationError represents a certificate validation error
type ValidationError struct {
	Type    ValidationErrorType
	Message string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type.String(), e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(errorType ValidationErrorType, message string) *ValidationError {
	return &ValidationError{
		Type:    errorType,
		Message: message,
	}
}

// AggregatedValidationError represents multiple validation errors
type AggregatedValidationError struct {
	Errors []error
}

// Error implements the error interface
func (e *AggregatedValidationError) Error() string {
	if len(e.Errors) == 0 {
		return "no validation errors"
	}
	
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	
	var messages []string
	for _, err := range e.Errors {
		messages = append(messages, err.Error())
	}
	
	return fmt.Sprintf("multiple validation errors: %s", strings.Join(messages, "; "))
}

// NewAggregatedValidationError creates a new aggregated validation error
func NewAggregatedValidationError(errors []error) *AggregatedValidationError {
	return &AggregatedValidationError{
		Errors: errors,
	}
}