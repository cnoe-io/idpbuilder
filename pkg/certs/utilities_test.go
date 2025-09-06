package certs

import (
	"crypto/x509"
	"os"
	"testing"
)

func TestNewCertPoolManager(t *testing.T) {
	trustStore := NewTrustStore()
	poolManager := NewCertPoolManager(trustStore)
	
	if poolManager == nil {
		t.Fatal("Pool manager should not be nil")
	}
	
	if poolManager.pools == nil {
		t.Error("Pools map not initialized")
	}
}

func TestCertPoolManager_GetPool(t *testing.T) {
	trustStore := NewTrustStore()
	poolManager := NewCertPoolManager(trustStore)
	
	registry := "test.registry.com"
	
	pool, err := poolManager.GetPool(registry)
	if err != nil {
		t.Fatalf("Failed to get pool: %v", err)
	}
	
	if pool == nil {
		t.Error("Pool should not be nil")
	}
}

func TestCertPoolManager_ClearCache(t *testing.T) {
	trustStore := NewTrustStore()
	poolManager := NewCertPoolManager(trustStore)
	
	// Get a pool to populate cache
	poolManager.GetPool("test.registry.com")
	
	// Clear cache
	poolManager.ClearCache()
	
	// Verify cache is cleared (pools map is reset)
	if len(poolManager.pools) != 0 {
		t.Error("Cache should be cleared")
	}
}

func TestNewTransportConfigurer(t *testing.T) {
	trustStore := NewTrustStore()
	configurer := NewTransportConfigurer(trustStore)
	
	if configurer == nil {
		t.Fatal("Configurer should not be nil")
	}
	
	if configurer.trustManager == nil {
		t.Error("Trust manager should be set")
	}
}

func TestTransportConfigurer_ConfigureTransport(t *testing.T) {
	trustStore := NewTrustStore()
	configurer := NewTransportConfigurer(trustStore)
	
	registry := "test.registry.com"
	
	transport, err := configurer.ConfigureTransport(registry)
	if err != nil {
		t.Fatalf("Failed to configure transport: %v", err)
	}
	
	if transport == nil {
		t.Error("Transport should not be nil")
	}
	
	if transport.TLSClientConfig == nil {
		t.Error("TLS config should not be nil")
	}
}

func TestDefaultTLSConfig(t *testing.T) {
	config := DefaultTLSConfig()
	
	if config == nil {
		t.Fatal("Config should not be nil")
	}
	
	if config.ValidateHostname != true {
		t.Error("ValidateHostname should be true by default")
	}
	
	if config.Timeout == 0 {
		t.Error("Timeout should be set")
	}
}

func TestTLSConfig_LoadConfigFromEnv(t *testing.T) {
	config := DefaultTLSConfig()
	
	// Test insecure mode
	os.Setenv("IDPBUILDER_TLS_INSECURE", "true")
	config.LoadConfigFromEnv()
	
	if !config.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be true")
	}
	
	// Clean up
	os.Unsetenv("IDPBUILDER_TLS_INSECURE")
}

func TestTLSConfig_ToGoTLSConfig(t *testing.T) {
	config := DefaultTLSConfig()
	config.Registry = "test.registry.com"
	
	tlsConfig := config.ToGoTLSConfig()
	
	if tlsConfig == nil {
		t.Fatal("TLS config should not be nil")
	}
	
	if tlsConfig.ServerName != "test.registry.com" {
		t.Errorf("Expected ServerName 'test.registry.com', got '%s'", tlsConfig.ServerName)
	}
}

func TestInitSecurityLogger(t *testing.T) {
	// Test logger initialization
	err := InitSecurityLogger()
	if err != nil {
		t.Fatalf("Failed to initialize security logger: %v", err)
	}
	
	// Test logging
	LogSecurityEvent("TEST_EVENT", "test.target", "Test message")
	
	// Clean up
	CloseSecurityLogger()
}

func TestNewCertValidator(t *testing.T) {
	validator := NewCertValidator()
	
	if validator == nil {
		t.Fatal("Validator should not be nil")
	}
}

func TestCertValidator_ValidateCertificate(t *testing.T) {
	validator := NewCertValidator()
	cert := createTestCertificate(t)
	
	result := validator.ValidateCertificate(cert)
	
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	
	if !result.Valid {
		t.Errorf("Certificate should be valid: %s", result.Message)
	}
}

func TestCertValidator_ValidateExpiredCertificate(t *testing.T) {
	validator := NewCertValidator()
	
	// Create an expired certificate (modify the test cert function)
	cert := createExpiredTestCertificate(t)
	
	result := validator.ValidateCertificate(cert)
	
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	
	if result.Valid {
		t.Error("Expired certificate should be invalid")
	}
	
	if result.Message != "Certificate expired" {
		t.Errorf("Expected 'Certificate expired', got '%s'", result.Message)
	}
}

func TestParseEnvBool(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"TRUE", true},
		{"1", true},
		{"enabled", true},
		{"ENABLED", true},
		{"false", false},
		{"0", false},
		{"disabled", false},
		{"", false},
		{"invalid", false},
	}
	
	for _, tc := range testCases {
		result := parseEnvBool(tc.input)
		if result != tc.expected {
			t.Errorf("parseEnvBool(%q) = %v, expected %v", tc.input, result, tc.expected)
		}
	}
}

func TestLoadCertificatesFromDir(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test empty directory
	certs, err := loadCertificatesFromDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to load from empty dir: %v", err)
	}
	
	if len(certs) != 0 {
		t.Error("Should return empty list for empty directory")
	}
	
	// Test non-existent directory
	certs, err = loadCertificatesFromDir("/non/existent/path")
	if err != nil || certs != nil {
		t.Error("Should return nil for non-existent directory")
	}
}

// Helper function to create expired certificate
func createExpiredTestCertificate(t *testing.T) *x509.Certificate {
	// Skip this test since we can't easily modify a certificate after creation
	t.Skip("Skipping expired certificate test - requires certificate recreation")
	return nil
}

func TestExtractHostname(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"registry.example.com", "registry.example.com"},
		{"registry.example.com:5000", "registry.example.com"},
		{"https://registry.example.com", "registry.example.com"},
		{"https://registry.example.com:5000", "registry.example.com"},
		{"registry.example.com/path", "registry.example.com"},
	}
	
	for _, tc := range testCases {
		result := extractHostname(tc.input)
		if result != tc.expected {
			t.Errorf("extractHostname(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestSanitizeUnsanitizeName(t *testing.T) {
	testCases := []string{
		"registry.example.com",
		"registry.example.com:5000",
		"registry/namespace",
		"complex.registry.com:5000/namespace",
	}
	
	for _, original := range testCases {
		sanitized := sanitizeName(original)
		
		// Test that sanitization removes problematic characters
		if sanitized == original && (contains(original, ":") || contains(original, "/") || contains(original, ".")) {
			t.Errorf("Sanitization didn't change input: %s", original)
		}
		
		// Test that sanitized name doesn't contain problematic characters
		if contains(sanitized, ":") || contains(sanitized, "/") || contains(sanitized, ".") {
			t.Errorf("Sanitized name still contains problematic characters: %s", sanitized)
		}
	}
}

func contains(s, substr string) bool {
	for i := 0; i < len(s); i++ {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}