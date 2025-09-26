package push

import (
	"context"
	"fmt"

	"github.com/cnoe-io/idpbuilder/pkg/auth"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/helpers"
	"github.com/spf13/cobra"
)

// PushCmd represents the push command
var PushCmd = &cobra.Command{
	Use:   "push [IMAGE]",
	Short: "Push container images to a registry",
	Long: `Push container images to a registry with authentication support.

Examples:
  # Push an image without authentication
  idpbuilder push myimage:latest

  # Push an image with username and password
  idpbuilder push myimage:latest --username myuser --password mypass

  # Push an image with short flags
  idpbuilder push myimage:latest -u myuser -p mypass`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPush(cmd, cmd.Context(), args[0])
	},
}

func init() {
	// Add authentication flags to the push command
	auth.AddAuthenticationFlags(PushCmd)

	// Add common flags
	PushCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")
	PushCmd.Flags().Bool("insecure", false, "Allow insecure registry connections")
}

// runPush executes the push command with the provided image name
func runPush(cmd *cobra.Command, ctx context.Context, imageName string) error {
	// Extract credentials from flags
	creds, err := auth.ExtractCredentialsFromFlags(cmd)
	if err != nil {
		return fmt.Errorf("failed to extract credentials: %w", err)
	}

	// Validate credentials
	validator := &auth.DefaultValidator{}
	if err := validator.ValidateCredentials(creds); err != nil {
		return fmt.Errorf("credential validation failed: %w", err)
	}

	// Create auth config
	authConfig := auth.NewAuthConfig(creds)

	// Log authentication status
	if authConfig.Required {
		helpers.CmdLogger.Info("Pushing with authentication", "username", creds.Username)
	} else {
		helpers.CmdLogger.Info("Pushing without authentication")
	}

	// Log what we would push (stub implementation for now)
	helpers.CmdLogger.Info("Push command executed", "image", imageName, "auth_required", authConfig.Required)

	// TODO: Implement actual push logic in Phase 2
	fmt.Printf("Successfully prepared push for image: %s\n", imageName)

	if authConfig.Required {
		fmt.Printf("Authentication configured for user: %s\n", creds.Username)
	} else {
		fmt.Println("No authentication configured")
	}

	return nil
}