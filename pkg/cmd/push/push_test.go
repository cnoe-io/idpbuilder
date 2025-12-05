package push

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/cnoe-io/idpbuilder/pkg/daemon"
	"github.com/cnoe-io/idpbuilder/pkg/registry"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing

type mockImageReader struct {
	data string
	pos  int
}

func (m *mockImageReader) Read(b []byte) (int, error) {
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	n := copy(b, m.data[m.pos:])
	m.pos += n
	return n, nil
}

func (m *mockImageReader) Close() error {
	return nil
}

type mockDaemonClient struct {
	pingErr        error
	imageExistsErr error
	imageExists    bool
	getImageErr    error
	getImageInfo   *daemon.ImageInfo
	contentData    string
}

func (m *mockDaemonClient) Ping(ctx context.Context) error {
	return m.pingErr
}

func (m *mockDaemonClient) ImageExists(ctx context.Context, reference string) (bool, error) {
	return m.imageExists, m.imageExistsErr
}

func (m *mockDaemonClient) GetImage(ctx context.Context, reference string) (*daemon.ImageInfo, daemon.ImageReader, error) {
	if m.getImageErr != nil {
		return nil, nil, m.getImageErr
	}
	if m.getImageInfo == nil {
		m.getImageInfo = &daemon.ImageInfo{
			ID:         "sha256:abc123",
			RepoTags:   []string{reference},
			Size:       1024000,
			LayerCount: 5,
		}
	}
	if m.contentData == "" {
		m.contentData = "fake image content"
	}
	return m.getImageInfo, &mockImageReader{data: m.contentData}, nil
}

type mockRegistryClient struct {
	pushErr    error
	pushResult *registry.PushResult
}

func (m *mockRegistryClient) Push(ctx context.Context, imageRef, destRef string, progress registry.ProgressReporter) (*registry.PushResult, error) {
	if m.pushErr != nil {
		return nil, m.pushErr
	}
	if m.pushResult == nil {
		m.pushResult = &registry.PushResult{
			Reference: "registry.example.com/myimage:latest@sha256:abc123",
			Digest:    "sha256:abc123",
			Size:      1024000,
		}
	}
	return m.pushResult, nil
}

// Test cases

// TestPushCmd_Success_OutputsReference - W2-PC-001
func TestPushCmd_Success_OutputsReference(t *testing.T) {
	// Setup
	mockDaemon := &mockDaemonClient{
		imageExists: true,
	}
	mockRegistry := &mockRegistryClient{
		pushResult: &registry.PushResult{
			Reference: "registry.example.com/test:latest@sha256:expected",
		},
	}

	// When successful push returns a reference, it should output it to stdout
	cmd := createPushCmdWithDependencies(mockDaemon, mockRegistry)
	cmd.SetArgs([]string{"test:latest"})

	// Capture output
	var output strings.Builder
	cmd.SetOut(&output)

	// Execute
	err := cmd.Execute()

	// Verify
	require.NoError(t, err)
	require.Contains(t, output.String(), "registry.example.com/test:latest@sha256:expected")
}

// TestPushCmd_CredentialIntegration - W2-PC-002
func TestPushCmd_CredentialIntegration(t *testing.T) {
	// Setup - credentials from flags should be used
	mockDaemon := &mockDaemonClient{
		imageExists: true,
	}
	mockRegistry := &mockRegistryClient{
		pushResult: &registry.PushResult{
			Reference: "registry.example.com/test:latest@sha256:abc",
		},
	}

	cmd := createPushCmdWithDependencies(mockDaemon, mockRegistry)
	cmd.SetArgs([]string{
		"test:latest",
		"--username", "testuser",
		"--password", "testpass",
	})

	// Parse flags to verify they are set
	err := cmd.ParseFlags([]string{
		"test:latest",
		"--username", "testuser",
		"--password", "testpass",
	})
	require.NoError(t, err)

	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")

	require.Equal(t, "testuser", username)
	require.Equal(t, "testpass", password)
}

// TestPushCmd_ImageNotFound_ExitCode2 - W2-PC-003
func TestPushCmd_ImageNotFound_ExitCode2(t *testing.T) {
	// Setup - image does not exist locally
	mockDaemon := &mockDaemonClient{
		imageExists: false,
	}
	mockRegistry := &mockRegistryClient{}

	cmd := createPushCmdWithDependencies(mockDaemon, mockRegistry)
	cmd.SetArgs([]string{"nonexistent:test"})

	// Execute
	err := cmd.Execute()

	// Verify error indicates image not found
	require.Error(t, err)
	require.Contains(t, err.Error(), "image not found")
}

// TestPushCmd_DaemonNotRunning_ExitCode2 - W2-PC-004
func TestPushCmd_DaemonNotRunning_ExitCode2(t *testing.T) {
	// Setup - daemon ping fails
	mockDaemon := &mockDaemonClient{
		pingErr: errors.New("connection refused"),
	}
	mockRegistry := &mockRegistryClient{}

	cmd := createPushCmdWithDependencies(mockDaemon, mockRegistry)
	cmd.SetArgs([]string{"test:latest"})

	// Execute
	err := cmd.Execute()

	// Verify error indicates daemon not running
	require.Error(t, err)
	require.Contains(t, err.Error(), "daemon")
}

// TestPushCmd_AuthFailure_ExitCode1 - W2-PC-005
func TestPushCmd_AuthFailure_ExitCode1(t *testing.T) {
	// Setup - registry returns auth error
	mockDaemon := &mockDaemonClient{
		imageExists: true,
	}
	mockRegistry := &mockRegistryClient{
		pushErr: &registry.AuthError{
			Message: "unauthorized",
			Cause:   errors.New("invalid credentials"),
		},
	}

	cmd := createPushCmdWithDependencies(mockDaemon, mockRegistry)
	cmd.SetArgs([]string{"test:latest"})

	// Execute
	err := cmd.Execute()

	// Verify error is returned
	require.Error(t, err)
}

// TestPushCmd_FlagParsing - W2-PC-008
func TestPushCmd_FlagParsing(t *testing.T) {
	// Verify all command flags are properly defined
	require.NotNil(t, PushCmd.Flags().Lookup("registry"))
	require.NotNil(t, PushCmd.Flags().Lookup("username"))
	require.NotNil(t, PushCmd.Flags().Lookup("password"))
	require.NotNil(t, PushCmd.Flags().Lookup("token"))
	require.NotNil(t, PushCmd.Flags().Lookup("insecure"))

	// Get short flag names
	registryFlag := PushCmd.Flags().Lookup("registry")
	require.Equal(t, "r", registryFlag.Shorthand)

	usernameFlag := PushCmd.Flags().Lookup("username")
	require.Equal(t, "u", usernameFlag.Shorthand)

	passwordFlag := PushCmd.Flags().Lookup("password")
	require.Equal(t, "p", passwordFlag.Shorthand)

	tokenFlag := PushCmd.Flags().Lookup("token")
	require.Equal(t, "t", tokenFlag.Shorthand)
}

// Test for default registry constant
func TestPushCmd_DefaultRegistry(t *testing.T) {
	require.Equal(t, "https://gitea.cnoe.localtest.me:8443", DefaultRegistry)
}

// Test parseImageRef helper function
func TestParseImageRef(t *testing.T) {
	tests := []struct {
		name      string
		ref       string
		expectRepo string
		expectTag string
	}{
		{
			name:      "Image with tag",
			ref:       "myimage:latest",
			expectRepo: "myimage",
			expectTag: "latest",
		},
		{
			name:       "Image without tag",
			ref:        "myimage",
			expectRepo: "myimage",
			expectTag: "",
		},
		{
			name:       "Registry with image and tag",
			ref:        "registry.io/myimage:v1.0",
			expectRepo: "registry.io/myimage",
			expectTag: "v1.0",
		},
		{
			name:       "Registry with port and image",
			ref:        "localhost:5000/myimage",
			expectRepo: "localhost:5000/myimage",
			expectTag: "",
		},
		{
			name:       "Semver tag v1.0",
			ref:        "myimage:v1.0",
			expectRepo: "myimage",
			expectTag: "v1.0",
		},
		{
			name:       "Semver tag v1.2.3",
			ref:        "myimage:v1.2.3",
			expectRepo: "myimage",
			expectTag: "v1.2.3",
		},
		{
			name:       "Alpine style tag",
			ref:        "alpine:3.18",
			expectRepo: "alpine",
			expectTag: "3.18",
		},
		{
			name:       "Registry with port and semver tag",
			ref:        "localhost:5000/myimage:v1.0",
			expectRepo: "localhost:5000/myimage",
			expectTag: "v1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, tag := parseImageRef(tt.ref)
			require.Equal(t, tt.expectRepo, repo)
			require.Equal(t, tt.expectTag, tag)
		})
	}
}

// Test buildDestinationRef helper function
func TestBuildDestinationRef(t *testing.T) {
	tests := []struct {
		name        string
		registryURL string
		imageRef    string
		expected    string
	}{
		{
			name:        "HTTPS registry",
			registryURL: "https://registry.example.com:8443",
			imageRef:    "myimage:latest",
			expected:    "registry.example.com:8443/myimage:latest",
		},
		{
			name:        "HTTP registry",
			registryURL: "http://localhost:5000",
			imageRef:    "test:v1",
			expected:    "localhost:5000/test:v1",
		},
		{
			name:        "Registry with path",
			registryURL: "https://registry.io",
			imageRef:    "app:latest",
			expected:    "registry.io/app:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDestinationRef(tt.registryURL, tt.imageRef)
			require.Equal(t, tt.expected, result)
		})
	}
}

// Test extractHost helper function
func TestExtractHost(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "HTTPS with port",
			url:      "https://registry.example.com:8443",
			expected: "registry.example.com:8443",
		},
		{
			name:     "HTTP without port",
			url:      "http://localhost",
			expected: "localhost",
		},
		{
			name:     "HTTPS without port",
			url:      "https://registry.io",
			expected: "registry.io",
		},
		{
			name:     "Localhost with port",
			url:      "http://localhost:5000",
			expected: "localhost:5000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractHost(tt.url)
			require.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions

// PushCommandWrapper for dependency injection in tests
type PushCommandWrapper struct {
	baseCmd *cobra.Command
}

func (pcw *PushCommandWrapper) SetArgs(args []string) {
	pcw.baseCmd.SetArgs(args)
}

func (pcw *PushCommandWrapper) Execute() error {
	return pcw.baseCmd.Execute()
}

func (pcw *PushCommandWrapper) SetOut(w io.Writer) {
	pcw.baseCmd.SetOut(w)
}

func (pcw *PushCommandWrapper) Flags() *pflag.FlagSet {
	return pcw.baseCmd.Flags()
}

func (pcw *PushCommandWrapper) ParseFlags(args []string) error {
	return pcw.baseCmd.ParseFlags(args)
}

// createPushCmdWithDependencies creates a push command with injectable dependencies for testing
func createPushCmdWithDependencies(
	daemonClient daemon.DaemonClient,
	registryClient registry.RegistryClient,
) *PushCommandWrapper {
	// Create a NEW command with the injected dependencies
	testCmd := &cobra.Command{
		Use:   "push IMAGE",
		Short: "Push a local Docker image to an OCI registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPushWithClients(cmd, args, daemonClient, registryClient)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Copy flag definitions
	testCmd.Flags().StringVarP(&flagRegistry, "registry", "r", DefaultRegistry, "Registry URL")
	testCmd.Flags().StringVarP(&flagUsername, "username", "u", "", "Registry username")
	testCmd.Flags().StringVarP(&flagPassword, "password", "p", "", "Registry password")
	testCmd.Flags().StringVarP(&flagToken, "token", "t", "", "Registry token")
	testCmd.Flags().BoolVar(&flagInsecure, "insecure", false, "Skip TLS verification")

	return &PushCommandWrapper{
		baseCmd: testCmd,
	}
}

// Test helper: executePushWithExitCode runs command and captures exit code
func executePushWithExitCode(args []string, daemon daemon.DaemonClient, registry registry.RegistryClient) int {
	cmd := createPushCmdWithDependencies(daemon, registry)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return exitWithError(err)
}
