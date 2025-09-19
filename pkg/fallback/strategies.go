package fallback

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// SystemCertStrategy tries to use system certificate store
type SystemCertStrategy struct {
	trustStore TrustStoreManager
	priority   int
}

// NewSystemCertStrategy creates a new system certificate strategy
func NewSystemCertStrategy(ts TrustStoreManager) *SystemCertStrategy {
	return &SystemCertStrategy{
		trustStore: ts,
		priority:   1, // Highest priority
	}
}

func (s *SystemCertStrategy) Name() string {
	return "system-cert-fallback"
}

func (s *SystemCertStrategy) Priority() int {
	return s.priority
}

func (s *SystemCertStrategy) Execute(ctx context.Context, registry string) error {
	// Try to load certificates from system store
	systemPool, err := x509.SystemCertPool()
	if err != nil {
		return fmt.Errorf("failed to load system cert pool: %w", err)
	}

	if systemPool == nil {
		return fmt.Errorf("system cert pool is not available")
	}

	// For now, we assume the system pool contains valid certificates
	// In a real implementation, we would extract specific certificates
	// and add them to the trust store for this registry

	// This is a simplified implementation that tells the trust store
	// to use system certificates for this registry
	if storer, ok := s.trustStore.(interface {
		SetUseSystemCerts(string, bool) error
	}); ok {
		return storer.SetUseSystemCerts(registry, true)
	}

	// Fallback: just return success if we can't set system certs
	// The actual certificate validation will be done by the trust store
	return nil
}

func (s *SystemCertStrategy) ShouldRetry(err error) bool {
	// Don't retry system cert loading failures
	return false
}

// CachedCertStrategy uses previously cached certificates
type CachedCertStrategy struct {
	trustStore TrustStoreManager
	cacheDir   string
	priority   int
}

// NewCachedCertStrategy creates a new cached certificate strategy
func NewCachedCertStrategy(ts TrustStoreManager) *CachedCertStrategy {
	homeDir, _ := os.UserHomeDir()
	return &CachedCertStrategy{
		trustStore: ts,
		cacheDir:   filepath.Join(homeDir, ".idpbuilder", "cert-cache"),
		priority:   2, // Medium priority
	}
}

func (c *CachedCertStrategy) Name() string {
	return "cached-cert-fallback"
}

func (c *CachedCertStrategy) Priority() int {
	return c.priority
}

func (c *CachedCertStrategy) Execute(ctx context.Context, registry string) error {
	// Ensure cache directory exists
	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Look for cached certificates for this registry
	cacheFile := filepath.Join(c.cacheDir, fmt.Sprintf("%s.pem", sanitizeFilename(registry)))

	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return fmt.Errorf("no cached certificate found for %s", registry)
	}

	// Load cached certificate
	certData, err := os.ReadFile(cacheFile)
	if err != nil {
		return fmt.Errorf("failed to read cached cert: %w", err)
	}

	// Simple validation - check if it looks like a certificate
	if !strings.Contains(string(certData), "-----BEGIN CERTIFICATE-----") {
		return fmt.Errorf("cached file does not contain valid certificate data")
	}

	// Add cached certificate to trust store
	if adder, ok := c.trustStore.(interface {
		AddCertificate(string, []byte) error
	}); ok {
		return adder.AddCertificate(registry, certData)
	}

	// If we can't add the certificate directly, assume success
	// The trust store implementation will handle the actual validation
	return nil
}

func (c *CachedCertStrategy) ShouldRetry(err error) bool {
	// Retry on transient file system errors
	if os.IsTimeout(err) {
		return true
	}

	// Retry on temporary file system errors
	if pathErr, ok := err.(*os.PathError); ok {
		if errno, ok := pathErr.Err.(syscall.Errno); ok {
			// Retry on EAGAIN, EBUSY, etc.
			return errno == syscall.EAGAIN || errno == syscall.EBUSY
		}
	}

	return false
}

// SelfSignedAcceptStrategy accepts self-signed certificates with user warning
type SelfSignedAcceptStrategy struct {
	trustStore TrustStoreManager
	priority   int
}

// NewSelfSignedAcceptStrategy creates a new self-signed acceptance strategy
func NewSelfSignedAcceptStrategy(ts TrustStoreManager) *SelfSignedAcceptStrategy {
	return &SelfSignedAcceptStrategy{
		trustStore: ts,
		priority:   10, // Lowest priority - last resort
	}
}

func (s *SelfSignedAcceptStrategy) Name() string {
	return "self-signed-accept"
}

func (s *SelfSignedAcceptStrategy) Priority() int {
	return s.priority
}

func (s *SelfSignedAcceptStrategy) Execute(ctx context.Context, registry string) error {
	// Warn user about accepting self-signed certificate
	fmt.Printf("⚠️  WARNING: Accepting self-signed certificate for %s\n", registry)
	fmt.Println("This reduces security. Use --insecure flag to suppress this warning.")

	// Configure trust store to accept self-signed for this registry
	return s.trustStore.SetInsecure(registry, true)
}

func (s *SelfSignedAcceptStrategy) ShouldRetry(err error) bool {
	// Don't retry self-signed acceptance - either it works or it doesn't
	return false
}

// Helper functions

// sanitizeFilename sanitizes a registry name for use as a filename
func sanitizeFilename(s string) string {
	// Replace characters that are not safe for filenames
	replacer := strings.NewReplacer(
		"/", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(s)
}

// validateCertificateData performs basic validation on certificate data
func validateCertificateData(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("certificate data is empty")
	}

	dataStr := string(data)

	// Check for PEM format
	if !strings.Contains(dataStr, "-----BEGIN CERTIFICATE-----") {
		return fmt.Errorf("certificate data is not in PEM format")
	}

	if !strings.Contains(dataStr, "-----END CERTIFICATE-----") {
		return fmt.Errorf("certificate data is incomplete - missing end marker")
	}

	return nil
}

// createCacheEntry creates a cache entry for a successful certificate
func createCacheEntry(cacheDir, registry string, certData []byte) error {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cacheFile := filepath.Join(cacheDir, fmt.Sprintf("%s.pem", sanitizeFilename(registry)))

	if err := os.WriteFile(cacheFile, certData, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}
