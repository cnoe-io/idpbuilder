package push

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cnoe-io/idpbuilder/pkg/daemon"
	"github.com/cnoe-io/idpbuilder/pkg/registry"
	"github.com/spf13/cobra"
)

const (
	// DefaultRegistry is the standard idpbuilder Gitea registry URL
	DefaultRegistry = "https://gitea.cnoe.localtest.me:8443"
)

// PushCmd represents the push command
var PushCmd = &cobra.Command{
	Use:   "push IMAGE",
	Short: "Push a local Docker image to an OCI registry",
	Long: `Push a local Docker image to an OCI-compliant registry.

The push command takes a local Docker image and uploads it to the specified
OCI registry. It integrates with the idpbuilder daemon to verify the image
exists locally before pushing, and handles authentication via flags or
environment variables.

Examples:
  # Push with default registry
  idpbuilder push myimage:latest

  # Push to custom registry with authentication
  idpbuilder push myimage:latest --registry https://registry.example.com --username user --password pass

  # Push with token authentication
  idpbuilder push myimage:latest --registry https://registry.example.com --token mytoken`,
	Args:          cobra.ExactArgs(1),
	RunE:          runPush,
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Command flags (private package-level)
var (
	flagRegistry string
	flagUsername string
	flagPassword string
	flagToken    string
	flagInsecure bool
)

// init registers flags
func init() {
	PushCmd.Flags().StringVarP(&flagRegistry, "registry", "r", DefaultRegistry, "Registry URL")
	PushCmd.Flags().StringVarP(&flagUsername, "username", "u", "", "Registry username")
	PushCmd.Flags().StringVarP(&flagPassword, "password", "p", "", "Registry password")
	PushCmd.Flags().StringVarP(&flagToken, "token", "t", "", "Registry token")
	PushCmd.Flags().BoolVar(&flagInsecure, "insecure", false, "Skip TLS verification")
}

// runPushWithClients is the internal implementation with injectable dependencies
// This allows tests to provide mock clients
func runPushWithClients(cmd *cobra.Command, args []string,
	daemonClient daemon.DaemonClient,
	registryClient registry.RegistryClient) error {
	imageRef := args[0]

	// Setup context with signal handling for Ctrl+C (REQ-013)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGINT and SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Resolve credentials
	env := &DefaultEnvironment{}
	credFlags := CredentialFlags{
		Username: flagUsername,
		Password: flagPassword,
		Token:    flagToken,
	}
	resolver := &DefaultCredentialResolver{}
	_, err := resolver.Resolve(credFlags, env)
	if err != nil {
		return fmt.Errorf("credential resolution failed: %w", err)
	}

	// Check if daemon is running
	if err := daemonClient.Ping(ctx); err != nil {
		return &daemonNotRunningError{err: err}
	}

	// Check if image exists in daemon
	exists, err := daemonClient.ImageExists(ctx, imageRef)
	if err != nil {
		return &daemonError{err: err}
	}
	if !exists {
		return &imageNotFoundError{imageRef: imageRef}
	}

	// Get image from daemon
	_, imageReader, err := daemonClient.GetImage(ctx, imageRef)
	if err != nil {
		return &daemonError{err: err}
	}
	defer imageReader.Close()

	// Build destination reference with registry host
	destRef := buildDestinationRef(flagRegistry, imageRef)

	// Create progress reporter
	progressReporter := &registry.NoOpProgressReporter{}

	// Push to registry
	result, err := registryClient.Push(ctx, imageRef, destRef, progressReporter)
	if err != nil {
		return err
	}

	// Output the pushed reference to stdout (REQ-001)
	fmt.Fprintln(cmd.OutOrStdout(), result.Reference)
	return nil
}

// runPush is the production entry point
func runPush(cmd *cobra.Command, args []string) error {
	// Create clients for dependency injection
	// Note: These will be properly initialized in E1.2.2 and E1.2.3
	var daemonClient daemon.DaemonClient
	var registryClient registry.RegistryClient

	// NOTE: This check ensures production code has properly initialized clients.
	// During testing, use runPushWithClients which receives mock clients directly.
	if daemonClient == nil || registryClient == nil {
		return fmt.Errorf("daemon or registry client not initialized")
	}

	return runPushWithClients(cmd, args, daemonClient, registryClient)
}

// buildDestinationRef constructs the full registry reference
// Takes a registry URL like "https://registry.example.com" and an image ref like "myimage:latest"
// Returns "registry.example.com/myimage:latest"
func buildDestinationRef(registryURL, imageRef string) string {
	host := extractHost(registryURL)
	return fmt.Sprintf("%s/%s", host, imageRef)
}

// extractHost extracts the host:port from a registry URL
// Handles URLs like "https://registry.example.com:8443" -> "registry.example.com:8443"
// and "https://registry.example.com" -> "registry.example.com"
func extractHost(registryURL string) string {
	u, err := url.Parse(registryURL)
	if err != nil {
		// If parsing fails, try to extract manually
		parts := strings.Split(registryURL, "://")
		if len(parts) > 1 {
			return parts[1]
		}
		return registryURL
	}

	// URL.Host includes port if present
	return u.Host
}

// parseImageRef extracts repository and tag from image reference
// "myimage:latest" -> ("myimage", "latest")
// "myimage" -> ("myimage", "")
// "registry.io/myimage:v1.0" -> ("registry.io/myimage", "v1.0")
func parseImageRef(ref string) (repo, tag string) {
	// Find the last colon (tag separator)
	lastColon := strings.LastIndex(ref, ":")

	// Check if colon is part of a registry with port or a tag
	if lastColon > 0 {
		// Check if the part after colon looks like a tag or port number
		potentialTag := ref[lastColon+1:]

		// Port numbers are purely numeric
		// Tags can contain alphanumeric, dots, dashes, underscores
		// Key insight: registry:port/image pattern has "/" after the port

		// If there's a "/" after the colon, it's definitely a port (registry:port/image)
		if strings.Contains(potentialTag, "/") {
			return ref, ""
		}

		// If it's purely numeric, likely a port number at start of ref
		// But we need to check if there's a "/" before the colon
		beforeColon := ref[:lastColon]
		if !strings.Contains(beforeColon, "/") {
			// No "/" before colon: could be "localhost:5000" (port) or "image:tag"
			// Check if numeric - if so, likely a port
			isAllDigits := true
			for _, c := range potentialTag {
				if c < '0' || c > '9' {
					isAllDigits = false
					break
				}
			}
			if isAllDigits && len(potentialTag) <= 5 {
				// Looks like a port number (1-65535 range, max 5 digits)
				return ref, ""
			}
		}

		// It's a tag (including semver like v1.0, alpine3.18)
		return ref[:lastColon], potentialTag
	}

	return ref, ""
}

// Custom error types for proper exit code handling

type daemonNotRunningError struct {
	err error
}

func (e *daemonNotRunningError) Error() string {
	return fmt.Sprintf("daemon not running: %v", e.err)
}

type daemonError struct {
	err error
}

func (e *daemonError) Error() string {
	return fmt.Sprintf("daemon error: %v", e.err)
}

type imageNotFoundError struct {
	imageRef string
}

func (e *imageNotFoundError) Error() string {
	return fmt.Sprintf("image not found: %s", e.imageRef)
}

type authError struct {
	err error
}

func (e *authError) Error() string {
	return fmt.Sprintf("authentication failed: %v", e.err)
}

type registryError struct {
	err error
}

func (e *registryError) Error() string {
	return fmt.Sprintf("registry error: %v", e.err)
}

// exitWithError handles error classification and returns appropriate exit codes
// Exit codes per POSIX conventions:
// 0 = Success
// 1 = General error, auth failure, registry error
// 2 = Resource not found (image not found, daemon not running)
// 130 = Interrupted (Ctrl+C)
func exitWithError(err error) int {
	if err == nil {
		return 0
	}

	// Check for context cancellation (Ctrl+C)
	if err == context.Canceled {
		return 130
	}

	// Check for specific error types
	switch err.(type) {
	case *imageNotFoundError, *daemonNotRunningError:
		return 2
	case *authError, *registryError:
		return 1
	case *daemonError:
		return 2
	default:
		return 1
	}
}
