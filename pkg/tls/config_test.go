package tls

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name     string
		insecure bool
		expected bool
	}{
		{
			name:     "secure mode (default)",
			insecure: false,
			expected: false,
		},
		{
			name:     "insecure mode for development",
			insecure: true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig(tt.insecure)

			require.NotNil(t, cfg)
			assert.Equal(t, tt.expected, cfg.InsecureSkipVerify)
		})
	}
}

func TestToTLSConfig(t *testing.T) {
	tests := []struct {
		name     string
		insecure bool
	}{
		{
			name:     "secure TLS config",
			insecure: false,
		},
		{
			name:     "insecure TLS config",
			insecure: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig(tt.insecure)
			tlsConfig := cfg.ToTLSConfig()

			require.NotNil(t, tlsConfig)
			assert.Equal(t, tt.insecure, tlsConfig.InsecureSkipVerify)
			assert.IsType(t, &tls.Config{}, tlsConfig)
		})
	}
}

func TestApplyToHTTPClient(t *testing.T) {
	t.Run("apply to client with nil transport", func(t *testing.T) {
		client := &http.Client{}
		cfg := NewConfig(true)

		cfg.ApplyToHTTPClient(client)

		require.NotNil(t, client.Transport)
		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)
		require.NotNil(t, transport.TLSClientConfig)
		assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	})

	t.Run("apply to client with existing transport", func(t *testing.T) {
		transport := &http.Transport{}
		client := &http.Client{Transport: transport}
		cfg := NewConfig(false)

		cfg.ApplyToHTTPClient(client)

		require.NotNil(t, transport.TLSClientConfig)
		assert.False(t, transport.TLSClientConfig.InsecureSkipVerify)
	})

	t.Run("apply insecure config", func(t *testing.T) {
		client := &http.Client{}
		cfg := NewConfig(true)

		cfg.ApplyToHTTPClient(client)

		transport := client.Transport.(*http.Transport)
		assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	})
}

func TestApplyToTransport(t *testing.T) {
	t.Run("apply secure config to transport", func(t *testing.T) {
		transport := &http.Transport{}
		cfg := NewConfig(false)

		cfg.ApplyToTransport(transport)

		require.NotNil(t, transport.TLSClientConfig)
		assert.False(t, transport.TLSClientConfig.InsecureSkipVerify)
	})

	t.Run("apply insecure config to transport", func(t *testing.T) {
		transport := &http.Transport{}
		cfg := NewConfig(true)

		cfg.ApplyToTransport(transport)

		require.NotNil(t, transport.TLSClientConfig)
		assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	})
}

func TestIsSecure(t *testing.T) {
	tests := []struct {
		name     string
		insecure bool
		expected bool
	}{
		{
			name:     "secure configuration",
			insecure: false,
			expected: true,
		},
		{
			name:     "insecure configuration",
			insecure: true,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig(tt.insecure)
			assert.Equal(t, tt.expected, cfg.IsSecure())
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		insecure bool
		expected string
	}{
		{
			name:     "secure configuration string",
			insecure: false,
			expected: "TLS Config: SECURE (certificate verification enabled)",
		},
		{
			name:     "insecure configuration string",
			insecure: true,
			expected: "TLS Config: INSECURE (certificate verification disabled)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig(tt.insecure)
			assert.Equal(t, tt.expected, cfg.String())
		})
	}
}

// Integration test to verify the entire configuration flow
func TestConfigurationIntegration(t *testing.T) {
	t.Run("complete secure configuration flow", func(t *testing.T) {
		// Create secure configuration
		cfg := NewConfig(false)

		// Verify configuration properties
		assert.True(t, cfg.IsSecure())
		assert.False(t, cfg.InsecureSkipVerify)
		assert.Contains(t, cfg.String(), "SECURE")

		// Apply to HTTP client
		client := &http.Client{}
		cfg.ApplyToHTTPClient(client)

		// Verify transport configuration
		transport := client.Transport.(*http.Transport)
		assert.False(t, transport.TLSClientConfig.InsecureSkipVerify)
	})

	t.Run("complete insecure configuration flow", func(t *testing.T) {
		// Create insecure configuration
		cfg := NewConfig(true)

		// Verify configuration properties
		assert.False(t, cfg.IsSecure())
		assert.True(t, cfg.InsecureSkipVerify)
		assert.Contains(t, cfg.String(), "INSECURE")

		// Apply to HTTP transport directly
		transport := &http.Transport{}
		cfg.ApplyToTransport(transport)

		// Verify transport configuration
		assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	})
}