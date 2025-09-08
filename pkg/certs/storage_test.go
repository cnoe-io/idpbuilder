package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewLocalCertStorage(t *testing.T) {
	// Create temporary directory
	tmpDir, err := ioutil.TempDir("", "cert_storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewLocalCertStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewLocalCertStorage() failed: %v", err)
	}

	if storage == nil {
		t.Fatal("NewLocalCertStorage() returned nil")
	}

	// Check directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Storage directory was not created")
	}
}

func TestNewLocalCertStorage_HomeDir(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Create temp home
	tmpHome, err := ioutil.TempDir("", "home_test")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tmpHome)
	os.Setenv("HOME", tmpHome)

	storage, err := NewLocalCertStorage("~/certs")
	if err != nil {
		t.Fatalf("NewLocalCertStorage() failed: %v", err)
	}

	expectedDir := filepath.Join(tmpHome, "certs")
	if storage.baseDir != expectedDir {
		t.Errorf("Expected baseDir %s, got %s", expectedDir, storage.baseDir)
	}

	// Check directory was created
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Error("Home-based storage directory was not created")
	}
}

func TestLocalCertStorage_StoreAndLoad(t *testing.T) {
	// Setup
	tmpDir, err := ioutil.TempDir("", "cert_storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewLocalCertStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewLocalCertStorage() failed: %v", err)
	}

	// Create test certificate
	cert := createTestCert(t)

	// Test Store
	err = storage.Store("test-cert", cert)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	// Test Load
	loadedCert, err := storage.Load("test-cert")
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if !cert.Equal(loadedCert) {
		t.Error("Loaded certificate does not match stored certificate")
	}
}

func TestLocalCertStorage_LoadNonexistent(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "cert_storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewLocalCertStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewLocalCertStorage() failed: %v", err)
	}

	_, err = storage.Load("nonexistent")
	if err != ErrCertNotFound {
		t.Errorf("Expected ErrCertNotFound, got: %v", err)
	}
}

func TestLocalCertStorage_Exists(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "cert_storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewLocalCertStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewLocalCertStorage() failed: %v", err)
	}

	// Test non-existent certificate
	if storage.Exists("test-cert") {
		t.Error("Exists() should return false for non-existent certificate")
	}

	// Store a certificate
	cert := createTestCert(t)
	err = storage.Store("test-cert", cert)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	// Test existing certificate
	if !storage.Exists("test-cert") {
		t.Error("Exists() should return true for existing certificate")
	}
}

func TestLocalCertStorage_Remove(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "cert_storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewLocalCertStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewLocalCertStorage() failed: %v", err)
	}

	// Store a certificate
	cert := createTestCert(t)
	err = storage.Store("test-cert", cert)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	// Remove the certificate
	err = storage.Remove("test-cert")
	if err != nil {
		t.Fatalf("Remove() failed: %v", err)
	}

	// Verify it's gone
	if storage.Exists("test-cert") {
		t.Error("Certificate should not exist after removal")
	}

	// Try to remove non-existent certificate
	err = storage.Remove("nonexistent")
	if err != ErrCertNotFound {
		t.Errorf("Expected ErrCertNotFound, got: %v", err)
	}
}

func TestLocalCertStorage_ListCertificates(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "cert_storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewLocalCertStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewLocalCertStorage() failed: %v", err)
	}

	// Test empty directory
	certs, err := storage.ListCertificates()
	if err != nil {
		t.Fatalf("ListCertificates() failed: %v", err)
	}
	if len(certs) != 0 {
		t.Errorf("Expected 0 certificates, got %d", len(certs))
	}

	// Store some certificates
	cert := createTestCert(t)
	names := []string{"cert1", "cert2", "cert3"}
	for _, name := range names {
		err = storage.Store(name, cert)
		if err != nil {
			t.Fatalf("Store(%s) failed: %v", name, err)
		}
	}

	// List certificates
	certs, err = storage.ListCertificates()
	if err != nil {
		t.Fatalf("ListCertificates() failed: %v", err)
	}

	if len(certs) != len(names) {
		t.Errorf("Expected %d certificates, got %d", len(names), len(certs))
	}

	// Check all names are present
	for _, expectedName := range names {
		found := false
		for _, actualName := range certs {
			if actualName == expectedName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Certificate %s not found in list", expectedName)
		}
	}
}

func TestLocalCertStorage_StoreAt(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "cert_storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewLocalCertStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewLocalCertStorage() failed: %v", err)
	}

	cert := createTestCert(t)
	customPath := filepath.Join(tmpDir, "custom-cert.pem")

	// Store at custom path
	err = storage.StoreAt(cert, customPath)
	if err != nil {
		t.Fatalf("StoreAt() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(customPath); os.IsNotExist(err) {
		t.Error("Certificate file was not created at custom path")
	}

	// Verify file has correct permissions
	fileInfo, err := os.Stat(customPath)
	if err != nil {
		t.Fatalf("Failed to stat certificate file: %v", err)
	}
	if fileInfo.Mode().Perm() != 0600 {
		t.Errorf("Expected file permissions 0600, got %o", fileInfo.Mode().Perm())
	}
}

// Helper function to create test certificate
func createTestCert(t *testing.T) *x509.Certificate {
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
			CommonName:   "test.example.com",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    []string{"test.example.com", "localhost"},
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
