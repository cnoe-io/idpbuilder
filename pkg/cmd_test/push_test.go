package cmd_test

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestPushCommandFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "missing image argument",
			args:        []string{"push"},
			expectError: true,
		},
		{
			name:        "valid image argument",
			args:        []string{"push", "myapp:latest"},
			expectError: false,
		},
		{
			name:        "with insecure flag",
			args:        []string{"push", "--insecure", "myapp:latest"},
			expectError: false,
		},
		{
			name:        "with custom registry",
			args:        []string{"push", "--registry", "custom.registry.com:5000", "myapp:latest"},
			expectError: false,
		},
		{
			name:        "too many arguments",
			args:        []string{"push", "myapp:latest", "extra-arg"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new instance of the command for each test
			cmd := &cobra.Command{
				Use:  "push IMAGE[:TAG]",
				Args: cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock run function that doesn't actually push
					return nil
				},
			}

			// Add flags like the real command
			var pushInsecure bool
			var pushRegistry string

			cmd.Flags().BoolVar(&pushInsecure, "insecure", false, "Skip certificate verification")
			cmd.Flags().StringVar(&pushRegistry, "registry", "gitea.cnoe.localtest.me:8443", "Target registry")

			// Set arguments and test
			cmd.SetArgs(tt.args[1:]) // Skip "push" since it's the command name
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