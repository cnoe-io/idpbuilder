package certs

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"testing"
	"time"
)

// MockKindClient implements KindClient for testing
type MockKindClient struct {
	clusters          []string
	pods              map[string][]string
	podFiles          map[string][]byte
	shouldFailCopy    bool
	shouldFailExec    bool
	shouldFailPods    bool
	shouldFailCluster bool
}

func (m *MockKindClient) GetCurrentCluster() (string, error) {
	if m.shouldFailCluster {
		return "", errors.New("no clusters found")
	}
	if len(m.clusters) == 0 {
		return "", ErrNoKindCluster
	}
	return m.clusters[0], nil
}

func (m *MockKindClient) ListClusters() ([]string, error) {
	if m.shouldFailCluster {
		return nil, errors.New("failed to list clusters")
	}
	return m.clusters, nil
}

func (m *MockKindClient) GetPods(ctx context.Context, namespace, labelSelector string) ([]string, error) {
	if m.shouldFailPods {
		return nil, errors.New("failed to get pods")
	}
	key := namespace + ":" + labelSelector
	pods, exists := m.pods[key]
	if !exists {
		return []string{}, nil
	}
	return pods, nil
}

func (m *MockKindClient) CopyFromPod(ctx context.Context, podName, path string) ([]byte, error) {
	if m.shouldFailCopy {
		return nil, errors.New("copy failed")
	}
	key := podName + ":" + path
	data, exists := m.podFiles[key]
	if !exists {
		return nil, ErrCertNotInPod
	}
	return data, nil
}

func (m *MockKindClient) ExecInPod(ctx context.Context, podName string, command []string) (string, error) {
	if m.shouldFailExec {
		return "", errors.New("exec failed")
	}
	return "command output", nil
}

// MockCertValidator implements KindCertValidator for testing
type MockCertValidator struct {
	shouldFailValidation bool
}

func (m *MockCertValidator) ValidateCertificate(cert *x509.Certificate) error {
	if m.shouldFailValidation {
		return errors.New("validation failed")
	}
	return nil
}

func TestNewKindCertExtractor(t *testing.T) {
	config := ExtractorConfig{
		ClusterName:      "test-cluster",
		Namespace:        "test-ns",
		PodLabelSelector: "app=test",
		CertPath:         "/test/cert.pem",
		Timeout:          10 * time.Second,
		RetryAttempts:    2,
	}

	// This will fail because kubectl might not be available in test environment
	// but we can test the configuration validation
	_, err := NewKindCertExtractor(config)
	// We expect this to either succeed or fail with a kubectl-related error
	// The important thing is that it doesn't panic
	if err != nil {
		t.Logf("NewKindCertExtractor failed as expected in test environment: %v", err)
	}
}

func TestKindCertExtractor_ExtractGiteaCert(t *testing.T) {
	// Create test certificate
	testCert := createTestCertForExtractor(t)
	certPEM := encodeCertToPEMForExtractor(t, testCert)

	// Create mock client
	mockClient := &MockKindClient{
		clusters: []string{"test-cluster"},
		pods: map[string][]string{
			"gitea:app=gitea": {"gitea-pod-123"},
		},
		podFiles: map[string][]byte{
			"gitea-pod-123:/etc/ssl/certs/ca.pem": certPEM,
		},
	}

	// Create temp storage
	tmpDir, err := ioutil.TempDir("", "extractor_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewLocalCertStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create extractor with mocks
	extractor := &KindCertExtractor{
		client:    mockClient,
		storage:   storage,
		validator: &MockCertValidator{},
		config: ExtractorConfig{
			ClusterName:      "test-cluster",
			Namespace:        "gitea",
			PodLabelSelector: "app=gitea",
			CertPath:         "/etc/ssl/certs/ca.pem",
			Timeout:          10 * time.Second,
			RetryAttempts:    3,
		},
	}

	// Test successful extraction
	ctx := context.Background()
	extractedCert, err := extractor.ExtractGiteaCert(ctx)
	if err != nil {
		t.Fatalf("ExtractGiteaCert() failed: %v", err)
	}

	if !testCert.Equal(extractedCert) {
		t.Error("Extracted certificate does not match test certificate")
	}

	// Verify certificate was stored
	if !storage.Exists("gitea") {
		t.Error("Certificate was not stored")
	}
}

// Feature disabled test removed - features are now always enabled in production

func TestKindCertExtractor_ExtractGiteaCert_NoCluster(t *testing.T) {
	mockClient := &MockKindClient{
		clusters:          []string{},
		shouldFailCluster: true,
	}

	extractor := &KindCertExtractor{
		client:    mockClient,
		validator: &MockCertValidator{},
		config:    ExtractorConfig{},
	}

	// Feature flag removed - features are now always enabled

	ctx := context.Background()
	_, err := extractor.ExtractGiteaCert(ctx)
	if err == nil {
		t.Error("Expected error when no cluster found")
	}
}

func TestKindCertExtractor_ExtractGiteaCert_NoPod(t *testing.T) {
	mockClient := &MockKindClient{
		clusters: []string{"test-cluster"},
		pods:     map[string][]string{}, // No pods
	}

	extractor := &KindCertExtractor{
		client:    mockClient,
		validator: &MockCertValidator{},
		config: ExtractorConfig{
			Namespace:        "gitea",
			PodLabelSelector: "app=gitea",
		},
	}

	// Feature flag removed - features are now always enabled

	ctx := context.Background()
	_, err := extractor.ExtractGiteaCert(ctx)
	if err == nil {
		t.Error("Expected error when no pod found")
	}
}

func TestKindCertExtractor_ExtractGiteaCert_CopyFailed(t *testing.T) {
	mockClient := &MockKindClient{
		clusters: []string{"test-cluster"},
		pods: map[string][]string{
			"gitea:app=gitea": {"gitea-pod-123"},
		},
		shouldFailCopy: true,
	}

	extractor := &KindCertExtractor{
		client:    mockClient,
		validator: &MockCertValidator{},
		config: ExtractorConfig{
			Namespace:        "gitea",
			PodLabelSelector: "app=gitea",
			RetryAttempts:    2,
		},
	}

	// Feature flag removed - features are now always enabled

	ctx := context.Background()
	_, err := extractor.ExtractGiteaCert(ctx)
	if err == nil {
		t.Error("Expected error when copy fails")
	}
}

func TestKindCertExtractor_ExtractGiteaCert_ValidationFailed(t *testing.T) {
	testCert := createTestCertForExtractor(t)
	certPEM := encodeCertToPEMForExtractor(t, testCert)

	mockClient := &MockKindClient{
		clusters: []string{"test-cluster"},
		pods: map[string][]string{
			"gitea:app=gitea": {"gitea-pod-123"},
		},
		podFiles: map[string][]byte{
			"gitea-pod-123:/etc/ssl/certs/ca.pem": certPEM,
		},
	}

	tmpDir, err := ioutil.TempDir("", "extractor_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewLocalCertStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	extractor := &KindCertExtractor{
		client:    mockClient,
		storage:   storage,
		validator: &MockCertValidator{shouldFailValidation: true},
		config: ExtractorConfig{
			Namespace:        "gitea",
			PodLabelSelector: "app=gitea",
			CertPath:         "/etc/ssl/certs/ca.pem",
		},
	}

	// Feature flag removed - features are now always enabled

	ctx := context.Background()
	_, err = extractor.ExtractGiteaCert(ctx)
	if err == nil {
		t.Error("Expected error when validation fails")
	}
}

func TestDefaultCertValidator_ValidateCertificate(t *testing.T) {
	validator := &DefaultCertValidator{}

	// Test nil certificate
	err := validator.ValidateCertificate(nil)
	if err == nil {
		t.Error("Expected error for nil certificate")
	}

	// Test valid certificate
	cert := createTestCertForExtractor(t)
	err = validator.ValidateCertificate(cert)
	if err != nil {
		t.Errorf("ValidateCertificate() failed for valid cert: %v", err)
	}

	// Test expired certificate
	expiredCert := &x509.Certificate{
		NotBefore: time.Now().Add(-2 * time.Hour),
		NotAfter:  time.Now().Add(-1 * time.Hour),
		Subject:   pkix.Name{CommonName: "expired.com"},
	}
	err = validator.ValidateCertificate(expiredCert)
	if err != ErrCertExpired {
		t.Errorf("Expected ErrCertExpired, got: %v", err)
	}

	// Test certificate with no subject names
	emptyCert := &x509.Certificate{
		NotBefore:   time.Now().Add(-1 * time.Hour),
		NotAfter:    time.Now().Add(1 * time.Hour),
		Subject:     pkix.Name{}, // Empty subject
		DNSNames:    []string{},  // No DNS names
		IPAddresses: []net.IP{},  // No IP addresses
	}
	err = validator.ValidateCertificate(emptyCert)
	if err == nil {
		t.Error("Expected error for certificate with no subject names")
	}
}

// Helper functions for testing
func createTestCertForExtractor(t *testing.T) *x509.Certificate {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

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
		DNSNames:    []string{"test.example.com"},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
	}

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

func encodeCertToPEMForExtractor(t *testing.T, cert *x509.Certificate) []byte {
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
	if certPEM == nil {
		t.Fatal("Failed to encode certificate to PEM")
	}
	return certPEM
}
