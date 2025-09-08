package fallback

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// FallbackStrategy defines the interface for fallback mechanisms
type FallbackStrategy interface {
	Name() string
	Priority() int
	Execute(ctx context.Context, registry string) error
	ShouldRetry(err error) bool
}

// FallbackManager coordinates fallback strategies
type FallbackManager struct {
	strategies      []FallbackStrategy
	trustStore      TrustStoreManager
	insecureMode    bool
	maxRetries      int
	retryDelay      time.Duration
	warningCallback func(string)
}

// NewFallbackManager creates a new fallback manager
func NewFallbackManager(trustStore TrustStoreManager, opts ...Option) *FallbackManager {
	fm := &FallbackManager{
		trustStore:      trustStore,
		strategies:      make([]FallbackStrategy, 0),
		maxRetries:      3,
		retryDelay:      time.Second,
		warningCallback: defaultWarning,
	}

	// Apply options
	for _, opt := range opts {
		opt(fm)
	}

	// Initialize default strategies
	fm.initDefaultStrategies()
	return fm
}

// Option configures the FallbackManager
type Option func(*FallbackManager)

// WithInsecureMode enables insecure mode
func WithInsecureMode(insecure bool) Option {
	return func(fm *FallbackManager) {
		fm.insecureMode = insecure
	}
}

// WithMaxRetries sets maximum retry attempts
func WithMaxRetries(max int) Option {
	return func(fm *FallbackManager) {
		fm.maxRetries = max
	}
}

// WithRetryDelay sets the base retry delay
func WithRetryDelay(delay time.Duration) Option {
	return func(fm *FallbackManager) {
		fm.retryDelay = delay
	}
}

// WithWarningCallback sets a custom warning callback
func WithWarningCallback(callback func(string)) Option {
	return func(fm *FallbackManager) {
		fm.warningCallback = callback
	}
}

// AddStrategy adds a custom fallback strategy
func (fm *FallbackManager) AddStrategy(strategy FallbackStrategy) {
	fm.strategies = append(fm.strategies, strategy)
	// Re-sort after adding
	fm.sortStrategies()
}

// HandleValidationFailure processes certificate validation failures
func (fm *FallbackManager) HandleValidationFailure(ctx context.Context, registry string, err error) error {
	// Check if insecure mode is enabled
	if fm.insecureMode {
		fm.warningCallback(fmt.Sprintf("⚠️  INSECURE MODE: Bypassing certificate validation for %s", registry))
		return fm.trustStore.SetInsecure(registry, true)
	}

	// Try fallback strategies in order of priority
	for _, strategy := range fm.strategies {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := fm.executeWithRetry(ctx, strategy, registry); err == nil {
				fm.warningCallback(fmt.Sprintf("✅ Fallback strategy '%s' succeeded for %s", strategy.Name(), registry))
				return nil
			}
		}
	}

	return fmt.Errorf("all fallback strategies failed for %s: %w", registry, err)
}

// executeWithRetry executes a strategy with retry logic
func (fm *FallbackManager) executeWithRetry(ctx context.Context, strategy FallbackStrategy, registry string) error {
	var lastErr error

	for attempt := 0; attempt < fm.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := fm.retryDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		lastErr = strategy.Execute(ctx, registry)
		if lastErr == nil {
			return nil
		}

		if !strategy.ShouldRetry(lastErr) {
			break
		}

		fm.warningCallback(fmt.Sprintf("⚠️  Strategy '%s' attempt %d failed for %s: %v",
			strategy.Name(), attempt+1, registry, lastErr))
	}

	return lastErr
}

// initDefaultStrategies sets up the default fallback strategies
func (fm *FallbackManager) initDefaultStrategies() {
	fm.strategies = []FallbackStrategy{
		NewSystemCertStrategy(fm.trustStore),
		NewCachedCertStrategy(fm.trustStore),
		NewSelfSignedAcceptStrategy(fm.trustStore),
	}

	// Sort by priority
	fm.sortStrategies()
}

// sortStrategies sorts strategies by priority (lower number = higher priority)
func (fm *FallbackManager) sortStrategies() {
	sort.Slice(fm.strategies, func(i, j int) bool {
		return fm.strategies[i].Priority() < fm.strategies[j].Priority()
	})
}

// IsInsecureMode returns whether insecure mode is enabled
func (fm *FallbackManager) IsInsecureMode() bool {
	return fm.insecureMode
}

// SetInsecureMode enables or disables insecure mode
func (fm *FallbackManager) SetInsecureMode(enabled bool) {
	fm.insecureMode = enabled
	if enabled {
		fm.warningCallback("⚠️  GLOBAL INSECURE MODE ENABLED - Certificate validation disabled")
	}
}

// GetStrategies returns the list of configured strategies
func (fm *FallbackManager) GetStrategies() []FallbackStrategy {
	return fm.strategies
}

// ResetStrategies clears all strategies and reinitializes with defaults
func (fm *FallbackManager) ResetStrategies() {
	fm.strategies = make([]FallbackStrategy, 0)
	fm.initDefaultStrategies()
}

// defaultWarning is the default warning callback that prints to stderr
func defaultWarning(message string) {
	fmt.Println(message)
}

// ValidateConfiguration checks if the manager is properly configured
func (fm *FallbackManager) ValidateConfiguration() error {
	if fm.trustStore == nil {
		return fmt.Errorf("trust store manager is required")
	}

	if len(fm.strategies) == 0 {
		return fmt.Errorf("at least one fallback strategy must be configured")
	}

	if fm.maxRetries < 1 {
		return fmt.Errorf("max retries must be at least 1")
	}

	if fm.retryDelay < 0 {
		return fmt.Errorf("retry delay must be non-negative")
	}

	return nil
}
