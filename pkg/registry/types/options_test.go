package types

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestConnectionOptions(t *testing.T) {
	tests := []struct {
		name string
		opts *ConnectionOptions
	}{
		{
			name: "basic connection options",
			opts: &ConnectionOptions{
				UserAgent: "test-client/1.0",
				Debug:     true,
			},
		},
		{
			name: "connection options with TLS config",
			opts: &ConnectionOptions{
				TLSConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
				UserAgent: "secure-client/2.0",
				Debug:     false,
			},
		},
		{
			name: "connection options with custom HTTP client",
			opts: &ConnectionOptions{
				HTTPClient: &http.Client{
					Timeout: 30 * time.Second,
				},
				Headers: map[string]string{
					"X-Custom-Header": "value",
				},
				UserAgent: "custom-client/1.0",
			},
		},
		{
			name: "connection options with headers",
			opts: &ConnectionOptions{
				Headers: map[string]string{
					"Authorization": "Bearer token",
					"Content-Type":  "application/json",
				},
				UserAgent: "api-client/1.0",
				Debug:     true,
			},
		},
		{
			name: "empty connection options - edge case",
			opts: &ConnectionOptions{},
		},
		{
			name: "connection options with empty user agent - edge case",
			opts: &ConnectionOptions{
				UserAgent: "",
				Debug:     false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test UserAgent field
			if tt.name == "basic connection options" {
				if tt.opts.UserAgent != "test-client/1.0" {
					t.Errorf("UserAgent = %q, want test-client/1.0", tt.opts.UserAgent)
				}
			}

			// Test Debug field consistency
			if tt.name == "basic connection options" && !tt.opts.Debug {
				t.Error("Debug should be true for basic options")
			}

			// Test TLS config presence
			if tt.name == "connection options with TLS config" {
				if tt.opts.TLSConfig == nil {
					t.Error("TLSConfig should not be nil")
				}
				if !tt.opts.TLSConfig.InsecureSkipVerify {
					t.Error("InsecureSkipVerify should be true")
				}
			}

			// Test HTTP client presence
			if tt.name == "connection options with custom HTTP client" {
				if tt.opts.HTTPClient == nil {
					t.Error("HTTPClient should not be nil")
				}
				if tt.opts.HTTPClient.Timeout != 30*time.Second {
					t.Errorf("HTTPClient.Timeout = %v, want 30s", tt.opts.HTTPClient.Timeout)
				}
			}

			// Test headers
			if tt.name == "connection options with headers" {
				if len(tt.opts.Headers) == 0 {
					t.Error("Headers should not be empty")
				}
				if tt.opts.Headers["Authorization"] != "Bearer token" {
					t.Error("Authorization header not set correctly")
				}
				if tt.opts.Headers["Content-Type"] != "application/json" {
					t.Error("Content-Type header not set correctly")
				}
			}

			// Test nil safety
			if tt.opts != nil {
				// Should not panic when accessing fields
				_ = tt.opts.UserAgent
				_ = tt.opts.Debug
			}
		})
	}
}

func TestPushOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     *PushOptions
		wantErr  bool
		hasValid bool
	}{
		{
			name: "valid push options",
			opts: &PushOptions{
				ProgressWriter: os.Stdout,
				ParallelLayers: 4,
				Force:          false,
				Platform:       "linux/amd64",
			},
			wantErr:  false,
			hasValid: true,
		},
		{
			name: "push options with force",
			opts: &PushOptions{
				Force:          true,
				ParallelLayers: 1,
				Platform:       "linux/arm64",
			},
			wantErr:  false,
			hasValid: true,
		},
		{
			name: "push options with zero parallel layers - edge case",
			opts: &PushOptions{
				ParallelLayers: 0,
				Platform:       "linux/amd64",
			},
			wantErr:  true,
			hasValid: false,
		},
		{
			name: "push options with negative parallel layers - edge case",
			opts: &PushOptions{
				ParallelLayers: -1,
				Platform:       "linux/amd64",
			},
			wantErr:  true,
			hasValid: false,
		},
		{
			name: "push options without platform - edge case",
			opts: &PushOptions{
				ParallelLayers: 2,
				Platform:       "",
			},
			wantErr:  true,
			hasValid: false,
		},
		{
			name:     "nil push options - edge case",
			opts:     nil,
			wantErr:  true,
			hasValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts == nil {
				if !tt.wantErr {
					t.Error("nil push options should be invalid")
				}
				return
			}

			// Validate parallel layers
			isValidParallel := tt.opts.ParallelLayers > 0
			if !isValidParallel && !tt.wantErr {
				t.Error("valid push options should have positive parallel layers")
			}

			// Validate platform
			isValidPlatform := tt.opts.Platform != ""
			if !isValidPlatform && !tt.wantErr {
				t.Error("valid push options should have platform specified")
			}

			// Test overall validity
			isValid := isValidParallel && isValidPlatform
			if isValid != tt.hasValid {
				t.Errorf("push options validity = %v, want %v", isValid, tt.hasValid)
			}

			// Test progress writer (if set)
			if tt.opts.ProgressWriter != nil {
				// Should be a valid writer
				var _ io.Writer = tt.opts.ProgressWriter
			}

			// Test force flag
			if tt.name == "push options with force" && !tt.opts.Force {
				t.Error("Force should be true for force push options")
			}
		})
	}
}

func TestPullOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     *PullOptions
		wantErr  bool
		hasValid bool
	}{
		{
			name: "valid pull options",
			opts: &PullOptions{
				ProgressWriter:  os.Stderr,
				VerifySignature: true,
				Platform:        "linux/amd64",
			},
			wantErr:  false,
			hasValid: true,
		},
		{
			name: "pull options without signature verification",
			opts: &PullOptions{
				VerifySignature: false,
				Platform:        "darwin/arm64",
			},
			wantErr:  false,
			hasValid: true,
		},
		{
			name: "pull options without platform - edge case",
			opts: &PullOptions{
				VerifySignature: true,
				Platform:        "",
			},
			wantErr:  true,
			hasValid: false,
		},
		{
			name:     "nil pull options - edge case",
			opts:     nil,
			wantErr:  true,
			hasValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts == nil {
				if !tt.wantErr {
					t.Error("nil pull options should be invalid")
				}
				return
			}

			// Validate platform
			isValidPlatform := tt.opts.Platform != ""
			if !isValidPlatform && !tt.wantErr {
				t.Error("valid pull options should have platform specified")
			}

			// Test overall validity
			isValid := isValidPlatform
			if isValid != tt.hasValid {
				t.Errorf("pull options validity = %v, want %v", isValid, tt.hasValid)
			}

			// Test progress writer (if set)
			if tt.opts.ProgressWriter != nil {
				// Should be a valid writer
				var _ io.Writer = tt.opts.ProgressWriter
			}

			// Test signature verification flag
			if tt.name == "valid pull options" && !tt.opts.VerifySignature {
				t.Error("VerifySignature should be true for valid pull options")
			}
		})
	}
}

func TestListOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     *ListOptions
		wantErr  bool
		hasValid bool
	}{
		{
			name: "valid list options",
			opts: &ListOptions{
				Limit:  10,
				Offset: 0,
				Filter: "myapp*",
			},
			wantErr:  false,
			hasValid: true,
		},
		{
			name: "list options with pagination",
			opts: &ListOptions{
				Limit:  50,
				Offset: 100,
			},
			wantErr:  false,
			hasValid: true,
		},
		{
			name: "list options with zero limit - edge case",
			opts: &ListOptions{
				Limit:  0,
				Offset: 10,
			},
			wantErr:  false, // zero limit might mean no limit
			hasValid: true,
		},
		{
			name: "list options with negative limit - edge case",
			opts: &ListOptions{
				Limit:  -1,
				Offset: 0,
			},
			wantErr:  true,
			hasValid: false,
		},
		{
			name: "list options with negative offset - edge case",
			opts: &ListOptions{
				Limit:  10,
				Offset: -1,
			},
			wantErr:  true,
			hasValid: false,
		},
		{
			name: "list options with filter only",
			opts: &ListOptions{
				Filter: "production/*",
			},
			wantErr:  false,
			hasValid: true,
		},
		{
			name:     "nil list options - edge case",
			opts:     nil,
			wantErr:  true,
			hasValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts == nil {
				if !tt.wantErr {
					t.Error("nil list options should be invalid")
				}
				return
			}

			// Validate limit
			isValidLimit := tt.opts.Limit >= 0
			if !isValidLimit && !tt.wantErr {
				t.Error("valid list options should have non-negative limit")
			}

			// Validate offset
			isValidOffset := tt.opts.Offset >= 0
			if !isValidOffset && !tt.wantErr {
				t.Error("valid list options should have non-negative offset")
			}

			// Test overall validity
			isValid := isValidLimit && isValidOffset
			if isValid != tt.hasValid {
				t.Errorf("list options validity = %v, want %v", isValid, tt.hasValid)
			}

			// Test specific values for known test cases
			if tt.name == "valid list options" {
				if tt.opts.Limit != 10 {
					t.Errorf("Limit = %d, want 10", tt.opts.Limit)
				}
				if tt.opts.Filter != "myapp*" {
					t.Errorf("Filter = %q, want myapp*", tt.opts.Filter)
				}
			}

			if tt.name == "list options with pagination" {
				if tt.opts.Limit != 50 {
					t.Errorf("Limit = %d, want 50", tt.opts.Limit)
				}
				if tt.opts.Offset != 100 {
					t.Errorf("Offset = %d, want 100", tt.opts.Offset)
				}
			}
		})
	}
}
