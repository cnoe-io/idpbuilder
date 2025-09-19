package registry

import (
	"fmt"
	"log"
	"time"
)

// retryWithExponentialBackoff executes the given operation with exponential backoff retry logic
func retryWithExponentialBackoff(operation func() error, operationName, target string) error {
	const (
		maxRetries   = 3
		initialDelay = time.Second
		backoffFactor = 2.0
	)

	var lastErr error
	delay := initialDelay

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying %s for %s (attempt %d/%d) after %v delay",
				operationName, target, attempt+1, maxRetries, delay)
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * backoffFactor)
		}

		lastErr = operation()
		if lastErr == nil {
			if attempt > 0 {
				log.Printf("Successfully completed %s for %s after %d retries",
					operationName, target, attempt)
			}
			return nil
		}

		log.Printf("Failed %s for %s (attempt %d/%d): %v",
			operationName, target, attempt+1, maxRetries, lastErr)
	}

	return fmt.Errorf("failed %s for %s after %d attempts: %v",
		operationName, target, maxRetries, lastErr)
}
