// Copyright 2024 The IDP Builder Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package certs

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"time"
)

// Certificate wraps an x509.Certificate with additional functionality
type Certificate struct {
	*x509.Certificate
	// Raw PEM data for the certificate
	PEMData []byte
}

// CertificateBundle represents a complete certificate configuration
type CertificateBundle struct {
	// CACert is the Certificate Authority certificate
	CACert *Certificate
	
	// ClientCert is the client certificate for mutual TLS
	ClientCert *Certificate
	
	// ClientKey is the private key for the client certificate
	ClientKey interface{}
	
	// CertificateChain contains the full certificate chain
	CertificateChain []*Certificate
	
	// ValidFrom indicates when the bundle becomes valid
	ValidFrom time.Time
	
	// ValidUntil indicates when the bundle expires
	ValidUntil time.Time
}

// TLSConfig represents TLS configuration for registry connections
type TLSConfig struct {
	// InsecureSkipVerify disables certificate verification (NOT recommended)
	InsecureSkipVerify bool
	
	// ServerName for certificate verification
	ServerName string
	
	// RootCAs defines the certificate authorities to trust
	RootCAs *x509.CertPool
	
	// ClientCAs defines the certificate authorities for client certificates
	ClientCAs *x509.CertPool
	
	// Certificates contains client certificates for mutual TLS
	Certificates []tls.Certificate
	
	// MinVersion specifies the minimum TLS version
	MinVersion uint16
	
	// MaxVersion specifies the maximum TLS version
	MaxVersion uint16
	
	// CipherSuites specifies the preferred cipher suites
	CipherSuites []uint16
	
	// CurvePreferences specifies the preferred elliptic curves
	CurvePreferences []tls.CurveID
	
	// ClientAuth determines client certificate requirements
	ClientAuth tls.ClientAuthType
	
	// Registry specifies the registry hostname
	Registry string
	
	// ValidateHostname determines if hostname validation is enabled
	ValidateHostname bool
	
	// Timeout specifies the connection timeout
	Timeout time.Duration
}

// CertificateValidator defines the interface for certificate validation
type CertificateValidator interface {
	// Validate checks if the certificate is valid
	Validate(cert *Certificate) error
	
	// ValidateChain checks if the certificate chain is valid
	ValidateChain(chain []*Certificate) error
	
	// IsExpired checks if the certificate has expired
	IsExpired(cert *Certificate) bool
	
	// WillExpireSoon checks if certificate will expire within threshold
	WillExpireSoon(cert *Certificate, threshold time.Duration) bool
}

// BasicCertificateValidator provides basic certificate validation
type BasicCertificateValidator struct {
	// TrustedCAs contains trusted certificate authorities
	TrustedCAs *x509.CertPool
	
	// AllowSelfSigned permits self-signed certificates
	AllowSelfSigned bool
	
	// CheckExpiry enables expiration checking
	CheckExpiry bool
}

// NewCertificate creates a new Certificate from x509.Certificate
func NewCertificate(cert *x509.Certificate, pemData []byte) *Certificate {
	return &Certificate{
		Certificate: cert,
		PEMData:     pemData,
	}
}

// NewCertificateBundle creates a new empty certificate bundle
func NewCertificateBundle() *CertificateBundle {
	return &CertificateBundle{
		CertificateChain: make([]*Certificate, 0),
	}
}

// NewTLSConfig creates a new TLS configuration with secure defaults
func NewTLSConfig() *TLSConfig {
	return &TLSConfig{
		MinVersion:       MinTLSVersion,
		MaxVersion:       PreferredTLSVersion,
		CipherSuites:     PreferredCipherSuites,
		CurvePreferences: SecureTLSCurves,
		ClientAuth:       tls.NoClientCert,
	}
}

// DefaultTLSConfig creates a TLS configuration with default settings
func DefaultTLSConfig() *TLSConfig {
	return &TLSConfig{
		MinVersion:       tls.VersionTLS12,
		ValidateHostname: true,
		Timeout:          10 * time.Second,
	}
}

// NewBasicCertificateValidator creates a new basic certificate validator
func NewBasicCertificateValidator(trustedCAs *x509.CertPool) *BasicCertificateValidator {
	return &BasicCertificateValidator{
		TrustedCAs:  trustedCAs,
		CheckExpiry: true,
	}
}

// IsValid checks if the certificate is currently valid
func (c *Certificate) IsValid() bool {
	now := time.Now()
	return now.After(c.NotBefore) && now.Before(c.NotAfter)
}

// IsExpired checks if the certificate has expired
func (c *Certificate) IsExpired() bool {
	return time.Now().After(c.NotAfter)
}

// WillExpireSoon checks if the certificate will expire within the given duration
func (c *Certificate) WillExpireSoon(threshold time.Duration) bool {
	return time.Until(c.NotAfter) < threshold
}

// GetCommonName returns the certificate's Common Name
func (c *Certificate) GetCommonName() string {
	return c.Subject.CommonName
}

// GetSANs returns the Subject Alternative Names
func (c *Certificate) GetSANs() []string {
	return c.DNSNames
}

// HasPrivateKey checks if the bundle has a client private key
func (cb *CertificateBundle) HasPrivateKey() bool {
	return cb.ClientKey != nil
}

// IsValid checks if the certificate bundle is currently valid
func (cb *CertificateBundle) IsValid() bool {
	now := time.Now()
	
	// Check bundle validity period
	if !cb.ValidFrom.IsZero() && now.Before(cb.ValidFrom) {
		return false
	}
	if !cb.ValidUntil.IsZero() && now.After(cb.ValidUntil) {
		return false
	}
	
	// Check individual certificates
	if cb.CACert != nil && !cb.CACert.IsValid() {
		return false
	}
	if cb.ClientCert != nil && !cb.ClientCert.IsValid() {
		return false
	}
	
	return true
}

// UpdateValidityPeriod updates the bundle validity period based on certificates
func (cb *CertificateBundle) UpdateValidityPeriod() {
	if cb.CACert != nil {
		if cb.ValidFrom.IsZero() || cb.CACert.NotBefore.After(cb.ValidFrom) {
			cb.ValidFrom = cb.CACert.NotBefore
		}
		if cb.ValidUntil.IsZero() || cb.CACert.NotAfter.Before(cb.ValidUntil) {
			cb.ValidUntil = cb.CACert.NotAfter
		}
	}
	
	if cb.ClientCert != nil {
		if cb.ValidFrom.IsZero() || cb.ClientCert.NotBefore.After(cb.ValidFrom) {
			cb.ValidFrom = cb.ClientCert.NotBefore
		}
		if cb.ValidUntil.IsZero() || cb.ClientCert.NotAfter.Before(cb.ValidUntil) {
			cb.ValidUntil = cb.ClientCert.NotAfter
		}
	}
}

// ToTLSConfig converts the TLSConfig to Go's tls.Config
func (tc *TLSConfig) ToTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: tc.InsecureSkipVerify,
		ServerName:         tc.ServerName,
		RootCAs:            tc.RootCAs,
		ClientCAs:          tc.ClientCAs,
		Certificates:       tc.Certificates,
		MinVersion:         tc.MinVersion,
		MaxVersion:         tc.MaxVersion,
		CipherSuites:       tc.CipherSuites,
		CurvePreferences:   tc.CurvePreferences,
		ClientAuth:         tc.ClientAuth,
	}
}

// AddClientCertificate adds a client certificate to the TLS configuration
func (tc *TLSConfig) AddClientCertificate(cert tls.Certificate) {
	tc.Certificates = append(tc.Certificates, cert)
}

// SetRootCA sets the root certificate authority
func (tc *TLSConfig) SetRootCA(caCert *Certificate) error {
	if tc.RootCAs == nil {
		tc.RootCAs = x509.NewCertPool()
	}
	
	if caCert.Certificate == nil {
		return fmt.Errorf("invalid CA certificate")
	}
	
	tc.RootCAs.AddCert(caCert.Certificate)
	return nil
}

// Validate implements CertificateValidator interface
func (v *BasicCertificateValidator) Validate(cert *Certificate) error {
	if cert == nil || cert.Certificate == nil {
		return fmt.Errorf(ErrInvalidCertificate)
	}
	
	// Check expiration if enabled
	if v.CheckExpiry {
		if cert.IsExpired() {
			return fmt.Errorf(ErrCertificateExpired)
		}
		
		now := time.Now()
		if now.Before(cert.NotBefore) {
			return fmt.Errorf(ErrCertificateNotYetValid)
		}
	}
	
	return nil
}

// ValidateChain implements CertificateValidator interface
func (v *BasicCertificateValidator) ValidateChain(chain []*Certificate) error {
	if len(chain) == 0 {
		return fmt.Errorf("empty certificate chain")
	}
	
	// Validate each certificate in the chain
	for i, cert := range chain {
		if err := v.Validate(cert); err != nil {
			return fmt.Errorf("certificate %d in chain: %w", i, err)
		}
	}
	
	// If we have trusted CAs, validate the chain against them
	if v.TrustedCAs != nil && len(chain) > 0 {
		leafCert := chain[0].Certificate
		intermediates := x509.NewCertPool()
		
		// Add intermediate certificates to the pool
		for i := 1; i < len(chain); i++ {
			intermediates.AddCert(chain[i].Certificate)
		}
		
		verifyOpts := x509.VerifyOptions{
			Roots:         v.TrustedCAs,
			Intermediates: intermediates,
		}
		
		_, err := leafCert.Verify(verifyOpts)
		if err != nil {
			return fmt.Errorf(ErrCertificateChainInvalid+": %w", err)
		}
	}
	
	return nil
}

// IsExpired implements CertificateValidator interface
func (v *BasicCertificateValidator) IsExpired(cert *Certificate) bool {
	if cert == nil || cert.Certificate == nil {
		return true
	}
	return cert.IsExpired()
}

// WillExpireSoon implements CertificateValidator interface
func (v *BasicCertificateValidator) WillExpireSoon(cert *Certificate, threshold time.Duration) bool {
	if cert == nil || cert.Certificate == nil {
		return true
	}
	return cert.WillExpireSoon(threshold)
}

// ParseCertificateFromPEM parses a certificate from PEM data
func ParseCertificateFromPEM(pemData []byte) (*Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf(ErrInvalidPEMBlock)
	}
	
	if block.Type != PEMBlockCertificate {
		return nil, fmt.Errorf("expected certificate PEM block, got %s", block.Type)
	}
	
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf(ErrInvalidCertificate+": %w", err)
	}
	
	return NewCertificate(cert, pemData), nil
}

// ParseCertificatesFromPEM parses multiple certificates from PEM data
func ParseCertificatesFromPEM(pemData []byte) ([]*Certificate, error) {
	var certificates []*Certificate
	remaining := pemData
	
	for len(remaining) > 0 {
		block, rest := pem.Decode(remaining)
		if block == nil {
			break
		}
		
		if block.Type == PEMBlockCertificate {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate: %w", err)
			}
			
			// Extract the PEM block for this specific certificate
			blockEnd := len(pemData) - len(rest)
			blockStart := blockEnd - len(pem.EncodeToMemory(block))
			certPEM := pemData[blockStart:blockEnd]
			
			certificates = append(certificates, NewCertificate(cert, certPEM))
		}
		
		remaining = rest
	}
	
	if len(certificates) == 0 {
		return nil, fmt.Errorf("no certificates found in PEM data")
	}
	
	return certificates, nil
}

// LoadCertificateFromReader loads a certificate from an io.Reader
func LoadCertificateFromReader(reader io.Reader) (*Certificate, error) {
	pemData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate data: %w", err)
	}
	
	return ParseCertificateFromPEM(pemData)
}

// CertificateError represents a certificate-related error
type CertificateError struct {
	Type    string
	Message string
	Cause   error
}

// Error implements the error interface
func (e *CertificateError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Cause.Error())
	}
	return e.Message
}

// Unwrap implements error unwrapping
func (e *CertificateError) Unwrap() error {
	return e.Cause
}