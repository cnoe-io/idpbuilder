package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFixturePath(t *testing.T) {
	path := GetFixturePath(t, "auth/credentials.yaml")
	assert.Contains(t, path, "fixtures/auth/credentials.yaml")

	// Verify file exists
	_, err := os.Stat(path)
	require.NoError(t, err)
}

func TestLoadFixture(t *testing.T) {
	content := LoadFixture(t, "auth/credentials.yaml")
	assert.NotEmpty(t, content)
	assert.Contains(t, string(content), "testuser")
}

func TestLoadJSONFixture(t *testing.T) {
	var tokens map[string]interface{}
	LoadJSONFixture(t, "auth/tokens.json", &tokens)

	assert.NotEmpty(t, tokens)
	assert.Contains(t, tokens, "tokens")
}

func TestCreateTempRegistry(t *testing.T) {
	registryPath, cleanup := CreateTempRegistry(t)
	defer cleanup()

	assert.NotEmpty(t, registryPath)

	// Verify registry directory structure
	v2Path := filepath.Join(registryPath, "v2")
	_, err := os.Stat(v2Path)
	require.NoError(t, err)
}

func TestSetupTestCredentials(t *testing.T) {
	username, password, cleanup := SetupTestCredentials(t)
	defer cleanup()

	assert.Equal(t, "testuser", username)
	assert.Equal(t, "testpass123", password)

	// Verify environment variables are set
	assert.Equal(t, "testuser", os.Getenv("GITEA_USERNAME"))
	assert.Equal(t, "testpass123", os.Getenv("GITEA_PASSWORD"))
}

func TestSetupTestTLS(t *testing.T) {
	caCert, clientCert, clientKey, cleanup := SetupTestTLS(t)
	defer cleanup()

	// Verify paths are not empty
	assert.NotEmpty(t, caCert)
	assert.NotEmpty(t, clientCert)
	assert.NotEmpty(t, clientKey)

	// Verify files exist
	_, err := os.Stat(caCert)
	require.NoError(t, err)
	_, err = os.Stat(clientCert)
	require.NoError(t, err)
	_, err = os.Stat(clientKey)
	require.NoError(t, err)
}

func TestMockGiteaRegistry(t *testing.T) {
	config := MockGiteaRegistry(t)

	assert.Equal(t, "localhost:3000", config.URL)
	assert.Equal(t, "test-org", config.Namespace)
	assert.Equal(t, "test-app", config.Repository)
}

func TestCreateTestImageFixture(t *testing.T) {
	imagePath := CreateTestImageFixture(t, "test")
	defer os.RemoveAll(imagePath)

	// Verify manifest and config files exist
	manifestPath := filepath.Join(imagePath, "manifest.json")
	configPath := filepath.Join(imagePath, "config.json")

	_, err := os.Stat(manifestPath)
	require.NoError(t, err)
	_, err = os.Stat(configPath)
	require.NoError(t, err)
}

func TestCompareManifests(t *testing.T) {
	manifest1 := LoadFixture(t, "images/manifest.json")
	manifest2 := LoadFixture(t, "images/manifest.json")

	// Should not panic and should pass (identical manifests)
	CompareManifests(t, manifest1, manifest2)
}

func TestSetupTestEnvironment(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Cleanup()

	// Verify all components are set up
	assert.NotEmpty(t, env.TempDir)
	assert.NotEmpty(t, env.Registry)
	assert.NotNil(t, env.Credentials)
	assert.NotNil(t, env.TLS)

	// Verify credentials
	assert.Equal(t, "testuser", env.Credentials.Username)
	assert.Equal(t, "testpass123", env.Credentials.Password)
	assert.Equal(t, "test-bearer-token", env.Credentials.Token)

	// Verify TLS config
	assert.NotEmpty(t, env.TLS.CACert)
	assert.NotEmpty(t, env.TLS.ClientCert)
	assert.NotEmpty(t, env.TLS.ClientKey)
	assert.False(t, env.TLS.InsecureSkipVerify)
}