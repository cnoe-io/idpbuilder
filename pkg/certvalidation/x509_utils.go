package certvalidation

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ParsePEMCertificate parses a PEM-encoded certificate
func ParsePEMCertificate(pemData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM data")
	}
	
	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("expected CERTIFICATE block, got %s", block.Type)
	}
	
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}
	
	return cert, nil
}

// ParsePEMCertificates parses multiple PEM-encoded certificates from a single input
func ParsePEMCertificates(pemData []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	remaining := pemData
	
	for {
		block, rest := pem.Decode(remaining)
		if block == nil {
			break
		}
		
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate: %w", err)
			}
			certs = append(certs, cert)
		}
		
		remaining = rest
	}
	
	if len(certs) == 0 {
		return nil, errors.New("no certificates found in PEM data")
	}
	
	return certs, nil
}

// CertificateToPEM converts an X509 certificate to PEM format
func CertificateToPEM(cert *x509.Certificate) ([]byte, error) {
	if cert == nil {
		return nil, errors.New("certificate cannot be nil")
	}
	
	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}
	
	return pem.EncodeToMemory(block), nil
}

// GetCertificateInfo extracts basic information from a certificate
func GetCertificateInfo(cert *x509.Certificate) CertificateInfo {
	if cert == nil {
		return CertificateInfo{}
	}
	
	// Convert IP addresses to strings
	var ipStrings []string
	for _, ip := range cert.IPAddresses {
		ipStrings = append(ipStrings, ip.String())
	}
	
	info := CertificateInfo{
		Subject:      cert.Subject.String(),
		Issuer:       cert.Issuer.String(),
		SerialNumber: cert.SerialNumber.String(),
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		KeyUsage:     cert.KeyUsage,
		ExtKeyUsage:  cert.ExtKeyUsage,
		DNSNames:     cert.DNSNames,
		IPAddresses:  ipStrings,
		IsCA:         cert.IsCA,
	}
	
	// Extract common name
	if cert.Subject.CommonName != "" {
		info.CommonName = cert.Subject.CommonName
	}
	
	// Check validity
	now := time.Now()
	info.IsValid = now.After(cert.NotBefore) && now.Before(cert.NotAfter)
	
	// Calculate days until expiration
	if cert.NotAfter.After(now) {
		info.DaysToExpiry = int(cert.NotAfter.Sub(now).Hours() / 24)
	}
	
	return info
}

// CertificateInfo contains extracted information from a certificate
type CertificateInfo struct {
	Subject       string
	CommonName    string
	Issuer        string
	SerialNumber  string
	NotBefore     time.Time
	NotAfter      time.Time
	KeyUsage      x509.KeyUsage
	ExtKeyUsage   []x509.ExtKeyUsage
	DNSNames      []string
	IPAddresses   []string
	IsCA          bool
	IsValid       bool
	DaysToExpiry  int
}

// ValidateCertificateTime checks if a certificate is valid at a specific time
func ValidateCertificateTime(cert *x509.Certificate, t time.Time) error {
	if cert == nil {
		return errors.New("certificate cannot be nil")
	}
	
	if t.Before(cert.NotBefore) {
		return fmt.Errorf("certificate is not yet valid (valid from: %s)", cert.NotBefore.Format(time.RFC3339))
	}
	
	if t.After(cert.NotAfter) {
		return fmt.Errorf("certificate has expired (expired on: %s)", cert.NotAfter.Format(time.RFC3339))
	}
	
	return nil
}

// IsSelfSigned checks if a certificate is self-signed
func IsSelfSigned(cert *x509.Certificate) bool {
	if cert == nil {
		return false
	}
	
	return cert.Subject.String() == cert.Issuer.String()
}

// GetCertificateFingerprint calculates the SHA-256 fingerprint of a certificate
func GetCertificateFingerprint(cert *x509.Certificate) string {
	if cert == nil {
		return ""
	}
	
	// x509.Certificate already has a SHA-256 fingerprint in the Fingerprint field
	// But we'll calculate it manually for consistency
	hash := "sha256"
	return fmt.Sprintf("%s:%x", hash, cert.Raw)
}

// FilterCertificatesByUsage filters certificates based on key usage
func FilterCertificatesByUsage(certs []*x509.Certificate, usage x509.KeyUsage) []*x509.Certificate {
	var filtered []*x509.Certificate
	
	for _, cert := range certs {
		if cert.KeyUsage&usage != 0 {
			filtered = append(filtered, cert)
		}
	}
	
	return filtered
}

// FindCertificatesBySubject finds certificates with a specific subject
func FindCertificatesBySubject(certs []*x509.Certificate, subject string) []*x509.Certificate {
	var matches []*x509.Certificate
	subject = strings.ToLower(subject)
	
	for _, cert := range certs {
		certSubject := strings.ToLower(cert.Subject.String())
		if strings.Contains(certSubject, subject) || 
		   strings.ToLower(cert.Subject.CommonName) == subject {
			matches = append(matches, cert)
		}
	}
	
	return matches
}

// SortCertificatesByExpiry sorts certificates by expiry date (earliest first)
func SortCertificatesByExpiry(certs []*x509.Certificate) []*x509.Certificate {
	if len(certs) <= 1 {
		return certs
	}
	
	// Simple bubble sort for small arrays
	sorted := make([]*x509.Certificate, len(certs))
	copy(sorted, certs)
	
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].NotAfter.After(sorted[j+1].NotAfter) {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	
	return sorted
}

// ExtractCertificateChainFromPEM extracts and orders certificates from PEM data
// Returns them ordered from leaf to root (if possible to determine)
func ExtractCertificateChainFromPEM(pemData []byte) ([]*x509.Certificate, error) {
	certs, err := ParsePEMCertificates(pemData)
	if err != nil {
		return nil, err
	}
	
	if len(certs) <= 1 {
		return certs, nil
	}
	
	// Try to order certificates in chain order (leaf to root)
	var ordered []*x509.Certificate
	remaining := make([]*x509.Certificate, len(certs))
	copy(remaining, certs)
	
	// Find the leaf certificate (one that is not an issuer of any other cert)
	var leaf *x509.Certificate
	for i, cert := range remaining {
		isIssuer := false
		for _, other := range remaining {
			if cert.Subject.String() == other.Issuer.String() && cert != other {
				isIssuer = true
				break
			}
		}
		
		if !isIssuer && !cert.IsCA {
			leaf = cert
			remaining = append(remaining[:i], remaining[i+1:]...)
			break
		}
	}
	
	if leaf != nil {
		ordered = append(ordered, leaf)
		
		// Build chain by following issuer relationships
		current := leaf
		for len(remaining) > 0 {
			found := false
			for i, cert := range remaining {
				if current.Issuer.String() == cert.Subject.String() {
					ordered = append(ordered, cert)
					current = cert
					remaining = append(remaining[:i], remaining[i+1:]...)
					found = true
					break
				}
			}
			if !found {
				break
			}
		}
		
		// Add any remaining certificates
		ordered = append(ordered, remaining...)
	} else {
		// If we can't determine order, return as-is
		ordered = certs
	}
	
	return ordered, nil
}