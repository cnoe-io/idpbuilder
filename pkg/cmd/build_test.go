package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateImageTag(t *testing.T) {
	tests := []struct {
		name      string
		tag       string
		expectErr bool
	}{
		{
			name:      "valid simple tag",
			tag:       "myapp",
			expectErr: false,
		},
		{
			name:      "valid tag with version",
			tag:       "myapp:v1.0.0",
			expectErr: false,
		},
		{
			name:      "valid tag with latest",
			tag:       "myapp:latest",
			expectErr: false,
		},
		{
			name:      "valid tag with registry",
			tag:       "registry.example.com/myapp:v1.0.0",
			expectErr: false,
		},
		{
			name:      "empty tag",
			tag:       "",
			expectErr: true,
		},
		{
			name:      "just colon",
			tag:       ":",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateImageTag(tt.tag)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParsePlatform(t *testing.T) {
	tests := []struct {
		name         string
		platform     string
		expectedOS   string
		expectedArch string
		expectErr    bool
	}{
		{
			name:         "valid linux/amd64",
			platform:     "linux/amd64",
			expectedOS:   "linux",
			expectedArch: "amd64",
			expectErr:    false,
		},
		{
			name:         "valid linux/arm64",
			platform:     "linux/arm64",
			expectedOS:   "linux",
			expectedArch: "arm64",
			expectErr:    false,
		},
		{
			name:         "valid windows/amd64",
			platform:     "windows/amd64",
			expectedOS:   "windows",
			expectedArch: "amd64",
			expectErr:    false,
		},
		{
			name:         "valid darwin/arm64",
			platform:     "darwin/arm64",
			expectedOS:   "darwin",
			expectedArch: "arm64",
			expectErr:    false,
		},
		{
			name:         "valid linux/arm",
			platform:     "linux/arm",
			expectedOS:   "linux",
			expectedArch: "arm",
			expectErr:    false,
		},
		{
			name:         "valid linux/386",
			platform:     "linux/386",
			expectedOS:   "linux",
			expectedArch: "386",
			expectErr:    false,
		},
		{
			name:      "invalid format - no slash",
			platform:  "linux",
			expectErr: true,
		},
		{
			name:      "invalid format - too many parts",
			platform:  "linux/amd64/extra",
			expectErr: true,
		},
		{
			name:      "invalid OS",
			platform:  "freebsd/amd64",
			expectErr: true,
		},
		{
			name:      "invalid architecture",
			platform:  "linux/sparc",
			expectErr: true,
		},
		{
			name:      "empty platform",
			platform:  "",
			expectErr: true,
		},
		{
			name:      "just slash",
			platform:  "/",
			expectErr: true,
		},
		{
			name:      "multiple slashes",
			platform:  "linux//amd64",
			expectErr: false, // Current implementation treats empty parts as valid
			expectedOS: "linux",
			expectedArch: "amd64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			osName, arch, err := parsePlatform(tt.platform)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOS, osName)
				assert.Equal(t, tt.expectedArch, arch)
			}
		})
	}
}

func TestBuildCmdInit(t *testing.T) {
	// Test that the command is properly initialized
	require.NotNil(t, BuildCmd)
	assert.Equal(t, "build", BuildCmd.Use)
	assert.Equal(t, "Assemble OCI image from context", BuildCmd.Short)
	assert.NotEmpty(t, BuildCmd.Long)

	// Test flags are set up
	contextFlag := BuildCmd.Flags().Lookup("context")
	require.NotNil(t, contextFlag)
	assert.Equal(t, ".", contextFlag.DefValue)

	tagFlag := BuildCmd.Flags().Lookup("tag")
	require.NotNil(t, tagFlag)
	assert.Equal(t, "", tagFlag.DefValue)

	platformFlag := BuildCmd.Flags().Lookup("platform")
	require.NotNil(t, platformFlag)
	assert.Equal(t, "linux/amd64", platformFlag.DefValue)
}

func TestBuildCmdValidation(t *testing.T) {
	// Test that the command accepts no arguments
	// cobra.NoArgs returns nil for no args, error for any args
	assert.Nil(t, BuildCmd.Args(BuildCmd, []string{}))
	assert.NotNil(t, BuildCmd.Args(BuildCmd, []string{"extra"}))
}