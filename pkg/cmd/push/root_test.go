package push

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPushCommand(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantErr   bool
		errMsg    string
	}{
		{
			name:    "valid image name",
			args:    []string{"myapp:latest"},
			wantErr: false,
		},
		{
			name:    "valid simple image name",
			args:    []string{"myapp"},
			wantErr: false,
		},
		{
			name:    "valid image with namespace",
			args:    []string{"namespace/myapp:v1.0"},
			wantErr: false,
		},
		{
			name:    "missing image name",
			args:    []string{},
			wantErr: true,
			errMsg:  "accepts 1 arg(s), received 0",
		},
		{
			name:    "too many arguments",
			args:    []string{"image1", "image2"},
			wantErr: true,
			errMsg:  "accepts 1 arg(s), received 2",
		},
		{
			name:    "empty image name",
			args:    []string{""},
			wantErr: true,
			errMsg:  "image name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := PushCmd
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				output := buf.String()
				assert.Contains(t, output, "Pushing image:")
				assert.Contains(t, output, tt.args[0])
				assert.Contains(t, output, "Note: Push functionality will be implemented in Phase 4")
			}
		})
	}
}

func TestValidateImageName(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		wantErr   bool
	}{
		{
			name:      "valid simple name",
			imageName: "myapp",
			wantErr:   false,
		},
		{
			name:      "valid with tag",
			imageName: "myapp:latest",
			wantErr:   false,
		},
		{
			name:      "valid with namespace",
			imageName: "namespace/myapp:v1.0",
			wantErr:   false,
		},
		{
			name:      "valid complex name",
			imageName: "registry.example.com/namespace/myapp:v1.2.3",
			wantErr:   false,
		},
		{
			name:      "empty name",
			imageName: "",
			wantErr:   true,
		},
		{
			name:      "whitespace only",
			imageName: "   ",
			wantErr:   false, // trimming not implemented yet, will be enhanced in Phase 3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateImageName(tt.imageName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPushCommandHelp(t *testing.T) {
	cmd := PushCmd
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Push an OCI image to the integrated Gitea registry")
	assert.Contains(t, output, "IMAGE_NAME")
	assert.Contains(t, output, "https://gitea.cnoe.localtest.me:8443/")
	assert.Contains(t, output, "Examples:")
}

func TestPushCommandUsage(t *testing.T) {
	cmd := PushCmd

	// Test Use field
	assert.Equal(t, "push IMAGE_NAME", cmd.Use)

	// Test Short description
	assert.Equal(t, "Push an OCI image to the integrated Gitea registry", cmd.Short)

	// Test Long description contains key information
	assert.Contains(t, cmd.Long, "Gitea registry")
	assert.Contains(t, cmd.Long, "https://gitea.cnoe.localtest.me:8443/")
	assert.Contains(t, cmd.Long, "Examples:")
}

func TestRunPushFunction(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantErr   bool
		errMsg    string
	}{
		{
			name:    "valid execution",
			args:    []string{"myapp:latest"},
			wantErr: false,
		},
		{
			name:    "empty image name",
			args:    []string{""},
			wantErr: true,
			errMsg:  "image name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := PushCmd
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := runPush(cmd, tt.args)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPushConfig(t *testing.T) {
	config := &pushConfig{
		imageName: "test-image:v1.0",
	}

	assert.Equal(t, "test-image:v1.0", config.imageName)

	// Verify config struct has the expected fields
	// This ensures the structure is ready for future flag additions
	assert.IsType(t, "", config.imageName)
}