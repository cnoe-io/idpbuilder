package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

// mockRoundTripper implements http.RoundTripper for testing
type mockRoundTripper struct {
	response   *http.Response
	err        error
	requests   []*http.Request
	callCount  int
	statusCode int
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.callCount++
	m.requests = append(m.requests, req)

	if m.err != nil {
		return nil, m.err
	}

	statusCode := m.statusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	resp := &http.Response{
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("test response")),
		Request:    req,
	}

	if m.response != nil {
		return m.response, nil
	}

	return resp, nil
}

func (m *mockRoundTripper) reset() {
	m.callCount = 0
	m.requests = nil
}

// mockRetryRoundTripper for testing 401 retry logic
type mockRetryRoundTripper struct {
	callCount int
}

func (m *mockRetryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.callCount++
	if m.callCount == 1 {
		// First call returns 401
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("unauthorized")),
			Request:    req,
		}, nil
	}
	// Second call returns 200
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("success")),
		Request:    req,
	}, nil
}

// mockAlwaysUnauthorizedRoundTripper for testing multiple retry behavior
type mockAlwaysUnauthorizedRoundTripper struct {
	callCount int
}

func (m *mockAlwaysUnauthorizedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.callCount++
	// Always return 401 to test that we only retry once
	return &http.Response{
		StatusCode: http.StatusUnauthorized,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("unauthorized")),
		Request:    req,
	}, nil
}

// mockAuthenticator implements Authenticator for testing
type mockAuthenticator struct {
	authError   error
	refreshError error
	isValid     bool
	authCalls   int
	refreshCalls int
	validCalls  int
	shouldRefreshOnAuth bool
}

func (m *mockAuthenticator) Authenticate(ctx context.Context, req *http.Request) error {
	m.authCalls++
	if m.shouldRefreshOnAuth {
		req.Header.Set("Authorization", "Bearer refreshed-token")
	} else {
		req.Header.Set("Authorization", "Bearer test-token")
	}
	return m.authError
}

func (m *mockAuthenticator) Refresh(ctx context.Context) error {
	m.refreshCalls++
	if m.refreshError == nil {
		m.shouldRefreshOnAuth = true
	}
	return m.refreshError
}

func (m *mockAuthenticator) IsValid() bool {
	m.validCalls++
	return m.isValid
}

func (m *mockAuthenticator) reset() {
	m.authCalls = 0
	m.refreshCalls = 0
	m.validCalls = 0
	m.shouldRefreshOnAuth = false
}

// mockContextAwareAuthenticator for testing context handling
type mockContextAwareAuthenticator struct {
	isValid bool
}

func (m *mockContextAwareAuthenticator) Authenticate(ctx context.Context, req *http.Request) error {
	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		req.Header.Set("Authorization", "Bearer test-token")
		return nil
	}
}

func (m *mockContextAwareAuthenticator) Refresh(ctx context.Context) error {
	return nil
}

func (m *mockContextAwareAuthenticator) IsValid() bool {
	return m.isValid
}

func TestNewTransport(t *testing.T) {
	tests := []struct {
		name string
		base http.RoundTripper
		auth Authenticator
	}{
		{
			name: "with custom base transport",
			base: &mockRoundTripper{},
			auth: &mockAuthenticator{isValid: true},
		},
		{
			name: "with nil base transport uses default",
			base: nil,
			auth: &mockAuthenticator{isValid: true},
		},
		{
			name: "with nil authenticator",
			base: &mockRoundTripper{},
			auth: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewTransport(tt.base, tt.auth)

			if transport == nil {
				t.Fatal("NewTransport() returned nil")
			}

			if tt.base == nil && transport.Base != http.DefaultTransport {
				t.Error("NewTransport() should use http.DefaultTransport when base is nil")
			}

			if tt.base != nil && transport.Base != tt.base {
				t.Error("NewTransport() did not set custom base transport correctly")
			}

			if transport.Authenticator != tt.auth {
				t.Error("NewTransport() did not set authenticator correctly")
			}
		})
	}
}

func TestTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name           string
		auth           *mockAuthenticator
		base           *mockRoundTripper
		expectAuthCall bool
		expectError    bool
	}{
		{
			name: "successful request with valid auth",
			auth: &mockAuthenticator{
				isValid: true,
			},
			base: &mockRoundTripper{
				statusCode: http.StatusOK,
			},
			expectAuthCall: true,
		},
		{
			name: "successful request with no authenticator",
			auth: nil,
			base: &mockRoundTripper{
				statusCode: http.StatusOK,
			},
			expectAuthCall: false,
		},
		{
			name: "authentication error",
			auth: &mockAuthenticator{
				isValid:   true,
				authError: fmt.Errorf("auth failed"),
			},
			base:        &mockRoundTripper{},
			expectError: true,
		},
		{
			name: "base transport error",
			auth: &mockAuthenticator{
				isValid: true,
			},
			base: &mockRoundTripper{
				err: fmt.Errorf("transport error"),
			},
			expectAuthCall: true,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var auth Authenticator
			if tt.auth != nil {
				auth = tt.auth
			}
			transport := NewTransport(tt.base, auth)

			req, err := http.NewRequest("GET", "http://example.com", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if tt.auth != nil {
				tt.auth.reset()
			}
			if tt.base != nil {
				tt.base.reset()
			}

			resp, err := transport.RoundTrip(req)

			if tt.expectError {
				if err == nil {
					t.Errorf("Transport.RoundTrip() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Transport.RoundTrip() unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Fatal("Transport.RoundTrip() returned nil response")
			}

			// Verify authentication was called if expected
			if tt.expectAuthCall && tt.auth != nil {
				if tt.auth.authCalls == 0 {
					t.Error("Expected authentication call, but none was made")
				}
			}

			// Verify base transport was called
			if tt.base != nil && tt.base.callCount == 0 {
				t.Error("Expected base transport call, but none was made")
			}

			// Close response body
			if resp.Body != nil {
				resp.Body.Close()
			}
		})
	}
}

func TestTransport_RoundTrip_RefreshOnInvalid(t *testing.T) {
	auth := &mockAuthenticator{
		isValid: false, // Auth is initially invalid
	}

	base := &mockRoundTripper{
		statusCode: http.StatusOK,
	}

	transport := NewTransport(base, auth)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Errorf("Transport.RoundTrip() error = %v", err)
	}

	// Verify that refresh was called due to invalid auth
	if auth.refreshCalls != 1 {
		t.Errorf("Expected 1 refresh call, got %d", auth.refreshCalls)
	}

	// Verify that authentication was called
	if auth.authCalls != 1 {
		t.Errorf("Expected 1 auth call, got %d", auth.authCalls)
	}

	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
}

func TestTransport_RoundTrip_401Retry(t *testing.T) {
	auth := &mockAuthenticator{
		isValid: true,
	}

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create custom round tripper for 401 retry test
	customBase := &mockRetryRoundTripper{}
	transport := NewTransport(customBase, auth)

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Errorf("Transport.RoundTrip() error = %v", err)
	}

	// Verify retry occurred
	if customBase.callCount != 2 {
		t.Errorf("Expected 2 calls to base transport (original + retry), got %d", customBase.callCount)
	}

	// Verify refresh was called after 401
	if auth.refreshCalls != 1 {
		t.Errorf("Expected 1 refresh call after 401, got %d", auth.refreshCalls)
	}

	// Verify authentication was called twice (original + retry)
	if auth.authCalls != 2 {
		t.Errorf("Expected 2 auth calls (original + retry), got %d", auth.authCalls)
	}

	// Verify final response is successful
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected final status 200, got %d", resp.StatusCode)
	}

	if resp.Body != nil {
		resp.Body.Close()
	}
}

func TestTransport_RoundTrip_401RetryRefreshError(t *testing.T) {
	auth := &mockAuthenticator{
		isValid:      true,
		refreshError: fmt.Errorf("refresh failed"),
	}

	base := &mockRoundTripper{
		statusCode: http.StatusUnauthorized,
	}

	transport := NewTransport(base, auth)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)

	// Should get an error due to refresh failure
	if err == nil {
		t.Error("Transport.RoundTrip() expected error due to refresh failure, got nil")
	}

	if resp != nil {
		t.Error("Transport.RoundTrip() expected nil response on error, got response")
	}

	// Verify refresh was attempted
	if auth.refreshCalls == 0 {
		t.Error("Expected refresh call after 401, got none")
	}
}

func TestTransport_RoundTrip_RequestCloning(t *testing.T) {
	auth := &mockAuthenticator{
		isValid: true,
	}

	base := &mockRoundTripper{
		statusCode: http.StatusOK,
	}

	transport := NewTransport(base, auth)

	originalReq, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Add a header to the original request
	originalReq.Header.Set("X-Original", "true")

	resp, err := transport.RoundTrip(originalReq)
	if err != nil {
		t.Errorf("Transport.RoundTrip() error = %v", err)
	}

	// Verify original request was not modified (should not have auth header)
	if originalReq.Header.Get("Authorization") != "" {
		t.Error("Original request was modified with Authorization header")
	}

	// Verify the original header is still there
	if originalReq.Header.Get("X-Original") != "true" {
		t.Error("Original request header was lost")
	}

	// Verify that the request sent to base transport has auth header
	if len(base.requests) > 0 {
		sentReq := base.requests[0]
		if sentReq.Header.Get("Authorization") == "" {
			t.Error("Request sent to base transport missing Authorization header")
		}

		// Verify the original header was preserved in the cloned request
		if sentReq.Header.Get("X-Original") != "true" {
			t.Error("Cloned request missing original headers")
		}
	}

	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
}

func TestTransport_RoundTrip_NoAuth(t *testing.T) {
	base := &mockRoundTripper{
		statusCode: http.StatusOK,
	}

	transport := NewTransport(base, nil)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	base.reset()

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Errorf("Transport.RoundTrip() error = %v", err)
	}

	// Verify base transport was called
	if base.callCount != 1 {
		t.Errorf("Expected 1 call to base transport, got %d", base.callCount)
	}

	// Verify no Authorization header was added
	if len(base.requests) > 0 {
		sentReq := base.requests[0]
		if sentReq.Header.Get("Authorization") != "" {
			t.Error("Authorization header was added despite no authenticator")
		}
	}

	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
}

func TestTransport_RoundTrip_RefreshError(t *testing.T) {
	auth := &mockAuthenticator{
		isValid:      false,
		refreshError: fmt.Errorf("refresh failed"),
	}

	base := &mockRoundTripper{
		statusCode: http.StatusOK,
	}

	transport := NewTransport(base, auth)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)

	// Should get an error due to refresh failure
	if err == nil {
		t.Error("Transport.RoundTrip() expected error due to refresh failure, got nil")
	}

	expectedError := "auth refresh failed: refresh failed"
	if err.Error() != expectedError {
		t.Errorf("Transport.RoundTrip() error = %q, want %q", err.Error(), expectedError)
	}

	if resp != nil {
		t.Error("Transport.RoundTrip() expected nil response on error, got response")
	}

	// Verify base transport was not called due to auth failure
	if base.callCount != 0 {
		t.Error("Base transport should not be called when auth refresh fails")
	}
}

func TestTransport_Interface(t *testing.T) {
	var _ http.RoundTripper = (*Transport)(nil)
}

func TestTransport_RoundTrip_ContextCancellation(t *testing.T) {
	auth := &mockContextAwareAuthenticator{
		isValid: true,
	}

	base := &mockRoundTripper{
		statusCode: http.StatusOK,
	}

	transport := NewTransport(base, auth)

	// Create a request with a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req, err := http.NewRequestWithContext(ctx, "GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	_, err = transport.RoundTrip(req)

	// Should get context cancellation error
	if err == nil {
		t.Error("Transport.RoundTrip() expected context cancellation error, got nil")
	}
}

func TestTransport_RoundTrip_MultipleRetries(t *testing.T) {
	auth := &mockAuthenticator{
		isValid: true,
	}

	base := &mockAlwaysUnauthorizedRoundTripper{}
	transport := NewTransport(base, auth)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Errorf("Transport.RoundTrip() error = %v", err)
	}

	// Verify only 2 calls were made (original + 1 retry)
	if base.callCount != 2 {
		t.Errorf("Expected exactly 2 calls (original + 1 retry), got %d", base.callCount)
	}

	// Verify final response is still 401
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected final status 401, got %d", resp.StatusCode)
	}

	if resp.Body != nil {
		resp.Body.Close()
	}
}