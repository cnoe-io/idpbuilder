package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/cnoe-io/idpbuilder/pkg/certs"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/helpers"
	"github.com/cnoe-io/idpbuilder/pkg/gitea"
	"github.com/spf13/cobra"
)

var (
	pushInsecure bool
	pushRegistry string

	PushCmd = &cobra.Command{
		Use:   "push IMAGE[:TAG]",
		Short: "Push image to Gitea registry",
		Long: `Push a container image to the builtin Gitea registry with certificate support.
The command automatically handles certificate extraction and configuration unless
the --insecure flag is specified.

Examples:
  # Push with automatic certificate handling
  idpbuilder push myapp:latest
  
  # Push to specific registry
  idpbuilder push --registry gitea.cnoe.localtest.me:8443 myapp:latest
  
  # Push without certificate verification (not recommended)
  idpbuilder push --insecure myapp:latest`,
		Args: cobra.ExactArgs(1),
		RunE: runPush,
	}
)

func init() {
	PushCmd.Flags().BoolVar(&pushInsecure, "insecure", false, "Skip certificate verification (not recommended)")
	PushCmd.Flags().StringVar(&pushRegistry, "registry", getDefaultRegistry(), "Target registry")
}

func runPush(cmd *cobra.Command, args []string) error {
	imageRef := args[0]

	// Validate image reference
	if err := validateImageReference(imageRef); err != nil {
		return fmt.Errorf("invalid image reference: %w", err)
	}

	// Progress feedback
	helpers.PrintColoredOutput("Pushing image: %s\n", imageRef)
	helpers.PrintColoredOutput("Target registry: %s\n", pushRegistry)

	// Initialize Gitea client with certificate handling
	var client *gitea.Client
	var err error

	if pushInsecure {
		helpers.PrintColoredOutput("Warning: Running in insecure mode, skipping certificate verification\n")
		client, err = gitea.NewInsecureClient(pushRegistry)
	} else {
		// Use certificate infrastructure from Phase 1
		helpers.PrintColoredOutput("Configuring certificates for secure connection...\n")
		certManager := certs.NewTrustStore()
		// Note: Certificate extraction/configuration is handled internally by registry
		
		client, err = gitea.NewClient(pushRegistry, certManager)
	}

	if err != nil {
		return fmt.Errorf("failed to create registry client: %w", err)
	}

	// Setup progress reporting
	progressChan := make(chan gitea.PushProgress, 10)
	progressDone := make(chan bool)

	// Start progress reporter goroutine
	go func() {
		defer close(progressDone)
		for progress := range progressChan {
			if progress.TotalLayers > 0 {
				helpers.PrintColoredOutput("Pushing layer %d/%d: %d%%\n",
					progress.CurrentLayer, progress.TotalLayers, progress.Percentage)
			} else {
				helpers.PrintColoredOutput("Pushing: %d%% complete\n", progress.Percentage)
			}
		}
	}()

	// Perform the push operation
	helpers.PrintColoredOutput("Starting push operation...\n")
	if err := client.Push(imageRef, progressChan); err != nil {
		close(progressChan)
		<-progressDone
		return fmt.Errorf("push failed: %w", err)
	}

	// Wait for progress reporting to complete
	close(progressChan)
	<-progressDone

	helpers.PrintColoredOutput("Successfully pushed %s to %s\n", imageRef, pushRegistry)
	return nil
}

// validateImageReference ensures the image reference format is valid
func validateImageReference(ref string) error {
	if ref == "" {
		return fmt.Errorf("image reference cannot be empty")
	}

	// Basic validation - should not start/end with special characters
	if strings.HasPrefix(ref, ":") || strings.HasSuffix(ref, ":") {
		return fmt.Errorf("invalid image reference format: %s", ref)
	}

	// Should not contain double colons
	if strings.Contains(ref, "::") {
		return fmt.Errorf("invalid image reference format: %s", ref)
	}

	return nil
}

// getDefaultRegistry returns the default Gitea registry URL
func getDefaultRegistry() string {
	if env := os.Getenv("IDPBUILDER_REGISTRY"); env != "" {
		return env
	}
	return "gitea.cnoe.localtest.me:8443"
}