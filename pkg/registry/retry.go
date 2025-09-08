package registry

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// retryConfig holds configuration for retry operations
type retryConfig struct {
	maxAttempts   int
	baseDelay     time.Duration
	maxDelay      time.Duration
	backoffFactor float64
}

// getDefaultRetryConfig returns the default retry configuration
func getDefaultRetryConfig() retryConfig {
	return retryConfig{
		maxAttempts:   3,
		baseDelay:     500 * time.Millisecond,
		maxDelay:      5 * time.Second,
		backoffFactor: 2.0,
	}
}

// retryWithExponentialBackoff executes an operation with exponential backoff retry logic.
// Uses a maximum of 3 attempts with exponential delay between retries.
// Returns the last error if all attempts fail.
func retryWithExponentialBackoff(operation func() error, operationName, context string) error {
	config := getDefaultRetryConfig()
	var lastErr error
	
	for attempt := 1; attempt <= config.maxAttempts; attempt++ {
		// Execute the operation
		err := operation()
		if err == nil {
			// Success
			if attempt > 1 {
				log.Printf("Operation %s succeeded on attempt %d/%d", operationName, attempt, config.maxAttempts)
			}
			return nil
		}
		
		lastErr = err
		
		// Check if this is the last attempt
		if attempt == config.maxAttempts {
			log.Printf("Operation %s failed on final attempt %d/%d: %v", operationName, attempt, config.maxAttempts, err)
			break
		}
		
		// Check if error is retryable
		if !isRetryableError(err) {
			log.Printf("Operation %s failed with non-retryable error: %v", operationName, err)
			break
		}
		
		// Calculate delay for this attempt
		delay := calculateDelay(attempt, config)
		
		log.Printf("Operation %s failed on attempt %d/%d, retrying in %v: %v", 
			operationName, attempt, config.maxAttempts, delay, err)
		
		// Wait before retry
		time.Sleep(delay)
	}
	
	return fmt.Errorf("operation %s failed after %d attempts, last error: %v", 
		operationName, config.maxAttempts, lastErr)
}

// calculateDelay calculates the delay for a given retry attempt using exponential backoff
func calculateDelay(attempt int, config retryConfig) time.Duration {
	// Calculate exponential delay: baseDelay * (backoffFactor ^ (attempt - 1))
	delay := float64(config.baseDelay) * pow(config.backoffFactor, float64(attempt-1))
	
	// Apply maximum delay limit
	if delay > float64(config.maxDelay) {
		delay = float64(config.maxDelay)
	}
	
	return time.Duration(delay)
}

// pow calculates base^exp for positive values (simple implementation for our use case)
func pow(base, exp float64) float64 {
	if exp == 0 {
		return 1
	}
	result := base
	for i := 1; i < int(exp); i++ {
		result *= base
	}
	return result
}

// isRetryableError determines if an error should trigger a retry attempt.
// Returns true for transient errors that may succeed on retry.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	errorMsg := strings.ToLower(err.Error())
	
	// Network-related errors that are typically transient
	retryablePatterns := []string{
		"timeout",
		"connection reset",
		"connection refused", 
		"network unreachable",
		"temporary failure",
		"service unavailable",
		"internal server error",
		"bad gateway",
		"gateway timeout",
		"too many requests",
		"rate limit",
		"throttle",
	}
	
	for _, pattern := range retryablePatterns {
		if strings.Contains(errorMsg, pattern) {
			return true
		}
	}
	
	// Non-retryable errors
	nonRetryablePatterns := []string{
		"unauthorized",
		"forbidden",
		"not found",
		"bad request",
		"invalid",
		"malformed",
		"unsupported",
		"authentication",
	}
	
	for _, pattern := range nonRetryablePatterns {
		if strings.Contains(errorMsg, pattern) {
			return false
		}
	}
	
	// Default to retryable for unknown errors
	return true
}

// retryableOperation wraps an operation to make it compatible with retry logic
type retryableOperation struct {
	name    string
	context string
	execute func() error
}

// newRetryableOperation creates a new retryable operation wrapper
func newRetryableOperation(name, context string, operation func() error) *retryableOperation {
	return &retryableOperation{
		name:    name,
		context: context,
		execute: operation,
	}
}

// run executes the retryable operation with exponential backoff
func (r *retryableOperation) run() error {
	return retryWithExponentialBackoff(r.execute, r.name, r.context)
}

// withRetry is a convenience function to wrap and execute an operation with retry logic
func withRetry(operationName, context string, operation func() error) error {
	retryOp := newRetryableOperation(operationName, context, operation)
	return retryOp.run()
}