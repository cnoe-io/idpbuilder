package certs

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// CertPoolManager manages certificate pools with basic caching
type CertPoolManager struct {
	mu    sync.RWMutex
	pools map[string]*x509.CertPool
	store TrustStoreManager
}

// NewCertPoolManager creates a certificate pool manager
func NewCertPoolManager(store TrustStoreManager) *CertPoolManager {
	return &CertPoolManager{
		pools: make(map[string]*x509.CertPool),
		store: store,
	}
}

// GetPool returns a certificate pool for a registry
func (m *CertPoolManager) GetPool(registry string) (*x509.CertPool, error) {
	m.mu.RLock()
	if pool, exists := m.pools[registry]; exists {
		m.mu.RUnlock()
		return pool, nil
	}
	m.mu.RUnlock()
	
	// Build new pool
	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}
	
	// Add custom certificates
	certs, err := m.store.GetTrustedCerts(registry)
	if err != nil {
		return nil, err
	}
	
	for _, cert := range certs {
		pool.AddCert(cert)
	}
	
	// Cache it
	m.mu.Lock()
	m.pools[registry] = pool
	m.mu.Unlock()
	
	return pool, nil
}

// ClearCache clears the certificate pool cache
func (m *CertPoolManager) ClearCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pools = make(map[string]*x509.CertPool)
}

// TransportConfigurer configures HTTP transports
type TransportConfigurer struct {
	trustManager TrustStoreManager
	timeout      time.Duration
}

// NewTransportConfigurer creates a transport configurer
func NewTransportConfigurer(manager TrustStoreManager) *TransportConfigurer {
	return &TransportConfigurer{
		trustManager: manager,
		timeout:      30 * time.Second,
	}
}

// ConfigureTransport creates a configured transport
func (c *TransportConfigurer) ConfigureTransport(registry string) (*http.Transport, error) {
	tlsConfig, err := c.buildTLSConfig(registry)
	if err != nil {
		return nil, err
	}
	
	return &http.Transport{
		TLSClientConfig:     tlsConfig,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}, nil
}

// buildTLSConfig builds TLS configuration
func (c *TransportConfigurer) buildTLSConfig(registry string) (*tls.Config, error) {
	if c.trustManager.IsInsecure(registry) {
		return &tls.Config{InsecureSkipVerify: true}, nil
	}
	
	// Get certificates
	certs, err := c.trustManager.GetTrustedCerts(registry)
	if err != nil {
		return nil, err
	}
	
	// Build cert pool
	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}
	
	for _, cert := range certs {
		pool.AddCert(cert)
	}
	
	return &tls.Config{
		RootCAs:    pool,
		MinVersion: tls.VersionTLS12,
		ServerName: extractHostname(registry),
	}, nil
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Registry           string
	InsecureSkipVerify bool
	MinVersion         uint16
	ValidateHostname   bool
	Timeout            time.Duration
}

// DefaultTLSConfig returns default TLS config
func DefaultTLSConfig() *TLSConfig {
	return &TLSConfig{
		MinVersion:       tls.VersionTLS12,
		ValidateHostname: true,
		Timeout:          10 * time.Second,
	}
}

// LoadConfigFromEnv loads configuration from environment
func (c *TLSConfig) LoadConfigFromEnv() {
	if os.Getenv("IDPBUILDER_TLS_INSECURE") == "true" {
		c.InsecureSkipVerify = true
	}
	
	if timeout := os.Getenv("IDPBUILDER_TLS_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			c.Timeout = d
		}
	}
}

// ToGoTLSConfig converts to standard tls.Config
func (c *TLSConfig) ToGoTLSConfig() *tls.Config {
	config := &tls.Config{
		InsecureSkipVerify: c.InsecureSkipVerify,
		MinVersion:         c.MinVersion,
	}
	
	if c.ValidateHostname && c.Registry != "" {
		config.ServerName = extractHostname(c.Registry)
	}
	
	return config
}

// SecurityLogger handles security audit logging
type SecurityLogger struct {
	mu   sync.Mutex
	file *os.File
}

var globalLogger *SecurityLogger

// InitSecurityLogger initializes security logging
func InitSecurityLogger() error {
	logPath := filepath.Join(getConfigDir(), "security.log")
	
	os.MkdirAll(filepath.Dir(logPath), 0700)
	
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	
	globalLogger = &SecurityLogger{file: file}
	return nil
}

// LogSecurityEvent logs a security event
func LogSecurityEvent(event, target, message string) {
	if globalLogger == nil {
		// Fallback to stdout
		fmt.Printf("[SECURITY] %s: %s - %s\n", event, target, message)
		return
	}
	
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	logLine := fmt.Sprintf("%s [SECURITY] %s: %s - %s\n", timestamp, event, target, message)
	globalLogger.file.WriteString(logLine)
}

// CloseSecurityLogger closes the security logger
func CloseSecurityLogger() error {
	if globalLogger != nil && globalLogger.file != nil {
		return globalLogger.file.Close()
	}
	return nil
}

// ValidationResult represents certificate validation result
type ValidationResult struct {
	Valid   bool
	Message string
	Actions []string
}

// RegistryCertValidator provides certificate validation
type RegistryCertValidator struct{}

// NewRegistryCertValidator creates a certificate validator
func NewRegistryCertValidator() *RegistryCertValidator {
	return &RegistryCertValidator{}
}

// NewCertValidator creates a certificate validator (alias for NewRegistryCertValidator)
func NewCertValidator() *RegistryCertValidator {
	return NewRegistryCertValidator()
}

// ValidateCertificate validates a certificate
func (v *RegistryCertValidator) ValidateCertificate(cert *x509.Certificate) *ValidationResult {
	now := time.Now()
	
	if now.Before(cert.NotBefore) {
		return &ValidationResult{
			Valid:   false,
			Message: "Certificate not yet valid",
		}
	}
	
	if now.After(cert.NotAfter) {
		return &ValidationResult{
			Valid:   false,
			Message: "Certificate expired",
		}
	}
	
	// Warn if expiring soon
	if now.Add(30 * 24 * time.Hour).After(cert.NotAfter) {
		return &ValidationResult{
			Valid:   true,
			Message: "Certificate expires within 30 days",
			Actions: []string{"renew certificate"},
		}
	}
	
	return &ValidationResult{
		Valid:   true,
		Message: "Certificate is valid",
	}
}

// parseEnvBool parses boolean from environment variable
func parseEnvBool(value string) bool {
	lower := strings.ToLower(value)
	return lower == "true" || lower == "1" || lower == "yes" || lower == "enabled"
}

// Utility functions for configuration directories
func getConfigDirPath() string {
	return getConfigDir()
}

// Helper function to load certificates from directory
func loadCertificatesFromDir(dir string) ([]*x509.Certificate, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}
	
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	
	var certs []*x509.Certificate
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".pem") {
			continue
		}
		
		path := filepath.Join(dir, file.Name())
		cert, err := loadCertFromPEM(path)
		if err != nil {
			continue // Skip invalid certificates
		}
		
		certs = append(certs, cert)
	}
	
	return certs, nil
}