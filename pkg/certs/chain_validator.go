package certs

import (
	"crypto/x509"
	"fmt"
	"strings"
	"time"
)


// ChainValidationOptions defines options for certificate chain validation
type ChainValidationOptions struct {
	// MaxChainLength specifies the maximum allowed certificate chain length
	MaxChainLength int
	
	// AllowSelfSigned allows self-signed certificates in the chain
	AllowSelfSigned bool
	
	// RequireLeafKeyUsage specifies required key usage for leaf certificate
	RequireLeafKeyUsage []x509.KeyUsage
	
	// RequireLeafExtKeyUsage specifies required extended key usage for leaf certificate
	RequireLeafExtKeyUsage []x509.ExtKeyUsage
	
	// AllowExpiredIntermediates allows expired intermediate certificates in chain
	AllowExpiredIntermediates bool
	
	// VerifyHostname specifies whether to verify hostname against leaf certificate
	VerifyHostname bool
	
	// Hostname is the expected hostname for verification
	Hostname string
	
	// CurrentTime specifies the time for validity checks (defaults to now)
	CurrentTime time.Time
}

// TrustStoreProvider abstracts trust store operations for chain validation
type TrustStoreProvider interface {
	GetTrustedCerts(registry string) ([]*x509.Certificate, error)
	IsInsecure(registry string) bool
}

// ChainValidator performs certificate chain validation
type ChainValidator struct {
	trustStore TrustStoreProvider
	mode       ValidationMode
}

// NewChainValidator creates a new certificate chain validator
func NewChainValidator(trustStore TrustStoreProvider, mode ValidationMode) *ChainValidator {
	return &ChainValidator{
		trustStore: trustStore,
		mode:       mode,
	}
}

// ValidateChain validates a certificate chain according to validation mode and options
func (cv *ChainValidator) ValidateChain(chain []*x509.Certificate, registry string, options *ChainValidationOptions) error {
	if len(chain) == 0 {
		return NewValidationError(ChainIncomplete, "certificate chain is empty", "unknown")
	}

	// Set default options if not provided
	if options == nil {
		options = cv.getDefaultOptions()
	}
	
	// Set current time if not specified
	if options.CurrentTime.IsZero() {
		options.CurrentTime = time.Now()
	}

	// Validate chain length
	if err := cv.validateChainLength(chain, options); err != nil {
		return err
	}

	// Validate chain ordering
	if err := cv.validateChainOrdering(chain); err != nil {
		return err
	}

	// Validate each certificate in the chain
	for i, cert := range chain {
		if err := cv.validateIndividualCertificate(cert, i == 0, options); err != nil {
			return err
		}
	}

	// Validate certificate signatures (chain of trust)
	if err := cv.validateSignatures(chain); err != nil {
		return err
	}

	// Build trust verification based on mode
	switch cv.mode {
	case StrictMode:
		return cv.validateTrustStrict(chain, registry, options)
	case LenientMode:
		return cv.validateTrustLenient(chain, registry, options)
	case InsecureMode:
		// Minimal trust validation
		return cv.validateTrustMinimal(chain, registry, options)
	default:
		return fmt.Errorf("unsupported validation mode: %v", cv.mode)
	}
}

// validateChainLength checks if chain length is within acceptable limits
func (cv *ChainValidator) validateChainLength(chain []*x509.Certificate, options *ChainValidationOptions) error {
	maxLength := options.MaxChainLength
	if maxLength == 0 {
		maxLength = cv.getMaxChainLengthForMode()
	}
	
	if len(chain) > maxLength {
		leafSubject := "unknown"
		if len(chain) > 0 {
			leafSubject = chain[0].Subject.String()
		}
		
		validationErr := NewValidationError(ChainTooLong, 
			fmt.Sprintf("certificate chain too long: %d certificates (max: %d)", len(chain), maxLength),
			leafSubject)
		validationErr.AddDetail("chain_length", len(chain))
		validationErr.AddDetail("max_allowed", maxLength)
		return validationErr
	}
	
	return nil
}

// validateChainOrdering ensures certificates are in proper order (leaf to root)
func (cv *ChainValidator) validateChainOrdering(chain []*x509.Certificate) error {
	if len(chain) < 2 {
		return nil // Single certificate or empty chain
	}
	
	for i := 0; i < len(chain)-1; i++ {
		cert := chain[i]
		issuer := chain[i+1]
		
		// Check if the next certificate is the issuer of the current certificate
		if !cv.isIssuedBy(cert, issuer) {
			return NewValidationError(ChainIncomplete,
				fmt.Sprintf("certificate chain ordering invalid at position %d", i),
				cert.Subject.String())
		}
	}
	
	return nil
}

// isIssuedBy checks if cert was issued by issuer
func (cv *ChainValidator) isIssuedBy(cert, issuer *x509.Certificate) bool {
	// Compare issuer DN with subject DN
	if cert.Issuer.String() != issuer.Subject.String() {
		return false
	}
	
	// Verify authority key identifier if present
	if len(cert.AuthorityKeyId) > 0 && len(issuer.SubjectKeyId) > 0 {
		return string(cert.AuthorityKeyId) == string(issuer.SubjectKeyId)
	}
	
	return true
}

// validateIndividualCertificate validates a single certificate
func (cv *ChainValidator) validateIndividualCertificate(cert *x509.Certificate, isLeaf bool, options *ChainValidationOptions) error {
	if cert == nil {
		return NewValidationError(InvalidCertificate, "certificate is nil", "unknown")
	}

	// Validate certificate times
	if err := cv.validateCertificateTimes(cert, options, isLeaf); err != nil {
		return err
	}

	// Validate key usage for leaf certificate
	if isLeaf {
		if err := cv.validateLeafKeyUsage(cert, options); err != nil {
			return err
		}
	}

	// Validate hostname for leaf certificate
	if isLeaf && options.VerifyHostname && options.Hostname != "" {
		if err := cv.validateHostname(cert, options.Hostname); err != nil {
			return err
		}
	}

	// Validate signature algorithm strength
	if err := cv.validateSignatureAlgorithm(cert); err != nil {
		return err
	}

	return nil
}

// validateCertificateTimes checks certificate validity periods
func (cv *ChainValidator) validateCertificateTimes(cert *x509.Certificate, options *ChainValidationOptions, isLeaf bool) error {
	now := options.CurrentTime
	
	// Check if certificate is not yet valid
	if now.Before(cert.NotBefore) {
		return NewValidationError(NotYetValid,
			fmt.Sprintf("certificate not yet valid (valid from: %v)", cert.NotBefore),
			cert.Subject.String())
	}
	
	// Check if certificate is expired
	if now.After(cert.NotAfter) {
		// For intermediate certificates in lenient mode, allow expired if configured
		if !isLeaf && options.AllowExpiredIntermediates && cv.mode == LenientMode {
			return nil // Allow expired intermediate
		}
		
		return NewValidationError(Expired,
			fmt.Sprintf("certificate expired (expired on: %v)", cert.NotAfter),
			cert.Subject.String())
	}
	
	return nil
}

// validateLeafKeyUsage validates key usage for leaf certificate
func (cv *ChainValidator) validateLeafKeyUsage(cert *x509.Certificate, options *ChainValidationOptions) error {
	// Check required key usage
	for _, requiredUsage := range options.RequireLeafKeyUsage {
		if cert.KeyUsage&requiredUsage == 0 {
			return NewValidationError(KeyUsageMismatch,
				fmt.Sprintf("certificate missing required key usage: %v", requiredUsage),
				cert.Subject.String())
		}
	}
	
	// Check required extended key usage
	certExtUsages := make(map[x509.ExtKeyUsage]bool)
	for _, usage := range cert.ExtKeyUsage {
		certExtUsages[usage] = true
	}
	
	for _, requiredExtUsage := range options.RequireLeafExtKeyUsage {
		if !certExtUsages[requiredExtUsage] {
			return NewValidationError(ExtendedKeyUsageMismatch,
				fmt.Sprintf("certificate missing required extended key usage: %v", requiredExtUsage),
				cert.Subject.String())
		}
	}
	
	return nil
}

// validateHostname validates hostname against certificate
func (cv *ChainValidator) validateHostname(cert *x509.Certificate, hostname string) error {
	// Check common name
	if cert.Subject.CommonName == hostname {
		return nil
	}
	
	// Check DNS names
	for _, dnsName := range cert.DNSNames {
		if cv.matchHostname(hostname, dnsName) {
			return nil
		}
	}
	
	// Check IP addresses
	for _, ip := range cert.IPAddresses {
		if ip.String() == hostname {
			return nil
		}
	}
	
	return NewValidationError(HostnameMismatch,
		fmt.Sprintf("certificate hostname mismatch: %s", hostname),
		cert.Subject.String())
}

// matchHostname performs hostname matching including wildcards
func (cv *ChainValidator) matchHostname(hostname, pattern string) bool {
	if pattern == hostname {
		return true
	}
	
	// Basic wildcard matching
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[2:]
		if strings.HasSuffix(hostname, suffix) {
			// Ensure hostname has one more label than suffix
			hostLabels := strings.Split(hostname, ".")
			suffixLabels := strings.Split(suffix, ".")
			return len(hostLabels) == len(suffixLabels)+1
		}
	}
	
	return false
}

// validateSignatureAlgorithm checks for weak signature algorithms
func (cv *ChainValidator) validateSignatureAlgorithm(cert *x509.Certificate) error {
	// Reject weak algorithms
	weakAlgorithms := map[x509.SignatureAlgorithm]bool{
		x509.MD2WithRSA:    true,
		x509.MD5WithRSA:    true,
		x509.SHA1WithRSA:   true,
		x509.DSAWithSHA1:   true,
		x509.ECDSAWithSHA1: true,
	}
	
	if weakAlgorithms[cert.SignatureAlgorithm] {
		return NewValidationError(WeakSignatureAlgorithm,
			fmt.Sprintf("weak signature algorithm: %v", cert.SignatureAlgorithm),
			cert.Subject.String())
	}
	
	return nil
}

// validateSignatures verifies the signature chain
func (cv *ChainValidator) validateSignatures(chain []*x509.Certificate) error {
	for i := 0; i < len(chain)-1; i++ {
		cert := chain[i]
		issuer := chain[i+1]
		
		if err := cert.CheckSignatureFrom(issuer); err != nil {
			return NewValidationError(SignatureVerificationFailed,
				fmt.Sprintf("signature verification failed: %v", err),
				cert.Subject.String())
		}
	}
	
	return nil
}

// validateTrustStrict performs strict trust validation
func (cv *ChainValidator) validateTrustStrict(chain []*x509.Certificate, registry string, options *ChainValidationOptions) error {
	if cv.trustStore.IsInsecure(registry) {
		return NewValidationError(UntrustedCA,
			"registry marked insecure but strict mode requires trusted certificates",
			chain[0].Subject.String())
	}
	
	// Get trusted certificates for the registry
	trustedCerts, err := cv.trustStore.GetTrustedCerts(registry)
	if err != nil {
		return fmt.Errorf("failed to get trusted certificates: %w", err)
	}
	
	// Root certificate must be in trust store
	rootCert := chain[len(chain)-1]
	
	// Check if root is self-signed
	if cv.isSelfSigned(rootCert) && !options.AllowSelfSigned {
		return NewValidationError(UntrustedCA,
			"self-signed root certificate not allowed in strict mode",
			rootCert.Subject.String())
	}
	
	// Verify root is trusted
	if !cv.isCertificateTrusted(rootCert, trustedCerts) {
		return NewValidationError(UntrustedCA,
			"root certificate not found in trust store",
			rootCert.Subject.String())
	}
	
	return nil
}

// validateTrustLenient performs lenient trust validation
func (cv *ChainValidator) validateTrustLenient(chain []*x509.Certificate, registry string, options *ChainValidationOptions) error {
	if cv.trustStore.IsInsecure(registry) {
		// Allow insecure registries in lenient mode
		return nil
	}
	
	// Get trusted certificates for the registry
	trustedCerts, err := cv.trustStore.GetTrustedCerts(registry)
	if err != nil {
		return fmt.Errorf("failed to get trusted certificates: %w", err)
	}
	
	// If no trusted certificates configured, allow if self-signed is permitted
	if len(trustedCerts) == 0 {
		rootCert := chain[len(chain)-1]
		if cv.isSelfSigned(rootCert) && options.AllowSelfSigned {
			return nil
		}
		return NewValidationError(UntrustedCA,
			"no trusted certificates configured for registry",
			rootCert.Subject.String())
	}
	
	// Check if any certificate in chain is trusted
	for _, cert := range chain {
		if cv.isCertificateTrusted(cert, trustedCerts) {
			return nil
		}
	}
	
	return NewValidationError(UntrustedCA,
		"no certificates in chain found in trust store",
		chain[0].Subject.String())
}

// validateTrustMinimal performs minimal trust validation
func (cv *ChainValidator) validateTrustMinimal(chain []*x509.Certificate, registry string, options *ChainValidationOptions) error {
	// In insecure mode, we only check for basic certificate structure
	// Trust is not validated
	return nil
}

// isSelfSigned checks if certificate is self-signed
func (cv *ChainValidator) isSelfSigned(cert *x509.Certificate) bool {
	return cert.Subject.String() == cert.Issuer.String()
}

// isCertificateTrusted checks if certificate exists in trusted certificates
func (cv *ChainValidator) isCertificateTrusted(cert *x509.Certificate, trustedCerts []*x509.Certificate) bool {
	for _, trusted := range trustedCerts {
		if cert.Equal(trusted) {
			return true
		}
	}
	return false
}

// getDefaultOptions returns default validation options based on mode
func (cv *ChainValidator) getDefaultOptions() *ChainValidationOptions {
	options := &ChainValidationOptions{
		MaxChainLength:            cv.getMaxChainLengthForMode(),
		AllowSelfSigned:          cv.mode != StrictMode,
		AllowExpiredIntermediates: cv.mode == LenientMode,
		VerifyHostname:           cv.mode == StrictMode,
		CurrentTime:              time.Now(),
	}
	
	// Set default key usage requirements for strict mode
	if cv.mode == StrictMode {
		options.RequireLeafKeyUsage = []x509.KeyUsage{
			x509.KeyUsageDigitalSignature,
		}
		options.RequireLeafExtKeyUsage = []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		}
	}
	
	return options
}

// getMaxChainLengthForMode returns default max chain length for validation mode
func (cv *ChainValidator) getMaxChainLengthForMode() int {
	switch cv.mode {
	case StrictMode:
		return 4 // Leaf + 2 intermediates + root
	case LenientMode:
		return 6 // More relaxed
	case InsecureMode:
		return 10 // Very relaxed
	default:
		return 4
	}
}