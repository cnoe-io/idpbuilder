package certvalidation

import (
	"crypto/x509"
	"errors"
	"fmt"
	"time"
)

// ChainValidator provides certificate chain validation functionality
type ChainValidator struct {
	// intermediates holds intermediate certificates for chain building
	intermediates *x509.CertPool
	// roots holds root CA certificates for verification
	roots *x509.CertPool
	// verifyOpts contains additional verification options
	verifyOpts x509.VerifyOptions
}

// NewChainValidator creates a new certificate chain validator
func NewChainValidator() *ChainValidator {
	return &ChainValidator{
		intermediates: x509.NewCertPool(),
		roots:         x509.NewCertPool(),
		verifyOpts: x509.VerifyOptions{
			KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		},
	}
}

// AddIntermediateCert adds an intermediate certificate to the chain validator
func (cv *ChainValidator) AddIntermediateCert(cert *x509.Certificate) {
	if cert != nil {
		cv.intermediates.AddCert(cert)
	}
}

// AddRootCert adds a root CA certificate to the chain validator
func (cv *ChainValidator) AddRootCert(cert *x509.Certificate) {
	if cert != nil {
		cv.roots.AddCert(cert)
	}
}

// SetVerifyTime sets the time at which certificates should be verified
func (cv *ChainValidator) SetVerifyTime(t time.Time) {
	cv.verifyOpts.CurrentTime = t
}

// ValidateChain validates a certificate chain starting from the leaf certificate
func (cv *ChainValidator) ValidateChain(leafCert *x509.Certificate) error {
	if leafCert == nil {
		return errors.New("leaf certificate cannot be nil")
	}

	// Set up verification options
	opts := cv.verifyOpts
	opts.Roots = cv.roots
	opts.Intermediates = cv.intermediates

	// If no roots are provided, use system root pool
	if opts.Roots == nil || len(opts.Roots.Subjects()) == 0 {
		systemRoots, err := x509.SystemCertPool()
		if err != nil {
			return fmt.Errorf("failed to get system cert pool: %w", err)
		}
		opts.Roots = systemRoots
	}

	// Perform certificate chain verification
	chains, err := leafCert.Verify(opts)
	if err != nil {
		return fmt.Errorf("certificate chain validation failed: %w", err)
	}

	if len(chains) == 0 {
		return errors.New("no valid certificate chains found")
	}

	return nil
}

// ValidateChainWithHostname validates a certificate chain for a specific hostname
func (cv *ChainValidator) ValidateChainWithHostname(leafCert *x509.Certificate, hostname string) error {
	// First validate the basic chain
	if err := cv.ValidateChain(leafCert); err != nil {
		return err
	}

	// Then validate hostname binding
	if err := leafCert.VerifyHostname(hostname); err != nil {
		return fmt.Errorf("hostname verification failed: %w", err)
	}

	return nil
}

// BuildChain attempts to build a certificate chain from available certificates
func (cv *ChainValidator) BuildChain(leafCert *x509.Certificate) ([]*x509.Certificate, error) {
	if leafCert == nil {
		return nil, errors.New("leaf certificate cannot be nil")
	}

	// Set up verification options for chain building
	opts := cv.verifyOpts
	opts.Roots = cv.roots
	opts.Intermediates = cv.intermediates

	// If no roots are provided, use system root pool
	if opts.Roots == nil || len(opts.Roots.Subjects()) == 0 {
		systemRoots, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("failed to get system cert pool: %w", err)
		}
		opts.Roots = systemRoots
	}

	// Verify and get the certificate chains
	chains, err := leafCert.Verify(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to build certificate chain: %w", err)
	}

	if len(chains) == 0 {
		return nil, errors.New("no certificate chains could be built")
	}

	// Return the first valid chain
	return chains[0], nil
}

// GetChainInfo returns information about a certificate chain
func (cv *ChainValidator) GetChainInfo(chain []*x509.Certificate) ChainInfo {
	if len(chain) == 0 {
		return ChainInfo{}
	}

	info := ChainInfo{
		Length:    len(chain),
		LeafCert:  chain[0],
		RootCert:  chain[len(chain)-1],
		IsValid:   true,
		ExpiresAt: chain[0].NotAfter,
	}

	// Find the earliest expiration time in the chain
	for _, cert := range chain {
		if cert.NotAfter.Before(info.ExpiresAt) {
			info.ExpiresAt = cert.NotAfter
		}
	}

	// Check if any certificate in the chain has expired or will expire soon
	now := time.Now()
	for _, cert := range chain {
		if cert.NotAfter.Before(now) {
			info.IsValid = false
			info.Issues = append(info.Issues, fmt.Sprintf("Certificate '%s' has expired", cert.Subject.CommonName))
		} else if cert.NotAfter.Before(now.Add(24 * time.Hour * 30)) { // 30 days
			info.Issues = append(info.Issues, fmt.Sprintf("Certificate '%s' will expire soon", cert.Subject.CommonName))
		}
	}

	return info
}

// ChainInfo contains information about a certificate chain
type ChainInfo struct {
	Length    int
	LeafCert  *x509.Certificate
	RootCert  *x509.Certificate
	IsValid   bool
	ExpiresAt time.Time
	Issues    []string
}