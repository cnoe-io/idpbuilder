package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cnoe-io/idpbuilder/pkg/build"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/helpers"
	"github.com/spf13/cobra"
)

var (
	buildContext  string
	buildTag      string
	buildPlatform string

	BuildCmd = &cobra.Command{
		Use:   "build",
		Short: "Assemble OCI image from context",
		Long: `Assemble a single-layer OCI image from a directory using go-containerregistry.
Certificate handling is applied during the push phase, not during build.

Examples:
  # Build from current directory
  idpbuilder build --tag myapp:latest
  
  # Build from specific context
  idpbuilder build --context ./app --tag myapp:v1.0.0
  
  # Build with platform specification
  idpbuilder build --context ./app --tag myapp:latest --platform linux/amd64`,
		Args: cobra.NoArgs,
		RunE: runBuild,
	}
)

func init() {
	BuildCmd.Flags().StringVarP(&buildContext, "context", "c", ".", "Build context directory")
	BuildCmd.Flags().StringVarP(&buildTag, "tag", "t", "", "Image tag (required)")
	BuildCmd.Flags().StringVar(&buildPlatform, "platform", "linux/amd64", "Target platform")

	BuildCmd.MarkFlagRequired("tag")
}

func runBuild(cmd *cobra.Command, args []string) error {
	// Validate context exists
	absContext, err := filepath.Abs(buildContext)
	if err != nil {
		return fmt.Errorf("failed to resolve context path: %w", err)
	}

	if _, err := os.Stat(absContext); err != nil {
		return fmt.Errorf("context directory not found: %w", err)
	}

	// Validate image tag format
	if err := validateImageTag(buildTag); err != nil {
		return fmt.Errorf("invalid image tag: %w", err)
	}

	// Validate platform format
	osName, arch, err := parsePlatform(buildPlatform)
	if err != nil {
		return fmt.Errorf("invalid platform: %w", err)
	}

	// Progress feedback
	helpers.PrintColoredOutput("Building image from context: %s\n", absContext)
	helpers.PrintColoredOutput("Target tag: %s\n", buildTag)
	helpers.PrintColoredOutput("Platform: %s/%s\n", osName, arch)

	// Call image-builder package from Phase 2 Wave 1
	builder := build.NewBuilder()
	builder.SetPlatform(osName, arch)

	if helpers.LogLevel == "debug" {
		builder.SetVerbose(true)
	}

	image, err := builder.BuildFromContext(absContext, buildTag)
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Success feedback
	helpers.PrintColoredOutput("Successfully built image: %s\n", image.Tag())
	if size := image.Size(); size > 0 {
		helpers.PrintColoredOutput("Image size: %d bytes\n", size)
	}

	helpers.PrintColoredOutput("Build completed successfully!\n")
	return nil
}

// validateImageTag ensures the tag format is correct
func validateImageTag(tag string) error {
	if tag == "" {
		return fmt.Errorf("image tag cannot be empty")
	}

	// Basic validation - tag should not be just a colon
	if tag == ":" {
		return fmt.Errorf("invalid tag format")
	}

	// Allow tags with or without version (registry will handle defaults)
	return nil
}

// parsePlatform validates and parses platform string
func parsePlatform(platform string) (osName, arch string, err error) {
	// Split by slash
	parts := []string{}
	current := ""
	for _, char := range platform {
		if char == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}

	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid platform format: %s (expected OS/ARCH)", platform)
	}

	osName = parts[0]
	arch = parts[1]

	// Validate OS
	validOS := map[string]bool{
		"linux":   true,
		"windows": true,
		"darwin":  true,
	}

	// Validate architecture
	validArch := map[string]bool{
		"amd64": true,
		"arm64": true,
		"arm":   true,
		"386":   true,
	}

	if !validOS[osName] {
		return "", "", fmt.Errorf("unsupported OS: %s", osName)
	}

	if !validArch[arch] {
		return "", "", fmt.Errorf("unsupported architecture: %s", arch)
	}

	return osName, arch, nil
}