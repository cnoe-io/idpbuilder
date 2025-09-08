package certs

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// TrustStoreManager manages trusted certificates for registries (consolidated interface)
type TrustStoreManager interface {
	AddCertificate(registry string, cert *x509.Certificate) error
	GetTrustedCerts(registry string) ([]*x509.Certificate, error)
	SetInsecure(registry string, insecure bool) error
	IsInsecure(registry string) bool
	CreateHTTPClient(registry string) (*http.Client, error)
	ConfigureTransport(registry string) (remote.Option, error)
}

// DefaultTrustStore implements TrustStoreManager with minimal functionality
type DefaultTrustStore struct {
	mu           sync.RWMutex
	trustedCerts map[string][]*x509.Certificate
	insecure     map[string]bool
	certDir      string
}

// NewTrustStore creates a new trust store
func NewTrustStore() *DefaultTrustStore {
	certDir := getConfigDir() + "/certs"
	os.MkdirAll(certDir, 0700)

	store := &DefaultTrustStore{
		trustedCerts: make(map[string][]*x509.Certificate),
		insecure:     make(map[string]bool),
		certDir:      certDir,
	}
	store.loadFromDisk()
	return store
}

// AddCertificate adds a certificate for a registry
func (ts *DefaultTrustStore) AddCertificate(registry string, cert *x509.Certificate) error {
	if !isRegistryFeatureEnabled("REGISTRY_TLS_TRUST_ENABLED") {
		return fmt.Errorf("REGISTRY_TLS_TRUST feature disabled")
	}

	if err := validateCert(cert); err != nil {
		logSecurityEvent("CERT_INVALID", registry, err.Error())
		return err
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.trustedCerts[registry] = append(ts.trustedCerts[registry], cert)

	// Save to disk
	path := filepath.Join(ts.certDir, sanitizeName(registry)+".pem")
	if err := saveCertToPEM(path, cert); err != nil {
		return fmt.Errorf("failed to save cert: %w", err)
	}

	logSecurityEvent("CERT_ADDED", registry, "Certificate added")
	return nil
}

// GetTrustedCerts returns certificates for a registry
func (ts *DefaultTrustStore) GetTrustedCerts(registry string) ([]*x509.Certificate, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.trustedCerts[registry], nil
}

// SetInsecure marks a registry as insecure
func (ts *DefaultTrustStore) SetInsecure(registry string, insecure bool) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if insecure {
		ts.insecure[registry] = true
		logSecurityEvent("INSECURE_SET", registry, "Registry marked insecure")
	} else {
		delete(ts.insecure, registry)
	}
	return nil
}

// IsInsecure checks if registry is insecure
func (ts *DefaultTrustStore) IsInsecure(registry string) bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.insecure[registry]
}

// CreateHTTPClient creates an HTTP client with proper TLS config
func (ts *DefaultTrustStore) CreateHTTPClient(registry string) (*http.Client, error) {
	tlsConfig, err := ts.buildTLSConfig(registry)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		TLSClientConfig:     tlsConfig,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}, nil
}

// ConfigureTransport creates go-containerregistry transport option
func (ts *DefaultTrustStore) ConfigureTransport(registry string) (remote.Option, error) {
	client, err := ts.CreateHTTPClient(registry)
	if err != nil {
		return nil, err
	}
	return remote.WithTransport(client.Transport), nil
}

// buildTLSConfig builds TLS configuration for a registry
func (ts *DefaultTrustStore) buildTLSConfig(registry string) (*tls.Config, error) {
	if ts.IsInsecure(registry) {
		logSecurityEvent("TLS_INSECURE", registry, "Insecure mode")
		return &tls.Config{InsecureSkipVerify: true}, nil
	}

	// Create cert pool with system + custom certs
	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}

	// Add custom certs
	certs, _ := ts.GetTrustedCerts(registry)
	for _, cert := range certs {
		pool.AddCert(cert)
	}

	config := &tls.Config{
		RootCAs:    pool,
		MinVersion: tls.VersionTLS12,
		ServerName: extractHostname(registry),
	}

	return config, nil
}

// loadFromDisk loads certificates from disk
func (ts *DefaultTrustStore) loadFromDisk() {
	files, err := os.ReadDir(ts.certDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".pem") {
			continue
		}

		registry := unsanitizeName(strings.TrimSuffix(file.Name(), ".pem"))
		path := filepath.Join(ts.certDir, file.Name())

		cert, err := loadCertFromPEM(path)
		if err != nil {
			continue
		}

		ts.trustedCerts[registry] = append(ts.trustedCerts[registry], cert)
	}
}

// Helper functions

func validateCert(cert *x509.Certificate) error {
	if time.Now().After(cert.NotAfter) {
		return fmt.Errorf("certificate expired")
	}
	return nil
}

func saveCertToPEM(path string, cert *x509.Certificate) error {
	os.MkdirAll(filepath.Dir(path), 0700)
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
	return os.WriteFile(path, pemData, 0600)
}

func loadCertFromPEM(path string) (*x509.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}

	return x509.ParseCertificate(block.Bytes)
}

func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "/", "_")
	return strings.ReplaceAll(name, ".", "_")
}

func unsanitizeName(name string) string {
	name = strings.ReplaceAll(name, "_", ":")
	if strings.Count(name, ":") > 1 {
		// Fix over-replacement
		parts := strings.Split(name, ":")
		if len(parts) >= 2 {
			name = parts[0] + "." + strings.Join(parts[1:], ":")
		}
	}
	return name
}

func extractHostname(registry string) string {
	registry = strings.TrimPrefix(registry, "https://")
	registry = strings.TrimPrefix(registry, "http://")

	if host, _, err := net.SplitHostPort(registry); err == nil {
		return host
	}

	if idx := strings.Index(registry, "/"); idx != -1 {
		return registry[:idx]
	}

	return registry
}

func getConfigDir() string {
	if dir := os.Getenv("IDPBUILDER_CONFIG_DIR"); dir != "" {
		return dir
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".idpbuilder")
}

func isRegistryFeatureEnabled(flag string) bool {
	value := strings.ToLower(os.Getenv("IDPBUILDER_" + flag))
	return value == "true" || value == "1" || value == "enabled"
}

func logSecurityEvent(event, target, message string) {
	fmt.Printf("[SECURITY] %s: %s - %s\n", event, target, message)
}
