package fallback

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// InsecureManager handles insecure mode decisions and user confirmations
type InsecureManager struct {
	mu              sync.Mutex
	securityLogger  *SecurityLogger
	confirmedRegistries map[string]*InsecureDecision
	config          *InsecureConfig
}

// InsecureConfig holds configuration for insecure mode behavior
type InsecureConfig struct {
	// Require explicit confirmation for each registry
	RequireConfirmation bool

	// Show detailed warnings before confirmation
	ShowWarnings bool

	// Remember confirmations for the session
	RememberDecisions bool

	// Allow environment variable bypass
	AllowEnvOverride bool

	// Maximum time to wait for user input
	ConfirmationTimeout time.Duration
}

// InsecureDecision tracks a user's decision about insecure mode
type InsecureDecision struct {
	Registry      string
	Confirmed     bool
	Timestamp     time.Time
	Reason        string
	UserInput     string
	SessionOnly   bool
}

// InsecureWarning represents different types of warnings for insecure mode
type InsecureWarning struct {
	Level    WarningLevel
	Title    string
	Message  string
	Risks    []string
	Mitigations []string
}

// WarningLevel indicates the severity of insecure mode warnings
type WarningLevel int

const (
	WarningInfo WarningLevel = iota
	WarningCaution
	WarningDanger
	WarningCritical
)

// NewInsecureManager creates a new insecure mode manager
func NewInsecureManager(securityLogger *SecurityLogger) *InsecureManager {
	return &InsecureManager{
		securityLogger:     securityLogger,
		confirmedRegistries: make(map[string]*InsecureDecision),
		config:             DefaultInsecureConfig(),
	}
}

// ConfirmInsecureMode requests user confirmation for insecure mode usage
func (m *InsecureManager) ConfirmInsecureMode(registry string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already confirmed for this registry in this session
	if decision, exists := m.confirmedRegistries[registry]; exists && m.config.RememberDecisions {
		if time.Since(decision.Timestamp) < time.Hour { // Expire after 1 hour
			m.securityLogger.LogSecurityDecision("INSECURE_MODE_REUSED",
				registry, "Using previous insecure mode confirmation")
			return decision.Confirmed, nil
		}
		// Expired, remove old decision
		delete(m.confirmedRegistries, registry)
	}

	// Check for environment variable override
	if m.config.AllowEnvOverride {
		if envDecision := m.checkEnvironmentOverride(registry); envDecision != nil {
			m.confirmedRegistries[registry] = envDecision
			m.securityLogger.LogInsecureModeUsed(registry, envDecision.Confirmed,
				"Environment variable override: "+envDecision.Reason)
			return envDecision.Confirmed, nil
		}
	}

	// Show warnings and get user confirmation
	confirmed, reason, err := m.promptUserConfirmation(registry)
	if err != nil {
		return false, fmt.Errorf("failed to get user confirmation: %w", err)
	}

	// Record the decision
	decision := &InsecureDecision{
		Registry:    registry,
		Confirmed:   confirmed,
		Timestamp:   time.Now(),
		Reason:      reason,
		SessionOnly: true,
	}

	m.confirmedRegistries[registry] = decision
	m.securityLogger.LogInsecureModeUsed(registry, confirmed, reason)

	return confirmed, nil
}

// checkEnvironmentOverride checks for environment variable overrides
func (m *InsecureManager) checkEnvironmentOverride(registry string) *InsecureDecision {
	// Global insecure mode
	if os.Getenv("IDPBUILDER_INSECURE_REGISTRIES") == "true" ||
		os.Getenv("IDPBUILDER_SKIP_TLS_VERIFY") == "true" {
		return &InsecureDecision{
			Registry:    registry,
			Confirmed:   true,
			Timestamp:   time.Now(),
			Reason:      "Global insecure mode enabled via environment variable",
			UserInput:   "ENV_VAR_OVERRIDE",
			SessionOnly: true,
		}
	}

	// Registry-specific insecure mode
	envVar := fmt.Sprintf("IDPBUILDER_INSECURE_%s", sanitizeRegistryForEnv(registry))
	if os.Getenv(envVar) == "true" {
		return &InsecureDecision{
			Registry:    registry,
			Confirmed:   true,
			Timestamp:   time.Now(),
			Reason:      fmt.Sprintf("Registry-specific insecure mode enabled via %s", envVar),
			UserInput:   "ENV_VAR_OVERRIDE",
			SessionOnly: true,
		}
	}

	return nil
}

// promptUserConfirmation shows warnings and prompts for user confirmation
func (m *InsecureManager) promptUserConfirmation(registry string) (bool, string, error) {
	if !m.config.RequireConfirmation {
		return true, "Confirmation not required by configuration", nil
	}

	// Show comprehensive warnings
	if m.config.ShowWarnings {
		m.displayInsecureWarnings(registry)
	}

	// Prompt for confirmation
	fmt.Printf("\n🚨 CRITICAL SECURITY WARNING 🚨\n")
	fmt.Printf("You are about to disable ALL certificate validation for registry: %s\n", registry)
	fmt.Printf("This creates a CRITICAL security vulnerability!\n\n")

	fmt.Printf("Do you understand the risks and want to proceed? (type 'yes' to confirm): ")

	// Set up input with timeout
	input := make(chan string, 1)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		input <- strings.TrimSpace(response)
	}()

	// Wait for input with timeout
	select {
	case response := <-input:
		confirmed := strings.ToLower(response) == "yes"
		reason := fmt.Sprintf("User %s insecure mode with input: %s",
			map[bool]string{true: "confirmed", false: "rejected"}[confirmed], response)

		if confirmed {
			m.showPostConfirmationWarning(registry)
		}

		return confirmed, reason, nil

	case <-time.After(m.config.ConfirmationTimeout):
		return false, "User confirmation timed out", nil
	}
}

// displayInsecureWarnings shows detailed warnings about insecure mode
func (m *InsecureManager) displayInsecureWarnings(registry string) {
	warnings := m.generateInsecureWarnings(registry)

	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("🚨 INSECURE MODE WARNINGS FOR: %s\n", registry)
	fmt.Printf(strings.Repeat("=", 80) + "\n")

	for i, warning := range warnings {
		fmt.Printf("\n%d. %s [%s]\n", i+1, warning.Title, m.formatWarningLevel(warning.Level))
		fmt.Printf("   %s\n", warning.Message)

		if len(warning.Risks) > 0 {
			fmt.Printf("   RISKS:\n")
			for _, risk := range warning.Risks {
				fmt.Printf("   • %s\n", risk)
			}
		}

		if len(warning.Mitigations) > 0 {
			fmt.Printf("   MITIGATIONS:\n")
			for _, mitigation := range warning.Mitigations {
				fmt.Printf("   • %s\n", mitigation)
			}
		}
	}

	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
}

// generateInsecureWarnings creates comprehensive warnings for insecure mode
func (m *InsecureManager) generateInsecureWarnings(registry string) []InsecureWarning {
	return []InsecureWarning{
		{
			Level:   WarningCritical,
			Title:   "Certificate Validation Completely Disabled",
			Message: "ALL TLS certificate validation will be bypassed",
			Risks: []string{
				"Man-in-the-middle attacks possible",
				"Connection to malicious registries possible",
				"No verification of registry identity",
				"Encrypted traffic can be intercepted and modified",
			},
			Mitigations: []string{
				"Use only in isolated development environments",
				"Never use with production data",
				"Obtain proper certificates as soon as possible",
			},
		},
		{
			Level:   WarningDanger,
			Title:   "Data Integrity at Risk",
			Message: "Container images and metadata cannot be verified as authentic",
			Risks: []string{
				"Malicious images could be pulled without detection",
				"Image tampering will not be detected",
				"Supply chain attacks possible",
				"Compromised registry won't be detected",
			},
			Mitigations: []string{
				"Verify image signatures separately if available",
				"Use trusted base images only",
				"Monitor for unexpected image changes",
			},
		},
		{
			Level:   WarningCaution,
			Title:   "Compliance and Audit Issues",
			Message: "Insecure connections may violate security policies",
			Risks: []string{
				"May violate organizational security policies",
				"Could fail security audits",
				"Regulatory compliance issues possible",
				"Security team notifications may be triggered",
			},
			Mitigations: []string{
				"Document temporary usage with business justification",
				"Set calendar reminder to fix certificate issues",
				"Notify security team if required by policy",
			},
		},
		{
			Level:   WarningInfo,
			Title:   "Alternative Solutions Available",
			Message: "Consider these alternatives before using insecure mode",
			Risks: []string{
				"Missing opportunity to properly fix the issue",
				"Creating technical debt",
				"Setting precedent for insecure practices",
			},
			Mitigations: []string{
				"Extract and trust the registry's CA certificate",
				"Contact registry administrator for proper certificates",
				"Use a properly configured registry",
				"Set up certificate auto-renewal",
			},
		},
	}
}

// showPostConfirmationWarning shows final warning after user confirmation
func (m *InsecureManager) showPostConfirmationWarning(registry string) {
	fmt.Printf("\n⚠️  INSECURE MODE ACTIVATED FOR: %s\n", registry)
	fmt.Printf("⚠️  Certificate validation is now DISABLED\n")
	fmt.Printf("⚠️  This connection is NOT SECURE\n")
	fmt.Printf("⚠️  Please fix certificate issues as soon as possible\n\n")

	// Also write to stderr for logging/monitoring systems
	fmt.Fprintf(os.Stderr, "SECURITY_ALERT: Insecure mode activated for registry %s\n", registry)
}

// IsRegistryInsecure checks if a registry is currently configured for insecure mode
func (m *InsecureManager) IsRegistryInsecure(registry string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if decision, exists := m.confirmedRegistries[registry]; exists {
		// Check if decision is still valid
		if time.Since(decision.Timestamp) < time.Hour {
			return decision.Confirmed
		}
		// Expired, remove old decision
		delete(m.confirmedRegistries, registry)
	}

	return false
}

// RevokeInsecureMode removes insecure mode for a registry
func (m *InsecureManager) RevokeInsecureMode(registry string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.confirmedRegistries[registry]; exists {
		delete(m.confirmedRegistries, registry)
		m.securityLogger.LogSecurityDecision("INSECURE_MODE_REVOKED",
			registry, "Insecure mode manually revoked")

		fmt.Printf("✅ Insecure mode revoked for registry: %s\n", registry)
		fmt.Printf("✅ Certificate validation will be enforced on next connection\n")
	}
}

// GetInsecureRegistries returns all registries currently in insecure mode
func (m *InsecureManager) GetInsecureRegistries() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	var registries []string
	for registry, decision := range m.confirmedRegistries {
		if decision.Confirmed && time.Since(decision.Timestamp) < time.Hour {
			registries = append(registries, registry)
		}
	}

	return registries
}

// ShowInsecureStatus displays current insecure mode status
func (m *InsecureManager) ShowInsecureStatus() {
	insecureRegistries := m.GetInsecureRegistries()

	if len(insecureRegistries) == 0 {
		fmt.Printf("✅ No registries are currently in insecure mode\n")
		return
	}

	fmt.Printf("⚠️  WARNING: The following registries are in INSECURE mode:\n")
	for _, registry := range insecureRegistries {
		if decision, exists := m.confirmedRegistries[registry]; exists {
			fmt.Printf("   • %s (since %s)\n", registry,
				decision.Timestamp.Format("15:04:05"))
		}
	}
	fmt.Printf("\n⚠️  These connections are NOT SECURE\n")
	fmt.Printf("💡 Use 'idpbuilder registry secure <registry>' to re-enable certificate validation\n")
}

// Helper functions

// DefaultInsecureConfig returns secure default configuration
func DefaultInsecureConfig() *InsecureConfig {
	return &InsecureConfig{
		RequireConfirmation:  true,
		ShowWarnings:         true,
		RememberDecisions:    true,
		AllowEnvOverride:     true,
		ConfirmationTimeout:  30 * time.Second,
	}
}

// formatWarningLevel returns a colored string for warning levels
func (m *InsecureManager) formatWarningLevel(level WarningLevel) string {
	switch level {
	case WarningInfo:
		return "INFO"
	case WarningCaution:
		return "CAUTION"
	case WarningDanger:
		return "DANGER"
	case WarningCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// sanitizeRegistryForEnv converts registry name to valid environment variable format
func sanitizeRegistryForEnv(registry string) string {
	// Replace problematic characters with underscores
	sanitized := strings.ToUpper(registry)
	sanitized = strings.ReplaceAll(sanitized, ":", "_")
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, ".", "_")
	sanitized = strings.ReplaceAll(sanitized, "-", "_")
	return sanitized
}

// ValidateInsecureConfig validates insecure mode configuration
func ValidateInsecureConfig(config *InsecureConfig) error {
	if config.ConfirmationTimeout < time.Second {
		return fmt.Errorf("confirmation timeout too short: %v", config.ConfirmationTimeout)
	}
	if config.ConfirmationTimeout > 5*time.Minute {
		return fmt.Errorf("confirmation timeout too long: %v", config.ConfirmationTimeout)
	}
	return nil
}