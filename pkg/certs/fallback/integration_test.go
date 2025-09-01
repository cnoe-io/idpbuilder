package fallback

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestLoggerRecommendationIntegration tests the integration between logger and recommendation systems
func TestLoggerRecommendationIntegration(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := NewDefaultSecurityLogger(&logBuffer)
	engine := NewDefaultRecommendationEngine()

	t.Run("Log recommendations for all error types", func(t *testing.T) {
		context := map[string]interface{}{
			"host":        "example.com",
			"port":        443,
			"environment": "production",
		}

		errorTypes := []CertErrorType{
			CertExpired, CertNotYetValid, CertHostnameMismatch,
			CertUntrustedRoot, CertSelfSigned, CertRevoked,
			CertInvalidSignature, CertUnknownError,
		}

		for _, errorType := range errorTypes {
			rec := engine.GetRecommendation(errorType, context)
			if rec == nil {
				t.Fatalf("No recommendation for error type %v", errorType)
			}

			// Log the recommendation with appropriate security level
			var level SecurityLevel
			switch rec.RiskLevel {
			case "critical":
				level = CriticalLevel
			case "high":
				level = ErrorLevel
			case "medium":
				level = WarnLevel
			default:
				level = InfoLevel
			}

			logger.Log(level, "cert-fallback", rec.Title, context, rec.RiskLevel, rec.UserGuidance)
		}

		// Verify all error types were logged
		logOutput := logBuffer.String()
		for _, errorType := range errorTypes {
			if !strings.Contains(logOutput, errorType.String()) {
				t.Errorf("Log output should contain reference to error type: %v", errorType)
			}
		}
	})

	t.Run("Security level consistency", func(t *testing.T) {
		testCases := []struct {
			errorType     CertErrorType
			expectedLevel SecurityLevel
		}{
			{CertRevoked, CriticalLevel},
			{CertExpired, ErrorLevel},
			{CertInvalidSignature, ErrorLevel},
			{CertUntrustedRoot, ErrorLevel},
			{CertHostnameMismatch, WarnLevel},
			{CertSelfSigned, WarnLevel},
			{CertNotYetValid, WarnLevel},
			{CertUnknownError, WarnLevel},
		}

		for _, tc := range testCases {
			rec := engine.GetRecommendation(tc.errorType, nil)
			var actualLevel SecurityLevel

			switch rec.RiskLevel {
			case "critical":
				actualLevel = CriticalLevel
			case "high":
				actualLevel = ErrorLevel
			case "medium":
				actualLevel = WarnLevel
			default:
				actualLevel = InfoLevel
			}

			if actualLevel != tc.expectedLevel {
				t.Errorf("Error type %v: expected security level %v, got %v",
					tc.errorType, tc.expectedLevel, actualLevel)
			}
		}
	})
}

// TestEndToEndScenarios tests complete certificate error handling workflows
func TestEndToEndScenarios(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := NewDefaultSecurityLogger(&logBuffer)
	engine := NewDefaultRecommendationEngine()

	t.Run("Certificate error to recommendation to log workflow", func(t *testing.T) {
		context := map[string]interface{}{
			"host":       "expired-cert.example.com",
			"port":       443,
			"expires_at": "2023-01-01T00:00:00Z",
		}

		// Step 1: Get recommendation for expired certificate
		recommendation := engine.GetRecommendation(CertExpired, context)
		if recommendation == nil {
			t.Fatal("Should get recommendation for expired certificate")
		}

		// Step 2: Log the recommendation
		logger.Log(ErrorLevel, "cert-handler", recommendation.Title, context,
			recommendation.RiskLevel, recommendation.UserGuidance)

		// Step 3: Verify complete workflow
		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, "expired") {
			t.Error("Log should mention certificate expiration")
		}
		if !strings.Contains(logOutput, "expired-cert.example.com") {
			t.Error("Log should include the problematic host")
		}

		// Parse log entry as JSON to verify structure
		lines := strings.Split(strings.TrimSpace(logOutput), "\n")
		if len(lines) == 0 {
			t.Fatal("Should have log output")
		}

		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(lines[0]), &logEntry); err != nil {
			t.Fatalf("Log entry should be valid JSON: %v", err)
		}

		if logEntry["level"] != "error" {
			t.Error("Log entry should have error level")
		}
		if logEntry["component"] != "cert-handler" {
			t.Error("Log entry should have correct component")
		}
	})

	t.Run("Multiple error handling", func(t *testing.T) {
		scenarios := []struct {
			host      string
			errorType CertErrorType
		}{
			{"self-signed.local", CertSelfSigned},
			{"hostname-mismatch.com", CertHostnameMismatch},
			{"untrusted-root.org", CertUntrustedRoot},
		}

		for _, scenario := range scenarios {
			context := map[string]interface{}{
				"host": scenario.host,
				"port": 443,
			}

			rec := engine.GetRecommendation(scenario.errorType, context)
			logger.Log(WarnLevel, "cert-validator", rec.Title, context,
				rec.RiskLevel, rec.UserGuidance)
		}

		// Verify all scenarios were logged
		logOutput := logBuffer.String()
		for _, scenario := range scenarios {
			if !strings.Contains(logOutput, scenario.host) {
				t.Errorf("Log should contain host: %s", scenario.host)
			}
		}

		// Count log entries
		lines := strings.Split(strings.TrimSpace(logOutput), "\n")
		// Filter out empty lines
		validLines := 0
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				validLines++
			}
		}

		if validLines < len(scenarios)+1 { // +1 for the previous test
			t.Errorf("Should have at least %d log entries, got %d",
				len(scenarios)+1, validLines)
		}
	})
}

// TestPerformanceIntegration tests performance of integrated components
func TestPerformanceIntegration(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := NewDefaultSecurityLogger(&logBuffer)
	engine := NewDefaultRecommendationEngine()

	t.Run("Recommendation generation speed", func(t *testing.T) {
		context := map[string]interface{}{
			"host": "performance-test.com",
			"port": 443,
		}

		start := time.Now()
		const iterations = 1000

		for i := 0; i < iterations; i++ {
			errorType := CertErrorType((i % 8) + 1) // Cycle through error types
			rec := engine.GetRecommendation(errorType, context)
			if rec == nil {
				t.Fatal("Should always get a recommendation")
			}
		}

		elapsed := time.Since(start)
		avgTime := elapsed / iterations

		// Should be fast enough for real-time use
		if avgTime > time.Millisecond {
			t.Errorf("Average recommendation generation time too slow: %v", avgTime)
		}

		t.Logf("Generated %d recommendations in %v (avg: %v)", iterations, elapsed, avgTime)
	})

	t.Run("Logger throughput", func(t *testing.T) {
		context := map[string]interface{}{"host": "throughput-test.com"}

		start := time.Now()
		const logCount = 1000

		for i := 0; i < logCount; i++ {
			logger.Log(InfoLevel, "throughput-test",
				"Performance test message", context, "low", "Monitor performance")
		}

		elapsed := time.Since(start)
		avgTime := elapsed / logCount

		// Should handle reasonable logging load
		if avgTime > time.Microsecond*100 {
			t.Errorf("Average logging time too slow: %v", avgTime)
		}

		t.Logf("Logged %d entries in %v (avg: %v)", logCount, elapsed, avgTime)

		// Verify all entries were written
		logOutput := logBuffer.String()
		lines := strings.Split(strings.TrimSpace(logOutput), "\n")
		nonEmptyLines := 0
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				nonEmptyLines++
			}
		}

		expectedTotal := logCount + 2 // +2 from previous performance test
		if nonEmptyLines < expectedTotal {
			t.Errorf("Expected at least %d log lines, got %d", expectedTotal, nonEmptyLines)
		}
	})
}