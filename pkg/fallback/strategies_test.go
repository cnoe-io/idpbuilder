package fallback

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestSystemCertStrategy(t *testing.T) {
	mockStore := NewMockTrustStore()
	strategy := NewSystemCertStrategy(mockStore)

	if strategy.Name() != "system-cert-fallback" {
		t.Errorf("Expected name 'system-cert-fallback', got '%s'", strategy.Name())
	}

	if strategy.Priority() != 1 {
		t.Errorf("Expected priority 1, got %d", strategy.Priority())
	}

	// Test execution with mock store that supports SetUseSystemCerts
	ctx := context.Background()
	err := strategy.Execute(ctx, "test.registry")

	// Should succeed because our mock store supports the interface
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify SetUseSystemCerts was called
	if len(mockStore.useSystemCertsCalls) != 1 {
		t.Errorf("Expected 1 SetUseSystemCerts call, got %d", len(mockStore.useSystemCertsCalls))
	}

	if mockStore.useSystemCertsCalls[0].Registry != "test.registry" {
		t.Errorf("Expected registry 'test.registry', got '%s'", mockStore.useSystemCertsCalls[0].Registry)
	}

	if !mockStore.useSystemCertsCalls[0].Use {
		t.Error("Expected Use to be true")
	}

	// Test ShouldRetry
	if strategy.ShouldRetry(nil) {
		t.Error("SystemCertStrategy should not retry on any error")
	}
}

func TestCachedCertStrategy(t *testing.T) {
	// Create temp directory for cache
	tmpDir := t.TempDir()

	mockStore := NewMockTrustStore()
	strategy := &CachedCertStrategy{
		trustStore: mockStore,
		cacheDir:   tmpDir,
		priority:   2,
	}

	if strategy.Name() != "cached-cert-fallback" {
		t.Errorf("Expected name 'cached-cert-fallback', got '%s'", strategy.Name())
	}

	if strategy.Priority() != 2 {
		t.Errorf("Expected priority 2, got %d", strategy.Priority())
	}

	// Test with no cached cert
	ctx := context.Background()
	err := strategy.Execute(ctx, "test.registry")
	if err == nil || !strings.Contains(err.Error(), "no cached certificate found") {
		t.Errorf("Expected 'no cached certificate found' error, got: %v", err)
	}

	// Create a cached cert file
	cacheFile := filepath.Join(tmpDir, "test.registry.pem")
	testCert := []byte("-----BEGIN CERTIFICATE-----\ntest certificate data\n-----END CERTIFICATE-----")
	if err := os.WriteFile(cacheFile, testCert, 0644); err != nil {
		t.Fatalf("Failed to write test cert file: %v", err)
	}

	// Test with cached cert
	err = strategy.Execute(ctx, "test.registry")
	if err != nil {
		t.Errorf("Expected no error with cached cert, got: %v", err)
	}

	// Verify AddCertificate was called
	if len(mockStore.addCertificateCalls) != 1 {
		t.Errorf("Expected 1 AddCertificate call, got %d", len(mockStore.addCertificateCalls))
	}

	if mockStore.addCertificateCalls[0].Registry != "test.registry" {
		t.Errorf("Expected registry 'test.registry', got '%s'", mockStore.addCertificateCalls[0].Registry)
	}

	// Test ShouldRetry with various errors
	if !strategy.ShouldRetry(&os.PathError{Err: syscall.EAGAIN}) {
		t.Error("Should retry on EAGAIN error")
	}

	if strategy.ShouldRetry(&os.PathError{Err: syscall.ENOENT}) {
		t.Error("Should not retry on ENOENT error")
	}
}

func TestCachedCertStrategy_InvalidCert(t *testing.T) {
	tmpDir := t.TempDir()

	mockStore := NewMockTrustStore()
	strategy := &CachedCertStrategy{
		trustStore: mockStore,
		cacheDir:   tmpDir,
		priority:   2,
	}

	// Create a cached file with invalid certificate data
	cacheFile := filepath.Join(tmpDir, "test.registry.pem")
	invalidCert := []byte("not a certificate")
	if err := os.WriteFile(cacheFile, invalidCert, 0644); err != nil {
		t.Fatalf("Failed to write invalid cert file: %v", err)
	}

	ctx := context.Background()
	err := strategy.Execute(ctx, "test.registry")
	if err == nil || !strings.Contains(err.Error(), "does not contain valid certificate data") {
		t.Errorf("Expected invalid certificate error, got: %v", err)
	}
}

func TestSelfSignedAcceptStrategy(t *testing.T) {
	mockStore := NewMockTrustStore()
	strategy := NewSelfSignedAcceptStrategy(mockStore)

	if strategy.Name() != "self-signed-accept" {
		t.Errorf("Expected name 'self-signed-accept', got '%s'", strategy.Name())
	}

	if strategy.Priority() != 10 {
		t.Errorf("Expected priority 10, got %d", strategy.Priority())
	}

	// Test execution
	ctx := context.Background()
	err := strategy.Execute(ctx, "test.registry")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify SetInsecure was called
	if len(mockStore.setInsecureCalls) != 1 {
		t.Errorf("Expected 1 SetInsecure call, got %d", len(mockStore.setInsecureCalls))
	}

	if mockStore.setInsecureCalls[0].Registry != "test.registry" {
		t.Errorf("Expected registry 'test.registry', got '%s'", mockStore.setInsecureCalls[0].Registry)
	}

	if !mockStore.setInsecureCalls[0].Insecure {
		t.Error("Expected Insecure to be true")
	}

	// Test ShouldRetry
	if strategy.ShouldRetry(nil) {
		t.Error("SelfSignedAcceptStrategy should not retry")
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "registry.example.com:5000",
			expected: "registry.example.com_5000",
		},
		{
			input:    "registry.example.com/path",
			expected: "registry.example.com_path",
		},
		{
			input:    "registry with spaces",
			expected: "registry_with_spaces",
		},
		{
			input:    "registry<>:*/|?\"",
			expected: "registry________",
		},
		{
			input:    "normal.registry",
			expected: "normal.registry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestValidateCertificateData(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty data",
			data:        []byte{},
			expectError: true,
			errorMsg:    "certificate data is empty",
		},
		{
			name:        "valid PEM certificate",
			data:        []byte("-----BEGIN CERTIFICATE-----\ndata\n-----END CERTIFICATE-----"),
			expectError: false,
		},
		{
			name:        "missing begin marker",
			data:        []byte("data\n-----END CERTIFICATE-----"),
			expectError: true,
			errorMsg:    "certificate data is not in PEM format",
		},
		{
			name:        "missing end marker",
			data:        []byte("-----BEGIN CERTIFICATE-----\ndata"),
			expectError: true,
			errorMsg:    "certificate data is incomplete - missing end marker",
		},
		{
			name:        "not PEM format",
			data:        []byte("just some random data"),
			expectError: true,
			errorMsg:    "certificate data is not in PEM format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCertificateData(tt.data)

			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %v", err)
				}
			}
		})
	}
}

func TestCreateCacheEntry(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	certData := []byte("-----BEGIN CERTIFICATE-----\ntest data\n-----END CERTIFICATE-----")

	err := createCacheEntry(cacheDir, "test.registry", certData)
	if err != nil {
		t.Errorf("Expected no error creating cache entry, got: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Error("Expected cache directory to be created")
	}

	// Verify file was created with correct content
	expectedFile := filepath.Join(cacheDir, "test.registry.pem")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Error("Expected cache file to be created")
	}

	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Errorf("Failed to read cache file: %v", err)
	}

	if string(content) != string(certData) {
		t.Errorf("Expected cache file content to match cert data")
	}
}

func TestNewCachedCertStrategy(t *testing.T) {
	mockStore := NewMockTrustStore()
	strategy := NewCachedCertStrategy(mockStore)

	if strategy == nil {
		t.Fatal("Expected strategy to be created")
	}

	if strategy.trustStore != mockStore {
		t.Error("Expected trust store to be set")
	}

	if strategy.priority != 2 {
		t.Errorf("Expected priority 2, got %d", strategy.priority)
	}

	// Verify cache directory is set to a reasonable default
	if strategy.cacheDir == "" {
		t.Error("Expected cache directory to be set")
	}

	if !strings.Contains(strategy.cacheDir, ".idpbuilder") {
		t.Error("Expected cache directory to contain .idpbuilder")
	}

	if !strings.Contains(strategy.cacheDir, "cert-cache") {
		t.Error("Expected cache directory to contain cert-cache")
	}
}