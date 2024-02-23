package cmd

import (
	"fmt"
	"os"

	"github.com/cnoe-io/idpbuilder/pkg/cmd/create"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/delete"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/list"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "idpbuilder",
	Short: "Manage reference IDPs",
	Long:  "",
}

func init() {
	rootCmd.AddCommand(create.CreateCmd)
	rootCmd.AddCommand(delete.DeleteCmd)
	rootCmd.AddCommand(list.ListCmd)
	rootCmd.AddCommand(version.VersionCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
