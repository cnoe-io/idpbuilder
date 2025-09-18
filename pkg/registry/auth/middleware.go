package auth

import (
	"fmt"
	"net/http"
)

// Transport wraps an http.RoundTripper with authentication
type Transport struct {
	Base          http.RoundTripper
	Authenticator Authenticator
}

// NewTransport creates a new authenticated transport
func NewTransport(base http.RoundTripper, auth Authenticator) *Transport {
	if base == nil {
		base = http.DefaultTransport
	}

	return &Transport{
		Base:          base,
		Authenticator: auth,
	}
}

// RoundTrip implements http.RoundTripper
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	clonedReq := req.Clone(req.Context())

	// Apply authentication
	if t.Authenticator != nil {
		// Check if auth needs refresh
		if !t.Authenticator.IsValid() {
			if err := t.Authenticator.Refresh(req.Context()); err != nil {
				return nil, fmt.Errorf("auth refresh failed: %w", err)
			}
		}

		if err := t.Authenticator.Authenticate(req.Context(), clonedReq); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Execute the request
	resp, err := t.Base.RoundTrip(clonedReq)
	if err != nil {
		return nil, err
	}

	// Handle 401 Unauthorized by refreshing auth and retrying once
	if resp.StatusCode == http.StatusUnauthorized && t.Authenticator != nil {
		resp.Body.Close()

		if err := t.Authenticator.Refresh(req.Context()); err != nil {
			return nil, fmt.Errorf("auth refresh after 401 failed: %w", err)
		}

		// Retry with refreshed auth
		retryReq := req.Clone(req.Context())
		if err := t.Authenticator.Authenticate(req.Context(), retryReq); err != nil {
			return nil, fmt.Errorf("re-authentication failed: %w", err)
		}

		return t.Base.RoundTrip(retryReq)
	}

	return resp, nil
}