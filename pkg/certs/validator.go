package certs

import (
	"crypto/x509"
	"time"
)

// ValidationMode defines the strictness of certificate validation
type ValidationMode int

const (
	// StrictMode enforces all certificate validation rules
	StrictMode ValidationMode = iota
	// LenientMode allows some common certificate issues
	LenientMode
	// PermissiveMode allows most certificate issues for development
	PermissiveMode
)

// String returns the string representation of ValidationMode
func (v ValidationMode) String() string {
	switch v {
	case StrictMode:
		return "STRICT"
	case LenientMode:
		return "LENIENT"
	case PermissiveMode:
		return "PERMISSIVE"
	default:
		return "UNKNOWN"
	}
}

// CertificateValidator defines the interface for certificate validation operations
type CertificateValidator interface {
	// ValidateChain validates a complete certificate chain from leaf to root
	ValidateChain(certs []*x509.Certificate) error
	
	// ValidateCertificate validates a single certificate
	ValidateCertificate(cert *x509.Certificate) error
	
	// VerifyHostname verifies that a certificate is valid for a given hostname
	VerifyHostname(cert *x509.Certificate, hostname string) error
	
	// GenerateDiagnostics creates diagnostic information for troubleshooting
	GenerateDiagnostics() (*CertDiagnostics, error)
	
	// SetValidationMode changes the validation strictness
	SetValidationMode(mode ValidationMode)
	
	// GetValidationMode returns the current validation mode
	GetValidationMode() ValidationMode
}

// TrustStoreProvider defines an interface for accessing trusted root certificates
// This will integrate with Wave 1's TrustStoreManager when available
type TrustStoreProvider interface {
	// GetRootCAs returns the pool of trusted root CA certificates
	GetRootCAs() (*x509.CertPool, error)
	
	// IsRootTrusted checks if a certificate is a trusted root CA
	IsRootTrusted(cert *x509.Certificate) bool
}

// DefaultCertificateValidator implements the CertificateValidator interface
type DefaultCertificateValidator struct {
	mode             ValidationMode
	trustStore       TrustStoreProvider
	lastValidation   *ValidationResult
	diagnostics      *CertDiagnostics
}

// ValidationResult holds the results of a validation operation
type ValidationResult struct {
	IsValid      bool
	Errors       []error
	Warnings     []string
	ValidatedAt  time.Time
	Certificate  *x509.Certificate
	Chain        []*x509.Certificate
}

// NewDefaultCertificateValidator creates a new certificate validator
func NewDefaultCertificateValidator(trustStore TrustStoreProvider) *DefaultCertificateValidator {
	return &DefaultCertificateValidator{
		mode:        StrictMode,
		trustStore:  trustStore,
		diagnostics: &CertDiagnostics{},
	}
}

// ValidateChain validates a certificate chain from leaf to root
func (v *DefaultCertificateValidator) ValidateChain(certs []*x509.Certificate) error {
	if len(certs) == 0 {
		return NewValidationError(InvalidChain, "certificate chain is empty")
	}

	// Reset diagnostics for new validation
	v.diagnostics = &CertDiagnostics{
		ValidationStarted: time.Now(),
		ChainLength:      len(certs),
	}

	var validationErrors []error
	
	// Validate leaf certificate first
	leafCert := certs[0]
	if err := v.ValidateCertificate(leafCert); err != nil {
		validationErrors = append(validationErrors, err)
	}

	// Build and validate the chain
	chainValidator := NewChainValidator(v.trustStore, v.mode)
	if err := chainValidator.ValidateChain(certs); err != nil {
		validationErrors = append(validationErrors, err)
	}

	// Store validation result
	v.lastValidation = &ValidationResult{
		IsValid:     len(validationErrors) == 0,
		Errors:      validationErrors,
		ValidatedAt: time.Now(),
		Certificate: leafCert,
		Chain:       certs,
	}

	// Update diagnostics
	v.updateDiagnostics(leafCert, certs, validationErrors)

	if len(validationErrors) > 0 {
		return NewAggregatedValidationError(validationErrors)
	}

	return nil
}

// ValidateCertificate validates a single certificate
func (v *DefaultCertificateValidator) ValidateCertificate(cert *x509.Certificate) error {
	if cert == nil {
		return NewValidationError(InvalidCertificate, "certificate is nil")
	}

	var errors []error

	// Check time validity
	now := time.Now()
	if now.Before(cert.NotBefore) {
		if v.mode == StrictMode {
			errors = append(errors, NewValidationError(NotYetValid, "certificate not yet valid"))
		}
	}
	
	if now.After(cert.NotAfter) {
		if v.mode != PermissiveMode {
			errors = append(errors, NewValidationError(Expired, "certificate has expired"))
		}
	}

	// Check basic constraints and key usage
	if err := v.validateConstraints(cert); err != nil && v.mode == StrictMode {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return NewAggregatedValidationError(errors)
	}

	return nil
}

// VerifyHostname verifies hostname against certificate
func (v *DefaultCertificateValidator) VerifyHostname(cert *x509.Certificate, hostname string) error {
	if cert == nil {
		return NewValidationError(InvalidCertificate, "certificate is nil")
	}

	if hostname == "" {
		return NewValidationError(InvalidInput, "hostname cannot be empty")
	}

	// Use Go's built-in hostname verification
	if err := cert.VerifyHostname(hostname); err != nil {
		return NewValidationError(HostnameMismatch, err.Error())
	}

	return nil
}

// GenerateDiagnostics creates diagnostic information
func (v *DefaultCertificateValidator) GenerateDiagnostics() (*CertDiagnostics, error) {
	if v.diagnostics == nil {
		return &CertDiagnostics{
			Message: "No validation has been performed yet",
		}, nil
	}

	return v.diagnostics, nil
}

// SetValidationMode changes the validation strictness
func (v *DefaultCertificateValidator) SetValidationMode(mode ValidationMode) {
	v.mode = mode
}

// GetValidationMode returns the current validation mode
func (v *DefaultCertificateValidator) GetValidationMode() ValidationMode {
	return v.mode
}

// validateConstraints checks basic constraints and key usage
func (v *DefaultCertificateValidator) validateConstraints(cert *x509.Certificate) error {
	// Check if certificate can be used for TLS
	if cert.KeyUsage != 0 {
		if cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 && 
		   cert.KeyUsage&x509.KeyUsageKeyEncipherment == 0 {
			return NewValidationError(InvalidKeyUsage, "certificate lacks required key usage for TLS")
		}
	}

	// Check extended key usage for server certificates
	if len(cert.ExtKeyUsage) > 0 {
		hasServerAuth := false
		for _, usage := range cert.ExtKeyUsage {
			if usage == x509.ExtKeyUsageServerAuth {
				hasServerAuth = true
				break
			}
		}
		
		if !hasServerAuth && v.mode == StrictMode {
			return NewValidationError(InvalidExtKeyUsage, "certificate lacks server authentication extended key usage")
		}
	}

	return nil
}

// updateDiagnostics updates the diagnostic information after validation
func (v *DefaultCertificateValidator) updateDiagnostics(cert *x509.Certificate, chain []*x509.Certificate, errors []error) {
	v.diagnostics.Subject = cert.Subject.String()
	v.diagnostics.Issuer = cert.Issuer.String()
	v.diagnostics.NotBefore = cert.NotBefore
	v.diagnostics.NotAfter = cert.NotAfter
	v.diagnostics.ChainLength = len(chain)
	v.diagnostics.IsExpired = time.Now().After(cert.NotAfter)
	v.diagnostics.IsSelfSigned = cert.Subject.String() == cert.Issuer.String()
	
	// Convert errors to string messages
	v.diagnostics.ValidationErrors = make([]string, len(errors))
	for i, err := range errors {
		v.diagnostics.ValidationErrors[i] = err.Error()
	}
	
	v.diagnostics.ValidationCompleted = time.Now()
}