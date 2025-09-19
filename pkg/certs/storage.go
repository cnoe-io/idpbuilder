package certs

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// CertificateStorage interface for certificate persistence
type CertificateStorage interface {
	Store(name string, cert *x509.Certificate) error
	StoreAt(cert *x509.Certificate, path string) error
	Load(name string) (*x509.Certificate, error)
	Exists(name string) bool
	Remove(name string) error
	ListCertificates() ([]string, error)
}

// LocalCertStorage implements file-based certificate storage
type LocalCertStorage struct {
	baseDir string
}

// NewLocalCertStorage creates a new local certificate storage
func NewLocalCertStorage(baseDir string) (*LocalCertStorage, error) {
	// Expand home directory if needed
	expandedDir := expandHomeDir(baseDir)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(expandedDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create cert directory: %w", err)
	}

	return &LocalCertStorage{baseDir: expandedDir}, nil
}

// Store saves a certificate with the given name
func (s *LocalCertStorage) Store(name string, cert *x509.Certificate) error {
	path := filepath.Join(s.baseDir, fmt.Sprintf("%s.pem", name))
	return s.StoreAt(cert, path)
}

// StoreAt saves a certificate at a specific path
func (s *LocalCertStorage) StoreAt(cert *x509.Certificate, path string) error {
	// Encode certificate to PEM
	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}

	pemData := pem.EncodeToMemory(pemBlock)
	if pemData == nil {
		return fmt.Errorf("failed to encode certificate")
	}

	// Write to file with secure permissions
	if err := ioutil.WriteFile(path, pemData, 0600); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	return nil
}

// Load retrieves a certificate by name
func (s *LocalCertStorage) Load(name string) (*x509.Certificate, error) {
	path := filepath.Join(s.baseDir, fmt.Sprintf("%s.pem", name))

	// Read certificate file
	pemData, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrCertNotFound
		}
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	// Parse PEM block
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// Exists checks if a certificate exists
func (s *LocalCertStorage) Exists(name string) bool {
	path := filepath.Join(s.baseDir, fmt.Sprintf("%s.pem", name))
	_, err := os.Stat(path)
	return err == nil
}

// Remove deletes a certificate
func (s *LocalCertStorage) Remove(name string) error {
	path := filepath.Join(s.baseDir, fmt.Sprintf("%s.pem", name))
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return ErrCertNotFound
		}
		return fmt.Errorf("failed to remove certificate: %w", err)
	}
	return nil
}

// ListCertificates returns all stored certificate names
func (s *LocalCertStorage) ListCertificates() ([]string, error) {
	// Read directory
	files, err := ioutil.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate directory: %w", err)
	}

	var certNames []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		name := file.Name()
		if strings.HasSuffix(name, ".pem") {
			// Return name without extension
			certName := strings.TrimSuffix(name, ".pem")
			certNames = append(certNames, certName)
		}
	}

	return certNames, nil
}