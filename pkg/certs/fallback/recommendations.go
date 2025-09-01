// Package fallback provides user-friendly certificate error recommendations
package fallback

import (
	"fmt"
	"strings"
	"time"
)

// CertErrorType represents different types of certificate errors
type CertErrorType int

const (
	// CertExpired indicates the certificate has expired
	CertExpired CertErrorType = iota
	// CertNotYetValid indicates the certificate is not yet valid
	CertNotYetValid
	// CertHostnameMismatch indicates hostname doesn't match certificate
	CertHostnameMismatch
	// CertUntrustedRoot indicates untrusted certificate authority
	CertUntrustedRoot
	// CertSelfSigned indicates self-signed certificate
	CertSelfSigned
	// CertRevoked indicates certificate has been revoked
	CertRevoked
	// CertInvalidSignature indicates invalid certificate signature
	CertInvalidSignature
	// CertUnknownError indicates an unrecognized certificate error
	CertUnknownError
)

// String returns the string representation of CertErrorType
func (c CertErrorType) String() string {
	switch c {
	case CertExpired:
		return "EXPIRED"
	case CertNotYetValid:
		return "NOT_YET_VALID"
	case CertHostnameMismatch:
		return "HOSTNAME_MISMATCH"
	case CertUntrustedRoot:
		return "UNTRUSTED_ROOT"
	case CertSelfSigned:
		return "SELF_SIGNED"
	case CertRevoked:
		return "REVOKED"
	case CertInvalidSignature:
		return "INVALID_SIGNATURE"
	default:
		return "UNKNOWN"
	}
}

// ErrorDetails contains specific information about a certificate error
type ErrorDetails struct {
	// ErrorType categorizes the type of certificate error
	ErrorType CertErrorType `json:"error_type"`
	// ErrorMessage contains the raw error message
	ErrorMessage string `json:"error_message"`
	// Registry identifies the OCI registry where the error occurred
	Registry string `json:"registry"`
	// CertificateInfo contains parsed certificate information if available
	CertificateInfo *CertificateInfo `json:"certificate_info,omitempty"`
	// Timestamp when the error was detected
	Timestamp time.Time `json:"timestamp"`
	// Severity indicates the severity of the error (1-10, 10 being most severe)
	Severity int `json:"severity"`
	// Recoverable indicates whether the error can potentially be worked around
	Recoverable bool `json:"recoverable"`
}

// CertificateInfo contains parsed certificate information
type CertificateInfo struct {
	// Subject contains the certificate subject information
	Subject string `json:"subject"`
	// Issuer contains the certificate issuer information
	Issuer string `json:"issuer"`
	// NotBefore is when the certificate becomes valid
	NotBefore time.Time `json:"not_before"`
	// NotAfter is when the certificate expires
	NotAfter time.Time `json:"not_after"`
	// DNSNames contains subject alternative names
	DNSNames []string `json:"dns_names,omitempty"`
	// SerialNumber contains the certificate serial number
	SerialNumber string `json:"serial_number,omitempty"`
	// Fingerprint contains the certificate fingerprint
	Fingerprint string `json:"fingerprint,omitempty"`
}

// Recommendation represents a user-friendly recommendation for handling certificate errors
type Recommendation struct {
	// Title provides a brief, user-friendly title for the recommendation
	Title string `json:"title"`
	// Description explains the recommendation in detail
	Description string `json:"description"`
	// Actions contains specific steps the user can take
	Actions []RecommendationAction `json:"actions"`
	// Priority indicates the urgency of this recommendation (1-10, 10 being most urgent)
	Priority int `json:"priority"`
	// Category groups similar recommendations
	Category string `json:"category"`
	// IsSecurityRisk indicates if ignoring this recommendation poses security risks
	IsSecurityRisk bool `json:"is_security_risk"`
	// EstimatedEffort indicates the expected effort to implement (LOW, MEDIUM, HIGH)
	EstimatedEffort string `json:"estimated_effort"`
	// References contains links or references to additional documentation
	References []string `json:"references,omitempty"`
}

// RecommendationAction represents a specific action a user can take
type RecommendationAction struct {
	// Type indicates the type of action (COMMAND, CONFIGURATION, MANUAL)
	Type string `json:"type"`
	// Description explains what this action does
	Description string `json:"description"`
	// Command contains the exact command to run (for COMMAND type actions)
	Command string `json:"command,omitempty"`
	// ConfigPath indicates the configuration file path (for CONFIGURATION type actions)
	ConfigPath string `json:"config_path,omitempty"`
	// ConfigContent contains the configuration content to add/modify
	ConfigContent string `json:"config_content,omitempty"`
	// Note contains additional notes or warnings for this action
	Note string `json:"note,omitempty"`
}

// RecommendationEngine provides user-friendly recommendations for certificate errors
type RecommendationEngine interface {
	// GetRecommendations returns recommendations for a certificate error
	GetRecommendations(details ErrorDetails) ([]Recommendation, error)
	// GetQuickFix returns a quick fix command for simple errors
	GetQuickFix(details ErrorDetails) (string, error)
	// GetDiagnosticInfo returns diagnostic information for troubleshooting
	GetDiagnosticInfo(registry string) (string, error)
	// GetSecurityAssessment evaluates the security risk of ignoring an error
	GetSecurityAssessment(details ErrorDetails) (SecurityAssessment, error)
}

// SecurityAssessment evaluates the security risk of ignoring a certificate error
type SecurityAssessment struct {
	// RiskLevel indicates the security risk (LOW, MEDIUM, HIGH, CRITICAL)
	RiskLevel string `json:"risk_level"`
	// RiskScore provides a numerical risk score (0-100)
	RiskScore int `json:"risk_score"`
	// Impact describes the potential impact of ignoring this error
	Impact string `json:"impact"`
	// Mitigations suggests ways to reduce the risk if the error must be ignored
	Mitigations []string `json:"mitigations"`
	// Compliance indicates which compliance frameworks are affected
	Compliance []string `json:"compliance,omitempty"`
}

// DefaultRecommendationEngine implements RecommendationEngine with comprehensive recommendations
type DefaultRecommendationEngine struct {
	// registryConfigs contains registry-specific configuration knowledge
	registryConfigs map[string]RegistryConfig
	// securityPolicies contains security policy configurations
	securityPolicies SecurityPolicies
}

// RegistryConfig contains registry-specific configuration information
type RegistryConfig struct {
	// Name is the registry identifier
	Name string
	// DocumentationURL points to registry-specific documentation
	DocumentationURL string
	// SupportsCertPinning indicates if the registry supports certificate pinning
	SupportsCertPinning bool
	// SupportsInsecureMode indicates if insecure mode is available
	SupportsInsecureMode bool
	// CommonIssues lists common certificate issues for this registry
	CommonIssues []string
	// ConfigExamples contains configuration examples
	ConfigExamples map[string]string
}

// SecurityPolicies contains security policy configurations
type SecurityPolicies struct {
	// AllowSelfSigned indicates if self-signed certificates are permitted
	AllowSelfSigned bool
	// AllowExpiredCerts indicates if expired certificates are permitted
	AllowExpiredCerts bool
	// RequireHostnameVerification indicates if hostname verification is required
	RequireHostnameVerification bool
	// MaxCertAge is the maximum allowed certificate age
	MaxCertAge time.Duration
	// TrustedCAs contains paths to trusted CA certificates
	TrustedCAs []string
}

// NewDefaultRecommendationEngine creates a new recommendation engine with default settings
func NewDefaultRecommendationEngine() *DefaultRecommendationEngine {
	engine := &DefaultRecommendationEngine{
		registryConfigs: make(map[string]RegistryConfig),
		securityPolicies: SecurityPolicies{
			AllowSelfSigned:             false,
			AllowExpiredCerts:           false,
			RequireHostnameVerification: true,
			MaxCertAge:                  365 * 24 * time.Hour, // 1 year
			TrustedCAs:                  []string{"/etc/ssl/certs", "/usr/local/share/ca-certificates"},
		},
	}

	// Initialize common registry configurations
	engine.initializeRegistryConfigs()

	return engine
}

// initializeRegistryConfigs sets up configurations for common OCI registries
func (d *DefaultRecommendationEngine) initializeRegistryConfigs() {
	// Docker Hub configuration
	d.registryConfigs["docker.io"] = RegistryConfig{
		Name:                "Docker Hub",
		DocumentationURL:    "https://docs.docker.com/docker-hub/",
		SupportsCertPinning: false,
		SupportsInsecureMode: true,
		CommonIssues: []string{
			"Rate limiting with anonymous access",
			"Authentication token expiration",
			"Regional certificate differences",
		},
		ConfigExamples: map[string]string{
			"insecure": `"insecure-registries": ["docker.io"]`,
			"mirror":   `"registry-mirrors": ["https://mirror.gcr.io"]`,
		},
	}

	// GitHub Container Registry configuration
	d.registryConfigs["ghcr.io"] = RegistryConfig{
		Name:                "GitHub Container Registry",
		DocumentationURL:    "https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry",
		SupportsCertPinning: true,
		SupportsInsecureMode: false,
		CommonIssues: []string{
			"GitHub token authentication required",
			"Package visibility permissions",
			"Organization-level restrictions",
		},
		ConfigExamples: map[string]string{
			"auth": `"auths": {"ghcr.io": {"auth": "base64-encoded-token"}}`,
		},
	}

	// Google Container Registry configuration
	d.registryConfigs["gcr.io"] = RegistryConfig{
		Name:                "Google Container Registry",
		DocumentationURL:    "https://cloud.google.com/container-registry/docs",
		SupportsCertPinning: true,
		SupportsInsecureMode: false,
		CommonIssues: []string{
			"Service account authentication",
			"IAM permissions for registry access",
			"Regional endpoint variations",
		},
		ConfigExamples: map[string]string{
			"gcp-auth": `gcloud auth configure-docker`,
		},
	}

	// Amazon Elastic Container Registry configuration
	d.registryConfigs["amazonaws.com"] = RegistryConfig{
		Name:                "Amazon ECR",
		DocumentationURL:    "https://docs.aws.amazon.com/ecr/",
		SupportsCertPinning: true,
		SupportsInsecureMode: false,
		CommonIssues: []string{
			"AWS credentials configuration",
			"ECR repository policies",
			"Cross-region access",
			"VPC endpoint configurations",
		},
		ConfigExamples: map[string]string{
			"aws-auth": `aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin`,
		},
	}

	// Azure Container Registry configuration
	d.registryConfigs["azurecr.io"] = RegistryConfig{
		Name:                "Azure Container Registry",
		DocumentationURL:    "https://docs.microsoft.com/en-us/azure/container-registry/",
		SupportsCertPinning: true,
		SupportsInsecureMode: false,
		CommonIssues: []string{
			"Azure AD authentication",
			"Service principal configuration",
			"Network security group rules",
		},
		ConfigExamples: map[string]string{
			"azure-auth": `az acr login --name myregistry`,
		},
	}
}

// GetRecommendations returns recommendations for a certificate error
func (d *DefaultRecommendationEngine) GetRecommendations(details ErrorDetails) ([]Recommendation, error) {
	var recommendations []Recommendation

	switch details.ErrorType {
	case CertExpired:
		recommendations = d.getExpiredCertRecommendations(details)
	case CertNotYetValid:
		recommendations = d.getNotYetValidRecommendations(details)
	case CertHostnameMismatch:
		recommendations = d.getHostnameMismatchRecommendations(details)
	case CertUntrustedRoot:
		recommendations = d.getUntrustedRootRecommendations(details)
	case CertSelfSigned:
		recommendations = d.getSelfSignedRecommendations(details)
	case CertRevoked:
		recommendations = d.getRevokedCertRecommendations(details)
	case CertInvalidSignature:
		recommendations = d.getInvalidSignatureRecommendations(details)
	default:
		recommendations = d.getGenericRecommendations(details)
	}

	return recommendations, nil
}

// getExpiredCertRecommendations provides recommendations for expired certificates
func (d *DefaultRecommendationEngine) getExpiredCertRecommendations(details ErrorDetails) []Recommendation {
	recommendations := []Recommendation{
		{
			Title:           "Contact Registry Administrator",
			Description:     "The certificate for this registry has expired and needs to be renewed by the registry administrator.",
			Priority:        9,
			Category:        "Certificate Management",
			IsSecurityRisk:  true,
			EstimatedEffort: "LOW",
			Actions: []RecommendationAction{
				{
					Type:        "MANUAL",
					Description: "Contact the registry administrator to renew the expired certificate",
					Note:        "Provide the certificate expiration date and error details",
				},
			},
		},
		{
			Title:           "Verify System Time",
			Description:     "Ensure your system clock is accurate, as certificate validation depends on correct time.",
			Priority:        7,
			Category:        "System Configuration",
			IsSecurityRisk:  false,
			EstimatedEffort: "LOW",
			Actions: []RecommendationAction{
				{
					Type:        "COMMAND",
					Description: "Check current system time",
					Command:     "date",
				},
				{
					Type:        "COMMAND",
					Description: "Sync system time with NTP",
					Command:     "sudo ntpdate -s time.nist.gov",
					Note:        "May require NTP package installation",
				},
			},
		},
		{
			Title:           "Temporary Insecure Access (High Risk)",
			Description:     "As a last resort, you can temporarily bypass certificate validation, but this poses significant security risks.",
			Priority:        3,
			Category:        "Workaround",
			IsSecurityRisk:  true,
			EstimatedEffort: "LOW",
			Actions: []RecommendationAction{
				{
					Type:        "CONFIGURATION",
					Description: "Add registry to insecure registries list",
					ConfigPath:  "/etc/docker/daemon.json",
					ConfigContent: fmt.Sprintf(`{
  "insecure-registries": ["%s"]
}`, details.Registry),
					Note: "WARNING: This disables all certificate validation for this registry",
				},
			},
			References: []string{
				"https://docs.docker.com/registry/insecure/",
			},
		},
	}

	return recommendations
}

// getNotYetValidRecommendations provides recommendations for not-yet-valid certificates
func (d *DefaultRecommendationEngine) getNotYetValidRecommendations(details ErrorDetails) []Recommendation {
	return []Recommendation{
		{
			Title:           "Check System Time",
			Description:     "The certificate is not yet valid according to your system time. Verify your system clock is correct.",
			Priority:        8,
			Category:        "System Configuration",
			IsSecurityRisk:  false,
			EstimatedEffort: "LOW",
			Actions: []RecommendationAction{
				{
					Type:        "COMMAND",
					Description: "Check current system time and timezone",
					Command:     "timedatectl status",
				},
				{
					Type:        "COMMAND",
					Description: "Enable NTP synchronization",
					Command:     "sudo timedatectl set-ntp true",
				},
			},
		},
	}
}

// getHostnameMismatchRecommendations provides recommendations for hostname mismatches
func (d *DefaultRecommendationEngine) getHostnameMismatchRecommendations(details ErrorDetails) []Recommendation {
	return []Recommendation{
		{
			Title:           "Verify Registry URL",
			Description:     "Ensure you're using the correct hostname for the registry. Check for typos or incorrect subdomain.",
			Priority:        8,
			Category:        "Configuration",
			IsSecurityRisk:  false,
			EstimatedEffort: "LOW",
			Actions: []RecommendationAction{
				{
					Type:        "MANUAL",
					Description: "Double-check the registry URL for typos or incorrect subdomains",
					Note:        "Common mistakes include www. prefixes or incorrect region specifiers",
				},
				{
					Type:        "COMMAND",
					Description: "Test DNS resolution for the registry",
					Command:     fmt.Sprintf("nslookup %s", details.Registry),
				},
			},
		},
		{
			Title:           "Check Certificate Subject Alternative Names",
			Description:     "The certificate may be valid for the server but not for the specific hostname you're using.",
			Priority:        6,
			Category:        "Certificate Analysis",
			IsSecurityRisk:  false,
			EstimatedEffort: "LOW",
			Actions: []RecommendationAction{
				{
					Type:        "COMMAND",
					Description: "View certificate details including SANs",
					Command:     fmt.Sprintf("openssl s_client -connect %s:443 -servername %s | openssl x509 -text -noout", details.Registry, details.Registry),
				},
			},
		},
	}
}

// getUntrustedRootRecommendations provides recommendations for untrusted root certificates
func (d *DefaultRecommendationEngine) getUntrustedRootRecommendations(details ErrorDetails) []Recommendation {
	return []Recommendation{
		{
			Title:           "Install Missing CA Certificate",
			Description:     "The certificate is signed by a Certificate Authority that is not trusted by your system.",
			Priority:        7,
			Category:        "Certificate Management",
			IsSecurityRisk:  false,
			EstimatedEffort: "MEDIUM",
			Actions: []RecommendationAction{
				{
					Type:        "COMMAND",
					Description: "Download the certificate chain",
					Command:     fmt.Sprintf("openssl s_client -connect %s:443 -showcerts", details.Registry),
				},
				{
					Type:        "MANUAL",
					Description: "Install the CA certificate to your system's trust store",
					Note:        "Specific steps vary by operating system and distribution",
				},
			},
			References: []string{
				"https://ubuntu.com/server/docs/security-trust-store",
				"https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/8/html/securing_networks/using-shared-system-certificates_securing-networks",
			},
		},
	}
}

// getSelfSignedRecommendations provides recommendations for self-signed certificates
func (d *DefaultRecommendationEngine) getSelfSignedRecommendations(details ErrorDetails) []Recommendation {
	return []Recommendation{
		{
			Title:           "Add Certificate Exception",
			Description:     "For internal registries with self-signed certificates, you can add a specific exception.",
			Priority:        6,
			Category:        "Certificate Management",
			IsSecurityRisk:  true,
			EstimatedEffort: "MEDIUM",
			Actions: []RecommendationAction{
				{
					Type:        "COMMAND",
					Description: "Extract the self-signed certificate",
					Command:     fmt.Sprintf("echo | openssl s_client -connect %s:443 2>/dev/null | openssl x509 -out %s.crt", details.Registry, details.Registry),
				},
				{
					Type:        "COMMAND",
					Description: "Add certificate to system trust store",
					Command:     fmt.Sprintf("sudo cp %s.crt /usr/local/share/ca-certificates/ && sudo update-ca-certificates", details.Registry),
				},
			},
		},
	}
}

// getRevokedCertRecommendations provides recommendations for revoked certificates
func (d *DefaultRecommendationEngine) getRevokedCertRecommendations(details ErrorDetails) []Recommendation {
	return []Recommendation{
		{
			Title:           "Certificate Revocation Issue",
			Description:     "The certificate has been revoked and should not be trusted. Contact the registry administrator immediately.",
			Priority:        10,
			Category:        "Security",
			IsSecurityRisk:  true,
			EstimatedEffort: "HIGH",
			Actions: []RecommendationAction{
				{
					Type:        "MANUAL",
					Description: "Contact registry administrator about certificate revocation",
					Note:        "This is a serious security issue that requires immediate attention",
				},
			},
		},
	}
}

// getInvalidSignatureRecommendations provides recommendations for invalid signatures
func (d *DefaultRecommendationEngine) getInvalidSignatureRecommendations(details ErrorDetails) []Recommendation {
	return []Recommendation{
		{
			Title:           "Certificate Integrity Issue",
			Description:     "The certificate signature is invalid, which could indicate tampering or corruption.",
			Priority:        9,
			Category:        "Security",
			IsSecurityRisk:  true,
			EstimatedEffort: "HIGH",
			Actions: []RecommendationAction{
				{
					Type:        "MANUAL",
					Description: "Report potential certificate tampering to registry administrator",
					Note:        "Invalid signatures can indicate man-in-the-middle attacks",
				},
			},
		},
	}
}

// getGenericRecommendations provides fallback recommendations for unknown errors
func (d *DefaultRecommendationEngine) getGenericRecommendations(details ErrorDetails) []Recommendation {
	return []Recommendation{
		{
			Title:           "General Certificate Troubleshooting",
			Description:     "General steps to diagnose and resolve certificate issues.",
			Priority:        5,
			Category:        "Troubleshooting",
			IsSecurityRisk:  false,
			EstimatedEffort: "MEDIUM",
			Actions: []RecommendationAction{
				{
					Type:        "COMMAND",
					Description: "Test connectivity to the registry",
					Command:     fmt.Sprintf("curl -v https://%s", details.Registry),
				},
				{
					Type:        "COMMAND",
					Description: "Check certificate details",
					Command:     fmt.Sprintf("openssl s_client -connect %s:443 -servername %s", details.Registry, details.Registry),
				},
			},
		},
	}
}

// GetQuickFix returns a quick fix command for simple errors
func (d *DefaultRecommendationEngine) GetQuickFix(details ErrorDetails) (string, error) {
	switch details.ErrorType {
	case CertExpired, CertNotYetValid:
		return "sudo ntpdate -s time.nist.gov", nil
	case CertHostnameMismatch:
		return fmt.Sprintf("nslookup %s", details.Registry), nil
	case CertSelfSigned:
		return fmt.Sprintf("echo | openssl s_client -connect %s:443 2>/dev/null | openssl x509 -out %s.crt", details.Registry, details.Registry), nil
	default:
		return fmt.Sprintf("curl -v https://%s", details.Registry), nil
	}
}

// GetDiagnosticInfo returns diagnostic information for troubleshooting
func (d *DefaultRecommendationEngine) GetDiagnosticInfo(registry string) (string, error) {
	var diagnostics strings.Builder

	diagnostics.WriteString(fmt.Sprintf("Diagnostic Information for %s:\n", registry))
	diagnostics.WriteString("=====================================\n\n")

	// Registry-specific information
	if config, exists := d.registryConfigs[registry]; exists {
		diagnostics.WriteString(fmt.Sprintf("Registry: %s\n", config.Name))
		diagnostics.WriteString(fmt.Sprintf("Documentation: %s\n", config.DocumentationURL))
		diagnostics.WriteString(fmt.Sprintf("Supports Certificate Pinning: %t\n", config.SupportsCertPinning))
		diagnostics.WriteString(fmt.Sprintf("Supports Insecure Mode: %t\n", config.SupportsInsecureMode))
		diagnostics.WriteString("\nCommon Issues:\n")
		for _, issue := range config.CommonIssues {
			diagnostics.WriteString(fmt.Sprintf("- %s\n", issue))
		}
		diagnostics.WriteString("\n")
	}

	// System information
	diagnostics.WriteString("System Diagnostic Commands:\n")
	diagnostics.WriteString("---------------------------\n")
	diagnostics.WriteString(fmt.Sprintf("Test connectivity: curl -v https://%s\n", registry))
	diagnostics.WriteString(fmt.Sprintf("Check DNS: nslookup %s\n", registry))
	diagnostics.WriteString(fmt.Sprintf("View certificate: openssl s_client -connect %s:443 -servername %s\n", registry, registry))
	diagnostics.WriteString("Check system time: timedatectl status\n")
	diagnostics.WriteString("View trusted CAs: ls /etc/ssl/certs/\n\n")

	return diagnostics.String(), nil
}

// GetSecurityAssessment evaluates the security risk of ignoring an error
func (d *DefaultRecommendationEngine) GetSecurityAssessment(details ErrorDetails) (SecurityAssessment, error) {
	assessment := SecurityAssessment{
		Compliance: []string{"SOX", "PCI-DSS", "HIPAA", "SOC2"},
	}

	switch details.ErrorType {
	case CertExpired:
		assessment.RiskLevel = "HIGH"
		assessment.RiskScore = 85
		assessment.Impact = "Expired certificates can be exploited by attackers to intercept communications"
		assessment.Mitigations = []string{
			"Use certificate pinning if possible",
			"Monitor certificate expiration dates",
			"Implement automated certificate renewal",
		}

	case CertRevoked:
		assessment.RiskLevel = "CRITICAL"
		assessment.RiskScore = 95
		assessment.Impact = "Revoked certificates indicate known compromise or security issues"
		assessment.Mitigations = []string{
			"Never ignore revoked certificates",
			"Contact security team immediately",
			"Implement certificate revocation checking",
		}

	case CertSelfSigned:
		assessment.RiskLevel = "MEDIUM"
		assessment.RiskScore = 60
		assessment.Impact = "Self-signed certificates cannot be verified against trusted authorities"
		assessment.Mitigations = []string{
			"Use certificate fingerprint verification",
			"Implement out-of-band certificate verification",
			"Consider using private CA infrastructure",
		}

	case CertUntrustedRoot:
		assessment.RiskLevel = "MEDIUM"
		assessment.RiskScore = 55
		assessment.Impact = "Unknown certificate authorities may not follow proper security practices"
		assessment.Mitigations = []string{
			"Research the certificate authority",
			"Install CA certificate only if trustworthy",
			"Consider using certificate pinning",
		}

	case CertHostnameMismatch:
		assessment.RiskLevel = "HIGH"
		assessment.RiskScore = 80
		assessment.Impact = "Hostname mismatches can indicate man-in-the-middle attacks"
		assessment.Mitigations = []string{
			"Verify you're connecting to the correct server",
			"Check for DNS manipulation",
			"Use alternative verification methods",
		}

	default:
		assessment.RiskLevel = "MEDIUM"
		assessment.RiskScore = 50
		assessment.Impact = "Unknown certificate issues should be investigated thoroughly"
		assessment.Mitigations = []string{
			"Investigate the specific error thoroughly",
			"Consult security team for guidance",
			"Document the decision and rationale",
		}
	}

	return assessment, nil
}