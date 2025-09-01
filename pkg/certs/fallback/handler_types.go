// Package fallback provides certificate error detection and fallback handling
// for when TLS operations fail due to certificate issues.
package fallback

import (
	"context"
	"crypto/tls"
	"crypto/x509"
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