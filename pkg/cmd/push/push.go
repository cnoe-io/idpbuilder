package push

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	insecureUsage = "Skip TLS certificate verification (use for self-signed certificates)"
)

var (
	// Flags
	insecure bool
)

var PushCmd = &cobra.Command{
	Use:          "push",
	Short:        "Push OCI artifacts to Gitea registry",
	Long:         `Push OCI artifacts to Gitea registry with support for TLS configuration`,
	RunE:         pushRun,
	SilenceUsage: true,
}

func init() {
	// Add insecure flag for TLS configuration
	PushCmd.Flags().BoolVar(&insecure, "insecure", false, insecureUsage)
}

func pushRun(cmd *cobra.Command, args []string) error {
	// Display warning when insecure mode is used
	if insecure {
		fmt.Printf("⚠️  WARNING: Running in insecure mode - TLS certificate verification disabled\n")
		fmt.Printf("   This should only be used with self-signed certificates in development environments\n\n")
	}

	// TODO: Implement actual push logic in future efforts
	fmt.Printf("Push command executed with insecure flag: %v\n", insecure)

	return nil
}