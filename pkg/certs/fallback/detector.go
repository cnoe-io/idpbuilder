// Package fallback provides certificate error detection and fallback handling
// for when TLS operations fail due to certificate issues.
package fallback

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"
	"time"
)

// CertErrorType represents different types of certificate errors
type CertErrorType int

const (
	// ErrorTypeUnknown represents an unclassified error
	ErrorTypeUnknown CertErrorType = iota
	
	// ErrorTypeExpired indicates the certificate has expired
	ErrorTypeExpired
	
	// ErrorTypeNotYetValid indicates the certificate is not yet valid
	ErrorTypeNotYetValid
	
	// ErrorTypeHostnameMismatch indicates the hostname doesn't match the certificate
	ErrorTypeHostnameMismatch
	
	// ErrorTypeUntrustedCA indicates the certificate authority is not trusted
	ErrorTypeUntrustedCA
	
	// ErrorTypeSelfSigned indicates the certificate is self-signed
	ErrorTypeSelfSigned
	
	// ErrorTypeBadCertificate indicates a malformed or invalid certificate
	ErrorTypeBadCertificate
	
	// ErrorTypeChainIncomplete indicates the certificate chain is incomplete
	ErrorTypeChainIncomplete
	
	// ErrorTypeRevoked indicates the certificate has been revoked
	ErrorTypeRevoked
	
	// ErrorTypeKeyUsage indicates improper key usage
	ErrorTypeKeyUsage
)

// String returns a human-readable representation of the error type
func (e CertErrorType) String() string {
	switch e {
	case ErrorTypeExpired:
		return "expired"
	case ErrorTypeNotYetValid:
		return "not_yet_valid"
	case ErrorTypeHostnameMismatch:
		return "hostname_mismatch"
	case ErrorTypeUntrustedCA:
		return "untrusted_ca"
	case ErrorTypeSelfSigned:
		return "self_signed"
	case ErrorTypeBadCertificate:
		return "bad_certificate"
	case ErrorTypeChainIncomplete:
		return "chain_incomplete"
	case ErrorTypeRevoked:
		return "revoked"
	case ErrorTypeKeyUsage:
		return "key_usage"
	default:
		return "unknown"
	}
}

// ErrorDetails contains detailed information about a certificate error
type ErrorDetails struct {
	// Type is the classified error type
	Type CertErrorType
	
	// OriginalError is the underlying error that occurred
	OriginalError error
	
	// Certificate is the problematic certificate (if available)
	Certificate *x509.Certificate
	
	// Hostname is the hostname that was being verified
	Hostname string
	
	// Chain contains the certificate chain (if available)
	Chain []*x509.Certificate
	
	// Timestamp is when the error was detected
	Timestamp time.Time
	
	// Message is a human-readable description of the error
	Message string
	
	// IsRecoverable indicates whether this error might be recoverable with fallback
	IsRecoverable bool
	
	// SuggestedActions contains recommended actions to resolve the error
	SuggestedActions []string
}

// CertErrorDetector interface defines methods for detecting and classifying certificate errors
type CertErrorDetector interface {
	// DetectError analyzes a TLS error and returns detailed error information
	DetectError(err error, hostname string) (*ErrorDetails, error)
	
	// ClassifyError determines the type of certificate error
	ClassifyError(err error) CertErrorType
	
	// ValidateChain performs validation on a certificate chain
	ValidateChain(chain []*x509.Certificate, hostname string) *ErrorDetails
	
	// IsTrustedCA checks if the certificate authority is trusted
	IsTrustedCA(cert *x509.Certificate) bool
	
	// ExtractCertFromError attempts to extract certificate information from an error
	ExtractCertFromError(err error) (*x509.Certificate, []*x509.Certificate)
	
	// IsRecoverable determines if an error is potentially recoverable
	IsRecoverable(errorType CertErrorType) bool
}

// DefaultCertErrorDetector implements the CertErrorDetector interface
type DefaultCertErrorDetector struct {
	// systemRoots contains the system's trusted certificate roots
	systemRoots *x509.CertPool
	
	// additionalRoots contains additional trusted roots
	additionalRoots *x509.CertPool
	
	// allowSelfSigned indicates whether self-signed certificates should be considered recoverable
	allowSelfSigned bool
	
	// timeSkewTolerance is the amount of time skew to tolerate for expiry checks
	timeSkewTolerance time.Duration
}

// NewDefaultCertErrorDetector creates a new instance of DefaultCertErrorDetector
func NewDefaultCertErrorDetector() *DefaultCertErrorDetector {
	systemRoots, _ := x509.SystemCertPool()
	if systemRoots == nil {
		systemRoots = x509.NewCertPool()
	}

	return &DefaultCertErrorDetector{
		systemRoots:       systemRoots,
		additionalRoots:   x509.NewCertPool(),
		allowSelfSigned:   true,  // For development environments
		timeSkewTolerance: 5 * time.Minute,
	}
}

// DetectError analyzes a TLS error and returns detailed error information
func (d *DefaultCertErrorDetector) DetectError(err error, hostname string) (*ErrorDetails, error) {
	if err == nil {
		return nil, fmt.Errorf("no error to analyze")
	}

	errorType := d.ClassifyError(err)
	cert, chain := d.ExtractCertFromError(err)
	
	details := &ErrorDetails{
		Type:          errorType,
		OriginalError: err,
		Certificate:   cert,
		Hostname:      hostname,
		Chain:         chain,
		Timestamp:     time.Now(),
		IsRecoverable: d.IsRecoverable(errorType),
	}
	
	// Generate human-readable message and suggestions
	d.populateErrorDetails(details)
	
	return details, nil
}

// ClassifyError determines the type of certificate error
func (d *DefaultCertErrorDetector) ClassifyError(err error) CertErrorType {
	if err == nil {
		return ErrorTypeUnknown
	}
	
	errorStr := err.Error()
	errorLower := strings.ToLower(errorStr)
	
	// Check for specific error patterns
	switch {
	case strings.Contains(errorLower, "certificate has expired"):
		return ErrorTypeExpired
	case strings.Contains(errorLower, "certificate is not valid yet"):
		return ErrorTypeNotYetValid
	case strings.Contains(errorLower, "hostname") && strings.Contains(errorLower, "doesn't match"):
		return ErrorTypeHostnameMismatch
	case strings.Contains(errorLower, "certificate signed by unknown authority"):
		return ErrorTypeUntrustedCA
	case strings.Contains(errorLower, "self signed certificate"):
		return ErrorTypeSelfSigned
	case strings.Contains(errorLower, "bad certificate"):
		return ErrorTypeBadCertificate
	case strings.Contains(errorLower, "certificate chain"):
		return ErrorTypeChainIncomplete
	case strings.Contains(errorLower, "revoked"):
		return ErrorTypeRevoked
	case strings.Contains(errorLower, "key usage"):
		return ErrorTypeKeyUsage
	}
	
	// Check for TLS-specific errors
	if tlsErr, ok := err.(tls.RecordHeaderError); ok {
		_ = tlsErr // Prevent unused variable warning
		return ErrorTypeBadCertificate
	}
	
	// Check for x509 errors
	if certErr, ok := err.(x509.CertificateInvalidError); ok {
		switch certErr.Reason {
		case x509.Expired:
			return ErrorTypeExpired
		case x509.NotAuthorizedToSign:
			return ErrorTypeUntrustedCA
		default:
			return ErrorTypeBadCertificate
		}
	}
	
	// Check for hostname verification errors
	if _, ok := err.(x509.HostnameError); ok {
		return ErrorTypeHostnameMismatch
	}
	
	// Check for unknown authority errors
	if _, ok := err.(x509.UnknownAuthorityError); ok {
		return ErrorTypeUntrustedCA
	}
	
	return ErrorTypeUnknown
}

// ValidateChain performs validation on a certificate chain
func (d *DefaultCertErrorDetector) ValidateChain(chain []*x509.Certificate, hostname string) *ErrorDetails {
	if len(chain) == 0 {
		return &ErrorDetails{
			Type:             ErrorTypeBadCertificate,
			Hostname:         hostname,
			Timestamp:        time.Now(),
			Message:          "Empty certificate chain provided",
			IsRecoverable:    false,
			SuggestedActions: []string{"Verify server configuration", "Check certificate installation"},
		}
	}
	
	leafCert := chain[0]
	now := time.Now()
	
	// Check expiration with time skew tolerance
	if leafCert.NotAfter.Add(d.timeSkewTolerance).Before(now) {
		return &ErrorDetails{
			Type:             ErrorTypeExpired,
			Certificate:      leafCert,
			Chain:            chain,
			Hostname:         hostname,
			Timestamp:        now,
			Message:          fmt.Sprintf("Certificate expired on %s", leafCert.NotAfter.Format(time.RFC3339)),
			IsRecoverable:    false,
			SuggestedActions: []string{"Renew the certificate", "Update certificate on server"},
		}
	}
	
	// Check if certificate is not yet valid
	if leafCert.NotBefore.Add(-d.timeSkewTolerance).After(now) {
		return &ErrorDetails{
			Type:             ErrorTypeNotYetValid,
			Certificate:      leafCert,
			Chain:            chain,
			Hostname:         hostname,
			Timestamp:        now,
			Message:          fmt.Sprintf("Certificate not valid until %s", leafCert.NotBefore.Format(time.RFC3339)),
			IsRecoverable:    true,
			SuggestedActions: []string{"Check system time", "Wait until certificate becomes valid"},
		}
	}
	
	// Check hostname verification
	if hostname != "" {
		if err := leafCert.VerifyHostname(hostname); err != nil {
			return &ErrorDetails{
				Type:             ErrorTypeHostnameMismatch,
				OriginalError:    err,
				Certificate:      leafCert,
				Chain:            chain,
				Hostname:         hostname,
				Timestamp:        now,
				Message:          fmt.Sprintf("Hostname %s doesn't match certificate", hostname),
				IsRecoverable:    true,
				SuggestedActions: []string{"Use correct hostname", "Add hostname to certificate SAN", "Enable hostname verification bypass"},
			}
		}
	}
	
	// Check if certificate is self-signed
	if leafCert.Subject.String() == leafCert.Issuer.String() {
		isRecoverable := d.allowSelfSigned
		return &ErrorDetails{
			Type:             ErrorTypeSelfSigned,
			Certificate:      leafCert,
			Chain:            chain,
			Hostname:         hostname,
			Timestamp:        now,
			Message:          "Certificate is self-signed",
			IsRecoverable:    isRecoverable,
			SuggestedActions: []string{"Add certificate to trust store", "Use CA-signed certificate", "Enable self-signed certificate acceptance"},
		}
	}
	
	// If we get here, the certificate appears valid
	return nil
}

// IsTrustedCA checks if the certificate authority is trusted
func (d *DefaultCertErrorDetector) IsTrustedCA(cert *x509.Certificate) bool {
	if cert == nil {
		return false
	}
	
	// Create verification options
	opts := x509.VerifyOptions{
		Roots:       d.systemRoots,
		CurrentTime: time.Now(),
	}
	
	// Try to verify with system roots
	_, err := cert.Verify(opts)
	if err == nil {
		return true
	}
	
	// Try with additional roots if system verification failed
	if d.additionalRoots != nil {
		opts.Roots = d.additionalRoots
		_, err = cert.Verify(opts)
		return err == nil
	}
	
	return false
}

// ExtractCertFromError attempts to extract certificate information from an error
func (d *DefaultCertErrorDetector) ExtractCertFromError(err error) (*x509.Certificate, []*x509.Certificate) {
	if err == nil {
		return nil, nil
	}
	
	// Try to extract from x509 hostname error
	if hostnameErr, ok := err.(x509.HostnameError); ok {
		return hostnameErr.Certificate, nil
	}
	
	// Try to extract from unknown authority error
	if unknownAuthErr, ok := err.(x509.UnknownAuthorityError); ok {
		return unknownAuthErr.Cert, nil
	}
	
	// Try to extract from certificate invalid error
	if certInvalidErr, ok := err.(x509.CertificateInvalidError); ok {
		return certInvalidErr.Cert, nil
	}
	
	// For other error types, we can't extract certificate information
	return nil, nil
}

// IsRecoverable determines if an error is potentially recoverable
func (d *DefaultCertErrorDetector) IsRecoverable(errorType CertErrorType) bool {
	switch errorType {
	case ErrorTypeHostnameMismatch:
		return true // Can bypass hostname verification
	case ErrorTypeSelfSigned:
		return d.allowSelfSigned // Configurable
	case ErrorTypeUntrustedCA:
		return true // Can add to trust store or bypass
	case ErrorTypeNotYetValid:
		return true // Might resolve with time
	case ErrorTypeChainIncomplete:
		return true // Can potentially complete chain
	case ErrorTypeExpired:
		return false // Cannot recover from expired certificates
	case ErrorTypeBadCertificate:
		return false // Cannot recover from malformed certificates
	case ErrorTypeRevoked:
		return false // Cannot recover from revoked certificates
	case ErrorTypeKeyUsage:
		return false // Cannot recover from key usage violations
	default:
		return false // Unknown errors are not recoverable
	}
}

// populateErrorDetails fills in the message and suggested actions for an error
func (d *DefaultCertErrorDetector) populateErrorDetails(details *ErrorDetails) {
	if details == nil {
		return
	}
	
	switch details.Type {
	case ErrorTypeExpired:
		details.Message = "Certificate has expired"
		details.SuggestedActions = []string{
			"Renew the certificate",
			"Update server configuration",
			"Check certificate validity period",
		}
		
	case ErrorTypeNotYetValid:
		details.Message = "Certificate is not yet valid"
		details.SuggestedActions = []string{
			"Check system time synchronization",
			"Wait until certificate becomes valid",
			"Verify certificate validity period",
		}
		
	case ErrorTypeHostnameMismatch:
		details.Message = fmt.Sprintf("Hostname '%s' doesn't match certificate", details.Hostname)
		details.SuggestedActions = []string{
			"Use the correct hostname from certificate SAN",
			"Add hostname to certificate Subject Alternative Names",
			"Enable hostname verification bypass for development",
		}
		
	case ErrorTypeUntrustedCA:
		details.Message = "Certificate authority is not trusted"
		details.SuggestedActions = []string{
			"Add CA certificate to system trust store",
			"Use a certificate from a trusted CA",
			"Enable untrusted certificate acceptance for development",
		}
		
	case ErrorTypeSelfSigned:
		details.Message = "Certificate is self-signed"
		details.SuggestedActions = []string{
			"Add certificate to trust store",
			"Use a CA-signed certificate",
			"Enable self-signed certificate acceptance",
		}
		
	case ErrorTypeBadCertificate:
		details.Message = "Certificate is malformed or invalid"
		details.SuggestedActions = []string{
			"Replace with a valid certificate",
			"Check certificate format and encoding",
			"Verify certificate installation",
		}
		
	case ErrorTypeChainIncomplete:
		details.Message = "Certificate chain is incomplete"
		details.SuggestedActions = []string{
			"Install intermediate certificates",
			"Complete the certificate chain",
			"Verify certificate chain configuration",
		}
		
	case ErrorTypeRevoked:
		details.Message = "Certificate has been revoked"
		details.SuggestedActions = []string{
			"Replace with a new certificate",
			"Check certificate revocation status",
			"Contact certificate authority",
		}
		
	case ErrorTypeKeyUsage:
		details.Message = "Certificate key usage is inappropriate for this operation"
		details.SuggestedActions = []string{
			"Use a certificate with appropriate key usage",
			"Generate new certificate with correct extensions",
			"Verify certificate purpose and usage",
		}
		
	default:
		details.Message = "Unknown certificate error"
		details.SuggestedActions = []string{
			"Check certificate configuration",
			"Verify TLS connection settings",
			"Review server and client configurations",
		}
	}
	
	// Add hostname-specific context if available
	if details.Hostname != "" && details.Message != "" {
		details.Message += fmt.Sprintf(" (hostname: %s)", details.Hostname)
	}
	
	// Add certificate subject if available
	if details.Certificate != nil {
		details.Message += fmt.Sprintf(" (subject: %s)", details.Certificate.Subject.String())
	}
}

// AddTrustedCA adds a certificate to the additional trusted roots
func (d *DefaultCertErrorDetector) AddTrustedCA(cert *x509.Certificate) error {
	if cert == nil {
		return fmt.Errorf("certificate cannot be nil")
	}
	
	if d.additionalRoots == nil {
		d.additionalRoots = x509.NewCertPool()
	}
	
	d.additionalRoots.AddCert(cert)
	return nil
}

// SetSelfSignedAllowed configures whether self-signed certificates are considered recoverable
func (d *DefaultCertErrorDetector) SetSelfSignedAllowed(allowed bool) {
	d.allowSelfSigned = allowed
}

// SetTimeSkewTolerance configures the time skew tolerance for certificate validation
func (d *DefaultCertErrorDetector) SetTimeSkewTolerance(tolerance time.Duration) {
	d.timeSkewTolerance = tolerance
}