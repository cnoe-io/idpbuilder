package certs

import (
	"crypto/x509"
	"errors"
	"fmt"
)

// ChainValidator handles certificate chain validation logic
type ChainValidator struct {
	trustStore TrustStoreProvider
	mode       ValidationMode
}

// ChainValidationOptions configures chain validation behavior
type ChainValidationOptions struct {
	// AllowSelfSigned permits self-signed certificates
	AllowSelfSigned bool
	
	// RequireLeafInChain ensures the leaf certificate is present
	RequireLeafInChain bool
	
	// MaxChainLength limits the maximum chain depth
	MaxChainLength int
}

// NewChainValidator creates a new chain validator
func NewChainValidator(trustStore TrustStoreProvider, mode ValidationMode) *ChainValidator {
	return &ChainValidator{
		trustStore: trustStore,
		mode:       mode,
	}
}

// ValidateChain validates the complete certificate chain
func (cv *ChainValidator) ValidateChain(certs []*x509.Certificate) error {
	if len(certs) == 0 {
		return NewValidationError(InvalidChain, "certificate chain is empty")
	}

	// Get default validation options based on mode
	opts := cv.getValidationOptions()
	
	// Validate chain length
	if len(certs) > opts.MaxChainLength {
		return NewValidationError(InvalidChain, 
			fmt.Sprintf("certificate chain too long: %d certificates (max: %d)", 
				len(certs), opts.MaxChainLength))
	}

	// Validate each certificate in the chain
	var validationErrors []error
	
	for i, cert := range certs {
		if err := cv.validateCertificateInChain(cert, i, len(certs)); err != nil {
			validationErrors = append(validationErrors, err)
		}
	}

	// Validate chain structure and signatures
	if err := cv.validateChainStructure(certs, opts); err != nil {
		validationErrors = append(validationErrors, err)
	}

	// Validate trust path to root
	if err := cv.validateTrustPath(certs, opts); err != nil {
		validationErrors = append(validationErrors, err)
	}

	if len(validationErrors) > 0 {
		return NewAggregatedValidationError(validationErrors)
	}

	return nil
}

// validateCertificateInChain validates a single certificate within the chain context
func (cv *ChainValidator) validateCertificateInChain(cert *x509.Certificate, position int, chainLength int) error {
	if cert == nil {
		return NewValidationError(InvalidCertificate, 
			fmt.Sprintf("certificate at position %d is nil", position))
	}

	// Position-specific validation
	if position == 0 {
		// Leaf certificate - should not be a CA unless self-signed
		if cert.IsCA && cert.Subject.String() != cert.Issuer.String() {
			if cv.mode == StrictMode {
				return NewValidationError(InvalidChain, "leaf certificate is marked as CA")
			}
		}
	} else {
		// Intermediate/root certificates - should be CAs
		if !cert.IsCA {
			return NewValidationError(InvalidChain, 
				fmt.Sprintf("certificate at position %d is not a CA but appears in chain", position))
		}
	}

	// Check basic constraints path length
	if cert.IsCA && cert.BasicConstraintsValid {
		if cert.MaxPathLen >= 0 {
			remainingDepth := chainLength - position - 1
			if remainingDepth > cert.MaxPathLen {
				return NewValidationError(InvalidChain, 
					fmt.Sprintf("basic constraints path length exceeded at position %d", position))
			}
		}
	}

	return nil
}

// validateChainStructure validates the structural integrity of the chain
func (cv *ChainValidator) validateChainStructure(certs []*x509.Certificate, opts *ChainValidationOptions) error {
	if len(certs) < 2 && !opts.AllowSelfSigned {
		return NewValidationError(InvalidChain, "certificate chain too short (need at least issuer)")
	}

	// Validate parent-child relationships
	for i := 0; i < len(certs)-1; i++ {
		cert := certs[i]
		issuer := certs[i+1]

		// Check if issuer actually issued the certificate
		if cert.Issuer.String() != issuer.Subject.String() {
			return NewValidationError(InvalidChain, 
				fmt.Sprintf("certificate at position %d issuer does not match certificate at position %d subject", 
					i, i+1))
		}

		// Verify signature
		if err := cert.CheckSignatureFrom(issuer); err != nil {
			return NewValidationError(InvalidSignature, 
				fmt.Sprintf("signature verification failed for certificate at position %d: %v", i, err))
		}
	}

	// Handle self-signed case
	if len(certs) == 1 {
		cert := certs[0]
		if cert.Subject.String() == cert.Issuer.String() {
			// Self-signed certificate
			if !opts.AllowSelfSigned {
				return NewValidationError(UntrustedRoot, "self-signed certificates not allowed")
			}
			
			// Verify self-signature
			if err := cert.CheckSignatureFrom(cert); err != nil {
				return NewValidationError(InvalidSignature, 
					fmt.Sprintf("self-signature verification failed: %v", err))
			}
		}
	}

	return nil
}

// validateTrustPath validates the path to a trusted root
func (cv *ChainValidator) validateTrustPath(certs []*x509.Certificate, opts *ChainValidationOptions) error {
	if cv.trustStore == nil {
		// No trust store available - rely on system default or allow all in permissive mode
		if cv.mode == PermissiveMode {
			return nil
		}
		return NewValidationError(UntrustedRoot, "no trust store configured")
	}

	// Check if root certificate is trusted
	rootCert := certs[len(certs)-1]
	
	// Handle self-signed certificates
	if len(certs) == 1 && rootCert.Subject.String() == rootCert.Issuer.String() {
		if opts.AllowSelfSigned {
			// In permissive mode, allow any self-signed certificate
			if cv.mode == PermissiveMode {
				return nil
			}
			// In other modes, check if it's explicitly trusted
			if cv.trustStore.IsRootTrusted(rootCert) {
				return nil
			}
		}
		return NewValidationError(UntrustedRoot, "self-signed certificate is not trusted")
	}

	// Check if root is in trust store
	if !cv.trustStore.IsRootTrusted(rootCert) {
		return NewValidationError(UntrustedRoot, "root certificate is not trusted")
	}

	return nil
}

// getValidationOptions returns validation options based on the current mode
func (cv *ChainValidator) getValidationOptions() *ChainValidationOptions {
	switch cv.mode {
	case StrictMode:
		return &ChainValidationOptions{
			AllowSelfSigned:    false,
			RequireLeafInChain: true,
			MaxChainLength:     10,
		}
	case LenientMode:
		return &ChainValidationOptions{
			AllowSelfSigned:    true,
			RequireLeafInChain: true,
			MaxChainLength:     15,
		}
	case PermissiveMode:
		return &ChainValidationOptions{
			AllowSelfSigned:    true,
			RequireLeafInChain: false,
			MaxChainLength:     20,
		}
	default:
		return &ChainValidationOptions{
			AllowSelfSigned:    false,
			RequireLeafInChain: true,
			MaxChainLength:     10,
		}
	}
}

// VerifyChainWithOptions validates a chain with custom options
func (cv *ChainValidator) VerifyChainWithOptions(certs []*x509.Certificate, opts *ChainValidationOptions) error {
	if opts == nil {
		opts = cv.getValidationOptions()
	}

	// Use the same validation logic but with custom options
	if len(certs) == 0 {
		return NewValidationError(InvalidChain, "certificate chain is empty")
	}

	if len(certs) > opts.MaxChainLength {
		return NewValidationError(InvalidChain, 
			fmt.Sprintf("certificate chain too long: %d certificates (max: %d)", 
				len(certs), opts.MaxChainLength))
	}

	var validationErrors []error
	
	for i, cert := range certs {
		if err := cv.validateCertificateInChain(cert, i, len(certs)); err != nil {
			validationErrors = append(validationErrors, err)
		}
	}

	if err := cv.validateChainStructure(certs, opts); err != nil {
		validationErrors = append(validationErrors, err)
	}

	if err := cv.validateTrustPath(certs, opts); err != nil {
		validationErrors = append(validationErrors, err)
	}

	if len(validationErrors) > 0 {
		return NewAggregatedValidationError(validationErrors)
	}

	return nil
}

// BuildChain attempts to build a complete certificate chain from available certificates
func (cv *ChainValidator) BuildChain(leafCert *x509.Certificate, availableCerts []*x509.Certificate) ([]*x509.Certificate, error) {
	if leafCert == nil {
		return nil, errors.New("leaf certificate cannot be nil")
	}

	chain := []*x509.Certificate{leafCert}
	current := leafCert

	// Build chain by finding issuers
	for {
		// Check if current certificate is self-signed (root)
		if current.Subject.String() == current.Issuer.String() {
			break
		}

		// Look for issuer in available certificates
		var issuer *x509.Certificate
		for _, cert := range availableCerts {
			if cert.Subject.String() == current.Issuer.String() {
				issuer = cert
				break
			}
		}

		if issuer == nil {
			// Cannot find issuer - chain is incomplete
			break
		}

		// Verify signature before adding to chain
		if err := current.CheckSignatureFrom(issuer); err != nil {
			return nil, fmt.Errorf("invalid signature from issuer: %v", err)
		}

		chain = append(chain, issuer)
		current = issuer

		// Prevent infinite loops
		if len(chain) > 20 {
			return nil, errors.New("chain building exceeded maximum depth")
		}
	}

	return chain, nil
}