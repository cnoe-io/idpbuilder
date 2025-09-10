package gitea

import (
	"os"
	"testing"

	"github.com/cnoe-io/idpbuilder/pkg/certs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	registryURL := "https://gitea.example.com"
	certManager := certs.NewTrustStore()

	client, err := NewClient(registryURL, certManager)

	// Since we're using placeholder credentials, we expect this to succeed in initialization
	// The actual registry connection would fail, but client creation should work
	require.NotNil(t, client)
	require.NoError(t, err)

	// Verify client configuration
	assert.Equal(t, registryURL, client.config.URL)
	assert.Equal(t, "admin", client.config.Username) // From getRegistryUsername()
	assert.Equal(t, "password", client.config.Token) // From getRegistryPassword()
	assert.False(t, client.config.Insecure)
}

func TestNewInsecureClient(t *testing.T) {
	registryURL := "http://gitea.example.com"

	client, err := NewInsecureClient(registryURL)

	// Client creation should succeed
	require.NotNil(t, client)
	require.NoError(t, err)

	// Verify client configuration
	assert.Equal(t, registryURL, client.config.URL)
	assert.Equal(t, "admin", client.config.Username)
	assert.Equal(t, "password", client.config.Token)
	assert.True(t, client.config.Insecure)
}

func TestGetRegistryUsername(t *testing.T) {
	// Test default username
	username := getRegistryUsername()
	assert.Equal(t, "admin", username)
}

func TestGetRegistryPassword(t *testing.T) {
	// Test default password
	password := getRegistryPassword()
	assert.Equal(t, "password", password)
}

func TestGetImageContentForReference(t *testing.T) {
	registryURL := "https://gitea.example.com"
	certManager := certs.NewTrustStore()

	client, err := NewClient(registryURL, certManager)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Test getting image content (placeholder implementation)
	imageRef := "myapp:latest"
	content, err := client.getImageContentForReference(imageRef)

	require.NoError(t, err)
	require.NotNil(t, content)

	// Read the content to verify it's a valid JSON manifest
	buffer := make([]byte, 1024)
	n, err := content.Read(buffer)
	require.NoError(t, err)
	require.Greater(t, n, 0)

	// Verify it looks like a Docker manifest
	contentStr := string(buffer[:n])
	assert.Contains(t, contentStr, "mediaType")
	assert.Contains(t, contentStr, "schemaVersion")
	assert.Contains(t, contentStr, "config")
	assert.Contains(t, contentStr, "layers")
}

func TestPushProgressStruct(t *testing.T) {
	// Test PushProgress struct
	progress := PushProgress{
		CurrentLayer: 3,
		TotalLayers:  10,
		Percentage:   30,
	}

	assert.Equal(t, 3, progress.CurrentLayer)
	assert.Equal(t, 10, progress.TotalLayers)
	assert.Equal(t, 30, progress.Percentage)
}

func TestClientConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		insecure bool
	}{
		{
			name:     "secure client",
			url:      "https://gitea.secure.com",
			insecure: false,
		},
		{
			name:     "insecure client",
			url:      "http://gitea.insecure.com",
			insecure: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client *Client
			var err error

			if tt.insecure {
				client, err = NewInsecureClient(tt.url)
			} else {
				certManager := certs.NewTrustStore()
				client, err = NewClient(tt.url, certManager)
			}

			require.NoError(t, err)
			require.NotNil(t, client)
			assert.Equal(t, tt.url, client.config.URL)
			assert.Equal(t, tt.insecure, client.config.Insecure)
		})
	}
}

// Test environment variable integration (if applicable)
func TestClientWithEnvironmentVariables(t *testing.T) {
	// Save original environment
	originalUsername := os.Getenv("REGISTRY_USERNAME")
	originalPassword := os.Getenv("REGISTRY_PASSWORD")
	
	// Clean up after test
	defer func() {
		if originalUsername != "" {
			os.Setenv("REGISTRY_USERNAME", originalUsername)
		} else {
			os.Unsetenv("REGISTRY_USERNAME")
		}
		if originalPassword != "" {
			os.Setenv("REGISTRY_PASSWORD", originalPassword)
		} else {
			os.Unsetenv("REGISTRY_PASSWORD")
		}
	}()

	// Test with current implementation (uses hardcoded values)
	// This test verifies the current behavior, but the implementation
	// should be enhanced to read from environment variables
	username := getRegistryUsername()
	password := getRegistryPassword()

	assert.Equal(t, "admin", username)
	assert.Equal(t, "password", password)
}