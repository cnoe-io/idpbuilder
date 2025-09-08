package registry

import (
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"time"
)

type RetryPolicy struct {
	MaxRetries    int
	InitialDelay  time.Duration
	BackoffFactor float64
}

func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{3, time.Second, 2.0}
}

func (p *RetryPolicy) IsRetryableError(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout() || netErr.Temporary()
	}
	return false
}

func RetryWithPolicy(ctx context.Context, policy *RetryPolicy, operation func() error) error {
	if policy == nil {
		policy = DefaultRetryPolicy()
	}
	
	var lastErr error
	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		if err := operation(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		
		if attempt >= policy.MaxRetries || !policy.IsRetryableError(lastErr) {
			break
		}
		
		delay := time.Duration(float64(policy.InitialDelay) * math.Pow(policy.BackoffFactor, float64(attempt)))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return fmt.Errorf("failed after %d attempts: %w", policy.MaxRetries+1, lastErr)
}

type WithRetry struct {
	registry Registry
	policy   *RetryPolicy
}

func NewWithRetry(registry Registry, policy *RetryPolicy) *WithRetry {
	if policy == nil {
		policy = DefaultRetryPolicy()
	}
	return &WithRetry{registry, policy}
}

func (r *WithRetry) Push(ctx context.Context, image string, content io.Reader) error {
	return RetryWithPolicy(ctx, r.policy, func() error { return r.registry.Push(ctx, image, content) })
}

func (r *WithRetry) List(ctx context.Context) ([]string, error) {
	var result []string
	err := RetryWithPolicy(ctx, r.policy, func() error { 
		var e error
		result, e = r.registry.List(ctx)
		return e
	})
	return result, err
}

func (r *WithRetry) Exists(ctx context.Context, repository string) (bool, error) {
	var result bool
	err := RetryWithPolicy(ctx, r.policy, func() error { 
		var e error
		result, e = r.registry.Exists(ctx, repository)
		return e
	})
	return result, err
}

func (r *WithRetry) Delete(ctx context.Context, repository string) error {
	return RetryWithPolicy(ctx, r.policy, func() error { return r.registry.Delete(ctx, repository) })
}

func (r *WithRetry) Close() error { return r.registry.Close() }