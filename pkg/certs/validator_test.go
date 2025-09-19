package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"testing"
	"time"
)

// createTestCertificateWithTimes creates a test certificate with custom validity times
func createTestCertificateWithTimes(t *testing.T, notBefore, notAfter time.Time) *x509.Certificate {
	// Generate a private key
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
			Country:      []string{"US"},
			Locality:     []string{"San Francisco"},
		},
		NotBefore:   notBefore,
		NotAfter:    notAfter,
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	return cert
}

func TestDefaultCertValidator_ValidateCertificate_ValidCert(t *testing.T) {
	validator := &DefaultCertValidator{}

	// Create a valid certificate for testing
	cert := createTestCertificateWithTimes(t, time.Now().Add(-24*time.Hour), time.Now().Add(24*time.Hour))
	cert.Subject.CommonName = "test.example.com"

	err := validator.ValidateCertificate(cert)
	if err != nil {
		t.Fatalf("Expected no error for valid certificate, got: %v", err)
	}
}

func TestDefaultCertValidator_ValidateCertificate_NilCert(t *testing.T) {
	validator := &DefaultCertValidator{}

	err := validator.ValidateCertificate(nil)
	if err == nil {
		t.Fatal("Expected error for nil certificate")
	}

	expectedMsg := "certificate is nil"
	if err.Error() != expectedMsg {
		t.Fatalf("Expected error message '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestDefaultCertValidator_ValidateCertificate_ExpiredCert(t *testing.T) {
	validator := &DefaultCertValidator{}

	// Create expired certificate
	cert := createTestCertificateWithTimes(t, time.Now().Add(-48*time.Hour), time.Now().Add(-24*time.Hour))
	cert.Subject.CommonName = "expired.example.com"

	err := validator.ValidateCertificate(cert)
	if err == nil {
		t.Fatal("Expected error for expired certificate")
	}
}

func TestDefaultCertValidator_ValidateCertificate_NoSubjectNames(t *testing.T) {
	validator := &DefaultCertValidator{}

	// Create certificate with no subject names by modifying the template
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
			Country:      []string{"US"},
			Locality:     []string{"San Francisco"},
			// No CommonName
		},
		NotBefore:   time.Now().Add(-24 * time.Hour),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		// No DNS names or IP addresses
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	err = validator.ValidateCertificate(cert)
	if err == nil {
		t.Fatal("Expected error for certificate with no subject names")
	}

	expectedMsg := "certificate has no valid subject names"
	if err.Error() != expectedMsg {
		t.Fatalf("Expected error message '%s', got: %s", expectedMsg, err.Error())
	}
}
