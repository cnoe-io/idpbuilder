package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"testing"
	"time"
)

func TestParseCertificate(t *testing.T) {
	// Create a test certificate
	cert := createTestCertificate(t)
	pemData := encodeCertToPEM(t, cert)

	// Test successful parsing
	parsedCert, err := parseCertificate(pemData)
	if err != nil {
		t.Fatalf("parseCertificate() failed: %v", err)
	}

	if !cert.Equal(parsedCert) {
		t.Error("parseCertificate() returned different certificate")
	}

	// Test empty PEM data
	_, err = parseCertificate([]byte(""))
	if err == nil {
		t.Error("parseCertificate() should fail with empty data")
	}

	// Test invalid PEM block
	invalidPEM := []byte("-----BEGIN PRIVATE KEY-----\ninvalid data\n-----END PRIVATE KEY-----")
	_, err = parseCertificate(invalidPEM)
	if err == nil {
		t.Error("parseCertificate() should fail with non-certificate PEM")
	}
}

// Feature flag test removed - features are now always enabled in production

func TestValidateCertificateExpiry(t *testing.T) {
	now := time.Now()

	// Valid certificate (not expired, not yet valid)
	validCert := &x509.Certificate{
		NotBefore: now.Add(-1 * time.Hour),
		NotAfter:  now.Add(1 * time.Hour),
	}
	if err := validateCertificateExpiry(validCert); err != nil {
		t.Errorf("validateCertificateExpiry() should pass for valid cert, got: %v", err)
	}

	// Expired certificate
	expiredCert := &x509.Certificate{
		NotBefore: now.Add(-2 * time.Hour),
		NotAfter:  now.Add(-1 * time.Hour),
	}
	if err := validateCertificateExpiry(expiredCert); err != ErrCertExpired {
		t.Errorf("validateCertificateExpiry() should return ErrCertExpired, got: %v", err)
	}

	// Not yet valid certificate
	futureCart := &x509.Certificate{
		NotBefore: now.Add(1 * time.Hour),
		NotAfter:  now.Add(2 * time.Hour),
	}
	if err := validateCertificateExpiry(futureCart); err != ErrCertNotYetValid {
		t.Errorf("validateCertificateExpiry() should return ErrCertNotYetValid, got: %v", err)
	}
}

func TestExpandHomeDir(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set test HOME
	os.Setenv("HOME", "/test/home")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"path with tilde", "~/certs", "/test/home/certs"},
		{"nested path with tilde", "~/config/certs", "/test/home/config/certs"},
		{"absolute path", "/absolute/path", "/absolute/path"},
		{"relative path", "relative/path", "relative/path"},
		{"empty path", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandHomeDir(tt.input)
			if got != tt.expected {
				t.Errorf("expandHomeDir(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizeRegistryName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"registry with port", "registry:5000", "registry_5000"},
		{"registry with path", "registry.io/path", "registry_io_path"},
		{"complex registry", "my.registry.io:443/namespace/repo", "my_registry_io_443_namespace_repo"},
		{"simple name", "localhost", "localhost"},
		{"ip with port", "192.168.1.1:5000", "192_168_1_1_5000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeRegistryName(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeRegistryName(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

// Helper functions for testing

func createTestCertificate(t *testing.T) *x509.Certificate {
	// Generate a private key
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Test Org"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
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

func encodeCertToPEM(t *testing.T, cert *x509.Certificate) []byte {
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
	if certPEM == nil {
		t.Fatal("Failed to encode certificate to PEM")
	}
	return certPEM
}
