package push

import (
	"fmt"

	"github.com/spf13/cobra"
)

var PushCmd = &cobra.Command{
	Use:   "push IMAGE_NAME",
	Short: "Push an OCI image to the integrated Gitea registry",
	Long: `Push an OCI image to the integrated Gitea registry.

The push command uploads container images to the Gitea registry
at https://gitea.cnoe.localtest.me:8443/.

Examples:
  # Push an image with authentication
  idpbuilder push myapp:latest --username admin --password secret

  # Push with insecure TLS (self-signed certificates)
  idpbuilder push myapp:latest --username admin --password secret --insecure`,
	Args: cobra.ExactArgs(1),
	RunE: runPush,
}

// pushConfig holds the configuration for the push command
type pushConfig struct {
	imageName string
	// Placeholder for future flags (auth, TLS) that will be added in subsequent efforts
}

func init() {
	// Future flags will be added in subsequent efforts:
	// - Authentication flags (effort 1.1.2)
	// - TLS configuration flags (effort 1.1.3)
}

// runPush executes the push command
func runPush(cmd *cobra.Command, args []string) error {
	config := &pushConfig{
		imageName: args[0],
	}

	// Validate image name format
	if err := validateImageName(config.imageName); err != nil {
		return fmt.Errorf("invalid image name: %w", err)
	}

	// Log command execution (temporary until implementation)
	cmd.Printf("Pushing image: %s\n", config.imageName)
	cmd.Println("Note: Push functionality will be implemented in Phase 4")

	return nil
}

// validateImageName performs basic validation on the image name
func validateImageName(name string) error {
	if name == "" {
		return fmt.Errorf("image name cannot be empty")
	}

	// Basic validation - will be enhanced in Phase 3
	// For now, just ensure non-empty
	// Future enhancements will validate:
	// - Registry format: [registry/]namespace/name[:tag]
	// - Character restrictions
	// - Tag format validation

	return nil
}