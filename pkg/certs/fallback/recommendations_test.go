package fallback

import (
	"errors"
	"strings"
	"testing"
)

func TestNewRecommendationEngine(t *testing.T) {
	engine := NewRecommendationEngine()

	if engine == nil {
		t.Fatal("Expected non-nil recommendation engine")
	}

	if len(engine.knownIssues) == 0 {
		t.Error("Expected known issues to be initialized")
	}
}

func TestGenerateRecommendations(t *testing.T) {
	engine := NewRecommendationEngine()

	tests := []struct {
		name        string
		registry    string
		err         error
		expectCount int
	}{
		{
			name:        "Unknown authority error",
			registry:    "test.registry.com",
			err:         errors.New("certificate signed by unknown authority"),
			expectCount: 1, // Should match at least one known issue
		},
		{
			name:        "Expired certificate",
			registry:    "expired.registry.com",
			err:         errors.New("certificate has expired"),
			expectCount: 1,
		},
		{
			name:        "Hostname mismatch",
			registry:    "mismatch.registry.com",
			err:         errors.New("hostname doesn't match certificate"),
			expectCount: 1,
		},
		{
			name:        "Self-signed certificate",
			registry:    "selfsigned.registry.com",
			err:         errors.New("self-signed certificate"),
			expectCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendations := engine.GenerateRecommendations(tt.registry, tt.err)

			if len(recommendations) == 0 {
				t.Error("Expected at least one recommendation")
				return
			}

			// Check that recommendations are meaningful
			for _, rec := range recommendations {
				if len(rec) == 0 {
					t.Error("Found empty recommendation")
				}
			}
		})
	}
}

func TestRegistrySpecificRecommendations(t *testing.T) {
	engine := NewRecommendationEngine()

	tests := []struct {
		name     string
		registry string
		expected string
	}{
		{
			name:     "Kind registry",
			registry: "kind-registry:5000",
			expected: "Kind",
		},
		{
			name:     "Docker Hub",
			registry: "docker.io",
			expected: "Docker Hub",
		},
		{
			name:     "Harbor registry",
			registry: "harbor.company.com",
			expected: "registry administrator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorMsg := "unknown authority"
			recommendations := engine.generateRegistrySpecificRecommendations(tt.registry, errorMsg)

			found := false
			for _, rec := range recommendations {
				if strings.Contains(rec, tt.expected) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected to find %q in recommendations for %s", tt.expected, tt.registry)
			}
		})
	}
}

func TestPrioritizeRecommendations(t *testing.T) {
	engine := NewRecommendationEngine()

	input := []string{
		"NEVER use --insecure in production environments",
		"Test with curl: curl -v https://registry-url/v2/",
		"Add the registry's CA certificate to your trust store",
		"NEVER use --insecure in production environments", // Duplicate
		"Check registry status and availability",
	}

	result := engine.prioritizeRecommendations(input)

	// Check duplicates removed
	if len(result) != 4 {
		t.Errorf("Expected 4 unique recommendations, got %d", len(result))
	}

	// Check security warning comes first
	if !strings.Contains(result[0], "NEVER") {
		t.Error("Expected security warning to be prioritized first")
	}

	// Check that all original items are present (deduplicated)
	expectedItems := []string{
		"NEVER use --insecure in production environments",
		"Test with curl: curl -v https://registry-url/v2/",
		"Add the registry's CA certificate to your trust store",
		"Check registry status and availability",
	}

	for _, expected := range expectedItems {
		found := false
		for _, actual := range result {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected item %q not found in result", expected)
		}
	}
}

func TestGetDetailedRecommendation(t *testing.T) {
	engine := NewRecommendationEngine()

	tests := []struct {
		name     string
		registry string
		err      error
		checkFn  func(*DetailedRecommendation) error
	}{
		{
			name:     "Unknown authority",
			registry: "test.registry.com",
			err:      errors.New("certificate signed by unknown authority"),
			checkFn: func(rec *DetailedRecommendation) error {
				if rec.Category != CategoryUnknownCA {
					return errors.New("expected CategoryUnknownCA")
				}
				if rec.Severity != SeverityError {
					return errors.New("expected SeverityError")
				}
				return nil
			},
		},
		{
			name:     "Expired certificate",
			registry: "expired.com",
			err:      errors.New("certificate has expired"),
			checkFn: func(rec *DetailedRecommendation) error {
				if rec.Category != CategoryExpiredCert {
					return errors.New("expected CategoryExpiredCert")
				}
				if rec.Severity != SeverityCritical {
					return errors.New("expected SeverityCritical")
				}
				return nil
			},
		},
		{
			name:     "Unknown error",
			registry: "unknown.com",
			err:      errors.New("some unknown certificate error"),
			checkFn: func(rec *DetailedRecommendation) error {
				if rec.Category != CategoryConfiguration {
					return errors.New("expected CategoryConfiguration for unknown error")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendation := engine.GetDetailedRecommendation(tt.registry, tt.err)

			if recommendation == nil {
				t.Fatal("Expected non-nil detailed recommendation")
			}

			if len(recommendation.Solutions) == 0 {
				t.Error("Expected at least one solution")
			}

			if len(recommendation.NextSteps) == 0 {
				t.Error("Expected at least one next step")
			}

			if tt.checkFn != nil {
				if err := tt.checkFn(recommendation); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestImpactAnalysis(t *testing.T) {
	engine := NewRecommendationEngine()

	tests := []struct {
		category IssueCategory
		expected string
	}{
		{CategoryUnknownCA, "Registry connection will fail"},
		{CategoryExpiredCert, "Critical security issue"},
		{CategoryHostnameMismatch, "Certificate validation fails"},
		{CategorySelfSigned, "Self-signed certificates need explicit trust"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.category)), func(t *testing.T) {
			impact := engine.getImpactAnalysis(tt.category)

			if !strings.Contains(impact, tt.expected) {
				t.Errorf("Expected impact to contain %q, got %q", tt.expected, impact)
			}
		})
	}
}

func TestNextSteps(t *testing.T) {
	engine := NewRecommendationEngine()

	registry := "test.registry.com"
	errorMsg := "unknown authority"

	steps := engine.getNextSteps(registry, errorMsg)

	if len(steps) == 0 {
		t.Error("Expected at least one next step")
	}

	// Check that basic connectivity tests are included
	foundPing := false
	foundCurl := false

	for _, step := range steps {
		if strings.Contains(step, "ping") {
			foundPing = true
		}
		if strings.Contains(step, "curl") {
			foundCurl = true
		}
	}

	if !foundPing {
		t.Error("Expected ping test in next steps")
	}

	if !foundCurl {
		t.Error("Expected curl test in next steps")
	}
}

func TestEstimatedTime(t *testing.T) {
	engine := NewRecommendationEngine()

	tests := []struct {
		category IssueCategory
		expected string
	}{
		{CategoryUnknownCA, "10-20 minutes"},
		{CategoryExpiredCert, "Depends on admin"},
		{CategoryHostnameMismatch, "5-15 minutes"},
		{CategorySelfSigned, "5-10 minutes"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.category)), func(t *testing.T) {
			time := engine.getEstimatedTime(tt.category)

			if !strings.Contains(time, tt.expected) {
				t.Errorf("Expected time estimate to contain %q, got %q", tt.expected, time)
			}
		})
	}
}

func TestContainsHelper(t *testing.T) {
	slice := []string{"apple", "banana", "orange"}

	if !contains(slice, "banana") {
		t.Error("Expected contains to return true for 'banana'")
	}

	if contains(slice, "grape") {
		t.Error("Expected contains to return false for 'grape'")
	}

	if contains([]string{}, "anything") {
		t.Error("Expected contains to return false for empty slice")
	}
}