package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateImageReference(t *testing.T) {
	tests := []struct {
		name      string
		ref       string
		expectErr bool
	}{
		{
			name:      "valid simple reference",
			ref:       "myapp",
			expectErr: false,
		},
		{
			name:      "valid reference with tag",
			ref:       "myapp:latest",
			expectErr: false,
		},
		{
			name:      "valid reference with version",
			ref:       "myapp:v1.0.0",
			expectErr: false,
		},
		{
			name:      "valid reference with registry",
			ref:       "registry.example.com/myapp:latest",
			expectErr: false,
		},
		{
			name:      "empty reference",
			ref:       "",
			expectErr: true,
		},
		{
			name:      "starts with colon",
			ref:       ":myapp",
			expectErr: true,
		},
		{
			name:      "ends with colon",
			ref:       "myapp:",
			expectErr: true,
		},
		{
			name:      "double colon",
			ref:       "myapp::latest",
			expectErr: true,
		},
		{
			name:      "just colon",
			ref:       ":",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateImageReference(tt.ref)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetDefaultRegistry(t *testing.T) {
	// Test default value
	originalEnv := os.Getenv("IDPBUILDER_REGISTRY")
	os.Unsetenv("IDPBUILDER_REGISTRY")
	defer os.Setenv("IDPBUILDER_REGISTRY", originalEnv)

	registry := getDefaultRegistry()
	assert.Equal(t, "gitea.cnoe.localtest.me:8443", registry)
}

func TestGetDefaultRegistryWithEnvironment(t *testing.T) {
	// Test with environment variable
	customRegistry := "custom-registry.example.com:5000"
	originalEnv := os.Getenv("IDPBUILDER_REGISTRY")
	os.Setenv("IDPBUILDER_REGISTRY", customRegistry)
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("IDPBUILDER_REGISTRY")
		} else {
			os.Setenv("IDPBUILDER_REGISTRY", originalEnv)
		}
	}()

	registry := getDefaultRegistry()
	assert.Equal(t, customRegistry, registry)
}

func TestPushCmdInit(t *testing.T) {
	// Test that the command is properly initialized
	require.NotNil(t, PushCmd)
	assert.Equal(t, "push IMAGE[:TAG]", PushCmd.Use)
	assert.Equal(t, "Push image to Gitea registry", PushCmd.Short)
	assert.NotEmpty(t, PushCmd.Long)

	// Test flags are set up
	insecureFlag := PushCmd.Flags().Lookup("insecure")
	require.NotNil(t, insecureFlag)
	assert.Equal(t, "false", insecureFlag.DefValue)

	registryFlag := PushCmd.Flags().Lookup("registry")
	require.NotNil(t, registryFlag)
	assert.Equal(t, getDefaultRegistry(), registryFlag.DefValue)
}

func TestPushCmdValidation(t *testing.T) {
	// Test that the command requires exactly one argument
	// cobra.ExactArgs(1) returns nil for valid args, error for invalid
	assert.Nil(t, PushCmd.Args(PushCmd, []string{"image:tag"}))
	assert.NotNil(t, PushCmd.Args(PushCmd, []string{}))
	assert.NotNil(t, PushCmd.Args(PushCmd, []string{"image1", "image2"}))
}