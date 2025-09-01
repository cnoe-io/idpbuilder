// Package fallback provides certificate error detection and fallback handling
// for when TLS operations fail due to certificate issues.
package fallback

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"
	"time"
)

// FallbackAction represents the action to take when a certificate error occurs
type FallbackAction int

const (
	// ActionDeny rejects the connection (secure default)
	ActionDeny FallbackAction = iota
	
	// ActionAccept accepts the connection despite the certificate error
	ActionAccept
	
	// ActionPrompt asks the user to decide (interactive mode)
	ActionPrompt
	
	// ActionLog logs the error but allows the connection
	ActionLog
	
	// ActionRetry attempts to retry the connection with different parameters
	ActionRetry
)

// String returns a human-readable representation of the fallback action
func (a FallbackAction) String() string {
	switch a {
	case ActionDeny:
		return "deny"
	case ActionAccept:
		return "accept"
	case ActionPrompt:
		return "prompt"
	case ActionLog:
		return "log"
	case ActionRetry:
		return "retry"
	default:
		return "unknown"
	}
}

// FallbackDecision represents a decision made by the fallback handler
type FallbackDecision struct {
	// Action is the recommended action
	Action FallbackAction
	
	// Reason explains why this action was chosen
	Reason string
	
	// TLSConfig contains modified TLS configuration if applicable
	TLSConfig *tls.Config
	
	// Timeout specifies how long to wait before timing out the connection
	Timeout time.Duration
	
	// Permanent indicates whether this decision should be remembered
	Permanent bool
	
	// RetryCount indicates how many retries are recommended
	RetryCount int
	
	// SecurityRisk indicates the assessed security risk level (0-10)
	SecurityRisk int
	
	// Metadata contains additional context about the decision
	Metadata map[string]interface{}
}

// FallbackStrategy contains configuration for how to handle different certificate errors
type FallbackStrategy struct {
	// Mode determines the overall security posture
	Mode FallbackMode
	
	// Rules contains specific rules for different error types
	Rules map[CertErrorType]FallbackAction
	
	// Hostnames contains hostname-specific overrides
	Hostnames map[string]FallbackAction
	
	// InsecureHosts contains hosts where insecure connections are always allowed
	InsecureHosts map[string]bool
	
	// PromptEnabled indicates whether user prompting is available
	PromptEnabled bool
	
	// LoggingEnabled indicates whether to log security decisions
	LoggingEnabled bool
	
	// MaxRetries specifies the maximum number of connection retries
	MaxRetries int
	
	// RetryDelay specifies the delay between retries
	RetryDelay time.Duration
	
	// TrustNewCerts indicates whether to automatically trust new certificates
	TrustNewCerts bool
	
	// RememberDecisions indicates whether to remember user decisions
	RememberDecisions bool
	
	// DecisionTimeout is how long to wait for user decisions
	DecisionTimeout time.Duration
}

// FallbackMode represents different security modes
type FallbackMode int

const (
	// ModeSecure denies all certificate errors by default
	ModeSecure FallbackMode = iota
	
	// ModePermissive allows most certificate errors with logging
	ModePermissive
	
	// ModeDevelopment allows all certificate errors (unsafe for production)
	ModeDevelopment
	
	// ModeInteractive prompts the user for decisions
	ModeInteractive
	
	// ModeCustom uses the configured rules
	ModeCustom
)

// String returns a human-readable representation of the fallback mode
func (m FallbackMode) String() string {
	switch m {
	case ModeSecure:
		return "secure"
	case ModePermissive:
		return "permissive"
	case ModeDevelopment:
		return "development"
	case ModeInteractive:
		return "interactive"
	case ModeCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// FallbackHandler interface defines methods for handling certificate errors
type FallbackHandler interface {
	// HandleError processes a certificate error and returns a decision
	HandleError(ctx context.Context, errorDetails *ErrorDetails) (*FallbackDecision, error)
	
	// GetStrategy returns the current fallback strategy
	GetStrategy() *FallbackStrategy
	
	// UpdateStrategy updates the fallback strategy
	UpdateStrategy(strategy *FallbackStrategy) error
	
	// IsHostTrusted checks if a hostname is in the trusted hosts list
	IsHostTrusted(hostname string) bool
	
	// AddTrustedHost adds a hostname to the trusted hosts list
	AddTrustedHost(hostname string) error
	
	// RemoveTrustedHost removes a hostname from the trusted hosts list
	RemoveTrustedHost(hostname string) error
	
	// CreateTLSConfig creates a TLS configuration with appropriate settings
	CreateTLSConfig(decision *FallbackDecision) (*tls.Config, error)
	
	// LogSecurityDecision logs a security decision (implementation in Split 002)
	LogSecurityDecision(decision *FallbackDecision, errorDetails *ErrorDetails)
}

// UserPrompter interface defines methods for prompting users for decisions
type UserPrompter interface {
	// PromptForDecision asks the user to make a decision about a certificate error
	PromptForDecision(ctx context.Context, errorDetails *ErrorDetails) (FallbackAction, error)
	
	// PromptForTrust asks the user whether to trust a certificate permanently
	PromptForTrust(ctx context.Context, cert *x509.Certificate) (bool, error)
}

// SecurityLogger interface defines methods for logging security-related events
// Implementation will be provided in Split 002
type SecurityLogger interface {
	// LogCertificateError logs a certificate error
	LogCertificateError(errorDetails *ErrorDetails)
	
	// LogFallbackDecision logs a fallback decision
	LogFallbackDecision(decision *FallbackDecision, errorDetails *ErrorDetails)
	
	// LogSecurityRisk logs a security risk assessment
	LogSecurityRisk(level int, reason string, context map[string]interface{})
}

// DefaultFallbackHandler implements the FallbackHandler interface
type DefaultFallbackHandler struct {
	// strategy contains the current fallback strategy
	strategy *FallbackStrategy
	
	// detector is used to analyze certificate errors
	detector CertErrorDetector
	
	// prompter is used for user interaction (optional)
	prompter UserPrompter
	
	// logger is used for security logging (optional)
	logger SecurityLogger
	
	// decisions stores remembered user decisions
	decisions map[string]*FallbackDecision
	
	// mutex protects concurrent access to decisions
	mutex sync.RWMutex
	
	// trustedHosts contains hosts that are explicitly trusted
	trustedHosts map[string]bool
	
	// hostMutex protects concurrent access to trustedHosts
	hostMutex sync.RWMutex
}

// NewDefaultFallbackHandler creates a new instance of DefaultFallbackHandler
func NewDefaultFallbackHandler(detector CertErrorDetector) *DefaultFallbackHandler {
	return &DefaultFallbackHandler{
		strategy:     NewSecureStrategy(),
		detector:     detector,
		decisions:    make(map[string]*FallbackDecision),
		trustedHosts: make(map[string]bool),
	}
}

// NewSecureStrategy creates a secure fallback strategy (denies all errors)
func NewSecureStrategy() *FallbackStrategy {
	rules := make(map[CertErrorType]FallbackAction)
	
	// Secure mode denies all certificate errors by default
	for errorType := ErrorTypeUnknown; errorType <= ErrorTypeKeyUsage; errorType++ {
		rules[errorType] = ActionDeny
	}
	
	return &FallbackStrategy{
		Mode:              ModeSecure,
		Rules:             rules,
		Hostnames:         make(map[string]FallbackAction),
		InsecureHosts:     make(map[string]bool),
		PromptEnabled:     false,
		LoggingEnabled:    true,
		MaxRetries:        0,
		RetryDelay:        0,
		TrustNewCerts:     false,
		RememberDecisions: false,
		DecisionTimeout:   30 * time.Second,
	}
}

// NewDevelopmentStrategy creates a permissive fallback strategy (allows most errors)
func NewDevelopmentStrategy() *FallbackStrategy {
	rules := make(map[CertErrorType]FallbackAction)
	
	// Development mode allows most certificate errors with logging
	rules[ErrorTypeExpired] = ActionDeny // Still deny expired certificates
	rules[ErrorTypeRevoked] = ActionDeny // Still deny revoked certificates
	rules[ErrorTypeBadCertificate] = ActionDeny // Still deny malformed certificates
	
	// Allow these errors in development
	rules[ErrorTypeHostnameMismatch] = ActionLog
	rules[ErrorTypeSelfSigned] = ActionLog
	rules[ErrorTypeUntrustedCA] = ActionLog
	rules[ErrorTypeNotYetValid] = ActionLog
	rules[ErrorTypeChainIncomplete] = ActionLog
	rules[ErrorTypeKeyUsage] = ActionLog
	rules[ErrorTypeUnknown] = ActionLog
	
	return &FallbackStrategy{
		Mode:              ModeDevelopment,
		Rules:             rules,
		Hostnames:         make(map[string]FallbackAction),
		InsecureHosts:     make(map[string]bool),
		PromptEnabled:     false,
		LoggingEnabled:    true,
		MaxRetries:        3,
		RetryDelay:        1 * time.Second,
		TrustNewCerts:     false,
		RememberDecisions: true,
		DecisionTimeout:   30 * time.Second,
	}
}

// NewInteractiveStrategy creates an interactive fallback strategy (prompts user)
func NewInteractiveStrategy(prompter UserPrompter) *FallbackStrategy {
	rules := make(map[CertErrorType]FallbackAction)
	
	// Interactive mode prompts for most decisions
	rules[ErrorTypeExpired] = ActionDeny // Always deny expired certificates
	rules[ErrorTypeRevoked] = ActionDeny // Always deny revoked certificates
	rules[ErrorTypeBadCertificate] = ActionDeny // Always deny malformed certificates
	
	// Prompt for these errors
	rules[ErrorTypeHostnameMismatch] = ActionPrompt
	rules[ErrorTypeSelfSigned] = ActionPrompt
	rules[ErrorTypeUntrustedCA] = ActionPrompt
	rules[ErrorTypeNotYetValid] = ActionPrompt
	rules[ErrorTypeChainIncomplete] = ActionPrompt
	rules[ErrorTypeKeyUsage] = ActionPrompt
	rules[ErrorTypeUnknown] = ActionPrompt
	
	return &FallbackStrategy{
		Mode:              ModeInteractive,
		Rules:             rules,
		Hostnames:         make(map[string]FallbackAction),
		InsecureHosts:     make(map[string]bool),
		PromptEnabled:     true,
		LoggingEnabled:    true,
		MaxRetries:        2,
		RetryDelay:        1 * time.Second,
		TrustNewCerts:     false,
		RememberDecisions: true,
		DecisionTimeout:   60 * time.Second,
	}
}

// HandleError processes a certificate error and returns a decision
func (h *DefaultFallbackHandler) HandleError(ctx context.Context, errorDetails *ErrorDetails) (*FallbackDecision, error) {
	if errorDetails == nil {
		return nil, fmt.Errorf("error details cannot be nil")
	}
	
	// Check if we have a remembered decision for this error
	h.mutex.RLock()
	decisionKey := h.createDecisionKey(errorDetails)
	if remembered, exists := h.decisions[decisionKey]; exists && h.strategy.RememberDecisions {
		h.mutex.RUnlock()
		return remembered, nil
	}
	h.mutex.RUnlock()
	
	// Check hostname-specific rules first
	if action, exists := h.strategy.Hostnames[errorDetails.Hostname]; exists {
		decision := h.createDecision(action, errorDetails, fmt.Sprintf("hostname-specific rule for %s", errorDetails.Hostname))
		h.rememberDecision(decisionKey, decision)
		return decision, nil
	}
	
	// Check if host is in insecure hosts list
	if h.strategy.InsecureHosts[errorDetails.Hostname] {
		decision := h.createDecision(ActionAccept, errorDetails, fmt.Sprintf("hostname %s is in insecure hosts list", errorDetails.Hostname))
		return decision, nil
	}
	
	// Get the action based on error type and mode
	action := h.determineAction(errorDetails)
	
	// Handle user prompting if needed
	if action == ActionPrompt && h.strategy.PromptEnabled && h.prompter != nil {
		userAction, err := h.prompter.PromptForDecision(ctx, errorDetails)
		if err != nil {
			// If prompting fails, fall back to deny for security
			action = ActionDeny
		} else {
			action = userAction
		}
	}
	
	decision := h.createDecision(action, errorDetails, h.getReasonForAction(action, errorDetails))
	
	// Remember the decision if configured to do so
	if h.strategy.RememberDecisions {
		h.rememberDecision(decisionKey, decision)
	}
	
	// Log the decision
	if h.logger != nil && h.strategy.LoggingEnabled {
		h.logger.LogFallbackDecision(decision, errorDetails)
	}
	
	return decision, nil
}

// GetStrategy returns the current fallback strategy
func (h *DefaultFallbackHandler) GetStrategy() *FallbackStrategy {
	return h.strategy
}

// UpdateStrategy updates the fallback strategy
func (h *DefaultFallbackHandler) UpdateStrategy(strategy *FallbackStrategy) error {
	if strategy == nil {
		return fmt.Errorf("strategy cannot be nil")
	}
	
	h.strategy = strategy
	
	// Clear remembered decisions when strategy changes
	h.mutex.Lock()
	h.decisions = make(map[string]*FallbackDecision)
	h.mutex.Unlock()
	
	return nil
}

// IsHostTrusted checks if a hostname is in the trusted hosts list
func (h *DefaultFallbackHandler) IsHostTrusted(hostname string) bool {
	h.hostMutex.RLock()
	defer h.hostMutex.RUnlock()
	return h.trustedHosts[hostname] || h.strategy.InsecureHosts[hostname]
}

// AddTrustedHost adds a hostname to the trusted hosts list
func (h *DefaultFallbackHandler) AddTrustedHost(hostname string) error {
	if hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}
	
	h.hostMutex.Lock()
	defer h.hostMutex.Unlock()
	h.trustedHosts[hostname] = true
	
	return nil
}

// RemoveTrustedHost removes a hostname from the trusted hosts list
func (h *DefaultFallbackHandler) RemoveTrustedHost(hostname string) error {
	h.hostMutex.Lock()
	defer h.hostMutex.Unlock()
	
	delete(h.trustedHosts, hostname)
	delete(h.strategy.InsecureHosts, hostname)
	
	return nil
}

// CreateTLSConfig creates a TLS configuration with appropriate settings
func (h *DefaultFallbackHandler) CreateTLSConfig(decision *FallbackDecision) (*tls.Config, error) {
	if decision == nil {
		return nil, fmt.Errorf("decision cannot be nil")
	}
	
	// If decision already includes TLS config, use it
	if decision.TLSConfig != nil {
		return decision.TLSConfig, nil
	}
	
	// Create base TLS config
	config := &tls.Config{
		MinVersion: tls.VersionTLS12, // Secure default
	}
	
	// Modify based on the decision
	switch decision.Action {
	case ActionAccept, ActionLog:
		// For accept/log actions, disable certificate verification
		config.InsecureSkipVerify = true
		
	case ActionDeny:
		// For deny actions, use strict verification (default)
		config.InsecureSkipVerify = false
		
	case ActionRetry:
		// For retry actions, keep verification enabled but allow retries
		config.InsecureSkipVerify = false
	}
	
	return config, nil
}

// LogSecurityDecision logs a security decision
// NOTE: This is a stub implementation. Full implementation will be in Split 002
func (h *DefaultFallbackHandler) LogSecurityDecision(decision *FallbackDecision, errorDetails *ErrorDetails) {
	if h.logger != nil {
		h.logger.LogFallbackDecision(decision, errorDetails)
	}
	// Stub implementation - actual logging will be implemented in Split 002
}

// SetUserPrompter sets the user prompter for interactive decisions
func (h *DefaultFallbackHandler) SetUserPrompter(prompter UserPrompter) {
	h.prompter = prompter
}

// SetSecurityLogger sets the security logger
func (h *DefaultFallbackHandler) SetSecurityLogger(logger SecurityLogger) {
	h.logger = logger
}

// determineAction determines the appropriate action based on the error and strategy
func (h *DefaultFallbackHandler) determineAction(errorDetails *ErrorDetails) FallbackAction {
	// Check type-specific rules first
	if action, exists := h.strategy.Rules[errorDetails.Type]; exists {
		return action
	}
	
	// Fall back to mode-based defaults
	switch h.strategy.Mode {
	case ModeSecure:
		return ActionDeny
		
	case ModePermissive, ModeDevelopment:
		if errorDetails.IsRecoverable {
			return ActionLog
		}
		return ActionDeny
		
	case ModeInteractive:
		if errorDetails.IsRecoverable {
			return ActionPrompt
		}
		return ActionDeny
		
	case ModeCustom:
		// Custom mode should have explicit rules, but fall back to deny if not found
		return ActionDeny
		
	default:
		return ActionDeny
	}
}

// createDecision creates a FallbackDecision with appropriate settings
func (h *DefaultFallbackHandler) createDecision(action FallbackAction, errorDetails *ErrorDetails, reason string) *FallbackDecision {
	decision := &FallbackDecision{
		Action:       action,
		Reason:       reason,
		Timeout:      30 * time.Second,
		Permanent:    h.strategy.RememberDecisions,
		RetryCount:   h.strategy.MaxRetries,
		SecurityRisk: h.assessSecurityRisk(errorDetails),
		Metadata:     make(map[string]interface{}),
	}
	
	// Add metadata
	decision.Metadata["error_type"] = errorDetails.Type.String()
	decision.Metadata["hostname"] = errorDetails.Hostname
	decision.Metadata["timestamp"] = errorDetails.Timestamp
	decision.Metadata["recoverable"] = errorDetails.IsRecoverable
	
	// Create appropriate TLS config
	config, _ := h.CreateTLSConfig(decision)
	decision.TLSConfig = config
	
	return decision
}

// createDecisionKey creates a unique key for remembering decisions
func (h *DefaultFallbackHandler) createDecisionKey(errorDetails *ErrorDetails) string {
	return fmt.Sprintf("%s:%s:%s", errorDetails.Type.String(), errorDetails.Hostname, 
		errorDetails.Certificate.Subject.String())
}

// rememberDecision stores a decision for future use
func (h *DefaultFallbackHandler) rememberDecision(key string, decision *FallbackDecision) {
	if !h.strategy.RememberDecisions {
		return
	}
	
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.decisions[key] = decision
}

// getReasonForAction provides a human-readable reason for an action
func (h *DefaultFallbackHandler) getReasonForAction(action FallbackAction, errorDetails *ErrorDetails) string {
	switch action {
	case ActionAccept:
		return fmt.Sprintf("Accepting connection despite %s error due to policy configuration", errorDetails.Type.String())
	case ActionDeny:
		return fmt.Sprintf("Denying connection due to %s error for security", errorDetails.Type.String())
	case ActionLog:
		return fmt.Sprintf("Allowing connection but logging %s error", errorDetails.Type.String())
	case ActionPrompt:
		return fmt.Sprintf("User decision required for %s error", errorDetails.Type.String())
	case ActionRetry:
		return fmt.Sprintf("Retrying connection due to %s error", errorDetails.Type.String())
	default:
		return "Unknown action"
	}
}

// assessSecurityRisk assesses the security risk level of an error (0-10)
func (h *DefaultFallbackHandler) assessSecurityRisk(errorDetails *ErrorDetails) int {
	switch errorDetails.Type {
	case ErrorTypeExpired:
		return 8 // High risk - certificate could be compromised
	case ErrorTypeRevoked:
		return 10 // Maximum risk - certificate is known to be compromised
	case ErrorTypeBadCertificate:
		return 7 // High risk - malformed certificates are suspicious
	case ErrorTypeUntrustedCA:
		return 6 // Medium-high risk - could be man-in-the-middle
	case ErrorTypeHostnameMismatch:
		return 5 // Medium risk - could be configuration error or attack
	case ErrorTypeSelfSigned:
		return 4 // Medium-low risk - common in development
	case ErrorTypeChainIncomplete:
		return 3 // Low-medium risk - often configuration issue
	case ErrorTypeNotYetValid:
		return 2 // Low risk - usually time synchronization issue
	case ErrorTypeKeyUsage:
		return 3 // Low-medium risk - configuration issue
	default:
		return 5 // Medium risk for unknown errors
	}
}