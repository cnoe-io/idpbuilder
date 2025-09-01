// Package fallback provides certificate error detection and fallback handling
// for when TLS operations fail due to certificate issues.
package fallback

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"
)

// DefaultFallbackHandler is the default implementation of the FallbackHandler interface
type DefaultFallbackHandler struct {
	// detector is the certificate error detector interface
	detector CertErrorDetector
	
	// strategy is the current fallback strategy
	strategy *FallbackStrategy
	
	// logger is the security logger interface (implementation in Split 002)
	logger SecurityLogger
	
	// prompter is the user prompter interface
	prompter UserPrompter
	
	// trustedHosts contains the list of trusted hostnames
	trustedHosts map[string]bool
	
	// tlsConfigCache caches TLS configurations for performance
	tlsConfigCache map[string]*tls.Config
	
	// decisionCache caches user decisions to avoid repeated prompts
	decisionCache map[string]*FallbackDecision
	
	// mutex protects concurrent access to internal maps
	mutex sync.RWMutex
}

// NewDefaultFallbackHandler creates a new DefaultFallbackHandler with secure defaults
func NewDefaultFallbackHandler(detector CertErrorDetector, logger SecurityLogger, prompter UserPrompter) *DefaultFallbackHandler {
	return &DefaultFallbackHandler{
		detector:       detector,
		logger:         logger,
		prompter:       prompter,
		strategy:       NewSecureStrategy(),
		trustedHosts:   make(map[string]bool),
		tlsConfigCache: make(map[string]*tls.Config),
		decisionCache:  make(map[string]*FallbackDecision),
	}
}

// NewSecureStrategy returns a secure fallback strategy that denies all certificate errors by default
func NewSecureStrategy() *FallbackStrategy {
	return &FallbackStrategy{
		Mode:              ModeSecure,
		Rules:             map[CertErrorType]FallbackAction{
			ErrorTypeExpired:          ActionDeny,
			ErrorTypeNotYetValid:      ActionDeny,
			ErrorTypeHostnameMismatch: ActionDeny,
			ErrorTypeUntrustedCA:      ActionDeny,
			ErrorTypeSelfSigned:       ActionDeny,
			ErrorTypeBadCertificate:   ActionDeny,
			ErrorTypeChainIncomplete:  ActionDeny,
			ErrorTypeRevoked:          ActionDeny,
			ErrorTypeKeyUsage:         ActionDeny,
		},
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

// NewDevelopmentStrategy returns a development strategy that allows certificate errors with logging
func NewDevelopmentStrategy() *FallbackStrategy {
	return &FallbackStrategy{
		Mode:              ModeDevelopment,
		Rules:             map[CertErrorType]FallbackAction{
			ErrorTypeExpired:          ActionLog,
			ErrorTypeNotYetValid:      ActionLog,
			ErrorTypeHostnameMismatch: ActionLog,
			ErrorTypeUntrustedCA:      ActionLog,
			ErrorTypeSelfSigned:       ActionLog,
			ErrorTypeBadCertificate:   ActionDeny, // Still deny malformed certs
			ErrorTypeChainIncomplete:  ActionLog,
			ErrorTypeRevoked:          ActionDeny, // Still deny revoked certs
			ErrorTypeKeyUsage:         ActionLog,
		},
		Hostnames:         make(map[string]FallbackAction),
		InsecureHosts:     make(map[string]bool),
		PromptEnabled:     false,
		LoggingEnabled:    true,
		MaxRetries:        2,
		RetryDelay:        1 * time.Second,
		TrustNewCerts:     false,
		RememberDecisions: false,
		DecisionTimeout:   30 * time.Second,
	}
}

// NewInteractiveStrategy returns an interactive strategy that prompts users for decisions
func NewInteractiveStrategy() *FallbackStrategy {
	return &FallbackStrategy{
		Mode:              ModeInteractive,
		Rules:             map[CertErrorType]FallbackAction{
			ErrorTypeExpired:          ActionPrompt,
			ErrorTypeNotYetValid:      ActionPrompt,
			ErrorTypeHostnameMismatch: ActionPrompt,
			ErrorTypeUntrustedCA:      ActionPrompt,
			ErrorTypeSelfSigned:       ActionPrompt,
			ErrorTypeBadCertificate:   ActionDeny, // Never prompt for malformed certs
			ErrorTypeChainIncomplete:  ActionPrompt,
			ErrorTypeRevoked:          ActionDeny, // Never prompt for revoked certs
			ErrorTypeKeyUsage:         ActionPrompt,
		},
		Hostnames:         make(map[string]FallbackAction),
		InsecureHosts:     make(map[string]bool),
		PromptEnabled:     true,
		LoggingEnabled:    true,
		MaxRetries:        1,
		RetryDelay:        500 * time.Millisecond,
		TrustNewCerts:     false,
		RememberDecisions: true,
		DecisionTimeout:   60 * time.Second,
	}
}

// HandleError processes a certificate error and returns a fallback decision
func (h *DefaultFallbackHandler) HandleError(ctx context.Context, errorDetails *ErrorDetails) (*FallbackDecision, error) {
	if errorDetails == nil {
		return nil, fmt.Errorf("error details cannot be nil")
	}

	// Check for cached decision first
	decisionKey := h.createDecisionKey(errorDetails)
	if h.strategy.RememberDecisions {
		h.mutex.RLock()
		if cachedDecision, exists := h.decisionCache[decisionKey]; exists {
			h.mutex.RUnlock()
			return cachedDecision, nil
		}
		h.mutex.RUnlock()
	}

	// Determine the appropriate action
	action := h.determineAction(errorDetails)

	// Handle user prompt if needed
	if action == ActionPrompt && h.prompter != nil {
		promptCtx, cancel := context.WithTimeout(ctx, h.strategy.DecisionTimeout)
		defer cancel()
		
		userAction, err := h.prompter.PromptForDecision(promptCtx, errorDetails)
		if err != nil {
			// If user prompt fails, fall back to deny for security
			action = ActionDeny
		} else {
			action = userAction
		}
	}

	// Create the decision
	decision := h.createDecision(action, errorDetails)

	// Cache the decision if configured
	if h.strategy.RememberDecisions && decision.Permanent {
		h.rememberDecision(decisionKey, decision)
	}

	return decision, nil
}

// GetStrategy returns the current fallback strategy
func (h *DefaultFallbackHandler) GetStrategy() *FallbackStrategy {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	// Return a copy to prevent external modification
	strategyCopy := *h.strategy
	return &strategyCopy
}

// UpdateStrategy updates the fallback strategy
func (h *DefaultFallbackHandler) UpdateStrategy(strategy *FallbackStrategy) error {
	if strategy == nil {
		return fmt.Errorf("strategy cannot be nil")
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	h.strategy = strategy
	
	// Clear caches when strategy changes
	h.tlsConfigCache = make(map[string]*tls.Config)
	if !strategy.RememberDecisions {
		h.decisionCache = make(map[string]*FallbackDecision)
	}
	
	return nil
}

// IsHostTrusted checks if a hostname is in the trusted hosts list
func (h *DefaultFallbackHandler) IsHostTrusted(hostname string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	// Check explicit trusted hosts
	if trusted, exists := h.trustedHosts[hostname]; exists && trusted {
		return true
	}
	
	// Check strategy insecure hosts
	if insecure, exists := h.strategy.InsecureHosts[hostname]; exists && insecure {
		return true
	}
	
	return false
}

// AddTrustedHost adds a hostname to the trusted hosts list
func (h *DefaultFallbackHandler) AddTrustedHost(hostname string) error {
	if hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	h.trustedHosts[hostname] = true
	
	// Clear TLS config cache for this hostname
	delete(h.tlsConfigCache, hostname)
	
	return nil
}

// RemoveTrustedHost removes a hostname from the trusted hosts list
func (h *DefaultFallbackHandler) RemoveTrustedHost(hostname string) error {
	if hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	delete(h.trustedHosts, hostname)
	
	// Clear TLS config cache for this hostname
	delete(h.tlsConfigCache, hostname)
	
	return nil
}

// CreateTLSConfig creates a TLS configuration based on the fallback decision
func (h *DefaultFallbackHandler) CreateTLSConfig(decision *FallbackDecision) (*tls.Config, error) {
	if decision == nil {
		return nil, fmt.Errorf("decision cannot be nil")
	}

	// Use cached config if available
	configKey := fmt.Sprintf("%s-%d", decision.Action.String(), decision.SecurityRisk)
	h.mutex.RLock()
	if cachedConfig, exists := h.tlsConfigCache[configKey]; exists {
		h.mutex.RUnlock()
		return cachedConfig, nil
	}
	h.mutex.RUnlock()

	// Create new TLS config based on action
	config := &tls.Config{}
	
	switch decision.Action {
	case ActionAccept, ActionLog:
		config.InsecureSkipVerify = true
		
	case ActionDeny:
		config.InsecureSkipVerify = false
		
	case ActionRetry:
		// Use default secure settings for retry
		config.InsecureSkipVerify = false
		
	default:
		return nil, fmt.Errorf("unsupported action: %v", decision.Action)
	}
	
	// Apply timeouts
	if decision.Timeout > 0 {
		// TLS config doesn't directly support timeouts, but we can store it in metadata
		if decision.Metadata == nil {
			decision.Metadata = make(map[string]interface{})
		}
		decision.Metadata["timeout"] = decision.Timeout
	}

	// Cache the config
	h.mutex.Lock()
	h.tlsConfigCache[configKey] = config
	h.mutex.Unlock()
	
	return config, nil
}

// LogSecurityDecision logs a security decision (stub - full implementation in Split 002)
func (h *DefaultFallbackHandler) LogSecurityDecision(decision *FallbackDecision, errorDetails *ErrorDetails) {
	if h.logger != nil {
		h.logger.LogFallbackDecision(decision, errorDetails)
	}
	// Stub implementation - full logging will be implemented in Split 002
}

// SetUserPrompter sets the user prompter interface
func (h *DefaultFallbackHandler) SetUserPrompter(prompter UserPrompter) {
	h.mutex.Lock()
	h.prompter = prompter
	h.mutex.Unlock()
}

// SetSecurityLogger sets the security logger interface
func (h *DefaultFallbackHandler) SetSecurityLogger(logger SecurityLogger) {
	h.mutex.Lock()
	h.logger = logger
	h.mutex.Unlock()
}

// determineAction determines the appropriate action based on the error details and strategy
func (h *DefaultFallbackHandler) determineAction(errorDetails *ErrorDetails) FallbackAction {
	// Check hostname-specific rules first
	if action, exists := h.strategy.Hostnames[errorDetails.Hostname]; exists {
		return action
	}
	
	// Check if hostname is in trusted/insecure hosts
	if h.IsHostTrusted(errorDetails.Hostname) {
		return ActionAccept
	}
	
	// Apply error type specific rules
	if action, exists := h.strategy.Rules[errorDetails.Type]; exists {
		return action
	}
	
	// Fall back to mode defaults
	switch h.strategy.Mode {
	case ModeSecure:
		return ActionDeny
	case ModePermissive:
		return ActionLog
	case ModeDevelopment:
		return ActionAccept
	case ModeInteractive:
		return ActionPrompt
	case ModeCustom:
		// For custom mode, deny by default if no specific rule
		return ActionDeny
	default:
		return ActionDeny
	}
}

// createDecision creates a FallbackDecision based on the action and error details
func (h *DefaultFallbackHandler) createDecision(action FallbackAction, errorDetails *ErrorDetails) *FallbackDecision {
	decision := &FallbackDecision{
		Action:       action,
		Reason:       fmt.Sprintf("Certificate error %s for host %s", errorDetails.Type.String(), errorDetails.Hostname),
		Timeout:      30 * time.Second,
		Permanent:    h.strategy.RememberDecisions,
		RetryCount:   h.strategy.MaxRetries,
		SecurityRisk: h.assessSecurityRisk(action, errorDetails),
		Metadata:     map[string]interface{}{
			"errorType": errorDetails.Type.String(),
			"hostname":  errorDetails.Hostname,
			"timestamp": errorDetails.Timestamp,
		},
	}
	return decision
}

// createDecisionKey creates a unique key for caching decisions
func (h *DefaultFallbackHandler) createDecisionKey(errorDetails *ErrorDetails) string {
	return fmt.Sprintf("%s:%s:%s", errorDetails.Hostname, errorDetails.Type.String(), errorDetails.Message)
}

// rememberDecision caches a decision for future use
func (h *DefaultFallbackHandler) rememberDecision(key string, decision *FallbackDecision) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	// Only cache permanent decisions
	if decision.Permanent {
		h.decisionCache[key] = decision
	}
}


// assessSecurityRisk evaluates the security risk level for a given action and error
func (h *DefaultFallbackHandler) assessSecurityRisk(action FallbackAction, errorDetails *ErrorDetails) int {
	if action == ActionDeny {
		return 0 // No risk if we deny
	}
	
	// Base risk by error type  
	risk := map[CertErrorType]int{
		ErrorTypeRevoked: 10, ErrorTypeBadCertificate: 9, ErrorTypeUntrustedCA: 8,
		ErrorTypeSelfSigned: 7, ErrorTypeHostnameMismatch: 6, ErrorTypeChainIncomplete: 5,
		ErrorTypeExpired: 4, ErrorTypeNotYetValid: 3, ErrorTypeKeyUsage: 2,
	}
	
	baseRisk, exists := risk[errorDetails.Type]
	if !exists {
		baseRisk = 5
	}
	
	// Adjust by action
	switch action {
	case ActionPrompt: return baseRisk / 2
	case ActionRetry: return baseRisk / 3
	default: return baseRisk
	}
}