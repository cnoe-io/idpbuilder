package fallback

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestNewFallbackHandler(t *testing.T) {
	config := DefaultFallbackConfig()
	handler, err := NewFallbackHandler(config)

	if err != nil {
		t.Fatalf("Failed to create fallback handler: %v", err)
	}

	if handler == nil {
		t.Fatal("Expected non-nil fallback handler")
	}

	if handler.config != config {
		t.Error("Config not properly set")
	}
}

func TestDefaultFallbackConfig(t *testing.T) {
	config := DefaultFallbackConfig()

	if !config.EnableAutoFallback {
		t.Error("Expected auto fallback to be enabled by default")
	}

	if config.AllowInsecureMode {
		t.Error("Expected insecure mode to be disabled by default")
	}

	if config.FallbackTimeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.FallbackTimeout)
	}
}

func TestAnalyzeError(t *testing.T) {
	handler := &FallbackHandler{
		config: DefaultFallbackConfig(),
	}

	tests := []struct {
		name          string
		err           error
		expectedFirst FallbackStrategy
	}{
		{
			name:          "Unknown authority error",
			err:           errors.New("certificate signed by unknown authority"),
			expectedFirst: StrategyRetryWithSystemCA,
		},
		{
			name:          "Hostname mismatch",
			err:           errors.New("hostname doesn't match certificate"),
			expectedFirst: StrategyRetryWithoutSNI,
		},
		{
			name:          "TLS version error",
			err:           errors.New("tls handshake failure"),
			expectedFirst: StrategyRetryWithLowerTLS,
		},
		{
			name:          "Expired certificate",
			err:           errors.New("certificate has expired"),
			expectedFirst: StrategyNone, // No strategies for expired cert when insecure disabled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategies := handler.analyzeError(tt.err)

			if len(strategies) == 0 && tt.expectedFirst != StrategyNone {
				t.Error("Expected at least one strategy")
				return
			}

			if len(strategies) > 0 && strategies[0] != tt.expectedFirst {
				t.Errorf("Expected first strategy %v, got %v", tt.expectedFirst, strategies[0])
			}
		})
	}
}

func TestFallbackStrategyString(t *testing.T) {
	tests := []struct {
		strategy FallbackStrategy
		expected string
	}{
		{StrategyRetryWithSystemCA, "Retry with system CA certificates"},
		{StrategyRetryWithoutSNI, "Retry without SNI (hostname verification)"},
		{StrategyInsecureMode, "Use insecure mode (no certificate validation)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.strategy.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, tt.strategy.String())
			}
		})
	}
}

func TestSecurityRiskLevelString(t *testing.T) {
	tests := []struct {
		risk     SecurityRiskLevel
		expected string
	}{
		{RiskNone, "No additional security risk"},
		{RiskLow, "Low security risk"},
		{RiskCritical, "Critical security risk"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.risk.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, tt.risk.String())
			}
		})
	}
}

func TestRetryWithSystemCA(t *testing.T) {
	handler := &FallbackHandler{
		config: DefaultFallbackConfig(),
	}

	ctx := context.Background()
	registry := "test.registry.com"

	// This test would require a mock transport for full testing
	// For now, test the basic structure
	result, err := handler.retryWithSystemCA(ctx, registry)

	// Expect failure since we don't have a real registry
	if err != nil {
		t.Logf("Expected error for non-existent registry: %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result even on failure")
	}

	if result.Strategy != StrategyRetryWithSystemCA {
		t.Errorf("Expected strategy %v, got %v", StrategyRetryWithSystemCA, result.Strategy)
	}
}

func TestTestConnection(t *testing.T) {
	handler := &FallbackHandler{}

	// Test with a transport that will fail
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	ctx := context.Background()
	registry := "nonexistent.registry.invalid"

	err := handler.testConnection(ctx, registry, transport)

	// Expect an error since registry doesn't exist
	if err == nil {
		t.Error("Expected error for non-existent registry")
	}
}

// Mock implementation for testing
type mockSecurityLogger struct {
	entries []string
}

func (m *mockSecurityLogger) LogCertificateFailure(registry string, err error) {
	m.entries = append(m.entries, "cert_failure:"+registry)
}

func (m *mockSecurityLogger) LogFallbackAttempt(registry string, strategy FallbackStrategy, success bool, reason string) {
	m.entries = append(m.entries, "fallback_attempt:"+registry)
}

func (m *mockSecurityLogger) LogFallbackSuccess(registry string, strategy FallbackStrategy, risk SecurityRiskLevel) {
	m.entries = append(m.entries, "fallback_success:"+registry)
}

func (m *mockSecurityLogger) LogInsecureModeUsed(registry string, confirmed bool, reason string) {
	m.entries = append(m.entries, "insecure_mode:"+registry)
}

func (m *mockSecurityLogger) LogSecurityDecision(decision, target, reason string) {
	m.entries = append(m.entries, "security_decision:"+target)
}

func TestHandleCertificateErrorWithMockLogger(t *testing.T) {
	mockLogger := &mockSecurityLogger{}

	handler := &FallbackHandler{
		securityLogger: mockLogger,
		config:         DefaultFallbackConfig(),
	}

	ctx := context.Background()
	registry := "test.registry.com"
	originalErr := errors.New("certificate signed by unknown authority")

	_, err := handler.HandleCertificateError(ctx, registry, originalErr)

	// Expect failure since we don't have real infrastructure
	if err == nil {
		t.Error("Expected error for fallback without proper setup")
	}

	// Check that logging occurred
	if len(mockLogger.entries) == 0 {
		t.Error("Expected security logging to occur")
	}

	// Check that certificate failure was logged
	found := false
	for _, entry := range mockLogger.entries {
		if entry == "cert_failure:"+registry {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected certificate failure to be logged")
	}
}