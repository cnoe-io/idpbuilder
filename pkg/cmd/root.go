package cmd

import (
	"fmt"
	"os"

	"git.autodesk.com/forge-cd-services/idpbuilder/pkg/cmd/create"
	"git.autodesk.com/forge-cd-services/idpbuilder/pkg/cmd/gitserver"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ucp-dev",
	Short: "Development tooling for the Unified Control Plane",
	Long:  "",
}

func init() {
	rootCmd.AddCommand(create.CreateCmd)
	rootCmd.AddCommand(gitserver.GitServerCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
