package fallback

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestCertErrorType tests the CertErrorType enum
func TestCertErrorType(t *testing.T) {
	tests := []struct {
		name     string
		errType  CertErrorType
		expected string
	}{
		{"Expired certificate", CertExpired, "expired"},
		{"Not yet valid certificate", CertNotYetValid, "not_yet_valid"},
		{"Hostname mismatch", CertHostnameMismatch, "hostname_mismatch"},
		{"Untrusted root", CertUntrustedRoot, "untrusted_root"},
		{"Self-signed certificate", CertSelfSigned, "self_signed"},
		{"Revoked certificate", CertRevoked, "revoked"},
		{"Invalid signature", CertInvalidSignature, "invalid_signature"},
		{"Unknown error", CertUnknownError, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errType.String(); got != tt.expected {
				t.Errorf("CertErrorType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestCertErrorType_AllTypes ensures all error types are covered
func TestCertErrorType_AllTypes(t *testing.T) {
	allTypes := []CertErrorType{
		CertExpired,
		CertNotYetValid,
		CertHostnameMismatch,
		CertUntrustedRoot,
		CertSelfSigned,
		CertRevoked,
		CertInvalidSignature,
		CertUnknownError,
	}

	for _, errType := range allTypes {
		if errType.String() == "" {
			t.Errorf("CertErrorType %v should have non-empty string representation", errType)
		}
	}
}

// TestRecommendation tests the Recommendation structure
func TestRecommendation(t *testing.T) {
	rec := &Recommendation{
		ErrorType:     CertExpired,
		Title:         "Certificate Expired",
		Description:   "The server certificate has expired and is no longer valid",
		RiskLevel:     "high",
		Actions:       []string{"Renew certificate", "Update certificate store"},
		TechnicalInfo: "Certificate expired on 2023-12-01T00:00:00Z",
		UserGuidance:  "Contact your system administrator to renew the certificate",
		AutoFix:       false,
		Severity:      3,
		Context: map[string]interface{}{
			"host":        "example.com",
			"port":        443,
			"expires_at":  "2023-12-01T00:00:00Z",
		},
	}

	t.Run("Field validation", func(t *testing.T) {
		if rec.ErrorType == 0 {
			t.Error("ErrorType should not be zero")
		}
		if rec.Title == "" {
			t.Error("Title should not be empty")
		}
		if rec.Description == "" {
			t.Error("Description should not be empty")
		}
		if rec.RiskLevel == "" {
			t.Error("RiskLevel should not be empty")
		}
		if len(rec.Actions) == 0 {
			t.Error("Actions should not be empty")
		}
		if rec.Severity < 1 || rec.Severity > 5 {
			t.Error("Severity should be between 1 and 5")
		}
	})

	t.Run("JSON serialization", func(t *testing.T) {
		data, err := json.Marshal(rec)
		if err != nil {
			t.Fatalf("Failed to marshal Recommendation: %v", err)
		}

		if !strings.Contains(string(data), "Certificate Expired") {
			t.Error("JSON should contain title")
		}
		if !strings.Contains(string(data), "high") {
			t.Error("JSON should contain risk level")
		}

		// Test deserialization
		var unmarshaled Recommendation
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal Recommendation: %v", err)
		}

		if unmarshaled.ErrorType != rec.ErrorType {
			t.Errorf("ErrorType mismatch: got %v, want %v", unmarshaled.ErrorType, rec.ErrorType)
		}
		if unmarshaled.Title != rec.Title {
			t.Errorf("Title mismatch: got %v, want %v", unmarshaled.Title, rec.Title)
		}
		if unmarshaled.Severity != rec.Severity {
			t.Errorf("Severity mismatch: got %v, want %v", unmarshaled.Severity, rec.Severity)
		}
	})
}

// TestDefaultRecommendationEngine tests the main recommendation engine
func TestDefaultRecommendationEngine(t *testing.T) {
	engine := NewDefaultRecommendationEngine()

	if engine == nil {
		t.Fatal("NewDefaultRecommendationEngine should not return nil")
	}

	context := map[string]interface{}{
		"host":     "example.com",
		"port":     443,
		"protocol": "https",
	}

	t.Run("CertExpired recommendations", func(t *testing.T) {
		rec := engine.GetRecommendation(CertExpired, context)

		if rec == nil {
			t.Fatal("GetRecommendation should not return nil")
		}
		if rec.ErrorType != CertExpired {
			t.Errorf("Expected CertExpired, got %v", rec.ErrorType)
		}
		if rec.RiskLevel != "high" {
			t.Errorf("Expired certificates should have high risk level, got %v", rec.RiskLevel)
		}
		if rec.Severity < 3 {
			t.Errorf("Expired certificates should have severity >= 3, got %v", rec.Severity)
		}
		if len(rec.Actions) == 0 {
			t.Error("Should provide actions for expired certificates")
		}
		if !strings.Contains(strings.ToLower(rec.Title), "expir") {
			t.Error("Title should mention expiration")
		}
	})

	t.Run("CertNotYetValid recommendations", func(t *testing.T) {
		rec := engine.GetRecommendation(CertNotYetValid, context)

		if rec.ErrorType != CertNotYetValid {
			t.Errorf("Expected CertNotYetValid, got %v", rec.ErrorType)
		}
		if rec.RiskLevel != "medium" {
			t.Errorf("Not yet valid certificates should have medium risk level, got %v", rec.RiskLevel)
		}
		if len(rec.Actions) == 0 {
			t.Error("Should provide actions for not yet valid certificates")
		}
		if !strings.Contains(strings.ToLower(rec.Description), "time") {
			t.Error("Description should mention timing issues")
		}
	})

	t.Run("CertHostnameMismatch recommendations", func(t *testing.T) {
		rec := engine.GetRecommendation(CertHostnameMismatch, context)

		if rec.ErrorType != CertHostnameMismatch {
			t.Errorf("Expected CertHostnameMismatch, got %v", rec.ErrorType)
		}
		if rec.RiskLevel != "medium" {
			t.Errorf("Hostname mismatch should have medium risk level, got %v", rec.RiskLevel)
		}
		if len(rec.Actions) == 0 {
			t.Error("Should provide actions for hostname mismatch")
		}
		if !strings.Contains(strings.ToLower(rec.Description), "hostname") &&
		   !strings.Contains(strings.ToLower(rec.Description), "domain") {
			t.Error("Description should mention hostname or domain")
		}
	})

	t.Run("CertUntrustedRoot recommendations", func(t *testing.T) {
		rec := engine.GetRecommendation(CertUntrustedRoot, context)

		if rec.ErrorType != CertUntrustedRoot {
			t.Errorf("Expected CertUntrustedRoot, got %v", rec.ErrorType)
		}
		if rec.RiskLevel != "high" {
			t.Errorf("Untrusted root should have high risk level, got %v", rec.RiskLevel)
		}
		if rec.Severity < 3 {
			t.Errorf("Untrusted root should have severity >= 3, got %v", rec.Severity)
		}
		if len(rec.Actions) == 0 {
			t.Error("Should provide actions for untrusted root")
		}
	})

	t.Run("CertSelfSigned recommendations", func(t *testing.T) {
		rec := engine.GetRecommendation(CertSelfSigned, context)

		if rec.ErrorType != CertSelfSigned {
			t.Errorf("Expected CertSelfSigned, got %v", rec.ErrorType)
		}
		if rec.RiskLevel != "medium" {
			t.Errorf("Self-signed certificates should have medium risk level, got %v", rec.RiskLevel)
		}
		if len(rec.Actions) == 0 {
			t.Error("Should provide actions for self-signed certificates")
		}
		if !strings.Contains(strings.ToLower(rec.Description), "self") {
			t.Error("Description should mention self-signed")
		}
	})

	t.Run("CertRevoked recommendations", func(t *testing.T) {
		rec := engine.GetRecommendation(CertRevoked, context)

		if rec.ErrorType != CertRevoked {
			t.Errorf("Expected CertRevoked, got %v", rec.ErrorType)
		}
		if rec.RiskLevel != "critical" {
			t.Errorf("Revoked certificates should have critical risk level, got %v", rec.RiskLevel)
		}
		if rec.Severity < 4 {
			t.Errorf("Revoked certificates should have severity >= 4, got %v", rec.Severity)
		}
		if len(rec.Actions) == 0 {
			t.Error("Should provide actions for revoked certificates")
		}
	})

	t.Run("CertInvalidSignature recommendations", func(t *testing.T) {
		rec := engine.GetRecommendation(CertInvalidSignature, context)

		if rec.ErrorType != CertInvalidSignature {
			t.Errorf("Expected CertInvalidSignature, got %v", rec.ErrorType)
		}
		if rec.RiskLevel != "high" {
			t.Errorf("Invalid signature should have high risk level, got %v", rec.RiskLevel)
		}
		if rec.Severity < 3 {
			t.Errorf("Invalid signature should have severity >= 3, got %v", rec.Severity)
		}
		if len(rec.Actions) == 0 {
			t.Error("Should provide actions for invalid signature")
		}
	})

	t.Run("CertUnknownError recommendations", func(t *testing.T) {
		rec := engine.GetRecommendation(CertUnknownError, context)

		if rec.ErrorType != CertUnknownError {
			t.Errorf("Expected CertUnknownError, got %v", rec.ErrorType)
		}
		if rec.RiskLevel != "medium" {
			t.Errorf("Unknown errors should have medium risk level, got %v", rec.RiskLevel)
		}
		if len(rec.Actions) == 0 {
			t.Error("Should provide actions for unknown errors")
		}
	})

	t.Run("Context-aware recommendations", func(t *testing.T) {
		// Test with development context
		devContext := map[string]interface{}{
			"host":        "localhost",
			"port":        8443,
			"environment": "development",
		}
		devRec := engine.GetRecommendation(CertSelfSigned, devContext)

		// Test with production context
		prodContext := map[string]interface{}{
			"host":        "api.example.com",
			"port":        443,
			"environment": "production",
		}
		prodRec := engine.GetRecommendation(CertSelfSigned, prodContext)

		// Development should be more permissive
		if devRec.Severity >= prodRec.Severity {
			t.Error("Development context should result in lower severity than production")
		}
	})

	t.Run("Risk level assignment", func(t *testing.T) {
		riskLevels := map[CertErrorType]string{
			CertExpired:          "high",
			CertNotYetValid:      "medium",
			CertHostnameMismatch: "medium",
			CertUntrustedRoot:    "high",
			CertSelfSigned:       "medium",
			CertRevoked:          "critical",
			CertInvalidSignature: "high",
			CertUnknownError:     "medium",
		}

		for errorType, expectedRisk := range riskLevels {
			rec := engine.GetRecommendation(errorType, context)
			if rec.RiskLevel != expectedRisk {
				t.Errorf("Error type %v should have risk level %v, got %v",
					errorType, expectedRisk, rec.RiskLevel)
			}
		}
	})

	t.Run("Action list generation", func(t *testing.T) {
		allErrorTypes := []CertErrorType{
			CertExpired, CertNotYetValid, CertHostnameMismatch,
			CertUntrustedRoot, CertSelfSigned, CertRevoked,
			CertInvalidSignature, CertUnknownError,
		}

		for _, errorType := range allErrorTypes {
			rec := engine.GetRecommendation(errorType, context)
			if len(rec.Actions) < 2 {
				t.Errorf("Error type %v should have at least 2 actions, got %d",
					errorType, len(rec.Actions))
			}

			// Check that actions are meaningful (not empty or too short)
			for i, action := range rec.Actions {
				if len(action) < 10 {
					t.Errorf("Action %d for error type %v is too short: %v",
						i, errorType, action)
				}
			}
		}
	})
}

// TestRecommendationEngineEdgeCases tests edge cases and error conditions
func TestRecommendationEngineEdgeCases(t *testing.T) {
	engine := NewDefaultRecommendationEngine()

	t.Run("Nil context", func(t *testing.T) {
		rec := engine.GetRecommendation(CertExpired, nil)
		if rec == nil {
			t.Error("Should handle nil context gracefully")
		}
		if rec.ErrorType != CertExpired {
			t.Error("Should still return correct error type with nil context")
		}
	})

	t.Run("Empty context", func(t *testing.T) {
		emptyContext := make(map[string]interface{})
		rec := engine.GetRecommendation(CertHostnameMismatch, emptyContext)
		if rec == nil {
			t.Error("Should handle empty context gracefully")
		}
		if len(rec.Actions) == 0 {
			t.Error("Should still provide actions with empty context")
		}
	})

	t.Run("Invalid error type", func(t *testing.T) {
		invalidErrorType := CertErrorType(999)
		rec := engine.GetRecommendation(invalidErrorType, map[string]interface{}{})
		
		// Should default to unknown error handling
		if rec == nil {
			t.Error("Should handle invalid error types gracefully")
		}
		if rec.ErrorType != CertUnknownError && rec.ErrorType != invalidErrorType {
			t.Error("Should handle invalid error type appropriately")
		}
	})

	t.Run("Complex context values", func(t *testing.T) {
		complexContext := map[string]interface{}{
			"nested": map[string]interface{}{
				"deep": map[string]string{
					"value": "test",
				},
			},
			"array":     []string{"item1", "item2"},
			"number":    42,
			"timestamp": time.Now(),
			"boolean":   true,
		}

		rec := engine.GetRecommendation(CertExpired, complexContext)
		if rec == nil {
			t.Error("Should handle complex context structures")
		}
	})

	t.Run("Recommendation consistency", func(t *testing.T) {
		context := map[string]interface{}{"host": "test.com", "port": 443}
		
		// Get multiple recommendations for the same error type
		rec1 := engine.GetRecommendation(CertExpired, context)
		rec2 := engine.GetRecommendation(CertExpired, context)
		
		if !reflect.DeepEqual(rec1, rec2) {
			t.Error("Recommendations should be consistent for same input")
		}
	})

	t.Run("Memory safety with large contexts", func(t *testing.T) {
		largeContext := make(map[string]interface{})
		for i := 0; i < 1000; i++ {
			largeContext[string(rune('A'+i%26))+string(rune('a'+i%26))] = i
		}
		
		rec := engine.GetRecommendation(CertUntrustedRoot, largeContext)
		if rec == nil {
			t.Error("Should handle large contexts without issues")
		}
	})
}

// BenchmarkRecommendationEngine benchmarks the recommendation engine performance
func BenchmarkRecommendationEngine(b *testing.B) {
	engine := NewDefaultRecommendationEngine()
	context := map[string]interface{}{
		"host": "example.com",
		"port": 443,
	}

	b.Run("GetRecommendation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			errorType := CertErrorType(i%8 + 1) // Cycle through error types
			_ = engine.GetRecommendation(errorType, context)
		}
	})

	b.Run("ConcurrentRecommendations", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = engine.GetRecommendation(CertExpired, context)
			}
		})
	})
}