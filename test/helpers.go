package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEnvironment contains all test resources for comprehensive testing
type TestEnvironment struct {
	TempDir     string
	Registry    string
	Credentials *TestCredentials
	TLS         *TestTLSConfig
	Cleanup     func()
}

// TestCredentials for authentication testing
type TestCredentials struct {
	Username string
	Password string
	Token    string
}

// TestTLSConfig for TLS testing
type TestTLSConfig struct {
	CACert             string
	ClientCert         string
	ClientKey          string
	InsecureSkipVerify bool
}

// GiteaRegistryConfig for Gitea-specific testing
type GiteaRegistryConfig struct {
	URL        string
	Namespace  string
	Repository string
}

// GetFixturePath returns absolute path to fixture file
// relativePath should be relative to test/fixtures/ directory
func GetFixturePath(t *testing.T, relativePath string) string {
	t.Helper()

	// Get the directory where this test file is located
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "Failed to get caller information")

	testDir := filepath.Dir(filename)
	fixturePath := filepath.Join(testDir, "fixtures", relativePath)

	// Ensure the fixture file exists
	_, err := os.Stat(fixturePath)
	require.NoError(t, err, "Fixture file not found: %s", fixturePath)

	return fixturePath
}

// LoadFixture loads fixture content as bytes
func LoadFixture(t *testing.T, relativePath string) []byte {
	t.Helper()

	fixturePath := GetFixturePath(t, relativePath)
	content, err := ioutil.ReadFile(fixturePath)
	require.NoError(t, err, "Failed to read fixture file: %s", fixturePath)

	return content
}

// LoadJSONFixture loads and unmarshals JSON fixture into target
func LoadJSONFixture(t *testing.T, relativePath string, target interface{}) {
	t.Helper()

	content := LoadFixture(t, relativePath)
	err := json.Unmarshal(content, target)
	require.NoError(t, err, "Failed to unmarshal JSON fixture: %s", relativePath)
}

// CreateTempRegistry creates a temporary directory simulating a registry
// Returns the registry path and a cleanup function
func CreateTempRegistry(t *testing.T) (string, func()) {
	t.Helper()

	tempDir, err := ioutil.TempDir("", "test-registry-*")
	require.NoError(t, err, "Failed to create temp registry directory")

	// Create basic registry structure
	err = os.MkdirAll(filepath.Join(tempDir, "v2"), 0755)
	require.NoError(t, err, "Failed to create registry v2 directory")

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// SetupTestCredentials sets up test authentication environment
// Returns username, password, and cleanup function
func SetupTestCredentials(t *testing.T) (string, string, func()) {
	t.Helper()

	username := "testuser"
	password := "testpass123"

	// Set environment variables for testing
	originalUsername := os.Getenv("GITEA_USERNAME")
	originalPassword := os.Getenv("GITEA_PASSWORD")

	err := os.Setenv("GITEA_USERNAME", username)
	require.NoError(t, err, "Failed to set GITEA_USERNAME")

	err = os.Setenv("GITEA_PASSWORD", password)
	require.NoError(t, err, "Failed to set GITEA_PASSWORD")

	cleanup := func() {
		if originalUsername != "" {
			os.Setenv("GITEA_USERNAME", originalUsername)
		} else {
			os.Unsetenv("GITEA_USERNAME")
		}

		if originalPassword != "" {
			os.Setenv("GITEA_PASSWORD", originalPassword)
		} else {
			os.Unsetenv("GITEA_PASSWORD")
		}
	}

	return username, password, cleanup
}

// SetupTestTLS creates test certificates and returns paths
// Returns CA cert, client cert, client key paths, and cleanup function
func SetupTestTLS(t *testing.T) (string, string, string, func()) {
	t.Helper()

	tempDir, err := ioutil.TempDir("", "test-tls-*")
	require.NoError(t, err, "Failed to create temp TLS directory")

	caCertPath := filepath.Join(tempDir, "ca.crt")
	clientCertPath := filepath.Join(tempDir, "client.crt")
	clientKeyPath := filepath.Join(tempDir, "client.key")

	// Copy fixture certificates to temp directory for modification
	caCertContent := LoadFixture(t, "certs/ca.crt")
	clientCertContent := LoadFixture(t, "certs/client.crt")
	clientKeyContent := LoadFixture(t, "certs/client.key")

	err = ioutil.WriteFile(caCertPath, caCertContent, 0644)
	require.NoError(t, err, "Failed to write CA cert")

	err = ioutil.WriteFile(clientCertPath, clientCertContent, 0644)
	require.NoError(t, err, "Failed to write client cert")

	err = ioutil.WriteFile(clientKeyPath, clientKeyContent, 0600)
	require.NoError(t, err, "Failed to write client key")

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return caCertPath, clientCertPath, clientKeyPath, cleanup
}

// MockGiteaRegistry creates a mock Gitea registry configuration
func MockGiteaRegistry(t *testing.T) *GiteaRegistryConfig {
	t.Helper()

	return &GiteaRegistryConfig{
		URL:        "localhost:3000",
		Namespace:  "test-org",
		Repository: "test-app",
	}
}

// CreateTestImageFixture creates a minimal OCI image for testing
// Returns the path to the created image directory
func CreateTestImageFixture(t *testing.T, name string) string {
	t.Helper()

	tempDir, err := ioutil.TempDir("", fmt.Sprintf("test-image-%s-*", name))
	require.NoError(t, err, "Failed to create temp image directory")

	// Create manifest.json
	manifestPath := filepath.Join(tempDir, "manifest.json")
	manifestContent := LoadFixture(t, "images/manifest.json")
	err = ioutil.WriteFile(manifestPath, manifestContent, 0644)
	require.NoError(t, err, "Failed to write manifest.json")

	// Create config.json
	configPath := filepath.Join(tempDir, "config.json")
	configContent := LoadFixture(t, "images/config.json")
	err = ioutil.WriteFile(configPath, configContent, 0644)
	require.NoError(t, err, "Failed to write config.json")

	return tempDir
}

// CompareManifests compares two OCI manifests for testing
func CompareManifests(t *testing.T, expected, actual []byte) {
	t.Helper()

	var expectedManifest, actualManifest map[string]interface{}

	err := json.Unmarshal(expected, &expectedManifest)
	require.NoError(t, err, "Failed to unmarshal expected manifest")

	err = json.Unmarshal(actual, &actualManifest)
	require.NoError(t, err, "Failed to unmarshal actual manifest")

	require.Equal(t, expectedManifest, actualManifest, "Manifests do not match")
}

// SetupTestEnvironment prepares complete test environment
// Returns a TestEnvironment with all resources configured
func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()

	// Create temp directory for the entire test environment
	tempDir, err := ioutil.TempDir("", "test-env-*")
	require.NoError(t, err, "Failed to create test environment directory")

	// Setup registry
	registryPath, registryCleanup := CreateTempRegistry(t)

	// Setup credentials
	username, password, credCleanup := SetupTestCredentials(t)
	credentials := &TestCredentials{
		Username: username,
		Password: password,
		Token:    "test-bearer-token",
	}

	// Setup TLS
	caCert, clientCert, clientKey, tlsCleanup := SetupTestTLS(t)
	tlsConfig := &TestTLSConfig{
		CACert:             caCert,
		ClientCert:         clientCert,
		ClientKey:          clientKey,
		InsecureSkipVerify: false,
	}

	// Combined cleanup function
	cleanup := func() {
		tlsCleanup()
		credCleanup()
		registryCleanup()
		os.RemoveAll(tempDir)
	}

	return &TestEnvironment{
		TempDir:     tempDir,
		Registry:    registryPath,
		Credentials: credentials,
		TLS:         tlsConfig,
		Cleanup:     cleanup,
	}
}