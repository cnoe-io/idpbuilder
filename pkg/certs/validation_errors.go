package certs

import (
	"fmt"
	"strings"
	"time"
)

// ValidationErrorType represents the type of certificate validation error
type ValidationErrorType int

const (
	// InvalidCertificate indicates the certificate is malformed or invalid
	InvalidCertificate ValidationErrorType = iota
	
	// Expired indicates the certificate has expired
	Expired
	
	// NotYetValid indicates the certificate is not yet valid
	NotYetValid
	
	// UntrustedCA indicates the certificate authority is not trusted
	UntrustedCA
	
	// HostnameMismatch indicates the certificate hostname does not match
	HostnameMismatch
	
	// KeyUsageMismatch indicates the certificate key usage is incorrect
	KeyUsageMismatch
	
	// ExtendedKeyUsageMismatch indicates extended key usage is incorrect
	ExtendedKeyUsageMismatch
	
	// ChainTooLong indicates the certificate chain exceeds maximum length
	ChainTooLong
	
	// ChainIncomplete indicates the certificate chain is incomplete
	ChainIncomplete
	
	// SignatureVerificationFailed indicates signature verification failed
	SignatureVerificationFailed
	
	// WeakSignatureAlgorithm indicates the signature algorithm is weak
	WeakSignatureAlgorithm
	
	// InvalidKeySize indicates the key size is invalid or too small
	InvalidKeySize
	
	// CertificateRevoked indicates the certificate has been revoked
	CertificateRevoked
	
	// PolicyConstraintViolation indicates certificate policy constraints violated
	PolicyConstraintViolation
	
	// NameConstraintViolation indicates certificate name constraints violated
	NameConstraintViolation
	
	// PathLengthExceeded indicates path length constraint exceeded
	PathLengthExceeded
	
	// UnknownCriticalExtension indicates an unknown critical extension
	UnknownCriticalExtension
)

// ValidationError represents a certificate validation error with detailed information
type ValidationError struct {
	Type        ValidationErrorType
	Message     string
	Certificate string // Subject or identifier of the certificate
	Timestamp   time.Time
	Details     map[string]interface{}
}

// NewValidationError creates a new ValidationError
func NewValidationError(errorType ValidationErrorType, message, certificate string) *ValidationError {
	return &ValidationError{
		Type:        errorType,
		Message:     message,
		Certificate: certificate,
		Timestamp:   time.Now(),
		Details:     make(map[string]interface{}),
	}
}

// Error implements the error interface
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("certificate validation failed: %s - %s (certificate: %s)",
		ve.Type.String(), ve.Message, ve.Certificate)
}

// String returns a human-readable string representation of the ValidationError
func (ve *ValidationError) String() string {
	var details []string
	for key, value := range ve.Details {
		details = append(details, fmt.Sprintf("%s: %v", key, value))
	}
	
	result := fmt.Sprintf("ValidationError{Type: %s, Message: %s, Certificate: %s, Timestamp: %s",
		ve.Type.String(), ve.Message, ve.Certificate, ve.Timestamp.Format(time.RFC3339))
	
	if len(details) > 0 {
		result += fmt.Sprintf(", Details: {%s}", strings.Join(details, ", "))
	}
	
	return result + "}"
}

// AddDetail adds additional detail information to the error
func (ve *ValidationError) AddDetail(key string, value interface{}) {
	ve.Details[key] = value
}

// String returns the string representation of ValidationErrorType
func (vet ValidationErrorType) String() string {
	switch vet {
	case InvalidCertificate:
		return "InvalidCertificate"
	case Expired:
		return "Expired"
	case NotYetValid:
		return "NotYetValid"
	case UntrustedCA:
		return "UntrustedCA"
	case HostnameMismatch:
		return "HostnameMismatch"
	case KeyUsageMismatch:
		return "KeyUsageMismatch"
	case ExtendedKeyUsageMismatch:
		return "ExtendedKeyUsageMismatch"
	case ChainTooLong:
		return "ChainTooLong"
	case ChainIncomplete:
		return "ChainIncomplete"
	case SignatureVerificationFailed:
		return "SignatureVerificationFailed"
	case WeakSignatureAlgorithm:
		return "WeakSignatureAlgorithm"
	case InvalidKeySize:
		return "InvalidKeySize"
	case CertificateRevoked:
		return "CertificateRevoked"
	case PolicyConstraintViolation:
		return "PolicyConstraintViolation"
	case NameConstraintViolation:
		return "NameConstraintViolation"
	case PathLengthExceeded:
		return "PathLengthExceeded"
	case UnknownCriticalExtension:
		return "UnknownCriticalExtension"
	default:
		return fmt.Sprintf("Unknown(%d)", int(vet))
	}
}

// IsTemporary returns true if the error might be temporary and retry could succeed
func (vet ValidationErrorType) IsTemporary() bool {
	return vet == CertificateRevoked // Only revocation checks might be temporary network issues
}

// IsFatal returns true if the error is fatal and no retry should be attempted
func (vet ValidationErrorType) IsFatal() bool {
	switch vet {
	case InvalidCertificate, Expired, UntrustedCA, HostnameMismatch,
		KeyUsageMismatch, ExtendedKeyUsageMismatch, ChainTooLong,
		SignatureVerificationFailed, WeakSignatureAlgorithm, InvalidKeySize,
		PolicyConstraintViolation, NameConstraintViolation, PathLengthExceeded,
		UnknownCriticalExtension:
		return true
	default:
		return false
	}
}