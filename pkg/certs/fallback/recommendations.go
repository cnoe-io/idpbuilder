package fallback

import (
	"fmt"
	"regexp"
	"strings"
)

// RecommendationEngine generates actionable recommendations for certificate failures
type RecommendationEngine struct {
	knownIssues map[string]*KnownIssue
}

// KnownIssue represents a common certificate problem with solutions
type KnownIssue struct {
	Pattern        *regexp.Regexp
	Category       IssueCategory
	Severity       IssueSeverity
	Description    string
	Recommendations []string
	LearnMoreURL   string
}

// IssueCategory categorizes different types of certificate issues
type IssueCategory int

const (
	CategoryUnknownCA IssueCategory = iota
	CategoryExpiredCert
	CategoryHostnameMismatch
	CategorySelfSigned
	CategoryTLSVersion
	CategoryNetworkIssue
	CategoryConfiguration
)

// IssueSeverity indicates how critical the issue is
type IssueSeverity int

const (
	SeverityInfo IssueSeverity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

// Recommendation represents a single actionable recommendation
type Recommendation struct {
	Title       string
	Description string
	Action      string
	Command     string
	RiskLevel   SecurityRiskLevel
	Urgency     RecommendationUrgency
}

// RecommendationUrgency indicates how urgently a recommendation should be followed
type RecommendationUrgency int

const (
	UrgencyLow RecommendationUrgency = iota
	UrgencyMedium
	UrgencyHigh
	UrgencyImmediate
)

// NewRecommendationEngine creates a new recommendation engine with predefined issues
func NewRecommendationEngine() *RecommendationEngine {
	engine := &RecommendationEngine{
		knownIssues: make(map[string]*KnownIssue),
	}

	// Initialize known issues and their solutions
	engine.initializeKnownIssues()

	return engine
}

// GenerateRecommendations analyzes a certificate error and generates actionable recommendations
func (r *RecommendationEngine) GenerateRecommendations(registry string, err error) []string {
	var recommendations []string

	// Get the error message for analysis
	errorMsg := strings.ToLower(err.Error())

	// Check for known issues
	for _, issue := range r.knownIssues {
		if issue.Pattern.MatchString(errorMsg) {
			recommendations = append(recommendations, issue.Recommendations...)
		}
	}

	// Add registry-specific recommendations
	registryRecs := r.generateRegistrySpecificRecommendations(registry, errorMsg)
	recommendations = append(recommendations, registryRecs...)

	// Add general fallback recommendations
	generalRecs := r.generateGeneralRecommendations(errorMsg)
	recommendations = append(recommendations, generalRecs...)

	// Deduplicate and prioritize recommendations
	return r.prioritizeRecommendations(recommendations)
}

// initializeKnownIssues sets up the database of known certificate issues
func (r *RecommendationEngine) initializeKnownIssues() {
	// Unknown Certificate Authority
	r.knownIssues["unknown_authority"] = &KnownIssue{
		Pattern:     regexp.MustCompile(`unknown authority|certificate signed by unknown authority`),
		Category:    CategoryUnknownCA,
		Severity:    SeverityError,
		Description: "Certificate signed by unknown or untrusted Certificate Authority",
		Recommendations: []string{
			"Add the registry's CA certificate to your trust store",
			"Extract the CA certificate from the Kind cluster if using self-signed certs",
			"Verify the registry is using a valid certificate chain",
			"Use --insecure flag only for testing (NOT recommended for production)",
		},
		LearnMoreURL: "https://docs.docker.com/registry/insecure/",
	}

	// Expired Certificate
	r.knownIssues["expired_cert"] = &KnownIssue{
		Pattern:     regexp.MustCompile(`certificate has expired|expired certificate`),
		Category:    CategoryExpiredCert,
		Severity:    SeverityCritical,
		Description: "Registry certificate has expired",
		Recommendations: []string{
			"Contact registry administrator to renew the certificate",
			"Check if registry has updated certificates available",
			"Verify system clock is correct",
			"Use --insecure flag as temporary workaround (SECURITY RISK)",
		},
		LearnMoreURL: "https://letsencrypt.org/docs/certificate-renewals/",
	}

	// Hostname Mismatch
	r.knownIssues["hostname_mismatch"] = &KnownIssue{
		Pattern:     regexp.MustCompile(`hostname.*doesn't match|certificate is valid for.*but not`),
		Category:    CategoryHostnameMismatch,
		Severity:    SeverityError,
		Description: "Certificate hostname doesn't match the registry URL",
		Recommendations: []string{
			"Verify you're using the correct registry hostname",
			"Check if registry is accessible via Subject Alternative Names (SANs)",
			"Update /etc/hosts if using custom hostname mapping",
			"Configure registry with correct hostname in certificate",
		},
		LearnMoreURL: "https://tools.ietf.org/html/rfc6125",
	}

	// Self-Signed Certificate
	r.knownIssues["self_signed"] = &KnownIssue{
		Pattern:     regexp.MustCompile(`self.*signed|self-signed certificate`),
		Category:    CategorySelfSigned,
		Severity:    SeverityWarning,
		Description: "Registry is using self-signed certificate",
		Recommendations: []string{
			"Add the self-signed certificate to your trust store",
			"Extract certificate from Kind cluster: kubectl cp <pod>:/etc/ssl/certs/ca.pem ./ca.pem",
			"Configure registry with proper CA-signed certificate for production",
			"Use --insecure flag for development only",
		},
		LearnMoreURL: "https://kind.sigs.k8s.io/docs/user/private-registries/",
	}

	// TLS Version Issues
	r.knownIssues["tls_version"] = &KnownIssue{
		Pattern:     regexp.MustCompile(`tls.*version|protocol version|handshake failure`),
		Category:    CategoryTLSVersion,
		Severity:    SeverityError,
		Description: "TLS version compatibility issue",
		Recommendations: []string{
			"Update registry to support TLS 1.2 or higher",
			"Check if client TLS configuration is too restrictive",
			"Verify cipher suite compatibility",
			"Update container registry client libraries",
		},
		LearnMoreURL: "https://wiki.mozilla.org/Security/Server_Side_TLS",
	}

	// Network/Connection Issues
	r.knownIssues["network_issue"] = &KnownIssue{
		Pattern:     regexp.MustCompile(`connection.*refused|timeout|network.*unreachable`),
		Category:    CategoryNetworkIssue,
		Severity:    SeverityError,
		Description: "Network connectivity issue to registry",
		Recommendations: []string{
			"Verify registry is running and accessible",
			"Check firewall and network policies",
			"Confirm registry port (usually 443 for HTTPS, 5000 for HTTP)",
			"Test connectivity: curl -v https://registry-url/v2/",
		},
		LearnMoreURL: "https://docs.docker.com/registry/spec/api/",
	}
}

// generateRegistrySpecificRecommendations provides recommendations based on registry type
func (r *RecommendationEngine) generateRegistrySpecificRecommendations(registry string, errorMsg string) []string {
	var recommendations []string

	// Kind cluster registry recommendations
	if strings.Contains(registry, "kind") || strings.Contains(registry, "localhost") {
		recommendations = append(recommendations, []string{
			"For Kind clusters: Extract CA certificate using: kind get kubeconfig",
			"Verify Kind cluster is running: kind get clusters",
			"Check Gitea pod is accessible: kubectl get pods -n gitea",
			"Use Kind's built-in registry if available: localhost:5001",
		}...)
	}

	// Docker Hub recommendations
	if strings.Contains(registry, "docker.io") || strings.Contains(registry, "registry-1.docker.io") {
		recommendations = append(recommendations, []string{
			"Docker Hub connectivity issue - check internet connection",
			"Verify Docker Hub status: https://status.docker.com/",
			"Try alternative Docker Hub mirrors if available",
			"Check for rate limiting issues",
		}...)
	}

	// Self-hosted registry recommendations
	if strings.Contains(registry, "harbor") || strings.Contains(registry, "nexus") {
		recommendations = append(recommendations, []string{
			"Contact registry administrator for certificate issues",
			"Check registry documentation for TLS configuration",
			"Verify registry health endpoint is accessible",
			"Review registry logs for additional error details",
		}...)
	}

	return recommendations
}

// generateGeneralRecommendations provides general fallback and troubleshooting advice
func (r *RecommendationEngine) generateGeneralRecommendations(errorMsg string) []string {
	var recommendations []string

	// Time-based recommendations
	if strings.Contains(errorMsg, "expired") || strings.Contains(errorMsg, "not yet valid") {
		recommendations = append(recommendations, []string{
			"Verify system clock is correct: date",
			"Sync system time: sudo ntpdate -s time.nist.gov",
			"Check timezone settings",
		}...)
	}

	// General troubleshooting
	recommendations = append(recommendations, []string{
		"Enable verbose logging for more details: export GODEBUG=x509verifier=1",
		"Test with curl: curl -v https://registry-url/v2/",
		"Check registry status and availability",
		"Review registry documentation for TLS requirements",
	}...)

	// Security best practices
	recommendations = append(recommendations, []string{
		"NEVER use --insecure in production environments",
		"Always verify certificate fingerprints when adding to trust store",
		"Consider using a proper CA-signed certificate for the registry",
		"Monitor certificate expiration dates",
	}...)

	return recommendations
}

// prioritizeRecommendations removes duplicates and orders recommendations by importance
func (r *RecommendationEngine) prioritizeRecommendations(recommendations []string) []string {
	// Remove duplicates
	seen := make(map[string]bool)
	var unique []string

	for _, rec := range recommendations {
		if !seen[rec] {
			seen[rec] = true
			unique = append(unique, rec)
		}
	}

	// Prioritize by type (security first, then actionable steps)
	var prioritized []string

	// High priority: Security warnings
	for _, rec := range unique {
		if strings.Contains(strings.ToLower(rec), "never") ||
			strings.Contains(strings.ToLower(rec), "security") ||
			strings.Contains(strings.ToLower(rec), "production") {
			prioritized = append(prioritized, rec)
		}
	}

	// Medium priority: Direct actions
	for _, rec := range unique {
		if strings.Contains(rec, ":") || // Commands with colons
			strings.Contains(rec, "kubectl") ||
			strings.Contains(rec, "curl") ||
			strings.Contains(rec, "extract") {
			if !contains(prioritized, rec) {
				prioritized = append(prioritized, rec)
			}
		}
	}

	// Low priority: General advice
	for _, rec := range unique {
		if !contains(prioritized, rec) {
			prioritized = append(prioritized, rec)
		}
	}

	return prioritized
}

// GetDetailedRecommendation returns detailed information about a specific error
func (r *RecommendationEngine) GetDetailedRecommendation(registry string, err error) *DetailedRecommendation {
	errorMsg := strings.ToLower(err.Error())

	// Find the most relevant known issue
	var bestMatch *KnownIssue
	for _, issue := range r.knownIssues {
		if issue.Pattern.MatchString(errorMsg) {
			bestMatch = issue
			break
		}
	}

	if bestMatch == nil {
		// Create generic recommendation for unknown issues
		return &DetailedRecommendation{
			Issue:       "Unknown certificate error",
			Category:    CategoryConfiguration,
			Severity:    SeverityError,
			Description: fmt.Sprintf("Certificate validation failed: %v", err),
			ImpactAnalysis: "Unable to establish secure connection to registry. " +
				"This prevents pulling/pushing container images securely.",
			Solutions:     r.GenerateRecommendations(registry, err),
			NextSteps:     r.getNextSteps(registry, errorMsg),
			EstimatedTime: "5-30 minutes depending on issue complexity",
		}
	}

	return &DetailedRecommendation{
		Issue:          bestMatch.Description,
		Category:       bestMatch.Category,
		Severity:       bestMatch.Severity,
		Description:    fmt.Sprintf("%s: %v", bestMatch.Description, err),
		ImpactAnalysis: r.getImpactAnalysis(bestMatch.Category),
		Solutions:      bestMatch.Recommendations,
		NextSteps:      r.getNextSteps(registry, errorMsg),
		LearnMoreURL:   bestMatch.LearnMoreURL,
		EstimatedTime:  r.getEstimatedTime(bestMatch.Category),
	}
}

// DetailedRecommendation provides comprehensive guidance for certificate issues
type DetailedRecommendation struct {
	Issue          string
	Category       IssueCategory
	Severity       IssueSeverity
	Description    string
	ImpactAnalysis string
	Solutions      []string
	NextSteps      []string
	LearnMoreURL   string
	EstimatedTime  string
}

// getImpactAnalysis provides analysis of the issue's impact
func (r *RecommendationEngine) getImpactAnalysis(category IssueCategory) string {
	switch category {
	case CategoryUnknownCA:
		return "Registry connection will fail until CA certificate is trusted. All image operations blocked."
	case CategoryExpiredCert:
		return "Critical security issue. Connection failures until certificate is renewed."
	case CategoryHostnameMismatch:
		return "Certificate validation fails due to hostname mismatch. Connection blocked for security."
	case CategorySelfSigned:
		return "Self-signed certificates need explicit trust. Development workflows affected."
	case CategoryTLSVersion:
		return "TLS compatibility issue preventing secure handshake. All HTTPS operations fail."
	case CategoryNetworkIssue:
		return "Network connectivity problem. Registry may be unreachable or misconfigured."
	default:
		return "Certificate validation failure affecting registry connectivity."
	}
}

// getNextSteps provides immediate next steps for the user
func (r *RecommendationEngine) getNextSteps(registry string, errorMsg string) []string {
	var steps []string

	// Immediate diagnostics
	steps = append(steps, "1. Test basic connectivity: ping "+registry)
	steps = append(steps, "2. Check registry status: curl -I https://"+registry+"/v2/")
	steps = append(steps, "3. Examine certificate: openssl s_client -connect "+registry+":443")

	// Issue-specific steps
	if strings.Contains(errorMsg, "unknown authority") {
		steps = append(steps, "4. Extract and trust the CA certificate")
		steps = append(steps, "5. Verify certificate chain is complete")
	} else if strings.Contains(errorMsg, "expired") {
		steps = append(steps, "4. Verify system time is correct")
		steps = append(steps, "5. Contact registry admin for certificate renewal")
	} else if strings.Contains(errorMsg, "hostname") {
		steps = append(steps, "4. Verify registry hostname matches certificate")
		steps = append(steps, "5. Check Subject Alternative Names (SANs)")
	}

	return steps
}

// getEstimatedTime provides time estimates for resolution
func (r *RecommendationEngine) getEstimatedTime(category IssueCategory) string {
	switch category {
	case CategoryUnknownCA:
		return "10-20 minutes (extract and install CA certificate)"
	case CategoryExpiredCert:
		return "Depends on admin response (certificate renewal required)"
	case CategoryHostnameMismatch:
		return "5-15 minutes (verify hostname configuration)"
	case CategorySelfSigned:
		return "5-10 minutes (add certificate to trust store)"
	case CategoryTLSVersion:
		return "15-30 minutes (update TLS configuration)"
	case CategoryNetworkIssue:
		return "5-60 minutes (network troubleshooting)"
	default:
		return "10-30 minutes (depending on issue complexity)"
	}
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}