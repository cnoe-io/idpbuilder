package fallback

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SecurityLogger handles audit logging for all fallback decisions
type SecurityLogger struct {
	mu       sync.Mutex
	logger   *log.Logger
	file     *os.File
	logPath  string
	entries  []SecurityLogEntry
}

// SecurityLogEntry represents a single security audit log entry
type SecurityLogEntry struct {
	Timestamp   time.Time         `json:"timestamp"`
	EventType   SecurityEventType `json:"event_type"`
	Registry    string            `json:"registry"`
	Decision    string            `json:"decision"`
	Reason      string            `json:"reason"`
	Risk        SecurityRiskLevel `json:"risk_level"`
	User        string            `json:"user"`
	Success     bool              `json:"success"`
	Details     map[string]string `json:"details,omitempty"`
	FallbackID  string            `json:"fallback_id"`
}

// SecurityEventType categorizes different types of security events
type SecurityEventType string

const (
	EventCertificateFailure SecurityEventType = "CERTIFICATE_FAILURE"
	EventFallbackAttempt    SecurityEventType = "FALLBACK_ATTEMPT"
	EventFallbackSuccess    SecurityEventType = "FALLBACK_SUCCESS"
	EventInsecureModeUsed   SecurityEventType = "INSECURE_MODE_USED"
	EventSecurityDecision   SecurityEventType = "SECURITY_DECISION"
	EventRiskAccepted       SecurityEventType = "RISK_ACCEPTED"
	EventCertificateAdded   SecurityEventType = "CERTIFICATE_ADDED"
)

// NewSecurityLogger creates a new security logger
func NewSecurityLogger() (*SecurityLogger, error) {
	// Get configuration directory
	configDir := getSecurityLogDir()
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create security log directory: %w", err)
	}

	// Create log file path
	logPath := filepath.Join(configDir, "certificate-fallback-security.log")

	// Open log file with append mode
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open security log file: %w", err)
	}

	// Create logger with structured format
	logger := log.New(file, "", 0) // No default timestamp, we'll add our own

	secLogger := &SecurityLogger{
		logger:  logger,
		file:    file,
		logPath: logPath,
		entries: make([]SecurityLogEntry, 0),
	}

	// Log initialization
	secLogger.logEntry(SecurityLogEntry{
		Timestamp: time.Now().UTC(),
		EventType: EventSecurityDecision,
		Decision:  "SECURITY_LOGGER_INITIALIZED",
		Reason:    "Certificate fallback security logging started",
		Risk:      RiskNone,
		Success:   true,
	})

	return secLogger, nil
}

// LogCertificateFailure logs initial certificate validation failures
func (s *SecurityLogger) LogCertificateFailure(registry string, err error) {
	entry := SecurityLogEntry{
		Timestamp:  time.Now().UTC(),
		EventType:  EventCertificateFailure,
		Registry:   registry,
		Decision:   "CERTIFICATE_VALIDATION_FAILED",
		Reason:     err.Error(),
		Risk:       RiskNone, // Original failure is not a risk, just the starting point
		User:       getCurrentUser(),
		Success:    false,
		FallbackID: generateFallbackID(),
		Details: map[string]string{
			"error_type":    categorizeError(err),
			"error_message": err.Error(),
		},
	}

	s.logEntry(entry)
}

// LogFallbackAttempt logs attempts to use fallback strategies
func (s *SecurityLogger) LogFallbackAttempt(registry string, strategy FallbackStrategy, success bool, reason string) {
	riskLevel := calculateStrategyRisk(strategy)

	entry := SecurityLogEntry{
		Timestamp:  time.Now().UTC(),
		EventType:  EventFallbackAttempt,
		Registry:   registry,
		Decision:   fmt.Sprintf("FALLBACK_ATTEMPT_%s", strategy.String()),
		Reason:     reason,
		Risk:       riskLevel,
		User:       getCurrentUser(),
		Success:    success,
		FallbackID: generateFallbackID(),
		Details: map[string]string{
			"strategy":    strategy.String(),
			"risk_level":  riskLevel.String(),
			"attempt_id":  fmt.Sprintf("%d", time.Now().Unix()),
		},
	}

	s.logEntry(entry)
}

// LogFallbackSuccess logs successful fallback strategy usage
func (s *SecurityLogger) LogFallbackSuccess(registry string, strategy FallbackStrategy, risk SecurityRiskLevel) {
	entry := SecurityLogEntry{
		Timestamp:  time.Now().UTC(),
		EventType:  EventFallbackSuccess,
		Registry:   registry,
		Decision:   fmt.Sprintf("FALLBACK_SUCCESS_%s", strategy.String()),
		Reason:     fmt.Sprintf("Successfully connected using %s", strategy.String()),
		Risk:       risk,
		User:       getCurrentUser(),
		Success:    true,
		FallbackID: generateFallbackID(),
		Details: map[string]string{
			"strategy":         strategy.String(),
			"security_impact":  describePriorityLevel(risk),
			"connection_time":  time.Now().Format(time.RFC3339),
		},
	}

	// For high/critical risk strategies, add additional warnings
	if risk >= RiskHigh {
		entry.Details["WARNING"] = "HIGH SECURITY RISK - Review connection security"
		entry.Details["RECOMMENDATION"] = "Use proper certificates in production"
	}

	s.logEntry(entry)
}

// LogInsecureModeUsed logs when insecure mode is explicitly used
func (s *SecurityLogger) LogInsecureModeUsed(registry string, userConfirmed bool, reason string) {
	entry := SecurityLogEntry{
		Timestamp:  time.Now().UTC(),
		EventType:  EventInsecureModeUsed,
		Registry:   registry,
		Decision:   "INSECURE_MODE_USED",
		Reason:     reason,
		Risk:       RiskCritical,
		User:       getCurrentUser(),
		Success:    userConfirmed,
		FallbackID: generateFallbackID(),
		Details: map[string]string{
			"user_confirmed":    fmt.Sprintf("%v", userConfirmed),
			"security_warning":  "ALL CERTIFICATE VALIDATION DISABLED",
			"production_risk":   "CRITICAL - DO NOT USE IN PRODUCTION",
			"mitigation":        "Use proper certificates as soon as possible",
		},
	}

	s.logEntry(entry)

	// Also log to stderr for immediate visibility
	fmt.Fprintf(os.Stderr, "🚨 SECURITY WARNING: Insecure mode used for %s - ALL certificate validation disabled\n", registry)
}

// LogSecurityDecision logs general security-related decisions
func (s *SecurityLogger) LogSecurityDecision(decision, target, reason string) {
	entry := SecurityLogEntry{
		Timestamp:  time.Now().UTC(),
		EventType:  EventSecurityDecision,
		Registry:   target,
		Decision:   decision,
		Reason:     reason,
		Risk:       RiskMedium, // General security decisions are medium risk
		User:       getCurrentUser(),
		Success:    true,
		FallbackID: generateFallbackID(),
		Details: map[string]string{
			"decision_type": decision,
			"affected_target": target,
		},
	}

	s.logEntry(entry)
}

// LogRiskAcceptance logs when users explicitly accept security risks
func (s *SecurityLogger) LogRiskAcceptance(registry string, risk SecurityRiskLevel, strategy string, userChoice bool) {
	entry := SecurityLogEntry{
		Timestamp:  time.Now().UTC(),
		EventType:  EventRiskAccepted,
		Registry:   registry,
		Decision:   fmt.Sprintf("RISK_ACCEPTANCE_%s", risk.String()),
		Reason:     fmt.Sprintf("User %s risk for strategy: %s",
			map[bool]string{true: "accepted", false: "rejected"}[userChoice], strategy),
		Risk:       risk,
		User:       getCurrentUser(),
		Success:    userChoice,
		FallbackID: generateFallbackID(),
		Details: map[string]string{
			"risk_level":     risk.String(),
			"strategy":       strategy,
			"user_decision":  fmt.Sprintf("%v", userChoice),
			"risk_summary":   describePriorityLevel(risk),
		},
	}

	s.logEntry(entry)
}

// GetSecuritySummary returns a summary of recent security events
func (s *SecurityLogger) GetSecuritySummary(since time.Duration) SecuritySummary {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-since)
	summary := SecuritySummary{
		Period: since,
		Events: make(map[SecurityEventType]int),
		RisksByLevel: make(map[SecurityRiskLevel]int),
		Registries: make(map[string]int),
	}

	for _, entry := range s.entries {
		if entry.Timestamp.After(cutoff) {
			summary.Events[entry.EventType]++
			summary.RisksByLevel[entry.Risk]++
			summary.Registries[entry.Registry]++

			if entry.Risk >= RiskHigh {
				summary.HighRiskEvents++
			}
			if entry.EventType == EventInsecureModeUsed {
				summary.InsecureModeUsages++
			}
		}
	}

	return summary
}

// SecuritySummary provides an overview of security events
type SecuritySummary struct {
	Period            time.Duration
	Events            map[SecurityEventType]int
	RisksByLevel      map[SecurityRiskLevel]int
	Registries        map[string]int
	HighRiskEvents    int
	InsecureModeUsages int
}

// logEntry writes a security log entry to the log file and memory
func (s *SecurityLogger) logEntry(entry SecurityLogEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Add to memory cache
	s.entries = append(s.entries, entry)

	// Keep only last 1000 entries in memory
	if len(s.entries) > 1000 {
		s.entries = s.entries[len(s.entries)-1000:]
	}

	// Format as JSON for structured logging
	jsonData, err := json.Marshal(entry)
	if err != nil {
		// Fallback to plain text if JSON fails
		s.logger.Printf("SECURITY_LOG_ERROR: Failed to marshal log entry: %v", err)
		return
	}

	// Write to log file
	s.logger.Printf("%s", string(jsonData))
}

// Close closes the security logger and its file handle
func (s *SecurityLogger) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file != nil {
		// Log shutdown
		shutdownEntry := SecurityLogEntry{
			Timestamp: time.Now().UTC(),
			EventType: EventSecurityDecision,
			Decision:  "SECURITY_LOGGER_SHUTDOWN",
			Reason:    "Certificate fallback security logging stopped",
			Risk:      RiskNone,
			Success:   true,
		}

		// Write shutdown entry
		if jsonData, err := json.Marshal(shutdownEntry); err == nil {
			s.logger.Printf("%s", string(jsonData))
		}

		return s.file.Close()
	}
	return nil
}

// Helper functions

// getSecurityLogDir returns the directory for security logs
func getSecurityLogDir() string {
	if dir := os.Getenv("IDPBUILDER_SECURITY_LOG_DIR"); dir != "" {
		return dir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".idpbuilder/security"
	}

	return filepath.Join(home, ".idpbuilder", "security")
}

// getCurrentUser returns the current user for audit purposes
func getCurrentUser() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	return "unknown"
}

// generateFallbackID creates a unique identifier for fallback operations
func generateFallbackID() string {
	return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
}

// categorizeError categorizes error types for better analysis
func categorizeError(err error) string {
	errorStr := err.Error()
	switch {
	case strings.Contains(errorStr, "unknown authority"):
		return "unknown_ca"
	case strings.Contains(errorStr, "expired"):
		return "expired_certificate"
	case strings.Contains(errorStr, "hostname"):
		return "hostname_mismatch"
	case strings.Contains(errorStr, "self-signed"):
		return "self_signed"
	case strings.Contains(errorStr, "tls"):
		return "tls_handshake"
	default:
		return "generic_certificate_error"
	}
}

// calculateStrategyRisk determines the security risk level for each strategy
func calculateStrategyRisk(strategy FallbackStrategy) SecurityRiskLevel {
	switch strategy {
	case StrategyRetryWithSystemCA:
		return RiskLow
	case StrategyRetryWithoutSNI:
		return RiskMedium
	case StrategyRetryWithLowerTLS:
		return RiskMedium
	case StrategyAcceptSelfSigned:
		return RiskHigh
	case StrategyInsecureMode:
		return RiskCritical
	default:
		return RiskNone
	}
}

// describePriorityLevel provides human-readable risk descriptions
func describePriorityLevel(risk SecurityRiskLevel) string {
	switch risk {
	case RiskNone:
		return "No additional security risk"
	case RiskLow:
		return "Minimal security impact, acceptable for most environments"
	case RiskMedium:
		return "Moderate security risk, review before production use"
	case RiskHigh:
		return "High security risk, not recommended for production"
	case RiskCritical:
		return "Critical security risk, never use in production"
	default:
		return "Unknown risk level"
	}
}