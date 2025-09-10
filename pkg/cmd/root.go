package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/cnoe-io/idpbuilder/pkg/cmd/create"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/delete"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/get"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/helpers"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "idpbuilder",
	Short: "IDP Builder OCI image management and reference IDP operations",
	Long: `IDP Builder provides tools to build and push OCI images to the embedded 
Gitea registry with automatic certificate handling, as well as manage reference IDPs.`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&helpers.LogLevel, "log-level", "l", "info", helpers.LogLevelMsg)
	rootCmd.PersistentFlags().BoolVar(&helpers.ColoredOutput, "color", false, helpers.ColoredOutputMsg)
	
	// Existing commands for IDP management
	rootCmd.AddCommand(create.CreateCmd)
	rootCmd.AddCommand(get.GetCmd)
	rootCmd.AddCommand(delete.DeleteCmd)
	rootCmd.AddCommand(version.VersionCmd)
	
	// New OCI image commands
	rootCmd.AddCommand(BuildCmd)
	rootCmd.AddCommand(PushCmd)
}

func Execute(ctx context.Context) {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
