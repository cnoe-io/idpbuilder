package certvalidation

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"
)

// Helper function to create a test certificate
func createTestCert(subject, issuer pkix.Name, parent *x509.Certificate, parentKey *rsa.PrivateKey, isCA bool) (*x509.Certificate, *rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      subject,
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour * 24 * 365), // 1 year
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:         isCA,
	}

	if isCA {
		template.KeyUsage |= x509.KeyUsageCertSign
		template.BasicConstraintsValid = true
	}

	var signerCert *x509.Certificate
	var signerKey *rsa.PrivateKey

	if parent != nil {
		signerCert = parent
		signerKey = parentKey
	} else {
		signerCert = &template
		signerKey = key
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, signerCert, &key.PublicKey, signerKey)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

func TestNewChainValidator(t *testing.T) {
	validator := NewChainValidator()

	if validator == nil {
		t.Fatal("Expected non-nil ChainValidator")
	}

	if validator.intermediates == nil {
		t.Error("Expected non-nil intermediates pool")
	}

	if validator.roots == nil {
		t.Error("Expected non-nil roots pool")
	}
}

func TestChainValidator_AddCertificates(t *testing.T) {
	validator := NewChainValidator()

	// Create test certificates
	rootSubject := pkix.Name{CommonName: "Test Root CA"}
	rootCert, rootKey, err := createTestCert(rootSubject, rootSubject, nil, nil, true)
	if err != nil {
		t.Fatalf("Failed to create root certificate: %v", err)
	}

	intermediateSubject := pkix.Name{CommonName: "Test Intermediate CA"}
	intermediateCert, _, err := createTestCert(intermediateSubject, rootSubject, rootCert, rootKey, true)
	if err != nil {
		t.Fatalf("Failed to create intermediate certificate: %v", err)
	}

	// Test adding certificates
	validator.AddRootCert(rootCert)
	validator.AddIntermediateCert(intermediateCert)

	// Test adding nil certificates (should not panic)
	validator.AddRootCert(nil)
	validator.AddIntermediateCert(nil)
}

func TestChainValidator_SetVerifyTime(t *testing.T) {
	validator := NewChainValidator()
	testTime := time.Now().Add(time.Hour)

	validator.SetVerifyTime(testTime)

	if !validator.verifyOpts.CurrentTime.Equal(testTime) {
		t.Errorf("Expected verify time %v, got %v", testTime, validator.verifyOpts.CurrentTime)
	}
}

func TestChainValidator_ValidateChain(t *testing.T) {
	validator := NewChainValidator()

	// Test with nil certificate
	err := validator.ValidateChain(nil)
	if err == nil {
		t.Error("Expected error for nil certificate")
	}

	// Create a valid certificate chain
	rootSubject := pkix.Name{CommonName: "Test Root CA"}
	rootCert, rootKey, err := createTestCert(rootSubject, rootSubject, nil, nil, true)
	if err != nil {
		t.Fatalf("Failed to create root certificate: %v", err)
	}

	leafSubject := pkix.Name{CommonName: "test.example.com"}
	leafCert, _, err := createTestCert(leafSubject, rootSubject, rootCert, rootKey, false)
	if err != nil {
		t.Fatalf("Failed to create leaf certificate: %v", err)
	}

	// Add root to validator
	validator.AddRootCert(rootCert)

	// Validate chain
	err = validator.ValidateChain(leafCert)
	if err != nil {
		t.Errorf("Expected valid chain, got error: %v", err)
	}
}

func TestChainValidator_ValidateChainWithHostname(t *testing.T) {
	validator := NewChainValidator()

	// Create certificate with specific DNS name
	rootSubject := pkix.Name{CommonName: "Test Root CA"}
	rootCert, rootKey, err := createTestCert(rootSubject, rootSubject, nil, nil, true)
	if err != nil {
		t.Fatalf("Failed to create root certificate: %v", err)
	}

	leafSubject := pkix.Name{CommonName: "test.example.com"}
	leafCert, _, err := createTestCert(leafSubject, rootSubject, rootCert, rootKey, false)
	if err != nil {
		t.Fatalf("Failed to create leaf certificate: %v", err)
	}

	// Add DNS name to certificate
	leafCert.DNSNames = []string{"test.example.com"}

	validator.AddRootCert(rootCert)

	// Test valid hostname
	err = validator.ValidateChainWithHostname(leafCert, "test.example.com")
	if err != nil {
		t.Errorf("Expected valid hostname validation, got error: %v", err)
	}

	// Test invalid hostname
	err = validator.ValidateChainWithHostname(leafCert, "invalid.example.com")
	if err == nil {
		t.Error("Expected error for invalid hostname")
	}
}

func TestChainValidator_BuildChain(t *testing.T) {
	validator := NewChainValidator()

	// Test with nil certificate
	chain, err := validator.BuildChain(nil)
	if err == nil {
		t.Error("Expected error for nil certificate")
	}
	if chain != nil {
		t.Error("Expected nil chain for nil certificate")
	}

	// Create a three-level certificate chain
	rootSubject := pkix.Name{CommonName: "Test Root CA"}
	rootCert, rootKey, err := createTestCert(rootSubject, rootSubject, nil, nil, true)
	if err != nil {
		t.Fatalf("Failed to create root certificate: %v", err)
	}

	intermediateSubject := pkix.Name{CommonName: "Test Intermediate CA"}
	intermediateCert, intermediateKey, err := createTestCert(intermediateSubject, rootSubject, rootCert, rootKey, true)
	if err != nil {
		t.Fatalf("Failed to create intermediate certificate: %v", err)
	}

	leafSubject := pkix.Name{CommonName: "test.example.com"}
	leafCert, _, err := createTestCert(leafSubject, intermediateSubject, intermediateCert, intermediateKey, false)
	if err != nil {
		t.Fatalf("Failed to create leaf certificate: %v", err)
	}

	// Add certificates to validator
	validator.AddRootCert(rootCert)
	validator.AddIntermediateCert(intermediateCert)

	// Build chain
	chain, err = validator.BuildChain(leafCert)
	if err != nil {
		t.Errorf("Expected successful chain building, got error: %v", err)
	}

	if len(chain) == 0 {
		t.Error("Expected non-empty chain")
	}

	if chain[0] != leafCert {
		t.Error("Expected first certificate in chain to be leaf certificate")
	}
}

func TestChainValidator_GetChainInfo(t *testing.T) {
	validator := NewChainValidator()

	// Test with empty chain
	info := validator.GetChainInfo([]*x509.Certificate{})
	if info.Length != 0 {
		t.Errorf("Expected length 0 for empty chain, got %d", info.Length)
	}

	// Create a test certificate
	rootSubject := pkix.Name{CommonName: "Test Root CA"}
	rootCert, rootKey, err := createTestCert(rootSubject, rootSubject, nil, nil, true)
	if err != nil {
		t.Fatalf("Failed to create root certificate: %v", err)
	}

	leafSubject := pkix.Name{CommonName: "test.example.com"}
	leafCert, _, err := createTestCert(leafSubject, rootSubject, rootCert, rootKey, false)
	if err != nil {
		t.Fatalf("Failed to create leaf certificate: %v", err)
	}

	chain := []*x509.Certificate{leafCert, rootCert}
	info = validator.GetChainInfo(chain)

	if info.Length != 2 {
		t.Errorf("Expected length 2, got %d", info.Length)
	}

	if info.LeafCert != leafCert {
		t.Error("Expected leaf cert to match")
	}

	if info.RootCert != rootCert {
		t.Error("Expected root cert to match")
	}

	if !info.IsValid {
		t.Error("Expected chain to be valid")
	}
}

func TestChainValidator_GetChainInfoExpired(t *testing.T) {
	validator := NewChainValidator()

	// Create an expired certificate
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Expired Test Cert"},
		NotBefore:    time.Now().Add(-time.Hour * 24 * 2), // 2 days ago
		NotAfter:     time.Now().Add(-time.Hour * 24),     // 1 day ago (expired)
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	chain := []*x509.Certificate{cert}
	info := validator.GetChainInfo(chain)

	if info.IsValid {
		t.Error("Expected expired certificate to be invalid")
	}

	if len(info.Issues) == 0 {
		t.Error("Expected issues to be reported for expired certificate")
	}
}
