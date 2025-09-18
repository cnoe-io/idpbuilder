package certs

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// KindCertExtractor handles certificate extraction from Kind clusters
type KindCertExtractor struct {
	client    KindClient
	storage   CertificateStorage
	validator KindCertValidator
	config    ExtractorConfig
}

// ExtractorConfig holds configuration for the extractor
type ExtractorConfig struct {
	ClusterName      string
	Namespace        string
	PodLabelSelector string
	CertPath         string // Path inside the pod
	Timeout          time.Duration
	RetryAttempts    int
}

// KindCertValidator interface for certificate validation
type KindCertValidator interface {
	ValidateCertificate(cert *x509.Certificate) error
}

// DefaultCertValidator implements KindCertValidator interface for basic certificate validation
type DefaultCertValidator struct{}

// ValidateCertificate performs basic certificate validation
func (v *DefaultCertValidator) ValidateCertificate(cert *x509.Certificate) error {
	if cert == nil {
		return fmt.Errorf("certificate is nil")
	}

	// Check certificate expiry
	if err := validateCertificateExpiry(cert); err != nil {
		return err
	}

	// Check basic certificate properties
	if len(cert.Subject.CommonName) == 0 && len(cert.DNSNames) == 0 && len(cert.IPAddresses) == 0 {
		return fmt.Errorf("certificate has no valid subject names")
	}

	return nil
}

// NewKindCertExtractor creates a new certificate extractor
func NewKindCertExtractor(config ExtractorConfig) (*KindCertExtractor, error) {
	// Initialize kubectl client
	client, err := NewKubectlKindClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create kubectl client: %w", err)
	}

	// Validate configuration
	if config.CertPath == "" {
		config.CertPath = "/etc/ssl/certs/ca.pem" // Default certificate path
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}

	// Setup storage directory
	storageDir := filepath.Join(os.Getenv("HOME"), ".idpbuilder", "certs")
	storage, err := NewLocalCertStorage(storageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate storage: %w", err)
	}

	return &KindCertExtractor{
		client:    client,
		storage:   storage,
		validator: &DefaultCertValidator{},
		config:    config,
	}, nil
}

// ExtractGiteaCert extracts the Gitea certificate from Kind cluster
func (e *KindCertExtractor) ExtractGiteaCert(ctx context.Context) (*x509.Certificate, error) {
	// 1. Get cluster information
	clusterName, err := e.getClusterName()
	if err != nil {
		return nil, NewCertError("extraction", "cluster_discovery",
			fmt.Errorf("failed to get cluster name: %w", err))
	}

	// 2. Find Gitea pod
	podName, err := e.findGiteaPod(ctx, clusterName)
	if err != nil {
		return nil, NewCertError("extraction", "pod_discovery",
			fmt.Errorf("failed to find Gitea pod: %w", err))
	}

	// 3. Extract certificate data with retry
	var certData []byte
	for attempt := 0; attempt < e.config.RetryAttempts; attempt++ {
		certData, err = e.client.CopyFromPod(ctx, podName, e.config.CertPath)
		if err == nil {
			break
		}
		if attempt < e.config.RetryAttempts-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}
	if err != nil {
		return nil, NewCertError("extraction", "file_copy",
			fmt.Errorf("failed to copy certificate after %d attempts: %w", e.config.RetryAttempts, err))
	}

	// 4. Parse certificate
	cert, err := parseCertificate(certData)
	if err != nil {
		return nil, NewCertError("extraction", "certificate_parse",
			fmt.Errorf("failed to parse certificate: %w", err))
	}

	// 5. Validate certificate
	if err := e.validator.ValidateCertificate(cert); err != nil {
		return nil, NewCertError("extraction", "certificate_validation",
			fmt.Errorf("certificate validation failed: %w", err))
	}

	// 6. Store locally
	if err := e.storage.Store("gitea", cert); err != nil {
		return nil, NewCertError("extraction", "storage",
			fmt.Errorf("failed to store certificate: %w", err))
	}

	return cert, nil
}

// GetClusterName returns the current Kind cluster name
func (e *KindCertExtractor) GetClusterName() (string, error) {
	if e.config.ClusterName != "" {
		return e.config.ClusterName, nil
	}
	return e.client.GetCurrentCluster()
}

// ValidateCertificate performs basic certificate validation
func (e *KindCertExtractor) ValidateCertificate(cert *x509.Certificate) error {
	return e.validator.ValidateCertificate(cert)
}

// StoreCertificate saves the certificate to local storage
func (e *KindCertExtractor) StoreCertificate(cert *x509.Certificate, name string) error {
	return e.storage.Store(name, cert)
}

// StoreCertificateAt saves the certificate to a specific path
func (e *KindCertExtractor) StoreCertificateAt(cert *x509.Certificate, path string) error {
	return e.storage.StoreAt(cert, path)
}

// LoadCertificate loads a certificate from storage
func (e *KindCertExtractor) LoadCertificate(name string) (*x509.Certificate, error) {
	return e.storage.Load(name)
}

// CertificateExists checks if a certificate exists in storage
func (e *KindCertExtractor) CertificateExists(name string) bool {
	return e.storage.Exists(name)
}

// ListCertificates returns all stored certificate names
func (e *KindCertExtractor) ListCertificates() ([]string, error) {
	return e.storage.ListCertificates()
}

// Close cleans up resources
func (e *KindCertExtractor) Close() error {
	// In this implementation, there's nothing to clean up
	// But this method is provided for future extensibility
	return nil
}
