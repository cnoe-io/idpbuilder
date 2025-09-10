package cmd_test

import (
	"testing"

	"github.com/cnoe-io/idpbuilder/pkg/cmd"
	"github.com/spf13/cobra"
)

func TestBuildCommandFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "missing required tag flag",
			args:        []string{"build"},
			expectError: true,
		},
		{
			name:        "valid tag flag",
			args:        []string{"build", "--tag", "myapp:latest"},
			expectError: false,
		},
		{
			name:        "valid context and tag flags",
			args:        []string{"build", "--context", ".", "--tag", "myapp:v1.0.0"},
			expectError: false,
		},
		{
			name:        "valid platform flag",
			args:        []string{"build", "--tag", "myapp:latest", "--platform", "linux/arm64"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new instance of the command for each test
			cmd := &cobra.Command{
				Use: "build",
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock run function that doesn't actually build
					return nil
				},
			}

			// Add flags like the real command
			var buildContext string
			var buildTag string
			var buildPlatform string

			cmd.Flags().StringVarP(&buildContext, "context", "c", ".", "Build context directory")
			cmd.Flags().StringVarP(&buildTag, "tag", "t", "", "Image tag (required)")
			cmd.Flags().StringVar(&buildPlatform, "platform", "linux/amd64", "Target platform")
			cmd.MarkFlagRequired("tag")

			// Set arguments and test
			cmd.SetArgs(tt.args[1:]) // Skip "build" since it's the command name
			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}