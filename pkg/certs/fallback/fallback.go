package fallback

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// FallbackHandler manages certificate validation fallback strategies
type FallbackHandler struct {
	recommendations *RecommendationEngine
	securityLogger  *SecurityLogger
	insecureManager *InsecureManager
	config          *FallbackConfig
}

// FallbackConfig holds configuration for fallback behavior
type FallbackConfig struct {
	// Enable automatic fallback strategies
	EnableAutoFallback bool

	// Maximum time to spend on fallback attempts
	FallbackTimeout time.Duration

	// Whether to allow insecure mode as last resort
	AllowInsecureMode bool

	// Registry-specific overrides
	RegistryOverrides map[string]*RegistryFallbackConfig
}

// RegistryFallbackConfig holds registry-specific fallback settings
type RegistryFallbackConfig struct {
	AllowInsecure     bool
	CustomCA          string
	SkipVerification  bool
	MaxRetryAttempts  int
}

// FallbackResult represents the outcome of a fallback attempt
type FallbackResult struct {
	Strategy       FallbackStrategy
	Success        bool
	Transport      http.RoundTripper
	SecurityRisk   SecurityRiskLevel
	Message        string
	Recommendations []string
}

// FallbackStrategy represents different fallback approaches
type FallbackStrategy int

const (
	StrategyNone FallbackStrategy = iota
	StrategyRetryWithSystemCA
	StrategyRetryWithoutSNI
	StrategyRetryWithLowerTLS
	StrategyAcceptSelfSigned
	StrategyInsecureMode
)

// SecurityRiskLevel indicates the security implications
type SecurityRiskLevel int

const (
	RiskNone SecurityRiskLevel = iota
	RiskLow
	RiskMedium
	RiskHigh
	RiskCritical
)

// NewFallbackHandler creates a new fallback handler
func NewFallbackHandler(config *FallbackConfig) (*FallbackHandler, error) {
	if config == nil {
		config = DefaultFallbackConfig()
	}

	securityLogger, err := NewSecurityLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize security logger: %w", err)
	}

	recommendations := NewRecommendationEngine()
	insecureManager := NewInsecureManager(securityLogger)

	return &FallbackHandler{
		recommendations: recommendations,
		securityLogger:  securityLogger,
		insecureManager: insecureManager,
		config:          config,
	}, nil
}

// HandleCertificateError processes certificate validation errors and attempts fallback
func (h *FallbackHandler) HandleCertificateError(ctx context.Context, registry string, originalErr error) (*FallbackResult, error) {
	// Log the initial certificate failure
	h.securityLogger.LogCertificateFailure(registry, originalErr)

	// Analyze the error to determine appropriate fallback strategies
	strategies := h.analyzeError(originalErr)

	// Generate recommendations for the user
	recommendations := h.recommendations.GenerateRecommendations(registry, originalErr)

	// Attempt each fallback strategy in order of preference
	for _, strategy := range strategies {
		result, err := h.attemptFallback(ctx, registry, strategy, originalErr)
		if err != nil {
			h.securityLogger.LogFallbackAttempt(registry, strategy, false, err.Error())
			continue
		}

		if result.Success {
			result.Recommendations = recommendations
			h.securityLogger.LogFallbackSuccess(registry, strategy, result.SecurityRisk)
			return result, nil
		}

		h.securityLogger.LogFallbackAttempt(registry, strategy, false, result.Message)
	}

	// If all strategies failed, return comprehensive failure information
	return &FallbackResult{
		Strategy:        StrategyNone,
		Success:         false,
		SecurityRisk:    RiskNone,
		Message:         "All fallback strategies failed",
		Recommendations: recommendations,
	}, fmt.Errorf("certificate validation failed and all fallback strategies exhausted")
}

// analyzeError examines the certificate error to determine appropriate fallback strategies
func (h *FallbackHandler) analyzeError(err error) []FallbackStrategy {
	var strategies []FallbackStrategy

	errorStr := strings.ToLower(err.Error())

	// Check for common certificate errors and suggest appropriate fallbacks
	switch {
	case strings.Contains(errorStr, "unknown authority"):
		strategies = append(strategies, StrategyRetryWithSystemCA, StrategyAcceptSelfSigned)
	case strings.Contains(errorStr, "hostname"):
		strategies = append(strategies, StrategyRetryWithoutSNI)
	case strings.Contains(errorStr, "tls"):
		strategies = append(strategies, StrategyRetryWithLowerTLS)
	case strings.Contains(errorStr, "expired") || strings.Contains(errorStr, "not yet valid"):
		// For time-based errors, only insecure mode helps
		if h.config.AllowInsecureMode {
			strategies = append(strategies, StrategyInsecureMode)
		}
	default:
		// Generic fallback sequence
		strategies = append(strategies,
			StrategyRetryWithSystemCA,
			StrategyRetryWithoutSNI,
			StrategyRetryWithLowerTLS,
		)
	}

	// Always add insecure mode as last resort if allowed
	if h.config.AllowInsecureMode {
		strategies = append(strategies, StrategyInsecureMode)
	}

	return strategies
}

// attemptFallback tries a specific fallback strategy
func (h *FallbackHandler) attemptFallback(ctx context.Context, registry string, strategy FallbackStrategy, originalErr error) (*FallbackResult, error) {
	switch strategy {
	case StrategyRetryWithSystemCA:
		return h.retryWithSystemCA(ctx, registry)
	case StrategyRetryWithoutSNI:
		return h.retryWithoutSNI(ctx, registry)
	case StrategyRetryWithLowerTLS:
		return h.retryWithLowerTLS(ctx, registry)
	case StrategyAcceptSelfSigned:
		return h.acceptSelfSigned(ctx, registry)
	case StrategyInsecureMode:
		return h.useInsecureMode(ctx, registry)
	default:
		return nil, fmt.Errorf("unknown fallback strategy: %v", strategy)
	}
}

// retryWithSystemCA attempts connection using only system CA certificates
func (h *FallbackHandler) retryWithSystemCA(ctx context.Context, registry string) (*FallbackResult, error) {
	systemPool, err := x509.SystemCertPool()
	if err != nil {
		return &FallbackResult{
			Strategy:     StrategyRetryWithSystemCA,
			Success:      false,
			SecurityRisk: RiskNone,
			Message:      "System CA pool unavailable",
		}, nil
	}

	tlsConfig := &tls.Config{
		RootCAs:    systemPool,
		MinVersion: tls.VersionTLS12,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// Test the connection
	if err := h.testConnection(ctx, registry, transport); err != nil {
		return &FallbackResult{
			Strategy:     StrategyRetryWithSystemCA,
			Success:      false,
			SecurityRisk: RiskNone,
			Message:      fmt.Sprintf("System CA retry failed: %v", err),
		}, nil
	}

	return &FallbackResult{
		Strategy:     StrategyRetryWithSystemCA,
		Success:      true,
		Transport:    transport,
		SecurityRisk: RiskLow,
		Message:      "Successfully connected using system CA certificates",
	}, nil
}

// retryWithoutSNI attempts connection without Server Name Indication
func (h *FallbackHandler) retryWithoutSNI(ctx context.Context, registry string) (*FallbackResult, error) {
	tlsConfig := &tls.Config{
		ServerName: "", // Disable SNI
		MinVersion: tls.VersionTLS12,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	if err := h.testConnection(ctx, registry, transport); err != nil {
		return &FallbackResult{
			Strategy:     StrategyRetryWithoutSNI,
			Success:      false,
			SecurityRisk: RiskNone,
			Message:      fmt.Sprintf("No-SNI retry failed: %v", err),
		}, nil
	}

	return &FallbackResult{
		Strategy:     StrategyRetryWithoutSNI,
		Success:      true,
		Transport:    transport,
		SecurityRisk: RiskMedium,
		Message:      "Successfully connected without SNI (hostname verification disabled)",
	}, nil
}

// retryWithLowerTLS attempts connection with lower TLS version
func (h *FallbackHandler) retryWithLowerTLS(ctx context.Context, registry string) (*FallbackResult, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS10, // Allow older TLS versions
		MaxVersion: tls.VersionTLS13,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	if err := h.testConnection(ctx, registry, transport); err != nil {
		return &FallbackResult{
			Strategy:     StrategyRetryWithLowerTLS,
			Success:      false,
			SecurityRisk: RiskNone,
			Message:      fmt.Sprintf("Lower TLS retry failed: %v", err),
		}, nil
	}

	return &FallbackResult{
		Strategy:     StrategyRetryWithLowerTLS,
		Success:      true,
		Transport:    transport,
		SecurityRisk: RiskMedium,
		Message:      "Successfully connected using lower TLS version (security reduced)",
	}, nil
}

// acceptSelfSigned accepts self-signed certificates with explicit warning
func (h *FallbackHandler) acceptSelfSigned(ctx context.Context, registry string) (*FallbackResult, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	if err := h.testConnection(ctx, registry, transport); err != nil {
		return &FallbackResult{
			Strategy:     StrategyAcceptSelfSigned,
			Success:      false,
			SecurityRisk: RiskNone,
			Message:      fmt.Sprintf("Self-signed acceptance failed: %v", err),
		}, nil
	}

	// Log security decision
	h.securityLogger.LogSecurityDecision("ACCEPT_SELF_SIGNED", registry,
		"Accepting self-signed certificate for registry connection")

	return &FallbackResult{
		Strategy:     StrategyAcceptSelfSigned,
		Success:      true,
		Transport:    transport,
		SecurityRisk: RiskHigh,
		Message:      "Connected accepting self-signed certificates (HIGH SECURITY RISK)",
	}, nil
}

// useInsecureMode completely disables certificate verification
func (h *FallbackHandler) useInsecureMode(ctx context.Context, registry string) (*FallbackResult, error) {
	// Insecure mode requires explicit user confirmation
	confirmed, err := h.insecureManager.ConfirmInsecureMode(registry)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm insecure mode: %w", err)
	}

	if !confirmed {
		return &FallbackResult{
			Strategy:     StrategyInsecureMode,
			Success:      false,
			SecurityRisk: RiskNone,
			Message:      "Insecure mode rejected by user",
		}, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &FallbackResult{
		Strategy:     StrategyInsecureMode,
		Success:      true,
		Transport:    transport,
		SecurityRisk: RiskCritical,
		Message:      "Connected in INSECURE mode - ALL certificate validation disabled",
	}, nil
}

// testConnection performs a basic connectivity test with the given transport
func (h *FallbackHandler) testConnection(ctx context.Context, registry string, transport http.RoundTripper) error {
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	// Construct test URL
	testURL := fmt.Sprintf("https://%s/v2/", registry)

	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	defer resp.Body.Close()

	// Accept any HTTP response as successful connection
	// We're only testing certificate/TLS connectivity
	return nil
}

// DefaultFallbackConfig returns a secure default configuration
func DefaultFallbackConfig() *FallbackConfig {
	return &FallbackConfig{
		EnableAutoFallback: true,
		FallbackTimeout:    30 * time.Second,
		AllowInsecureMode:  false, // Insecure mode disabled by default
		RegistryOverrides:  make(map[string]*RegistryFallbackConfig),
	}
}

// String returns a human-readable description of the fallback strategy
func (s FallbackStrategy) String() string {
	switch s {
	case StrategyRetryWithSystemCA:
		return "Retry with system CA certificates"
	case StrategyRetryWithoutSNI:
		return "Retry without SNI (hostname verification)"
	case StrategyRetryWithLowerTLS:
		return "Retry with lower TLS version"
	case StrategyAcceptSelfSigned:
		return "Accept self-signed certificates"
	case StrategyInsecureMode:
		return "Use insecure mode (no certificate validation)"
	default:
		return "Unknown strategy"
	}
}

// String returns a human-readable description of the security risk level
func (r SecurityRiskLevel) String() string {
	switch r {
	case RiskNone:
		return "No additional security risk"
	case RiskLow:
		return "Low security risk"
	case RiskMedium:
		return "Medium security risk"
	case RiskHigh:
		return "High security risk"
	case RiskCritical:
		return "Critical security risk"
	default:
		return "Unknown risk level"
	}
}